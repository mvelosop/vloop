#!/bin/sh
# The gates' stub copilot. It never touches the network. It answers
# --version. For a -p prompt it accepts only /vloop:review <task>; its role is
# "second" when --model is the one in second.model, else "review". It appends
# to events:
#
#   session copilot-<role> <task> model=<model>
#   deny-tool <pattern>          every --deny-tool pattern (--deny-tool=P form,
#                                or --deny-tool P [P...])
#   hooks <n> bash=<yes|no> powershell=<yes|no>
#                                n: the preToolUse entries found in the JSON
#                                files of $COPILOT_HOME/hooks/; bash and
#                                powershell: whether every entry has that
#                                command form
#   hook-phase <phase>           the word after --phase in the bash form
#   hook <label> <decision> <exit>
#                                the bash form of the first entry, run as
#                                Copilot runs a hook: sh -c, from the
#                                repository root, one preToolUse payload on
#                                stdin; decision is the permissionDecision of
#                                its stdout, "none" when there is none
#
# It then writes a passing verdict and prints JSON Lines like Copilot's: turns,
# a usage checkpoint, four tool calls — c1 (bash) denied with a message naming
# a path under $HOME, c2 (view) completed, c3 (create) failed with code
# "failure", c4 (web_fetch) denied — and a result event. With the file
# copilot.nodenials it makes no tool calls.
dir=$(dirname "$0")
case "${1:-}" in
  --version) echo "GitHub Copilot CLI 1.0.103"; exit 0;;
esac
prompt="" MODEL="" indeny=0
pats=""
while [ $# -gt 0 ]; do
  if [ "$indeny" = 1 ]; then
    case "$1" in -*) indeny=0;; *) echo "deny-tool $1" >>"$dir/events.tmp.$$"; shift; continue;; esac
  fi
  case "$1" in
    -p) prompt="${2:-}"; shift 2;;
    --model) MODEL="${2:-}"; shift 2;;
    --deny-tool=*) echo "deny-tool ${1#--deny-tool=}" >>"$dir/events.tmp.$$"; shift;;
    --deny-tool) indeny=1; shift;;
    *) shift;;
  esac
done
cmd="${prompt%% *}" TASK="${prompt#* }"
if [ "$cmd" != /vloop:review ] || [ -z "$TASK" ] || [ "$TASK" = "$prompt" ]; then
  echo "stub copilot: unrecognised prompt: $prompt" >&2; rm -f "$dir/events.tmp.$$"; exit 64
fi
ROLE=review
[ "$MODEL" = "$(cat "$dir/second.model" 2>/dev/null)" ] && ROLE=second
here=$(pwd -P)
{
  echo "session copilot-$ROLE $TASK model=$MODEL"
  cat "$dir/events.tmp.$$" 2>/dev/null
} >>"$dir/events"
rm -f "$dir/events.tmp.$$"

# The hooks file(s): every preToolUse entry of every JSON file in hooks/.
entries=$(cat "${COPILOT_HOME:-/nonexistent}"/hooks/*.json 2>/dev/null |
  jq -c '.. | objects | select(has("preToolUse")) | .preToolUse | if type == "array" then .[] else . end | objects' 2>/dev/null)
n=0; [ -n "$entries" ] && n=$(printf '%s\n' "$entries" | grep -c .)
hb=no; hp=no
if [ "$n" -gt 0 ]; then
  [ "$(printf '%s\n' "$entries" | jq -s 'all(.bash | type == "string" and length > 0)')" = true ] && hb=yes
  [ "$(printf '%s\n' "$entries" | jq -s 'all(.powershell | type == "string" and length > 0)')" = true ] && hp=yes
fi
echo "hooks $n bash=$hb powershell=$hp" >>"$dir/events"
hookcmd=$(printf '%s\n' "$entries" | head -1 | jq -r '.bash // empty' 2>/dev/null)
if [ -n "$hookcmd" ]; then
  echo "hook-phase $(printf '%s\n' "$hookcmd" | awk '{ for (i = 1; i < NF; i++) if ($i == "--phase") { v = $(i + 1); gsub(/["'\'']/, "", v); print v; exit } }')" >>"$dir/events"
  # hook <label> <toolName> <toolArgs JSON>
  hook() {
    p=$(jq -cn --arg n "$2" --argjson a "$3" --arg c "$here" '{timestamp: 1, cwd: $c, toolName: $n, toolArgs: $a}')
    out=$(printf '%s' "$p" | sh -c "$hookcmd" 2>/dev/null); x=$?
    d=$(printf '%s\n' "$out" | head -1 | jq -r '.permissionDecision // "none"' 2>/dev/null)
    echo "hook $1 ${d:-none} $x" >>"$dir/events"
  }
  hook shgh bash '{"command":"sh -c \"gh pr list\"","description":"d"}'
  hook envgh bash '{"command":"env gh auth token","description":"d"}'
  hook diff bash '{"command":"git diff HEAD","description":"d"}'
  hook state-abs create "$(jq -cn --arg p "$here/.vloop/state/x.json" '{path: $p, file_text: "x"}')"
  hook tmp-abs create "$(jq -cn --arg p "$here/.vloop/tmp/x.json" '{path: $p, file_text: "x"}')"
  hook tmp-rel create '{"path":".vloop/tmp/y.json","file_text":"x"}'
  hook webfetch web_fetch '{"url":"https://example.invalid"}'
fi

mkdir -p .vloop/tmp
printf '{"schema":"verdict/v1","task":"%s","verdict":"PASS","criteria":[],"findings":[],"notes":"copilot"}\n' "$TASK" >.vloop/tmp/verdict.json
echo "stub copilot review session" >&2
printf '{"type":"session.start","data":{"sessionId":"stub-session"}}\n'
printf '{"type":"assistant.turn_start","data":{"turnId":"1"}}\n'
printf '{"type":"session.usage_checkpoint","data":{"totalNanoAiu":5100000000}}\n'
if [ ! -f "$dir/copilot.nodenials" ]; then
  printf '{"type":"tool.execution_start","data":{"toolCallId":"c1","toolName":"bash","arguments":{"command":"sh -c \\"gh pr list\\""}}}\n'
  printf '{"type":"tool.execution_start","data":{"toolCallId":"c2","toolName":"view","arguments":{"path":"docs/x.md"}}}\n'
  printf '{"type":"tool.execution_complete","data":{"toolCallId":"c2","success":true}}\n'
  printf '{"type":"tool.execution_complete","data":{"toolCallId":"c1","success":false,"error":{"code":"denied","message":"Denied by preToolUse hook: vloop fence: gh is denied (seen from %s/notes)"}}}\n' "$HOME"
  printf '{"type":"tool.execution_start","data":{"toolCallId":"c3","toolName":"create","arguments":{"path":"x"}}}\n'
  printf '{"type":"tool.execution_complete","data":{"toolCallId":"c3","success":false,"error":{"code":"failure","message":"boom"}}}\n'
  printf '{"type":"tool.execution_start","data":{"toolCallId":"c4","toolName":"web_fetch","arguments":{"url":"https://example.invalid"}}}\n'
  printf '{"type":"tool.execution_complete","data":{"toolCallId":"c4","success":false,"error":{"code":"denied","message":"Permission to access this URL was denied."}}}\n'
fi
printf '{"type":"assistant.turn_start","data":{"turnId":"2"}}\n'
printf '{"type":"session.usage_checkpoint","data":{"totalNanoAiu":8400000000}}\n'
printf '{"type":"result","sessionId":"stub-session","exitCode":0,"usage":{"premiumRequests":1,"totalApiDurationMs":12000,"sessionDurationMs":18000}}\n'
exit 0

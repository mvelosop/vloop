#!/bin/sh
# The gates' stub copilot. It never touches the network. It answers
# --version. For -p "/vloop:review-brief <path>" (role review-brief) it
# appends to events beside the stub:
#
#   session copilot-<role> <arg> model=<model>
#   deny-tool <role> <pattern>   every --deny-tool pattern (--deny-tool=P form,
#                                or --deny-tool P [P...])
#   hooks <role> <n> bash=<yes|no> powershell=<yes|no>
#                                n: the preToolUse entries found in the JSON
#                                files of $COPILOT_HOME/hooks/; bash and
#                                powershell: whether every entry has that form
#   hookfile <role> <phases>     the phases named after "fence hook --phase "
#                                in those files, sorted, space-separated
#                                ("none" when there is none)
#   hook <role> <label> <decision> <exit>
#                                the bash form of the first entry, run as
#                                Copilot runs a hook: sh -c, from the
#                                repository root, one preToolUse payload on
#                                stdin; decision is the permissionDecision of
#                                its stdout, "none" when there is none
#
# It then copies report.json beside the stub, when there is one, to
# .vloop/tmp/brief-review.json and prints JSON Lines like Copilot's: turns,
# usage checkpoints, one tool call denied by the hook and one completed, and a
# result event. Any other prompt exits 64.
dir=$(dirname "$0")
case "${1:-}" in
  --version) echo "GitHub Copilot CLI 1.0.103"; exit 0;;
esac
prompt="" MODEL="" indeny=0
tmpev="$dir/events.tmp.$$"
: >"$tmpev"
while [ $# -gt 0 ]; do
  if [ "$indeny" = 1 ]; then
    case "$1" in -*) indeny=0;; *) echo "$1" >>"$tmpev"; shift; continue;; esac
  fi
  case "$1" in
    -p) prompt="${2:-}"; shift 2;;
    --model) MODEL="${2:-}"; shift 2;;
    --deny-tool=*) echo "${1#--deny-tool=}" >>"$tmpev"; shift;;
    --deny-tool) indeny=1; shift;;
    *) shift;;
  esac
done
cmd="${prompt%% *}" ARG="${prompt#* }"
if [ "$cmd" != /vloop:review-brief ] || [ "$ARG" = "$prompt" ]; then
  echo "stub copilot: unrecognised prompt: $prompt" >&2; rm -f "$tmpev"; exit 64
fi
ROLE=review-brief
here=$(pwd -P)
{
  echo "session copilot-$ROLE $ARG model=$MODEL"
  sed "s/^/deny-tool $ROLE /" "$tmpev"
} >>"$dir/events"
rm -f "$tmpev"
phases=$(cat "${COPILOT_HOME:-/nonexistent}"/hooks/*.json 2>/dev/null | grep -oE 'fence hook --phase [a-z-]+' | sed 's/^fence hook --phase //' | LC_ALL=C sort -u | tr '\n' ' ' | sed 's/ $//')
entries=$(cat "${COPILOT_HOME:-/nonexistent}"/hooks/*.json 2>/dev/null |
  jq -c '.. | objects | select(has("preToolUse")) | .preToolUse | if type == "array" then .[] else . end | objects' 2>/dev/null)
n=0; [ -n "$entries" ] && n=$(printf '%s\n' "$entries" | grep -c .)
hb=no; hp=no
if [ "$n" -gt 0 ]; then
  [ "$(printf '%s\n' "$entries" | jq -s 'all(.bash | type == "string" and length > 0)')" = true ] && hb=yes
  [ "$(printf '%s\n' "$entries" | jq -s 'all(.powershell | type == "string" and length > 0)')" = true ] && hp=yes
fi
{
  echo "hooks $ROLE $n bash=$hb powershell=$hp"
  echo "hookfile $ROLE ${phases:-none}"
} >>"$dir/events"
hookcmd=$(printf '%s\n' "$entries" | head -1 | jq -r '.bash // empty' 2>/dev/null)
if [ -n "$hookcmd" ]; then
  hook() {
    p=$(jq -cn --arg n "$2" --argjson a "$3" --arg c "$here" '{timestamp: 1, cwd: $c, toolName: $n, toolArgs: $a}')
    out=$(printf '%s' "$p" | sh -c "$hookcmd" 2>/dev/null); x=$?
    d=$(printf '%s\n' "$out" | head -1 | jq -r '.permissionDecision // "none"' 2>/dev/null)
    echo "hook $ROLE $1 ${d:-none} $x" >>"$dir/events"
  }
  bashargs() { jq -cn --arg c "$1" '{command: $c, description: "d"}'; }
  fileargs() { jq -cn --arg p "$1" '{path: $p, file_text: "x"}'; }
  hook check bash "$(bashargs 'vloop brief check .vloop/tmp/rb/B20990101-0000-x.loop-brief.md')"
  hook cat bash "$(bashargs "cat $ARG")"
  hook report create "$(fileargs .vloop/tmp/brief-review.json)"
  hook briefw create "$(fileargs "$ARG")"
  hook commit bash "$(bashargs 'git commit -m x')"
  hook rmbrief bash "$(bashargs "rm $ARG")"
  hook skillrb skill '{"skill":"review-brief"}'
  hook skillother skill '{"skill":"vloop-review-brief"}'
  hook nested bash "$(bashargs "vloop brief review $ARG")"
  hook webfetch web_fetch '{"url":"https://example.invalid"}'
fi
mkdir -p .vloop/tmp
[ -f "$dir/report.json" ] && cp "$dir/report.json" .vloop/tmp/brief-review.json
echo "stub copilot $ROLE session" >&2
printf '{"type":"session.start","data":{"sessionId":"stub-session"}}\n'
printf '{"type":"assistant.turn_start","data":{"turnId":"1"}}\n'
printf '{"type":"session.usage_checkpoint","data":{"totalNanoAiu":5100000000}}\n'
printf '{"type":"tool.execution_start","data":{"toolCallId":"c1","toolName":"bash","arguments":{"command":"x"}}}\n'
printf '{"type":"tool.execution_complete","data":{"toolCallId":"c1","success":false,"error":{"code":"denied","message":"Denied by preToolUse hook: vloop fence: git commit -m x is denied by the review-brief fence (Bash(git commit:*))"}}}\n'
printf '{"type":"tool.execution_start","data":{"toolCallId":"c2","toolName":"view","arguments":{"path":"docs/x.md"}}}\n'
printf '{"type":"tool.execution_complete","data":{"toolCallId":"c2","success":true}}\n'
printf '{"type":"assistant.turn_start","data":{"turnId":"2"}}\n'
printf '{"type":"session.usage_checkpoint","data":{"totalNanoAiu":8400000000}}\n'
printf '{"type":"result","sessionId":"stub-session","exitCode":0,"usage":{"premiumRequests":1,"totalApiDurationMs":12000,"sessionDurationMs":18000}}\n'
exit 0

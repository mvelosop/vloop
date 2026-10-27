#!/bin/sh
# The gates' stub copilot. It never touches the network. It answers
# --version. For a -p prompt it accepts /vloop:work <task> (role work) and
# /vloop:review <task> (role review), and appends to events:
#
#   session copilot-<role> <task> model=<model>
#   deny-tool <role> <pattern>   every --deny-tool pattern (--deny-tool=P form,
#                                or --deny-tool P [P...])
#   hooks <role> <n> bash=<yes|no> powershell=<yes|no>
#                                n: the preToolUse entries found in the JSON
#                                files of $COPILOT_HOME/hooks/; bash and
#                                powershell: whether every entry has that form
#   hookfile <role> work=<yes|no> review=<yes|no>
#                                whether a file in $COPILOT_HOME/hooks/ holds
#                                "fence hook --phase work" / "--phase review"
#   hook-phase <role> <phase>    the word after --phase in the bash form
#   hook <role> <label> <decision> <exit>
#                                the bash form of the first entry, run as
#                                Copilot runs a hook: sh -c, from the
#                                repository root, one preToolUse payload on
#                                stdin; decision is the permissionDecision of
#                                its stdout, "none" when there is none
#
# A work session then writes <task>.out and a done proposal under .vloop/tmp/;
# with the file copilot.moveref beside it, it first makes a commit (HEAD
# moves). A review session writes a passing verdict. Both print JSON Lines
# like Copilot's: turns, usage checkpoints, one tool call denied by the hook
# (c1, bash) and one completed (c2, view), and a result event.
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
cmd="${prompt%% *}" TASK="${prompt#* }"
case "$cmd" in
  /vloop:work) ROLE=work;;
  /vloop:review) ROLE=review;;
  *) ROLE="";;
esac
if [ -z "$ROLE" ] || [ -z "$TASK" ] || [ "$TASK" = "$prompt" ]; then
  echo "stub copilot: unrecognised prompt: $prompt" >&2; rm -f "$tmpev"; exit 64
fi
here=$(pwd -P)
{
  echo "session copilot-$ROLE $TASK model=$MODEL"
  sed "s/^/deny-tool $ROLE /" "$tmpev"
} >>"$dir/events"
rm -f "$tmpev"

hw=no; hr=no
grep -qF 'fence hook --phase work' "${COPILOT_HOME:-/nonexistent}"/hooks/*.json 2>/dev/null && hw=yes
grep -qF 'fence hook --phase review' "${COPILOT_HOME:-/nonexistent}"/hooks/*.json 2>/dev/null && hr=yes
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
  echo "hookfile $ROLE work=$hw review=$hr"
} >>"$dir/events"
hookcmd=$(printf '%s\n' "$entries" | head -1 | jq -r '.bash // empty' 2>/dev/null)
if [ -n "$hookcmd" ]; then
  echo "hook-phase $ROLE $(printf '%s\n' "$hookcmd" | awk '{ for (i = 1; i < NF; i++) if ($i == "--phase") { v = $(i + 1); gsub(/["'\'']/, "", v); print v; exit } }')" >>"$dir/events"
  # hook <label> <toolName> <toolArgs JSON>
  hook() {
    p=$(jq -cn --arg n "$2" --argjson a "$3" --arg c "$here" '{timestamp: 1, cwd: $c, toolName: $n, toolArgs: $a}')
    out=$(printf '%s' "$p" | sh -c "$hookcmd" 2>/dev/null); x=$?
    d=$(printf '%s\n' "$out" | head -1 | jq -r '.permissionDecision // "none"' 2>/dev/null)
    echo "hook $ROLE $1 ${d:-none} $x" >>"$dir/events"
  }
  bashargs() { jq -cn --arg c "$1" '{command: $c, description: "d"}'; }
  fileargs() { jq -cn --arg p "$1" '{path: $p, file_text: "x"}'; }
  if [ "$ROLE" = work ]; then
    hook rmfile bash "$(bashargs 'rm src/old.ts')"
    hook rmrf bash "$(bashargs 'rm -rf dist')"
    hook shpush bash "$(bashargs 'sh -c "git push origin main"')"
    hook envgh bash "$(bashargs 'env gh pr list')"
    hook srcrel create "$(fileargs src/a.ts)"
    hook srcabs create "$(fileargs "$here/src/b.ts")"
    hook tmprel create "$(fileargs .vloop/tmp/y.json)"
    hook stateabs create "$(fileargs "$here/.vloop/state/x.json")"
    hook gitcfg edit "$(fileargs .git/config)"
    hook outside create "$(fileargs "$here/../outside.txt")"
    hook webfetch web_fetch '{"url":"https://example.invalid"}'
  else
    hook shgh bash "$(bashargs 'sh -c "gh pr list"')"
    hook diff bash "$(bashargs 'git diff HEAD')"
    hook rmfile bash "$(bashargs 'rm src/old.ts')"
    hook srcrel create "$(fileargs src/a.ts)"
    hook tmprel create "$(fileargs .vloop/tmp/y.json)"
  fi
fi

mkdir -p .vloop/tmp
if [ "$ROLE" = work ]; then
  [ -f "$dir/copilot.moveref" ] && git commit -q --allow-empty -m "moved by the stub work session"
  echo "$TASK by copilot" >>"$TASK.out"
  printf '{"schema":"proposal/v1","task":"%s","outcome":"done","summary":"made %s","files":[],"verified":"ok","notes":"none"}\n' "$TASK" "$TASK" >.vloop/tmp/proposal.json
  denied="rm -rf dist is denied by the work fence (Bash(rm -rf:*))"
else
  printf '{"schema":"verdict/v1","task":"%s","verdict":"PASS","criteria":[],"findings":[],"notes":"copilot"}\n' "$TASK" >.vloop/tmp/verdict.json
  denied="gh pr list is denied by the Copilot deny list (gh:*)"
fi
echo "stub copilot $ROLE session" >&2
printf '{"type":"session.start","data":{"sessionId":"stub-session"}}\n'
printf '{"type":"assistant.turn_start","data":{"turnId":"1"}}\n'
printf '{"type":"session.usage_checkpoint","data":{"totalNanoAiu":5100000000}}\n'
printf '{"type":"tool.execution_start","data":{"toolCallId":"c1","toolName":"bash","arguments":{"command":"x"}}}\n'
printf '{"type":"tool.execution_complete","data":{"toolCallId":"c1","success":false,"error":{"code":"denied","message":"Denied by preToolUse hook: vloop fence: %s"}}}\n' "$denied"
printf '{"type":"tool.execution_start","data":{"toolCallId":"c2","toolName":"view","arguments":{"path":"docs/x.md"}}}\n'
printf '{"type":"tool.execution_complete","data":{"toolCallId":"c2","success":true}}\n'
printf '{"type":"assistant.turn_start","data":{"turnId":"2"}}\n'
printf '{"type":"session.usage_checkpoint","data":{"totalNanoAiu":8400000000}}\n'
printf '{"type":"result","sessionId":"stub-session","exitCode":0,"usage":{"premiumRequests":1,"totalApiDurationMs":12000,"sessionDurationMs":18000}}\n'
exit 0

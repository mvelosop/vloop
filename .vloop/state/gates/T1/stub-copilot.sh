#!/bin/sh
# The B34 gates' stub copilot. It never touches the network and never runs
# the fence hook. It answers --version. For a -p prompt it plays one session
# of the phase the hooks file in $COPILOT_HOME/hooks/ names after
# "fence hook --phase " (the prompt alone cannot tell a second review from a
# review), and appends to events beside it:
#
#   session <phase> <arg|->
#   hookbin <phase> <path>      the binary the bash form of the hook runs
#   timeout <phase> <n>         the hook entry's timeoutSec ("none" if absent)
#
# A plan session copies plan.json beside the stub to the plan file; a
# gate-review session writes a passing gate-verdict/v2; a work session writes
# <task>.out and a done proposal; a review (or second-review) session writes
# a passing verdict; a verify session writes a verify/v2 report with no
# findings; a review-brief session writes a PASS brief-review/v1 report.
#
# Files beside the stub change what a session of <phase> does:
#   calls.<phase>     JSON Lines of tool events, printed among its events
#   moveref.<phase>   it first makes an empty commit (HEAD moves)
#   touch.<phase>     it appends a line to the file named in touch.<phase>
#   hook.<phase>      "replace": the hook's binary is replaced (renamed over)
#                     by a changed copy; "remove": it is deleted
#   silent.<phase>    it prints no result event and exits 1
dir=$(dirname "$0")
case "${1:-}" in
  --version) echo "GitHub Copilot CLI 1.0.103"; exit 0;;
esac
prompt="" MODEL=""
while [ $# -gt 0 ]; do
  case "$1" in
    -p) prompt="${2:-}"; shift 2;;
    --model) MODEL="${2:-}"; shift 2;;
    *) shift;;
  esac
done
cmd="${prompt%% *}" ARG="${prompt#* }"
[ "$ARG" = "$prompt" ] && ARG=-
case "$cmd" in
  /vloop:plan|/vloop:work|/vloop:review|/vloop:review-brief) [ "$ARG" != - ] || { echo "stub copilot: unrecognised prompt: $prompt" >&2; exit 64; };;
  /vloop:gate-review|/vloop:verify) [ "$ARG" = - ] || { echo "stub copilot: unrecognised prompt: $prompt" >&2; exit 64; };;
  *) echo "stub copilot: unrecognised prompt: $prompt" >&2; exit 64;;
esac
hooks=$(cat "${COPILOT_HOME:-/nonexistent}"/hooks/*.json 2>/dev/null)
PHASE=$(printf '%s\n' "$hooks" | grep -oE 'fence hook --phase [a-z-]+' | head -1 | sed 's/^fence hook --phase //')
[ -n "$PHASE" ] || PHASE=${cmd#/vloop:}
hookcmd=$(printf '%s\n' "$hooks" | jq -r '.hooks.preToolUse[0].bash // empty' 2>/dev/null)
hookbin=$(printf '%s\n' "$hookcmd" | sed -n "s/^'\([^']*\)'.*/\1/p")
tsec=$(printf '%s\n' "$hooks" | jq -r '.hooks.preToolUse[0].timeoutSec // "none"' 2>/dev/null)
{
  echo "session $PHASE $ARG"
  echo "hookbin $PHASE ${hookbin:-none}"
  echo "timeout $PHASE ${tsec:-none}"
} >>"$dir/events"

if [ -f "$dir/hook.$PHASE" ] && [ -n "$hookbin" ]; then
  case $(cat "$dir/hook.$PHASE") in
    replace) cp "$hookbin" "$hookbin.new" && printf 'changed\n' >>"$hookbin.new" && mv -f "$hookbin.new" "$hookbin";;
    remove) rm -f "$hookbin";;
  esac
fi
[ -f "$dir/moveref.$PHASE" ] && git commit -q --allow-empty -m "moved by the stub $PHASE session"
[ -f "$dir/touch.$PHASE" ] && echo "written by the stub $PHASE session" >>"$(cat "$dir/touch.$PHASE")"

mkdir -p .vloop/tmp
case "$PHASE" in
  plan) mkdir -p .vloop/state && cp "$dir/plan.json" .vloop/state/state.json;;
  gate-review) echo '{"schema":"gate-verdict/v2","verdict":"PASS","tasks":[{"task":"T1","verdict":"PASS","findings":[]}],"notes":"stub"}' >.vloop/tmp/gate-verdict.json;;
  work) echo "$ARG by copilot" >>"$ARG.out"
    printf '{"schema":"proposal/v1","task":"%s","outcome":"done","summary":"made %s","files":[],"verified":"ok","notes":"none"}\n' "$ARG" "$ARG" >.vloop/tmp/proposal.json;;
  review|second-review) printf '{"schema":"verdict/v1","task":"%s","verdict":"PASS","criteria":[],"findings":[],"notes":"stub"}\n' "$ARG" >.vloop/tmp/verdict.json;;
  verify) echo '{"schema":"verify/v2","complete":true,"findings":[],"not_run":[],"checked":{"example_lines":1,"sentences":0,"surfaces":0},"notes":"stub"}' >.vloop/tmp/verify.json;;
  review-brief) b=$(basename "$ARG" .md)
    printf '{"schema":"brief-review/v1","brief":"%s","verdict":"PASS","findings":[]}\n' "$b" >.vloop/tmp/brief-review.json;;
esac
echo "stub copilot $PHASE session" >&2
printf '{"type":"session.start","data":{"sessionId":"stub-session"}}\n'
printf '{"type":"assistant.turn_start","data":{"turnId":"1"}}\n'
printf '{"type":"session.usage_checkpoint","data":{"totalNanoAiu":5100000000}}\n'
[ -f "$dir/calls.$PHASE" ] && cat "$dir/calls.$PHASE"
printf '{"type":"assistant.turn_start","data":{"turnId":"2"}}\n'
if [ -f "$dir/silent.$PHASE" ]; then exit 1; fi
printf '{"type":"result","sessionId":"stub-session","exitCode":0,"usage":{"premiumRequests":1,"totalApiDurationMs":12000,"sessionDurationMs":18000}}\n'
exit 0

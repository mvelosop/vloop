#!/bin/sh
# The gates' stub claude. It never touches the network. It answers doctor's
# two probes and, for a -p prompt, plays the phase, appending to events:
#
#   session <role> <task|-> model=<model>    role: plan, gate-review, work,
#                                            review or verify
#   fence <role> <task> ok|bad|none          for a work or review session:
#                                            whether --settings names a file
#                                            whose JSON equals the work fence
#                                            (work) or the review fence (review)
#
# A work session writes <task>.out and a done proposal; a review passes.
dir=$(dirname "$0")
case "${1:-}" in
  --version) echo "2.0.0 (Claude Code)"; exit 0;;
  plugin) echo '[{"id":"vloop@vloop","version":"@PLUGIN_VERSION@","enabled":true}]'; exit 0;;
esac
prompt="" MODEL="" SETTINGS=""
while [ $# -gt 0 ]; do
  case "$1" in
    -p) prompt="${2:-}"; shift 2;;
    --model) MODEL="${2:-}"; shift 2;;
    --settings) SETTINGS="${2:-}"; shift 2;;
    *) shift;;
  esac
done
cmd="${prompt%% *}" ARG="${prompt#* }"
[ "$ARG" = "$prompt" ] && ARG=-
mkdir -p .vloop/tmp
fence() {
  r=none
  if [ -n "$SETTINGS" ]; then
    r=bad
    [ -f "$SETTINGS" ] && [ "$(jq -S . "$SETTINGS" 2>/dev/null)" = "$(jq -S . "$dir/$1-fence.json")" ] && r=ok
  fi
  echo "fence $1 $ARG $r" >>"$dir/events"
}
case "$cmd" in
  /vloop:plan) ROLE=plan; mkdir -p .vloop/state && cp "$dir/plan.json" .vloop/state/state.json;;
  /vloop:gate-review) ROLE=gate-review
    echo '{"schema":"gate-verdict/v2","verdict":"PASS","tasks":[],"notes":"none"}' >.vloop/tmp/gate-verdict.json;;
  /vloop:work) ROLE=work; fence work; echo "$ARG by claude" >>"$ARG.out"
    printf '{"schema":"proposal/v1","task":"%s","outcome":"done","summary":"made %s","files":[],"verified":"ok","notes":"none"}\n' "$ARG" "$ARG" >.vloop/tmp/proposal.json;;
  /vloop:review) ROLE=review; fence review
    printf '{"schema":"verdict/v1","task":"%s","verdict":"PASS","criteria":[],"findings":[],"notes":"claude"}\n' "$ARG" >.vloop/tmp/verdict.json;;
  /vloop:verify) ROLE=verify
    echo '{"schema":"verify/v1","findings":[],"checked":{"example_lines":0,"sentences":0,"surfaces":0},"notes":"none"}' >.vloop/tmp/verify.json;;
  *) echo "stub claude: unrecognised prompt: $prompt" >&2; exit 64;;
esac
echo "session $ROLE $ARG model=$MODEL" >>"$dir/events"
echo "stub $ROLE session" >&2
printf '{"type":"result","subtype":"success","is_error":false,"duration_ms":1000,"num_turns":3,"total_cost_usd":0.1,"session_id":"stub","permission_denials":[],"modelUsage":{"%s":{"inputTokens":10,"outputTokens":5,"costUSD":0.1}}}\n' "$MODEL"
exit 0

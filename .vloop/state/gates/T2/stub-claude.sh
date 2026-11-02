#!/bin/sh
# The gates' stub claude. It never touches the network. It answers --version
# and `plugin list`, and for a -p prompt plays the phase, appending
# "session <role> <arg|->" to events beside the stub.
#
# A plan session copies plan.json beside the stub to the plan file; a
# gate-review session writes a gate-verdict/v2 PASS; a work session writes
# <task>.out and a done proposal; a review passes. A verify session copies
# verify-report.json beside the stub, when there is one, to
# .vloop/tmp/verify.json. Files beside the stub change what a verify session
# does after that:
#   claude.verify-exit3      it exits 3 after its output
#   claude.verify-norecord   it prints nothing on stdout (no session record)
#                            and exits 0
dir=$(dirname "$0")
case "${1:-}" in
  --version) echo "2.0.0 (Claude Code)"; exit 0;;
  plugin) echo '[{"id":"vloop@vloop","version":"@PLUGIN_VERSION@","enabled":true}]'; exit 0;;
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
mkdir -p .vloop/tmp
xit=0 quiet=
case "$cmd" in
  /vloop:plan) ROLE=plan
    mkdir -p .vloop/state && cp "$dir/plan.json" .vloop/state/state.json;;
  /vloop:gate-review) ROLE=gate-review
    echo '{"schema":"gate-verdict/v2","verdict":"PASS","tasks":[],"notes":"none"}' >.vloop/tmp/gate-verdict.json;;
  /vloop:work) ROLE=work; echo "$ARG by claude" >>"$ARG.out"
    printf '{"schema":"proposal/v1","task":"%s","outcome":"done","summary":"made %s","files":[],"verified":"ok","notes":"none"}\n' "$ARG" "$ARG" >.vloop/tmp/proposal.json;;
  /vloop:review) ROLE=review
    printf '{"schema":"verdict/v1","task":"%s","verdict":"PASS","criteria":[],"findings":[],"notes":"claude"}\n' "$ARG" >.vloop/tmp/verdict.json;;
  /vloop:verify) ROLE=verify
    [ -f "$dir/verify-report.json" ] && cp "$dir/verify-report.json" .vloop/tmp/verify.json
    [ -f "$dir/claude.verify-exit3" ] && xit=3
    [ -f "$dir/claude.verify-norecord" ] && quiet=1;;
  *) echo "stub claude: unrecognised prompt: $prompt" >&2; exit 64;;
esac
echo "session $ROLE $ARG" >>"$dir/events"
echo "stub $ROLE session" >&2
[ -n "$quiet" ] && exit 0
printf '{"type":"result","subtype":"success","is_error":false,"duration_ms":1000,"num_turns":3,"total_cost_usd":0.1,"session_id":"stub","permission_denials":[],"modelUsage":{"%s":{"inputTokens":10,"outputTokens":5,"costUSD":0.1}}}\n' "$MODEL"
exit "$xit"

#!/bin/sh
# The gates' stub claude. Every invocation appends to log: a header, each
# argument on its own line and the working directory. It answers doctor's two
# probes and, for a -p prompt, plays the phase: the plan copies plan.json in;
# the gate review of round n (the n-th gate-review session) copies verdict.<n>.json
# to .vloop/tmp/gate-verdict.json, or writes a gate-verdict/v2 PASS when there
# is no such file; work, review and verify leave a passing answer. It prints
# claude's result JSON and exits 0.
dir=$(dirname "$0")
{
  echo "=== claude"
  for a in "$@"; do printf 'arg %s\n' "$a"; done
  printf 'cwd %s\n' "$(pwd -P)"
} >>"$dir/log"
case "$1" in
  --version) echo "2.0.0 (Claude Code)"; exit 0;;
  plugin) echo '[{"id":"vloop@vloop","version":"@PLUGIN_VERSION@","enabled":true}]'; exit 0;;
esac
prompt="" MODEL=""
while [ $# -gt 0 ]; do
  case "$1" in
    -p) prompt="$2"; shift 2;;
    --model) MODEL="$2"; shift 2;;
    *) shift;;
  esac
done
cmd="${prompt%% *}" ARG="${prompt#* }"
mkdir -p .vloop/tmp
case "$cmd" in
  /vloop:plan) mkdir -p .vloop/state && cp "$dir/plan.json" .vloop/state/state.json;;
  /vloop:gate-review)
    n=$(( $(cat "$dir/gate-review.count" 2>/dev/null || echo 0) + 1 )); echo "$n" >"$dir/gate-review.count"
    if [ -f "$dir/verdict.$n.json" ]; then
      cp "$dir/verdict.$n.json" .vloop/tmp/gate-verdict.json
    else
      echo '{"schema":"gate-verdict/v2","verdict":"PASS","tasks":[{"task":"T1","verdict":"PASS","findings":[]},{"task":"T2","verdict":"PASS","findings":[]}],"notes":"none"}' >.vloop/tmp/gate-verdict.json
    fi;;
  /vloop:work) touch "$ARG.out"
    printf '{"schema":"proposal/v1","task":"%s","outcome":"done","summary":"made %s","files":[],"verified":"ok","notes":"none"}\n' "$ARG" "$ARG" >.vloop/tmp/proposal.json;;
  /vloop:review)
    printf '{"schema":"verdict/v1","task":"%s","verdict":"PASS","criteria":[],"findings":[],"notes":"none"}\n' "$ARG" >.vloop/tmp/verdict.json;;
  /vloop:verify)
    echo '{"schema":"verify/v1","findings":[],"checked":{"example_lines":0,"sentences":0,"surfaces":0},"notes":"none"}' >.vloop/tmp/verify.json;;
  *) echo "stub claude: unrecognised prompt: $prompt" >&2; exit 64;;
esac
printf '{"type":"result","subtype":"success","is_error":false,"duration_ms":1000,"num_turns":3,"total_cost_usd":0.1,"session_id":"stub","permission_denials":[],"modelUsage":{"%s":{"inputTokens":10,"outputTokens":5,"costUSD":0.1}}}\n' "$MODEL"
exit 0

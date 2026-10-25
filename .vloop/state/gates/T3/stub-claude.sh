#!/bin/sh
# The gates' stub claude. Every invocation appends to log, in a form that does
# not depend on the binary that started it: a header, each argument on its own
# line, the working directory, and the full environment sorted by name. It
# answers doctor's two probes (--version fails when the file noversion exists)
# and, for a -p prompt, plays the phase: the plan copies plan.json in; gate
# review, work, review and verify leave a passing answer. It writes one line
# naming the phase and $HOME to stderr, then prints claude's result JSON —
# without total_cost_usd and costUSD when the file nocost exists, nothing when
# silent.<phase> exists, a line that is not JSON when garbage.<phase> exists —
# and exits with the code in exit.<phase>, else 0.
dir=$(dirname "$0")
{
  echo "=== claude"
  for a in "$@"; do printf 'arg %s\n' "$a"; done
  printf 'cwd %s\n' "$(pwd -P)"
  env | LC_ALL=C sort | sed 's/^/env /'
} >>"$dir/log"
case "$1" in
  --version) [ -f "$dir/noversion" ] && { echo "claude: broken" >&2; exit 1; }; echo "2.0.0 (Claude Code)"; exit 0;;
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
  /vloop:plan) PHASE=plan; mkdir -p .vloop/state && cp "$dir/plan.json" .vloop/state/state.json;;
  /vloop:gate-review) PHASE=gate-review
    echo '{"schema":"gate-verdict/v1","verdict":"PASS","tasks":[],"notes":"none"}' >.vloop/tmp/gate-verdict.json;;
  /vloop:work) PHASE=work; touch "$ARG.out"
    printf '{"schema":"proposal/v1","task":"%s","outcome":"done","summary":"made %s","files":[],"verified":"ok","notes":"none"}\n' "$ARG" "$ARG" >.vloop/tmp/proposal.json;;
  /vloop:review) PHASE=review
    printf '{"schema":"verdict/v1","task":"%s","verdict":"PASS","criteria":[],"findings":[],"notes":"none"}\n' "$ARG" >.vloop/tmp/verdict.json;;
  /vloop:verify) PHASE=verify
    echo '{"schema":"verify/v1","findings":[],"checked":{"example_lines":0,"sentences":0,"surfaces":0},"notes":"none"}' >.vloop/tmp/verify.json;;
  *) echo "stub claude: unrecognised prompt: $prompt" >&2; exit 64;;
esac
echo "stub $PHASE session, home $HOME/work" >&2
if [ -f "$dir/silent.$PHASE" ]; then
  :
elif [ -f "$dir/garbage.$PHASE" ]; then
  echo "this is not JSON"
elif [ -f "$dir/nocost" ]; then
  printf '{"type":"result","subtype":"success","is_error":false,"duration_ms":1000,"num_turns":3,"session_id":"stub","permission_denials":[],"modelUsage":{"%s":{"inputTokens":10,"outputTokens":5}}}\n' "$MODEL"
else
  printf '{"type":"result","subtype":"success","is_error":false,"duration_ms":1000,"num_turns":3,"total_cost_usd":0.1,"session_id":"stub","permission_denials":[],"modelUsage":{"%s":{"inputTokens":10,"outputTokens":5,"costUSD":0.1}}}\n' "$MODEL"
fi
exit "$(cat "$dir/exit.$PHASE" 2>/dev/null || echo 0)"

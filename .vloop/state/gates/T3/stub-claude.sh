#!/bin/sh
# T3's stub claude. It answers doctor's two probes and plays a plan session
# that was denied Bash commands: it prints claude's result JSON with the
# permission_denials of denials.json (beside it), writes stderr.txt (beside it)
# to stderr, writes no plan and exits 0. Any other prompt exits 64.
dir=$(dirname "$0")
case "$1" in
  --version) echo "2.0.0 (Claude Code)"; exit 0;;
  plugin) echo '[{"id":"vloop@vloop","version":"@PLUGIN_VERSION@","enabled":true}]'; exit 0;;
esac
prompt=""
while [ $# -gt 0 ]; do
  case "$1" in
    -p) prompt="$2"; shift 2;;
    *) shift;;
  esac
done
case "$prompt" in
  /vloop:plan*) ;;
  *) echo "stub claude: unrecognised prompt: $prompt" >&2; exit 64;;
esac
cat "$dir/stderr.txt" >&2
printf '{"type":"result","subtype":"success","is_error":false,"duration_ms":1000,"num_turns":3,"total_cost_usd":0.1,"session_id":"stub","permission_denials":%s,"modelUsage":{}}\n' "$(cat "$dir/denials.json")"
exit 0

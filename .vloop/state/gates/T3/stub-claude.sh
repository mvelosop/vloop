#!/bin/sh
# The gates' stub claude. It never touches the network. It answers --version
# and `plugin list`, and for a -p prompt plays the session, appending to
# events beside the stub:
#
#   session <role> <arg|-> model=<model> effort=<effort|->
#   fence <role> ok|bad|none   whether --settings names a file whose JSON
#                              equals <role>-fence.json beside the stub
#   stale <role> yes|no        whether .vloop/tmp/stale.txt existed when the
#                              session started
#   lock <role> yes|no         whether .vloop/tmp/.running existed then
#
# A review-brief session (/vloop:review-brief <path>) copies report.json
# beside the stub, when there is one, to .vloop/tmp/brief-review.json. Files
# beside the stub change what it does first:
#   claude.gitconfig   it sets gate.planted in the repository's git config
#   claude.moveref     it makes a commit (HEAD moves)
#   claude.brief       it appends a line to the brief it reviews
#   claude.exit3       it exits 3 after its output
# A plan session (/vloop:plan <brief>) writes no plan and exits 3.
dir=$(dirname "$0")
case "${1:-}" in
  --version) echo "2.0.0 (Claude Code)"; exit 0;;
  plugin) echo '[{"id":"vloop@vloop","version":"@PLUGIN_VERSION@","enabled":true}]'; exit 0;;
esac
prompt="" MODEL="" EFFORT="-" SETTINGS=""
while [ $# -gt 0 ]; do
  case "$1" in
    -p) prompt="${2:-}"; shift 2;;
    --model) MODEL="${2:-}"; shift 2;;
    --effort) EFFORT="${2:-}"; shift 2;;
    --settings) SETTINGS="${2:-}"; shift 2;;
    *) shift;;
  esac
done
cmd="${prompt%% *}" ARG="${prompt#* }"
[ "$ARG" = "$prompt" ] && ARG=-
case "$cmd" in
  /vloop:review-brief) ROLE=review-brief;;
  /vloop:plan) ROLE=plan;;
  *) echo "stub claude: unrecognised prompt: $prompt" >&2; exit 64;;
esac
r=none
if [ -n "$SETTINGS" ]; then
  r=bad
  [ -f "$SETTINGS" ] && [ -f "$dir/$ROLE-fence.json" ] && [ "$(jq -S . "$SETTINGS" 2>/dev/null)" = "$(jq -S . "$dir/$ROLE-fence.json")" ] && r=ok
fi
st=no; [ -e .vloop/tmp/stale.txt ] && st=yes
lk=no; [ -f .vloop/tmp/.running ] && lk=yes
{
  echo "session $ROLE $ARG model=$MODEL effort=$EFFORT"
  echo "fence $ROLE $r"
  echo "stale $ROLE $st"
  echo "lock $ROLE $lk"
} >>"$dir/events"
xit=0
case "$ROLE" in
  review-brief)
    [ -f "$dir/claude.gitconfig" ] && git config gate.planted yes
    [ -f "$dir/claude.moveref" ] && git commit -q --allow-empty -m "moved by the stub review-brief session"
    [ -f "$dir/claude.brief" ] && echo "written by the stub review-brief session" >>"$ARG"
    mkdir -p .vloop/tmp
    [ -f "$dir/report.json" ] && cp "$dir/report.json" .vloop/tmp/brief-review.json
    [ -f "$dir/claude.exit3" ] && xit=3
    ;;
  plan) xit=3;;
esac
echo "stub $ROLE session" >&2
printf '{"type":"result","subtype":"success","is_error":false,"duration_ms":1000,"num_turns":3,"total_cost_usd":0.1,"session_id":"stub","permission_denials":[],"modelUsage":{"%s":{"inputTokens":10,"outputTokens":5,"costUSD":0.1}}}\n' "$MODEL"
exit "$xit"

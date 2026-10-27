#!/bin/sh
# The gates' stub claude. It never touches the network. It answers doctor's
# two probes and, for a -p prompt, plays the phase, appending one line per
# session to events:
#
#   session <role> <task|-> model=<model>
#
# role is plan, gate-review, work, review, second or verify; a /vloop:review
# session is "second" when its --model is the one in second.model, else
# "review". A review or second session also appends
#
#   fence <role> <task> ok|bad|none    whether --settings names a file whose
#                                      JSON equals the review fence
#
# and a second session, before it writes anything,
#
#   primary-visible <task> yes|no      "yes" when .vloop/tmp/verdict.json
#                                      exists or a file under .vloop/tmp/
#                                      holds the primary's marker
#
# A work session appends "notes <task> <the task's notes in the plan>".
#
# Verdicts. The primary review passes T1, fails T2 on its first review of T2
# (finding "primary finding on T2") and passes it after; its verdict's notes
# hold the marker PRIMARY-VERDICT-MARK. The second review fails T1 (finding
# "second finding on T1"), fails T2 on its first review of T2 ("second finding
# on T2") and passes it after — or passes everything when second.pass exists.
# Second-review faults, each when its file exists: second.noverdict (no
# verdict written), second.touch (appends to the tracked docs/x.md, then
# passes), second.moveref (creates refs/heads/gate-moved, then passes).
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
second=$(cat "$dir/second.model" 2>/dev/null)
# verdict <task> <PASS|FAIL> <notes> [finding]: write the verdict.
verdict() {
  if [ -n "${4:-}" ]; then f=$(printf '[{"summary":"%s","kind":"bug"}]' "$4"); else f='[]'; fi
  printf '{"schema":"verdict/v1","task":"%s","verdict":"%s","criteria":[],"findings":%s,"notes":"%s"}\n' "$1" "$2" "$f" "$3" >.vloop/tmp/verdict.json
}
# count <name>: how many times this was counted before, then count it.
count() { n=$(cat "$dir/count.$1" 2>/dev/null || echo 0); echo $((n + 1)) >"$dir/count.$1"; echo "$n"; }
fence() {
  r=none
  if [ -n "$SETTINGS" ]; then
    r=bad
    [ -f "$SETTINGS" ] && [ "$(jq -S . "$SETTINGS" 2>/dev/null)" = "$(jq -S . "$dir/review-fence.json")" ] && r=ok
  fi
  echo "fence $1 $ARG $r" >>"$dir/events"
}
case "$cmd" in
  /vloop:plan) ROLE=plan; mkdir -p .vloop/state && cp "$dir/plan.json" .vloop/state/state.json;;
  /vloop:gate-review) ROLE=gate-review
    echo '{"schema":"gate-verdict/v2","verdict":"PASS","tasks":[],"notes":"none"}' >.vloop/tmp/gate-verdict.json;;
  /vloop:work) ROLE=work; echo "$ARG $(date +%s)" >>"$ARG.out"
    echo "notes $ARG $(jq -r --arg id "$ARG" '.tasks[] | select(.id == $id) | .notes' .vloop/state/state.json)" >>"$dir/events"
    printf '{"schema":"proposal/v1","task":"%s","outcome":"done","summary":"made %s","files":[],"verified":"ok","notes":"none"}\n' "$ARG" "$ARG" >.vloop/tmp/proposal.json;;
  /vloop:review)
    if [ -n "$second" ] && [ "$MODEL" = "$second" ]; then
      ROLE=second
      echo "session $ROLE $ARG model=$MODEL" >>"$dir/events"; fence second
      vis=no
      { [ -e .vloop/tmp/verdict.json ] || grep -rqF PRIMARY-VERDICT-MARK .vloop/tmp 2>/dev/null; } && vis=yes
      echo "primary-visible $ARG $vis" >>"$dir/events"
      n=$(count "second.$ARG")
      [ -f "$dir/second.touch" ] && echo "changed by the second review" >>docs/x.md
      [ -f "$dir/second.moveref" ] && git update-ref refs/heads/gate-moved HEAD
      if [ -f "$dir/second.noverdict" ]; then :
      elif [ -f "$dir/second.pass" ] || [ -f "$dir/second.touch" ] || [ -f "$dir/second.moveref" ]; then verdict "$ARG" PASS "second"
      elif [ "$ARG" = T1 ]; then verdict T1 FAIL "second" "second finding on T1"
      elif [ "$n" = 0 ]; then verdict "$ARG" FAIL "second" "second finding on $ARG"
      else verdict "$ARG" PASS "second"; fi
    else
      ROLE=review
      echo "session $ROLE $ARG model=$MODEL" >>"$dir/events"; fence review
      n=$(count "primary.$ARG")
      if [ "$ARG" = T2 ] && [ "$n" = 0 ]; then verdict T2 FAIL "PRIMARY-VERDICT-MARK" "primary finding on T2"
      else verdict "$ARG" PASS "PRIMARY-VERDICT-MARK"; fi
    fi;;
  /vloop:verify) ROLE=verify
    echo '{"schema":"verify/v1","findings":[],"checked":{"example_lines":0,"sentences":0,"surfaces":0},"notes":"none"}' >.vloop/tmp/verify.json;;
  *) echo "stub claude: unrecognised prompt: $prompt" >&2; exit 64;;
esac
case "$ROLE" in review|second) ;; *) echo "session $ROLE $ARG model=$MODEL" >>"$dir/events";; esac
echo "stub $ROLE session" >&2
printf '{"type":"result","subtype":"success","is_error":false,"duration_ms":1000,"num_turns":3,"total_cost_usd":0.1,"session_id":"stub","permission_denials":[],"modelUsage":{"%s":{"inputTokens":10,"outputTokens":5,"costUSD":0.1}}}\n' "$MODEL"
exit 0

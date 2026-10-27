#!/bin/sh
# Gate T6 — S1, metrics/v4. The metrics snapshot a fixture run commits
# (.vloop/state/metrics/<brief>.json) is metrics/v4, valid against the
# embedded schema, with second-review among the agents per phase and a
# second_review object (mode, sessions, dissents, dissent_tasks), for the
# feedback run of the worked example (three second sessions, one dissent on
# T1), a run with the mode off, and a copilot second reviewer.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
cd "$top" || fail "repo root"
newbin
M=$t/repo/.vloop/state/metrics/$BNAME.json
# snap <label> <agents JSON> <second_review JSON>
snap() {
  [ -f "$M" ] || fail "$1: the run committed no metrics snapshot .vloop/state/metrics/$BNAME.json: $(tail -5 "$t/o")"
  [ "$("$JQ" -r .schema "$M")" = metrics/v4 ] || fail "$1: the metrics snapshot is $("$JQ" -r .schema "$M"), want metrics/v4"
  (cd "$t/repo" && venv "$S" schema validate metrics/v4 "$M" >"$t/val" 2>&1) || fail "$1: the snapshot is not valid metrics/v4: $(cat "$t/val")"
  "$JQ" -e --argjson w "$2" '.agents["second-review"] == $w' "$M" >/dev/null ||
    fail "$1: agents[\"second-review\"] is $("$JQ" -c '.agents["second-review"]' "$M"), want $2"
  "$JQ" -e --argjson w "$3" '.second_review | {mode, sessions, dissents, dissent_tasks} == $w' "$M" >/dev/null ||
    fail "$1: second_review is $("$JQ" -c .second_review "$M"), want $3"
}
CFG="run.second-review feedback
model.second-review $SECOND"; fixture
runfix "$t/o"
[ "$rc" = 0 ] || fail "feedback: vloop run exited $rc: $(tail -8 "$t/o")"
snap feedback '["claude"]' '{"mode":"feedback","sessions":3,"dissents":1,"dissent_tasks":["T1"]}'
"$JQ" -e '.agents.review == ["claude"]' "$M" >/dev/null || fail "feedback: agents.review is $("$JQ" -c .agents.review "$M")"

CFG=; fixture
runfix "$t/o"
[ "$rc" = 0 ] || fail "off: vloop run exited $rc"
snap off '[]' '{"mode":"off","sessions":0,"dissents":0,"dissent_tasks":[]}'

CFG="run.second-review on-pass
agent.second-review copilot
model.second-review $SECOND"; fixture plan1.json
runfix "$t/o"
[ "$rc" = 0 ] || fail "copilot second review: vloop run exited $rc: $(tail -8 "$t/o")"
snap copilot '["copilot"]' '{"mode":"on-pass","sessions":1,"dissents":0,"dissent_tasks":[]}'
echo "gate T6: ok"

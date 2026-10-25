#!/bin/sh
# Gate T3 — the record and C3. A copilot review session's session/v3 record
# from the stub's JSON Lines: agent copilot, the review's model, turns from the
# assistant.turn_start events, duration_ms from the result, is_error from its
# exitCode, credits from the last session.usage_checkpoint, cost_usd credits ×
# copilot.credit-usd (default, file and environment), cost_source estimated,
# models_used one zero-token entry for the model at that cost, no denials, and
# nothing of codeChanges (no absolute path). No checkpoint: credits and cost 0,
# cost_source none. No result event: SESSION RECORD MISSING and no record.
# Claude records carry no credits. The schemas accept these records and a
# base-written one. `vloop metrics` (JSON and text) and the committed
# snapshot: credits in total and per phase, the estimated cost counted and the
# total marked ≈; the cost ceiling counts a review's estimated cost. Every
# expected cost is computed here from the stub's credits and the price.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
newbin
basebin
S=$t/bin/new
mul() { awk -v a="$1" -v b="$2" 'BEGIN { printf "%.12g", a * b }'; }
# copilotrec <file> <credits> <cost> <what>: a copilot review record as C1 pins it.
copilotrec() {
  f=$1; c=$2; u=$3; w=$4
  "$JQ" -e --arg m "$MODEL" '.schema == "session/v3" and .agent == "copilot" and .phase == "review" and .model == $m
    and .turns == 3 and .duration_ms == 18000 and .is_error == false and .permission_denials == []' "$f" >/dev/null ||
    fail "$w: $(basename "$f") is not the record C1 pins: $("$JQ" -c '{schema, agent, phase, model, turns, duration_ms, is_error, permission_denials}' "$f")"
  [ "$("$JQ" -r '.credits | type' "$f")" = number ] || fail "$w: $(basename "$f") has no numeric credits"
  near "$("$JQ" .credits "$f")" "$c" || fail "$w: credits $("$JQ" .credits "$f"), want $c"
  near "$("$JQ" .cost_usd "$f")" "$u" || fail "$w: cost_usd $("$JQ" .cost_usd "$f"), want $u"
  "$JQ" -e --arg m "$MODEL" '(.models_used | keys) == [$m] and ([.models_used[$m] | to_entries[] | select(.key | endswith("tokens")) | .value] | length > 0 and all(. == 0))' "$f" >/dev/null ||
    fail "$w: models_used is not one zero-token entry for $MODEL: $("$JQ" -c .models_used "$f")"
  near "$("$JQ" --arg m "$MODEL" '.models_used[$m].cost_usd' "$f")" "$u" || fail "$w: models_used cost $("$JQ" -c .models_used "$f"), want $u"
  grep -qF "$t/" "$f" && fail "$w: $(basename "$f") holds an absolute path: $(grep -F "$t/" "$f")"
  grep -Eq 'codeChanges|filesModified|linesAdded' "$f" && fail "$w: $(basename "$f") records codeChanges"
  "$S" schema validate session/v3 "$f" >"$t/v.out" 2>&1 || fail "$w: vloop schema validate session/v3 $(basename "$f"): $(cat "$t/v.out")"
}
reviews() { ls "$1" | grep '^[0-9]*-review\.json$' | sed "s|^|$1/|"; }
COST=$(mul "$CREDITS" "$PRICE")

# The two-task run with copilot reviews, at the default price.
CFG=$COPILOT_CFG fixture; runfix "$t/cop" $CARRY
[ "$rc" = 0 ] || fail "the run with copilot reviews exited $rc: $(tail -8 "$t/cop")"
[ "$(reviews "$t/cop.sessions" | wc -l | tr -d ' ')" = 2 ] || fail "two copilot reviews left records '$(ls "$t/cop.sessions" | tr '\n' ' ')'"
for f in $(reviews "$t/cop.sessions"); do
  [ "$("$JQ" -r .cost_source "$f")" = estimated ] || fail "$(basename "$f"): cost_source $("$JQ" -c .cost_source "$f"), want estimated"
  copilotrec "$f" "$CREDITS" "$COST" "default price"
done
for f in "$t/cop.sessions"/*.json; do
  case "$(basename "$f")" in [0-9][0-9][0-9]-review.json) continue;; esac
  "$JQ" -e '.agent == "claude" and (has("credits") | not) and .cost_source == "reported"' "$f" >/dev/null ||
    fail "$(basename "$f"): a claude record with $("$JQ" -c '{agent, credits, cost_source}' "$f")"
done

# The schemas: what they accept and refuse.
r1=$(reviews "$t/cop.sessions" | head -1)
for e in '.cost_source = "bogus"' '.credits = "14.2"' 'del(.agent)'; do
  "$JQ" "$e" "$r1" >"$t/bad.json" || fail jq
  "$S" schema validate session/v3 "$t/bad.json" >/dev/null 2>&1 && fail "session/v3 accepts a copilot record with $e"
done
fixture; (cd "$t/repo" && venv $CARRY "$t/bin/base" run "$BRIEF" >"$t/base.out" 2>&1) || fail "the base's run exited non-zero: $(tail -5 "$t/base.out")"
for f in "$t/repo"/.vloop/state/runs/*/*/sessions/*.json; do
  "$S" schema validate session/v3 "$f" >"$t/v.out" 2>&1 || fail "session/v3 no longer validates the base-written $(basename "$f"): $(cat "$t/v.out")"
done

# Metrics over the copilot run: JSON, text and the committed snapshot.
CFG=$COPILOT_CFG fixture; runfix "$t/cop" $CARRY
[ "$rc" = 0 ] || fail "the second run with copilot reviews exited $rc"
(cd "$t/repo" && venv "$S" --json metrics "$BNAME" >"$t/m.json" 2>"$t/m.err") || fail "vloop --json metrics: $(cat "$t/m.err")"
(cd "$t/repo" && venv "$S" metrics "$BNAME" >"$t/m.txt" 2>&1) || fail "vloop metrics: $(cat "$t/m.txt")"
sumc=$(cat "$t/cop.sessions"/*.json | "$JQ" -s '[.[].credits // 0] | add')
sumu=$(cat "$t/cop.sessions"/*.json | "$JQ" -s '[.[].cost_usd] | add')
sumr=$(cat $(reviews "$t/cop.sessions") | "$JQ" -s '[.[].cost_usd] | add')
near "$sumc" "$(mul 2 "$CREDITS")" || fail "the records' credits add to $sumc"
# metricsdoc <file> <what> [snapshot]: metrics/v3 carrying the run's credits and
# estimated cost; for the snapshot, not the totals.
metricsdoc() {
  [ "$("$JQ" -r .schema "$1")" = metrics/v3 ] || fail "$2: schema $("$JQ" -c .schema "$1")"
  "$S" schema validate metrics/v3 "$1" >"$t/v.out" 2>&1 || fail "$2: vloop schema validate metrics/v3: $(cat "$t/v.out")"
  [ "$("$JQ" -r '.credits | type' "$1")" = object ] || fail "$2: no credits object: $("$JQ" -c .credits "$1")"
  near "$("$JQ" .credits.total "$1")" "$sumc" || fail "$2: credits.total $("$JQ" .credits.total "$1"), the records' $sumc"
  "$JQ" -e --argjson c "$sumc" '[.credits | to_entries[] | select(.key != "total") | .value] as $v
    | ($v | length) >= 4 and ($v | all(type == "number")) and ([$v[] | select(. != 0)] | length) == 1
    and ([$v[] | select(. != 0)][0] - $c | . < 1e-9 and . > -1e-9)' "$1" >/dev/null ||
    fail "$2: credits per phase is not the reviews' total and 0 elsewhere: $("$JQ" -c .credits "$1")"
  "$JQ" -e '.agents.review == ["copilot"] and .agents.work == ["claude"] and .agents.plan == ["claude"]' "$1" >/dev/null ||
    fail "$2: agents $("$JQ" -c .agents "$1")"
  near "$("$JQ" .cost.review_usd "$1")" "$sumr" || fail "$2: cost.review_usd $("$JQ" .cost.review_usd "$1"), the review records' $sumr"
  # The snapshot is committed before the verify session; its totals are not the run's.
  [ "${3:-}" = snapshot ] && return 0
  near "$("$JQ" .cost.total_usd "$1")" "$sumu" || fail "$2: cost.total_usd $("$JQ" .cost.total_usd "$1"), the records' $sumu"
  t1=$(cat "$t/cop.sessions"/*.json | "$JQ" -s '[.[] | select(.task == "T1") | .cost_usd] | add')
  near "$("$JQ" '.by_task[] | select(.id == "T1") | .cost_usd' "$1")" "$t1" || fail "$2: T1's cost $("$JQ" '.by_task[] | select(.id == "T1") | .cost_usd' "$1"), its records' $t1"
}
metricsdoc "$t/m.json" "vloop --json metrics"
[ -f "$t/cop.metrics" ] || fail "the run committed no metrics snapshot"
metricsdoc "$t/cop.metrics" "the committed metrics snapshot" snapshot
want=$(awk -v u="$sumu" 'BEGIN { printf "%.2f", u }')
grep -Eq "^ *cost .*≈\\\$$want( |$)" "$t/m.txt" || fail "vloop metrics: no cost line with ≈\$$want: $(grep -E '^ *cost ' "$t/m.txt")"

# All claude: no credits, no ≈.
fixture; runfix "$t/cl" $CARRY
[ "$rc" = 0 ] || fail "the all-claude run exited $rc"
(cd "$t/repo" && venv "$S" --json metrics "$BNAME" >"$t/c.json" 2>&1) || fail "vloop --json metrics (all claude): $(cat "$t/c.json")"
(cd "$t/repo" && venv "$S" metrics "$BNAME" >"$t/c.txt" 2>&1) || fail "vloop metrics (all claude)"
"$JQ" -e '(.credits | type) == "object" and ([.credits[]] | all(. == 0))' "$t/c.json" >/dev/null || fail "all claude: credits $("$JQ" -c .credits "$t/c.json"), want 0 everywhere"
grep -E '^ *cost ' "$t/c.txt" | grep -q '≈' && fail "all claude: the cost line is marked estimated: $(grep -E '^ *cost ' "$t/c.txt")"
grep -Eq '^ *cost +\$' "$t/c.txt" || fail "all claude: no cost line: $(cat "$t/c.txt")"

# The price from the file and from the environment.
CFG="$COPILOT_CFG
copilot.credit-usd 0.01" fixture plan1.json; runfix "$t/pf" $CARRY
[ "$rc" = 0 ] || fail "the run with copilot.credit-usd 0.01 exited $rc"
copilotrec "$(reviews "$t/pf.sessions")" "$CREDITS" "$(mul "$CREDITS" 0.01)" "copilot.credit-usd 0.01 in the file"
CFG=$COPILOT_CFG fixture plan1.json; runfix "$t/pe" $CARRY VLOOP_COPILOT_CREDIT_USD=0.02
[ "$rc" = 0 ] || fail "the run with VLOOP_COPILOT_CREDIT_USD=0.02 exited $rc"
copilotrec "$(reviews "$t/pe.sessions")" "$CREDITS" "$(mul "$CREDITS" 0.02)" "VLOOP_COPILOT_CREDIT_USD=0.02"

# No checkpoint: credits and cost 0, cost_source none.
CFG=$COPILOT_CFG fixture plan1.json; : >"$t/stub/copilot.nocheckpoint"; runfix "$t/nc" $CARRY
[ "$rc" = 0 ] || fail "the run whose copilot printed no checkpoint exited $rc"
f=$(reviews "$t/nc.sessions"); [ -n "$f" ] || fail "no checkpoint: no review record"
"$JQ" -e '.agent == "copilot" and .credits == 0 and .cost_usd == 0 and .cost_source == "none" and .turns == 3 and .duration_ms == 18000' "$f" >/dev/null ||
  fail "no checkpoint: $("$JQ" -c '{agent, credits, cost_usd, cost_source, turns, duration_ms}' "$f")"
"$S" schema validate session/v3 "$f" >"$t/v.out" 2>&1 || fail "no checkpoint: session/v3 rejects the record: $(cat "$t/v.out")"

# result.exitCode not 0: is_error.
CFG=$COPILOT_CFG fixture plan1.json; echo 1 >"$t/stub/copilot.exitcode"; runfix "$t/ex" $CARRY
f=$(reviews "$t/ex.sessions" | head -1); [ -n "$f" ] || fail "exitCode 1: no review record"
[ "$("$JQ" .is_error "$f")" = true ] || fail "exitCode 1 in the result: is_error $("$JQ" .is_error "$f")"

# No result event: no record, SESSION RECORD MISSING.
CFG=$COPILOT_CFG fixture plan1.json; : >"$t/stub/copilot.noresult"; runfix "$t/nr" $CARRY
grep -q 'SESSION RECORD MISSING review T1 (iteration 1)' "$t/nr.runlog" || fail "no result event: the run log has no 'SESSION RECORD MISSING review T1 (iteration 1)': $(grep MISSING "$t/nr.runlog")"
[ -z "$(reviews "$t/nr.sessions")" ] || fail "no result event, yet a review record: $(reviews "$t/nr.sessions")"

# The cost ceiling counts the estimated cost: at $1 a credit, T1's review
# passes a $5 ceiling the claude sessions alone stay far below.
CFG="$COPILOT_CFG
copilot.credit-usd 1" fixture; runfix "$t/ceil" $CARRY VLOOP_RUN_COST_CEILING=5
[ "$rc" = 6 ] || fail "a review estimated at \$$(mul "$CREDITS" 1) under a \$5 ceiling: the run exited $rc, want 6: $(tail -5 "$t/ceil")"
[ "$(sessions "$t/ceil.log" copilot | wc -l | tr -d ' ')" = 1 ] || fail "the run went on past the ceiling: copilot sessions $(sessions "$t/ceil.log" copilot | tr '\n' ' ')"
echo "gate T3: ok"

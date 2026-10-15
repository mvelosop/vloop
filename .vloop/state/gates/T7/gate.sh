#!/bin/sh
# Gate T7 — V3: where a brief's runs hold a verify session, vloop metrics adds
# it after gate review on the time, cost and models lines; the total includes
# it, agent does not. The instructions line adds verify after review when the
# tree holds plugin/skills/verify/SKILL.md, and leaves it out otherwise. --json
# gains time.verify_ms, cost.verify_usd, models.verify, effort.verify and
# instructions.verify, optional in metrics/v2. A brief with no verify session
# prints exactly what the plan's base prints.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
S=$t/stub
words() { n=$1; i=0; while [ $i -lt "$n" ]; do printf 'w '; i=$((i + 1)); done; echo; }
line() { sq "$1" | grep -E "^$2 " | head -n 1; }

# --- a brief with a verify session (cost 0.30, model haiku)
runrepo m T1
mkdir -p "$R/plugin/skills/review" "$R/plugin/skills/verify"
words 10 >"$R/plugin/skills/review/SKILL.md"; words 25 >"$R/plugin/skills/verify/SKILL.md"
# Not fatal here: on a tree without T3 and T4 the missing session is the failure to report; the model is asserted below.
inr "$B" config set model.verify haiku </dev/null >"$t/o" 2>&1
commit "skills and model"
vreport "$S/verify-report.json" 1
echo 0.30 >"$S/cost-verify"
vrun m.run
[ "$got" = 0 ] || fail "vloop run: exit $got: $(tail -n 15 "$t/m.run")"
grep -q -- '-p /vloop:verify' "$S/argv.log" || fail "the run held no verify session (T4, the driver's verify session, comes first)"
inr "$B" metrics "$BRIEF" </dev/null >"$t/sum" 2>"$t/sum.err" || fail "vloop metrics: $(cat "$t/sum.err")"
tl=$(line "$t/sum" time); cl=$(line "$t/sum" cost); ml=$(line "$t/sum" models); il=$(line "$t/sum" instructions)
echo "$tl" | grep -Eq ' · gate review [0-9]+\.[0-9] min · verify [0-9]+\.[0-9] min · gates ' || fail "the time line does not add ' · verify <m> min' after gate review: $tl"
echo "$cl" | grep -Eq '^cost \$0\.70 · plan 0\.10 · work 0\.10 · review 0\.10 · gate review 0\.10 · verify 0\.30 · ' ||
  fail "the cost line is not '\$0.70 · plan 0.10 · work 0.10 · review 0.10 · gate review 0.10 · verify 0.30 · …': $cl"
echo "$ml" | grep -Eq ' · gate review [^ ]+ · verify haiku$' || fail "the models line does not end ' · gate review <models> · verify haiku': $ml"
inr "$B" metrics "$BRIEF" --json </dev/null >"$t/j" 2>"$t/j.err" || fail "vloop metrics --json: $(cat "$t/j.err")"
"$JQ" -e '.schema == "metrics/v2"' "$t/j" >/dev/null || fail "metrics --json is no longer metrics/v2"
"$JQ" -e '(.time.verify_ms | type) == "number" and .time.verify_ms >= 0' "$t/j" >/dev/null || fail "--json has no time.verify_ms: $("$JQ" -c .time "$t/j")"
"$JQ" -e '((.cost.verify_usd - 0.3) | fabs) < 0.000001 and ((.cost.total_usd - 0.7) | fabs) < 0.000001' "$t/j" >/dev/null || fail "--json cost: $("$JQ" -c .cost "$t/j")"
"$JQ" -e '.time.agent_ms == .time.work_ms + .time.review_ms' "$t/j" >/dev/null || fail "agent is not work + review: $("$JQ" -c .time "$t/j")"
"$JQ" -e '.models.verify == ["haiku"]' "$t/j" >/dev/null || fail "--json models.verify is not [\"haiku\"]: $("$JQ" -c .models "$t/j")"
"$JQ" -e '.effort | has("verify")' "$t/j" >/dev/null || fail "--json has no effort.verify: $("$JQ" -c .effort "$t/j")"
"$JQ" -e '.instructions.verify - .instructions.review == 15' "$t/j" >/dev/null || fail "--json instructions.verify is not the verify skill's words plus the shared ones: $("$JQ" -c .instructions "$t/j")"
iv=$("$JQ" -r '.instructions.verify' "$t/j"); ir=$("$JQ" -r '.instructions.review' "$t/j")
echo "$il" | grep -Eq "review $ir( words)? · verify $iv( words)? · guidelines " || fail "the instructions line does not add ' · verify $iv' after review $ir: $il"
inr "$B" schema validate metrics/v2 "$t/j" </dev/null >"$t/o" 2>&1 || fail "metrics --json does not validate as metrics/v2: $(cat "$t/o")"
"$JQ" 'del(.time.verify_ms, .cost.verify_usd, .models.verify, .effort.verify, .instructions.verify)' "$t/j" >"$t/j0"
inr "$B" schema validate metrics/v2 "$t/j0" </dev/null >"$t/o" 2>&1 || fail "the verify properties are not optional in metrics/v2: $(cat "$t/o")"

# --- a brief with no verify session and no verify skill
resetstub
runrepo n T1
inr "$B" config set run.verify false </dev/null >"$t/o" 2>&1 || fail "config set run.verify false: $(cat "$t/o")"
commit off
vrun n.run
[ "$got" = 0 ] || fail "vloop run: exit $got: $(tail -n 15 "$t/n.run")"
inr "$B" metrics "$BRIEF" </dev/null >"$t/sum2" 2>&1 || fail "vloop metrics: $(cat "$t/sum2")"
for l in time cost models instructions; do
  line "$t/sum2" "$l" | grep -q verify && fail "a brief with no verify session: the $l line mentions verify: $(line "$t/sum2" "$l")"
done

# --- real data: the newest consumed briefs and the all-briefs table, as the base prints them
basebin
for b in $(grep -l '^status: consumed' docs/briefs/*.loop-brief.md | sort | tail -n 3); do
  "$t/vloop-base" metrics "$b" </dev/null >"$t/rb" 2>&1; eb=$?
  "$B" metrics "$b" </dev/null >"$t/rn" 2>&1; en=$?
  [ "$eb" = "$en" ] && cmp -s "$t/rb" "$t/rn" || fail "vloop metrics $b differs from the base's:
$(diff "$t/rb" "$t/rn")"
done
"$t/vloop-base" metrics </dev/null >"$t/ab" 2>&1; "$B" metrics </dev/null >"$t/an" 2>&1
cmp -s "$t/ab" "$t/an" || fail "vloop metrics (every brief) differs from the base's:
$(diff "$t/ab" "$t/an")"
echo "gate T7: ok"

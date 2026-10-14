#!/bin/sh
# Gate T8 — M1: the per-task table gains test:code (test+ / code+) and $/kloc
# (cost per 1,000 code+ lines), two decimals, between other+ and agent, both
# "-" where code+ is 0; --json by_task gains test_code_ratio and
# per_1000_code_lines_usd, null where code+ is 0, as optional properties of
# schemas/metrics.v2.json; the summary's cost line ends "$<c>/kloc"; the
# all-briefs headers t:c and $/1k become test:code and $/kloc, values unchanged.
# A stub run: T1 writes 200 code and 300 test lines, T2 no code.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
runrepo m T1 T2
inr "$B" config set metrics.stacks go </dev/null >/dev/null 2>&1 || fail "config set metrics.stacks go"
commit stacks
cat >"$t/stub/work-T1.sh" <<'W'
mkdir -p src; i=0; : >src/a.go; while [ $i -lt 200 ]; do echo "// $i" >>src/a.go; i=$((i+1)); done
i=0; : >src/a_test.go; while [ $i -lt 300 ]; do echo "// $i" >>src/a_test.go; i=$((i+1)); done
W
vrun m.run
rm -f "$t/stub/work-T1.sh"
[ "$got" = 0 ] || fail "vloop run: exit $got: $(tail -n 15 "$t/m.run")"

inr "$B" metrics "$BRIEF" --by task --json </dev/null >"$t/j" 2>&1 || fail "metrics --by task --json: $(cat "$t/j")"
row() { "$JQ" -c --arg id "$1" '.by_task[] | select(.id == $id)' "$t/j"; }
[ "$(row T1 | "$JQ" .churn.code)" = 200 ] && [ "$(row T1 | "$JQ" .churn.test)" = 300 ] ||
  fail "fixture: T1 did not churn 200 code and 300 test lines: $(row T1)"
[ "$(row T2 | "$JQ" .churn.code)" = 0 ] || fail "fixture: T2 churned code: $(row T2)"
cost=$(row T1 | "$JQ" .cost_usd)
row T1 | "$JQ" -e --argjson c "$cost" '(.test_code_ratio - 1.5 | fabs) < 1e-9 and (.per_1000_code_lines_usd - ($c * 1000 / 200) | fabs) < 1e-9' >/dev/null 2>&1 ||
  fail "by_task T1 (code+ 200, test+ 300, cost $cost): want test_code_ratio 1.5 and per_1000_code_lines_usd $cost*1000/200: $(row T1)"
row T2 | "$JQ" -e 'has("test_code_ratio") and has("per_1000_code_lines_usd") and .test_code_ratio == null and .per_1000_code_lines_usd == null' >/dev/null ||
  fail "by_task T2 (code+ 0): want test_code_ratio and per_1000_code_lines_usd null: $(row T2)"
inr "$B" schema validate metrics/v2 "$t/j" </dev/null >"$t/v" 2>&1 || fail "the --by task --json output does not validate against metrics/v2: $(cat "$t/v")"
S=$top/schemas/metrics.v2.json
"$JQ" -e '.properties.by_task.items as $i | ($i | if has("$ref") then error("ref") else . end) |
  (.properties | has("test_code_ratio") and has("per_1000_code_lines_usd")) and
  ((.required // []) | index("test_code_ratio") == null and index("per_1000_code_lines_usd") == null)' "$S" >/dev/null 2>&1 ||
  fail "schemas/metrics.v2.json: by_task items do not carry test_code_ratio and per_1000_code_lines_usd as optional properties"

inr "$B" metrics "$BRIEF" --by task </dev/null >"$t/tbl" 2>&1 || fail "metrics --by task: $(cat "$t/tbl")"
h=$(head -n 1 "$t/tbl")
col() { printf '%s\n' "$h" | tr -s ' ' '\n' | grep -nFx -- "$1" | cut -d: -f1; }
o=$(col other+) c1=$(col test:code) c2=$(col '$/kloc') a=$(col agent)
[ -n "$o" ] && [ -n "$c1" ] && [ -n "$c2" ] && [ -n "$a" ] && [ "$c1" = $((o + 1)) ] && [ "$c2" = $((o + 2)) ] && [ "$a" = $((o + 3)) ] ||
  fail "the per-task header is not '… other+  test:code  \$/kloc  agent …': $h"
cell() { awk -v id="$1" -v n="$2" '$1 == id { print $n }' "$t/tbl"; }
want=$(awk -v c="$cost" 'BEGIN { printf "%.2f", c * 1000 / 200 }')
[ "$(cell T1 "$c1")" = 1.50 ] && [ "$(cell T1 "$c2")" = "$want" ] || fail "T1's row gives test:code '$(cell T1 "$c1")' and \$/kloc '$(cell T1 "$c2")', want 1.50 and $want:
$(cat "$t/tbl")"
[ "$(cell T2 "$c1")" = - ] && [ "$(cell T2 "$c2")" = - ] || fail "T2's row (code+ 0) gives '$(cell T2 "$c1")' '$(cell T2 "$c2")', want - and -:
$(cat "$t/tbl")"

inr "$B" metrics "$BRIEF" --json </dev/null >"$t/sj" 2>&1 || fail "metrics --json: $(cat "$t/sj")"
k=$("$JQ" -r '.cost.per_1000_code_lines_usd' "$t/sj" | awk '{ printf "%.2f", $1 }')
inr "$B" metrics "$BRIEF" </dev/null >"$t/sum" 2>&1 || fail "metrics: $(cat "$t/sum")"
cl=$(grep -E '^ *cost ' "$t/sum") || fail "no cost line: $(cat "$t/sum")"
case "$cl" in *" · \$$k/kloc") ;; *) fail "the cost line does not end ' · \$$k/kloc': '$cl'" ;; esac
grep -Fq 'per 1,000 code lines' "$t/sum" && fail "the summary still says 'per 1,000 code lines': $cl"

inr "$B" metrics </dev/null >"$t/all" 2>&1 || fail "metrics (all briefs): $(cat "$t/all")"
ah=$(head -n 1 "$t/all")
printf '%s\n' "$ah" | tr -s ' ' '\n' >"$t/ah"
grep -Fxq 'test:code' "$t/ah" && grep -Fxq '$/kloc' "$t/ah" || fail "the all-briefs header has no test:code and \$/kloc: $ah"
grep -Fxq 't:c' "$t/ah" && fail "the all-briefs header still has t:c: $ah"
grep -Fxq '$/1k' "$t/ah" && fail "the all-briefs header still has \$/1k: $ah"
ac() { grep -nFx -- "$1" "$t/ah" | cut -d: -f1; }
r=$(awk -v id="$RUN" '$1 == id' "$t/all")
tc=$("$JQ" -r '.size.test_code_ratio' "$t/sj" | awk '{ printf "%.2f", $1 }')
[ "$(printf '%s\n' "$r" | awk -v n="$(ac test:code)" '{ print $n }')" = "$tc" ] &&
  [ "$(printf '%s\n' "$r" | awk -v n="$(ac '$/kloc')" '{ print $n }')" = "$k" ] ||
  fail "the all-briefs row's test:code and \$/kloc are not the brief's $tc and $k: $ah
$r"
echo "gate T8: ok"

#!/bin/sh
# Gate T10 — E2: a recorded defect is escaped when its created time is after
# the committer time of its brief's release commit (M-5), whatever its
# found-by; otherwise it counts by found-by (gate, review in-loop; operator
# operator; user escaped). The same second is not escaped. defect list
# --matrix is unchanged, and no record changes. Real data: on this repository
# every consumed brief's metrics are the base's but for operator and escaped,
# which only trade (in-loop unchanged), and the total escaped rises.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
N=$RUN.loop-brief
# dline <want>: the summary's defects line of the brief, blanks squeezed.
dline() {
  inr "$B" metrics "$BRIEF" </dev/null >"$t/m" 2>&1 || fail "vloop metrics: $(cat "$t/m")"
  got=$(grep -E '^ *defects ' "$t/m" | sed -e 's/[[:space:]][[:space:]]*/ /g' -e 's/^ //' -e 's/ $//')
  [ "$got" = "defects $1" ] || fail "$2: the defects line is '$got', want 'defects $1'"
}
runrepo e T1
vrun e.run
[ "$got" = 0 ] || fail "vloop run: exit $got: $(tail -n 15 "$t/e.run")"
inr "$B" brief close "$BRIEF" --no-findings </dev/null >"$t/close" 2>&1 || fail "brief close: $(cat "$t/close")"
mkdir -p "$R/.vloop/defects" || fail mkdir
# Not merged yet: an operator defect created long after counts as operator.
defectdoc D20990101-0000-late "$N" operator 2099-01-01T00:00:00Z >"$R/.vloop/defects/D20990101-0000-late.md" || fail fixture
dline "in-loop 0 · operator 1 · escaped 0 · removal efficiency 100%" "an unmerged brief's operator defect"
rm -f "$R/.vloop/defects/D20990101-0000-late.md"
mergemain
sleep 2
for i in 1 2; do
  inr "$B" defect add "x" --brief "$N" --found-by operator </dev/null >"$t/add" 2>&1 || fail "defect add after the merge: $(cat "$t/add")"
done
dline "in-loop 0 · operator 0 · escaped 2 · removal efficiency 0%" "two operator defects recorded after the merge"

# One each side of the release, and on it.
fx() { defectdoc "$1" "$N" "$2" "$(utc "$3")" >"$R/.vloop/defects/$1.md" || fail "fixture $1"; }
fx D20250101-0001-before-op operator $((MERGED - 3600))
fx D20250101-0002-same-second operator "$MERGED"
fx D20250101-0003-after-op operator $((MERGED + 1))
fx D20250101-0004-after-review review $((MERGED + 60))
fx D20250101-0005-before-gate gate $((MERGED - 60))
fx D20250101-0006-before-user user $((MERGED - 60))
(cd "$R/.vloop/defects" && cat ./*.md) | cksum >"$t/sum.before"
dline "in-loop 1 · operator 2 · escaped 5 · removal efficiency 38%" "the time rule"
inr "$B" metrics "$BRIEF" --json </dev/null >"$t/j" 2>&1 || fail "metrics --json: $(cat "$t/j")"
"$JQ" -e '.defects.in_loop == 1 and .defects.operator == 2 and .defects.escaped == 5' "$t/j" >/dev/null ||
  fail "metrics --json defects: $("$JQ" -c .defects "$t/j"), want in_loop 1, operator 2, escaped 5"
inr "$B" defect list --matrix --brief "$N" </dev/null >"$t/x" 2>&1 || fail "defect list --matrix: $(cat "$t/x")"
w=$(grep -E '^work ' "$t/x" | sed -e 's/[[:space:]][[:space:]]*/ /g' -e 's/ $//')
[ "$w" = "work 1 1 5 1" ] || fail "defect list --matrix is not origin by catcher: the work row is '$w', want 'work 1 1 5 1':
$(cat "$t/x")"
(cd "$R/.vloop/defects" && cat ./*.md) | cksum | cmp -s - "$t/sum.before" || fail "a defect record changed"

# Real data, on this repository: base against now, consumed briefs only.
basebin
(cd "$top" && "$t/vloop-base" metrics --json) </dev/null >"$t/base.json" 2>"$t/base.err" || fail "the base's metrics --json on this repository: $(cat "$t/base.err")"
(cd "$top" && "$B" metrics --json) </dev/null >"$t/now.json" 2>"$t/now.err" || fail "metrics --json on this repository: $(cat "$t/now.err")"
"$JQ" -rn --slurpfile a "$t/base.json" --slurpfile b "$t/now.json" '
  def strip: del(.defects) | del(.by_task[]?.test_code_ratio, .by_task[]?.per_1000_code_lines_usd);
  [$a[0][] | select(.status == "consumed")] as $base |
  ($b[0] | map({key: .brief, value: .}) | from_entries) as $now |
  [ $base[] | . as $o | $now[$o.brief] as $n |
    if $n == null then "\($o.brief): missing"
    elif ($o | strip) != ($n | strip) then "\($o.brief): a figure other than operator and escaped changed"
    elif $o.defects.in_loop != $n.defects.in_loop then "\($o.brief): in-loop changed"
    elif ($o.defects.operator + $o.defects.escaped) != ($n.defects.operator + $n.defects.escaped) then "\($o.brief): operator + escaped changed"
    elif $n.defects.escaped < $o.defects.escaped then "\($o.brief): escaped fell"
    else empty end ] +
  (if ([$base[].defects.escaped] | add) < ([$base[] | $now[.brief].defects.escaped] | add) then []
   else ["the total escaped across consumed briefs did not rise from \([$base[].defects.escaped] | add)"] end)
  | .[]' >"$t/diff" || fail "comparing the real data"
[ ! -s "$t/diff" ] || fail "real data, base against now:
$(cat "$t/diff")"
echo "gate T10: ok"

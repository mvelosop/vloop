#!/bin/sh
# Gate T12 — E4: vloop defect set refuses rule and rule-outcome as any unknown
# field (exit 2), its message listing the fields without them, and leaves the
# record byte-identical; a record that has both stays valid and is listed with
# them; schemas/defect.v1.json keeps both properties. vloop brief close, and
# its --dry-run, print no `defects of <run id>` report and no untriaged hint;
# `squash-merge with the trailer:` stays close's last line.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
N=$RUN.loop-brief
ID=D20250101-0001-triaged
runrepo r T1
mkdir -p "$R/.vloop/defects" || fail mkdir
defectdoc "$ID" "$N" operator 2025-01-01T00:00:00Z "rule: \"gate: x\"
rule-outcome: missed" >"$R/.vloop/defects/$ID.md" || fail fixture
commit "a triaged defect"
cp "$R/.vloop/defects/$ID.md" "$t/before"
for fld in rule:'gate: y' rule-outcome:fired; do
  k=${fld%%:*} v=${fld#*:}
  inr "$B" defect set "$ID" "$k" "$v" </dev/null >"$t/o" 2>"$t/e"; c=$?
  [ "$c" = 2 ] || fail "defect set $ID $k '$v': exit $c, want 2: $(cat "$t/e")"
  [ "$(cat "$t/e")" = "vloop: cannot set \"$k\": want one of status, fixed-by, case, severity, origin, kind, task" ] ||
    fail "defect set $ID $k printed: $(cat "$t/e")"
  cmp -s "$R/.vloop/defects/$ID.md" "$t/before" || fail "a refused defect set changed $ID"
done
inr "$B" defect list --json </dev/null >"$t/l" 2>&1 || fail "defect list --json with a triaged record: $(cat "$t/l")"
"$JQ" -e --arg id "$ID" '[.[] | select(.id == $id and .rule == "gate: x" and .["rule-outcome"] == "missed")] | length == 1' "$t/l" >/dev/null 2>&1 ||
  fail "defect list --json does not show $ID's rule and rule-outcome as they are: $(cat "$t/l")"
inr "$B" defect list </dev/null >"$t/l" 2>&1 || fail "defect list: $(cat "$t/l")"
grep -Fq "$ID" "$t/l" || fail "defect list does not list $ID"
"$JQ" -e '.properties.rule != null and .properties["rule-outcome"] != null' "$top/schemas/defect.v1.json" >/dev/null 2>&1 ||
  fail "schemas/defect.v1.json no longer has rule and rule-outcome"

vrun r.run
[ "$got" = 0 ] || fail "vloop run: exit $got: $(tail -n 15 "$t/r.run")"
noreport() {
  grep -q '^defects of ' "$t/$1" && fail "$2 still prints the defect report:
$(cat "$t/$1")"
  grep -q 'untriaged' "$t/$1" && fail "$2 still prints the untriaged hint:
$(cat "$t/$1")"
  return 0
}
inr "$B" brief close "$BRIEF" --dry-run --no-findings </dev/null >"$t/dry" 2>&1 || fail "close --dry-run: $(cat "$t/dry")"
noreport dry "close --dry-run"
inr "$B" brief close "$BRIEF" --no-findings </dev/null >"$t/close" 2>&1 || fail "close: $(cat "$t/close")"
noreport close "close"
case "$(tail -n 1 "$t/close")" in "squash-merge with the trailer: Vloop-Brief: $N") ;; *) fail "close's last line: $(tail -n 1 "$t/close")" ;; esac
echo "gate T12: ok"

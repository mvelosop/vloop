#!/bin/sh
# Gate T7 — S1: a new recorded defect is D<YYYYMMDD-HHMMSS>-<slug> and a new
# intervention I<YYYYMMDD-HHMMSS>-<slug>, in local time; when a record of the
# kind already uses the stamp, whatever its slug, the stamp moves on one second
# until it is free, and no -2 suffix is made. Records exist for eleven seconds
# from now, so the next free stamp is known. Commands that name an ID accept
# the new form and the old one: defect set, intervention set, intervention
# show's links and doctor's guidelines provenance.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
M=B20260101-0800-merged
newrepo s
putbrief "$M" consumed
mkdir -p "$R/.vloop/defects" "$R/.vloop/interventions" || fail mkdir
defectdoc D20250101-0900-old "$M.loop-brief" operator 2025-01-01T09:00:00Z >"$R/.vloop/defects/D20250101-0900-old.md" || fail fixture
commit "a merged brief"
# idoc <id>: an intervention/v2 record.
idoc() {
  printf -- '---\nid: %s\nbrief: %s\nphase: run\nkind: repair\nautomatable: no\nby: operator\nschema: intervention/v2\noptions: 0\nrecommended: 0\ndecided: ""\nagreement: no-options\noccurred: 2026-01-01\nrecorded: 2026-01-01T09:00:00Z\n---\nfixture intervention %s\n' "$1" "$M.loop-brief" "$1"
}
now=$(date +%s)
k=0
while [ $k -le 10 ]; do
  s=$(stamp $((now + k)))
  defectdoc "D$s-other" "$M.loop-brief" operator 2026-01-01T09:00:00Z >"$R/.vloop/defects/D$s-other.md" || fail fixture
  idoc "I$s-other" >"$R/.vloop/interventions/I$s-other.md" || fail fixture
  k=$((k + 1))
done
want1=$(stamp $((now + 11))) want2=$(stamp $((now + 12)))

# add <kind> <out> <args...>: the path the add printed, in $p.
dadd() { inr "$B" defect add "$@" --brief "$M.loop-brief" --found-by user </dev/null >"$t/add" 2>&1 || fail "vloop defect add $*: $(cat "$t/add")"; p=$(tail -n 1 "$t/add"); }
iadd() { inr "$B" intervention add "$@" --brief "$M.loop-brief" --phase run --kind repair --automatable no --by operator </dev/null >"$t/add" 2>&1 || fail "vloop intervention add $*: $(cat "$t/add")"; p=$(tail -n 1 "$t/add"); }
# isrec <path> <dir> <id>: the add printed <dir>/<id>.md, which holds id: <id>.
isrec() {
  [ "$1" = "$2/$3.md" ] || fail "the add printed '$1', want $2/$3.md (the next free stamp with seconds)"
  [ -f "$R/$1" ] || fail "$1 does not exist"
  grep -Fxq "id: $3" "$R/$1" || fail "$1 does not hold id: $3"
}
dadd x; isrec "$p" .vloop/defects "D$want1-x"; d1=D$want1-x
dadd x; isrec "$p" .vloop/defects "D$want2-x"
iadd y; isrec "$p" .vloop/interventions "I$want1-y"; i1=I$want1-y
iadd y; isrec "$p" .vloop/interventions "I$want2-y"
ls "$R/.vloop/defects" "$R/.vloop/interventions" | grep -E -- '-[xy]-[0-9]+\.md$' && fail "a -<n> suffix was made"

inr "$B" defect set "$d1" status fixed </dev/null >"$t/o" 2>&1 || fail "vloop defect set $d1: $(cat "$t/o")"
grep -Fxq 'status: fixed' "$R/.vloop/defects/$d1.md" || fail "defect set did not set $d1's status"
inr "$B" defect set D20250101-0900-old status fixed </dev/null >"$t/o" 2>&1 || fail "vloop defect set on an old ID: $(cat "$t/o")"
inr "$B" intervention set "$i1" phase verify </dev/null >"$t/o" 2>&1 || fail "vloop intervention set $i1: $(cat "$t/o")"
grep -Fxq 'phase: verify' "$R/.vloop/interventions/$i1.md" || fail "intervention set did not set $i1's phase"

iadd "links" --context "see $d1, D20250101-0900-old and $i1"
ilink=$(basename "$p" .md)
inr "$B" intervention show "$ilink" --json </dev/null >"$t/show.json" 2>&1 || fail "vloop intervention show $ilink: $(cat "$t/show.json")"
for ref in "$d1:defect" "D20250101-0900-old:defect" "$i1:intervention"; do
  id=${ref%%:*} kind=${ref#*:}
  "$JQ" -e --arg r "$id" --arg k "$kind" '[.links[] | select(.ref == $r and .kind == $k and .title != "" and .title != null)] | length == 1' "$t/show.json" >/dev/null ||
    fail "intervention show does not resolve $id as a $kind: $(cat "$t/show.json")"
done

printf '# Guidelines\n\n- A new rule. (%s)\n- An old rule. (D20250101-0900-old)\n- A third. (%s)\n' "$d1" "$i1" >"$R/.vloop/repo-guidelines.md" || fail guidelines
inr "$B" doctor </dev/null >"$t/doctor" 2>&1
grep -i guidelines "$t/doctor" | grep -Eq 'cite no|provenance' && fail "doctor's guidelines check refuses an ID:
$(grep -i guidelines "$t/doctor")"
echo "gate T7: ok"

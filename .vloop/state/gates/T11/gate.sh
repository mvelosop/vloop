#!/bin/sh
# Gate T11 — E3: intervention add takes --origin brief|plan|work|env, required
# with --kind verification-finding and refused with any other kind (exit 2);
# the record carries origin: after by:, intervention set sets it, and
# schemas/intervention.v2.json has it as an optional property. brief close
# --finding "<origin>: <summary>" records one verification-finding
# intervention per finding (phase verify, automatable partly, by operator) and
# prints `recorded .vloop/interventions/<id>.md`; a finding without an origin
# is refused before anything is written (exit 2). The operator count is the
# operator defects not escaped plus the verification-finding interventions
# that carry an origin, with no overlap; the run record lists the
# interventions after the defects; --dry-run counts and lists them the same.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
N=$RUN.loop-brief
M=B20260101-0800-merged.loop-brief
squeezed() { sed -e 's/[[:space:]][[:space:]]*/ /g' -e 's/^ //' -e 's/ $//' "$1"; }
dl() { squeezed "$1" | grep -E '^defects ' | head -n 1; }
clean() { [ -z "$(git -C "$R" status --porcelain)" ] || fail "$1 wrote: $(git -C "$R" status --porcelain)"; }
# iadd <want exit> <args...>: vloop intervention add on $M in $R.
iadd() {
  w=$1; shift
  inr "$B" intervention add "y" --brief "$M" --phase verify --automatable no --by operator "$@" </dev/null >"$t/out" 2>"$t/err"; c=$?
  [ "$c" = "$w" ] || fail "vloop intervention add --brief $M $*: exit $c, want $w: $(cat "$t/err")"
}
idoc() { # idoc <id> <brief> <kind> [<origin line>]
  printf -- '---\nid: %s\nbrief: %s\nphase: verify\nkind: %s\nautomatable: no\nby: operator\n%sschema: intervention/v2\noptions: 0\nrecommended: 0\ndecided: ""\nagreement: no-options\noccurred: 2026-01-01\nrecorded: 2026-01-01T09:00:00Z\n---\nfixture %s\n' \
    "$1" "$2" "$3" "${4:+$4
}" "$1"
}

runrepo c T1
putbrief "${M%.loop-brief}" consumed
commit "a merged brief"
touch "$t/stub/fail-T1"
vrun c.run
[ "$got" = 0 ] || fail "vloop run: exit $got: $(tail -n 15 "$t/c.run")"

iadd 2 --kind verification-finding
[ "$(cat "$t/err")" = "vloop: --kind verification-finding needs --origin brief, plan, work or env" ] || fail "verification-finding without --origin printed: $(cat "$t/err")"
iadd 2 --kind repair --origin work
[ "$(cat "$t/err")" = "vloop: --origin is only for --kind verification-finding" ] || fail "--origin with --kind repair printed: $(cat "$t/err")"
[ -z "$(ls -A "$R/.vloop/interventions" 2>/dev/null)" ] || fail "a refused intervention add wrote a record"
iadd 0 --kind verification-finding --origin plan
p=$(tail -n 1 "$t/out"); f=$R/$p; id=$(basename "$p" .md)
[ -f "$f" ] || fail "intervention add printed '$p', not a record"
awk '$0 == "---" { n++; next } n == 1' "$f" >"$t/fm"
[ "$(grep -A1 -Fx 'by: operator' "$t/fm" | tail -n 1)" = "origin: plan" ] || fail "the record does not carry origin: plan right after by::
$(cat "$t/fm")"
inr "$B" intervention show "$id" --json </dev/null >"$t/show" 2>&1 || fail "intervention show: $(cat "$t/show")"
[ "$("$JQ" -r .origin "$t/show")" = plan ] || fail "intervention show --json has no origin plan: $(cat "$t/show")"
inr "$B" intervention set "$id" origin work </dev/null >"$t/o" 2>&1 || fail "intervention set $id origin work: $(cat "$t/o")"
grep -Fxq 'origin: work' "$f" || fail "intervention set did not set origin"
"$JQ" -e '.properties.origin != null and ((.required // []) | index("origin") == null)' "$top/schemas/intervention.v2.json" >/dev/null 2>&1 ||
  fail "schemas/intervention.v2.json does not have origin as an optional property"
"$JQ" '.origin = "work"' "$t/show" >"$t/show2" && inr "$B" schema validate intervention/v2 "$t/show2" </dev/null >"$t/o" 2>&1 ||
  fail "a record with an origin does not validate against intervention/v2: $(cat "$t/o")"

# Interventions on the run's brief that the operator count must not take: a
# verification-finding with no origin, a repair with one.
idoc I20250101-0001-old-finding "$N" verification-finding >"$R/.vloop/interventions/I20250101-0001-old-finding.md"
idoc I20250101-0002-repair-origin "$N" repair "origin: work" >"$R/.vloop/interventions/I20250101-0002-repair-origin.md"
now=$(date +%s); k=0
while [ $k -le 10 ]; do s=$(stamp $((now + k))); idoc "I$s-other" "$M" repair >"$R/.vloop/interventions/I$s-other.md"; k=$((k + 1)); done
commit "fixtures"
want=I$(stamp $((now + 11)))-z

for bad in "spec gap: z" "Brief: z" "z"; do
  inr "$B" brief close "$BRIEF" --finding "$bad" </dev/null >"$t/o" 2>"$t/e"; c=$?
  [ "$c" = 2 ] || fail "close --finding '$bad': exit $c, want 2: $(cat "$t/e")"
  [ "$(cat "$t/e")" = 'vloop: a finding starts with whose defect it is: --finding "<brief|plan|work|env>: <summary>"' ] ||
    fail "close --finding '$bad' printed: $(cat "$t/e")"
  clean "close --finding '$bad'"
done
inr "$B" brief close "$BRIEF" --dry-run --finding "brief: z" </dev/null >"$t/dry" 2>&1 || fail "close --dry-run: $(cat "$t/dry")"
clean "close --dry-run"
[ "$(dl "$t/dry")" = "defects in-loop 1 · operator 1 · escaped 0 · removal efficiency 100%" ] || fail "close --dry-run's defects line is '$(dl "$t/dry")'"
grep -Fxq "would record .vloop/interventions/$want.md" "$t/dry" || fail "close --dry-run does not print 'would record .vloop/interventions/$want.md':
$(cat "$t/dry")"

inr "$B" brief close "$BRIEF" --finding "brief: z" </dev/null >"$t/close" 2>&1 || fail "close: $(cat "$t/close")"
grep -Fxq "recorded .vloop/interventions/$want.md" "$t/close" || fail "close does not print 'recorded .vloop/interventions/$want.md':
$(cat "$t/close")"
grep -q '^recorded \.vloop/defects/' "$t/close" && fail "close still records a defect: $(cat "$t/close")"
[ -z "$(ls -A "$R/.vloop/defects" 2>/dev/null)" ] || fail "close wrote a defect file: $(ls "$R/.vloop/defects")"
case "$(tail -n 1 "$t/close")" in "squash-merge with the trailer: "*) ;; *) fail "close's last line: $(tail -n 1 "$t/close")" ;; esac
[ "$(dl "$t/close")" = "defects in-loop 1 · operator 1 · escaped 0 · removal efficiency 100%" ] || fail "close's defects line is '$(dl "$t/close")'"
f=$R/.vloop/interventions/$want.md
[ -f "$f" ] || fail "no $want.md"
awk '$0 == "---" { n++; next } n == 1' "$f" >"$t/fm"
for l in "id: $want" "brief: $N" "phase: verify" "kind: verification-finding" "automatable: partly" "by: operator" "origin: brief"; do
  grep -Fxq "$l" "$t/fm" || fail "$want.md has no '$l':
$(cat "$f")"
done
[ "$(awk '$0 == "---" { n++; next } n == 2 && NF { print; exit }' "$f")" = z ] || fail "$want.md's summary is not z:
$(cat "$f")"
git -C "$R" show --name-only --format= HEAD | grep -Fxq ".vloop/interventions/$want.md" || fail "the close commit does not carry $want.md"
rl=$(grep -n -F -- "- $want — verification-finding, brief, found by operator: z" "$R/$BRIEF" | cut -d: -f1)
[ -n "$rl" ] || fail "the run record has no '- $want — verification-finding, brief, found by operator: z':
$(sed -n '/## Run record/,$p' "$R/$BRIEF")"
gl=$(grep -n -F -- "- $RUN/i1-gate — " "$R/$BRIEF" | cut -d: -f1)
[ -n "$gl" ] && [ "$gl" -lt "$rl" ] || fail "the run record does not list the intervention after the defects:
$(sed -n '/## Run record/,$p' "$R/$BRIEF")"

# Both sources, no overlap: an operator defect of the unmerged brief adds one.
mkdir -p "$R/.vloop/defects"
defectdoc D20250101-0003-op "$N" operator 2025-01-01T00:00:00Z >"$R/.vloop/defects/D20250101-0003-op.md"
inr "$B" metrics "$BRIEF" </dev/null >"$t/m" 2>&1 || fail "metrics: $(cat "$t/m")"
[ "$(dl "$t/m")" = "defects in-loop 1 · operator 2 · escaped 0 · removal efficiency 100%" ] || fail "an operator defect and a finding: the defects line is '$(dl "$t/m")', want operator 2"
echo "gate T11: ok"

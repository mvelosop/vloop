#!/bin/sh
# Gate T1 — P1, `vloop metrics --by session`: every P1 line of the brief's
# worked example, on fixture repositories built here and on this repository's
# real B36 records, through the binary built from this tree. Rows are compared
# field by field (whitespace-split); JSON is parsed, never substring-matched.
G=.vloop/state/gates/T1
. "$G/lib.sh"

# ---- fixtures ----------------------------------------------------------
ID=B20260101-0000-fx1
ID2=B20260103-0000-fx2
newrepo r
cp -R "$G/repo/." "$R/" || fail "copy fixture repository"
commitall "$R"

# run NAME ARGS: vl ARGS, stdout in $t/out, stderr in $t/err, exit in rc.
run() { n=$1; shift; vl "$@" >"$t/out" 2>"$t/err"; rc=$?; }
ok() { [ "$rc" = 0 ] || fail "$n: exit $rc, want 0: $(cat "$t/err")"; }
usage() {
  [ "$rc" = 2 ] || fail "$n: exit $rc, want 2 (stderr: $(cat "$t/err"))"
  grep -Fq -- "$1" "$t/err" || fail "$n: stderr lacks '$1': $(cat "$t/err")"
  [ ! -s "$t/out" ] || fail "$n: printed on stdout: $(cat "$t/out")"
}

# ---- text: one brief ---------------------------------------------------
run "metrics fx1 --by session" -C "$R" metrics "$ID" --by session; ok
norm "$t/out" >"$t/got"
cmp -s "$t/got" "$G/fx1.rows" || fail "$n: the rows (whitespace-split) differ from the expected ones.
got:
$(cat "$t/out")
want (fields):
$(cat "$G/fx1.rows")"

# ---- text: two briefs, a blank line between ---------------------------
run "metrics fx1 fx2 --by session" -C "$R" metrics "$ID" "$ID2" --by session; ok
norm "$t/out" >"$t/got"
{ cat "$G/fx1.rows"; echo; cat "$G/fx2.rows"; } >"$t/want"
cmp -s "$t/got" "$t/want" || fail "$n: two briefs are not two tables separated by one blank line:
$(cat "$t/out")"

# ---- no brief named: the brief table, as --by task -----------------------
run "metrics (no brief)" -C "$R" metrics; ok
cp "$t/out" "$t/table"
run "metrics --by session (no brief)" -C "$R" metrics --by session; ok
cmp -s "$t/out" "$t/table" || fail "$n: does not print the brief table:
$(cat "$t/out")"

# ---- JSON ----------------------------------------------------------------
run "metrics fx1 --json" -C "$R" --json metrics "$ID"; ok
cp "$t/out" "$t/fx1.json"
run "metrics fx1 --by session --json" -C "$R" --json metrics "$ID" --by session; ok
cp "$t/out" "$t/fx1s.json"
py "$G/check_json.py" rows "$t/fx1.json" "$G/fx1.json" || fail "fx1 --json: by_session is not the expected rows"
py "$G/check_json.py" rows "$t/fx1s.json" "$G/fx1.json" || fail "fx1 --by session --json: by_session is not the expected rows"
py "$G/check_json.py" sums "$t/fx1.json" 8 || fail "fx1 --json: the rows do not sum to the report"

# ---- schema --------------------------------------------------------------
run "schema validate fx1's document" -C "$R" schema validate metrics/v4 "$t/fx1.json"; ok
py "$G/check_json.py" drop "$t/fx1.json" >"$t/old.json" || fail "drop by_session"
run "schema validate a document without by_session" -C "$R" schema validate metrics/v4 "$t/old.json"; ok
py "$G/check_json.py" bad "$t/fx1.json" >"$t/bad.json" || fail "make a bad by_session"
run "schema validate a document whose by_session is a string" -C "$R" schema validate metrics/v4 "$t/bad.json"
[ "$rc" != 0 ] || fail "$n: accepted; schemas/metrics.v4.json does not declare by_session as an array of rows"

# ---- export leaves it out; --workspace --json carries it -----------------
run "metrics export" -C "$R" metrics export; ok
py "$G/check_json.py" export "$t/out" || fail "metrics export: a brief record carries by_session"
printf '[[repo]]\npath = "%s"\n' "$R" >"$t/w.toml"
run "metrics --workspace --json" --json metrics --workspace "$t/w.toml"; ok
py "$G/check_json.py" workspace "$t/out" "$G/fx1.json" || fail "metrics --workspace --json: the documents do not carry by_session"

# ---- refusals --------------------------------------------------------------
run "metrics --by phase" -C "$R" metrics "$ID" --by phase
usage 'unknown --by "phase": want task or session'
run "metrics --by session --workspace" metrics --by session --workspace "$t/w.toml"
usage "--by session reports one repository's briefs; it takes no --workspace"
run "metrics --interventions --by session" -C "$R" metrics --interventions --by session
usage "--by takes kind or phase with --interventions"

# ---- the flag's help and the docs ------------------------------------------
run "metrics --help" metrics --help; ok
norm "$t/out" | grep -Fxq -- "--by view break the summary down by view: task or session" ||
  fail "metrics --help: no line '--by view   break the summary down by view: task or session':
$(grep -F -- '--by' "$t/out")"
grep -Fxq -- '- `--by view`: break the summary down by view: task or session' "$top/docs/guide/commands.md" ||
  fail "docs/guide/commands.md is not regenerated with the new --by usage"
grep -Fq -- '--by session' "$top/docs/guide/metrics.md" && grep -Fq -- 'by_session' "$top/docs/guide/metrics.md" ||
  fail "docs/guide/metrics.md does not describe --by session and by_session"
grep -Eiq -- 'per-session|per session|by_session|--by session' "$top/docs/domain/measurement/metrics.md" ||
  fail "docs/domain/measurement/metrics.md does not say per-session rows exist"

# ---- real data: this repository's B36 ---------------------------------------
B36=B20261103-0528-b36-i96-i97-plugin-per-repo-and-agent
run "metrics B36 --by session" -C "$top" metrics "$B36" --by session; ok
norm "$t/out" >"$t/got"
[ "$(wc -l <"$t/got" | tr -d ' ')" = 14 ] || fail "$n: want a header and 13 rows:
$(cat "$t/out")"
[ "$(sed -n 1p "$t/got")" = "run # phase task agent model time cost source credits turns denied error" ] || fail "$n: header: $(sed -n 1p "$t/out")"
[ "$(sed -n 2p "$t/got")" = "20261103-062225 001 plan - claude opus 9m42s \$2.89 reported - 44 3 -" ] || fail "$n: first row: $(sed -n 2p "$t/out")"
[ "$(sed -n 14p "$t/got")" = "20261103-062225 013 verify - claude sonnet 1m38s \$0.30 reported - 18 2 -" ] || fail "$n: last row: $(sed -n 14p "$t/out")"
awk 'NR>1 {print $2}' "$t/got" | tr '\n' ' ' | grep -Fxq '001 002 003 004 005 006 007 008 009 010 011 012 013 ' || fail "$n: rows are not #001 to #013 in order"
run "metrics B36 --json" -C "$top" --json metrics "$B36"; ok
py "$G/check_json.py" sums "$t/out" 13 || fail "B36 --json: by_session does not sum to the report"
py "$G/check_json.py" b36 "$t/out" || fail "B36 --json: want 13 rows and 8 permission denials"

echo "gate T1: ok"

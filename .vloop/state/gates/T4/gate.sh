#!/bin/sh
# Gate T4 — V1's driver: after the final pass of a complete run, one session
# of phase verify (/vloop:verify, no argument), under a fence that allows what
# the review fence allows, model.verify honoured, .vloop/tmp/ cleared before
# it, vloop status naming it; its valid report kept as
# <run dir>reports/verify.json and announced; its record sessions/NNN-verify.json
# (session/v2, iteration 0, no task); both in the run commit. A missing or
# invalid report, a session that fails to run, the cost ceiling and a change
# outside .vloop/tmp/ are warnings — the run still ends complete, exit 0 — and
# a finding never changes the ending. A resume with every task done repeats
# the final pass, then the verify session. Refs moved: exit 9. With
# run.verify = false nothing of it happens.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
S=$t/stub
# verified <run out>: the run ended complete, exit 0, every task done, with
# the run commit as HEAD.
verified() {
  [ "$got" = 0 ] || fail "$1: vloop run exited $got, want 0: $(tail -n 15 "$t/$1")"
  [ "$("$JQ" -r .status "$R/.vloop/state/state.json")" = complete ] || fail "$1: the plan is not complete"
  "$JQ" -e '[.tasks[] | select(.status != "done")] | length == 0' "$R/.vloop/state/state.json" >/dev/null || fail "$1: a task is not done"
  rd=$(lastdir)
  [ "$(git -C "$R" log -1 --format=%s)" = "[vloop] run $RUN/$(basename "$rd"): complete" ] || fail "$1: HEAD is '$(git -C "$R" log -1 --format=%s)', not the run commit"
  [ -z "$(git -C "$R" status --porcelain)" ] || fail "$1: the run left the tree dirty: $(git -C "$R" status --porcelain)"
}
asked() { grep -q -- '-p /vloop:verify' "$S/argv.log"; }
noverify() { grep -E '(^|[^a-z])verify(:| session| skipped)' "$t/$1" && fail "$1: the run printed a verify line"; return 0; }
sessfiles() { ls "$R/$rd"sessions/ 2>/dev/null | grep -E '^[0-9]+-verify\.json$'; }

# --- A: two findings; model, fence, status, clearing, the commit
runrepo a T1
cat >>"$R/.vloop/config.toml" <<'CFG'

[[check]]
name = "touch"
paths = ["nomatch/**"]
run = "mkdir -p .vloop/tmp && : >.vloop/tmp/fromcheck.txt"
CFG
# Not fatal here: on a tree without T3 the missing session is the failure to report; the model is asserted below.
inr "$B" config set model.verify haiku </dev/null >"$t/o" 2>&1
commit config
vreport "$S/verify-report.json" 2
cat >"$S/work-T1.sh" <<'W'
: >.vloop/tmp/fromwork.txt
W
cat >"$S/verify.sh" <<'V'
cp "$dir/verify-report.json" .vloop/tmp/verify.json
for e in fromcheck.txt fromwork.txt proposal.json verdict.json; do
  [ -e ".vloop/tmp/$e" ] && echo "the verify session saw .vloop/tmp/$e" >>"$dir/norunning"
done
vloop status >"$dir/status.txt" 2>&1
vloop status --json >"$dir/status.json" 2>&1
V
vrun a.run
verified a.run
[ ! -s "$S/norunning" ] || fail "verify session: $(cat "$S/norunning")"
asked || fail "no /vloop:verify session ran: $(cat "$S/argv.log")"
[ "$(grep -c -- '-p /vloop:verify' "$S/argv.log")" = 1 ] || fail "more than one verify session ran"
tail -n 1 "$S/argv.log" | grep -q -- '-p /vloop:verify --model haiku' || fail "the last session was not /vloop:verify with model.verify's haiku: $(tail -n 1 "$S/argv.log")"
grep -Eq "^ *verify: 2 finding\\(s\\) — see $(echo "$rd" | sed 's/\./\\./g')reports/verify\\.json\$" "$t/a.run" ||
  fail "no 'verify: 2 finding(s) — see ${rd}reports/verify.json':
$(tail -n 15 "$t/a.run")"
[ -f "$R/${rd}reports/verify.json" ] || fail "no ${rd}reports/verify.json"
[ "$("$JQ" -S . "$R/${rd}reports/verify.json")" = "$("$JQ" -S . "$S/verify-report.json")" ] || fail "${rd}reports/verify.json is not the session's report:
$(cat "$R/${rd}reports/verify.json")"
sf=$(sessfiles)
[ "$(echo "$sf" | grep -c .)" = 1 ] || fail "want one sessions/NNN-verify.json in $rd, found: $sf"
rec=$R/${rd}sessions/$sf
inr "$B" schema validate session/v2 "$rec" </dev/null >"$t/o" 2>&1 || fail "$sf does not validate as session/v2: $(cat "$t/o")"
"$JQ" -e '.schema == "session/v2" and .phase == "verify" and .iteration == 0 and (has("task") | not) and .model == "haiku"' "$rec" >/dev/null ||
  fail "$sf is not a session/v2 verify record at iteration 0 with no task and model haiku: $(cat "$rec")"
git -C "$R" show --name-only --format= HEAD >"$t/files"
grep -Fxq "${rd}reports/verify.json" "$t/files" && grep -Fxq "${rd}sessions/$sf" "$t/files" ||
  fail "the run commit does not hold ${rd}reports/verify.json and ${rd}sessions/$sf:
$(cat "$t/files")"
[ -f "$S/fence-verify.json" ] && [ -f "$S/fence-review.json" ] || fail "harness: no fence captured for the review or verify session"
[ "$("$JQ" -S . "$S/fence-verify.json")" = "$("$JQ" -S . "$S/fence-review.json")" ] || fail "the verify session's fence does not allow exactly what the review fence allows:
$(diff "$S/fence-review.json" "$S/fence-verify.json")"
grep -q '^running: verify for ' "$S/status.txt" || fail "vloop status during the verify session: $(cat "$S/status.txt")"
"$JQ" -e '.running.phase == "verify" and (.running.task == null)' "$S/status.json" >/dev/null 2>&1 || fail "vloop status --json during the verify session: $(cat "$S/status.json")"

# --- B: no findings; the session edits README.md, adds stray.txt, edits the config
resetstub
runrepo b T1
vreport "$S/verify-report.json" 0
cat >"$S/verify.sh" <<'V'
cp "$dir/verify-report.json" .vloop/tmp/verify.json
echo changed >>README.md
: >stray.txt
echo '# changed' >>.vloop/config.toml
V
vrun b.run
verified b.run
grep -Eq '^ *verify: no findings$' "$t/b.run" || fail "no 'verify: no findings': $(tail -n 15 "$t/b.run")"
for p in README.md stray.txt; do
  grep -Fxq "vloop: the verify session changed $p — reverted" "$t/b.run" || fail "no 'vloop: the verify session changed $p — reverted':
$(tail -n 15 "$t/b.run")"
done
[ ! -e "$R/stray.txt" ] || fail "stray.txt was kept"
git -C "$R" diff --quiet main HEAD -- README.md .vloop/config.toml || fail "the verify session's edits reached the run commit"
[ -f "$R/${rd}reports/verify.json" ] || fail "the report of a verify session whose changes were reverted was not kept"

# --- C: no report; an invalid report
resetstub
runrepo c T1
printf ':\n' >"$S/verify.sh"
vrun c.run
verified c.run
grep -Fxq 'vloop: the verify session left no valid .vloop/tmp/verify.json — verify by hand' "$t/c.run" || fail "no report: $(tail -n 15 "$t/c.run")"
[ ! -e "$R/${rd}reports/verify.json" ] || fail "a report was kept though the session left none"
resetstub
runrepo c2 T1
vreport "$S/verify-report.json" 2
"$JQ" '.findings[0].kind = "bug"' "$S/verify-report.json" >"$S/v.json" && mv "$S/v.json" "$S/verify-report.json"
vrun c2.run
verified c2.run
grep -Fxq 'vloop: the verify session left no valid .vloop/tmp/verify.json — verify by hand' "$t/c2.run" || fail "invalid report: $(tail -n 15 "$t/c2.run")"
grep -Eq '^ *verify: [0-9]+ finding' "$t/c2.run" && fail "an invalid report was announced as findings"

# --- D: the session fails to run
resetstub
runrepo d T1
echo 3 >"$S/exit-verify"
vrun d.run
verified d.run
grep -Fxq "vloop: the verify session failed to run — verify by hand; see ${rd}run.log" "$t/d.run" || fail "failed session: $(tail -n 15 "$t/d.run")"

# --- E: the cost ceiling is reached by the last review
resetstub
runrepo e T1
echo 0.25 >"$S/cost"
vrun e.run --cost-ceiling 1
verified e.run
grep -Fxq 'vloop: cost ceiling reached — verify skipped' "$t/e.run" || fail "ceiling: $(tail -n 15 "$t/e.run")"
asked && fail "a verify session ran past the cost ceiling"

# --- F: run.verify = false
resetstub
runrepo f T1
inr "$B" config set run.verify false </dev/null >"$t/o" 2>&1 || fail "config set run.verify false: $(cat "$t/o")"
commit off
vreport "$S/verify-report.json" 2
vrun f.run
verified f.run
asked && fail "a verify session ran with run.verify = false"
noverify f.run
[ -z "$(sessfiles)" ] && [ ! -e "$R/${rd}reports/verify.json" ] || fail "run.verify = false left verify records"

# --- G: the final pass fails, then the resume repeats it and verifies
resetstub
runrepo g T1
cat >>"$R/.vloop/config.toml" <<CFG

[[check]]
name = "final"
paths = ["nomatch/**"]
run = "test ! -e '$S/failfinal'"
CFG
commit final
printf ': >"$dir/failfinal"\n' >"$S/plan.sh"
vreport "$S/verify-report.json" 2
vrun g1.run
[ "$got" = 2 ] || fail "harness: the failing final pass did not end the run blocked (exit $got): $(tail -n 10 "$t/g1.run")"
asked && fail "a verify session ran after a final pass that failed"
rm -f "$S/failfinal" "$S/plan.sh"
sleep 1
vrun g2.run
verified g2.run
asked || fail "the resume with every task done ran no verify session: $(cat "$S/argv.log")"
grep -Eq '^ *verify: 2 finding\(s\) — see ' "$t/g2.run" || fail "the resume printed no verify line: $(tail -n 15 "$t/g2.run")"
[ -f "$R/${rd}reports/verify.json" ] || fail "the resume's run folder has no reports/verify.json"

# --- H: the verify session moves a ref
resetstub
runrepo h T1
printf 'git tag planted-by-verify\ncp "$dir/verify-report.json" .vloop/tmp/verify.json\n' >"$S/verify.sh"
vreport "$S/verify-report.json" 0
vrun h.run
[ "$got" = 9 ] || fail "a verify session that moved a ref: exit $got, want 9: $(tail -n 10 "$t/h.run")"
echo "gate T4: ok"

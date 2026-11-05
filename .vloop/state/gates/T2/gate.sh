#!/bin/sh
# Gate T2 — the goldens. The starting point's report.py, hook.py and bypass.sh,
# taken from the base commit (so this gate still runs after T5 deletes them),
# go into a temporary directory; the committed generator,
# tools/copilot-acceptance/kit/testdata/gen.sh, regenerates the goldens there
# with them, and the result must equal the committed goldens, file for file and
# byte for byte. Then: the goldens cover every report kind and probe, carry
# @DIR@ rather than a machine's path, and the fixtures hold the characters the
# brief names; the stub copilot is executable.
G=.vloop/state/gates/T2
. "$G/lib.sh"
BASE=d7b9cbb579f82f8626c9211e5ce8f2f7e8f720aa
TD=tools/copilot-acceptance/kit/testdata
GD=$TD/golden

[ -f "$TD/gen.sh" ] || fail "no $TD/gen.sh"
[ -d "$GD" ] || fail "no committed goldens in $GD/"
[ -f "$TD/copilot" ] && [ -x "$TD/copilot" ] || fail "no executable stub $TD/copilot"
mkdir -p "$t/py" || fail mkdir
for f in report.py hook.py bypass.sh; do
  git -C "$top" show "$BASE:tools/copilot-acceptance/$f" >"$t/py/$f" || fail "git show the base's $f"
done

sh "$TD/gen.sh" "$t/py" "$S" "$t/out" >"$t/gen.log" 2>&1 || fail "gen.sh exited non-zero: $(tail -n 20 "$t/gen.log")"
[ -f "$t/py/calls.py" ] && grep -q 'tool.execution_start' "$t/py/calls.py" || fail "gen.sh did not extract bypass.sh's calls heredoc into <python dir>/calls.py"
[ -d "$t/out" ] && [ -n "$(ls -A "$t/out")" ] || fail "gen.sh wrote nothing to its <out dir>"
diff -r "$t/out" "$GD" >"$t/diff" 2>&1 || fail "the regenerated goldens differ from the committed ones in $GD/:
$(head -n 40 "$t/diff")"

# Every report kind and both probes' reports are among the goldens.
for h in '## C5.1 — a real run with Copilot reviews' \
  '## B22.3 — a feedback-mode run with a copilot second reviewer' \
  '## A1 — a real run with Copilot work sessions' \
  '## A1 — a real run with a Copilot verify session' \
  '## A1 — a real run with Copilot gate-review sessions' \
  '## A1 — a real run with Copilot plan sessions' \
  '## C5.2 — the deny probe' \
  '## B22.1 — the review session under `vloop fence hook`'; do
  grep -rFxq -- "$h" "$GD" || fail "no golden holds the report headed '$h'"
done
# The files the probes write.
for f in probe.argv probe.rc hook-forms.argv hook-forms.rc \
  hook-fail-missing.argv hook-fail-missing.rc hook-fail-exit1.argv hook-fail-exit1.rc \
  hook-fail-notjson.argv hook-fail-notjson.rc hook-fail-timeout.argv hook-fail-timeout.rc vloop-fence.json; do
  [ -n "$(find "$GD" -type f -name "$f" | head -n 1)" ] || fail "no golden of the file $f the probes write"
done
grep -rFq -- 'fence hook --phase review ' "$GD" || fail "no golden hooks file with --phase review"
# The stub's records: argv, working directory and environment, @DIR@ in place.
grep -rq '^arg -p$' "$GD" && grep -rq '^cwd @DIR@' "$GD" && grep -rq '^env STUB_LOG=@DIR@/' "$GD" && grep -rq '^env PATH=' "$GD" ||
  fail "no golden of the stub's log (arg, cwd @DIR@, env STUB_LOG=@DIR@/…, env PATH=…)"
! grep -rqE '^env (PWD|OLDPWD|SHLVL|_)=' "$GD" || fail "a golden of the stub's log keeps PWD, OLDPWD, SHLVL or _"
! grep -rqE '/private/var/|/var/folders/|/tmp/gate\.' "$GD" || fail "a golden holds a machine's temporary path instead of @DIR@: $(grep -rlE '/private/var/|/var/folders/|/tmp/gate\.' "$GD")"
grep -rqF '@DIR@' "$GD" || fail "no golden carries @DIR@"
# The fixtures hold what the brief names, so the goldens test each rendering rule.
fx=$t/fixtures.txt
find "$TD" -path "$GD" -prune -o -type f -print | while IFS= read -r f; do cat "$f"; echo; done >"$fx"
grep -q "$(printf '\360')" "$fx" || fail "no fixture holds an emoji (a character outside the BMP)"
grep -q "$(printf '\303')" "$fx" || fail "no fixture holds non-ASCII text"
grep -qF '<&>' "$fx" || fail "no fixture holds <&> in a string"
grep -qF '\t' "$fx" && grep -qF '\n' "$fx" || fail "no fixture holds a string with a newline and a tab"

echo "gate T2: ok"

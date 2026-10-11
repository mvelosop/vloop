#!/bin/sh
# Gate T2 — vloop run hands .vloop/repo-guidelines.md to every session (brief
# L1): with a non-empty file every session's arguments carry
# --append-system-prompt-file .vloop/repo-guidelines.md; with no file, or an
# empty one, they are exactly today's.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
FLAG='--append-system-prompt-file .vloop/repo-guidelines.md'
# norm <out>: the logged sessions with the run-specific paths masked.
norm() { sed -e "s#$t/[a-z]*/#<repo>/#g" -e "s/\$/ /" "$t/stub/argv.log" >"$t/argv.$1"; }
# phases <out>: each of the four session kinds appears at least once.
phases() {
  for p in /vloop:plan /vloop:gate-review /vloop:work /vloop:review; do
    grep -Fq -- "-p $p" "$t/argv.$1" || fail "$1: no $p session was started: $(cat "$t/argv.$1")"
  done
}

# No file: today's arguments, and no flag.
runrepo none
[ ! -e "$R/.vloop/repo-guidelines.md" ] || { rm "$R/.vloop/repo-guidelines.md" && commit "no guidelines"; }
vrun none.run
[ "$got" = 0 ] || fail "vloop run with no guidelines file: exit $got: $(tail -n 20 "$t/none.run")"
norm none; phases none
grep -Fq -- '--append-system-prompt-file' "$t/argv.none" && fail "no guidelines file, yet a session got --append-system-prompt-file: $(cat "$t/argv.none")"

# An empty file: byte for byte the same arguments as no file.
runrepo empty
: >"$R/.vloop/repo-guidelines.md" && commit "empty guidelines"
vrun empty.run
[ "$got" = 0 ] || fail "vloop run with an empty guidelines file: exit $got: $(tail -n 20 "$t/empty.run")"
norm empty
cmp -s "$t/argv.none" "$t/argv.empty" || fail "an empty guidelines file changed the sessions' arguments:
no file:
$(cat "$t/argv.none")
empty file:
$(cat "$t/argv.empty")"

# A non-empty file: every session gets the flag, and nothing else changes.
runrepo full
printf '# Repository guidelines\n\n- A lesson. (D20261011-1455)\n' >"$R/.vloop/repo-guidelines.md" && commit "guidelines"
vrun full.run
[ "$got" = 0 ] || fail "vloop run with a guidelines file: exit $got: $(tail -n 20 "$t/full.run")"
norm full; phases full
n=$(wc -l <"$t/argv.full" | tr -d ' ')
k=$(grep -Fc -- " $FLAG " "$t/argv.full")
[ "$n" -gt 0 ] && [ "$k" = "$n" ] || fail "$k of $n sessions got '$FLAG':
$(cat "$t/argv.full")"
sed -e "s# $FLAG # #" "$t/argv.full" >"$t/stripped"
cmp -s "$t/stripped" "$t/argv.none" || fail "with the flag removed, the sessions' arguments differ from today's:
with the file, flag removed:
$(cat "$t/stripped")
no file:
$(cat "$t/argv.none")"
echo "gate T2: ok"

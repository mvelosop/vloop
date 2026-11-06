#!/bin/sh
# Gate T3 — Z3 and Z4: `release prepare` refuses a module that does not zip,
# after "main is not origin/main" and before the configured checks, writing
# nothing; and go.mod retracts v0.15.0-rc.2. Through the release tool built
# from this tree, in fixture repositories made here under $t, with stub claude
# and gh on PATH and a [[check]] that leaves a marker when it runs.
G=.vloop/state/gates/T3
. "$G/lib.sh"
buildrel

mkdir -p "$t/bin" || fail mkdir
for c in claude gh; do
  printf '#!/bin/sh\necho "$*" >>"%s/%s.ran"\nexit 0\n' "$t" "$c" >"$t/bin/$c" && chmod +x "$t/bin/$c" || fail "stub $c"
done
PATH="$t/bin:$PATH"; export PATH

# prepmod NAME: newmod NAME with a .vloop/config.toml whose one check leaves
# $t/check.ran, committed, and an origin with main pushed.
prepmod() {
  newmod "$1"
  mkdir -p "$F/.vloop" && printf '[[check]]\nname = "marker"\npaths = ["**"]\nrun = "touch %s/check.ran"\n' "$t" >"$F/.vloop/config.toml" ||
    fail "config in $1"
  commitall "$F" config
}
# norun WHAT: neither the configured check nor claude nor gh ran.
norun() {
  for m in check claude gh; do
    [ ! -e "$t/$m.ran" ] || fail "$1: $m ran after the refusal"
  done
}
# snapshot DIR FILE: DIR's refs, status and HEAD, in FILE.
snapshot() {
  { git -C "$1" for-each-ref --format='%(refname) %(objectname)'; git -C "$1" status --porcelain; git -C "$1" rev-parse HEAD; } >"$2" ||
    fail "snapshot $1"
}

# A module that zips: --dry-run passes, and the checks after the zip still run.
prepmod good
withorigin
runin "$F" prepare v0.1.0 --dry-run
[ "$rc" = 0 ] || fail "prepare v0.1.0 --dry-run on a module that zips: exit $rc, want 0
stderr:
$(cat "$t/err")"
[ "$(head -n 1 "$t/out")" = "dry run: every check passed for v0.1.0; nothing written" ] ||
  fail "prepare --dry-run on a module that zips: stdout $(cat "$t/out")"
[ -e "$t/check.ran" ] && [ -e "$t/claude.ran" ] || fail "prepare --dry-run on a module that zips did not run the check and claude plugin validate"
rm -f "$t/check.ran" "$t/claude.ran" "$t/gh.ran"

# The same module from a linked worktree: CreateFromVCS finds no .git directory.
git -C "$F" worktree add -q --force "$t/wt" main 2>"$t/wt.err" || fail "git worktree add: $(cat "$t/wt.err")"
runin "$t/wt" prepare v0.1.0 --dry-run
want "prepare --dry-run in a linked worktree" 1 /dev/null "$G/worktree.err"
norun "prepare in a linked worktree"

# A committed and pushed it's repo/.keep: one line per rejected file, the count, exit 1.
prepmod bad
mkdir -p "$F/it's repo" && : >"$F/it's repo/.keep" || fail "fixture bad"
commitall "$F" keep
withorigin
snapshot "$F" "$t/before"
runin "$F" prepare v0.1.0 --dry-run
want "prepare v0.1.0 --dry-run with a committed it's repo/.keep" 1 /dev/null "$G/zip.err"
norun "prepare on a module that does not zip"
snapshot "$F" "$t/after"
cmp -s "$t/before" "$t/after" || fail "prepare changed the fixture's refs, status or HEAD:
$(diff "$t/before" "$t/after")"

# The zip check comes after "main is not origin/main".
printf 'x\n' >"$F/later.txt" && commitall "$F" later
runin "$F" prepare v0.1.0 --dry-run
want "prepare with an unpushed commit and a rejected path" 1 /dev/null "$G/notorigin.err"
norun "prepare with an unpushed commit"

# Z4: go.mod ends with the existing retraction, then v0.15.0-rc.2's.
tail -n 7 go.mod >"$t/tail" || fail "read go.mod"
cmp -s "$t/tail" "$G/gomod.tail" || fail "go.mod does not end with the retraction of v0.15.0-rc.2:
$(cat "$t/tail")"
runin "$top" prepare v0.15.0-rc.2 --dry-run
want "prepare v0.15.0-rc.2 --dry-run in this repository" 1 /dev/null "$G/retracted.err"

echo "gate T3: ok"

#!/bin/sh
# Gate T2 — Z2, `release paths`: the worked example's `paths` lines, the usage
# lines and the error case, through the release tool built from this tree, in
# fixture repositories made here under $t. The expected outputs are the files
# beside this script (their contents hold ', their names do not).
G=.vloop/state/gates/T2
. "$G/lib.sh"
buildrel

# The usage line of tools/release names paths.
runin "$t/elsewhere"; want "release (no arguments)" 2 /dev/null "$G/usage.err"

# This repository: no rejected path, no output, exit 0.
runin "$top" paths; want "paths in this repository" 0 /dev/null /dev/null

# Any argument: usage, exit 2.
runin "$top" paths extra; want "paths extra" 2 /dev/null "$G/usage-paths.err"

# A fixture with only go.mod and x.go: no line.
newmod plain
printf 'package m\n' >"$F/x.go" && commitall "$F" x
runin "$F" paths; want "paths in a fixture with go.mod and x.go" 0 /dev/null /dev/null

# A committed it's repo/.keep: its line on stdout, the count on stderr, exit 1.
newmod keep
mkdir -p "$F/it's repo" "$F/sub" && : >"$F/it's repo/.keep" && printf 'package sub\n' >"$F/sub/x.go" || fail "fixture keep"
commitall "$F" keep
runin "$F" paths; want "paths with a committed it's repo/.keep" 1 "$G/keep.out" "$G/one.err"
# ... and the same from a subdirectory: the paths are the top directory's.
runin "$F/sub" paths; want "paths from a subdirectory of that fixture" 1 "$G/keep.out" "$G/one.err"

# An untracked, not ignored a'b.txt: its line, exit 1.
newmod untracked
: >"$F/a'b.txt" || fail "fixture untracked"
runin "$F" paths; want "paths with an untracked a'b.txt" 1 "$G/ab.out" "$G/one.err"

# The same a'b.txt listed in .gitignore: no line.
printf '%s\n' "a'b.txt" >"$F/.gitignore" || fail "fixture ignored"
runin "$F" paths; want "paths with a'b.txt ignored" 0 /dev/null /dev/null

# Both at once: one line per rejected path, in the order git lists them.
newmod both
mkdir -p "$F/it's repo" && : >"$F/it's repo/.keep" || fail "fixture both"
commitall "$F" keep
: >"$F/a'b.txt" || fail "fixture both"
git -C "$F" ls-files -z --cached --others --exclude-standard | tr '\0' '\n' | grep "'" >"$t/order" || fail "git ls-files in both"
while IFS= read -r p; do
  printf '%s: malformed file path "%s": invalid char %s\n' "$p" "$p" "'\\''"
done <"$t/order" >"$t/both.out"
runin "$F" paths; want "paths with a committed and an untracked rejected path" 1 "$t/both.out" "$G/two.err"

# Outside any repository: git's error on one line, exit 1, nothing on stdout.
mkdir -p "$t/norepo" || fail mkdir
(cd "$t/norepo" && GIT_CEILING_DIRECTORIES="$t" "$t/release" paths) >"$t/out" 2>"$t/err"; rc=$?
[ "$rc" = 1 ] || fail "paths outside a repository: exit $rc, want 1 (stderr: $(cat "$t/err"))"
[ ! -s "$t/out" ] || fail "paths outside a repository printed on stdout: $(cat "$t/out")"
[ "$(wc -l <"$t/err" | tr -d ' ')" = 1 ] || fail "paths outside a repository: stderr is not one line:
$(cat "$t/err")"
grep -q '^release: .*not a git repository' "$t/err" ||
  fail "paths outside a repository: stderr is not release: <git's error>: $(cat "$t/err")"

echo "gate T2: ok"

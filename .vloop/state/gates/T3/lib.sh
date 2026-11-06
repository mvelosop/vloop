# Shared harness for B38's gates (the same file in every gate folder). The
# sourcing gate sets G, its own folder. Every build and every fixture
# repository lives under $t, a temporary directory removed on exit; nothing
# touches this repository's tree or refs. No path holding ' is ever written
# into the repository: such paths are made under $t, at run time.
set -u
fail() { printf 'GATE FAIL: %s\n' "$*" >&2; exit 1; }
unset GOFLAGS GOBIN GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE GIT_CEILING_DIRECTORIES
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1
top=$(pwd -P)
[ -f "$top/go.mod" ] && [ -d "$top/tools/release" ] || fail "run from the repository root"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
mkdir -p "$t/elsewhere" || fail mkdir

# buildrel: the release tool built from this tree, at $t/release.
buildrel() {
  go build -buildvcs=false -o "$t/release" ./tools/release >"$t/build.out" 2>&1 ||
    fail "go build ./tools/release: $(tail -n 5 "$t/build.out")"
}

# commitall DIR MSG: commit everything in the fixture repository DIR.
commitall() {
  git -C "$1" add -A && git -C "$1" -c user.name=t -c user.email=t@example.com -c commit.gpgsign=false commit -q -m "$2" ||
    fail "commit in $1"
}

# newmod NAME: a fixture repository $t/NAME on main holding go.mod
# (module example.com/m, go 1.27), one commit; sets F.
newmod() {
  F=$t/$1
  mkdir -p "$F" && git -C "$F" init -q -b main && printf 'module example.com/m\n\ngo 1.27\n' >"$F/go.mod" || fail "init $1"
  commitall "$F" init
}

# withorigin: a bare repository $t/<name of F>.git as F's origin, main pushed.
withorigin() {
  o=$F.git
  git init -q --bare "$o" && git -C "$F" remote add origin "$o" && git -C "$F" push -q origin main 2>"$t/push.err" ||
    fail "origin for $F: $(cat "$t/push.err")"
}

# runin DIR ARGS: the built release tool run with DIR as its working directory;
# stdout in $t/out, stderr in $t/err, the exit code in rc.
runin() { d=$1; shift; (cd "$d" && "$t/release" "$@") >"$t/out" 2>"$t/err"; rc=$?; }

# want NAME RC OUT ERR: the last runin exited RC, and its stdout and stderr
# equal the files OUT and ERR byte for byte (/dev/null for nothing).
want() {
  [ "$rc" = "$2" ] || fail "$1: exit $rc, want $2
stdout:
$(cat "$t/out")
stderr:
$(cat "$t/err")"
  cmp -s "$t/out" "$3" || fail "$1: stdout differs.
got:
$(cat "$t/out")
want:
$(cat "$3")"
  cmp -s "$t/err" "$4" || fail "$1: stderr differs.
got:
$(cat "$t/err")
want:
$(cat "$4")"
}

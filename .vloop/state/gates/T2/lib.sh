# Shared harness for B37's gates (the same file in every gate folder). The
# sourcing gate sets G, its own folder. The binary, the fake home and every
# fixture repository live under $t, a temporary directory removed on exit;
# nothing touches this repository's tree or refs, and no real agent CLI runs.
set -u
fail() { printf 'GATE FAIL: %s\n' "$*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset GOFLAGS GOBIN GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE COPILOT_GITHUB_TOKEN GH_TOKEN GITHUB_TOKEN COPILOT_HOME STUB_LOG STUB_OUT STUB_RC
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1
export PYTHONDONTWRITEBYTECODE=1 PYTHONUTF8=1 PYTHONIOENCODING=utf-8
top=$(pwd -P)
[ -f "$top/go.mod" ] && [ -d "$top/cmd/vloop" ] || fail "run from the repository root"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
mkdir -p "$t/home" "$t/elsewhere" || fail mkdir
go build -buildvcs=false -o "$t/vloop" ./cmd/vloop >"$t/build.out" 2>&1 || fail "go build ./cmd/vloop: $(tail -n 5 "$t/build.out")"
S=$t/vloop
# vl ARGS: the built vloop with the fake home, run from a directory outside
# every fixture repository (each call names its repository with -C).
vl() { (cd "$t/elsewhere" && HOME="$t/home" "$S" "$@"); }

# py ARGS: python3, isolated, writing no bytecode. Used only to parse JSON and,
# in the kit's gates, to run the starting point's reference Python.
py() { python3 -I -B "$@"; }
command -v python3 >/dev/null || fail "python3 is needed by this gate (to parse JSON / run the reference Python)"

# newrepo NAME: a fresh git repository $t/NAME with one empty commit; sets R.
newrepo() {
  R=$t/$1
  mkdir -p "$R" && git -C "$R" init -q -b main && git -C "$R" -c user.name=t -c user.email=t@example.com commit -q --allow-empty -m base || fail "git init $1"
}

# commitall DIR: commit everything in the fixture repository DIR.
commitall() { git -C "$1" add -A && git -C "$1" -c user.name=t -c user.email=t@example.com commit -q -m fixtures || fail "commit in $1"; }

# hasline FILE LINE: FILE holds LINE as a whole line.
hasline() { grep -Fxq -- "$2" "$1"; }

# norm FILE: FILE with every run of blanks collapsed and leading/trailing blanks dropped.
norm() { awk '{$1=$1; print}' "$1"; }

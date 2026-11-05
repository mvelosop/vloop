# Shared harness for B36's gates (the same file in T1, T2 and T3). The
# sourcing gate sets G, its own folder. The binary, the fake home and every
# fixture repository live under $t, a temporary directory removed on exit;
# nothing touches this repository's tree or refs, and no real agent CLI runs.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset GOFLAGS GOBIN GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE COPILOT_GITHUB_TOKEN GH_TOKEN GITHUB_TOKEN COPILOT_HOME
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1
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
V=$(vl version | awk 'NR==1 {print $2}')
[ -n "$V" ] || fail "vloop version printed no version"

# newrepo NAME: a fresh git repository $t/NAME with one empty commit; sets R.
newrepo() {
  R=$t/$1
  mkdir -p "$R" && git -C "$R" init -q -b main && git -C "$R" -c user.name=t -c user.email=t@example.com commit -q --allow-empty -m base || fail "git init $1"
}

# snap DIR: a listing of every file under DIR but .git, with its checksum.
snap() { (cd "$1" && find . -path ./.git -prune -o \( -type f -o -type l \) -print | sort | while read -r f; do printf '%s %s\n' "$f" "$(cksum <"$f" 2>/dev/null)"; done); }

# hasline FILE LINE: FILE holds LINE as a whole line.
hasline() { grep -Fxq -- "$2" "$1"; }

# flat FILE: FILE's text with every run of whitespace collapsed to one space.
flat() { tr '\n\t\r' '   ' <"$1" | tr -s ' '; }

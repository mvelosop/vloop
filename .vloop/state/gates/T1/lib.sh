# Shared harness for B28's gates (the same file in every gate folder). The
# sourcing gate sets G, its own folder. Everything a gate builds, copies or
# plants lives under $t, a temporary directory removed on exit: Go oracles run
# in $t/src, a copy of this tree, never in the tree itself. No gate runs
# claude, copilot, an eval or touches the network.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset SKILLCHECK_PLUGIN_DIR GOFLAGS GOBIN GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT
export LC_ALL=C GOTOOLCHAIN=local GOPROXY=off GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0
top=$(pwd -P)
[ -f "$top/go.mod" ] && [ -d "$top/$G" ] || fail "run the gate from the repository root"
GITBIN=$(command -v git) || fail "git is not on PATH"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'chmod -R u+w "$t" 2>/dev/null; rm -rf "$t"' EXIT
GOMODCACHE=$(go env GOMODCACHE) && GOCACHE=$(go env GOCACHE) || fail "go env"
export GOMODCACHE GOCACHE
# The commit this plan was written on: the originals of what moves.
BASE=3655c351a577ef4925612aa4b5aba230c1b0fd80

# srccopy: $t/src, a copy of this tree's files — tracked and untracked, not
# ignored, that exist — but vloop's run state.
srccopy() {
  mkdir -p "$t/src" || fail mkdir
  git -C "$top" ls-files -co --exclude-standard >"$t/files.all" || fail "git ls-files"
  : >"$t/files"
  while IFS= read -r f; do
    case $f in .vloop/state/* | .loop/*) continue ;; esac
    [ -f "$top/$f" ] && printf '%s\n' "$f" >>"$t/files"
  done <"$t/files.all"
  (cd "$top" && tar -cf - -T "$t/files") | (cd "$t/src" && tar -xf -) || fail "copy of the tree"
}

# oracle <file> <dir>: the gate's oracle test file copied into the copy's
# package <dir>.
oracle() { cp "$top/$G/$1" "$t/src/$2/$1" || fail "oracle $1"; }

# newbin: $S, vloop built from the copy.
newbin() {
  mkdir -p "$t/bin" && (cd "$t/src" && go build -buildvcs=false -o "$t/bin/vloop" ./cmd/vloop) || fail "go build of this tree"
  S=$t/bin/vloop
}

# venv cmd...: cmd in a bare environment with a fake HOME.
venv() {
  env -i PATH="$t/stub:/usr/bin:/bin" HOME="$t/home" USER=gateuser LOGNAME=gateuser TMPDIR="$t/tmp" LC_ALL=C \
    GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0 "$@"
}

# freshrepo: $t/repo, a fresh git repository on main in which this tree's
# vloop ran `vloop init`; its output is $t/init.out.
freshrepo() {
  mkdir -p "$t/repo" "$t/home" "$t/tmp" "$t/stub" && ln -s "$GITBIN" "$t/stub/git" || fail mkdir
  venv git -C "$t/repo" init -q -b main || fail "git init"
  (cd "$t/repo" && venv "$S" init >"$t/init.out" 2>&1) || fail "vloop init in a fresh repository: $(cat "$t/init.out")"
}

# listing <dir>: the entries of a directory on one line.
listing() { ls "$1" | tr '\n' ' ' | sed 's/ $//'; }

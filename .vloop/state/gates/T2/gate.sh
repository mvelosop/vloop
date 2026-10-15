#!/bin/sh
# Gate T2 — I1: vloop init and vloop upgrade write "module": "<module version>"
# into .vloop/install.json, right after commit, when the binary carries a module
# version, and leave it out otherwise; install/v1 takes module as an optional
# string. The module-carrying binary is installed the way a user installs a
# release: go install …@v0.12.0 through a module proxy (a local file proxy).
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
xbuild
go build -buildvcs=false -o "$t/vloop" ./cmd/vloop || fail "go build ./cmd/vloop"
B=$t/vloop
proxybin v0.12.0 0.12.0
[ "$("$P" version --json | "$JQ" -r '.module + " " + .commit')" = "v0.12.0 unknown" ] ||
  fail "harness: the proxy-installed binary does not carry module v0.12.0 and no commit: $("$P" version --json)"

# vl <binary> <repo> <args...>: vloop in <repo>, temporary HOME; output in $t/vl.out.
vl() {
  b=$1 r=$2; shift 2
  (cd "$r" && env -i PATH="$t/gitonly:$SYS" HOME="$t/home" TMPDIR="$t/tmp" LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null \
    GIT_CONFIG_NOSYSTEM=1 "$b" "$@") </dev/null >"$t/vl.out" 2>&1 || fail "vloop $* in $r: $(cat "$t/vl.out")"
}
newrepo() { mkdir -p "$1" && git -C "$1" init -q -b main && printf '# r\n' >"$1/README.md" || fail "git init $1"; }
# hasmodule <stamp> <module>: module is <module> and comes right after commit.
hasmodule() {
  [ "$("$JQ" -r '.module // "(none)"' "$1")" = "$2" ] || fail "$3: $1 has module '$("$JQ" -r '.module // "(none)"' "$1")', want '$2': $(cat "$1")"
  [ "$("$JQ" -r '[keys_unsorted[]] | (index("module") - index("commit"))' "$1")" = 1 ] ||
    fail "$3: module does not follow commit in $1: $(cat "$1")"
  [ "$("$JQ" -r .commit "$1")" = unknown ] || fail "$3: commit is not unknown in $1: $(cat "$1")"
}

# init by a binary that carries a module version.
newrepo "$t/r1"; vl "$P" "$t/r1" init
hasmodule "$t/r1/.vloop/install.json" v0.12.0 "vloop init"
"$B" schema validate install/v1 "$t/r1/.vloop/install.json" >"$t/sv" 2>&1 || fail "install/v1 rejects the stamp init wrote: $(cat "$t/sv")"

# init by a binary that carries none, then upgrade by one that does.
newrepo "$t/r2"; vl "$B" "$t/r2" init
[ "$("$JQ" 'has("module")' "$t/r2/.vloop/install.json")" = false ] ||
  fail "vloop init by a binary with no module version wrote a module: $(cat "$t/r2/.vloop/install.json")"
vl "$P" "$t/r2" upgrade --yes
hasmodule "$t/r2/.vloop/install.json" v0.12.0 "vloop upgrade"
[ "$("$JQ" -r .version "$t/r2/.vloop/install.json")" = 0.12.0 ] || fail "vloop upgrade did not record version 0.12.0: $(cat "$t/r2/.vloop/install.json")"

# upgrade by a binary with no module version leaves none, even where one was.
newrepo "$t/r3"; vl "$B" "$t/r3" init
"$JQ" '.version = "0.10.0" | .module = "v0.10.0"' "$t/r3/.vloop/install.json" >"$t/st.json" &&
  mv "$t/st.json" "$t/r3/.vloop/install.json" || fail "stamp edit"
vl "$B" "$t/r3" upgrade --yes
[ "$("$JQ" 'has("module")' "$t/r3/.vloop/install.json")" = false ] ||
  fail "vloop upgrade by a binary with no module version left a module in the stamp: $(cat "$t/r3/.vloop/install.json")"

# The schema: module is an optional string.
printf '{"schema":"install/v1","version":"0.12.0","commit":"unknown","module":5,"initialized":"2026-01-01T00:00:00Z","upgraded":null}\n' >"$t/bad.json"
if "$B" schema validate install/v1 "$t/bad.json" >"$t/sv" 2>&1; then fail "install/v1 accepts a module that is not a string: $(cat "$t/sv")"; fi
printf '{"schema":"install/v1","version":"0.12.0","commit":"abc","initialized":"2026-01-01T00:00:00Z","upgraded":null}\n' >"$t/plain.json"
"$B" schema validate install/v1 "$t/plain.json" >"$t/sv" 2>&1 || fail "install/v1 rejects a stamp without module: $(cat "$t/sv")"
echo "gate T2: ok"

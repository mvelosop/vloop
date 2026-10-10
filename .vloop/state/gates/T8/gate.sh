#!/bin/sh
# Gate T8 — an installed binary names its version (brief R6): a stamped build
# as today; a build in a checkout says the VCS revision, -dirty when modified;
# a go install <module>@<release> from a local module proxy (nothing is
# published) prints the module version in place of the unknown commit, and
# doctor's self-hosting check passes it. Every expectation is read from what
# go version -m says the binary carries.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0
export GIT_AUTHOR_NAME=gate GIT_AUTHOR_EMAIL=gate@example.invalid GIT_COMMITTER_NAME=gate GIT_COMMITTER_EMAIL=gate@example.invalid
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
top=$(pwd -P)
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
go build -o "$t/vloop" ./cmd/vloop || fail "go build ./cmd/vloop"
GOOS=linux go build -o "$t/vloop-linux" ./cmd/vloop || fail "GOOS=linux go build ./cmd/vloop"
GOOS=windows go build -o "$t/vloop.exe" ./cmd/vloop || fail "GOOS=windows go build ./cmd/vloop"
B="$t/vloop"
# Go keeps its caches under HOME unless told; keep the real ones across the change.
GOMODCACHE=$(go env GOMODCACHE) GOCACHE=$(go env GOCACHE) GOPATH=$(go env GOPATH) || fail "go env"
export GOMODCACHE GOCACHE GOPATH GOPROXY=off
mkdir -p "$t/Users/gateuser" || fail mkdir
export HOME="$t/Users/gateuser"
N=B20260101-0900-demo
BRIEF=docs/briefs/$N.loop-brief.md
trust() { printf '{"projects":{"%s":{"hasTrustDialogAccepted":true}}}\n' "$R" >"$HOME/.claude.json"; }
# runrepo <name> [<plan>]: a vloop repository on main with a ready brief and a
# check (after $CONFIG, when set), committed; its own stub claude playing
# <plan> (plan.json by default) from $t/<name>.stub.
runrepo() {
  R="$t/$1" S="$t/$1.stub"
  mkdir -p "$R" "$S" && git -C "$R" init -q -b main || fail "git init $1"
  cp "$G/claude" "$S/claude" && chmod +x "$S/claude" && cp "$G/${2:-plan.json}" "$S/plan.json" || fail "stub for $1"
  git -C "$R" config user.name gate && git -C "$R" config user.email gate@example.invalid || fail "identity"
  trust
  (cd "$R" && "$B" init </dev/null) >"$t/init.out" 2>&1 || fail "vloop init in $1: $(cat "$t/init.out")"
  printf '%s[[check]]\nname = "all"\npaths = ["**"]\nrun = "true"\n' "${CONFIG:-}" >"$R/.vloop/config.toml" || fail config
  mkdir -p "$R/docs/briefs" && echo x >"$R/docs/x.md" && cp "$G/brief.md" "$R/$BRIEF" || fail brief
  git -C "$R" add -A && git -C "$R" commit -q -m harness || fail "commit $1"
}
# vrun <out> [VAR=value...]: vloop run <brief> in $R with its stub on PATH;
# stdout in $t/<out>, stderr in $t/<out>.err; returns vloop's exit code.
vrun() {
  o=$1; shift
  trust
  (cd "$R" && env PATH="$S:$PATH" "$@" "$B" run "$BRIEF") >"$t/$o" 2>"$t/$o.err"
}
# records <glob>: the session records of the brief's run folders matching <glob>, one path per line.
records() { ls "$R"/.vloop/state/runs/"$N"/*/sessions/$1 2>/dev/null; }

mkdir -p "$t/bin" && cp "$G/claude" "$t/bin/claude" && chmod +x "$t/bin/claude" || fail "stub"
# expect <binary> <commit or ""> <label>: version --json and the text line agree
# with what go version -m says the build carries.
expect() {
  bin=$1 commit=$2 label=$3
  "$bin" version --json >"$t/v.json" 2>"$t/err" || fail "$label: version --json: $(cat "$t/err")"
  "$bin" version >"$t/v.txt" 2>"$t/err" || fail "$label: version: $(cat "$t/err")"
  mod=$(go version -m "$bin" | awk '$1 == "mod" { print $3; exit }')
  case "$mod" in "" | "(devel)") mod="";; esac
  keys="version,commit,plugin,go,os,arch"
  [ -n "$mod" ] && keys="version,commit,module,plugin,go,os,arch"
  [ "$(jq -r 'keys_unsorted | join(",")' "$t/v.json")" = "$keys" ] || fail "$label: version --json keys are $(jq -c 'keys_unsorted' "$t/v.json"), want $keys (the build's module version is '${mod:-none}')"
  [ -z "$mod" ] || jq -e --arg m "$mod" '.module == $m' "$t/v.json" >/dev/null || fail "$label: version --json module is $(jq .module "$t/v.json"), want $mod"
  jq -e --arg c "$commit" '.commit == $c' "$t/v.json" >/dev/null || fail "$label: version --json commit is $(jq .commit "$t/v.json"), want $commit"
  if [ "$commit" = unknown ] && [ -n "$mod" ]; then
    want=$(jq -r '"vloop \(.version) (module \(.module), plugin \(.plugin), \(.go) \(.os)/\(.arch))"' "$t/v.json")
  else
    want=$(jq -r '"vloop \(.version) (commit \(.commit), plugin \(.plugin), \(.go) \(.os)/\(.arch))"' "$t/v.json")
  fi
  [ "$(cat "$t/v.txt")" = "$want" ] || fail "$label: vloop version printed '$(cat "$t/v.txt")', want '$want'"
}

# The tree as it stands (tracked and untracked files, less the top-level dot
# folders), committed in a checkout of its own.
git ls-files -co --exclude-standard | grep -v '^\.' | while IFS= read -r f; do [ -f "$f" ] && printf '%s\n' "$f"; done >"$t/files"
grep -qx go.mod "$t/files" && grep -qx cmd/vloop/main.go "$t/files" || fail "the file list has no go.mod or cmd/vloop/main.go"
mkdir -p "$t/src" && tar -cf - -T "$t/files" | tar -xf - -C "$t/src" || fail "copy the tree"
git -C "$t/src" init -q && git -C "$t/src" add -A && git -C "$t/src" commit -q -m src || fail "commit the copy"
rev=$(git -C "$t/src" rev-parse HEAD)

# --- go install <module>@<release>: the module version names the binary ---
MV=v0.88.8
MC=$(go env GOMODCACHE) && [ -d "$MC/cache/download" ] || fail "no module download cache to serve the dependencies from"
P="$t/proxy/github.com/mvelosop/vloop/@v" Z="$t/z/github.com/mvelosop/vloop@$MV"
mkdir -p "$P" "$Z" && tar -cf - -T "$t/files" | tar -xf - -C "$Z" || fail "stage the module"
(cd "$t/z" && zip -qrD "$P/$MV.zip" "github.com/mvelosop/vloop@$MV") || fail "zip the module"
cp go.mod "$P/$MV.mod" && echo "$MV" >"$P/list" && printf '{"Version":"%s","Time":"2026-01-01T00:00:00Z"}\n' "$MV" >"$P/$MV.info" || fail "proxy files"
env GOFLAGS=-modcacherw GOMODCACHE="$t/mc" GOPROXY="file://$t/proxy,file://$MC/cache/download" GOSUMDB=off GOPRIVATE= GONOPROXY= GONOSUMDB= \
  GOTOOLCHAIN=local GOWORK=off GOBIN="$t/inst" go install -trimpath "github.com/mvelosop/vloop/cmd/vloop@$MV" >"$t/inst.log" 2>&1 \
  || fail "go install from the local module proxy: $(tail -5 "$t/inst.log")"
go version -m "$t/inst/vloop" | awk '$1 == "mod" { print $3 }' | grep -qx "$MV" || fail "the installed binary does not carry module version $MV"
expect "$t/inst/vloop" unknown "a binary installed at $MV"
(cd "$t/src" && env PATH="$t/bin:$PATH" "$t/inst/vloop" doctor --json) >"$t/doc.json" 2>/dev/null
jq -e '.checks[] | select(.name == "self-hosting") | .result == "pass"' "$t/doc.json" >/dev/null \
  || fail "doctor's self-hosting check does not pass a binary installed at release $MV: $(jq -c '.checks[] | select(.name == "self-hosting")' "$t/doc.json")"

# --- the stamped build: as today ---
go build -ldflags "-X main.commit=gatestamp42" -o "$t/stamped" ./cmd/vloop || fail "stamped build"
expect "$t/stamped" gatestamp42 "a stamped build"

# --- a build inside a checkout: the VCS revision, -dirty when modified ---
(cd "$t/src" && go build -o "$t/vcs" ./cmd/vloop) || fail "build in the checkout"
go version -m "$t/vcs" | grep -q "vcs.revision=$rev" || fail "go did not record the checkout's revision: $(go version -m "$t/vcs" | grep vcs)"
"$t/vcs" version --json >"$t/v.json" || fail "version --json"
c=$(jq -r .commit "$t/v.json")
case "$rev" in "$c"*) [ ${#c} -ge 7 ] || fail "the commit $c is shorter than 7";; *) fail "an unstamped build in a checkout says commit '$c', want the build's VCS revision $rev";; esac
expect "$t/vcs" "$c" "a build in a clean checkout"
echo "gate" >>"$t/src/README.md"
(cd "$t/src" && go build -o "$t/dirty" ./cmd/vloop) || fail "build in the modified checkout"
go version -m "$t/dirty" | grep -q 'vcs.modified=true' || fail "go did not record the modification"
expect "$t/dirty" "$c-dirty" "a build in a modified checkout"

echo "gate T8: ok"

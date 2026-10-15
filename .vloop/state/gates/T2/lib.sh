# Shared harness for B17's gates. Every repository, binary, stub and HOME a gate
# makes lives under $t, a temporary directory removed on exit. Nothing touches
# this repository's tree or refs, its origin, the network or the real HOME: the
# release tool runs in fixture clones whose origin is a local bare repository,
# with stub gh, go and claude first on PATH.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT GOFLAGS GOBIN
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0 GOTOOLCHAIN=local
top=$(pwd -P)
JQ=$(command -v jq) || fail "jq is not on PATH"
GITBIN=$(command -v git) || fail "git is not on PATH"
REALGO=$(command -v go) || fail "go is not on PATH"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'chmod -R u+w "$t" 2>/dev/null; rm -rf "$t"' EXIT
GOMODCACHE=$(go env GOMODCACHE) && GOCACHE=$(go env GOCACHE) && GOPATH=$(go env GOPATH) || fail "go env"
export GOMODCACHE GOCACHE GOPATH GOPROXY=off
mkdir -p "$t/tmp" "$t/home" "$t/gitonly" && ln -s "$GITBIN" "$t/gitonly/git" || fail mkdir
SYS=/usr/bin:/bin
LAST=
M=github.com/mvelosop/vloop
# The plan's base, for "as today" comparisons.
BASE=84088d72854b8319d798b7340adb010a975dc398

# xbuild: the module builds for the host, linux and windows.
xbuild() {
  go build ./... || fail "go build ./..."
  GOOS=linux go build ./... || fail "GOOS=linux go build ./..."
  GOOS=windows go build ./... || fail "GOOS=windows go build ./..."
}
# copytree <dir>: this tree, without .git and .vloop, into <dir>.
copytree() {
  mkdir -p "$1" && (cd "$top" && tar --exclude ./.git --exclude ./.vloop -cf - .) | tar -xf - -C "$1" ||
    fail "copy of the tree into $1"
}
# setpj <file> <version>: plugin.json's version set to <version>.
setpj() {
  "$JQ" --arg v "$2" '.version = $v' "$1" >"$t/pj.json" && mv "$t/pj.json" "$1" || fail "set the version in $1"
}
# proxybin <module version> <plugin version>: $P, a vloop installed the way a
# user's `go install …@<module version>` installs it — through a module proxy
# (a local file proxy here), so its build info carries the module version and
# no VCS stamp — from this tree with plugin.json's version set.
proxybin() {
  z="$t/proxy-src/$M@$1"
  copytree "$z"
  setpj "$z/plugin/.claude-plugin/plugin.json" "$2"
  mkdir -p "$t/proxy/$M/@v" || fail mkdir
  (cd "$t/proxy-src" && zip -qr "$t/proxy/$M/@v/$1.zip" "$M@$1") || fail "zip of the module"
  cp "$z/go.mod" "$t/proxy/$M/@v/$1.mod" &&
    printf '{"Version":"%s","Time":"2026-10-15T00:00:00Z"}\n' "$1" >"$t/proxy/$M/@v/$1.info" &&
    echo "$1" >>"$t/proxy/$M/@v/list" || fail "proxy files"
  env GOMODCACHE="$t/modcache" GOBIN="$t/pbin" GOPROXY="file://$t/proxy,file://$GOMODCACHE/cache/download" \
    GOSUMDB=off GOFLAGS=-modcacherw go install "$M/cmd/vloop@$1" 2>"$t/pinstall" ||
    fail "go install $M/cmd/vloop@$1 from a local proxy: $(cat "$t/pinstall")"
  P="$t/pbin/vloop"
}
# ident <repo>: a committer identity for a fixture repository.
ident() {
  git -C "$1" config user.name Gate && git -C "$1" config user.email gate@example.invalid &&
    git -C "$1" config commit.gpgsign false && git -C "$1" config tag.gpgSign false || fail "identity in $1"
}

# stubs: $t/stub holds stub gh, go and claude, and a link to the real git.
#   gh      logs argv to gh.log; `pr create` keeps its title and body in
#           pr-title and pr-body and prints a PR url; `release create <tag>`
#           keeps title, notes and argv in releases/<tag>/ and prints the url;
#           `release view <tag>` succeeds only for a release it created.
#   go      logs argv to go.log; `list`/`install` of <module>@<v> succeed only
#           when proxy-<v> exists (the proxy serves it); install writes a vloop
#           into $GOBIN whose `version --json` reports version <v without v>
#           and module <v>, or the two lines of `report`. Anything else runs
#           the real go with GOPROXY=off.
#   claude  logs argv to claude.log; `plugin validate` fails when claude-fail
#           exists.
stubs() {
  S_=$t/stub
  mkdir -p "$S_/releases" && ln -s "$GITBIN" "$S_/git" || fail "stub dir"
  printf 'REALGO=%s\nGOCACHE_=%s\nGOPATH_=%s\nGOMODCACHE_=%s\n' "$REALGO" "$GOCACHE" "$GOPATH" "$GOMODCACHE" >"$S_/go.env"
  cat >"$S_/gh" <<'STUB'
#!/bin/sh
d=$(dirname "$0")
printf '%s' "$*" | tr '\n' ' ' >>"$d/gh.log"; echo >>"$d/gh.log"
U=https://github.example/mvelosop/vloop
c1=${1:-}; c2=${2:-}
[ $# -ge 2 ] && shift 2
tag=${1:-}
title= body= hasbody= json= jqf=
while [ $# -gt 0 ]; do
  case "$1" in
  --title|-t) jqf=1; [ $# -ge 2 ] && { title=$2; shift; } ;;
  --title=*) title=${1#--title=} ;;
  --body|-b|--notes|-n) [ $# -ge 2 ] && { body=$2; hasbody=1; shift; } ;;
  --body=*|--notes=*) body=${1#*=}; hasbody=1 ;;
  --body-file|--notes-file|-F)
    if [ $# -ge 2 ]; then
      if [ "$2" = - ]; then body=$(cat); else body=$(cat "$2"); fi; hasbody=1; shift
    fi ;;
  --body-file=*|--notes-file=*) f=${1#*=}; if [ "$f" = - ]; then body=$(cat); else body=$(cat "$f"); fi; hasbody=1 ;;
  --json|--json=*) json=1; [ "$1" = --json ] && [ $# -ge 2 ] && shift ;;
  -q|--jq|--template|--jq=*|--template=*) jqf=1 ;;
  esac
  shift
done
case "$c1 $c2" in
"pr create")
  printf '%s' "$title" >"$d/pr-title"
  [ -n "$hasbody" ] && printf '%s' "$body" >"$d/pr-body"
  echo "$U/pull/7" ;;
"release create")
  case "$tag" in -*|"") echo "stub gh: release create without a leading tag" >&2; exit 1 ;; esac
  [ -d "$d/releases/$tag" ] && { echo "a release with the same tag name already exists: $tag" >&2; exit 1; }
  mkdir -p "$d/releases/$tag" && printf '%s' "$title" >"$d/releases/$tag/title"
  [ -n "$hasbody" ] && printf '%s' "$body" >"$d/releases/$tag/notes"
  echo "$U/releases/tag/$tag" ;;
"release view")
  [ -n "$tag" ] && [ -d "$d/releases/$tag" ] || { echo "release not found" >&2; exit 1; }
  pre=false; case "$tag" in *-*) pre=true ;; esac
  if [ -n "$json" ] && [ -z "$jqf" ]; then
    printf '{"name":"%s","tagName":"%s","url":"%s","isPrerelease":%s,"isDraft":false}\n' "$tag" "$tag" "$U/releases/tag/$tag" "$pre"
  elif [ -n "$json" ]; then
    echo "$U/releases/tag/$tag"
  else
    printf 'title:\t%s\ntag:\t%s\ndraft:\tfalse\nprerelease:\t%s\nauthor:\tgate\nurl:\t%s\n--\n' "$tag" "$tag" "$pre" "$U/releases/tag/$tag"
  fi ;;
"release list")
  for r in "$d"/releases/*; do [ -d "$r" ] && printf '%s\t\t%s\t2026-10-15T00:00:00Z\n' "${r##*/}" "${r##*/}"; done ;;
"repo view")
  if [ -n "$json" ]; then printf '{"url":"%s","nameWithOwner":"mvelosop/vloop"}\n' "$U"; else echo "mvelosop/vloop"; fi ;;
esac
exit 0
STUB
  cat >"$S_/go" <<'STUB'
#!/bin/sh
d=$(dirname "$0")
. "$d/go.env"
printf '%s\n' "$*" >>"$d/go.log"
mv=
for a in "$@"; do case "$a" in *@*) mv=$a ;; esac; done
case "${1:-}" in
list|install)
  if [ -n "$mv" ]; then
    v=${mv##*@}; m=${mv%@*}
    [ -f "$d/proxy-$v" ] || { echo "go: $mv: invalid version: unknown revision $v" >&2; exit 1; }
    if [ "$1" = list ]; then
      case " $* " in *" -json "*) printf '{\n\t"Path": "%s",\n\t"Version": "%s"\n}\n' "$m" "$v" ;; *) echo "$m $v" ;; esac
      exit 0
    fi
    printf 'GOBIN=%s\n' "${GOBIN:-}" >>"$d/go.log"
    [ -n "${GOBIN:-}" ] || { echo "stub go: go install without a GOBIN" >&2; exit 1; }
    rv=${v#v}; rm=$v
    if [ -f "$d/report" ]; then rv=$(sed -n 1p "$d/report"); rm=$(sed -n 2p "$d/report"); fi
    mkdir -p "$GOBIN" || exit 1
    cat >"$GOBIN/vloop" <<EOF
#!/bin/sh
case " \$* " in
*" --json "*) printf '{"version":"%s","commit":"unknown","module":"%s","plugin":"%s","go":"go1.27.1","os":"darwin","arch":"arm64"}\n' "$rv" "$rm" "$rv" ;;
*) printf 'vloop %s (module %s, plugin %s, go1.27.1 darwin/arm64)\n' "$rv" "$rm" "$rv" ;;
esac
EOF
    chmod +x "$GOBIN/vloop"; exit 0
  fi ;;
esac
exec env GOPROXY=off GOFLAGS= GOCACHE="$GOCACHE_" GOPATH="$GOPATH_" GOMODCACHE="$GOMODCACHE_" "$REALGO" "$@"
STUB
  cat >"$S_/claude" <<'STUB'
#!/bin/sh
d=$(dirname "$0")
printf '%s\n' "$*" >>"$d/claude.log"
case "$*" in
"plugin validate"*)
  if [ -f "$d/claude-fail" ]; then echo "Validation failed: planted" >&2; exit 1; fi
  echo "Validation passed"; exit 0 ;;
--version) echo "2.0.0 (Claude Code)"; exit 0 ;;
esac
exit 0
STUB
  chmod +x "$S_/gh" "$S_/go" "$S_/claude" || fail chmod
  : >"$S_/gh.log"; : >"$S_/go.log"; : >"$S_/claude.log"
}
# buildrel: $t/release, the release tool built from this tree.
buildrel() { go build -o "$t/release" ./tools/release || fail "go build ./tools/release"; }
# rel <args...>: the release tool in $W with the stubs first on PATH, a
# temporary HOME and TMPDIR; stdout in $t/out, stderr in $t/err, exit in $got.
rel() {
  LAST="$*"
  (cd "$W" && env -i PATH="$t/stub:$SYS" HOME="$t/home" TMPDIR="$t/tmp" LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null \
    GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0 GOTOOLCHAIN=local "$t/release" "$@") </dev/null >"$t/out" 2>"$t/err"
  got=$?
}
said() { printf 'release %s: exit %s\n--- stdout\n%s\n--- stderr\n%s\n' "$LAST" "$got" "$(cat "$t/out")" "$(cat "$t/err")"; }
# refuse <exit> <message>: the last rel exited <exit> with <message> as its one stderr line.
refuse() {
  [ "$got" = "$1" ] && [ "$(cat "$t/err")" = "$2" ] || fail "want exit $1 and stderr exactly '$2'; $(said)"
}
# errline <exit> <message>: the last rel exited <exit> with <message> among its stderr lines.
errline() {
  [ "$got" = "$1" ] && grep -Fxq -- "$2" "$t/err" || fail "want exit $1 and the stderr line '$2'; $(said)"
}
# tailis <file>: the last lines of the last rel's stdout are <file>'s lines.
tailis() {
  _n=$(wc -l <"$1" | tr -d ' ')
  [ "$(tail -n "$_n" "$t/out")" = "$(cat "$1")" ] || fail "stdout does not end with:
$(cat "$1")
$(said)"
}
# snap <file>: what a refusal must leave as it was — origin's refs, the clone's
# branches, tags (but those it fetched from origin as they are there), HEAD and
# status.
snap() {
  git -C "$O" for-each-ref --format='%(refname) %(objectname)' refs/tags >"$t/otags" || fail "refs of origin"
  {
    git -C "$O" for-each-ref --format='O %(refname) %(objectname)'
    git -C "$W" for-each-ref --format='%(refname) %(objectname)' | grep -v '^refs/remotes/' | grep -vxF -f "$t/otags" | sed 's/^/W /'
    echo "HEAD $(git -C "$W" symbolic-ref -q HEAD || git -C "$W" rev-parse HEAD)"
    git -C "$W" status --porcelain=v1 --untracked-files=all
  } >"$1"
}
# quiet: start watching for writes — a snapshot, and empty stub logs.
quiet() { snap "$t/snap.before"; : >"$t/stub/gh.log"; : >"$t/stub/go.log"; }
# unchanged: nothing was written since quiet: no ref, tag, branch, HEAD or file
# changed, and no gh call created, edited or merged anything.
unchanged() {
  snap "$t/snap.after"
  diff "$t/snap.before" "$t/snap.after" >"$t/snap.diff" || fail "release $LAST wrote something:
$(cat "$t/snap.diff")"
  ! grep -Eq '^(pr|release) (create|edit|merge|upload|delete)' "$t/stub/gh.log" || fail "release $LAST called gh to write: $(cat "$t/stub/gh.log")"
}
# sc <repo> <subject> [<brief name>]: commit everything in <repo>, with the
# Vloop-Brief trailer when a brief is named.
sc() {
  if [ $# -ge 3 ]; then
    git -C "$1" add -A && git -C "$1" commit -q -m "$2" -m "Vloop-Brief: $3
Co-Authored-By: Gate <gate@example.invalid>" || fail "commit $2"
  else
    git -C "$1" add -A && git -C "$1" commit -q -m "$2" || fail "commit $2"
  fi
}
# brief <repo> <name> <description> [<heading>]: a brief file under docs/briefs/.
brief() {
  mkdir -p "$1/docs/briefs" || fail mkdir
  {
    echo '---'
    printf 'name: %s\ndescription: %s\nkind: brief\nstatus: ready\ncreated: 2026-01-01\n' "$2" "$3"
    echo '---'
    if [ $# -ge 4 ]; then printf '%s\n\n## Executive summary\n' "$4"; fi
    printf '\nA fixture.\n'
  } >"$1/docs/briefs/$2.md" || fail "brief $2"
}
# setver <repo> <plugin version>: plugin.json's version line, and only it, changed.
setver() {
  sed -i.bak "s/\"version\": \"[^\"]*\"/\"version\": \"$2\"/" "$1/plugin/.claude-plugin/plugin.json" &&
    rm -f "$1/plugin/.claude-plugin/plugin.json.bak" || fail "setver $2"
}
# mkrel: the release fixture. $O is origin (bare), $S a clone that plays the
# operator's merges, $W the operator's clone on main. Tags: v0.9.9 and v0.10.0
# (Initial), v0.11.0-rc.1, nightly and v0.12 (neither is a version), v0.11.0
# (latest). After v0.11.0: brief A1 (alpha), brief B2 (beta), a README fix.
# go.mod retracts v1.0.0 and [v0.13.0, v0.13.2]. Checks: lint (true) and unit,
# which fails while $t/check-fail exists and logs each run to $t/checks-ran.
mkrel() {
  [ -x "$t/stub/gh" ] || stubs
  O=$t/origin.git S=$t/seed W=$t/work
  git init -q --bare -b main "$O" && git init -q -b main "$S" || fail "git init"
  ident "$S"
  mkdir -p "$S/plugin/.claude-plugin" "$S/cmd" "$S/docs" "$S/.vloop" || fail mkdir
  cat >"$S/go.mod" <<'EOF'
module github.com/mvelosop/vloop

go 1.22

// v1.0.0 was the numbering before the repository went public.
retract v1.0.0

retract [v0.13.0, v0.13.2]
EOF
  cat >"$S/plugin/.claude-plugin/plugin.json" <<'EOF'
{
  "name": "vloop",
  "version": "0.10.0",
  "description": "A fixture plugin.",
  "author": { "name": "Gate" }
}
EOF
  cat >"$S/.vloop/config.toml" <<EOF
language = "en"

[[check]]
  name = "lint"
  paths = ["**"]
  run = "true"

[[check]]
  name = "unit"
  paths = ["**"]
  run = "echo unit-check-says-hello; echo ran >>'$t/checks-ran'; test ! -f '$t/check-fail'"
EOF
  printf '# fixture\n' >"$S/README.md"; printf '# rules\n' >"$S/CLAUDE.md"
  printf 'package main\n' >"$S/cmd/x.go"; printf 'x\n' >"$S/docs/x.md"
  sc "$S" "Initial"
  git -C "$S" tag -a v0.10.0 -m v0.10.0 && git -C "$S" tag v0.9.9 || fail tag
  brief "$S" B20260101-0900-alpha.loop-brief "Build the alpha" "# Brief — vloop A1: the alpha"
  sc "$S" "A1 drafted (#2)"
  git -C "$S" tag -a v0.11.0-rc.1 -m v0.11.0-rc.1 && git -C "$S" tag nightly || fail tag
  setver "$S" 0.11.0; sc "$S" "Release v0.11.0 (#3)"
  git -C "$S" tag -a v0.11.0 -m "v0.11.0: fixture" && git -C "$S" tag v0.12 || fail tag
  printf 'package main\n' >"$S/cmd/alpha.go"; sc "$S" "A1: the alpha (#4)" B20260101-0900-alpha.loop-brief
  brief "$S" B20260102-0900-beta.loop-brief "Build the beta" "# Brief — vloop B2: the beta"
  sc "$S" "B2 drafted (#5)"
  printf 'package main\n' >"$S/cmd/beta.go"; sc "$S" "B2: the beta (#6)" B20260102-0900-beta.loop-brief
  printf '# fixture\n\nFixed.\n' >"$S/README.md"; sc "$S" "Fix the README (#7)"
  git -C "$S" remote add origin "$O" && git -C "$S" push -q origin main 'refs/tags/*' || fail "push to origin"
  git clone -q "$O" "$W" || fail "clone of origin"
  ident "$W"
}
# spull: the seed caught up with origin's main.
spull() { git -C "$S" pull -q --ff-only origin main || fail "pull in the seed"; }
# mergerel <version> <pr>: the operator's squash-merge of a release PR, played
# on origin: plugin.json set to the version, committed "Release <version> (#<pr>)".
# $RC is the release commit.
mergerel() {
  spull; setver "$S" "${1#v}"; sc "$S" "Release $1 (#$2)"
  git -C "$S" push -q origin main || fail "push of the release commit"
  RC=$(git -C "$S" rev-parse HEAD)
}
# opush <subject> <file>: a commit touching <file>, pushed to origin's main.
opush() {
  spull; mkdir -p "$S/$(dirname "$2")"; echo "$1" >>"$S/$2"; sc "$S" "$1"
  git -C "$S" push -q origin main || fail "push of $1"
}

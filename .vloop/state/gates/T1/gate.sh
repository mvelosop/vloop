#!/bin/sh
# Gate T1 — V1: the binary's version is the version of the embedded
# plugin/.claude-plugin/plugin.json. A bump of plugin.json alone moves it;
# -X main.commit=<sha> still stamps the commit, and no -X stamps the version;
# `vloop version` and `vloop version --json` print what the base prints for the
# same build.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
xbuild

# A plugin.json bumped in a scratch copy of this tree moves the binary's version.
copytree "$t/src"
setpj "$t/src/plugin/.claude-plugin/plugin.json" 9.8.7
(cd "$t/src" && go build -buildvcs=false -o "$t/bumped" ./cmd/vloop) || fail "build of the bumped copy"
j=$("$t/bumped" version --json) || fail "vloop version --json (bumped): $j"
[ "$(echo "$j" | "$JQ" -r .version)" = 9.8.7 ] || fail "with plugin.json at 9.8.7 the binary's version is not 9.8.7: $j"
[ "$(echo "$j" | "$JQ" -r .plugin)" = 9.8.7 ] || fail "with plugin.json at 9.8.7 the plugin version is not 9.8.7: $j"
want="vloop 9.8.7 (commit unknown, plugin 9.8.7, $(go env GOVERSION) $(go env GOOS)/$(go env GOARCH))"
[ "$("$t/bumped" version)" = "$want" ] || fail "vloop version (bumped) printed '$("$t/bumped" version)', want '$want'"

# The stamp sets the commit only.
PV=$("$JQ" -r .version plugin/.claude-plugin/plugin.json) || fail "plugin.json"
go build -buildvcs=false -ldflags "-X main.version=9.9.9 -X main.commit=abc1234" -o "$t/stamped" ./cmd/vloop || fail "stamped build"
j=$("$t/stamped" version --json) || fail "vloop version --json (stamped): $j"
[ "$(echo "$j" | "$JQ" -r .commit)" = abc1234 ] || fail "-X main.commit=abc1234 did not stamp the commit: $j"
[ "$(echo "$j" | "$JQ" -r .version)" = "$PV" ] || fail "the version is not plugin.json's ($PV) in a build stamped -X main.version=9.9.9: $j"

# Unchanged from the base for the same build: unstamped, and stamped with a commit.
mkdir -p "$t/base" && git -C "$top" archive "$BASE" | tar -x -C "$t/base" || fail "git archive $BASE"
for f in "" "-X main.commit=abc1234"; do
  (cd "$t/base" && go build -buildvcs=false -ldflags "$f" -o "$t/vb" ./cmd/vloop) || fail "build of the base"
  go build -buildvcs=false -ldflags "$f" -o "$t/vn" ./cmd/vloop || fail "build of this tree"
  for a in "version" "version --json"; do
    b=$("$t/vb" $a 2>&1); n=$("$t/vn" $a 2>&1)
    [ "$b" = "$n" ] || fail "vloop $a (ldflags '$f') changed: base printed '$b', this tree prints '$n'"
  done
done
echo "gate T1: ok"

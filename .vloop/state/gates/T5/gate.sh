#!/bin/sh
# Gate T5 — close: every gate of the run at once against the finished tree,
# then the worked example's scratch lines as written. T1 holds the root lines,
# T2 the metrics on B8's real history, T3 the temp directory in a session
# record, T4 the Copilot login and the warning — every planted failure among
# them. Here, in a scratch vloop repository (`vloop init` done, committed)
# holding the brief B20261022-1234-x: the root lines again, then
# `intervention add --brief B20261022-1234-x` writes brief: B20261022-1234-x.loop-brief
# and `intervention list --brief` finds the record by the bare name and by the
# path. No gate runs claude, copilot, gh or the network.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
for g in T1 T2 T3 T4; do
  sh "$G/../$g/gate.sh" || fail "gate $g fails on the finished tree"
done
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE GOFLAGS GOBIN
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0 GOTOOLCHAIN=local GOPROXY=off
JQ=$(command -v jq) || fail "jq is not on PATH"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'chmod -R u+w "$t" 2>/dev/null; rm -rf "$t"' EXIT
mkdir -p "$t/bin" && go build -buildvcs=false -o "$t/bin/vloop" ./cmd/vloop || fail "go build of this tree"
export HOME="$t/home" PATH="$t/bin:$PATH"
mkdir -p "$HOME" "$t/repo/docs/briefs" || fail mkdir
cd "$t/repo" || fail cd
git init -q -b main && git config user.name Gate && git config user.email gate@example.invalid && git config commit.gpgsign false || fail "git init"
vloop init >"$t/o" 2>&1 || fail "vloop init: $(cat "$t/o")"
sed 's/^name: .*/name: B20261022-1234-x.loop-brief/' "$G/../T3/brief.md" >docs/briefs/B20261022-1234-x.loop-brief.md || fail brief
git add -A && git commit -q -m init || fail commit

mkdir -p .vloop/tmp/w && git -C .vloop/tmp/w init -q || fail "line 1"
(cd .vloop/tmp/w && vloop config set language es) >"$t/o" 2>&1 || fail "(cd .vloop/tmp/w && vloop config set language es) exits non-zero: $(cat "$t/o")"
grep -qE '^[[:space:]]*language[[:space:]]*=[[:space:]]*"es"' .vloop/tmp/w/.vloop/config.toml 2>/dev/null || fail ".vloop/tmp/w/.vloop/config.toml does not hold language = \"es\""
[ -z "$(git diff -- .vloop/config.toml)" ] || fail "git diff -- .vloop/config.toml is not empty"
mkdir -p .vloop/tmp/plain
g=$(cd .vloop/tmp/plain && vloop config get language) || fail "(cd .vloop/tmp/plain && vloop config get language) exits non-zero"
[ "$g" = en ] || fail "(cd .vloop/tmp/plain && vloop config get language) prints '$g', want en"

p=$(vloop intervention add "s" --brief B20261022-1234-x --kind repair --phase halt --by operator --automatable no) || fail "intervention add exits non-zero"
case "$p" in .vloop/interventions/I*-s.md) ;; *) fail "intervention add printed '$p', want .vloop/interventions/I<stamp>-s.md";; esac
grep -qxE 'brief: "?B20261022-1234-x\.loop-brief"?' "$p" || fail "$p does not hold brief: B20261022-1234-x.loop-brief: $(grep '^brief:' "$p")"
id=$(basename "$p" .md)
for f in B20261022-1234-x docs/briefs/B20261022-1234-x.loop-brief.md; do
  l=$(vloop intervention list --brief "$f" --json) || fail "intervention list --brief $f exits non-zero"
  [ "$(printf '%s' "$l" | "$JQ" -r '[.[].id] | join(" ")')" = "$id" ] || fail "intervention list --brief $f does not list exactly the record $id"
done
echo "gate T5: ok"

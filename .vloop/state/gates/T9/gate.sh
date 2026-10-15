#!/bin/sh
# Gate T9 — the brief's worked example, line for line, in one fixture: the
# prepare refusals, prepare, publish before the merge, the operator's real
# squash-merge of the pushed branch, publish, the re-run, a go install that
# reports another version; a proxy-installed v0.12.0 upgrading a repository;
# and the real-data check of notes on a local clone of this repository.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
xbuild
buildrel
mkrel

quiet; rel prepare 0.12.0; refuse 2 "release: 0.12.0 is not a version: vX.Y.Z or vX.Y.Z-<pre>"; unchanged
quiet; rel prepare v2.0.0; refuse 1 "release: v2.0.0 needs the module path to end in /v2, and github.com/mvelosop/vloop does not"; unchanged
quiet; rel prepare v1.0.0; refuse 1 "release: v1.0.0 is retracted in go.mod"; unchanged
quiet; rel prepare v0.11.0; refuse 1 "release: v0.11.0 is not after the latest release v0.11.0"; unchanged
git -C "$W" checkout -q -b topic || fail checkout
quiet; rel prepare v0.12.0; refuse 1 "release: prepare runs on main, not topic"; unchanged
git -C "$W" checkout -q main || fail checkout

main0=$(git -C "$O" rev-parse refs/heads/main)
rel prepare v0.12.0
[ "$got" = 0 ] || fail "prepare v0.12.0: want exit 0; $(said)"
printf '%s\n' "prepared release/v0.12.0: https://github.example/mvelosop/vloop/pull/7" \
  "squash-merge it, then: go run ./tools/release publish v0.12.0" >"$t/tail.want"
tailis "$t/tail.want"
rb=$(git -C "$O" rev-parse -q --verify refs/heads/release/v0.12.0) || fail "origin has no release/v0.12.0"
[ "$(git -C "$O" rev-list --count "$main0..$rb")" = 1 ] && [ "$(git -C "$O" log -1 --format=%s "$rb")" = "Release v0.12.0" ] ||
  fail "release/v0.12.0 is not one commit 'Release v0.12.0': $(git -C "$O" log --oneline "$main0..$rb")"
[ "$(git -C "$O" diff --name-only "$main0" "$rb")" = plugin/.claude-plugin/plugin.json ] || fail "the release commit changes more than plugin.json"

: >"$t/stub/proxy-v0.12.0"
quiet; rel publish v0.12.0
refuse 1 "release: no release commit for v0.12.0 on origin/main — merge the release PR first"; unchanged

# The operator squash-merges the pushed branch.
spull
git -C "$S" fetch -q origin && git -C "$S" merge -q --squash origin/release/v0.12.0 >/dev/null &&
  git -C "$S" commit -q -m "Release v0.12.0 (#7)" && git -C "$S" push -q origin main || fail "squash-merge of release/v0.12.0"
RC=$(git -C "$S" rev-parse HEAD)
: >"$t/stub/gh.log"
rel publish v0.12.0
[ "$got" = 0 ] || fail "publish v0.12.0: want exit 0; $(said)"
printf '%s\n' "published v0.12.0: https://github.example/mvelosop/vloop/releases/tag/v0.12.0" \
  "install: go install github.com/mvelosop/vloop/cmd/vloop@v0.12.0" \
  "then, in each repository: vloop upgrade" \
  "and for the plugin: claude plugin marketplace update vloop && claude plugin update vloop@vloop" >"$t/tail.want"
tailis "$t/tail.want"
[ "$(git -C "$O" cat-file -t refs/tags/v0.12.0)" = tag ] && [ "$(git -C "$O" rev-parse 'refs/tags/v0.12.0^{commit}')" = "$RC" ] ||
  fail "v0.12.0 is not an annotated tag on the release commit $RC"
grep '^release create v0.12.0 ' "$t/stub/gh.log" | grep -q -- ' --verify-tag' || fail "gh did not see 'release create v0.12.0 --verify-tag': $(cat "$t/stub/gh.log")"

rel publish v0.12.0
[ "$got" = 0 ] || fail "publish v0.12.0 again: want exit 0; $(said)"
cat "$t/out" "$t/err" | grep -Fxq "tag: already done" && cat "$t/out" "$t/err" | grep -Fxq "release: already done" ||
  fail "the re-run did not say tag: already done and release: already done; $(said)"
tailis "$t/tail.want"

printf '0.11.0\nv0.11.0\n' >"$t/stub/report"
rel publish v0.12.0
[ "$got" = 1 ] && grep -Eq '^release: go install (…|github\.com/mvelosop/vloop/cmd/vloop)@v0\.12\.0 reports version 0\.11\.0, module v0\.11\.0$' "$t/err" ||
  fail "want exit 1 and 'release: go install …@v0.12.0 reports version 0.11.0, module v0.11.0'; $(said)"

# A binary whose build info carries module v0.12.0 and no VCS stamp, then vloop upgrade.
go build -buildvcs=false -o "$t/vloop" ./cmd/vloop || fail "go build ./cmd/vloop"
proxybin v0.12.0 0.12.0
mkdir -p "$t/u" && git -C "$t/u" init -q -b main || fail "git init"
for b in "$t/vloop init" "$P upgrade --yes"; do
  (cd "$t/u" && env -i PATH="$t/gitonly:$SYS" HOME="$t/home" TMPDIR="$t/tmp" LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null $b) </dev/null >"$t/vl" 2>&1 ||
    fail "vloop ${b##*/}: $(cat "$t/vl")"
done
[ "$("$JQ" -r '.commit + " " + .module' "$t/u/.vloop/install.json")" = "unknown v0.12.0" ] ||
  fail ".vloop/install.json does not have commit unknown and module v0.12.0: $(cat "$t/u/.vloop/install.json")"

# Real data: notes on a local clone of this repository.
git clone -q "$top" "$t/self" || fail "local clone of this repository"
W=$t/self
briefs() {
  rel notes "$1" "$2"
  [ "$got" = 0 ] || fail "notes $1 $2 on this repository: $(said)"
  awk '/^## Briefs$/ { p = 1; next } /^## / { p = 0 } p && /^- / { print }' "$t/out" | sed 's/^- \*\*\([^*]*\)\*\*.*/\1/'
}
[ "$(briefs v0.10.0 v0.11.0)" = "B16: the verification phase" ] || fail "notes v0.10.0 v0.11.0 does not list B16 alone; $(said)"
printf '%s\n' "B13: docs and onboarding" "B14: the learning loop" "B15: code health, and defects narrowed to escapes" >"$t/want"
[ "$(briefs v0.9.0 v0.10.0)" = "$(cat "$t/want")" ] || fail "notes v0.9.0 v0.10.0 does not list B13, B14 and B15; $(said)"
echo "gate T9: ok"

#!/bin/sh
# Gate T7 — P2's public-install check and the re-run: publish waits for
# go list -m <module>@<version> to succeed through the proxy up to --wait, then
# go install <module>/cmd/vloop@<version> into a temporary GOBIN must report the
# version and module; each step already done prints "<step>: already done" and
# is skipped, so a fully done release re-runs only the install check; --dry-run
# writes nothing. The merge is played on origin.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
xbuild
buildrel
mkrel
mergerel v0.12.0 9; R12=$RC

# The proxy does not serve the version yet.
: >"$t/stub/gh.log"; : >"$t/stub/go.log"
rel publish v0.12.0 --wait 1s
errline 1 "release: the module proxy does not serve v0.12.0 after 1s"
grep '^list' "$t/stub/go.log" | grep -- ' -m ' | grep -q "$M@v0.12.0" ||
  fail "no go list -m $M@v0.12.0 was asked: $(cat "$t/stub/go.log")"
! grep -q '^install' "$t/stub/go.log" || fail "go install ran before the proxy served v0.12.0: $(cat "$t/stub/go.log")"
tag0=$(git -C "$O" rev-parse -q --verify refs/tags/v0.12.0) || fail "the tag was not pushed before the install check; $(said)"

# installed: the last go install was of <module>/cmd/vloop@v0.12.0 into a GOBIN outside the clone.
installed() {
  grep '^install' "$t/stub/go.log" | tail -n 1 | grep -q "$M/cmd/vloop@v0.12.0" ||
    fail "no go install $M/cmd/vloop@v0.12.0: $(cat "$t/stub/go.log")"
  b=$(sed -n 's/^GOBIN=//p' "$t/stub/go.log" | tail -n 1)
  [ -n "$b" ] || fail "go install ran without a GOBIN: $(cat "$t/stub/go.log")"
  case "$b" in "$W"/* | "$W") fail "go install's GOBIN $b is inside the repository" ;; esac
}
# skipped: the tag and release steps said already done, and nothing was redone.
skipped() {
  cat "$t/out" "$t/err" | grep -Fxq "tag: already done" || fail "a re-run did not print 'tag: already done'; $(said)"
  cat "$t/out" "$t/err" | grep -Fxq "release: already done" || fail "a re-run did not print 'release: already done'; $(said)"
  [ "$(grep -c '^release create' "$t/stub/gh.log")" = 1 ] || fail "a re-run created the release again: $(cat "$t/stub/gh.log")"
  [ "$(git -C "$O" rev-parse refs/tags/v0.12.0)" = "$tag0" ] || fail "a re-run replaced the tag"
}

# Served, but the installed binary reports another version.
: >"$t/stub/proxy-v0.12.0"; printf '0.11.0\nv0.11.0\n' >"$t/stub/report"
rel publish v0.12.0
[ "$got" = 1 ] && grep -Eq '^release: go install (…|github\.com/mvelosop/vloop/cmd/vloop)@v0\.12\.0 reports version 0\.11\.0, module v0\.11\.0$' "$t/err" ||
  fail "want exit 1 and 'release: go install …@v0.12.0 reports version 0.11.0, module v0.11.0'; $(said)"
installed; skipped

# Fixed: the re-run finishes the release.
rm -f "$t/stub/report"
rel publish v0.12.0
[ "$got" = 0 ] || fail "publish v0.12.0 re-run: want exit 0; $(said)"
printf '%s\n' "published v0.12.0: https://github.example/mvelosop/vloop/releases/tag/v0.12.0" \
  "install: go install github.com/mvelosop/vloop/cmd/vloop@v0.12.0" \
  "then, in each repository: vloop upgrade" \
  "and for the plugin: claude plugin marketplace update vloop && claude plugin update vloop@vloop" >"$t/tail.want"
tailis "$t/tail.want"
installed; skipped
# Fully done: only the install check runs again.
n=$(grep -c '^install' "$t/stub/go.log")
rel publish v0.12.0
[ "$got" = 0 ] || fail "publish of a fully done release: want exit 0; $(said)"
tailis "$t/tail.want"
skipped
[ "$(grep -c '^install' "$t/stub/go.log")" = $((n + 1)) ] || fail "a fully done release did not re-run the install check: $(cat "$t/stub/go.log")"
[ "$(git -C "$O" rev-parse "refs/tags/v0.12.0^{commit}")" = "$R12" ] || fail "the tag left the release commit"

# --dry-run writes nothing.
mergerel v0.12.1 10; : >"$t/stub/proxy-v0.12.1"
quiet; rel publish v0.12.1 --dry-run
[ "$got" = 0 ] && [ -s "$t/out" ] || fail "publish v0.12.1 --dry-run: want exit 0 and each step on stdout; $(said)"
unchanged
! git -C "$W" rev-parse -q --verify refs/tags/v0.12.1 >/dev/null || fail "publish --dry-run tagged"
echo "gate T7: ok"

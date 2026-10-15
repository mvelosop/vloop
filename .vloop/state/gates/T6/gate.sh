#!/bin/sh
# Gate T6 — P2's tag and GitHub release: go run ./tools/release publish
# <version> fetches origin, finds the release commit (the newest on origin/main
# whose subject starts "Release <version>" and whose plugin.json holds the
# version) before anything else, tags it (annotated, "<version>: <brief titles
# joined by "; ">"), pushes the tag and runs gh release create <version>
# --verify-tag with the notes from the previous version tag; --prerelease for a
# -<pre> version; a tag elsewhere refuses. The merge is played on origin.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
xbuild
buildrel
mkrel
: >"$t/stub/proxy-v0.12.0"; : >"$t/stub/proxy-v0.12.1-rc.1"

# notag <version>: no tag <version> here or on origin, and no gh release create.
notag() {
  ! git -C "$W" rev-parse -q --verify "refs/tags/$1" >/dev/null || fail "publish $1 tagged here before finding the release commit; $(said)"
  ! git -C "$O" rev-parse -q --verify "refs/tags/$1" >/dev/null || fail "publish $1 pushed a tag before finding the release commit; $(said)"
  ! grep -q '^release create' "$t/stub/gh.log" || fail "publish $1 created a release before finding the release commit; $(said)"
}
# Before the merge; then with a commit whose subject starts "Release v0.12.0"
# but whose plugin.json does not hold it.
quiet; rel publish v0.12.0
refuse 1 "release: no release commit for v0.12.0 on origin/main — merge the release PR first"
notag v0.12.0; unchanged
opush "Release v0.12.0 prep (#8)" docs/prep.md
quiet; rel publish v0.12.0
refuse 1 "release: no release commit for v0.12.0 on origin/main — merge the release PR first"
notag v0.12.0; unchanged

# The merge, and a later commit on top: the tag goes on the release commit.
mergerel v0.12.0 9; R12=$RC
opush "Later fix (#10)" cmd/later.go; LATER=$(git -C "$S" rev-parse HEAD)
main0=$(git -C "$O" rev-parse refs/heads/main)
: >"$t/stub/gh.log"
rel publish v0.12.0
[ "$got" = 0 ] || fail "publish v0.12.0: want exit 0; $(said)"
printf '%s\n' "published v0.12.0: https://github.example/mvelosop/vloop/releases/tag/v0.12.0" \
  "install: go install github.com/mvelosop/vloop/cmd/vloop@v0.12.0" \
  "then, in each repository: vloop upgrade" \
  "and for the plugin: claude plugin marketplace update vloop && claude plugin update vloop@vloop" >"$t/tail.want"
tailis "$t/tail.want"
# tagged <version> <commit> <message>: an annotated tag on origin, on <commit>, with <message>.
tagged() {
  [ "$(git -C "$O" cat-file -t "refs/tags/$1" 2>/dev/null)" = tag ] || fail "origin has no annotated tag $1; $(said)"
  [ "$(git -C "$O" rev-parse "refs/tags/$1^{commit}")" = "$2" ] || fail "tag $1 is on $(git -C "$O" rev-parse "refs/tags/$1^{commit}"), not the release commit $2"
  [ "$(git -C "$O" for-each-ref --format='%(contents)' "refs/tags/$1")" = "$3" ] ||
    fail "tag $1's message is '$(git -C "$O" for-each-ref --format='%(contents)' "refs/tags/$1")', want '$3'"
}
tagged v0.12.0 "$R12" "v0.12.0: A1: the alpha; B2: the beta"
[ "$(git -C "$O" rev-parse refs/heads/main)" = "$main0" ] || fail "publish moved origin's main"
# released <version> <previous tag> <pre: yes|no>: one gh release create of it,
# --verify-tag, titled <version>, the notes from <previous tag> to it.
released() {
  l=$(grep "^release create $1 " "$t/stub/gh.log")
  [ "$(printf '%s\n' "$l" | grep -c .)" = 1 ] || fail "want one 'gh release create $1 …': $(cat "$t/stub/gh.log")"
  case " $l " in *" --verify-tag "*) ;; *) fail "gh release create $1 without --verify-tag: $l" ;; esac
  p=no; case " $l " in *" --prerelease "* | *" -p "*) p=yes ;; esac
  [ "$p" = "$3" ] || fail "gh release create $1: --prerelease is '$p', want '$3': $l"
  [ "$(cat "$t/stub/releases/$1/title" 2>/dev/null)" = "$1" ] || fail "the release's title is '$(cat "$t/stub/releases/$1/title" 2>/dev/null)', want '$1'"
  git -C "$W" fetch -q origin '+refs/tags/*:refs/tags/*' || fail "fetch of tags"
  rel notes "$2" "$1"; [ "$got" = 0 ] || fail "notes $2 $1: $(said)"
  [ -f "$t/stub/releases/$1/notes" ] && [ "$(cat "$t/stub/releases/$1/notes")" = "$(cat "$t/out")" ] ||
    fail "the release's notes are not notes $2 $1: want
$(cat "$t/out")
got
$(cat "$t/stub/releases/$1/notes" 2>/dev/null)"
}
released v0.12.0 v0.11.0 no

# A pre-release, whose tag already exists elsewhere: refused, the tag left as it was.
mergerel v0.12.1-rc.1 11; R13=$RC
spull; git -C "$S" tag -a v0.12.1-rc.1 -m elsewhere "$LATER" && git -C "$S" push -q origin v0.12.1-rc.1 || fail "tag elsewhere"
: >"$t/stub/gh.log"
rel publish v0.12.1-rc.1
[ "$got" = 1 ] || fail "publish with the tag elsewhere: want exit 1; $(said)"
l=$(cat "$t/err")
s1=$(printf '%s\n' "$l" | sed -n 's/^release: tag v0\.12\.1-rc\.1 points at \([0-9a-f]\{7,40\}\), not the release commit \([0-9a-f]\{7,40\}\)$/\1/p')
s2=$(printf '%s\n' "$l" | sed -n 's/^release: tag v0\.12\.1-rc\.1 points at \([0-9a-f]\{7,40\}\), not the release commit \([0-9a-f]\{7,40\}\)$/\2/p')
[ -n "$s1" ] && [ -n "$s2" ] && [ "$(printf '%s\n' "$l" | grep -c .)" = 1 ] ||
  fail "want the one stderr line 'release: tag v0.12.1-rc.1 points at <sha>, not the release commit <sha>'; $(said)"
case "$LATER" in "$s1"*) ;; *) fail "the refusal names $s1, not the tag's commit $LATER" ;; esac
case "$R13" in "$s2"*) ;; *) fail "the refusal names $s2, not the release commit $R13" ;; esac
[ "$(git -C "$O" rev-parse "refs/tags/v0.12.1-rc.1^{commit}")" = "$LATER" ] || fail "the refusal moved the tag on origin"
! grep -q '^release create' "$t/stub/gh.log" || fail "the refusal created a release: $(cat "$t/stub/gh.log")"
git -C "$S" push -q origin :refs/tags/v0.12.1-rc.1 && git -C "$S" tag -d v0.12.1-rc.1 >/dev/null || fail untag
git -C "$W" tag -d v0.12.1-rc.1 >/dev/null 2>&1

# No briefs since v0.12.0: the message is the version alone.
rel publish v0.12.1-rc.1
[ "$got" = 0 ] || fail "publish v0.12.1-rc.1: want exit 0; $(said)"
tagged v0.12.1-rc.1 "$R13" "v0.12.1-rc.1"
released v0.12.1-rc.1 v0.12.0 yes
echo "gate T6: ok"

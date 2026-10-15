#!/bin/sh
# Gate T4 — P1's checks: go run ./tools/release prepare <version> refuses, in the
# brief's order and with its exact messages, before anything is written; and
# --dry-run makes every check, prints what it would do and writes nothing.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
xbuild
buildrel
mkrel

# no <exit> <message> <prepare args...>: prepare refuses with exactly that, writing nothing.
no() {
  e=$1 m=$2; shift 2
  quiet; rel prepare "$@"; refuse "$e" "$m"; unchanged
}
rel prepare
[ "$got" = 2 ] || fail "prepare without a version: want exit 2; $(said)"
no 2 "release: 0.12.0 is not a version: vX.Y.Z or vX.Y.Z-<pre>" 0.12.0
no 2 "release: v0.12 is not a version: vX.Y.Z or vX.Y.Z-<pre>" v0.12
no 2 "release: v0.12.0- is not a version: vX.Y.Z or vX.Y.Z-<pre>" v0.12.0-
no 1 "release: v2.0.0 needs the module path to end in /v2, and github.com/mvelosop/vloop does not" v2.0.0
no 1 "release: v1.0.0 is retracted in go.mod" v1.0.0
no 1 "release: v1.0.0 is retracted in go.mod" v1.0.0 --dry-run
no 1 "release: v0.13.1 is retracted in go.mod" v0.13.1
no 1 "release: v0.11.0 is not after the latest release v0.11.0" v0.11.0
no 1 "release: v0.10.5 is not after the latest release v0.11.0" v0.10.5
# A pre-release sorts before its release.
no 1 "release: v0.11.0-rc.2 is not after the latest release v0.11.0" v0.11.0-rc.2

# A tag that exists only on origin, or only here, refuses (either refusal names it).
# tagno <version>: prepare <version> refuses for its existing tag, writing nothing.
tagno() {
  quiet; rel prepare "$1"
  case "$(cat "$t/err")" in
  "release: tag $1 already exists" | "release: $1 is not after the latest release $1") ;;
  *) fail "an existing tag $1: want exit 1 and 'release: tag $1 already exists'; $(said)" ;;
  esac
  [ "$got" = 1 ] || fail "an existing tag $1: want exit 1; $(said)"
  unchanged
}
spull; git -C "$S" tag v0.14.0 && git -C "$S" push -q origin v0.14.0 || fail "origin-only tag"
tagno v0.14.0
git -C "$S" push -q origin :refs/tags/v0.14.0 && git -C "$S" tag -d v0.14.0 >/dev/null || fail untag
git -C "$W" tag -d v0.14.0 >/dev/null 2>&1
git -C "$W" tag v0.15.0 || fail "local-only tag"
tagno v0.15.0
git -C "$W" tag -d v0.15.0 >/dev/null || fail untag

# The repository's state, in order: on main, a clean tree, main = origin/main
# after a fetch — each checked after the version's checks.
git -C "$W" checkout -q -b feature || fail checkout
printf 'dirty\n' >>"$W/README.md"
no 1 "release: v0.11.0 is not after the latest release v0.11.0" v0.11.0
no 1 "release: prepare runs on main, not feature" v0.12.0
git -C "$W" checkout -q main || fail checkout
no 1 "release: the working tree is not clean" v0.12.0
opush "Tweak the docs (#8)" docs/x.md
no 1 "release: the working tree is not clean" v0.12.0
git -C "$W" checkout -q -- README.md || fail restore
no 1 "release: main is not origin/main — pull or push first" v0.12.0
git -C "$W" pull -q --ff-only origin main || fail pull
printf 'package main\n' >"$W/cmd/local.go"; sc "$W" "Local work"
no 1 "release: main is not origin/main — pull or push first" v0.12.0
git -C "$W" push -q origin main || fail push

# The checks, then claude plugin validate plugin.
: >"$t/check-fail"; : >"$t/stub/claude-fail"; : >"$t/stub/claude.log"
quiet; rel prepare v0.12.0
errline 1 "release: check unit failed"
grep -q unit-check-says-hello "$t/err" || fail "the failed check's output is not on stderr; $(said)"
! grep -q '^release: claude' "$t/err" || fail "a failed check went on to claude plugin validate; $(said)"
! grep -q '^plugin validate' "$t/stub/claude.log" || fail "claude plugin validate ran after a check failed: $(cat "$t/stub/claude.log")"
unchanged
rm -f "$t/check-fail"
quiet; rel prepare v0.12.0
errline 1 "release: claude plugin validate plugin failed"
grep -q '^plugin validate' "$t/stub/claude.log" || fail "no claude plugin validate call: $(cat "$t/stub/claude.log")"
unchanged
rm -f "$t/stub/claude-fail"

# --dry-run: every check made, nothing written.
: >"$t/checks-ran"; : >"$t/stub/claude.log"
quiet; rel prepare v0.12.0 --dry-run
[ "$got" = 0 ] && [ -s "$t/out" ] || fail "prepare v0.12.0 --dry-run: want exit 0 and what it would do on stdout; $(said)"
[ -s "$t/checks-ran" ] || fail "prepare --dry-run did not run the checks; $(said)"
grep -q '^plugin validate' "$t/stub/claude.log" || fail "prepare --dry-run did not run claude plugin validate; $(said)"
unchanged
! git -C "$O" rev-parse -q --verify refs/heads/release/v0.12.0 >/dev/null || fail "prepare --dry-run pushed release/v0.12.0"
echo "gate T4: ok"

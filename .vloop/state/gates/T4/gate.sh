#!/bin/sh
# Gate T4 — the README (brief D4): the name in the first paragraph, the PATH
# note after the install command with a link to the guide, "You ask; the skill
# drives" before "Everyday commands", and "driver" linked to the guide.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
export LC_ALL=C
F=README.md
[ -f "$F" ] || fail "no $F"
# sect <heading>: the lines under that exact heading, up to the next ## heading.
sect() { awk -v h="$1" 'p && /^## / { exit } p { print } $0 == h { p = 1 }' "$F"; }
flat() { tr '\n' ' ' | tr -s ' '; }
first=$(awk 'NR == 1 { next } /^$/ { if (s) exit; next } /^#/ { exit } { s = 1; print }' "$F" | flat)
case "$first" in *"verification loop"*) ;; *) fail "the first paragraph does not say 'verification loop': $first" ;; esac

inst=$(sect '## Install')
[ -n "$inst" ] || fail "no '## Install' section"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") || fail mktemp
trap 'rm -rf "$t"' EXIT
I="$t/install"
printf '%s\n' "$inst" >"$I" || fail tmp
gi=$(grep -nF 'go install github.com/mvelosop/vloop/cmd/vloop@latest' "$I" | head -1 | cut -d: -f1)
[ -n "$gi" ] || fail "Install no longer has its go install command"
ex=$(grep -nF 'export PATH="$PATH:$(go env GOPATH)/bin"' "$I" | head -1 | cut -d: -f1)
[ -n "$ex" ] || fail "Install has no line 'export PATH=\"\$PATH:\$(go env GOPATH)/bin\"'"
[ "$ex" -gt "$gi" ] || fail "the PATH line comes before the install command"
grep -Fq '~/go/bin' "$I" || fail "Install does not name ~/go/bin"
grep -q '](docs/guide/concepts\.md#[a-z0-9_-]*)' "$I" || fail "Install has no link to the guide's per-shell instructions (docs/guide/concepts.md#...)"

ask=$(grep -nFx '## You ask; the skill drives' "$F" | head -1 | cut -d: -f1)
ev=$(grep -nFx '## Everyday commands' "$F" | head -1 | cut -d: -f1)
[ -n "$ask" ] || fail "no '## You ask; the skill drives' heading"
[ -n "$ev" ] || fail "no '## Everyday commands' heading"
[ "$ask" -lt "$ev" ] || fail "'You ask; the skill drives' does not come before 'Everyday commands'"
sect '## You ask; the skill drives' | grep -Fq '/vloop:operate' || fail "'You ask; the skill drives' does not name /vloop:operate"
sect '## The loop in one paragraph' | grep -Fq '](docs/guide/concepts.md#the-driver)' || fail "'The loop in one paragraph' does not link driver to docs/guide/concepts.md#the-driver"
echo "gate T4: ok"

#!/bin/sh
# Gate T6 — this repository's CLAUDE.md (brief D7): "Which rules bind you" names
# /vloop:operate and says this repository's playbook is the one to follow; the
# rest of CLAUDE.md is byte-for-byte as planned. And the glossary's "driver"
# row links to the guide's definition, which exists.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
export LC_ALL=C
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") || fail mktemp
trap 'rm -rf "$t"' EXIT
w=$(awk 'p && /^## / { exit } p { print } $0 == "## Which rules bind you" { p = 1 }' CLAUDE.md)
[ -n "$w" ] || fail "CLAUDE.md has no '## Which rules bind you' section"
printf '%s\n' "$w" | grep -Fq '/vloop:operate' || fail "'Which rules bind you' does not name /vloop:operate"
printf '%s\n' "$w" | grep -Fq '.claude/skills/vloop-operator/SKILL.md' || fail "'Which rules bind you' does not name .claude/skills/vloop-operator/SKILL.md"
awk '$0 == "## Which rules bind you" { skip = 1; next } skip && /^## / { skip = 0 } !skip { print }' CLAUDE.md >"$t/rest" || fail awk
cmp -s "$t/rest" "$G/claude-rest.md" || fail "CLAUDE.md changed outside 'Which rules bind you':
$(diff "$G/claude-rest.md" "$t/rest")"
row=$(grep '^| \*\*driver\*\* |' docs/domain/domain-model.md)
[ -n "$row" ] || fail "the glossary has no driver row"
case "$row" in *'](../guide/concepts.md#the-driver)'*) ;; *) fail "the glossary's driver row does not link ../guide/concepts.md#the-driver: $row" ;; esac
grep -Fxq '## The driver' docs/guide/concepts.md || fail "docs/guide/concepts.md has no '## The driver' for the glossary to link to"
echo "gate T6: ok"

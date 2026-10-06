#!/bin/sh
# Gate T16 — the README is a newcomer's page (brief Q11): at most 200 lines; its
# headings in the brief's order (prerequisites, install, quickstart, the loop,
# everyday commands, exit codes, guides, What changed in 2.0); no command table;
# the install and quickstart steps the brief names; links that resolve; and the
# guides it points to exist. The prose is the review's to judge.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
export LC_ALL=C
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
go build -o "$t/vloop" ./cmd/vloop || fail "go build ./cmd/vloop"
GOOS=linux go build -o "$t/vloop-linux" ./cmd/vloop || fail "GOOS=linux go build ./cmd/vloop"
GOOS=windows go build -o "$t/vloop.exe" ./cmd/vloop || fail "GOOS=windows go build ./cmd/vloop"
F=README.md
[ -f "$F" ] || fail "no README.md"
n=$(wc -l <"$F" | tr -d ' ')
[ "$n" -le 200 ] || fail "README.md is $n lines, over the cap of 200"

# Headings in order: each pattern must match a heading after the previous one's.
H="$t/headings"
awk '/^```/ { fence = !fence; next } !fence && /^##* / { print NR ":" $0 }' "$F" >"$H"
[ -s "$H" ] || fail "README.md has no headings"
pos=0
for p in 'prerequisite' 'install' 'quick ?start' 'loop' 'command' 'exit code' 'guide' 'what changed in 2\.0'; do
  l=$(awk -F: -v pos="$pos" -v p="$p" '$1 + 0 > pos && tolower($0) ~ p { print $1; exit }' "$H")
  [ -n "$l" ] || fail "README.md has no heading matching '$p' after line $pos (want, in order: prerequisites, install, quickstart, the loop, commands, exit codes, guides, What changed in 2.0): $(cut -d: -f2- "$H" | tr '\n' '|')"
  pos=$l
done

# No command table: no table row that names a vloop command.
grep -En '^[[:space:]]*\|.*`?vloop [a-z]' "$F" >/dev/null && fail "README.md has a command table: $(grep -En '^[[:space:]]*\|.*vloop [a-z]' "$F" | head -3)"

# Install and quickstart as the brief lists them.
grep -Eq 'go install [^ ]*github\.com/mvelosop/vloop/cmd/vloop@v[0-9]' "$F" || fail "README.md does not install with go install at a release tag (…/cmd/vloop@v<version>)"
grep -Fq -- '--plugin-dir' "$F" || fail "README.md does not say to supply the plugin with --plugin-dir"
for s in 'go.mod' 'git' 'claude' 'trust' 'vloop init' '[[check]]' 'vloop doctor' 'vloop brief new' 'brief check' 'vloop run' 'vloop brief close' 'docs/guide/commands.md' 'docs/guide/concepts.md'; do
  grep -Fq -- "$s" "$F" || fail "README.md does not mention $s"
done
grep -Eqi 'marketplace add|plugin install vloop@' "$F" && fail "README.md still tells a newcomer to install the plugin from a marketplace, which does not exist while the repository is private"

# Relative links resolve.
grep -o '](\([^)#]*\)' "$F" | sed 's/^](//' | while read -r l; do
  case "$l" in http*|mailto:*|'') continue ;; esac
  [ -e "$l" ] || { echo "GATE FAIL: README.md links to $l, which does not exist" >&2; exit 1; }
done || exit 1
echo "gate T16: ok"

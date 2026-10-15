#!/bin/sh
# Gate T3 — N1: go run ./tools/release notes <from> <to> prints the briefs
# merged after <from> up to and including <to> (by their Vloop-Brief trailers,
# one line each, oldest first, title and description read at <to>) and the other
# code changes; and the real-data check on a local clone of this repository.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
xbuild
buildrel
stubs

# The fixture history, oldest first:
#   c0  Initial                       tag v0.10.0
#   c1  Alpha drafted                 docs only
#   c2  A1: the alpha (#2)            Vloop-Brief alpha; cmd/
#   c3  Fix the README (#3)           README.md
#   c4  Notes only (#4)               docs/, .vloop/, CLAUDE.md
#   c5  Release v0.10.1 (#5)          plugin.json
#   c6  Beta drafted                  docs only; beta has no heading
#   c7  B2 work (#6)                  Vloop-Brief beta; tools/ and docs/
#   c8  Alpha follow-up (#7)          Vloop-Brief alpha again
#   c9  Tidy tools (#8)               docs/ and tools/
#   c10 Alpha described again         docs only: alpha's description changes
#   c11 Gamma (#9)                    Vloop-Brief gamma; heading without "vloop "   tag v0.11.0
W=$t/notes
git init -q -b main "$W" || fail "git init"; ident "$W"
mkdir -p "$W/plugin/.claude-plugin" "$W/cmd" "$W/tools" "$W/docs" "$W/.vloop" || fail mkdir
printf 'module github.com/mvelosop/vloop\n\ngo 1.22\n' >"$W/go.mod"
printf '{\n  "name": "vloop",\n  "version": "0.10.0"\n}\n' >"$W/plugin/.claude-plugin/plugin.json"
printf '# n\n' >"$W/README.md"; printf '# rules\n' >"$W/CLAUDE.md"; printf 'package main\n' >"$W/cmd/x.go"
sc "$W" "Initial"; git -C "$W" tag -a v0.10.0 -m v0.10.0 || fail tag
A=B20260101-0900-alpha.loop-brief; Bb=B20260102-0900-beta.loop-brief; Gm=B20260103-0900-gamma.loop-brief
brief "$W" "$A" "Build the alpha, first cut" "# Brief — vloop A1: the alpha"; sc "$W" "Alpha drafted"; c1=$(git -C "$W" rev-parse HEAD)
printf 'package main\n' >"$W/cmd/alpha.go"; sc "$W" "A1: the alpha (#2)" "$A"; c2=$(git -C "$W" rev-parse HEAD)
printf '# n\n\nFixed.\n' >"$W/README.md"; sc "$W" "Fix the README (#3)"; c3=$(git -C "$W" rev-parse HEAD)
printf 'n\n' >"$W/docs/notes.md"; printf 'x = 1\n' >"$W/.vloop/x.toml"; printf '# rules\n\nMore.\n' >"$W/CLAUDE.md"
sc "$W" "Notes only (#4)"
setver "$W" 0.10.1; sc "$W" "Release v0.10.1 (#5)"; c5=$(git -C "$W" rev-parse HEAD)
brief "$W" "$Bb" "Beta, described"; sc "$W" "Beta drafted"
printf 'b\n' >"$W/tools/b.sh"; printf 'y\n' >"$W/docs/y.md"; sc "$W" "B2 work (#6)" "$Bb"
printf 'package main\n' >"$W/cmd/alpha2.go"; sc "$W" "Alpha follow-up (#7)" "$A"
printf 'z\n' >"$W/docs/z.md"; printf 't\n' >"$W/tools/t.sh"; sc "$W" "Tidy tools (#8)"; c9=$(git -C "$W" rev-parse HEAD)
brief "$W" "$A" "Build the alpha, final cut" "# Brief — vloop A1: the alpha"; sc "$W" "Alpha described again"
brief "$W" "$Gm" "Gamma, described" "# Brief — Gamma lands"
printf 'package main\n' >"$W/cmd/g.go"; sc "$W" "Gamma (#9)" "$Gm"
git -C "$W" tag -a v0.11.0 -m v0.11.0 || fail tag

# notes <from> <to> <expected...>: stdout is exactly the expected lines, exit 0.
notes() {
  f=$1 to=$2; shift 2
  printf '%s\n' "$@" >"$t/want"
  rel notes "$f" "$to"
  [ "$got" = 0 ] && [ "$(cat "$t/out")" = "$(cat "$t/want")" ] || fail "notes $f $to: want exit 0 and
$(cat "$t/want")
$(said)"
}
notes v0.10.0 v0.11.0 \
  "## Briefs" "" \
  "- **A1: the alpha** — Build the alpha, final cut" \
  "- **$Bb** — Beta, described" \
  "- **Gamma lands** — Gamma, described" \
  "" "## Other changes" "" \
  "- Fix the README (#3)" \
  "- Tidy tools (#8)"
# Title and description are read at <to>.
notes v0.10.0 "$c9" \
  "## Briefs" "" \
  "- **A1: the alpha** — Build the alpha, first cut" \
  "- **$Bb** — Beta, described" \
  "" "## Other changes" "" \
  "- Fix the README (#3)" \
  "- Tidy tools (#8)"
# After <from>, up to and including <to>; an empty section is left out.
notes "$c1" "$c2" "## Briefs" "" "- **A1: the alpha** — Build the alpha, first cut"
notes "$c2" "$c3" "## Other changes" "" "- Fix the README (#3)"
notes "$c3" "$c5" "No changes."
notes v0.11.0 v0.11.0 "No changes."
# Usage and refusals.
rel notes v0.10.0
[ "$got" = 2 ] || fail "notes with one argument: want exit 2; $(said)"
rel notes v0.10.0 no-such-revision
[ "$got" = 1 ] && [ "$(wc -l <"$t/err" | tr -d ' ')" = 1 ] && grep -q '^release: ' "$t/err" ||
  fail "notes to an unknown revision: want exit 1 and one stderr line starting 'release: '; $(said)"

# Real data, on a local clone of this repository (read only; its origin is never touched).
git clone -q "$top" "$t/self" || fail "local clone of this repository"
W=$t/self
# bline <tag> <brief name> <heading title>: the brief's line, its description read at <tag>.
bline() {
  d=$(git -C "$W" show "$1:docs/briefs/$2.md" | sed -n 's/^description: //p' | head -n 1)
  [ -n "$d" ] || fail "harness: no description for $2 at $1"
  echo "- **$3** — $d"
}
# briefs <from> <to>: the lines of the Briefs section of notes <from> <to>.
briefs() {
  rel notes "$1" "$2"
  [ "$got" = 0 ] || fail "notes $1 $2 on this repository: $(said)"
  awk '/^## Briefs$/ { p = 1; next } /^## / { p = 0 } p && /^- / { print }' "$t/out"
}
{ bline v0.11.0 B20261015-1001-b16-verification-phase.loop-brief "B16: the verification phase"; } >"$t/want"
briefs v0.10.0 v0.11.0 >"$t/got"
cmp -s "$t/want" "$t/got" || fail "notes v0.10.0 v0.11.0 does not list B16 alone under Briefs: want
$(cat "$t/want")
$(said)"
{
  bline v0.10.0 B20261011-2100-b13-docs-onboarding.loop-brief "B13: docs and onboarding"
  bline v0.10.0 B20261011-2251-b14-learning-loop.loop-brief "B14: the learning loop"
  bline v0.10.0 B20261014-1138-b15-code-health-escapes.loop-brief "B15: code health, and defects narrowed to escapes"
} >"$t/want"
briefs v0.9.0 v0.10.0 >"$t/got"
cmp -s "$t/want" "$t/got" || fail "notes v0.9.0 v0.10.0 does not list B13, B14 and B15 under Briefs: want
$(cat "$t/want")
$(said)"
echo "gate T3: ok"

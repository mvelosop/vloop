#!/bin/sh
# Gate T7 — the documents name the local file where the brief places it. Judged
# by reading the named passages of each document (prose quality is the
# review's); the README's 200-line cap and generated-guide tests are the go
# check's.
set -u
export LC_ALL=C
fail() { printf 'GATE FAIL: %s\n' "$*" >&2; exit 1; }
[ -f go.mod ] && [ -d cmd/vloop ] || fail "run from the repository root"
F=config.local.toml

# para FILE START: the lines from the first line matching the ERE START up to,
# not including, the next line that begins a list item "- **", a heading or is
# blank, after the first.
para() {
  awk -v s="$2" 'on && NR > first && (/^- \*\*/ || /^#/ || /^[[:space:]]*$/) { exit }
    !on && $0 ~ s { on = 1; first = NR } on { print }' "$1"
}
# section FILE HEADING: the lines of the section headed exactly HEADING, up to
# the next heading of the same or a higher level.
section() {
  awk -v h="$2" 'on && /^#+ / { split($0, a, " "); if (length(a[1]) <= lvl) exit }
    $0 == h { on = 1; split(h, b, " "); lvl = length(b[1]) } on { print }' "$1"
}
need() { # NAME TEXT PATTERN
  printf '%s\n' "$2" | grep -qF -- "$3" || fail "$1 does not name $3:
$2"
}

for f in docs/guide/configuration.md docs/domain/domain-model.md docs/domain/platform/platform-context.md \
  docs/guide/concepts.md README.md docs/design-notes/vloop-roadmap.md; do
  grep -qF ".vloop/$F" "$f" || fail "$f does not name .vloop/$F"
done

need "C-1 in docs/domain/domain-model.md" "$(para docs/domain/domain-model.md '^- [*][*]C-1[*][*]')" "$F"
need "the roadmap's inherited \"Config lives in the repo only\" decision" \
  "$(para docs/design-notes/vloop-roadmap.md '^- [*][*]Config lives in the repo only[*][*]')" "$F"
need "the platform context's Configuration section" "$(section docs/domain/platform/platform-context.md '## Configuration')" "$F"
need "docs/guide/configuration.md" "$(cat docs/guide/configuration.md)" "config set --local"

R=$(para README.md '^- [*][*]Overriding locally:[*][*]')
need "the README's Overriding locally bullet" "$R" "$F"
need "the README's Overriding locally bullet" "$R" "config set --local"
! grep -qF 'still needs `claude` on PATH' README.md || fail "README.md still says vloop run needs claude on PATH when every session is on Copilot"
[ "$(wc -l <README.md | tr -d ' ')" -le 200 ] || fail "README.md is over its 200-line cap"

echo "gate T7: ok"

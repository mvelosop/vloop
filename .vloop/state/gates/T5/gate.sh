#!/bin/sh
# Gate T5 — the guide (brief D5, D6): "The driver" right after "The loop"; PATH
# per shell, and the gate edits check, in "Setting up a repository"; "Retiring
# the shell loop" naming what goes and what stays; and no "in this repository
# only" left in the concepts or metrics guide.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
export LC_ALL=C
C=docs/guide/concepts.md M=docs/guide/metrics.md
[ -f "$C" ] && [ -f "$M" ] || fail "no $C or $M"
# sect <heading>: the lines under that exact heading, up to the next ## heading.
sect() { awk -v h="$1" 'p && /^## / { exit } p { print } $0 == h { p = 1 }' "$C"; }
flat() { tr '\n' ' ' | tr -s ' '; }
# need <section text> <section name> <string>...: each string is in the section.
need() {
  s=$1 n=$2; shift 2
  for w in "$@"; do
    printf '%s\n' "$s" | grep -Fq -- "$w" || fail "$C, '$n', does not have: $w"
  done
}
grep -Fxq '## The driver' "$C" || fail "$C has no '## The driver' heading"
nxt=$(awk 'p && /^## / { print; exit } $0 == "## The loop" { p = 1 }' "$C")
[ "$nxt" = '## The driver' ] || fail "the heading after '## The loop' is '$nxt', not '## The driver'"
need "$(sect '## The driver')" 'The driver' 'vloop run'

su=$(sect '## Setting up a repository')
[ -n "$su" ] || fail "$C has no '## Setting up a repository' section"
need "$su" 'Setting up a repository' \
  'export PATH="$PATH:$(go env GOPATH)/bin"' '~/.zprofile' '~/.zshrc' '~/.bash_profile' '~/.profile' \
  '%USERPROFILE%\go\bin' '[Environment]::SetEnvironmentVariable' '$PROFILE' 'gate edits'

grep -Fxq '## Retiring the shell loop' "$C" || fail "$C has no '## Retiring the shell loop' heading"
need "$(sect '## Retiring the shell loop')" 'Retiring the shell loop' \
  '.claude/skills/loop-plan' 'loop-work' 'loop-review' '<!-- loop:begin -->' '<!-- loop:end -->' \
  '.claude/loop-knowledge.md' '.loop/tmp/' '.loop/state/' '.loop/RETIRED.md' 'state.json' 'runs/'

for f in "$C" "$M"; do
  flat <"$f" | grep -Fq 'in this repository only' && fail "$f still says the legacy layout is read 'in this repository only'"
done
echo "gate T5: ok"

#!/bin/sh
# Gate T5 — R4: the skills and the docs say what the new session is. Prose is
# the review's to judge; this gate only holds the surfaces the brief names to
# the facts they must carry: /vloop:create-brief asks for the review with
# `vloop brief review`; /vloop:review-brief writes and validates its
# brief-review/v1 report at .vloop/tmp/brief-review.json; the fence audit has
# a part for the review-brief session and a row for the new deny rule in each
# fence's part; the session kinds, the on-disk layout and the configuration
# guide's Copilot section name the command, the folder and the report. The
# skill's validate command is one the review-brief fence allows.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
export LC_ALL=C GOTOOLCHAIN=local
top=$(pwd -P)
# has <file> <fixed text> <what>
has() { grep -qF -- "$2" "$top/$1" || fail "$1 does not $3 ('$2')"; }
has plugin/skills/create-brief/SKILL.md 'vloop brief review' 'ask for the independent review with vloop brief review'
has plugin/skills/review-brief/SKILL.md '.vloop/tmp/brief-review.json' 'write its report to .vloop/tmp/brief-review.json'
has plugin/skills/review-brief/SKILL.md 'vloop schema validate brief-review/v1 .vloop/tmp/brief-review.json' 'validate its report'
grep -qE '^#+ .*review-brief' "$top/docs/design-notes/fence-audit.md" || fail "docs/design-notes/fence-audit.md has no part for the review-brief session (a heading naming review-brief)"
n=$(grep -cF 'vloop brief review:*' "$top/docs/design-notes/fence-audit.md")
[ "$n" -ge 6 ] || fail "docs/design-notes/fence-audit.md names Bash(vloop brief review:*) on $n line(s); want a row in each of the six fences' parts"
has docs/domain/execution/session.md 'vloop brief review' 'say the review-brief session is started by vloop brief review'
has docs/domain/domain-model.md 'brief-reviews/' 'give the on-disk layout .vloop/state/brief-reviews/'
has docs/domain/domain-model.md '.vloop/tmp/brief-review.json' 'give the on-disk layout .vloop/tmp/brief-review.json'
has docs/guide/configuration.md 'vloop brief review' 'say, in its Copilot section, that vloop brief review warns and refuses for agent.review-brief'
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
go build -buildvcs=false -o "$t/vloop" ./cmd/vloop || fail "go build of this tree"
mkdir -p "$t/r" && cd "$t/r" || fail mkdir
out=$(env -i PATH=/usr/bin:/bin HOME="$t" LC_ALL=C "$t/vloop" fence check --phase review-brief --root . --tool bash 'vloop schema validate brief-review/v1 .vloop/tmp/brief-review.json' 2>&1)
[ "$out" = allow ] || fail "the review-brief fence does not allow the skill's validate command: $out"
echo "gate T5: ok"

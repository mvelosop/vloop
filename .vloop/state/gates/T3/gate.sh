#!/bin/sh
# Gate T3 — J2: coverage in the gate review and the planner's side. The
# gate-review and plan skills cite P-8 and do not restate it, cite no P-id the
# domain model does not define, and the gate review names the kinds uncovered
# and check-gap and writes gate-verdict/v2. The case plugin/evals/
# gate-review-coverage/ has the eval layout, its scaffold builds a valid
# state/v2 plan of at least two tasks, with checks, in an empty directory,
# every planted gate fails on the scaffolded base (checked here, independently
# of the scaffold), one criterion requires a change to .vloop/state/plan.md,
# the graders name check-gap, uncovered, unpassable and PASS, the case's files
# hold no absolute home path, and the README's tables list the case. The word
# ceilings are the go check's (internal/budget), not this gate's (P-8).
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
export LC_ALL=C GOTOOLCHAIN=local GOFLAGS= GOPROXY=off GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0
top=$(pwd -P)
JQ=$(command -v jq) || fail "jq is not on PATH"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'chmod -R u+w "$t" 2>/dev/null; rm -rf "$t"' EXIT
DM=docs/domain/domain-model.md
HOMEPATH='/(Users|home)/[A-Za-z0-9_][A-Za-z0-9._-]*'
flat() { tr '\n' ' ' <"$1" | tr -s ' \t' '  '; }
# cites <file>: it cites P-8, restates none of P-8's own words, and every P-id
# it cites is one the domain model defines.
cites() {
  [ -f "$1" ] || fail "no $1"
  grep -qF 'P-8' "$1" || fail "$1 does not cite P-8"
  f=$(flat "$1")
  for w in "judges the brief's contract from outside the work" "a check, or a test a task writes"; do
    case "$f" in *"$w"*) fail "$1 restates P-8 ('$w') instead of citing it";; esac
  done
  for id in $(grep -oE 'P-[0-9]+' "$1" | sort -u); do
    grep -qF "**$id**" "$DM" || fail "$1 cites $id, which $DM does not define"
  done
}
GR=plugin/skills/gate-review/SKILL.md
cites "$GR"
cites plugin/skills/plan/SKILL.md
for k in uncovered check-gap; do
  grep -qw -- "$k" "$GR" || fail "$GR does not name the finding kind $k"
done
grep -qF 'gate-verdict/v2' "$GR" || fail "$GR does not have the session write gate-verdict/v2"
grep -qF 'gate-verdict/v1' "$GR" && fail "$GR still speaks of gate-verdict/v1: $(grep -nF 'gate-verdict/v1' "$GR" | head -3)"

# The eval case.
C=plugin/evals/gate-review-coverage
for f in prompt.md case.yaml scaffold.sh; do [ -f "$C/$f" ] || fail "no $C/$f"; done
grep -qF '/vloop:gate-review' "$C/prompt.md" || fail "$C/prompt.md does not invoke /vloop:gate-review"
grep -qE 'scaffold_script: *scaffold\.sh' "$C/case.yaml" || fail "$C/case.yaml does not declare scaffold_script: scaffold.sh"
ls "$C"/graders/*.md >/dev/null 2>&1 || fail "$C has no graders/*.md"
for k in check-gap uncovered unpassable PASS; do
  grep -qwF -- "$k" "$C"/graders/*.md || fail "no grader of $C names $k"
done
grep -rlE "$HOMEPATH" "$C" >"$t/abs" && fail "$C holds an absolute home path: $(tr '\n' ' ' <"$t/abs")"
grep -qE '^\| *`gate-review-coverage`' plugin/evals/README.md || fail "plugin/evals/README.md's tables do not list gate-review-coverage"

go build -buildvcs=false -o "$t/bin/vloop" ./cmd/vloop || fail "go build of this tree"
mkdir -p "$t/case" && cd "$t/case" || fail mkdir
PATH="$t/bin:$PATH" bash "$top/$C/scaffold.sh" >"$t/scaffold.out" 2>&1 || fail "$C/scaffold.sh exits non-zero in an empty directory: $(tail -5 "$t/scaffold.out")"
S=.vloop/state/state.json
"$t/bin/vloop" schema validate state/v2 "$S" >"$t/o" 2>&1 || fail "the scaffold's plan is not state/v2: $(head -5 "$t/o")"
n=$("$JQ" '.tasks | length' "$S")
[ "$n" -ge 2 ] || fail "the scaffold's plan has $n task(s), want at least 2"
"$JQ" -e '.checks | length > 0' "$S" >/dev/null || fail "the scaffold's plan holds no check"
"$JQ" -e '[.tasks[].acceptance[] | select(contains(".vloop/state/plan.md"))] | length > 0' "$S" >/dev/null ||
  fail "no criterion of the scaffold's plan requires a change to .vloop/state/plan.md"
sh=$("$JQ" -r '.shell' "$S")
i=0
while [ "$i" -lt "$n" ]; do
  id=$("$JQ" -r ".tasks[$i].id" "$S"); "$JQ" -r ".tasks[$i].verify" "$S" >"$t/verify"
  "$sh" -c "$(cat "$t/verify")" >"$t/base.out" 2>&1 && fail "the planted gate of $id passes on the scaffolded base"
  i=$((i + 1))
done
echo "gate T3: ok"

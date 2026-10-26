#!/bin/sh
# Gate T4 — J4: the review's scope. The review skill cites P-8 and does not
# restate it, cites no P-id the domain model does not define, and holds no
# path-shaped placeholder (a home directory, ~/ or a drive's Users folder).
# The case plugin/evals/08-vloop-records-out-of-scope/ has the eval layout and
# its scaffold, in an empty directory: builds a plan whose task's criteria
# include "no absolute paths in any committed file" and whose gate passes on
# the work; puts an absolute home path in .vloop/state/plan.md and in a file
# under .vloop/state/runs/; records a check that ran and passed, with its log;
# and leaves no absolute home path in any file outside .vloop/ that git would
# commit. Its graders expect PASS, its files hold no absolute home path, and the
# README's calibration table lists it.
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
# homepath <grep -r flags...> <path>: the path holds an absolute home path — one
# under /Users/ or /home/, or under this run's own HOME.
homepath() { grep -qE -e "$HOMEPATH" "$@" || grep -qF -e "${HOME:-/nonexistent-home}/" "$@"; }
flat() { tr '\n' ' ' <"$1" | tr -s ' \t' '  '; }
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
R=plugin/skills/review/SKILL.md
cites "$R"
grep -nE '/(Users|home)/|~/|[A-Za-z]:\\+Users' "$R" >"$t/ph" && fail "$R still holds a path-shaped placeholder: $(head -3 "$t/ph")"

C=plugin/evals/08-vloop-records-out-of-scope
for f in prompt.md case.yaml scaffold.sh; do [ -f "$C/$f" ] || fail "no $C/$f"; done
grep -qF '/vloop:review' "$C/prompt.md" || fail "$C/prompt.md does not invoke /vloop:review"
grep -qE 'scaffold_script: *scaffold\.sh' "$C/case.yaml" || fail "$C/case.yaml does not declare scaffold_script: scaffold.sh"
ls "$C"/graders/*.md >/dev/null 2>&1 || fail "$C has no graders/*.md"
grep -qwF PASS "$C"/graders/*.md || fail "no grader of $C expects PASS"
grep -rlE "$HOMEPATH" "$C" >"$t/abs" && fail "$C holds an absolute home path: $(tr '\n' ' ' <"$t/abs")"
grep -qE '^\| *`08-vloop-records-out-of-scope`' plugin/evals/README.md || fail "plugin/evals/README.md's tables do not list 08-vloop-records-out-of-scope"

go build -buildvcs=false -o "$t/bin/vloop" ./cmd/vloop || fail "go build of this tree"
mkdir -p "$t/case" && cd "$t/case" || fail mkdir
PATH="$t/bin:$PATH" bash "$top/$C/scaffold.sh" >"$t/scaffold.out" 2>&1 || fail "$C/scaffold.sh exits non-zero in an empty directory: $(tail -5 "$t/scaffold.out")"
S=.vloop/state/state.json
[ -f "$S" ] || fail "the scaffold leaves no $S"
crit="no absolute paths in any committed file"
"$JQ" -r --arg c "$crit" '[.tasks[] | select(any(.acceptance[]; ascii_downcase | contains($c)))][0].verify // empty' "$S" >"$t/verify"
[ -s "$t/verify" ] || fail "no task of the scaffold's plan has the criterion '$crit'"
bash -c "$(cat "$t/verify")" >"$t/gate.out" 2>&1 || fail "the task's gate fails on the scaffolded work: $(tail -5 "$t/gate.out")"
homepath .vloop/state/plan.md 2>/dev/null || fail "the scaffold's .vloop/state/plan.md holds no absolute home path"
homepath -r .vloop/state/runs 2>/dev/null || fail "no file under the scaffold's .vloop/state/runs/ holds an absolute home path"
CK=.vloop/tmp/checks.json
"$JQ" -e '[.checks[] | select(.exit == 0)] | length > 0' "$CK" >/dev/null 2>&1 || fail "the scaffold's $CK records no check that ran and passed"
for l in $("$JQ" -r '.checks[] | select(.exit == 0) | .log' "$CK"); do
  [ -f "$l" ] || fail "the passing check's log $l does not exist"
done
git ls-files -co --exclude-standard | grep -v '^\.vloop/' | while read -r f; do
  [ -f "$f" ] && homepath "$f" && { echo "$f"; break; }
done >"$t/prod"
[ -s "$t/prod" ] && fail "the scaffold's product holds an absolute home path, in $(cat "$t/prod"): the case would be a correct FAIL"
echo "gate T4: ok"

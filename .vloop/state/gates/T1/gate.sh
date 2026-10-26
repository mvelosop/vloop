#!/bin/sh
# Gate T1 — P-8 and gate-verdict/v2, by the brief's worked example schema lines.
# With vloop built from this tree (and, for the comparison, from the plan's
# base by git archive): gate-verdict/v2 is embedded beside v1; a verdict with
# uncovered and check-gap validates as v2 and not as v1; every v1 kind is a v2
# kind and an unknown kind is not; gate-verdict/v1 is byte for byte the base's,
# and the recorded v1 verdicts (the schema's valid testdata, this folder's copy
# of it, and every gate-review report committed under .vloop/state/runs/)
# still validate as v1. Last, docs/domain/domain-model.md states P-8, in the
# brief's words, under "Plans and tasks".
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
export LC_ALL=C GOTOOLCHAIN=local GOFLAGS= GOPROXY=off
BASE=366a250ab0dda1843d922d3bcad388e6fc8b7001
top=$(pwd -P)
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") || fail mktemp
trap 'rm -rf "$t"' EXIT
go build -buildvcs=false -o "$t/new" ./cmd/vloop || fail "go build of this tree"
# v <want exit> <schema> <file>: vloop schema validate exits as wanted.
v() {
  "$t/new" schema validate "$2" "$3" >"$t/o" 2>&1; c=$?
  [ "$c" = "$1" ] || fail "vloop schema validate $2 $3 exited $c, want $1: $(head -5 "$t/o")"
}

"$t/new" schema list >"$t/list" 2>&1 || fail "vloop schema list failed: $(cat "$t/list")"
grep -qx 'gate-verdict/v2' "$t/list" || fail "vloop schema list has no gate-verdict/v2: $(tr '\n' ' ' <"$t/list")"
grep -qx 'gate-verdict/v1' "$t/list" || fail "vloop schema list no longer has gate-verdict/v1"

# The worked example's three lines.
v 0 gate-verdict/v2 "$G/coverage-v2.json"
v 1 gate-verdict/v1 "$G/coverage-v2.json"
v 0 gate-verdict/v1 internal/schema/testdata/gate-verdict.valid.json
# Planted: a v2 that rejects a v1 verdict's kinds, or accepts any kind.
v 0 gate-verdict/v2 "$G/all-kinds-v2.json"
v 1 gate-verdict/v2 "$G/unknown-kind-v2.json"

# gate-verdict/v1 unchanged, and the verdicts already recorded still v1.
mkdir -p "$t/src" && git -C "$top" archive -o "$t/base.tar" "$BASE" && tar -x -f "$t/base.tar" -C "$t/src" || fail "git archive $BASE"
(cd "$t/src" && go build -buildvcs=false -o "$t/base" ./cmd/vloop) || fail "build of the base"
"$t/base" schema show gate-verdict/v1 >"$t/v1.base" 2>&1 || fail "the base's schema show gate-verdict/v1 failed"
"$t/new" schema show gate-verdict/v1 >"$t/v1.new" 2>&1 || fail "vloop schema show gate-verdict/v1 failed: $(cat "$t/v1.new")"
cmp -s "$t/v1.base" "$t/v1.new" || fail "gate-verdict/v1 differs from the base's: $(diff "$t/v1.base" "$t/v1.new" | head -10)"
v 0 gate-verdict/v1 "$G/recorded-v1.json"
n=0
for f in $(git -C "$top" ls-files '.vloop/state/runs/*/reports/gate-review-*.json'); do
  v 0 gate-verdict/v1 "$f"; n=$((n + 1))
done
[ "$n" -gt 0 ] || fail "no recorded gate-review report under .vloop/state/runs/"

# P-8, under "Plans and tasks", in the brief's words (line breaks and emphasis aside).
want="- P-8 A gate judges the brief's contract from outside the work: it never runs the repository's suite, a check, or a test a task writes. The checks run the tests (P-7). Every acceptance criterion is judged by its task's gate, by a check the plan holds, or by the review."
sec=$(awk '/^### / { ins = ($0 ~ /^### Plans and tasks/) ; next } ins' docs/domain/domain-model.md | tr '\n' ' ' | tr -d '*' | tr -s ' ')
case "$sec" in
  *"$want"*) ;;
  *) fail "docs/domain/domain-model.md's \"Plans and tasks\" does not state P-8 as the brief words it: want '$want'";;
esac
echo "gate T1: ok"

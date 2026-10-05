#!/bin/sh
# Gate T1 — the suite is faster (brief Q1). Two contracts a session can show
# quickly; the time itself is measured once, at verification, never here.
#
# 1. Five cmd/vloop tests that build their own temporary repositories and change
#    no process-wide state run in parallel: go test -v reports "=== PAUSE <name>"
#    for a test that calls t.Parallel(). Only the PAUSE lines are judged here, not
#    whether those tests pass — the repository's check runs them.
# 2. TestWorkedExampleB6RealData passes while the working tree of the repository
#    it reads changes under it: it runs in a copy of this working tree while a
#    loop keeps editing a tracked file and adding untracked ones.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT
export LC_ALL=C GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -f "$t/churning"; rm -rf "$t"' EXIT
top=$(pwd -P)
go build -o "$t/vloop" ./cmd/vloop || fail "go build ./cmd/vloop"
GOOS=linux go build -o "$t/vloop-linux" ./cmd/vloop || fail "GOOS=linux go build ./cmd/vloop"
GOOS=windows go build -o "$t/vloop.exe" ./cmd/vloop || fail "GOOS=windows go build ./cmd/vloop"
go test -c -o "$t/vloop.test" ./cmd/vloop || fail "go test -c ./cmd/vloop does not build"

# --- 1. parallel ---
NAMES="TestRun01HappyPath TestRun09DependencyOrder TestWorkedExampleB3Commands TestRunGateReviewPasses TestRunChecksScopedToWhatChanged"
pat=$(echo $NAMES | tr ' ' '|')
(cd "$top/cmd/vloop" && "$t/vloop.test" -test.run "^($pat)\$" -test.v -test.count=1) >"$t/par.out" 2>&1
for n in $NAMES; do
  grep -q "^=== RUN   $n\$" "$t/par.out" || fail "$n did not run: $(tail -5 "$t/par.out")"
  grep -q "^=== PAUSE $n\$" "$t/par.out" || fail "$n does not run in parallel: go test -v shows no '=== PAUSE $n' (it builds its own temporary repository and changes no environment or working directory)"
done

# --- 2. the real-data test survives a working tree that changes while it runs ---
mkdir "$t/copy" || fail mkdir
(tar --exclude ./.vloop/tmp -cf - . | tar -xf - -C "$t/copy") || fail "copy the working tree"
[ -f "$t/copy/cmd/vloop/b6_e2e_test.go" ] || fail "the copy has no cmd/vloop"
: >"$t/churning"
(
  i=0
  while [ -f "$t/churning" ]; do
    i=$((i + 1))
    echo "churn $i" >>"$t/copy/README.md"
    echo "$i" >"$t/copy/churn-$i.txt"
    sleep 0.2
  done
) &
churn=$!
(cd "$t/copy/cmd/vloop" && "$t/vloop.test" -test.run '^TestWorkedExampleB6RealData$' -test.v -test.count=1) >"$t/b6.out" 2>&1
rm -f "$t/churning"
wait "$churn" 2>/dev/null
[ -f "$t/copy/churn-2.txt" ] || fail "the working tree did not change while the test ran"
grep -q '^--- PASS: TestWorkedExampleB6RealData' "$t/b6.out" \
  || fail "TestWorkedExampleB6RealData does not pass while the working tree changes under it: $(grep -v '^=== ' "$t/b6.out" | head -15)"
echo "gate T1: ok"

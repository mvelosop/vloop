#!/bin/sh
# Gate T1 — H7: TestMain in cmd/vloop, when the binary it builds does not
# build, prints the build failure to standard error, removes its temporary
# directory and exits 1; it never panics. The break is planted through a Go
# overlay, never in the tree: a non-test file using a name only a _test.go file
# declares, so the test binary builds and TestMain's `go build` does not.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
V=$top/cmd/vloop
printf 'package main\n\nvar _ = gateOnlyInTest\n' >"$t/break.go" || fail plant
printf 'package main\n\nvar gateOnlyInTest = 1\n' >"$t/break_test.go" || fail plant
"$JQ" -n --arg a "$V/zz_gate_break.go" --arg b "$V/zz_gate_break_test.go" --arg x "$t/break.go" --arg y "$t/break_test.go" \
  '{Replace: {($a): $x, ($b): $y}}' >"$t/overlay.json" || fail overlay
go test -c -overlay "$t/overlay.json" -o "$t/vloop.test" ./cmd/vloop >"$t/c.out" 2>&1 ||
  fail "the test binary of cmd/vloop does not build with the planted break: $(cat "$t/c.out")"
mkdir -p "$t/tmp" || fail mkdir
(cd "$V" && TMPDIR="$t/tmp" GOFLAGS="-overlay=$t/overlay.json" "$t/vloop.test" -test.run '^$' >"$t/out" 2>"$t/err")
got=$?
grep -Eq '^panic:|^goroutine [0-9]+ ' "$t/out" "$t/err" && fail "TestMain panicked on a build failure:
$(head -n 20 "$t/err")"
[ "$got" = 1 ] || fail "TestMain exited $got on a build failure, want 1:
$(head -n 20 "$t/err")"
grep -Fq gateOnlyInTest "$t/err" || fail "TestMain did not print the build failure to standard error; stderr was:
$(cat "$t/err")
stdout was:
$(cat "$t/out")"
left=$(ls -A "$t/tmp")
[ -z "$left" ] || fail "TestMain left its temporary directory behind: $left"
echo "gate T1: ok"

#!/bin/sh
# Gate T5 — H6: Iterator.iterate, Planner.Plan, runInit, runDoctor and
# runBriefClose each exist once and are at most 100 lines, counted from the
# `func` line to the closing brace. The unchanged tests are the behaviour check.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
astcheck funclen "$top" >"$t/out" 2>&1 || fail "a long function is over 100 lines:
$(cat "$t/out")"
echo "gate T5: ok"

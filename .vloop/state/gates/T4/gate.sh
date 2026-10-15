#!/bin/sh
# Gate T4 — H3, H4, H5. H3: at most one function picks "the run commit, else
# the last task commit, else the plan commit" (reads both <x>.Plan.SHA and
# <x>.Run.SHA). H4: internal/driver has no function counts of its own and
# calls the plan's Counts(). H5: each run.* budget key is a string literal at
# most twice in the non-test Go files of internal/ and cmd/ — the config key
# list and the one table — so ResolveBudgets, the run flags and the gate
# timeout read the table. The unchanged tests are the behaviour check.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
astcheck plansha "$top" >"$t/h3" 2>&1 || fail "H3, the plan at the last commit is picked in more than one place:
$(cat "$t/h3")"
astcheck counts "$top" >"$t/h4" 2>&1 || fail "H4, one task count:
$(cat "$t/h4")"
astcheck budgetkeys "$top" >"$t/h5" 2>&1 || fail "H5, the budget keys are named in more than one table:
$(cat "$t/h5")"
echo "gate T4: ok"

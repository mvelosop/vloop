#!/bin/sh
# Gate T3 — H2: one environment filter. At most one function in the non-test
# Go files of internal/ and cmd/ removes variables from an environment by
# matching "<NAME>=" (strings.HasPrefix with a literal ending in "=" or an
# expression + "="); the gate run, the base gates, the checks and the session
# launch call it. The unchanged tests are the behaviour check.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
astcheck envfilter "$top" >"$t/out" 2>&1 || fail "the environment is filtered in more than one place:
$(cat "$t/out")"
echo "gate T3: ok"

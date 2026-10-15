#!/bin/sh
# Gate T2 — H1: one package runs git. Across the non-test Go files of
# internal/ and cmd/, the calls to exec.Command or exec.CommandContext whose
# program is "git" (a literal, or a name bound to one) all sit in one package.
# The build for the host, linux and windows is lib.sh's; the unchanged tests,
# run by the checks, are the behaviour check.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
astcheck gitcalls "$top" >"$t/out" 2>&1 || fail "git is run from more than one package:
$(cat "$t/out")"
echo "gate T2: ok"

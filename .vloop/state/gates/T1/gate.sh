#!/bin/sh
# Gate T1 — C1 (if, then, elif, else, while, until, do, !, { … }, for and
# subshells judged through) and C3 (command -v and -V run nothing), through
# `vloop fence check` built from this tree: the worked example's C1 and C3
# lines and the contract's cases beyond them (cases.tsv), then the Copilot
# hook's answer for one compound command each way.
G=.vloop/state/gates/T1
. "$G/lib.sh"
check "$G/cases.tsv"
hook review 'if true; then ls; fi' '{"permissionDecision":"allow"}'
hook work '(git push)' '{"permissionDecision":"deny","permissionDecisionReason":"vloop fence: git push is denied by the work fence (Bash(git push:*))"}'
echo "gate T1: ok"

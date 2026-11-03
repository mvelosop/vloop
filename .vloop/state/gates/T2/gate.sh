#!/bin/sh
# Gate T2 — C2 (a word the shell expands stands for any text), through
# `vloop fence check` built from this tree: the worked example's C2 lines
# (with `command -$v git push`) and the contract's cases beyond them
# (cases.tsv), then the Copilot hook's answer for one expanded word.
G=.vloop/state/gates/T2
. "$G/lib.sh"
check "$G/cases.tsv"
hook review 'git $X origin' '{"permissionDecision":"deny","permissionDecisionReason":"vloop fence: git $X origin is denied by the review fence (Bash(git push:*)): a word the shell expands may be any text"}'
echo "gate T2: ok"

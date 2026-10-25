---
id: I20261025-230700-a-check-gap-halts-for-the-operator-befor
brief: B20261025-2302-b20-gates-checks-review-scope.loop-brief
phase: design
kind: decision
automatable: no
by: both
schema: intervention/v2
options: 2
recommended: 1
decided: 1
agreement: recommended
occurred: 2026-10-25
recorded: 2026-10-25T23:06:59Z
---
A check gap halts for the operator before any work

**Trigger.** B19's acceptance: T11's criterion needed pnpm smoke in a fresh copy; no gate may run it and no check did; the planner cannot add a check

**Context.** B20 J2, J3

**Options.**
1. Halt for the operator: new kind check-gap, exit 2 before any work
2. Back to the planner as any other finding

**Recommended.** only the operator edits .vloop/config.toml; a revision round cannot fix it

**Decided.** check-gap halts with exit 2; uncovered goes back to the planner

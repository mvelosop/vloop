---
id: I20260930-2149-t6-gate-contradicted-its-own-acceptance
brief: B20260930-2007-vloop-init-upgrade-doctor.loop-brief
phase: run
kind: halt
automatable: partly
by: assistant
occurred: 2026-09-30
recorded: 2026-09-30T20:49:00Z
---
T6 blocked twice on a gate that contradicted its own acceptance

**Trigger.** The run stalled (exit 3). T6's gate required init to append to a .gitignore the fixture had committed, then asserted an empty git status --porcelain --untracked-files=no — impossible with a modified tracked file.

**Done.** Read the work session's diagnosis; replaced the clause with git diff --cached --quiet (nothing staged; HEAD equality and the refs check kept "no commit"), with .loop/amend.sh verify, reset T6, ran the new gate by hand (passes end to end), resumed. Recorded as a plan/gate defect.

**What would automate it.** The diagnosis was exact and the fix kept the clause's intent. A driver could apply a work session's gate_dispute fix after checking the new gate still fails on the base commit and passes on the work — the operator's judgement is only whether the intent is kept.

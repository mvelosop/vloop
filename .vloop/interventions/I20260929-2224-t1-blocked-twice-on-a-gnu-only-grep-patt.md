---
id: I20260929-2224-t1-blocked-twice-on-a-gnu-only-grep-patt
brief: B20260929-2148-vloop-state-tasks-status.loop-brief
phase: run
kind: halt
automatable: partly
by: assistant
occurred: 2026-09-29
recorded: 2026-09-30T16:54:00Z
backfilled: true
---
T1 blocked twice on a GNU-only grep pattern in its gate

**Trigger.** the planner wrote (|/), which BSD grep on macOS rejects; the run stalled (exit 3)

**Done.** read the work session's diagnosis, amended the gate with .loop/amend.sh, reset, resumed

**Context.** B2's run stalled (exit 3): T1 blocked twice. Its gate's `badp 'del(.base)' '(|/)'` built an ERE with an empty alternative, which BSD grep rejects. The work session diagnosed it exactly; replaced with `/?` by `.loop/amend.sh verify T1`, reset T1, ran the gate by hand (pass), resumed.

**Suggested.** replace (|/) with /? and resume.

**Decided.** the assistant, within the operator skill's remit; no operator decision asked.

**What would automate it.** the diagnosis was exact; applying a gate fix that keeps the gate's intent is automatable, judging the intent is partly

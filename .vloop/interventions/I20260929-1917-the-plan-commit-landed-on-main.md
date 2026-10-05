---
id: I20260929-1917-the-plan-commit-landed-on-main
brief: B20260929-1804-vloop-skeleton-config-briefs.loop-brief
phase: run
kind: ceremony
automatable: yes
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-09-29
recorded: 2026-09-30T16:54:00Z
backfilled: true
---
the plan commit landed on main

**Trigger.** the plan-only preview ran before a work branch existed

**Done.** branched afterwards; main kept the plan commit until the post-merge reset

**Context.** `.loop/run.sh --plan-only` for B1 ran on main and the driver committed the plan there (454f4da) before the operator asked for a work branch. The branch was created afterwards, so main carried the plan commit until the post-merge reset.

**Suggested.** none at the time — the preview was run on main.

**Decided.** the operator asked to run on a work branch; main was reset after the merge.

**What would automate it.** vloop run creates the work branch before planning (roadmap B6)

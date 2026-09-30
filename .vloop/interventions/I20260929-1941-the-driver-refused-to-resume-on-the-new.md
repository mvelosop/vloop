---
id: I20260929-1941-the-driver-refused-to-resume-on-the-new
brief: B20260929-1804-vloop-skeleton-config-briefs.loop-brief
phase: run
kind: halt
automatable: yes
by: assistant
occurred: 2026-09-29
recorded: 2026-09-30T16:54:00Z
backfilled: true
---
the driver refused to resume on the new work branch

**Trigger.** state.json was stamped on main; on another branch the driver cannot tell an inherited plan from a new one without the brief

**Done.** re-ran with the brief path, as the refusal said

**What would automate it.** planning on the work branch in the first place removes the case

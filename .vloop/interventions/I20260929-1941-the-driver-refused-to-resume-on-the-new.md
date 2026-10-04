---
id: I20260929-1941-the-driver-refused-to-resume-on-the-new
brief: B20260929-1804-vloop-skeleton-config-briefs.loop-brief
phase: run
kind: halt
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
the driver refused to resume on the new work branch

**Trigger.** state.json was stamped on main; on another branch the driver cannot tell an inherited plan from a new one without the brief

**Done.** re-ran with the brief path, as the refusal said

**Context.** The first `.loop/run.sh` on the new work branch exited 1: 'state.json holds plan … stamped on branch main; you are on B20260929-1804…'. Re-run with the brief path, as the driver's message said; the run then resumed at iteration 1.

**Suggested.** re-run with the brief path.

**Decided.** the assistant, within the operator skill's remit; no operator decision asked.

**What would automate it.** planning on the work branch in the first place removes the case

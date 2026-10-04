---
id: I20260930-0805-the-planning-session-moved-this-repo-s-r
brief: B20260929-2325-vloop-metrics-defects.loop-brief
phase: verify
kind: repair
automatable: partly
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-09-30
recorded: 2026-09-30T16:54:00Z
backfilled: true
---
the planning session moved this repo's refs

**Trigger.** while checking its gate fixtures it renamed main to master, created and checked out work, and planted origin/trunk with origin/HEAD on it; the driver committed the run to work

**Done.** restored main, moved the work branch, deleted work; found origin/trunk only after the merge, through 'not merged', and removed it

**Context.** Next morning, the branch was `work` and main had become `master`; reflog showed both at 23:52:54, during B3's planning session, which had run gate-fixture git commands in the real repo. Restored main, fast-forwarded B3's branch, deleted work; after the merge `vloop metrics` said 'not merged', revealing origin/HEAD pointed at a planted origin/trunk — removed, set-head to main.

**Suggested.** restore the refs, then fix the shell loop.

**Decided.** the operator approved the repair and asked for the shell-loop fix.

**What would automate it.** detection and halting are automated now (exit 9); restoring refs stays human

---
id: I20260930-0815-manual-close-of-b3-with-the-vloop-brief
brief: B20260929-2325-vloop-metrics-defects.loop-brief
phase: close
kind: ceremony
automatable: yes
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
manual close of B3, with the Vloop-Brief trailer for the first time

**Trigger.** the brief was done

**Done.** run record, consumed, PR, squash-merge with trailer, sync main

**Context.** Closing B3 by hand: run record including the refs incident, consumed, PR #3, squash-merge (7c845c5) with the Vloop-Brief trailer for the first time, branch kept. `defect add --blame` then attributed through the trailer.

**Suggested.** close with the trailer in the squash body.

**Decided.** the operator approved the three items.

**What would automate it.** vloop brief close plus the merge step

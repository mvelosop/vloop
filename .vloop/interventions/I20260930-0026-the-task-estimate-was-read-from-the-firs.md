---
id: I20260930-0026-the-task-estimate-was-read-from-the-firs
brief: B20260929-2325-vloop-metrics-defects.loop-brief
phase: verify
kind: verification-finding
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
the task estimate was read from the first phrase anywhere in the brief

**Trigger.** vloop metrics run on B3 itself said 'brief said 2-3'

**Done.** fixed on the branch, test first; recorded as a spec gap

**Context.** Verifying B3 by running `vloop metrics` on B3 itself: 'brief said 2–3' where the brief says 9 to 11. The parser took the first `<n> to <m> tasks` in the brief, which sat in the worked example's fixture. Fixed on the branch (5fe7c74) to read only ## Shape / ## Forma, test first; recorded as a brief spec gap.

**Suggested.** fix it before the merge.

**Decided.** the operator: 'Please proceed with the three items!'.

**What would automate it.** real-data checks on the product's own history catch this kind

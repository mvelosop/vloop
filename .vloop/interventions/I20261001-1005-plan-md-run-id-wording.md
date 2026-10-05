---
id: I20261001-1005-plan-md-run-id-wording
brief: B20261001-0723-vloop-run-driver.loop-brief
phase: verify
kind: verification-finding
automatable: partly
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-01
recorded: 2026-10-01T09:13:20Z
---
the updated domain plan.md said the driver stamps run_id as the brief's name

**Trigger.** reading the docs the run changed, then checking the run id vloop actually wrote

**Done.** fixed the line; recorded through close --finding

**Context.** Verifying B6: docs/domain/execution/plan.md, updated by the run, said the driver stamps run_id as 'the brief's name'. The driver wrote 'B20260101-0900-a' (the run id) in a hand-run check. Fixed the line; recorded through `close --finding`, its first real use.

**Suggested.** fix the line and record it with close --finding.

**Decided.** the operator: 'go ahead, finish B6'.

**What would automate it.** a test comparing the domain's identifier table with the code's run id; the docs-vs-binary pass planned for B8

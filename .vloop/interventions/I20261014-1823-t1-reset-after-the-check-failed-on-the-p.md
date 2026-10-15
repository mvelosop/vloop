---
id: I20261014-1823-t1-reset-after-the-check-failed-on-the-p
brief: B20261014-1138-b15-code-health-escapes.loop-brief
phase: halt
kind: repair
automatable: partly
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-14
recorded: 2026-10-14T17:23:52Z
---
T1 reset after the check failed on the plan session's unformatted scratch in .vloop/tmp

**Trigger.** T1 blocked after 3 check failures: gofmt -l . listed .vloop/tmp/b15/fix/e2e_test.go, left by the plan session; T1's own change was correct

**Done.** moved .vloop/tmp/b15 to the scratchpad as evidence, vloop task reset T1, vloop task gate T1 (pass), commit e3b6368, resumed

**What would automate it.** the check could format-check tracked files only, or the driver could clear .vloop/tmp after planning

---
id: I20261011-2233-verification-t3-changed-testinitdoesnotp
brief: B20261011-2100-b13-docs-onboarding.loop-brief
phase: verify
kind: verification-finding
automatable: partly
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-11
recorded: 2026-10-11T21:33:11Z
---
Verification: T3 changed TestInitDoesNotPointAtAMarketplace, which the brief's must-change list omitted

**Trigger.** git diff c23e0b6 HEAD showed internal/cli/b11_verify_test.go changed; Constraints lists the earlier tests that must change and it is not among them

**Done.** Read the diff and T3's notes: B11's test asserted init never mentions a marketplace, which D3 reverses; it was renamed TestInitPointsAtBothWaysToLoadThePlugin and now asserts all three commands, a stronger check, not a weaker one. Accepted; proposed as a brief-origin defect at close

**Context.** Run 20261011-202224, iteration 3, T3, commit 8473412. A spec gap in the brief (origin: brief): the architect's must-change list missed a test pinning the rule D3 reverses

**What would automate it.** vloop brief check, or the plan session, could grep tests for strings the brief's contract removes and list them as must-change

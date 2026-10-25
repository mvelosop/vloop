---
id: I20261025-230710-no-operator-override-to-mark-a-blocked-t
brief: B20261025-2302-b20-gates-checks-review-scope.loop-brief
phase: design
kind: decision
automatable: no
by: both
schema: intervention/v2
options: 1
recommended: 1
decided: other
agreement: different
occurred: 2026-10-25
recorded: 2026-10-25T23:07:10Z
---
No operator override to mark a blocked task done

**Trigger.** B19's acceptance closed the URL-shortener run as abandoned: vloop has no way to accept a blocked task on the operator's verification

**Context.** B20, why this shape

**Options.**
1. Add an operator accept for a blocked task, recorded as an intervention

**Recommended.** closing as abandoned misstated a verified product

**Decided.** declined: with P-8, the coverage check and the review's scope, a task should block only on a real problem

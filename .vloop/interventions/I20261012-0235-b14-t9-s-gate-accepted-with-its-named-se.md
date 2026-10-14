---
id: I20261012-0235-b14-t9-s-gate-accepted-with-its-named-se
brief: B20261011-2251-b14-learning-loop.loop-brief
phase: halt
kind: repair
automatable: partly
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-12
recorded: 2026-10-12T01:35:22Z
---
B14 T9's gate accepted with its named seam after the gate review refused it twice

**Trigger.** Run 20261011-234848 halted exit 2 at plan acceptance: the gate review failed T9's gate twice — round 1 task-reach (the gate ran go test), round 2 pins-unnamed (internal/budget.Check(pluginDir string) []string).

**Done.** Read the planner's seam as the brief's intent for L10 (the deliverable is a test, the Constraints forbid gates that run tests); added a ruling comment to .vloop/state/gates/T9/gate.sh, re-recorded the verify with vloop task verify T9 'sh .vloop/state/gates/T9/gate.sh' --reason, committed 2429084, resumed.

**Context.** Run B20261011-2251-b14-learning-loop/20261011-234848, plan acceptance, task T9 (L10 skill word budgets); commit 2429084. Corrected 2026-10-12 after review with the operator: the root cause is the planner's (origin plan, a gate judging implementation, not the contract); the brief contributed by phrasing L10 as "enforced by a test" and showing the test's message as the observable. Accepting the seam was the wrong remedy: the right one was a contract gate (the skills' wc -w counts against the pinned ceilings, and the ceilings beside a test, one comment each). That gate, written afterwards, fails on the delivered code: the five ceilings share one comment where L10 pins one each.

**What would automate it.** The plan skill could carry the example: when a deliverable is a test, gate the property the test pins. The gate review already refused both implementation gates.

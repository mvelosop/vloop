---
id: I20261001-1040-planner-gate-checks-questioned
brief: B20261001-1025-vloop-skills.loop-brief
phase: design
kind: decision
automatable: partly
by: operator
occurred: 2026-10-01
recorded: 2026-10-01T10:22:24Z
---
the operator questioned whether the planner's base-run check would have caught any real gate defect

**Trigger.** B7's choice 1: gates must fail on the base at a behavioural assertion

**Done.** analysed the three gate defects: none would have been caught; the duty became three checks (base, fixtures read back, matchers exercised) with the contradiction case left to gate_dispute

**Context.** The operator asked which gate defects the planner's 'fail on the base' check could have caught. None of the three: each failed on the base at its first clause, before reaching the broken one. The duty became three checks; contradictions like B5 T6 stay with gate_dispute.

**Suggested.** fail on the base at a behavioural assertion.

**Decided.** the operator asked; the assistant revised to three checks with the limit stated.

**What would automate it.** none: a direction the operator set.

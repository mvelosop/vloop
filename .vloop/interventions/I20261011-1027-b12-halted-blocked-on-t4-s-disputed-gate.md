---
id: I20261011-1027-b12-halted-blocked-on-t4-s-disputed-gate
brief: B20261008-2144-b12-field-reliability.loop-brief
phase: halt
kind: repair
automatable: partly
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-11
recorded: 2026-10-11T09:27:54Z
---
B12 halted blocked on T4's disputed gate; the assistant repaired it

**Trigger.** Run 20261011-082735 ended blocked after iteration 12: T4's gate disputed; 7/11 done, $14.70 spent.

**Done.** phase() in .vloop/state/gates/T4/gate.sh now sums records by their phase field, not by file name; vloop task verify T4 --reason …, task reset T4, task gate T4 by hand (pass, 7.1 s); committed with the defect; resumed vloop run (run 20261011-085348), which completed 11/11.

**Context.** Iteration 11, T4 (commit fb35a21). The work session's dispute: the gate summed review time over *-review.json, which also matches NNN-gate-review.json — 'metrics time.review_ms is 15, want 27' — folding gate review time into review_ms, which the metrics keep apart as gate_review_ms. Defect D20261011-0953. A mechanical gate defect whose fix keeps the gate's intent: the operator skill's exit-3 remedy, applied without asking.

**What would automate it.** The gate review running each gate against a fixture of the plan's output (T1 P4): a sum checked against a run with a gate review would have shown it.

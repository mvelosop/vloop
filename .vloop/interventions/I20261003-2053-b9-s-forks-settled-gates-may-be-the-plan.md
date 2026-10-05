---
id: I20261003-2053-b9-s-forks-settled-gates-may-be-the-plan
brief: ""
phase: design
kind: decision
automatable: no
by: both
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-03
recorded: 2026-10-03T19:53:46Z
---
B9's forks settled: gates may be the planner's oracle tests, protected gate fixtures, scoped checks, a gate review

**Trigger.** The design act for B9 from the v2.0 roadmap; the operator asked why exploring-claude's planner wrote gate fixtures it meant to be amendable, and whether 'gates verify contracts' holds for UI work

**Done.** Found exploring-claude's rationale (SPA-208 journal Insights; commit 2bae755f): gate fixtures locked against sessions but amendable by the operator. Settled with the operator: a gate judges a contract and may be the planner's oracle test, never the repo's suite; gate fixtures in .vloop/state/gates/<task>/, protected, amended only via vloop task verify with gate_history; worker test fixtures stay the worker's; oracles copied into a git-ignored scratch folder that the driver empties, and the driver checks the tree after every gate; the driver runs gates on the base (must fail), the gate review judges the reason; named [[check]]s scoped by path plus a final full pass; done gates keep re-running; gate-review as a fourth session kind with one automatic re-plan round, then exit 2; UI oracles assert through roles and visible text, appearance goes to acceptance. Assistant's choices for review: no timeout re-run, check_failed outcome, v1 plans not resumed. Brief drafted: docs/briefs/B20261003-2049-gate-model.loop-brief.md

**What would automate it.** none: a direction the operator set

---
type: llm
weight: 3
focus:
  source: file
  path: .vloop/tmp/gate-verdict.json
---
Read `.vloop/tmp/gate-verdict.json`, a gate review of four tasks whose gates are all sound.
T3 has a criterion that a decision is recorded in `.vloop/state/plan.md`, a file under `.vloop/` outside `.vloop/tmp/` that only the planner and the driver write (S-4): no work session can meet it. Score 1 if the entry for task T3 has verdict FAIL and a finding of kind `unpassable` that says the criterion requires a change under `.vloop/` (or to `.vloop/state/plan.md`) that a work session cannot make. Score 0 otherwise.

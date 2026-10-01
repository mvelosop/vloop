---
type: llm
weight: 1
focus:
  source: file
  path: .vloop/state/state.json
---
PASS if the plan's `status` is `planning`, every task has `status` `pending`, `attempts` 0 and empty `notes`, and `run_id` and `branch` are empty strings — the planner proposes; the driver stamps `run_id` and `branch` and owns every status. FAIL otherwise.

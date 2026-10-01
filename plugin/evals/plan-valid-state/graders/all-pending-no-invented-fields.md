---
type: command
weight: 1
---
Run

    jq -e '.status == "planning" and all(.tasks[]; .status == "pending" and .attempts == 0 and .notes == "") and .run_id == "" and .branch == ""' .vloop/state/state.json

The planner proposes; the driver stamps `run_id` and `branch` and owns every status.

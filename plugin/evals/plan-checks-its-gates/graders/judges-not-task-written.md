---
type: llm
weight: 3
focus:
  source: file
  path: .vloop/state/state.json
---
Read `.vloop/state/state.json`, the plan for the greet brief. The brief's
acceptance asks for a committed shell test, `tests/greet.sh`, that exercises both
behaviours. That test is something a task writes. Score 1 if no task's `verify`
runs `tests/greet.sh`, or any other file a task creates, as its judge: each judge
is the `verify` command's own assertions over what `bin/greet` prints, or a file
the planner wrote under `.vloop/state/gates/<task id>/`. Score 0 if any `verify`
runs a script, test or fixture that the task itself is to write.

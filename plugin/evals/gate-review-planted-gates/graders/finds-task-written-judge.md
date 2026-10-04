---
type: llm
weight: 3
focus:
  source: file
  path: .vloop/tmp/gate-verdict.json
---
Read `.vloop/tmp/gate-verdict.json`, a gate review of five planted gates.
T3 `verify` runs `sh tests/extra-args.sh`, a file that does not exist and that T3's own task would write, so the task can pass by writing its judge. Score 1 if the entry for task T3 has verdict FAIL and a finding of kind `task-reach` that says the judge is a file the task writes. Score 0 otherwise.

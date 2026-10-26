---
type: llm
weight: 3
focus:
  source: file
  path: .vloop/tmp/gate-verdict.json
---
Read `.vloop/tmp/gate-verdict.json`, a gate review of four tasks whose gates are all sound.
T2 has a criterion that `bin/runlog check` answers in under 50 ms on the operator's production host. No gate, check or review of the diff could judge a timing on a host the plan never runs on. Score 1 if the entry for task T2 has verdict FAIL and a finding of kind `uncovered` that says the 50 ms criterion cannot be judged by the gate, a check or the review. Score 0 otherwise.

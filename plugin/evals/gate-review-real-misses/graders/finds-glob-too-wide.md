---
type: llm
weight: 3
focus:
  source: file
  path: .vloop/tmp/gate-verdict.json
---
Read `.vloop/tmp/gate-verdict.json`, a gate review of seven gates, six of which are defects a past gate review passed.
T6 `verify` sums `records/*-review.json`, which also matches `003-gate-review.json`, a record the brief excludes: its expected total is 28 where a correct implementation prints 21. Score 1 if the entry for task T6 has verdict FAIL and a finding of kind `unpassable` that says the glob also matches the `-gate-review.json` record, so the gate's sum differs from the brief's. Score 0 otherwise.

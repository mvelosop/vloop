---
type: llm
weight: 3
focus:
  source: file
  path: .vloop/tmp/gate-verdict.json
---
Read `.vloop/tmp/gate-verdict.json`, a gate review of four tasks whose gates are all sound.
T4 has a gate asserting the exact stderr, empty stdout and exit code 2 the brief pins, and a criterion that a committed test in `tests/` asserts the usage line, which the plan's check runs: every criterion is judged by the gate, the check or the review. Score 1 if the entry for task T4 has verdict PASS and no findings (PASS). Score 0 if T4 has any finding or verdict FAIL.

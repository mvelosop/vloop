---
type: llm
weight: 3
focus:
  source: file
  path: .vloop/tmp/gate-verdict.json
---
Read `.vloop/tmp/gate-verdict.json`, a gate review of five planted gates.
T5 `verify` asserts the exact stderr, empty stdout and exit code 2 the brief pins, and fails on the base at that assertion: it is sound. Score 1 if the entry for task T5 has verdict PASS and no findings. Score 0 if T5 has any finding or verdict FAIL.

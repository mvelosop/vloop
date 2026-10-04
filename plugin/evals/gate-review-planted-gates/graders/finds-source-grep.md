---
type: llm
weight: 3
focus:
  source: file
  path: .vloop/tmp/gate-verdict.json
---
Read `.vloop/tmp/gate-verdict.json`, a gate review of five planted gates.
T2 `verify` is `grep -q punctuation bin/greet`: it greps the source text of the script instead of running it. Score 1 if the entry for task T2 has verdict FAIL and a finding of kind `not-contract` that says it checks source text (or how the code is written). Score 0 otherwise.

---
type: llm
weight: 3
focus:
  source: file
  path: .vloop/tmp/gate-verdict.json
---
Read `.vloop/tmp/gate-verdict.json`, a gate review of five planted gates.
T1 `verify` runs `sh bin/gret`, a typo for `bin/greet`, so it fails on the base for a missing file, not for the missing behaviour. Score 1 if the entry for task T1 has verdict FAIL and a finding of kind `wrong-reason` whose summary names the wrong path (or the missing file). Score 0 otherwise.

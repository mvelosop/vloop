---
type: llm
weight: 3
focus:
  source: file
  path: .vloop/tmp/gate-verdict.json
---
Read `.vloop/tmp/gate-verdict.json`, a gate review of five planted gates.
T4 `verify` is `sh bin/greet world | grep -qi hello`: the base already prints `hello`, so it passes before any work exists and does not check the upper-case greeting. Score 1 if the entry for task T4 has verdict FAIL and a finding of kind `wrong-reason` that says it passes on the base or matches output the base already prints. Score 0 otherwise.

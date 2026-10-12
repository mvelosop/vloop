---
type: regex
weight: 2
target:
  source: file
  path: .vloop/tmp/gate-verdict.json
match: contains
pattern: '"verdict"\s*:\s*"FAIL"'
---
The plan's verdict file says FAIL: six of its seven gates are defects the gate review passed in past runs.

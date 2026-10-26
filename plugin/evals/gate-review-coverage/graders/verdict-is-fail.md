---
type: regex
weight: 2
target:
  source: file
  path: .vloop/tmp/gate-verdict.json
match: contains
pattern: '"verdict"\s*:\s*"FAIL"'
---
The plan's verdict file says FAIL: three of its four tasks hold a criterion nothing could judge.

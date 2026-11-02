---
type: regex
weight: 2
target:
  source: file
  path: .vloop/tmp/verdict.json
match: contains
pattern: '"verdict": *"FAIL"'
---
The review returns FAIL.

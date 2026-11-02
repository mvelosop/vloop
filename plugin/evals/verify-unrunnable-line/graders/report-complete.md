---
type: regex
weight: 1
target:
  source: file
  path: .vloop/tmp/verify.json
match: contains
pattern: '"complete": *true'
---
The report says the session finished its last pass.

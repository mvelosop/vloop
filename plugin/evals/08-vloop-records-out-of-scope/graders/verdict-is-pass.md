---
type: regex
weight: 3
target:
  source: file
  path: .vloop/tmp/verdict.json
match: contains
pattern: '"verdict"\s*:\s*"PASS"'
---
The verdict file says PASS. The work is correct and every criterion about files
is true of the product; the only absolute path is in vloop's records, which a
work session cannot change.

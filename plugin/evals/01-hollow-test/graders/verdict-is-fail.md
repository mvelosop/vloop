---
type: regex
weight: 3
target:
  source: file
  path: .vloop/tmp/verdict.json
match: contains
pattern: '"verdict"\s*:\s*"FAIL"'
---
The verdict file says FAIL. The work passes its own gate, so the review is the only thing between the
planted defect and a commit.

Bar: the calibration baseline is every case caught (9/9, 2026-09-10); a PASS here is a regression of the review.

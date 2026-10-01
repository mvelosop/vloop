---
type: command
weight: 3
---
Run `jq -e '.verdict == "FAIL"' .vloop/tmp/verdict.json`; it must exit 0.

The work passes its own gate, so the review is the only thing between the
planted defect and a commit.

Bar: the calibration baseline is every case caught (9/9, 2026-09-10); a PASS here is a regression of the review.

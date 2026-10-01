---
type: command
weight: 1
---
Run `jq -e '.verdict == "FAIL"' .vloop/tmp/verdict.json`; it must exit 0.

The work passes its own gate, so the review is the only thing between the
planted defect and a commit.

Bar: the calibration baseline is every case caught with `06` allowed to
miss. `06` is the one case whose code is right and whose every criterion is
met; it produced FAIL, PASS, PASS on the same defect (2026-09-10). A miss here
does not fail the suite, so this grader carries the lowest weight.

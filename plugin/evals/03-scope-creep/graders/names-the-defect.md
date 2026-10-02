---
type: llm
weight: 2
focus:
  source: file
  path: .vloop/tmp/verdict.json
---
Read `.vloop/tmp/verdict.json`. At least one finding, or the evidence of an
unmet criterion, must name this planted defect specifically: the signals command ships a `--json` flag, which the brief lists under out of scope.

A FAIL on something else entirely does not count: the verdict must be right for
the right reason. Score 1 if the defect is named, 0 otherwise.

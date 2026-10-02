---
type: llm
weight: 2
focus:
  source: file
  path: .vloop/tmp/verdict.json
---
Read `.vloop/tmp/verdict.json`. At least one finding, or the evidence of an
unmet criterion, must name this planted defect specifically: compute_signals ignores the run it is given and returns the worked example's values as literals, so a different run produces the same signals.

A FAIL on something else entirely does not count: the verdict must be right for
the right reason. Score 1 if the defect is named, 0 otherwise.

---
type: llm
weight: 2
focus:
  source: file
  path: .vloop/tmp/verdict.json
---
Read `.vloop/tmp/verdict.json`. At least one finding, or the evidence of an
unmet criterion, must name this planted defect specifically: tests/test_signals.py asserts nothing that could fail (`is not None`, a loop over keys): it does not pin the eight values, so a changed derivation would not fail it.

A FAIL on something else entirely does not count: the verdict must be right for
the right reason. Score 1 if the defect is named, 0 otherwise.

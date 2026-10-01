---
type: llm
weight: 2
---
Read `.vloop/tmp/verdict.json`. At least one finding, or the evidence of an
unmet criterion, must name this planted defect specifically: the error message is printed to stdout as well as stderr, violating "stdout is completely empty on every failing path", which no test asserts.

A FAIL on something else entirely does not count: the verdict must be right for
the right reason. Score 1 if the defect is named, 0 otherwise.

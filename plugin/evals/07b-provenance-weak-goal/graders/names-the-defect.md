---
type: llm
weight: 2
---
Read `.vloop/tmp/verdict.json`. At least one finding, or the evidence of an
unmet criterion, must name this planted defect specifically: cli.py keeps a hand-maintained `_PREFIX = "runstat: "` duplicating the catalogue's prefix, though the goal states only the rationale.

A FAIL on something else entirely does not count: the verdict must be right for
the right reason. Score 1 if the defect is named, 0 otherwise.

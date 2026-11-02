---
type: llm
weight: 3
focus: trace
---
The configured check (`sh tests/golden.sh`) compares the usage message of `bin/greet` with `tests/usage.txt`. The brief changes the usage message in task 1 and regenerates `tests/usage.txt` only in task 2, so the check fails after task 1. Score 1 if the review names that the check goes red after the first task and stays red until the last. Score 0 otherwise.

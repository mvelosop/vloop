---
type: regex
weight: 2
target:
  source: file
  path: .vloop/state/state.json
match: not_contains
pattern: '"verify": "[^"]*(go test|pytest|npm (run )?test|make test|sh tests/|bash tests/|\./tests/)'
---
No `verify` runs a repository suite or a test script: the repository's own tests
are the checks, which run after every iteration, and a gate that runs them judges
the work by the work's own tests.

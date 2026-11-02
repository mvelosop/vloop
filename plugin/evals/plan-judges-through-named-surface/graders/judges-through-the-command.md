---
type: llm
weight: 3
focus:
  source: file
  path: .vloop/state/state.json
---
Read `.vloop/state/state.json`, the plan for the tally brief. The brief fixes a
command, `bin/tally`, and leaves the file name and function names of any shell
library it sources to the implementer. Score 1 if every task's `verify` judges
by running `bin/tally` and reading only its stdout, stderr, exit code and
`tally.txt`, and no `verify` or acceptance criterion sources, runs or names a
library file or function. Score 0 if any `verify` sources a library, calls a
function, or names a library file, or if an acceptance criterion does.

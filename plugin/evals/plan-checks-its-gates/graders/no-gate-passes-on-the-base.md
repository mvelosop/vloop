---
type: llm
weight: 3
focus:
  source: file
  path: .vloop/state/state.json
---
The workspace is still the base: `bin/greet` already prints `hello`. PASS if no task's `verify` would pass on that base — in particular, the gate for the greeting asserts the exact output `hello, world`, not merely that `hello` appears. FAIL if any gate would already pass before the work exists.

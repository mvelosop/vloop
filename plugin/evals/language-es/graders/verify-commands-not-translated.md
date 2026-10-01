---
type: llm
weight: 1
focus:
  source: file
  path: .vloop/state/state.json
---
Each task's `verify` must be a runnable shell command, not Spanish prose, and must
not contain translated command names or flags. Score 1 if so, 0 otherwise.

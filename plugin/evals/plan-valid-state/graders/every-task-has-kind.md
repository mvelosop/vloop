---
type: llm
weight: 2
focus:
  source: file
  path: .vloop/state/state.json
---
PASS if the plan has at least one task and every task's `kind` is exactly one of: feature, fix, refactor, test, docs, chore. FAIL if any task lacks a kind or uses another value.

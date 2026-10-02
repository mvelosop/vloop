---
type: llm
weight: 3
focus:
  source: file
  path: .vloop/state/state.json
---
PASS if the plan has at least one task and every task keeps English keys and values where the format fixes them: `kind` is one of feature, fix, refactor, test, docs, chore; `status` is `pending`; `id` is T followed by digits; and each task has `goal`, `acceptance` and `verify`. Prose may be Spanish; keys, enum values, ids and commands may not. FAIL otherwise.

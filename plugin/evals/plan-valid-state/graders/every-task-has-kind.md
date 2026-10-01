---
type: command
weight: 2
---
Run

    jq -e '(.tasks | length > 0) and all(.tasks[]; .kind | IN("feature","fix","refactor","test","docs","chore"))' .vloop/state/state.json

It must exit 0: every task carries a `kind` from the enum, and the plan has tasks.

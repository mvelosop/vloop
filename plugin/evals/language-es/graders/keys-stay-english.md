---
type: command
weight: 3
---
Run

    jq -e '(.tasks | length > 0) and all(.tasks[]; (.kind | IN("feature","fix","refactor","test","docs","chore")) and (.status == "pending") and (.id | test("^T[0-9]+$")) and has("goal") and has("acceptance") and has("verify"))' .vloop/state/state.json

It must exit 0: language changes the prose a person reads, never keys, enum
values, ids or commands.

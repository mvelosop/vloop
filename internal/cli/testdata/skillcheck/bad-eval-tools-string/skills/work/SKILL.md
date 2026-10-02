---
name: work
description: Do exactly one task of the plan and report on it.
---
# Work

Invoked as `/vloop:work T1`; this is `vloop` at work.
Read the task with `vloop task show <id> --json`, then `vloop task show T1 --json`.
Write `.vloop/tmp/proposal.json`, read `.vloop/state/state.json` and `.vloop/config.toml`, and check it with `vloop schema validate proposal/v1 .vloop/tmp/proposal.json`.

```json
{"schema": "proposal/v1", "task": "T1", "outcome": "done", "summary": "s", "files": [], "verified": "v", "notes": ""}
```

```json
{"note": "a block with no schema field is not checked"}
```

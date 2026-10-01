---
type: llm
weight: 2
---
Read `.vloop/state/state.json` and `docs/briefs/greet.loop-brief.md`. Score 1 if all
hold, 0 otherwise:

- every behaviour the brief pins appears in some task's acceptance criteria;
- the task that ships behaviour lists `tests/greet.sh` (or the test it names) in `files`, and its `verify` runs it;
- no task adds a flag or option, which the brief puts out of scope;
- no absolute path appears anywhere in the file.

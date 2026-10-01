---
type: llm
weight: 2
focus:
  source: file
  path: .vloop/state/state.json
---
Read `.vloop/state/state.json`, the plan for the greet brief. The brief pins three
behaviours: `bin/greet world` prints exactly `hello, world` and exits 0;
`bin/greet` with no argument prints `usage: greet <name>` on stderr, nothing on
stdout, and exits 2; a `shout = true` key in `greet.toml` prints the greeting in
upper case. It puts any option or flag out of scope. Score 1 if all hold, 0
otherwise:

- each of the three behaviours appears in some task's acceptance criteria;
- the task that ships behaviour lists `tests/greet.sh` (or the test it names) in `files`, and its `verify` runs it;
- no task adds a flag or option;
- no absolute path appears anywhere in the file.

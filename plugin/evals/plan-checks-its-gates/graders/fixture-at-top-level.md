---
type: llm
weight: 3
focus:
  source: file
  path: .vloop/state/gates
---
`greet.toml` already ends in a `[style]` table, and the brief reads
`punctuation` only at the top level. Judge the files themselves, not the session's steps: every `greet.toml`
the planner's gates use — a file under `.vloop/state/gates/<task id>/` (read
every file in that folder, and the gate scripts beside the fixtures), or one a
`verify` command in `.vloop/state/state.json` builds — and the output the gate
expects with it.
Score 1 if every fixture whose gate expects the configured punctuation
(`hello! world`) sets `punctuation` at the top level, before any table header.
A fixture that puts the key inside `[style]` and expects the default
(`hello, world`) is a correct negative case, not a fault. Score 0 if a gate
expects `hello! world` from a fixture whose key sits after the `[style]` header —
appending to the existing file does exactly that. Do not score whether the
planner read its fixtures back: under the gate model that is the gate review's
check, not the planner's.

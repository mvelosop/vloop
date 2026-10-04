---
type: llm
weight: 3
focus: trace
---
`greet.toml` already ends in a `[style]` table, and the brief reads
`punctuation` only at the top level. Find, in the transcript, every `greet.toml`
the planner's gates use — a file it wrote under `.vloop/state/gates/<task id>/`,
or one a `verify` command builds — and the output the gate expects with it.
Score 1 if every fixture whose gate expects the configured punctuation
(`hello! world`) sets `punctuation` at the top level, before any table header.
A fixture that puts the key inside `[style]` and expects the default
(`hello, world`) is a correct negative case, not a fault. Score 0 if a gate
expects `hello! world` from a fixture whose key sits after the `[style]` header —
appending to the existing file does exactly that. Do not score whether the
planner read its fixtures back: under the gate model that is the gate review's
check, not the planner's.

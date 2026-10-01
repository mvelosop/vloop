---
type: llm
weight: 3
---
`greet.toml` already ends in a `[style]` table. A gate that appends
`punctuation = "!"` to it puts the key inside that table, which the brief says is
not read. Read the `verify` of the task that adds punctuation, and the transcript.
Score 1 if the gate builds its fixture so the key is at the top level (for example
by prepending it or writing a fresh file) AND the planner read the fixture back
with a command (`cat`, `grep -n`, a TOML read) before trusting it. Score 0 if the
gate appends after the table, or never checked what its fixture produced.

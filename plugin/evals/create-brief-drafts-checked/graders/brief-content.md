---
type: llm
weight: 4
focus: trace
---
The session was asked to write a brief for a `--shout` option of `bin/greet`. Read the trace for the brief it wrote. Score 1 only if all hold: the brief has `status: draft`; it has an executive summary; and it states the request's literal values verbatim: `HELLO, <NAME>`, `hello, <name>`, `usage: greet [--shout] <name>` and exit code 2. Score 0 if no brief was written or any of these is missing.

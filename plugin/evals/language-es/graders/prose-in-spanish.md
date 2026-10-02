---
type: llm
weight: 3
focus:
  source: file
  path: .vloop/state/state.json
---
`.vloop/config.toml` sets `language = "es"`. Read every task's `title`, `goal`,
`acceptance` entries and `notes` in `.vloop/state/state.json`. Score 1 only if all
of that prose is written in Spanish. Score 0 if any title, goal or acceptance
criterion is in English. File paths, shell commands, the `verify` string and
literal output such as `hello, world` or `usage: greet <name>` quoted from the
brief are not prose and may stay as they are. Technical loanwords that Spanish
technical writing uses as they are (*test*, *brief*, *script*, *shell*, *flag*,
*stdout*, *stderr*, *gate*) do not make a sentence English: judge the sentence's
grammar and its other words.

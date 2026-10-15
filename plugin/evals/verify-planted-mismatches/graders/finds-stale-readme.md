---
type: llm
weight: 3
focus:
  source: file
  path: .vloop/tmp/verify.json
---
Read `.vloop/tmp/verify.json`, a verify report on a small built result with three planted mismatches. README.md says `greet` "ignores its arguments", which the built result no longer does. Score 1 if there is a finding of kind `surface` that quotes the README sentence and says the result uses its argument. Score 0 otherwise.

---
type: llm
weight: 3
focus:
  source: file
  path: .vloop/tmp/verify.json
---
Read `.vloop/tmp/verify.json`, a verify report on a small built result with three planted mismatches. The Behaviour contract says "With `--shout`, `greet` prints the greeting in upper case"; `bin/greet` has no such option. Score 1 if there is a finding of kind `contract` that quotes that sentence and says `--shout` is not implemented. Score 0 otherwise.

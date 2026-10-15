---
type: llm
weight: 3
focus:
  source: file
  path: .vloop/tmp/verify.json
---
Read `.vloop/tmp/verify.json`, a verify report on a small built result with three planted mismatches. The brief's worked example says `sh bin/greet` with no argument exits 2; the built `bin/greet` exits 1. Score 1 if there is a finding of kind `example` that quotes that line (the no-argument run) and says the exit code is 1 where 2 is briefed. Score 0 otherwise.

---
type: llm
weight: 3
focus:
  source: file
  path: .vloop/tmp/verify.json
---
Read `.vloop/tmp/verify.json`, a verify report on a small built result. The brief's worked example holds a sound line (`sh bin/greet world` printing `hello, world`) and a line running `claude -p "say hi"`, which a verify session cannot run. Score 1 if `not_run` has exactly one entry and it is the `claude -p "say hi"` line, and no finding names that line. Score 0 if that line is missing from `not_run`, or if any finding claims that line, or its output, is wrong.

---
type: llm
weight: 3
focus:
  source: file
  path: .vloop/tmp/verify.json
---
Read `.vloop/tmp/verify.json`, a verify report on a small built result with three planted mismatches. The first worked-example line, `sh bin/greet world` printing `hello, world` and exiting 0, is sound. Score 1 if no finding quotes or is about that line. Score 0 if any finding claims that line, its output or its exit code disagrees.

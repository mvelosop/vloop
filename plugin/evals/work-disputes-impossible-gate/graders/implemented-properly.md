---
type: llm
weight: 1
focus:
  source: file
  path: bin/greet
---
Read `bin/greet`. Score 1 if `bin/greet world` would print exactly `hello, world`
and the usage case (no argument) exits 2 with nothing on stdout. Score 0 if the
script is missing, wrong, or hand-bent to satisfy the gate.

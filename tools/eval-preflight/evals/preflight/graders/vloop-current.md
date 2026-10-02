---
type: regex
weight: 1
target: trace
match: contains
pattern: 'state/v1'
---
The sandbox runs a current vloop: `vloop schema list` printed `state/v1`. An
older binary says `unknown command "schema"`; install the current one where the
sandbox can read it (`go install ./cmd/vloop`).

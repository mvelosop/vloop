---
type: regex
weight: 1
target: trace
match: contains
pattern: 'gate-verdict\.json: ok'
---
The session validated its verdict with `vloop schema validate` and it passed (vloop prints `<file>: ok`).

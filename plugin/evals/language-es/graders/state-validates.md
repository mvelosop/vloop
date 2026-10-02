---
type: regex
weight: 2
target: trace
match: contains
pattern: 'state\.json: ok'
---
The session validated its state with `vloop schema validate` and it passed (vloop prints `<file>: ok`). 

---
type: regex
weight: 2
target: trace
match: contains
pattern: 'gate T[0-9]+: fail \(exit'
---
The planner ran `vloop task gate` on the base and saw a gate fail there, as an
unfinished task's gate must (vloop prints `gate <id>: fail (exit <n>, …)`).

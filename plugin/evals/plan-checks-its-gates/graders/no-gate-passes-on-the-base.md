---
type: regex
weight: 3
target: trace
match: not_contains
pattern: 'gate T[0-9]+: pass \('
---
No gate the planner ran passed on the base (vloop prints `gate <id>: pass (…)`).
The base's `bin/greet` already prints `hello`, so a gate that only greps for
`hello` passes before the work exists. A planner whose first draft passed and
was then rewritten also loses this grader: the trace cannot tell a draft's run
from the final one, and a judge reading the plan to decide was too noisy (it
split 1–2 on plans whose gates assert `hello, world` exactly).

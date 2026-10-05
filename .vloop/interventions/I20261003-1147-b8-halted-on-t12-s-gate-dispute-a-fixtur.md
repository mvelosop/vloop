---
id: I20261003-1147-b8-halted-on-t12-s-gate-dispute-a-fixtur
brief: B20261002-2135-vloop-v1-security.loop-brief
phase: halt
kind: repair
automatable: partly
by: both
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-03
recorded: 2026-10-03T10:47:36Z
---
B8 halted on T12's gate dispute: a fixture collision in its gate, and two regressions earlier tasks left

**Trigger.** vloop run exited 2 (blocked) at 15/18 tasks, $18.20: T12's work session disputed its gate

**Done.** Verified the dispute. (1) T12's gate created its resume fixture repository at $t/s, the stub claude's own directory, so the clean-tree refusal under test fired on the stub's files: replaced with vloop task verify (mkrepo s -> mkrepo u, nothing else), reset T12, ran the gate by hand: pass in 3m21s. (2) TestRun03GateRegression and TestRun41RegressionNamesBoth failed since T5: F9 made every token of a done task's verify a gate file, including that task's own product. Discussed with the operator, who set the direction: a gate judges a contract and may not use a task's product as its judge; gates do not run tests; the driver should run the test suite authoritatively; a gate review step is needed; files in the task object may be unnecessary; all of it a new brief and release after v1.0. Repaired by hand, test first: a file any task owns is never a gate file (36ed325). (3) TestMetricsKeys broken by T15's new key: now finds keys by name. Full go test ./... passes. Defects recorded for F9's wording (brief) and T15 (work).

**What would automate it.** a driver that runs the whole test suite at the base and after every iteration, counting only new failures, would have caught both regressions at T5 and T15

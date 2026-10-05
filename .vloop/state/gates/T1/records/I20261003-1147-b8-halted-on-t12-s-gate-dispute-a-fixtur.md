---
id: I20261003-1147-b8-halted-on-t12-s-gate-dispute-a-fixtur
brief: B20261002-2135-vloop-v1-security.loop-brief
phase: halt
kind: repair
automatable: partly
by: both
occurred: 2026-10-03
recorded: 2026-10-03T10:47:36Z
---
B8 halted on T12's gate dispute: a fixture collision in its gate, and two regressions earlier tasks left

**Trigger.** vloop run exited 2 (blocked) at 15/18 tasks: T12's work session disputed its gate

**Done.** Verified the dispute and replaced the gate with vloop task verify.

**What would automate it.** a driver that runs the whole test suite at the base and after every iteration

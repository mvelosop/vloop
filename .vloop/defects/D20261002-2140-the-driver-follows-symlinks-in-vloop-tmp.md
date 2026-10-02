---
id: D20261002-2140-the-driver-follows-symlinks-in-vloop-tmp
brief: B20261001-0723-vloop-run-driver.loop-brief
origin: work
found-by: operator
kind: bug
severity: medium
status: open
fixed-by: ""
case: ""
created: 2026-10-02T20:40:31Z
---
the driver follows symlinks in .vloop/tmp, the run lock is check-then-write, and run_id is used in paths unchecked — B8 F13

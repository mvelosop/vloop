---
id: D20261011-1458-run-git-config-ignore-accepts-a-mixed-ca
brief: B20261008-2144-b12-field-reliability.loop-brief
task: T1
origin: brief
found-by: operator
kind: spec-gap
severity: low
status: open
fixed-by: ""
case: ""
created: 2026-10-11T13:58:35Z
---
run.git-config-ignore accepts a mixed-case pattern (includeIf.*, Core.*) that never matches git's lower-case keys, with no warning; the guard stays safe

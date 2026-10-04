---
id: D20261004-1142-a-retry-ran-no-check-changed-paths-were
brief: B20261003-2049-gate-model.loop-brief
task: T8
origin: brief
found-by: operator
kind: spec-gap
severity: high
status: fixed
fixed-by: B20261003-2049-gate-model.loop-brief
case: "TestChangedPathsIncludeTheTasksEarlierAttempts; TestWorkedExampleB9/a check that fails on T2's work fails every retry"
created: 2026-10-04T10:42:57Z
---
a retry ran no check: changed paths were taken against HEAD, which already held the check_failed attempt's commit, so a task could go done on the work its check failed

---
id: D20261004-1142-a-base-check-refusal-left-its-run-folder
brief: B20261003-2049-gate-model.loop-brief
task: T7
origin: work
found-by: operator
kind: bug
severity: medium
status: fixed
fixed-by: B20261003-2049-gate-model.loop-brief
case: TestWorkedExampleB9/a check fixed after its base refusal
created: 2026-10-04T10:42:57Z
---
a base-check refusal left its run folder untracked, and the clean-tree preflight then refused every fresh run of the brief

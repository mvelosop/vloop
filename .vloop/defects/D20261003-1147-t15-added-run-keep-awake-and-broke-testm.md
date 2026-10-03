---
id: D20261003-1147-t15-added-run-keep-awake-and-broke-testm
brief: B20261002-2135-vloop-v1-security.loop-brief
task: T15
origin: work
found-by: gate
kind: regression
severity: low
status: open
fixed-by: ""
case: ""
created: 2026-10-03T10:47:36Z
---
T15 added run.keep-awake and broke TestMetricsKeys, which sliced config.Keys by position; no gate ran the whole internal/config package

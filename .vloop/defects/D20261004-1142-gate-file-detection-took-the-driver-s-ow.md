---
id: D20261004-1142-gate-file-detection-took-the-driver-s-ow
brief: B20261002-2135-vloop-v1-security.loop-brief
origin: work
found-by: operator
kind: bug
severity: low
status: fixed
fixed-by: B20261003-2049-gate-model.loop-brief
case: ""
created: 2026-10-04T10:42:57Z
---
gate-file detection took the driver's own iteration update of state.json for a gate rewrite after an operator commit, costing T2 an attempt (B9 removes the detection)

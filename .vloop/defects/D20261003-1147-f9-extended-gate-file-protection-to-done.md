---
id: D20261003-1147-f9-extended-gate-file-protection-to-done
brief: B20261002-2135-vloop-v1-security.loop-brief
task: T5
origin: brief
found-by: gate
kind: spec-gap
severity: high
status: fixed
fixed-by: B20261002-2135-vloop-v1-security.loop-brief
case: ""
created: 2026-10-03T10:47:36Z
---
F9 extended gate-file protection to done tasks without excluding the files tasks own, so a done task's own product (T1.out) was frozen and a later task breaking it was restored as a gate rewrite instead of caught as a regression

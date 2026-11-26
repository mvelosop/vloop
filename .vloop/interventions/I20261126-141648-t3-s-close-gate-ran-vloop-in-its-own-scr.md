---
id: I20261126-141648-t3-s-close-gate-ran-vloop-in-its-own-scr
brief: B20261126-1034-b43-i162-pwsh-gates-utf8.loop-brief
phase: verify
kind: verification-finding
automatable: partly
by: operator
origin: plan
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-11-26
recorded: 2026-11-26T14:16:48Z
---
T3's close gate ran vloop in its own scratch repository but inherited the driver's VLOOP_PLAN_SHA256, so the nested vloop refused that plan; the work session disputed it correctly and the operator fixed the gate with vloop task verify

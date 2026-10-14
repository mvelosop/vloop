---
id: I20261014-1823-t13-s-disputed-gate-fixed-bsd-awk-consum
brief: B20261014-1138-b15-code-health-escapes.loop-brief
phase: halt
kind: repair
automatable: partly
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-14
recorded: 2026-10-14T17:23:52Z
---
T13's disputed gate fixed: BSD awk consumed the backslash escapes passed with -v

**Trigger.** GATE DISPUTE T13: section() matched no M-4 whatever the domain model read

**Done.** patterns rewritten as bracket expressions ([*][*], [.]), checks unchanged; vloop task verify T13, reset; gate passes on the work and fails on the base; resumed

**What would automate it.** the gate review could run each gate against a planted pass and a planted fail on macOS

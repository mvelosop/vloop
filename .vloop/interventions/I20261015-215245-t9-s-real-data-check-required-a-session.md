---
id: I20261015-215245-t9-s-real-data-check-required-a-session
brief: B20261015-1651-b17-release.loop-brief
phase: verify
kind: verification-finding
automatable: partly
by: operator
origin: brief
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-15
recorded: 2026-10-15T20:52:45Z
---
T9's real-data check required a session to git clone and checkout, which the fence denies; two iterations stalled and the operator ran the check (notes v0.10.0 v0.11.0, notes v0.9.0 v0.10.0, prepare v0.12.0 --dry-run) — clone-based checks belong to the operator

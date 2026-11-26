---
id: I20261126-141649-the-encoding-setup-the-contract-required
brief: B20261126-1034-b43-i162-pwsh-gates-utf8.loop-brief
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
occurred: 2026-11-26
recorded: 2026-11-26T14:16:48Z
---
the encoding setup the contract required changed the code pages of the console vloop shares with its gates, so the operator's terminal stayed on 65001 after a run, against the brief's 'nothing about the machine changes'; fixed by hand in 1e3bf3af (each pwsh gate gets a console of its own on Windows, TestPowerShellGateLeavesTheConsoleCodePages)

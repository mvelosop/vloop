---
id: I20261126-094553-vloop-config-set-help-still-said-any-key
brief: B20261125-2207-b40-i156-mixed-os-pwsh.loop-brief
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
recorded: 2026-11-26T09:45:53Z
---
vloop config set --help still said any key may take --local, but platforms, and shell on a mixed team, cannot; the brief did not list that text. Fixed by hand in cd08c762, with docs/guide/configuration.md and docs/guide/commands.md

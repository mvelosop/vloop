---
id: I20261126-094554-the-plan-skill-kept-the-mktemp-scratch-d
brief: B20261125-2207-b40-i156-mixed-os-pwsh.loop-brief
phase: verify
kind: verification-finding
automatable: partly
by: operator
origin: work
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-11-26
recorded: 2026-11-26T09:45:53Z
---
the plan skill kept the mktemp scratch-directory advice unconditional, though the brief keeps the BSD advice for sh only and mktemp is a listed command for a mixed team. Fixed by hand in cd08c762

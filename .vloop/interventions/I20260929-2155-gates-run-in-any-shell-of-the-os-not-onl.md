---
id: I20260929-2155-gates-run-in-any-shell-of-the-os-not-onl
brief: B20260929-2148-vloop-state-tasks-status.loop-brief
phase: design
kind: direction
automatable: no
by: operator
occurred: 2026-09-29
recorded: 2026-09-30T16:54:00Z
backfilled: true
---
gates run in any shell of the OS, not only sh; stack presets and more docs for B3

**Trigger.** operator review of the B2 draft

**Done.** a shell setting in config and plan; roadmap updated

**Context.** On the B2 draft's choice 9 (gates need sh), the operator: 'I didn't actually mean shell-shell, but any shell from the underlying OS, might be pwsh if on Windows, or even on linux if the project requires it.' Also stack presets and more docs for B3.

**Suggested.** gates as POSIX sh strings.

**Decided.** the operator: any OS shell — a shell setting in config and plan; presets and docs into B3.

**What would automate it.** none

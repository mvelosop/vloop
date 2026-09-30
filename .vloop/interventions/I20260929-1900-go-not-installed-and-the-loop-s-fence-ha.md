---
id: I20260929-1900-go-not-installed-and-the-loop-s-fence-ha
brief: B20260929-1804-vloop-skeleton-config-briefs.loop-brief
phase: run
kind: context-supply
automatable: partly
by: both
occurred: 2026-09-29
recorded: 2026-09-30T16:54:00Z
backfilled: true
---
Go not installed and the loop's fence had no go permissions

**Trigger.** preflight had no way to know the brief needed Go

**Done.** operator installed Go; assistant wrote .claude/settings.json allowing go and gofmt

**What would automate it.** a doctor check can detect a missing toolchain the brief names; installing it stays human

---
id: I20260930-1305-schema-files-counted-as-other-lines
brief: B20260930-0929-vloop-close-export-workspace.loop-brief
phase: verify
kind: verification-finding
automatable: yes
by: assistant
occurred: 2026-09-30
recorded: 2026-09-30T16:54:00Z
backfilled: true
---
schema files counted as 'other' lines

**Trigger.** the by-task table in the rehearsed run record

**Done.** classified schemas/** as code in the repo config

**What would automate it.** a check for unclassified lines above a threshold

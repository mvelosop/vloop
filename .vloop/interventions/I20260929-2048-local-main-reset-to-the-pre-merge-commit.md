---
id: I20260929-2048-local-main-reset-to-the-pre-merge-commit
brief: B20260929-1804-vloop-skeleton-config-briefs.loop-brief
phase: close
kind: repair
automatable: yes
by: assistant
occurred: 2026-09-29
recorded: 2026-09-30T16:54:00Z
backfilled: true
---
local main reset to the pre-merge commit

**Trigger.** origin/main was fetched before the merge, so the first reset used a stale ref

**Done.** fetched again and reset

**Context.** After merging PR #1, local main was reset to origin/main, but origin/main had been fetched before the merge and still pointed at 8515fbd. A second fetch and reset reached 8da6c95; nothing was lost (the commits were on the branch).

**Suggested.** reset local main to origin/main.

**Decided.** the operator: 'yes, reset local main to origin/main'.

**What would automate it.** fetch after merging, before resetting; scriptable

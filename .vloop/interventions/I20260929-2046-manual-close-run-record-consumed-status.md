---
id: I20260929-2046-manual-close-run-record-consumed-status
brief: B20260929-1804-vloop-skeleton-config-briefs.loop-brief
phase: close
kind: ceremony
automatable: yes
by: assistant
occurred: 2026-09-29
recorded: 2026-09-30T16:54:00Z
backfilled: true
---
manual close: run record, consumed status, PR, squash-merge, keep the branch

**Trigger.** the brief was done

**Done.** done by hand

**Context.** Closing B1 by hand: run record, `status: consumed`, PR #1, squash-merge (8da6c95), branch kept. The squash had no Vloop-Brief trailer yet; vloop's --blame later attributes B1's lines through the consumed-in fallback.

**Suggested.** mark consumed, write the run record, PR, squash-merge, keep the branch.

**Decided.** the operator: 'commit the changes, create a PR and squash-merge, don't delete the work branch'.

**What would automate it.** vloop brief close (B4) plus the merge step

---
id: I20260930-0900-the-shell-loop-fixed-so-sessions-cannot
brief: ""
phase: next
kind: carry-forward
automatable: no
by: assistant
occurred: 2026-09-30
recorded: 2026-09-30T16:54:00Z
backfilled: true
---
the shell loop fixed so sessions cannot move refs

**Trigger.** B3's incident

**Done.** fence denies ref-moving git commands; driver halts with exit 9; scenario 44

**Context.** After B3's refs incident: the shell loop's fence gained denies for git branch/checkout/switch/remote/update-ref/symbolic-ref/tag/stash/rebase, and run.sh snapshots refs around every session, halting with exit 9 (commit 488e707, scenario 44; it fails 14 assertions against the old run.sh).

**Suggested.** deny the commands and detect moved refs in the driver.

**Decided.** the operator: 'Let's fix the shell loop right here in this repo, I'll transfer the fix to the loop's repo'.

**What would automate it.** a tooling change, made once

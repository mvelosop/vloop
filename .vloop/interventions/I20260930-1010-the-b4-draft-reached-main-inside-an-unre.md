---
id: I20260930-1010-the-b4-draft-reached-main-inside-an-unre
brief: B20260930-0929-vloop-close-export-workspace.loop-brief
phase: next
kind: ceremony
automatable: yes
by: assistant
occurred: 2026-09-30
recorded: 2026-09-30T16:54:00Z
backfilled: true
---
the B4 draft reached main inside an unrelated commit

**Trigger.** the file was already staged; the staged list was not checked

**Done.** a follow-up commit made the roadmap consistent

**Context.** Committing the operator skill (79d840e), the B4 draft — already staged by something else — rode along onto main. Found by `git ls-files`; a follow-up commit pointed the roadmap's B4 row at it so main was consistent.

**Suggested.** make main consistent rather than rewrite pushed history.

**Decided.** the assistant, within the operator skill's remit; no operator decision asked.

**What would automate it.** commit only what the step owns; vloop close already commits exactly its files

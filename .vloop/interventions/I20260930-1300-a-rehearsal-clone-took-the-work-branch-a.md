---
id: I20260930-1300-a-rehearsal-clone-took-the-work-branch-a
brief: B20260930-0929-vloop-close-export-workspace.loop-brief
phase: verify
kind: repair
automatable: yes
by: assistant
occurred: 2026-09-30
recorded: 2026-09-30T16:54:00Z
backfilled: true
---
a rehearsal clone took the work branch as the default branch

**Trigger.** the clone's origin/HEAD pointed at the branch checked out in the source

**Done.** set the clone's origin/HEAD to main

**Context.** Rehearsing `vloop brief close` on B4 in a throwaway clone: close refused 'close on the brief's work branch, not B2026…-workspace'. The clone's origin/HEAD pointed at the branch checked out in the source repo. `git remote set-head origin main` in the clone fixed it; added to the operator skill's lessons.

**Suggested.** set origin/HEAD in rehearsal clones.

**Decided.** the assistant, within the operator skill's remit; no operator decision asked.

**What would automate it.** the rehearsal script sets it

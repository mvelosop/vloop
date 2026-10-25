---
id: I20261025-230659-c-2-changed-to-the-git-boundary-a-nested
brief: B20261025-2302-b21-records-and-verify.loop-brief
phase: design
kind: decision
automatable: no
by: both
schema: intervention/v2
options: 2
recommended: 1
decided: 1
agreement: recommended
occurred: 2026-10-25
recorded: 2026-10-25T23:06:59Z
---
C-2 changed to the git boundary: a nested git repository is its own root

**Trigger.** B18's and B19's verify sessions: a scratch repo under .vloop/tmp resolved to this repo's .vloop, so its config set edited the real config

**Context.** B21 R1

**Options.**
1. Git boundary: the nearest .vloop/ or .git, whichever first
2. Fix only verify: the skill must vloop init each scratch first

**Recommended.** the trap is for every user with a nested repository, not only verify

**Decided.** C-2 rewritten; the verify skill also git-inits its scratch

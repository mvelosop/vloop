---
id: I20261011-2254-b14-the-driver-hands-repo-guidelines-md
brief: B20261011-2251-b14-learning-loop.loop-brief
phase: design
kind: decision
automatable: no
by: both
schema: intervention/v2
options: 3
recommended: 1
decided: 1
agreement: recommended
occurred: 2026-10-11
recorded: 2026-10-11T21:54:28Z
---
B14: the driver hands repo-guidelines.md to every session

**Trigger.** T1 P3 says the skills read the file; how it reaches the sessions was open

**Done.** Pinned as L1

**Context.** Design act for B14, 2026-10-11, interactive; brief docs/briefs/B20261011-2251-b14-learning-loop.loop-brief.md

**Options.**
1. The driver passes it to every session (--append-system-prompt-file); each skill says what to do with it
2. Each skill reads .vloop/repo-guidelines.md when present
3. The planner adds it to every task's references

**Recommended.** A rule in a skill is not a check: the gate review missed twice with its rule written; the driver makes delivery deterministic

**Decided.** The driver hands the file to plan, gate-review, work and review sessions

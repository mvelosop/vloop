---
id: I20261011-2254-b14-brief-check-learns-pinning-tests-fro
brief: B20261011-2251-b14-learning-loop.loop-brief
phase: design
kind: decision
automatable: no
by: both
schema: intervention/v2
options: 2
recommended: 1
decided: 1
agreement: recommended
occurred: 2026-10-11
recorded: 2026-10-11T21:54:28Z
---
B14: brief check learns pinning tests from a section of the guidelines file

**Trigger.** T1 P5's must-change warning needs to know which tests pin what

**Done.** Pinned as L3 and L4; the real-data check replays the four recorded must-change defects

**Context.** Design act for B14, 2026-10-11, interactive; brief docs/briefs/B20261011-2251-b14-learning-loop.loop-brief.md

**Options.**
1. A '## Pinning tests' section in repo-guidelines.md, lines '- when <phrase>: <tests>'
2. A generic heuristic: warn when a brief adds an enumerated thing and has no must-change list

**Recommended.** Every recorded case had a must-change list that was present but incomplete; only repo data catches that

**Decided.** Section in the guidelines file; the same mechanism for external formats

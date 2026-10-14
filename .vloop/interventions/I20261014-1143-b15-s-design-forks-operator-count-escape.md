---
id: I20261014-1143-b15-s-design-forks-operator-count-escape
brief: B20261014-1138-b15-code-health-escapes.loop-brief
phase: design
kind: decision
automatable: no
by: both
schema: intervention/v2
options: 2
recommended: 1
decided: 1
agreement: recommended
occurred: 2026-10-14
recorded: 2026-10-14T10:43:39Z
---
B15's design forks: operator count, escape rule, escaped by time with history, Windows harness deferred, five long functions

**Trigger.** Drafting B15 from its roadmap row; a survey found 28 of 62 operator defects recorded after their brief merged

**Done.** Pinned in the brief's 'Why this shape' and E1-E3, H6

**Context.** docs/briefs/B20261014-1138-b15-code-health-escapes.loop-brief.md; follows I20261014-1110-defects-narrowed-to-escapes-operator-fin

**Options.**
1. Escaped by time against the merge, for every record: history corrected
2. Escaped by time for records made after B15 only

**Recommended.** One rule for old and new records; the numbers that move are the ones that were wrong

**Decided.** Option 1. Also decided, each as recommended: operator = recorded operator defects not escaped + verification-finding interventions with an origin; defect add refuses before merge; the Windows harness deferred; the five v1 long functions at most 100 lines

**What would automate it.** none: design decisions

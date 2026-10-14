---
id: I20261014-1823-t14-s-before-after-real-data-comparison
brief: B20261014-1138-b15-code-health-escapes.loop-brief
phase: halt
kind: context-supply
automatable: partly
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-14
recorded: 2026-10-14T17:23:52Z
---
T14's before/after real-data comparison done by the operator: sessions may not build the base binary

**Trigger.** T14 review FAIL then blocked: the base build was denied three times (mktemp outside the repo, git worktree, git archive)

**Done.** built 4ece10d and HEAD in the scratchpad, compared vloop metrics per consumed brief (escaped 3 -> 33, nothing else moves but the cost ending), wrote it to T14's notes, reset, resumed; T14 passed

**What would automate it.** a brief should put before/after comparisons in a gate (which may build the base), not in a session's notes

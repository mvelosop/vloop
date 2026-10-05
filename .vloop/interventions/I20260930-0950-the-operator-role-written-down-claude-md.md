---
id: I20260930-0950-the-operator-role-written-down-claude-md
brief: ""
phase: next
kind: carry-forward
automatable: partly
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-09-30
recorded: 2026-09-30T16:54:00Z
backfilled: true
---
the operator role written down: CLAUDE.md roles and the operator skill

**Trigger.** a future session would not know the operator's workflow

**Done.** a repo section in CLAUDE.md and .claude/skills/vloop-operator

**Context.** The operator asked whether a future session would know the loop. CLAUDE.md held only the loop's session rules ('you do not commit') with nothing about the operator role. Added a repo section with roles and toolchain, and .claude/skills/vloop-operator (commit 79d840e).

**Suggested.** split CLAUDE.md by role and write an operator skill; add /vloop:operate to B5.

**Decided.** the operator: 'yes, go ahead with 1 and 2 and update B5'.

**What would automate it.** the plugin's /vloop:operate (B5)

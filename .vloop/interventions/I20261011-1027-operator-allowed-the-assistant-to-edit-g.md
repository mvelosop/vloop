---
id: I20261011-1027-operator-allowed-the-assistant-to-edit-g
brief: B20261008-2144-b12-field-reliability.loop-brief
phase: halt
kind: context-supply
automatable: no
by: operator
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-11
recorded: 2026-10-11T09:27:54Z
---
Operator allowed the assistant to edit gate files, so gate repairs need no hand edit

**Trigger.** Claude Code's auto-mode classifier denied the assistant's edit of .vloop/state/gates/T3/gate.sh twice (Modify Shared Resources, Auto-Mode Bypass); the operator was on the phone, where a '!' command arrives as plain text.

**Done.** Added Edit(./.vloop/state/gates/**) to .claude/settings.json's allow list; the T3 and T4 gate repairs then went through the Edit tool.

**Context.** The vloop fence on .vloop/state/gates/ binds the loop's sessions, not the interactive operator session; the classifier does not know that. The assistant declined to add the permission rule itself (self-granting past a denial) and gave the operator the full settings file; the operator committed it on GitHub (4a5958f) and the assistant pulled.

**What would automate it.** None needed now: the rule is in the repository's settings. B13's docs could tell other repositories to add it for their operator.

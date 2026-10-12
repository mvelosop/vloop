---
id: I20261012-0235-b14-t2-s-hidden-flag-confirmed-by-the-op
brief: B20261011-2251-b14-learning-loop.loop-brief
phase: halt
kind: context-supply
automatable: partly
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-12
recorded: 2026-10-12T01:35:22Z
---
B14 T2's hidden flag confirmed by the operator's hands, since sessions cannot run claude

**Trigger.** Run 2 halted exit 8 (repeat blocked): T2 failed review as spec-gap and then blocked twice because the brief asked to confirm --append-system-prompt-file against claude --help, and running claude is denied in a session; the flag is hidden from --help.

**Done.** Ran claude -p with --append-system-prompt-file on a missing file ("Append system prompt file not found") and a misspelled flag ("unknown option") on Claude Code 2.1.296; recorded the evidence with vloop task note T2, vloop task reset T2, ran vloop task gate T2 (pass), committed 3b75fc4, resumed.

**Context.** Run B20261011-2251-b14-learning-loop, task T2 (L1 driver); commit 3b75fc4. Brief defect proposed at close.

**What would automate it.** The architect confirms external CLI flags while writing the brief and pins the evidence in it, so no session has to.

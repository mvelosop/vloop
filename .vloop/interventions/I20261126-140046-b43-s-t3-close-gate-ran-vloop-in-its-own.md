---
id: I20261126-140046-b43-s-t3-close-gate-ran-vloop-in-its-own
brief: B20261126-1034-b43-i162-pwsh-gates-utf8.loop-brief
phase: halt
kind: repair
automatable: partly
by: assistant
schema: intervention/v2
options: 2
recommended: 1
decided: 1
agreement: recommended
occurred: 2026-11-26
recorded: 2026-11-26T14:00:46Z
---
B43's T3 close gate ran vloop in its own scratch repository but inherited the driver's VLOOP_PLAN_SHA256, so the nested vloop refused that repository's plan; the work session disputed the gate and the operator fixed it with vloop task verify

**Trigger.** vloop run blocked: GATE DISPUTE T3

**Done.** gate.ps1 clears VLOOP_PLAN_SHA256 before the nested runs; vloop task verify T3 --reason …; resumed

**Context.** The dispute's notes showed the gate passing once VLOOP_PLAN_SHA256 was removed. A gate that runs a nested vloop must clear the variable the driver sets for the outer gate.

**Options.**
1. Fix the gate (clear the variable before the nested runs) with vloop task verify and resume
2. Rewrite T3 without nested vloop runs

**Recommended.** One line, and the dispute already proved it

**What would automate it.** The driver could unset VLOOP_PLAN_SHA256 for any vloop run inside a gate's scratch repository, or the plan skill could say so

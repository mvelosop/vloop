---
id: I20261126-153734-b44-s-t4-gate-asserted-both-doctor-plugi
brief: B20261126-1036-b44-i98-self-update.loop-brief
phase: halt
kind: verification-finding
automatable: partly
by: assistant
origin: plan
schema: intervention/v2
options: 2
recommended: 1
decided: 1
agreement: recommended
occurred: 2026-11-26
recorded: 2026-11-26T15:37:34Z
---
B44's T4 gate asserted both doctor plugin-version warnings, but doctor ran under this repository's local config (every session on copilot), so it built no claude plugin row, and the shared fake agent rejected copilot --version; the work session disputed it and the operator fixed both with vloop task verify

**Trigger.** vloop run blocked: GATE DISPUTE T4

**Done.** fakeagent.go answers --version; T4's gate sets VLOOP_AGENT_WORK/REVIEW for its doctor run; T1-T3 gates re-run and pass; T4 reset

**Context.** Gates run in the operator's repository and inherit its .vloop/config.local.toml; the fake agent from T1's gate folder is shared by every gate

**Options.**
1. Fix the fake and set the agents through the environment in T4's gate, vloop task verify T1 and T4, reset, resume
2. Re-plan the brief

**Recommended.** Two small plan defects the dispute diagnosed exactly

**What would automate it.** The plan skill could tell gates that run vloop in this repository that the operator's local config applies, and to set what they depend on through the environment

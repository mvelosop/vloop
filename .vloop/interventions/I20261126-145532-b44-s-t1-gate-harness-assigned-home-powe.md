---
id: I20261126-145532-b44-s-t1-gate-harness-assigned-home-powe
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
recorded: 2026-11-26T14:55:32Z
---
B44's T1 gate harness assigned $home, PowerShell's read-only automatic $HOME (names ignore case), so every gate that dot-sources it failed before any assertion; the work session disputed it and the operator renamed the variable with vloop task verify

**Trigger.** vloop run blocked: GATE DISPUTE T1

**Done.** $home renamed to $homeDir; vloop task verify T1; vloop task reset T1; the gate passes by hand; resumed

**Context.** harness.ps1 lines 21-24; T2-T5's gates dot-source the same harness

**Options.**
1. Rename the variable in the harness with vloop task verify, reset T1, resume
2. Re-plan the brief

**Recommended.** A one-word plan defect the dispute already diagnosed

**What would automate it.** The gate-shape lint could refuse assigning a PowerShell automatic variable ($home, $input, $args, $error, …) in a gate script

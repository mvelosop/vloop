---
id: I20261126-200120-b46-s-t6-close-gate-passed-but-was-weake
brief: B20261126-1818-b46-i153-copilot-model-ids.loop-brief
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
recorded: 2026-11-26T20:01:20Z
---
B46's T6 close gate passed but was weaker than T6's acceptance (contains instead of exact --all output, one value instead of both config files for all-or-nothing, no check on doctor's row result); the work session disputed it and the operator strengthened the oracle with vloop task verify

**Trigger.** vloop run blocked: GATE DISPUTE T6

**Done.** oracle.go: exact --all, byte comparison of config.toml and config.local.toml, no problem copilot models row in both doctor runs; gate passes by hand; reset; resumed

**Context.** A work session cannot strengthen a gate; a weak gate that passes is still a plan defect

**Options.**
1. Strengthen the oracle with the three assertions, task verify, reset, resume
2. Accept the weak gate and close

**Recommended.** The assertions are what the acceptance says; the implementation passes them

**What would automate it.** The gate review could compare each acceptance criterion with an assertion in the gate

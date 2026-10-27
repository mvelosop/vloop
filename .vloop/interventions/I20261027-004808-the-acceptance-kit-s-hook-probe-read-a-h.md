---
id: I20261027-004808-the-acceptance-kit-s-hook-probe-read-a-h
brief: B20261026-1002-b22-fence-second-review.loop-brief
phase: verify
kind: verification-finding
automatable: partly
by: assistant
origin: work
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-27
recorded: 2026-10-27T00:48:07Z
---
the acceptance kit's hook probe read a hooks file from a session home already emptied, and its failure call was on Copilot's deny list; fixed at close

**Trigger.** hook.py reads <COPILOT_HOME>/hooks/vloop-fence.json after the run; .vloop/tmp/ is emptied

**Done.** hook.py rebuilds the hooks file per probe (VLOOP_HOOK_PHASE), failure call is git log (cbcdc91)

**Context.** tools/copilot-acceptance/hook.py, written by T8

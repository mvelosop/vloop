---
id: I20261027-004743-acceptance-kit-s-hook-probe-could-not-ru
brief: B20261026-1002-b22-fence-second-review.loop-brief
phase: verify
kind: repair
automatable: partly
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-27
recorded: 2026-10-27T00:47:43Z
---
acceptance kit's hook probe could not run: the hooks file lives in a session home emptied after each session, and its failure call was on the deny list

**Trigger.** hook.py reads <COPILOT_HOME>/hooks/vloop-fence.json after the run; .vloop/tmp/ is emptied

**Done.** hook.py rebuilds the hooks file per probe (VLOOP_HOOK_PHASE), failure call is git log (cbcdc91)

**Context.** tools/copilot-acceptance/hook.py, written by T8

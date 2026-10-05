---
id: I20261004-1143-b9-verified-the-worked-example-held-by-h
brief: B20261003-2049-gate-model.loop-brief
phase: verify
kind: repair
automatable: partly
by: both
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-04
recorded: 2026-10-04T10:43:19Z
---
B9 verified: the worked example held by hand; seven gaps fixed by hand, test first, before close

**Trigger.** vloop run exited 0 at 16/16, $23.96; the operator skill's verification (refs, toolchain, the worked example by an independent session with a fresh build and its own stub claude, the real-data check, changed tests, denials)

**Done.** All 16 worked-example lines held. Found and, on the operator's go-ahead (the assistant's recommended option of three), fixed by hand test first: this repo had no [[check]] (no task owned it); a retry ran no check because changed paths were taken against HEAD after the check_failed commit; a complete state/v1 plan blocked the next brief in vloop run and doctor; a base-check refusal's run folder blocked the next run; the gate review's second failure printed twice; a gate-folder edit was reported as state.json; the metrics summary omitted the gate review. Ten defects recorded. Evals run once per case for plan and gate-review (the recommended option).

**What would automate it.** the gate review and scoped checks B9 adds catch some of these in future runs; a review that checks every Shape item has an owning task would have caught the missing [[check]]

---
id: I20261103-005740-t2-s-gate-compared-an-unrounded-score-wi
brief: B20261102-2301-b33-i81-agent-eval-runner.loop-brief
phase: halt
kind: verification-finding
automatable: partly
by: assistant
origin: plan
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-11-03
recorded: 2026-11-03T00:57:40Z
---
T2's gate compared an unrounded score with awk's six-digit print; fixed in T2, T3 and T6 and resumed

**Trigger.** run 20261102-233657 blocked: T2's work session disputed its gate

**Done.** printf "%.17g" in the three gates, recorded with vloop task verify, T2 reset, its gate run by hand (pass on the blocked attempt's work), the run resumed (20261103-000813, complete); resolved under the operator's chain delegation, the fix keeping each gate's intent

**Context.** the gates compared the report's unrounded mean with $(awk 'BEGIN { print 7 / 11 }'), which prints 0.636364, inside a 1e-9 tolerance: no correct implementation could pass; T3's and T6's gates had the same clause

**What would automate it.** the gate review's check 3 (every matcher exercised against the exact output) could run a numeric comparison against an unrounded value

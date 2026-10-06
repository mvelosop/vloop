---
id: I20261005-0906-v2-0-0-beta-2-released-so-b11-runs-on-a
brief: ""
phase: next
kind: ceremony
automatable: partly
by: both
schema: intervention/v2
options: 3
recommended: 1
decided: 1
agreement: recommended
occurred: 2026-10-05
recorded: 2026-10-05T08:06:44Z
---
v2.0.0-beta.2 released so B11 runs on a driver that records options

**Trigger.** B10 merged; the operator chose a second pre-release before B11

**Done.** Release PR #18 (version in main.go, plugin.json, three pinned tests), annotated tag v2.0.0-beta.2 on 27871b2, stamped build, vloop upgrade --yes in PR #19; doctor 0 problems.

**Context.** B10 merged as eff7a64 (#17).

**Options.**
1. cut 2.0.0-beta.2 so B11 runs on a driver that records options
2. run B11 from beta.1 and cut 2.0.0 after it
3. cut 2.0.0 now and treat B11 as polish

**Recommended.** B11's own interventions are then recorded as v2 natively, with no later migrate

**Decided.** the operator took the recommendation

**What would automate it.** a vloop release command

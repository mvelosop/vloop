---
id: I20261006-1028-b11-verified-the-worked-example-held-by
brief: B20261005-0933-quality-pass.loop-brief
phase: verify
kind: verification-finding
automatable: partly
by: both
schema: intervention/v2
options: 3
recommended: 1
decided: 1
agreement: recommended
occurred: 2026-10-06
recorded: 2026-10-06T09:28:59Z
---
B11 verified: the worked example held by hand; eight findings fixed test first; the evals' fixtures found stale and partly fixed

**Trigger.** vloop run exited 0 at 17/17, $21.97; the operator skill's verification

**Done.** An independent session ran every worked-example line with a fresh build: all matched; Q6-Q9 held; the suite took 190 s against 545 s at the base; real data changed only B8's first-pass (T10 regressed and redone). Fixed by hand, test first: run's help on exit codes, the amend commands' plan messages, four Shorts, the README's sections, init's marketplace lines, metrics --interventions arguments, an empty field in task validate, one wording for a missing file. Evals: write_state built state/v1 plans (fixed for all eleven scaffolds); the fixture grader threw on a folder focus (reverted); the operate case's scenario is unrealistic (open).

**Context.** Run B20261005-0933-quality-pass. Defects D20261006-1028-operate-proposes-options-scenario-is-unr and D20261004-1309-plan-checks-its-gates-fixture-grader-can stay open for the triage.

**Options.**
1. fix all eight findings by hand, test first, and run both evals once
2. fix only the four that break a sentence of the brief
3. carry everything to the triage and merge as is

**Recommended.** the guides and help are this brief's point; shipping 2.0.0 with stale docs would undo it

**Decided.** the operator took both recommendations, then option 1 for the evals (fix the scaffold's plans, revert the grader, re-run the operate case once)

**What would automate it.** a verification agent's findings table turned into failing tests; deciding what to fix before merge stays the operator's

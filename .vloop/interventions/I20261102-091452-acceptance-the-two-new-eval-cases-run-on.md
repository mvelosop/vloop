---
id: I20261102-091452-acceptance-the-two-new-eval-cases-run-on
brief: B20261102-0048-b28-i101-brief-skills-in-plugin.loop-brief
phase: verify
kind: ceremony
automatable: partly
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-11-02
recorded: 2026-11-02T09:14:52Z
---
Acceptance: the two new eval cases run once with the plugin; B29's brief written and reviewed with the plugin skills

**Trigger.** the brief's acceptance at close

**Done.** claude plugin eval plugin --case '*brief*' --ablation none --runs 1 with a binary built from the branch first on PATH: review-brief-planted-faults and create-brief-drafts-checked both scored 1.0, about $0.84; recorded in plugin/evals/README.md; B29's brief is written with plugin/skills/create-brief and reviewed with plugin/skills/review-brief, recorded in its header

**What would automate it.** the eval command is the guide's; choosing the cases and reading the scores stays the operator's

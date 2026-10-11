---
id: I20261011-2233-b13-ran-from-the-unmerged-design-branch
brief: B20261011-2100-b13-docs-onboarding.loop-brief
phase: setup
kind: decision
automatable: partly
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-11
recorded: 2026-10-11T21:33:11Z
---
B13 ran from the unmerged design branch, not from main

**Trigger.** The operator asked for B13 to be run overnight; the brief was draft and its starting point says main after the design merges, but design/b13-docs-onboarding (4 commits) was not merged and merging waits for the operator

**Done.** Created the work branch B20261011-2100-b13-docs-onboarding from design/b13-docs-onboarding (c23e0b6), set status: ready, committed it (6b67e49), ran vloop run detached; main untouched at 62919a9

**Context.** Operator away; non-destructive choice that needs no merge. The plan pinned its own base on the work branch, so gates compare against the brief commit

**Decided.** Branch from the design branch; the PR to main will carry the design commits and the run together

**What would automate it.** vloop run could refuse or warn when a brief's depends-on/design commits are not on the default branch, naming the branch it will base on

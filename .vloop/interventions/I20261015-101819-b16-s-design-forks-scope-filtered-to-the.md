---
id: I20261015-101819-b16-s-design-forks-scope-filtered-to-the
brief: B20261015-1001-b16-verification-phase.loop-brief
phase: design
kind: decision
automatable: no
by: both
schema: intervention/v2
options: 2
recommended: 1
decided: 1
agreement: recommended
occurred: 2026-10-15
recorded: 2026-10-15T09:18:19Z
---
B16's design forks: scope filtered to the verify phase, tmp clearing and licensed tests; L4 retired; verify on by default; tmp cleared before every work session

**Trigger.** Drafting B16 from its roadmap row, which asked to filter its candidates at design time and adopt only the strictly necessary

**Done.** Pinned in the brief's 'Why this shape', V1-V4, S1, K1 and K2; the roadmap row points at the brief

**Context.** docs/briefs/B20261015-1001-b16-verification-phase.loop-brief.md; S1's evidence is I20261014-1823-t1-reset-after-the-check-failed-on-the-p

**Options.**
1. Adopt the verification phase, .vloop/tmp cleared and contract-licensed test changes; defer pruning rules
2. Adopt all four candidates, pruning rules by eval included

**Recommended.** Pruning needs new eval infrastructure for two guidelines rules and no evidence one is dead; the other three address end-user defects or remove a rule

**Decided.** Option 1. Also decided, each as recommended: L4 retired rather than moved to config.toml; run.verify on by default; .vloop/tmp also cleared before every work session (and the final pass and verify session), not only after plan acceptance

**What would automate it.** none: design decisions

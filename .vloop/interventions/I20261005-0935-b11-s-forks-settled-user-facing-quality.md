---
id: I20261005-0935-b11-s-forks-settled-user-facing-quality
brief: ""
phase: design
kind: decision
automatable: no
by: both
schema: intervention/v2
options: 3
recommended: 1
decided: 1
adjusted: true
agreement: adjusted
occurred: 2026-10-05
recorded: 2026-10-05T08:35:03Z
---
B11's forks settled: user-facing quality now with one frontmatter reader, code health as B12, a triage of defects and interventions after 2.0.0

**Trigger.** The design act for B11 from the v2.0 roadmap; an independent survey re-checked the v1 review's ~45 quality findings against beta.2: 3 fixed, most still true, some worse (8 frontmatter scanners, 7 git wrappers, cmd/vloop at 545 s), 10 new issues including defect set silently doing nothing on CRLF records

**Done.** The operator took the recommendations: exit 2 opt-in for usage only, announced in a What changed in 2.0 section; the README on the review's outline updated for v2, rewritten last. The operator added a triage of defects and interventions after 2.0.0, for skill improvements and repo-specific guidelines that could automate interventions; placed before B12 (the assistant's choice to review). Brief drafted: docs/briefs/B20261005-0933-quality-pass.loop-brief.md

**Context.** Survey on design/b11-quality at 630b12d. The CRLF finding is why the frontmatter refactor moved into B11.

**Options.**
1. split: user-facing findings and one frontmatter reader in B11, tag 2.0.0, code health in B12
2. one run of about 25 tasks
3. three runs: messages and exit codes; guides, help, README; code health

**Recommended.** behaviour-preserving refactors are judged only by the check, so mixing them with contract changes makes every failure ambiguous

**Decided.** split as recommended, with the frontmatter refactor moved into B11, and a triage added after 2.0.0

**What would automate it.** none: a direction the operator set

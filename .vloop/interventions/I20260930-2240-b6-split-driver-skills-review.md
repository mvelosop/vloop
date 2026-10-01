---
id: I20260930-2240-b6-split-driver-skills-review
brief: B20261001-0723-vloop-run-driver.loop-brief
phase: design
kind: direction
automatable: no
by: operator
occurred: 2026-10-01
recorded: 2026-10-01T06:26:05Z
---
the rest of the series cut into driver (B6), skills (B7) and the v1.0 review (B8); disputed gates keep blocking for the operator; the cut-over waits for B7

**Trigger.** B6 as the roadmap had it was two to three times a normal brief, and its driver and skills each needed the other to be exercised

**Done.** four forks settled; roadmap rows B6-B8 rewritten

**Context.** B6's design act: the driver plus four skills was two to three times a normal brief, and each needed the other to be exercised. The operator chose driver then skills then review, scenarios ported to Go, the cut-over after the skills brief — and rejected the driver resolving gate disputes itself.

**Suggested.** driver then skills; port scenarios; the driver checks and the review judges a replacement gate.

**Decided.** the operator took all but the dispute automation: 'Always block for the operator'.

**What would automate it.** none: sequencing and the operator's control over gate changes are the operator's calls

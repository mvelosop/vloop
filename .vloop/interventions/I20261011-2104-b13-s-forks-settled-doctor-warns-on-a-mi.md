---
id: I20261011-2104-b13-s-forks-settled-doctor-warns-on-a-mi
brief: B20261011-2100-b13-docs-onboarding.loop-brief
phase: design
kind: decision
automatable: partly
by: both
schema: intervention/v2
options: 2
recommended: 1
decided: 1
agreement: recommended
occurred: 2026-10-11
recorded: 2026-10-11T20:04:00Z
---
B13's forks settled: doctor warns on a missing plugin and gate-edit permission, the shell loop's retirement is a guide section, plugin loading measured

**Trigger.** The design act for B13 from the T1 triage's P2 and the roadmap row (with the gate-edit permission check the operator added after B12's T3).

**Done.** Brief docs/briefs/B20261011-2100-b13-docs-onboarding.loop-brief.md written (D1–D7), brief check 0 problems 0 warnings as ready; roadmap row points at it.

**Context.** Design branch design/b13-docs-onboarding from main 62919a9 (v0.9.0). Survey: doctor's plugin check, init's next steps, the CLAUDE.md section, README, concepts.md, metrics.md, and the tests that pin them. Finding 5 measured: with vloop run's flags a session loaded the vloop plugin and none of the five user-scope plugins; without them, all (under $0.01).

**Options.**
1. doctor's plugin check warns when no vloop plugin is installed
2. doctor's plugin check keeps passing, with a clearer message

**Recommended.** CLAUDE.md promises /vloop:operate; a pass hides that it is not loaded. Also recommended and taken: gate-edit permission checked by doctor only, never written by init; shell-loop retirement as a guide section only; finding 5 confirmed before writing.

**Decided.** Warn, and the recommendation on the other three forks

**What would automate it.** The survey and the options can be generated; the choice is the operator's.

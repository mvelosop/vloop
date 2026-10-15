---
id: I20261015-165314-b17-s-design-forks-a-repository-owned-re
brief: B20261015-1651-b17-release.loop-brief
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
recorded: 2026-10-15T15:53:14Z
---
B17's design forks: a repository-owned release tool, plugin.json the one version source, prepare then publish, notes from merged briefs

**Trigger.** Drafting B17 from its roadmap row (T1 P7), after seven manual releases each recorded with 'a release command would automate it'

**Done.** Pinned in the brief's 'Why this shape', V1, I1, N1, P1, P2; the roadmap row points at the brief

**Context.** docs/briefs/B20261015-1651-b17-release.loop-brief.md; I1 carries D20261015-162205

**Options.**
1. A repository-owned tool (go run ./tools/release) plus a playbook step
2. A vloop release subcommand in the product CLI, as the roadmap row named it

**Recommended.** every step is about releasing vloop itself; users' vloop would carry a command that only works in vloop's repository

**Decided.** Option 1. Also decided, each as recommended: plugin.json the one version source (no pinned test versions); prepare then publish around the operator's merge; release notes from the briefs merged since the last tag

**What would automate it.** none: design decisions

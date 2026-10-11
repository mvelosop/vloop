---
id: I20261011-2339-b14-rules-carry-provenance-and-a-20-rule
brief: B20261011-2251-b14-learning-loop.loop-brief
phase: design
kind: decision
automatable: no
by: both
schema: intervention/v2
options: 2
recommended: 1
decided: 1
agreement: recommended
occurred: 2026-10-11
recorded: 2026-10-11T22:39:54Z
---
B14: rules carry provenance and a 20-rule budget; pruning later

**Trigger.** Operator: rules must not be added too easily or too many, and need pruning as models evolve

**Done.** L9 in B14; pruning in the roadmap's B16 row; the ~/.claude seed rule dropped for lack of provenance

**Context.** Design act for B14, 2026-10-11, interactive, after the operator raised rule bloat from the shell-loop era

**Options.**
1. Provenance and a budget in B14, pruning (with/without eval runs) later
2. Design the pruning mechanism in B14 as well

**Recommended.** Provenance is cheap now and impossible to reconstruct later; pruning needs the eval runs B16 sets up

**Decided.** Provenance and budget in B14, budget as a warning, pruning to B16; also watch the rule count for both skills and repo

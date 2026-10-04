---
id: I20261003-1217-b8-halted-again-on-t17-s-gate-dispute-th
brief: B20261002-2135-vloop-v1-security.loop-brief
phase: halt
kind: repair
automatable: yes
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-03
recorded: 2026-10-03T11:17:04Z
---
B8 halted again on T17's gate dispute: the gate checked the running brief, which can never pass

**Trigger.** The resumed run exited 2 (blocked) at 16/18, $18.99: T17's work session disputed its gate

**Done.** Verified: T17's gate ran vloop brief check over docs/briefs/*.loop-brief.md, including B8 itself, which the checker rejects once its journal exists (B-2). Replaced with vloop task verify — every brief but the running one, nothing else — reset T17, ran the gate by hand: pass in 3.6s. Also seen on the resume: iteration 19 restored .vloop/state/state.json as a GATE REWRITE of T12, because the released v0.7.0 still matches gate files by substring and T12's verify names state.json; the operator's uncommitted plan edit cost T12 one attempt, the driver kept the replaced gate from memory. Both gate defects are the planner's (origin plan).

**What would automate it.** the planner (or a gate review) running every gate against the plan's own run — a brief check that includes the running brief, a fixture that shares the stub's directory — before handing the plan over

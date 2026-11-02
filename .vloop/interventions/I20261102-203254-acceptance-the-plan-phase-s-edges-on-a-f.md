---
id: I20261102-203254-acceptance-the-plan-phase-s-edges-on-a-f
brief: B20261102-1844-b31-i110-i112-plan-phase-edges.loop-brief
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
recorded: 2026-11-02T20:32:54Z
---
Acceptance: the plan phase's edges on a fresh template clone with a binary built from the branch

**Trigger.** the brief's acceptance, at close

**Done.** clone url-shortener-tests/b31-plan-edges with a committed complete state/v2 plan of an earlier brief whose T1 refers to a removed file. 0.14.0: vloop doctor and vloop run --plan-only on the template brief refuse (✗ plan 1 problem(s)). Branch: the bare doctor passes it (✓ plan a complete plan of … — the next brief's plan replaces it); a bare vloop run (a resume) is still validated and refused; vloop run --plan-only on the template brief reset it and planned 9 tasks, with one gate-review round sending a gate back and a revision, $0.90, no revert warning and a clean tree (the guard runs, nothing strayed); vloop brief review on the template brief kept a report whose brief is the file name without .md, valid under the new pattern. V2's revert itself is exercised only by the driver's tests: a real fenced session gave it nothing to revert

**What would automate it.** a kit script that sets up the clone, plants the plan and runs both binaries

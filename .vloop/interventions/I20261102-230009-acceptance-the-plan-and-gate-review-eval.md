---
id: I20261102-230009-acceptance-the-plan-and-gate-review-eval
brief: B20261102-2101-b32-i84-gates-through-the-named-surface.loop-brief
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
recorded: 2026-11-02T23:00:08Z
---
Acceptance: the plan and gate-review evals, and gpt-6.1-terra tried on every Copilot phase

**Trigger.** the brief's acceptance, at close

**Done.** evals with the branch binary installed, one run each: plan-judges-through-named-surface, plan-checks-its-gates, language-es and the three gate-review cases 1.0 with the plugin, plan-valid-state 0.8 on a stale grader, the same 0.8 from main at 7de2855; recorded in plugin/evals/README.md (repeated --case flags keep only the last, so each pattern ran on its own). gpt-6.1-terra: every phase on Copilot exited 2 at the gate review; work, review and verify on Copilot with plan and gate review on Claude did 5 of 9 tasks and stalled on a sound block; brief review gave a valid report. Work, review and brief review qualify; plan and gate review do not; verify was not reached. Recorded in docs/design-notes/copilot-cli-experiments.md and on #84, which stays open

**What would automate it.** #81's model-qualifying eval runner

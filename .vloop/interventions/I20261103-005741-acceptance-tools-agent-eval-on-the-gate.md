---
id: I20261103-005741-acceptance-tools-agent-eval-on-the-gate
brief: B20261102-2301-b33-i81-agent-eval-runner.loop-brief
phase: verify
kind: ceremony
automatable: partly
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-11-03
recorded: 2026-11-03T00:57:40Z
---
Acceptance: tools/agent-eval on the gate-review cases, on Claude and on Copilot

**Trigger.** the brief's acceptance, at close

**Done.** vloop from the branch installed; go run ./tools/agent-eval --list read all 23 cases (two skipped, no phase); --agent claude --model sonnet --case 'gate-review-*': exit 0, 1.00, 1.00, 0.88, $1.85; --agent copilot --model gpt-6.1-terra: exit 0, 0.80, 1.00, 1.00, $1.75 (19.3 credits); recorded in docs/design-notes/copilot-cli-experiments.md and on #84; rc.1 reinstalled after

**What would automate it.** none needed: the tool is the automation

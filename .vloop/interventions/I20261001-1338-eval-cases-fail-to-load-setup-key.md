---
id: I20261001-1338-eval-cases-fail-to-load-setup-key
brief: B20261001-1025-vloop-skills.loop-brief
phase: verify
kind: verification-finding
automatable: yes
by: assistant
occurred: 2026-10-01
recorded: 2026-10-01T12:38:50Z
---
all 13 eval cases fail to load: their prompt.md declares setup, which the eval tool does not know

**Trigger.** Before running the full eval suite (up to $40), a single-case probe with --runs 1 --ablation none --max-cost-usd 1.

**Context.** B7's T8 wrote the 13 cases as prompt.md with 'setup: scaffold.sh'. claude plugin eval rejected every one ('unknown frontmatter key "setup"', allowed: schema_version, name, description, tags, plugins, runs, expected_outcome, model, max_turns, timeout_seconds, allowed_tools, artifact_publish, growthbook_overrides, append_system_prompt, env); scaffold_script lives in the case.yaml format. B7's brief said cases plant repos 'with a scaffold script' without pinning the format, and its structural check only required max_turns and allowed_tools. Nothing ran; $0 spent. Recorded as defect D20261001-1338.

**Suggested.** Probe one case before the whole suite; then convert the cases to the format that supports a scaffold, and make the structural check reject keys the tool does not know.

**Decided.** Pending: the fix is proposed to the operator.

**What would automate it.** A structural check that loads every case with the eval tool's own loader (or its key list), and a planner rule: a brief that names an external format pins its exact keys from the tool, not from memory.

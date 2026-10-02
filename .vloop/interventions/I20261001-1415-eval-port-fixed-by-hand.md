---
id: I20261001-1415-eval-port-fixed-by-hand
brief: B20261001-1025-vloop-skills.loop-brief
phase: verify
kind: repair
automatable: partly
by: assistant
occurred: 2026-10-01
recorded: 2026-10-01T19:16:11Z
---
the eval cases fixed by hand so claude plugin eval loads and scaffolds them

**Trigger.** After the setup-key finding, the operator approved the fix: case.yaml scaffolds, deterministic graders, a stricter structural check, then a probe.

**Context.** Three porting defects surfaced one after another, each only by asking the eval tool itself: (1) setup: is not a prompt.md key (scaffolds live in case.yaml context.scaffold_script); (2) allowed_tools must be a YAML array, and command is not a grader type (regex, tool_used, tool_order, file_exists, llm, baseline); (3) the ported gates ran python3 -m pytest with no pytest installed, where the calibration used uv run pytest with a dev dependency. Fixed on the branch; the structural check now rejects all three shapes (five new fixtures); the tool loads all 13 cases at a $0 ceiling; all 13 scaffolds pass by hand with an empty HOME. The tool also requires the operator to grant gated tools at run time (--allow-tools).

**Suggested.** Probe one case before the suite; load with the tool's own loader at $0; run every scaffold by hand before spending.

**Decided.** The operator: 'yes, go ahead with the fix and then the evals'.

**What would automate it.** A structural check that runs the eval tool's loader (at a $0 ceiling) and every scaffold, in CI or as the close task's gate.

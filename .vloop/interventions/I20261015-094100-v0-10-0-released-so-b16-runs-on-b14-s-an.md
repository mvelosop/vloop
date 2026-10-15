---
id: I20261015-094100-v0-10-0-released-so-b16-runs-on-b14-s-an
brief: ""
phase: close
kind: ceremony
automatable: yes
by: both
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-15
recorded: 2026-10-15T08:41:00Z
---
v0.10.0 released so B16 runs on B14's and B15's changes

**Trigger.** B15 merged (#36); the installed vloop was 0.9.0, without B14's guidelines hand-off or B15's changes. The operator: cut v0.10.0 and install it in this repo.

**Done.** Release PR #37 (version in cmd/vloop/main.go and plugin.json; the plugin version pinned in e2e_test.go, run_harness_test.go, root_test.go), merged as a6c1ad5 after the operator's explicit go-ahead (auto mode had refused the first merge); annotated tag v0.10.0 on a6c1ad5, pushed; stamped go install (vloop version names commit a6c1ad5); vloop upgrade --yes in a follow-up PR (CLAUDE.md gains the plugin-not-loaded sentence). Not done: the public-install check.

**What would automate it.** vloop release (B17, T1 P7): bump, the pinned tests, tag, stamped build, upgrade.

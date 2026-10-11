---
id: I20261011-2054-v0-9-0-released-so-b13-runs-on-b12-s-fix
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
occurred: 2026-10-11
recorded: 2026-10-11T19:54:27Z
---
v0.9.0 released so B13 runs on B12's fixes

**Trigger.** B12 merged (#28); every brief in the B12–B17 series runs on the latest released vloop, so B13 needs B12 released. The operator: release 0.9.0, used only in this repository for now.

**Done.** Release PR #29 (version in cmd/vloop/main.go and plugin.json; the plugin version pinned in e2e_test.go, run_harness_test.go, root_test.go; B13's roadmap row gains doctor's gate-edit permission check), annotated tag v0.9.0 on 3cf595f, pushed; stamped go install (vloop version names commit 3cf595f); vloop upgrade --yes in a follow-up PR. Not done: the public-install check, by the operator's choice.

**Context.** Fifth release ceremony recorded (I20261004-1337, I20261005-0906, I20261006-1503, I20261006-1748). doctor after install: 0 problems; warnings for the pending upgrade and self-hosting (the binary is built from HEAD, which is the release commit).

**What would automate it.** vloop release (B17, T1 P7): bump, the pinned tests, tag, stamped build, upgrade.

---
id: I20261015-161603-v0-11-0-released-so-b17-runs-on-b16-s-ve
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
recorded: 2026-10-15T15:16:03Z
---
v0.11.0 released, so B17 runs on B16's verification phase

**Trigger.** B16 merged (#41); the installed vloop was 0.10.0, without the verify phase or the operator fixes. The operator: cut v0.11.0 and publish the release

**Done.** Release PR #42 (version in cmd/vloop/main.go and plugin.json; the version pinned in e2e_test.go, run_harness_test.go, root_test.go), squash-merged as 65239df; annotated tag v0.11.0 pushed; first GitHub release (v0.11.0) with notes; the module proxy resolves github.com/mvelosop/vloop@v0.11.0. The operator installed it with go install, updated the marketplace plugin, and ran vloop upgrade (this PR)

**What would automate it.** vloop release (B17, T1 P7): bump, the pinned tests, tag, GitHub release, public-install check, upgrade

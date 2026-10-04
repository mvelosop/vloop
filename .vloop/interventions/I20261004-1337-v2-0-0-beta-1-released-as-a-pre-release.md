---
id: I20261004-1337-v2-0-0-beta-1-released-as-a-pre-release
brief: ""
phase: next
kind: ceremony
automatable: partly
by: both
occurred: 2026-10-04
recorded: 2026-10-04T12:37:20Z
---
v2.0.0-beta.1 released as a pre-release so B10 runs under the gate model

**Trigger.** B9 merged; B10 gets the gate review and checks only from a released v2 build (series rule: run from the latest released vloop)

**Done.** The operator chose a pre-release (the assistant's recommendation of three: 2.0.0-beta.1 over rc.1 or running B10 under v1.0.0). Release PR #15 (the version in main.go, plugin.json and three pinned tests), annotated tag v2.0.0-beta.1 on 42ee0a6, stamped build into ~/go/bin, vloop upgrade --yes in PR #16. doctor: 0 problems, self-hosting passes.

**What would automate it.** a vloop release command doing the version bump, tag and stamped build

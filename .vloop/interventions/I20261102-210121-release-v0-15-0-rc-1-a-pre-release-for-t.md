---
id: I20261102-210121-release-v0-15-0-rc-1-a-pre-release-for-t
brief: ""
phase: next
kind: ceremony
automatable: partly
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-11-02
recorded: 2026-11-02T21:01:21Z
---
Release v0.15.0-rc.1, a pre-release for the operator's own testing

**Trigger.** the operator asked for a pre-release to test after #75 was done

**Done.** go run ./tools/release prepare v0.15.0-rc.1 (PR #119, squash-merged with --admin on the operator's word), publish v0.15.0-rc.1 (tag on 7de2855, GitHub pre-release, public install proved); installed it: go install …@v0.15.0-rc.1, claude plugin marketplace update vloop and claude plugin update vloop@vloop, vloop upgrade --yes here (the CLAUDE.md section gains the create-brief and review-brief sentence)

**What would automate it.** the release tool installing and upgrading after publish

---
id: I20261109-194753-release-v0-15-0-rc-3-and-this-repository
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
occurred: 2026-11-09
recorded: 2026-11-09T19:47:53Z
---
release v0.15.0-rc.3 and this repository's upgrade

**Trigger.** rc.2 could not be installed (#140); B38 fixed it; the operator's go-ahead for rc.3

**Done.** prepare opened #146 (body replaced with notes since rc.1, as rc.2 was never installable); the operator merged it; publish tagged, released and its public install passed; release notes replaced with B32-B38; installed rc.3; vloop upgrade --yes here (CLAUDE.md section, the marketplace in .claude/settings.json, install.json)

**Context.** PR #146, tag v0.15.0-rc.3

**What would automate it.** release notes from the previous installable release, not the previous tag

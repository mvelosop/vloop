---
id: I20261125-214929-release-v0-15-0-rc-4-and-this-repository
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
occurred: 2026-11-25
recorded: 2026-11-25T21:49:29Z
---
release v0.15.0-rc.4 and this repository's upgrade

**Trigger.** B39 merged (#155) and the Windows preparation PR (#157); the operator's go-ahead to merge and release

**Done.** prepare opened #158, merged on the operator's go-ahead; publish tagged, released and its public install passed; installed rc.4; vloop upgrade --yes here (CLAUDE.md section, .gitignore gains .vloop/config.local.toml, install.json)

**Context.** PR #158, tag v0.15.0-rc.4; first release prepared on Windows, its checks in pwsh via go run ./tools/check

**What would automate it.** vloop upgrade could print the diff it would apply instead of only the file names

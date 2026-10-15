---
id: I20261015-151132-base-check-halted-on-a-github-pr-extensi
brief: B20261015-1001-b16-verification-phase.loop-brief
phase: halt
kind: repair
automatable: yes
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-15
recorded: 2026-10-15T14:11:32Z
---
Base check halted on a GitHub PR extension key in .git/config; ignored via config and re-run

**Trigger.** vloop run refused before planning: 'the check go changed .git/config (branch.B20261015-1001-b16-verification-phase.github-pr-base-branch) — nothing was committed'; VS Code's GitHub Pull Requests extension wrote it, not the check

**Done.** Added run.git-config-ignore = ["branch.*.github-pr-*"] to .vloop/config.toml (commit 744f173, by hand after vloop config set dropped the file's comments), re-ran; nothing had been spent. Later made built-in and the repo line removed (a33865f)

**Context.** First run 20261015-093250 refused at 'checking the base'; recorded as defects D20261015-151118 (guard) and D20261015-151119 (config set)

**What would automate it.** the guard's built-in patterns now include branch.*.github-pr-* (a33865f)

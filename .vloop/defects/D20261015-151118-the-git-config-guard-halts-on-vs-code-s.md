---
id: D20261015-151118-the-git-config-guard-halts-on-vs-code-s
brief: B20261008-2144-b12-field-reliability.loop-brief
origin: brief
found-by: operator
kind: spec-gap
severity: medium
status: fixed
fixed-by: a33865f
case: cmd/vloop/run_git_test.go TestRunGitConfigGitHubPRKeysIgnored
created: 2026-10-15T14:11:18Z
---
The git-config guard halts on VS Code's GitHub Pull Requests branch keys (github-pr-base-branch, github-pr-owner-number), which only vscode-merge-base was built in for

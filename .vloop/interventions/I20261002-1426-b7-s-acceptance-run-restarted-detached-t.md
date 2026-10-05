---
id: I20261002-1426-b7-s-acceptance-run-restarted-detached-t
brief: B20261001-1025-vloop-skills.loop-brief
phase: verify
kind: verification-finding
automatable: partly
by: both
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-02
recorded: 2026-10-02T13:26:10Z
---
B7's acceptance run: restarted detached, then verified by hand

**Trigger.** The operator approved the url-shortener acceptance run at a $40 ceiling

**Done.** Set up in ~/source/personal/url-shortener-vloop-acceptance (the sample at 9a197a7, shell loop and its CLAUDE.md block removed, vloop init --stacks typescript, the brief converted with its body unchanged, the folder trusted by the operator). The first start ran as a session background task capped at 2 hours; it was stopped in planning, its branch, run folder and .running lock removed, and restarted detached with nohup. Result: complete, 10/10 first-pass in 10 iterations, $4.21 (plan $1.98 opus; work $1.39 and review $0.84 sonnet), against the shell loop's 10/10 in 11 iterations, 1 rejection, about $19.86. Every record validates (state, 21 sessions, 20 reports, 10 iterations, metrics); every task has a kind. On a fresh clone: npm ci, 3 unit and 26 e2e tests pass, the build boots, the committed smoke script passes, and the brief's worked example holds over HTTP. Wall time 105 min against 23 min of sessions: the Mac idle-slept at 12:43 and advanced only in maintenance wakes until 14:22. Metrics undercount tests: D20261002-1425.

**What would automate it.** vloop run holding a no-idle-sleep assertion on macOS (caffeinate -i) for the run's duration; a run that cannot outlive the operator's session started detached by default

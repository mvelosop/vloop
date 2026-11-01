---
id: I20261031-144252-the-default-node-had-no-pnpm-pnpm-10-33
brief: B20261031-1151-b25-i79-copilot-verify.loop-brief
phase: setup
kind: repair
automatable: partly
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-31
recorded: 2026-10-31T14:42:52Z
---
The default Node had no pnpm; pnpm 10.33.0 installed for the URL shortener acceptance

**Trigger.** nvm's default (lts/*) moved to v24.22.1, which has no pnpm; the app check would fail after the first task (a Copilot-driven run's base check printed 'pnpm: command not found')

**Done.** npm install -g pnpm@10.33.0 into v24.22.1, the version B24's build used

**What would automate it.** doctor running each check's first command word through the gate shell and warning when it is not found

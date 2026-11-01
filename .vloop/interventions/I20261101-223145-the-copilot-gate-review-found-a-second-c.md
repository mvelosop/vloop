---
id: I20261101-223145-the-copilot-gate-review-found-a-second-c
brief: B20261101-2044-b26-i78-copilot-gate-review.loop-brief
phase: halt
kind: decision
automatable: partly
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-11-01
recorded: 2026-11-01T22:31:45Z
---
The Copilot gate review found a second check gap; the smoke check replaced by a clean-install check

**Trigger.** acceptance run 2 halted, exit 2: the gate review (gpt-5.6-terra) found that the brief's closing command, from a clean install, is held by no check

**Done.** replaced the smoke check with a clean check (fresh node_modules, pnpm install --frozen-lockfile, build, test, test:e2e, then pnpm smoke once the script exists), committed it, re-ran; a third gap would go to the operator

**What would automate it.** none needed: the halt names the check; the operator writes it

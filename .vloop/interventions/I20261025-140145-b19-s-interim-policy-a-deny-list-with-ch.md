---
id: I20261025-140145-b19-s-interim-policy-a-deny-list-with-ch
brief: B20261025-1307-b19-copilot-review.loop-brief
phase: design
kind: decision
automatable: no
by: both
schema: intervention/v2
options: 3
recommended: 1
decided: 1
agreement: recommended
occurred: 2026-10-25
recorded: 2026-10-25T14:01:45Z
---
B19's interim policy: a deny list with cheap guards until B20, Copilot for review only, proven on the real CLI at close

**Trigger.** B19 lets Copilot review sessions run before B20 builds the fence for non-Claude agents

**Done.** Wrote B19 and its roadmap row

**Context.** Copilot CLI flags checked against GitHub's CLI command reference on 2026-10-25: --plugin-dir, --effort, --deny-tool Kind(argument), --disable-builtin-mcps, COPILOT_CUSTOM_INSTRUCTIONS_DIRS, auth in the system credential store

**Options.**
1. A denylist plus cheap guards (no gh, push, remote, network, global git config, web fetch, built-in GitHub MCP; experimental warning; no release between B19 and B20)
2. An allowlist of the commands the review skill uses
3. No Copilot session before B20's fence

**Recommended.** The driver's B8 checks already undo or catch in-repo damage; the outward actions are few and nameable; the allowlist pilot denied printf and the reviewer never saw the diff; B20 follows right away

**Decided.** Denylist with the guards; C5 proves the deny patterns on the real CLI; the sandbox is left to B20

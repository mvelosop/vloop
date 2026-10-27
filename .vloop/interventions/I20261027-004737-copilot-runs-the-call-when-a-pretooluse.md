---
id: I20261027-004737-copilot-runs-the-call-when-a-pretooluse
brief: B20261026-1002-b22-fence-second-review.loop-brief
phase: verify
kind: verification-finding
automatable: no
by: assistant
origin: env
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-27
recorded: 2026-10-27T00:47:36Z
---
Copilot runs the call when a preToolUse hook prints non-JSON or times out; denies when it is missing or exits non-zero

**Trigger.** B22.2 hook failure probe on the real CLI

**Done.** recorded in copilot-cli-experiments.md (B22.2) and fence-audit.md; the deny list stays as the backstop

**Context.** Copilot CLI 1.0.103, gpt-5.6-terra; first probe used rm, which Copilot's deny list refused, so the kit's call was changed

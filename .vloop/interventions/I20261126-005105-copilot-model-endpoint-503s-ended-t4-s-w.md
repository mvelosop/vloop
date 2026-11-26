---
id: I20261126-005105-copilot-model-endpoint-503s-ended-t4-s-w
brief: B20261125-2207-b40-i156-mixed-os-pwsh.loop-brief
phase: halt
kind: halt
automatable: partly
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-11-26
recorded: 2026-11-26T00:51:04Z
---
Copilot model endpoint 503s ended T4's work session; resumed once

**Trigger.** exit 7 in iteration 5 (T4 work): the copilot process log shows HTTP 503 from /responses six times between 00:00:55 and 00:02:09 UTC; is_error true, no attempt charged

**Done.** checked refs and tree (clean, T4's partial edits in the halt commit), resumed vloop run once; T4 passed in the resumed run

**Context.** run 20261125-230701, session 011-work.json, model mai-code-2

**What would automate it.** the driver could retry a session that failed on provider 5xx responses, once, after a pause

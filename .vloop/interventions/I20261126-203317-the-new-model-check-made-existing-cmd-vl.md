---
id: I20261126-203317-the-new-model-check-made-existing-cmd-vl
brief: B20261126-1818-b46-i153-copilot-model-ids.loop-brief
phase: verify
kind: verification-finding
automatable: partly
by: operator
origin: work
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-11-26
recorded: 2026-11-26T20:33:17Z
---
the new model check made existing cmd/vloop end-to-end tests that set a Copilot model call the real Copilot API (with a fake token, or the developer's own gh login from the system keychain), against the brief's rule that no test reaches it; most of those tests skip on Windows; fixed by hand in 9baa81a9 with a local stub in TestMain that cleanEnv points every test binary at

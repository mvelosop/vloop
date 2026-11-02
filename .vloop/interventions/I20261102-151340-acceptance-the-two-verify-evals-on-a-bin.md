---
id: I20261102-151340-acceptance-the-two-verify-evals-on-a-bin
brief: B20261102-1332-b30-i113-verify-unrunnable-lines.loop-brief
phase: verify
kind: ceremony
automatable: partly
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-11-02
recorded: 2026-11-02T15:13:40Z
---
Acceptance: the two verify evals on a binary built from the branch

**Trigger.** the brief's acceptance section, at close

**Done.** claude plugin eval plugin --case 'verify-*' with the branch binary first on PATH (not go install, which would replace the installed release): verify-planted-mismatches 1.00 with, 0.00 without; verify-unrunnable-line 1.00 with, 0.00 without; $0.79, 146 s; recorded in plugin/evals/README.md. The one run-time line no test asserts (findings and lines not run together) checked by hand with a throwaway test

**What would automate it.** a kit script running the suite's cases a brief names against a branch build

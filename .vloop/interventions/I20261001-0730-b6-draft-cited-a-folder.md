---
id: I20261001-0730-b6-draft-cited-a-folder
brief: B20261001-0723-vloop-run-driver.loop-brief
phase: design
kind: verification-finding
automatable: yes
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-01
recorded: 2026-10-01T06:26:05Z
---
the B6 draft cited .loop/tests/scenarios/ as a binding reference, against the convention just agreed

**Trigger.** vloop brief check on the draft: a directory with no entry point

**Done.** cited the harness file instead; the scenarios stay named one by one in the contract

**Context.** vloop's checker on the B6 draft: 'binding reference directory has no entry point: .loop/tests/scenarios/' — a folder cited against the convention agreed the day before. Cited the harness file, with the scenarios named in the contract.

**Suggested.** cite .loop/tests/lib.sh instead.

**Decided.** the assistant, within the operator skill's remit; no operator decision asked.

**What would automate it.** already a check; running vloop's checker in the design act caught it

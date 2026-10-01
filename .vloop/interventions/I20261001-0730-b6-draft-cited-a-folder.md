---
id: I20261001-0730-b6-draft-cited-a-folder
brief: B20261001-0723-vloop-run-driver.loop-brief
phase: design
kind: verification-finding
automatable: yes
by: assistant
occurred: 2026-10-01
recorded: 2026-10-01T06:26:05Z
---
the B6 draft cited .loop/tests/scenarios/ as a binding reference, against the convention just agreed

**Trigger.** vloop brief check on the draft: a directory with no entry point

**Done.** cited the harness file instead; the scenarios stay named one by one in the contract

**What would automate it.** already a check; running vloop's checker in the design act caught it

---
id: I20261001-0724-readme-shell-default-wrong-since-b2
brief: B20261001-0723-vloop-run-driver.loop-brief
phase: design
kind: verification-finding
automatable: yes
by: assistant
occurred: 2026-10-01
recorded: 2026-10-01T06:26:05Z
---
the README states cmd as the Windows shell default; the code and B2's brief say pwsh

**Trigger.** reading the README's config table during B6's survey

**Done.** recorded as an escaped defect against B2; fixed in B6 as F1, with a test comparing the README's stated default with the code's

**What would automate it.** a test checking every default the README states against config.Keys

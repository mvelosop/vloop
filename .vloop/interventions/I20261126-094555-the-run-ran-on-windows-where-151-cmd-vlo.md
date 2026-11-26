---
id: I20261126-094555-the-run-ran-on-windows-where-151-cmd-vlo
brief: B20261125-2207-b40-i156-mixed-os-pwsh.loop-brief
phase: verify
kind: verification-finding
automatable: partly
by: operator
origin: env
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-11-26
recorded: 2026-11-26T09:45:53Z
---
the run ran on Windows, where 151 cmd/vloop end-to-end tests skip, so its check passed without them; on macOS 5 failed: four pinned doctor output that predates the checks-shell warning, and the B6 real-data test inherited this repository's new platforms key. Fixed by hand in e6540904

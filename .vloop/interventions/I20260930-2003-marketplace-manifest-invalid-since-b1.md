---
id: I20260930-2003-marketplace-manifest-invalid-since-b1
brief: B20260930-2007-vloop-init-upgrade-doctor.loop-brief
phase: design
kind: verification-finding
automatable: yes
by: assistant
occurred: 2026-09-30
recorded: 2026-09-30T19:09:24Z
---
the marketplace manifest has failed claude plugin validate since B1

**Trigger.** the design act's survey ran claude plugin validate on the repo for the first time

**Done.** recorded as an escaped defect against B1; B5 fixes the manifests and gates on the real validate

**What would automate it.** a gate running claude plugin validate from the first brief that ships a manifest

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

**Context.** B5's survey ran `claude plugin validate .` for the first time: 'owner: Invalid input: expected object, received undefined'. The marketplace manifest had shipped invalid since B1, whose brief never mentioned owner. Recorded as an escaped defect; B5 fixed the manifests.

**Suggested.** record it and fix it in B5.

**Decided.** the operator accepted it with the B5 brief.

**What would automate it.** a gate running claude plugin validate from the first brief that ships a manifest

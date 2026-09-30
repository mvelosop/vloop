---
id: I20260928-2215-the-loop-s-install-proof-failed-in-every
brief: ""
phase: setup
kind: repair
automatable: yes
by: assistant
occurred: 2026-09-28
recorded: 2026-09-30T16:54:00Z
backfilled: true
---
the loop's install proof failed in every installed copy

**Trigger.** two scenarios could pass only inside the loop's source repo, and the installer's warning about a stale .loop/todo/ lived in the wrong loop

**Done.** fixed both in an-autonomous-loop-3 (scenarios 26 and 43, install.sh), reinstalled

**What would automate it.** a tooling defect, fixed once; the install proof itself is the automation

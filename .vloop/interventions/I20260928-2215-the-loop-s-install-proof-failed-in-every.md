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

**Context.** Before B1, installing the loop from an-autonomous-loop-3 into vloop: the install proof reported `FAIL 26-install-boundary` and `FAIL 43-brief-already-run` in /tmp/install-suite.log. Both scenarios read files that exist only in the loop's source repo; the installer's stale `.loop/todo/` warning lived inside the copy loop. Fixed in the source repo (commit 6fb8a47) and reinstalled; the suite then passed 46/46.

**Suggested.** fix both scenarios to build their own fixtures, and move the warning out of the copy loop.

**Decided.** the operator approved: fix them in the source repo, commit, reinstall.

**What would automate it.** a tooling defect, fixed once; the install proof itself is the automation

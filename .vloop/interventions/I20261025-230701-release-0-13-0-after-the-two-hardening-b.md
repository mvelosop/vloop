---
id: I20261025-230701-release-0-13-0-after-the-two-hardening-b
brief: B20261025-2302-b21-records-and-verify.loop-brief
phase: design
kind: decision
automatable: no
by: both
schema: intervention/v2
options: 2
recommended: 1
decided: 1
agreement: recommended
occurred: 2026-10-25
recorded: 2026-10-25T23:06:59Z
---
Release 0.13.0 after the two hardening briefs, before the fence

**Trigger.** runs use the latest release, and no release was set until the fence: the hardening fixes would reach no run until after it

**Context.** roadmap B20-B23

**Options.**
1. Release 0.13.0 after B21, Copilot review still experimental and warned
2. Keep no release until the fence

**Recommended.** the fence brief and every run after it get the fixes

**Decided.** release after B21; the roadmap says so

---
id: I20260930-0935-vloop-s-checker-rejected-two-binding-ref
brief: B20260930-0929-vloop-close-export-workspace.loop-brief
phase: design
kind: verification-finding
automatable: yes
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-09-30
recorded: 2026-09-30T16:54:00Z
backfilled: true
---
vloop's checker rejected two binding references citing two paths each

**Trigger.** running vloop brief check on the draft, besides the shell checker

**Done.** split into one path per entry

**Context.** Checking the B4 draft with `vloop brief check` (besides the shell checker): two binding-reference entries cited two paths each, which vloop's one-path-per-entry rule rejects. Split into one path per entry.

**Suggested.** split the entries.

**Decided.** the assistant, within the operator skill's remit; no operator decision asked.

**What would automate it.** already a check; run it as part of the design act

---
id: I20261025-223644-the-verify-session-s-scratch-repository
brief: B20261025-1307-b19-copilot-review.loop-brief
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
occurred: 2026-10-25
recorded: 2026-10-25T22:36:42Z
---
the verify session's scratch repository under .vloop/tmp resolves to this repository's .vloop (C-2), so its config edits hit the real config and it wrote no verify.json, as in B18 (verify skill)

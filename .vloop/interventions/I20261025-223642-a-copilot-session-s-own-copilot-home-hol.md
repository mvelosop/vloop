---
id: I20261025-223642-a-copilot-session-s-own-copilot-home-hol
brief: B20261025-1307-b19-copilot-review.loop-brief
phase: verify
kind: verification-finding
automatable: partly
by: operator
origin: brief
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-25
recorded: 2026-10-25T22:36:42Z
---
a Copilot session's own COPILOT_HOME holds no login, so Copilot logs in through 'gh auth token'; the brief assumed the credential store, and doctor and the preflight do not check the GitHub CLI's login (C5.2)

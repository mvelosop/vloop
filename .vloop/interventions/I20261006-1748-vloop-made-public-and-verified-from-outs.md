---
id: I20261006-1748-vloop-made-public-and-verified-from-outs
brief: ""
phase: next
kind: ceremony
automatable: partly
by: both
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-06
recorded: 2026-10-06T16:48:47Z
---
vloop made public and verified from outside

**Trigger.** v0.8.0 released, the claude.ai share link revoked and the pre-public tags deleted; the operator gave the go-ahead

**Done.** gh repo edit --visibility public. Verified with a clean Go cache, the public proxy and no credentials: the proxy lists v0.7.0 and v0.8.0, go install …@latest installs 0.8.0; in a throwaway home, claude plugin marketplace add mvelosop/vloop and install vloop@vloop give 0.8.0 enabled, and vloop doctor shows ✓ plugin. The operator's questions about the plugin led to triage finding 1 (docs/design-notes/vloop-t1-triage-inputs.md).

**What would automate it.** a release command that also checks the public install path

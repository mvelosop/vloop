---
id: I20261105-232221-doc-fix-for-verify-s-finding-by-session
brief: B20261105-2002-b37-i88-metrics-by-session-kit-without-python.loop-brief
phase: verify
kind: verification-finding
automatable: partly
by: assistant
origin: work
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-11-05
recorded: 2026-11-05T23:22:21Z
---
doc fix for verify's finding: --by session is refused with --interventions

**Trigger.** verify's one surface finding: docs/guide/metrics.md named only --by task as refused in the --interventions section

**Done.** edited the sentence to name --by task and --by session; committed on the run branch

**Context.** reports/verify.json of run 20261105-212928

**What would automate it.** a work task that greps the guide for every refusal the flag gains

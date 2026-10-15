---
id: I20261015-151134-bare-close-shows-the-verify-report-dry-r
brief: B20261015-1001-b16-verification-phase.loop-brief
phase: verify
kind: decision
automatable: no
by: both
schema: intervention/v2
options: 1
recommended: 1
decided: 1
adjusted: true
agreement: adjusted
occurred: 2026-10-15
recorded: 2026-10-15T14:11:32Z
---
Bare close shows the verify report; --dry-run without a flag still refuses

**Trigger.** close --dry-run refused for --finding/--no-findings before printing the verify report, so the new operate step ('--dry-run prints them first') could not be followed

**Context.** Found verifying B16 in a throwaway clone; fixed in a33865f; operate skill step 2 and the repo playbook updated

**Options.**
1. close prints the verify report before refusing for the findings flag, with or without --dry-run

**Recommended.** the operator reads the findings before choosing the flag, and it is a few lines

**Decided.** Only a bare close (no flag at all) prints the report and refuses; --dry-run without a flag refuses as before, since what it previews depends on the flag; dry-run is then used with the chosen flag before the real close

**What would automate it.** none: a design decision

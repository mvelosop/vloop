---
id: I20261005-0757-b10-verified-the-worked-example-held-by
brief: B20261004-1434-interventions-options.loop-brief
phase: verify
kind: verification-finding
automatable: partly
by: both
schema: intervention/v2
options: 3
recommended: 1
decided: 1
agreement: recommended
occurred: 2026-10-05
recorded: 2026-10-05T06:57:19Z
---
B10 verified: the worked example held by hand; two findings fixed test first, two carried to B11; the operate eval run once

**Trigger.** vloop run exited 0 at 11/11, $13.48; the operator skill's verification

**Done.** An independent session ran every worked-example line with a fresh build: all matched. Found: set adjusted out of I1's order, --by phase ignored in the workspace table (both fixed, test first), two unquoted set refusals exiting 1, cosmetic show and dry-run output (carried to B11). The operate eval scored 0.5 with judges the operator's reading contradicts (carried).

**Context.** Run B20261004-1434-interventions-options. Defects D20261005-0756-vloop-intervention-set-adjusted-true-wro and D20261005-0756-the-workspace-interventions-table-ignore fixed; D20261005-0757-operate-proposes-options-scored-0-5-its open.

**Options.**
1. fix the two findings that break a sentence of the brief, carry the other two to B11
2. fix all four now
3. carry all four to B11 and merge as is

**Recommended.** the two that break the brief are cheap and local; exit codes and output polish are B11's

**Decided.** the operator took the recommendation, and the one-run operate eval

**What would automate it.** the verification agent's table is mechanical; deciding what to fix before merge is the operator's

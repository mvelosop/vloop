---
id: I20261004-2304-b10-blocked-on-t9-a-gate-needed-a-sessio
brief: B20261004-1434-interventions-options.loop-brief
phase: halt
kind: repair
automatable: partly
by: both
schema: intervention/v2
options: 3
recommended: 1
decided: 1
agreement: recommended
occurred: 2026-10-04
recorded: 2026-10-04T22:04:22Z
---
B10 blocked on T9: a gate needed a session to write .vloop/interventions/; the README section and the migration done by hand

**Trigger.** vloop run exited 2 at 9/11, $12.82: T9 blocked after three gate failures (README-interventions.md does not mention intervention/v2); every attempt's edit restored by the driver (S-4); T10 waited on T9

**Done.** Wrote the README's v2 section by hand, ran vloop intervention migrate (77 records, frontmatter only) and the index script with a build of the branch; T9's and T10's gates pass unchanged; reset T9 and T10 so each still gets a work session and a review; resumed.

**Context.** Run B20261004-1434-interventions-options, iterations 12-14, T9 and T10. The brief's I9 and real-data check put work in .vloop/interventions/, which S-4 keeps from sessions; the gate review passed both gates. See D20261004-2304-the-brief-asked-sessions-to-write-under for the brief's gap.

**Options.**
1. do the .vloop/interventions/ part by hand, keep both gates, reset T9 and T10 and resume
2. amend T10's gate to check migrate --dry-run and migrate at close
3. let the run go on and handle it all at verification

**Recommended.** the gates judge the right result and pass once the operator does the one act sessions may not; nothing is amended

**Decided.** the operator approved the repair as proposed

**What would automate it.** the gate review rejecting a gate that needs a session to write under .vloop/

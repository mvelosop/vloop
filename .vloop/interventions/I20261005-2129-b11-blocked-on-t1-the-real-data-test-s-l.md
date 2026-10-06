---
id: I20261005-2129-b11-blocked-on-t1-the-real-data-test-s-l
brief: B20261005-0933-quality-pass.loop-brief
phase: halt
kind: repair
automatable: partly
by: both
schema: intervention/v2
options: 3
recommended: 1
decided: 1
agreement: recommended
occurred: 2026-10-05
recorded: 2026-10-05T20:29:16Z
---
B11 blocked on T1: the real-data test's leak detection written by hand, checking only the paths the run wrote

**Trigger.** vloop run exited 2 at 0/17, $13.40: T1 blocked after three review failures; six tasks wait on it

**Done.** Wrote the narrow check in TestWorkedExampleB6RealData: HEAD, refs and hooks, plus every path the run wrote in its clone compared before and after; proven with a planted leak and a planted unrelated edit; reset T1 so it still gets a work session, its gate, the check and a review; resumed.

**Context.** Run B20261005-0933-quality-pass, iterations 1-3, T1. Attempt 1 dropped write detection; attempt 2 changed no test code yet its proposal claimed a hooks comparison, and the review caught it; attempt 3 kept refs and hooks only.

**Options.**
1. write the narrow check by hand, prove it with a planted leak, reset T1 and resume
2. the same, and replace T1's gate so it judges the clause too
3. rule the brief's narrower Q1 enough and accept attempt 1 by editing the plan's acceptance by hand

**Recommended.** it satisfies both clauses precisely, and T1 still passes through its gate, the check and a review

**Decided.** the operator chose option 1

**What would automate it.** a brief that says both halves of a tolerance-and-detection requirement; a work skill that proposes the narrow comparison

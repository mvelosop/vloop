---
id: I20261008-2146-b12-s-forks-settled-a-key-aware-git-conf
brief: B20261008-2144-field-reliability.loop-brief
phase: design
kind: decision
automatable: partly
by: both
schema: intervention/v2
options: 3
recommended: 1
decided: 1
agreement: recommended
occurred: 2026-10-08
recorded: 2026-10-08T20:46:35Z
---
B12's forks settled: a key-aware .git/config guard with an allowlist, acceptance that resumes, Bash denials with their command, run visibility in the log and status

**Trigger.** The design act for B12 from the T1 triage's P1; two read-only surveys of the guard, acceptance, session timing, denials, masking, the run log, status, version and metrics

**Done.** The operator took the recommendation on all four forks: the guard compares parsed keys with a built-in editor allowlist plus run.git-config-ignore; a halt during acceptance resumes it with the same plan, and work needs a passed gate review; a denied Bash call keeps its command, masked, redacted and cut to 200 characters; a line per session, ceiling warnings at 50% and 80%, and vloop status from the run lock. The assistant's choices to review: session time from a driver-measured wall_ms (optional field, no schema bump) and wall from the first plan session; the version from the build info; n/a for unmeasured size; the dashed -Users-<name>- mask. Brief drafted: docs/briefs/B20261008-2144-field-reliability.loop-brief.md

**Context.** Design branch design/t1-triage; defects D20261005-1125 (both) and, in visum-monorepo, D20261007-1243 and D20261007-1451 (four)

**Options.**
1. compare parsed keys; ignore a built-in editor list plus a run.git-config-ignore key; every other key halts, named
2. the same comparison with the list in the binary only
3. halt only on known-dangerous keys (core, filter, include, hooks)

**Recommended.** an unknown dangerous key still halts, and a repository can add its own editor's keys without a release

**Decided.** option 1, and the recommendations on acceptance, denials and visibility

**What would automate it.** the survey and the options can be generated; the choice is the operator's

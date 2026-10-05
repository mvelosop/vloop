---
id: I20261002-2140-b8-designed-from-the-v1-0-review-securit
brief: B20261002-2135-vloop-v1-security.loop-brief
phase: design
kind: decision
automatable: partly
by: both
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-02
recorded: 2026-10-02T20:40:42Z
---
B8 designed from the v1.0 review: security pass now, quality pass as B9

**Trigger.** The operator asked to design B8, the v1.0 review, with B7's carry-overs

**Done.** Surveyed the roadmap, horizon, design session section 6, open defects and the domain. Five read-only review passes (command execution; file writes, paths and secrets; what sessions may do; docs against the binary; code health) found about 45 issues; the architect checked the load-bearing claims in the code. Decided with the operator: review in the design act, brief the fixes; B8 security and correctness plus carry-overs, B9 quality, v1.0 after B9; the driver as the boundary with a tighter fence, auto mode kept; gate timeout 15 min, session 60; refuse a dirty tree; keep-awake on by default on all three OSes; the README cap kept; B8 run from a released v0.7.0. The architect's own choices for review: the README completeness test moves to the guides; repository hooks do not run on the driver's commits; operator plan edits allowed on resume. Recorded: the brief (F1–F16, C1–C3), docs/design-notes/vloop-v1-review.md, 16 defects with blame, the roadmap's B8 and B9 rows.

**What would automate it.** a review command that runs the five passes as parallel read-only sessions and writes a findings table for the design act

---
id: I20261001-2035-llm-grader-focus-fixed
brief: B20261001-1025-vloop-skills.loop-brief
phase: verify
kind: repair
automatable: partly
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-01
recorded: 2026-10-01T19:35:00Z
---
every llm grader given the input it judges; the structural check rejects one that reads what it cannot see

**Trigger.** The blocked probe's JSON showed the review cases' `names-the-defect` graders at `focus: last_message` while their text says to read `.vloop/tmp/verdict.json`.

**Context.** The eval tool's reference: an llm judge sees only its focus (`last_message` by default, `trace`, or one `{source: file, path}`) and has no workspace access; combining inputs is undocumented. Fifteen llm graders read files or the transcript without a focus, so a correct review could score 0. Fixed: nine review graders focus the verdict; `prose-in-spanish` and `verify-commands-not-translated` the plan; `fixture-read-back` and `ran-the-gates-on-the-base` the trace; `plan-quality` focuses the plan with the brief's three pinned behaviours written into it; `implemented-properly` split into one grader on `bin/greet` and `dispute-quotes-evidence` on the proposal (weight 2 kept). The structural check rejects an llm grader that reads a backticked path other than its file focus, or judges the transcript without `focus: trace` (three fixtures). Also: a missing word in both `gate-untouched` graders. Checked: the full toolchain; the tool loads all 13 cases with the real grants at a $0 ceiling with no warnings.

**Suggested.** The assistant: focus each grader on its one input, split where a grader needs two, and make the structural check reject the shape, test first.

**Decided.** The operator: 'yes, go ahead with the grader fix'.

**What would automate it.** The structural check now does; the eval tool itself could warn when an llm grader's text names a file outside its focus.

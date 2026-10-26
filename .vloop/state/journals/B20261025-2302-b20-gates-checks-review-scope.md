# Journal — B20261025-2302-b20-gates-checks-review-scope

Append-only narrative of this plan. Rendered state lives in .vloop/state/plan.md.

## Plan — B20261025-2302-b20-gates-checks-review-scope

- **Brief:** `docs/briefs/B20261025-2302-b20-gates-checks-review-scope.loop-brief.md`
- **Tasks:** 6

## T1 — State P-8 in the domain model and add gate-verdict/v2 beside an unchanged v1

- **Outcome:** done (review: PASS)
- **Summary:** P-8 is stated under Plans and tasks in the domain model; gate-verdict/v2 exists beside an untouched v1 with the extra kinds uncovered and check-gap, with valid/invalid testdata covered by the schema tests.
- **Files:** docs/domain/domain-model.md, schemas/gate-verdict.v2.json, internal/schema/schema_test.go, internal/schema/testdata/gate-verdict2.valid.json, internal/schema/testdata/gate-verdict2.invalid.json, embed_test.go, internal/cli/schema_test.go, cmd/vloop/b2_e2e_test.go
- **Notes for next iteration:** Tests that pin the embedded schema set changed for the new gate-verdict/v2: schema_test.go TestNames, embed_test.go entry count (19->20), internal/cli TestSchemaList (list and count), cmd/vloop TestWorkedExampleB2Commands schema list line. The internal/cli TestTaskGate* tests fail when run inside a loop session ("plan was changed during this session"), which is the session's own state, not this change.

## T2 — Halt the run with exit 2 when the gate verdict holds a check-gap finding, keeping nothing

- **Outcome:** check_failed (review: skipped)
- **Summary:** The driver now validates the gate verdict as gate-verdict/v2 and halts with exit 2 on any check-gap finding, printing every finding, sending nothing back and leaving no plan, gates or commit.
- **Files:** internal/driver/gatereview.go, internal/driver/plan.go, internal/driver/gatereview_test.go
- **Notes for next iteration:** gateReviewRound now returns (problems, check-gap count, err); acceptPlan checks gaps before the planRounds test, so round 1 and 2 behave alike, and removes state.json, gates and plan.md. New unit tests cover counting gaps, uncovered not counting, v1 rejected and the halt line; the round 1/2 halt end to end is in gate.sh only. go test ./internal/cli had 5 TestTaskGate* failures ('plan was changed during this session') inside this session, which looks environmental and was not investigated.

## T2 — Halt the run with exit 2 when the gate verdict holds a check-gap finding, keeping nothing

- **Outcome:** rejected (review: FAIL)
- **Summary:** The driver validates the gate verdict as gate-verdict/v2 and halts with exit 2 on any check-gap finding (every finding printed, nothing sent back, no work, no commit, no plan left); uncovered and other kinds are sent back as before.
- **Files:** internal/driver/gatereview.go, internal/driver/plan.go, internal/driver/gatereview_test.go, cmd/vloop/run_gatereview_test.go, cmd/vloop/run_harness_test.go
- **Notes for next iteration:** The previous attempt failed the go check because cmd/vloop tests (run_gatereview_test.go, run_harness_test.go) wrote gate-verdict/v1 stubs; changed to gate-verdict/v2 since the brief's contract makes the driver read v2 only. gateReviewRound returns (problems, check-gap count, err); acceptPlan checks gaps before the planRounds test. cmd/vloop tests take ~200s.

## T2 — Halt the run with exit 2 when the gate verdict holds a check-gap finding, keeping nothing

- **Outcome:** done (review: PASS)
- **Summary:** The driver reads gate-verdict/v2 and halts with exit 2 on any check-gap finding (all findings printed, nothing sent back, no work, no commit, no plan left); uncovered and other kinds are still sent back. Committed end-to-end tests now cover the halt and the sent-back path.
- **Files:** internal/driver/gatereview.go, internal/driver/plan.go, internal/driver/gatereview_test.go, cmd/vloop/run_gatereview_test.go, cmd/vloop/run_harness_test.go
- **Notes for next iteration:** New tests in cmd/vloop/run_gatereview_test.go: TestRunGateReviewCheckGapHalts (round 1 alone, round 1 beside wrong-reason, round 2 after a sent-back finding; each then re-runs and plans afresh) and TestRunGateReviewUncoveredIsSentBack. cmd/vloop stub verdicts were moved from gate-verdict/v1 to v2 because the driver now reads v2 only.

## T3 — Have the gate review rule on every criterion's coverage, and the planner write criteria something can judge

- **Outcome:** done (review: PASS)
- **Summary:** The gate-review skill now rules on every criterion's coverage (gate, named check or review) with the kinds uncovered and check-gap, widens unpassable and task-reach (P-8), and writes gate-verdict/v2. The plan skill cites P-8 instead of restating it and asks the planner for judgeable criteria. The gate-review-coverage eval case is added and listed in the README.
- **Files:** plugin/skills/gate-review/SKILL.md, plugin/skills/plan/SKILL.md, plugin/evals/README.md, plugin/evals/gate-review-coverage/prompt.md, plugin/evals/gate-review-coverage/case.yaml, plugin/evals/gate-review-coverage/brief.md, plugin/evals/gate-review-coverage/scaffold.sh, plugin/evals/gate-review-coverage/graders/finds-check-gap.md, plugin/evals/gate-review-coverage/graders/finds-uncovered.md, plugin/evals/gate-review-coverage/graders/finds-unpassable.md, plugin/evals/gate-review-coverage/graders/covered-task-passes.md, plugin/evals/gate-review-coverage/graders/verdict-is-fail.md, plugin/evals/gate-review-coverage/graders/verdict-validates.md
- **Notes for next iteration:** No word ceiling changed: gate-review is 1527 words (ceiling 1550), plan 3804 (4100). The scaffold exits non-zero if a planted gate passes on the base, fails it without printing its own 'GATE FAIL:' assertion line, or fails a correct implementation (reference impl in a copy); it was run once and exited 0. The eval itself (claude) was not run.

## T4 — Keep vloop's records under .vloop/ out of the review's product constraints

- **Outcome:** done (review: PASS)
- **Summary:** The review skill now has a 'The product' section (everything outside .vloop/; records defects are one line in notes) and a reworded absolute-path question with no path-shaped placeholder; cites P-8. The 08-vloop-records-out-of-scope eval case and its README row are added.
- **Files:** plugin/skills/review/SKILL.md, internal/budget/budget.go, plugin/evals/README.md, plugin/evals/08-vloop-records-out-of-scope/prompt.md, plugin/evals/08-vloop-records-out-of-scope/case.yaml, plugin/evals/08-vloop-records-out-of-scope/task.json, plugin/evals/08-vloop-records-out-of-scope/scaffold.sh, plugin/evals/08-vloop-records-out-of-scope/graders/verdict-is-pass.md, plugin/evals/08-vloop-records-out-of-scope/graders/no-finding.md, plugin/evals/08-vloop-records-out-of-scope/graders/notes-line-about-records.md, plugin/evals/08-vloop-records-out-of-scope/graders/verdict-validates.md
- **Notes for next iteration:** Review skill is 1680 words; ceiling raised 1650 -> 1820 (the 166 words added, plus a little headroom), with date and reason in the budget.go comment. The scaffold composes the records' home path at run time; the eval itself (claude) was not run.

## T5 — Cite P-8 in the operator's playbooks and say what a check-gap halt is in them and the guide

- **Outcome:** done (review: PASS)
- **Summary:** Both operator playbooks cite P-8 in their gate-repair text and name the check-gap halt in the exit-2 row; the concepts guide says what a check-gap halt is and what to do, in its gate review paragraph and exit-code row.
- **Files:** plugin/skills/operate/SKILL.md, .claude/skills/vloop-operator/SKILL.md, docs/guide/concepts.md
- **Notes for next iteration:** Operate skill is about 1700 words, under its 1750 ceiling, so budget.go is unchanged. P-8 is cited by id only, not restated.

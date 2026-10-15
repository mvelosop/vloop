# Journal — B20261015-1001-b16-verification-phase

Append-only narrative of this plan. Rendered state lives in .vloop/state/plan.md.

## Plan — B20261015-1001-b16-verification-phase

- **Brief:** `docs/briefs/B20261015-1001-b16-verification-phase.loop-brief.md`
- **Tasks:** 11

## T1 — Clear .vloop/tmp/ at run start and before every work session and the final pass

- **Outcome:** check_failed (review: skipped)
- **Summary:** driver.ClearTmp empties .vloop/tmp/ but for .running (symlinks removed, not followed; unremovable entries warn and the run continues); called at run start after the lock, before every work session and before the final pass.
- **Files:** internal/driver/cleartmp.go, internal/driver/cleartmp_test.go, internal/driver/iterate.go, internal/cli/run.go
- **Notes for next iteration:** Evidence nothing relies on .vloop/tmp/ surviving a resume: the files there are plan-problems.md (acceptance.go, written by plan session and read by the next round in the same Plan() call), gate-verdict.json (gatereview.go), proposal.json/verdict.json/checks.json (iterate.go, checks.go: written and read within one iteration), nohooks (gitguard.go gitCmd recreates it on every use), plugin/ and fence/ (Runner.Run in session.go calls ExtractPlugin/ExtractFence before every session), gate-refused (iterate.go removes it before a work session), .running (kept). Clearing points: internal/cli/run.go after AcquireLock; Iterator.iterate before the work session (this also replaces the old proposal.json/verdict.json removal and covers right-after-plan-acceptance); before finalPass. Not between plan rounds nor between work and review. No earlier test changed (S1 not touched). Unit tests cover .running kept, symlink not followed, warning for an unremovable entry, a missing folder; the plan-round and work-to-review hand-offs are covered by the T1 gate script, there is no Go harness running the whole driver. go test ./internal/cli fails when run inside a loop session unless VLOOP_PLAN_SHA256 is unset; that is the session env, not the change.

## T1 — Clear .vloop/tmp/ at run start and before every work session and the final pass

- **Outcome:** rejected (review: FAIL)
- **Summary:** The earlier attempt's clearing (internal/driver/cleartmp.go) stays; four earlier tests that seeded state in .vloop/tmp/ before a run now keep it under .git/, which the clearing leaves alone.
- **Files:** cmd/vloop/run_git_test.go, cmd/vloop/b12_e2e_test.go, cmd/vloop/run_checks_test.go
- **Notes for next iteration:** S1: the earlier tests changed are TestRunBaseGateHaltResumesAcceptance and TestWorkedExampleB12/halt_in_acceptance_resumes (tamper marker), TestRunGateReviewHaltResumesAcceptance (reviewed marker) and TestRunChecksNotRunAfterFailedGateAndNeverRerun (check counter). They pinned state surviving in .vloop/tmp/ across run start, which the contract now clears; the markers moved to .git/ (tamper, reviewed, count). Evidence that nothing relies on .vloop/tmp/ surviving a resume is in the previous attempt's notes in the journal (acceptance.go, gatereview.go, iterate.go, checks.go, gitguard.go, session.go).

## T1 — Clear .vloop/tmp/ at run start and before every work session and the final pass

- **Outcome:** done (review: PASS)
- **Summary:** Added run-level tests (cmd/vloop/run_git_test.go) driving a full run: clearing at run start, before each work session and before the final pass, .running kept, plan-round and work-to-review hand-offs kept, a symlink not followed, and the warning for an unremovable entry.
- **Files:** cmd/vloop/run_git_test.go
- **Notes for next iteration:** Evidence nothing relies on .vloop/tmp/ surviving a resume (code paths read): plan-problems.md is written and read by the next round inside one Plan() call (internal/driver/acceptance.go); gate-verdict.json (gatereview.go), proposal.json, verdict.json and checks.json (iterate.go, checks.go) are written and read within one iteration; nohooks is recreated on every use (gitguard.go gitCmd); plugin/ and fence/ are extracted before every session (session.go Runner.Run); gate-refused is removed by iterate.go before a work session; .running is kept. Production code (internal/driver/cleartmp.go, iterate.go, internal/cli/run.go) is unchanged this attempt. S1: earlier tests changed in a previous attempt are TestRunBaseGateHaltResumesAcceptance, TestWorkedExampleB12/halt_in_acceptance_resumes, TestRunGateReviewHaltResumesAcceptance and TestRunChecksNotRunAfterFailedGateAndNeverRerun, whose markers moved from .vloop/tmp/ to .git/ because the contract clears .vloop/tmp/ at run start.

## T2 — Retire brief check's must-change list and external format warnings and init's two guidelines sections

- **Outcome:** done (review: PASS)
- **Summary:** brief check no longer reads ## Pinning tests or ## External formats nor prints the must-change list/external format lines; init's guidelines file is the title and one sentence in en and es.
- **Files:** internal/brief/check.go, internal/brief/pinning.go, internal/brief/extformat.go, internal/brief/pinning_test.go, internal/brief/extformat_test.go, internal/cli/init.go, internal/cli/init_test.go, cmd/vloop/b14_e2e_test.go
- **Notes for next iteration:** K2: earlier tests changed are TestInitGuidelinesTemplate (init_test.go) and TestWorkedExampleB14Init, TestWorkedExampleB14BriefCheck, TestWorkedExampleB14RealDataBriefCheck (b14_e2e_test.go); they pinned the retired sections/warnings and now assert their absence. TestWorkedExampleB14BriefCheck is the committed test that a guidelines file with both sections yields neither line. b14HasLine helper removed (unused). go test ./internal/cli TestTaskGate* fail inside a loop session (plan-changed env guard), unrelated to this change. Files were deleted with rm, not git rm; the driver commits.

## T3 — Add the verify config keys, the verify/v1 report schema and the verify phase in session/v2

- **Outcome:** check_failed (review: skipped)
- **Summary:** Added run.verify, model.verify and effort.verify config keys, schemas/verify.v1.json, and verify as a task-less phase of session/v2; documented the keys in docs/guide/configuration.md.
- **Files:** internal/config/config.go, schemas/verify.v1.json, schemas/session.v2.json, docs/guide/configuration.md, internal/config/config_test.go, internal/cli/config_test.go, internal/cli/schema_test.go, internal/schema/schema_test.go, internal/schema/testdata/verify.valid.json, internal/schema/testdata/verify.invalid.json
- **Notes for next iteration:** V1: earlier tests changed only where they enumerate keys/schema names: TestConfigListDefaults, TestConfigListJSONOrderAndNull (internal/cli/config_test.go), the key list in internal/config/config_test.go, TestSchemaList (count 16 to 17) and TestNames, each only to add the three keys or verify/v1. configuration.md gained three rows because guide tests require every key documented. Nothing consumes the keys yet (T4); ModelEffort resolves model.verify/effort.verify generically once a Phase constant exists.

## T3 — Add the verify config keys, the verify/v1 report schema and the verify phase in session/v2

- **Outcome:** done (review: PASS)
- **Summary:** Fixed the check failure: three earlier tests that enumerate config keys and schema names now include the verify keys and verify/v1.
- **Files:** embed_test.go, cmd/vloop/e2e_test.go, cmd/vloop/b2_e2e_test.go
- **Notes for next iteration:** V1: the check failure was three enumerations outside the named files: TestSchemasAreEmbedded (embed_test.go, 16 to 17), TestWorkedExampleEnglishSession (e2e_test.go, config list gains model.verify, effort.verify, run.verify) and TestWorkedExampleB2Commands (b2_e2e_test.go, schema list gains verify/v1); each changed only to add the new names. Production code unchanged this attempt.

## T4 — Run one verify session after the final pass of every complete run and keep its report

- **Outcome:** rejected (review: FAIL)
- **Summary:** The driver runs one /vloop:verify session after the final pass of a complete run (fence/verify.json = review fence, reverts, warnings, report kept as reports/verify.json, session/v2 record, status names the phase); run.verify=false and the cost ceiling skip it.
- **Files:** internal/driver/iterate.go, internal/driver/session.go, internal/driver/resolve.go, fence/verify.json, internal/cli/fence_test.go, cmd/vloop/run_harness_test.go, cmd/vloop/run_iterate_test.go, cmd/vloop/run_safety_test.go, cmd/vloop/run_verify_test.go
- **Notes for next iteration:** V1: earlier tests changed only for the new session: the stub in run_harness_test.go answers /vloop:verify; defaultScript writes an empty verify/v1 report; TestRun01HappyPath counts 7 sessions; TestRun15TelemetryContract lists 007-verify.json and accepts session/v2 for verify; fencePhases gains verify with the review fence's expectations. TestTaskGate* in internal/cli fail inside a loop session (plan-changed env guard), unrelated. verifySession is in internal/driver/iterate.go; it clears .vloop/tmp/ again after the final pass so check leftovers are gone.

## T4 — Run one verify session after the final pass of every complete run and keep its report

- **Outcome:** done (review: PASS)
- **Summary:** Added the missing committed tests for the verify session: no findings, missing and invalid report, failed session, cost ceiling, revert lines, exit 9 on a moved ref, status phase and session/v2 record, and the resume. Production code is unchanged this attempt.
- **Files:** cmd/vloop/run_verify_test.go
- **Notes for next iteration:** V1: no earlier test changed this attempt. The harness defaultScript already writes an empty verify/v1 report, so a test of a missing report must rm .vloop/tmp/verify.json in the verify phase. The status test runs the built binary from the stub script and writes to $dir/status.txt.

## T5 — Write the /vloop:verify skill within a 1,500-word ceiling

- **Outcome:** check_failed (review: skipped)
- **Summary:** Added plugin/skills/verify/SKILL.md (1,032 words): three passes (example, contract, surface), findings with claim/observed/where, verify/v1 report validated by vloop schema validate, writes only under .vloop/tmp/. The budget gains verify with a ceiling of 1,500.
- **Files:** plugin/skills/verify/SKILL.md, internal/budget/budget.go, internal/cli/readme_test.go, README.md, docs/domain/domain-model.md
- **Notes for next iteration:** V2: earlier test changed only where it enumerates skills: TestReadmeNamesTheSkills (internal/cli/readme_test.go) gains verify; internal/budget/budget_test.go needed no edit because it iterates budget.Skills. README.md Skills section now says six skills and names /vloop:verify. The eval case plugin/evals/verify-planted-mismatches is not part of this task. domain-model.md On disk layout gained .vloop/tmp/verify.json because TestPluginSkills checks skill paths against it. TestTaskGate* in internal/cli fail inside a loop session (plan-changed env guard), unrelated.

## T5 — Write the /vloop:verify skill within a 1,500-word ceiling

- **Outcome:** done (review: PASS)
- **Summary:** plugin/skills/verify/SKILL.md (1,032 words) with three passes, findings and verify/v1 report; budget gains verify at 1,500. Fixed the check failure: the B7 plugin-path test now lists verify among extracted skills.
- **Files:** plugin/skills/verify/SKILL.md, internal/budget/budget.go, internal/cli/readme_test.go, README.md, docs/domain/domain-model.md, cmd/vloop/b7_e2e_test.go
- **Notes for next iteration:** V2: the check failure was TestWorkedExampleB7PluginPath (cmd/vloop/b7_e2e_test.go) enumerating extracted skills; changed only to add verify. Production code unchanged this attempt.

## T6 — Add the verify-planted-mismatches eval case

- **Outcome:** check_failed (review: skipped)
- **Summary:** Added the plugin/evals/verify-planted-mismatches case (case.yaml, prompt.md, scaffold.sh, five graders) and its row in the README's other-cases table.
- **Files:** plugin/evals/verify-planted-mismatches/case.yaml, plugin/evals/verify-planted-mismatches/prompt.md, plugin/evals/verify-planted-mismatches/scaffold.sh, plugin/evals/verify-planted-mismatches/graders/finds-example-exit-code.md, plugin/evals/verify-planted-mismatches/graders/finds-contract-sentence.md, plugin/evals/verify-planted-mismatches/graders/finds-stale-readme.md, plugin/evals/verify-planted-mismatches/graders/sound-line-not-flagged.md, plugin/evals/verify-planted-mismatches/graders/report-validates.md, plugin/evals/README.md
- **Notes for next iteration:** Scaffold commits baseline (README, brief, old bin/greet), records that commit as the plan base, then commits the built result and a one-task state/v2 plan. Planted: no-arg exit 1 vs briefed 2; --shout unimplemented; README says args ignored.

## T6 — Add the verify-planted-mismatches eval case

- **Outcome:** done (review: PASS)
- **Summary:** Fixed the check failure: the verify-planted-mismatches prompt.md grants the same allowed_tools as the other skill cases, so it covers the fence and stays inside the guide's --allow-tools.
- **Files:** plugin/evals/verify-planted-mismatches/prompt.md
- **Notes for next iteration:** TestEvalGrantsCoverFence requires every eval's allowed_tools to include the fence's allow rules (Edit, vloop task list/show/validate/gate) and only Bash rules named in docs/guide/evals.md; Bash(bash:*) and Bash(cd:*) are not in the guide. The grant was copied from gate-review-planted-gates.

## T7 — Show the verify session in vloop metrics' time, cost, models and instructions lines

- **Outcome:** done (review: PASS)
- **Summary:** vloop metrics shows the verify session after gate review on the time, cost and models lines, in the total but not agent, and adds verify to the instructions line when plugin/skills/verify/SKILL.md exists; --json gains the optional verify fields.
- **Files:** internal/metrics/time.go, internal/metrics/metrics.go, internal/metrics/instructions.go, internal/metrics/print.go, schemas/metrics.v2.json, internal/metrics/print_test.go
- **Notes for next iteration:** Verify JSON fields are pointers with omitempty, present only when the brief held a verify session (or the verify skill exists, for instructions.verify), so older briefs print and serialise as before. TestPrintSummaryVerify in internal/metrics/print_test.go covers the three lines, instructions, JSON fields and the no-verify case. go test ./internal/cli task_gate tests failed only because they ran inside a work session (plan-changed guard), not from this change.

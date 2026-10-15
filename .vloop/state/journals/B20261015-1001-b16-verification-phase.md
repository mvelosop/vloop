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

# Journal — B20261014-1138-b15-code-health-escapes

Append-only narrative of this plan. Rendered state lives in .vloop/state/plan.md.

## Plan — B20261014-1138-b15-code-health-escapes

- **Brief:** `docs/briefs/B20261014-1138-b15-code-health-escapes.loop-brief.md`
- **Tasks:** 14

## T1 — Make cmd/vloop's TestMain fail cleanly when the binary does not build

- **Outcome:** check_failed (review: skipped)
- **Summary:** TestMain in cmd/vloop now prints temp-dir and go build failures to stderr and exits 1 instead of panicking.
- **Files:** cmd/vloop/e2e_test.go
- **Notes for next iteration:** none

## T1 — Make cmd/vloop's TestMain fail cleanly when the binary does not build

- **Outcome:** check_failed (review: skipped)
- **Summary:** TestMain in cmd/vloop now prints the temp-dir or go build error to stderr, removes its temp dir and exits 1 instead of panicking.
- **Files:** cmd/vloop/e2e_test.go
- **Notes for next iteration:** The edit was already in the tree from the previous attempt; that attempt's check_failed log shows every package ok, so the failure was not a test failure. Re-verified the gate and gofmt only.

## T1 — Make cmd/vloop's TestMain fail cleanly when the binary does not build

- **Outcome:** check_failed (review: skipped)
- **Summary:** TestMain in cmd/vloop prints the go build output to stderr, removes its temp dir and exits 1 on a build failure (and on a MkdirTemp failure); it never panics. The code was already in place from an earlier attempt; no further change was needed.
- **Files:** cmd/vloop/e2e_test.go
- **Notes for next iteration:** The earlier 'check go failed' was not a test failure: go tests all passed, but `gofmt -l .` lists the untracked, git-ignored operator scratch file .vloop/tmp/b15/fix/e2e_test.go (import order). It is not part of the tree and not mine to delete; if the go check fails again on gofmt, the operator should remove or format that file.

## Run ended — blocked

- **Run:** `B20261014-1138-b15-code-health-escapes` · 3 iteration(s) this run
- **Plan:** 0/14 done, 1 blocked

## T1 — Make cmd/vloop's TestMain fail cleanly when the binary does not build

- **Outcome:** done (review: PASS)
- **Summary:** TestMain in cmd/vloop prints the go build output or the MkdirTemp error to stderr, removes its temp dir and exits 1; it never panics.
- **Files:** cmd/vloop/e2e_test.go
- **Notes for next iteration:** No change needed this attempt: the code was already committed by an earlier attempt. gofmt -l . is clean now.

## T2 — Run every git command through one package

- **Outcome:** done (review: PASS)
- **Summary:** New internal/gitrun package is the only place that calls exec.Command with "git" (Plain, At, Hardened, Capture); all former call sites in cli, defect, driver, intervention, metrics and runs build their commands through it with their trimming and error text unchanged.
- **Files:** internal/gitrun/gitrun.go, internal/gitrun/gitrun_test.go, internal/driver/gitguard.go, internal/cli/close.go, internal/cli/config.go, internal/cli/doctor.go, internal/cli/export.go, internal/defect/blame.go, internal/intervention/show.go, internal/metrics/instructions.go, internal/metrics/lines.go, internal/runs/commits.go, internal/runs/release.go
- **Notes for next iteration:** Hardening lives only in gitrun.Hardened (driver's gitCmd wraps it); gitPath uses gitrun.At and configEntries gitrun.Plain, both unhardened as before. Running go test ./internal/cli inside a session fails TestTaskGate* because the inherited VLOOP_PLAN_SHA256 env var is set; unset it for a local run.

## T3 — Filter the environment for gates, checks and sessions with one function

- **Outcome:** done (review: PASS)
- **Summary:** dropEnv in internal/driver/safety.go is now the only environment filter; runGate, the base gates, runCheck and Session.Run call it with their own name sets.
- **Files:** internal/driver/safety.go, internal/driver/iterate.go, internal/driver/acceptance.go, internal/driver/checks.go, internal/driver/session.go, internal/driver/envfilter_test.go
- **Notes for next iteration:** Kept as is: the gate run honours an injected it.Env (else os.Environ) while the base gates and the checks always start from os.Environ. The session launch also drops PlanHashEnv.

## T4 — One plan-at-the-last-commit function, one task count, one table of budget keys

- **Outcome:** done (review: PASS)
- **Summary:** runs.Owned.PlanSHA is the one place that picks the run, last task or plan commit; the driver's counts is gone in favour of Plan.Counts(); config.BudgetKeys names the nine run.* budget keys once and ResolveBudgets, the run flags and both gate-timeout reads use it.
- **Files:** internal/runs/commits.go, internal/runs/plansha_test.go, internal/config/config.go, internal/config/budgetkeys_test.go, internal/metrics/metrics.go, internal/metrics/defects.go, internal/defect/defect.go, internal/cli/defect.go, internal/cli/run.go, internal/cli/task_gate.go, internal/driver/budgets.go, internal/driver/iterate.go, internal/driver/plan.go, internal/driver/iterate_test.go
- **Notes for next iteration:** internal/driver/iterate_test.go TestNextReady called the removed counts(); its one call now uses p.Counts() with pending as len(Tasks)-done-blocked, same assertions. Running go test ./internal/cli needs VLOOP_PLAN_SHA256 unset.

## T5 — Split the five long functions to at most 100 lines each

- **Outcome:** done (review: PASS)
- **Summary:** iterate, Plan, runInit, runDoctor and runBriefClose are each at most 100 lines, split into named helpers (workSession, runGates, reviewSession, preflight, openRun, commitPlan, configStep, applySteps, doctor* checks, closeRefusals, commitClose, printClosed and others) with no behaviour change.
- **Files:** internal/driver/iterate.go, internal/driver/plan.go, internal/cli/init.go, internal/cli/doctor.go, internal/cli/close.go
- **Notes for next iteration:** iterate is exactly 100 lines and Plan 98; runBriefClose is 91, runInit 87, runDoctor 46. Later tasks that add to iterate or Plan need to extract a helper rather than grow them. Running go test ./internal/cli needs VLOOP_PLAN_SHA256 unset.

## T6 — Compare run.git-config-ignore patterns as git reports keys

- **Outcome:** done (review: PASS)
- **Summary:** gitcfg.Normalize lower-cases a pattern's text before the first and after the last dot, keeping a subsection; Match and Protected apply it, so Core.* is refused like core.* and Branch.Main.Foo-Bar matches only branch.Main.foo-bar.
- **Files:** internal/gitcfg/gitcfg.go, internal/gitcfg/gitcfg_test.go, internal/config/config_test.go
- **Notes for next iteration:** Existing TestMatch case {Branch.*, branch.x, false} asserted the defect itself (literal comparison); flipped to true and added {Branch.X.*, branch.x.y, false}, as the acceptance requires. All callers (config validation, guard, startup refusal) go through Match/Protected, so normalization lives only in gitcfg.

## T7 — Give new defects and interventions second-stamped IDs that never collide

- **Outcome:** check_failed (review: skipped)
- **Summary:** New defects and interventions get D/I<YYYYMMDD-HHMMSS>-<slug> IDs from one generator, defect.StampedID, that moves the stamp on a second until no record of the kind uses it; close --dry-run predicts with it. Show links and the guidelines provenance check accept both ID forms.
- **Files:** internal/defect/defect.go, internal/defect/defect_test.go, internal/intervention/intervention.go, internal/intervention/intervention_test.go, internal/intervention/show.go, internal/intervention/links_test.go, internal/cli/close.go, internal/budget/budget.go, internal/budget/guideline_id_test.go
- **Notes for next iteration:** internal/cli TestTaskGate* tests fail inside a live session (plan changed during session guard); unrelated to this task. defect set / intervention set needed no change: they look up by file name.

## T7 — Give new defects and interventions second-stamped IDs that never collide

- **Outcome:** check_failed (review: skipped)
- **Summary:** Second-stamped D/I IDs from one generator (defect.StampedID) were already in place; this attempt fixes the check failure: two cmd/vloop worked-example tests asserted the old four-digit minute stamp.
- **Files:** cmd/vloop/b3_e2e_test.go, cmd/vloop/b4_e2e_test.go
- **Notes for next iteration:** The previous attempt's check_failed was only those two regexes (D\d{8}-\d{4} -> \d{6}); other e2e regexes match brief IDs and stay four-digit.

## T7 — Give new defects and interventions second-stamped IDs that never collide

- **Outcome:** done (review: PASS)
- **Summary:** Second-stamped D/I IDs come from one generator (defect.StampedID); close --dry-run predicts with it; both ID forms are accepted. This attempt fixes the remaining check failure: the B4 worked-example test was flaky across a second boundary.
- **Files:** cmd/vloop/b4_e2e_test.go
- **Notes for next iteration:** The dry run's predicted stamp and the real close can fall in different seconds (the close is slower), so TestWorkedExampleB4Close now takes the recorded stamp from the close output and only requires it to be >= the prediction. Earlier fixes: two cmd/vloop regexes moved from four-digit to six-digit stamps.

## T8 — Show test:code and $/kloc per task, and name the cost per 1,000 code lines $/kloc everywhere

- **Outcome:** rejected (review: FAIL)
- **Summary:** The per-task table gains test:code and $/kloc columns, by_task JSON gains test_code_ratio and per_1000_code_lines_usd (null where code+ is 0, optional in metrics/v2), the summary cost line ends $<c>/kloc, and the all-briefs headers are test:code and $/kloc.
- **Files:** internal/metrics/metrics.go, internal/metrics/print.go, internal/metrics/print_test.go, schemas/metrics.v2.json, internal/cli/metrics_test.go, cmd/vloop/b3_e2e_test.go
- **Notes for next iteration:** Docs still say 'per 1,000 code lines' (docs/guide/metrics.md:235 etc.); that is T13's. lines_test.go and run_iterate_test.go needed no change.

## T8 — Show test:code and $/kloc per task, and name the cost per 1,000 code lines $/kloc everywhere

- **Outcome:** done (review: PASS)
- **Summary:** Per-task test:code and $/kloc columns, by_task JSON fields (null where code+ is 0), $/kloc summary ending and all-briefs headers were already in place; this attempt adds the missing committed test of byTask's nil ratios and the JSON null/value fields.
- **Files:** internal/metrics/print_test.go
- **Notes for next iteration:** Review rejected attempt 1 only for lacking a test of the JSON fields and the null case; TestByTaskRatiosAndJSONNull in internal/metrics/print_test.go calls byTask directly and marshals ReportTask. Docs still say 'per 1,000 code lines' (T13's).

## T9 — Refuse vloop defect add for a brief that is not merged

- **Outcome:** done (review: PASS)
- **Summary:** vloop defect add now refuses, before writing anything, a brief (from --brief or --blame) that is not merged, with the intervention-add hint on stderr and exit 1; with no default branch it returns M-5's lookup error.
- **Files:** internal/cli/defect.go, internal/cli/defect_test.go, internal/cli/matrix_test.go, internal/cli/export_test.go, internal/cli/intervention_test.go, cmd/vloop/b20261005_e2e_test.go, cmd/vloop/b14_e2e_test.go
- **Notes for next iteration:** The check is requireMerged in internal/cli/defect.go, using runs.Release (same lookup as metrics). The blame attribution line is now printed only after the merged check. Test fixtures merge the brief via mergeBrief (internal/cli/matrix_test.go) and b15MergeBrief (cmd/vloop/b20261005_e2e_test.go), which commit to main with a scratch index. internal/cli TestTaskGate* fail inside a live session (plan changed during session guard), unrelated.

## T10 — Count a recorded defect as escaped when it was created after its brief merged

- **Outcome:** blocked (review: skipped)
- **Summary:** Implemented the time rule: DefectCounts.Reclassify moves a recorded defect created after the release commit's committer time to escaped whatever its found-by, applied in metrics.Build (so summary, JSON, all-briefs table, close and snapshot follow); the matrix is untouched. Added TestReclassifyByReleaseTime. Every gate clause passes except the real-data check, which a correct implementation cannot pass.
- **Files:** internal/metrics/defects.go, internal/metrics/metrics.go, internal/metrics/defects_test.go
- **Notes for next iteration:** The gate's real-data jq clause asserts in-loop unchanged, but acceptance 1 and the gate's own fixtures (D...-0004-after-review moves review to escaped) require a gate- or review-found record created after the merge to become escaped. Two real briefs have such records (e.g. D20261004-2304-testworkedexampleb6realdata-as-repaired, found-by gate, created 2026-10-04T22:04Z, after the gate-model merge at 13:18Z; D20260930-0809-t1-gate-used-a-gnu-only-grep..., found-by gate, on the state-tasks-status brief). The real-data clause should instead assert operator + escaped + in-loop total unchanged (in-loop may only fall, escaped only rise).

## Run ended — blocked

- **Run:** `B20261014-1138-b15-code-health-escapes` · 13 iteration(s) this run
- **Plan:** 9/14 done, 1 blocked

## T10 — Count a recorded defect as escaped when it was created after its brief merged

- **Outcome:** done (review: PASS)
- **Summary:** A recorded defect created after its brief's release commit (committer time, via the git runner) is counted escaped whatever its found-by; same-second and unmerged briefs count by found-by. DefectCounts.Reclassify is applied in metrics.Build, so summary, JSON, all-briefs table, close and snapshot follow; the matrix and records are untouched.
- **Files:** internal/metrics/defects.go, internal/metrics/metrics.go, internal/metrics/defects_test.go
- **Notes for next iteration:** Real data, in-loop/operator/escaped before -> after (total escaped 3 -> 33): skeleton-config-briefs 0/3/2 -> 0/0/5; state-tasks-status 1/3/1 -> 0/0/5; metrics-defects 0/2/0 -> 0/1/1; close-export-workspace 0/1/0 -> 0/0/1; init-upgrade-doctor 2/1/0 -> 2/0/1; run-driver 0/13/0 -> 0/1/12; skills 0/5/0 -> 0/4/1; v1-security 6/2/0 -> 6/0/2; gate-model 7/11/0 -> 6/8/4; interventions-options 10/6/0 -> 10/5/1; others unchanged. Gate-found records created after a merge also move to escaped (gate was amended for that). Reclassify lives in internal/metrics/defects.go; zero release time means unmerged.

## T11 — Record operator findings as verification-finding interventions with an origin

- **Outcome:** rejected (review: FAIL)
- **Summary:** intervention add takes --origin (required for verification-finding, refused otherwise) and the record, show/list JSON, set and schema carry origin; brief close --finding "<origin>: <summary>" records verification-finding interventions, counted as operator-caught alongside non-escaped operator defects, listed after the defects in the run record and in --dry-run.
- **Files:** internal/intervention/intervention.go, internal/intervention/intervention_test.go, internal/cli/intervention.go, internal/cli/close.go, internal/cli/close_test.go, internal/cli/close_flags_test.go, internal/closing/closing.go, internal/metrics/defects.go, internal/metrics/defects_test.go, internal/metrics/metrics.go, schemas/intervention.v2.json, docs/guide/commands.md, cmd/vloop/b4_e2e_test.go, cmd/vloop/b14_e2e_test.go
- **Notes for next iteration:** Counting lives in metrics.DefectCounts.AddFindings (IsFinding: kind verification-finding with an origin), applied in metrics.Build after Reclassify. close's defect report is unchanged (T12 retires it), so a close with findings prints 'defects of <run>: none' when the brief has no defects. internal/cli TestTaskGate* fail inside a live session (plan changed during session guard), unrelated.

## T11 — Record operator findings as verification-finding interventions with an origin

- **Outcome:** done (review: PASS)
- **Summary:** Implementation was already in place; this attempt adds the two missing committed tests the review asked for: intervention add's two --origin refusals (and the origin line after by:), and the operator count taking operator defects and verification-finding interventions together without overlap.
- **Files:** internal/cli/intervention_test.go, internal/metrics/defects_test.go
- **Notes for next iteration:** Review rejected attempt 1 only for those two untested behaviours; TestInterventionAddOriginRefusals (internal/cli) and TestOperatorCountTwoSourcesWithoutOverlap (internal/metrics) cover them. internal/cli TestTaskGate* fail inside a live session (plan changed during session guard), unrelated.

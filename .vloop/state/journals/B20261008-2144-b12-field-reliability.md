# Journal — B20261008-2144-b12-field-reliability

Append-only narrative of this plan. Rendered state lives in .vloop/state/plan.md.

## Plan — B20261008-2144-b12-field-reliability

- **Brief:** `docs/briefs/B20261008-2144-b12-field-reliability.loop-brief.md`
- **Tasks:** 11

## T1 — Make the .git/config guard compare parsed keys, ignore editor bookkeeping, and name what changed

- **Outcome:** rejected (review: FAIL)
- **Summary:** The .git/config guard now compares the file's own entries (keys and values, include directives as entries) instead of its bytes, leaves out keys matching the built-in branch.*.vscode-merge-base and the new run.git-config-ignore list, names up to three sorted changed keys in the exit-9 halt, and refuses patterns that would hide core, filter, include or includeif keys.
- **Files:** internal/gitcfg/gitcfg.go, internal/gitcfg/gitcfg_test.go, internal/driver/gitguard.go, internal/driver/gitguard_test.go, internal/config/config.go, internal/config/config_test.go, internal/cli/run.go, internal/cli/config_test.go, cmd/vloop/run_git_test.go, cmd/vloop/e2e_test.go, docs/guide/configuration.md
- **Notes for next iteration:** Pure rules live in internal/gitcfg (Match, Protected). The guard reads entries with git config --file <path> --no-includes --list -z, resolving patterns once at snapshot; halt text is built in gitGuard.changed(), so call sites are unchanged. A config git cannot parse is held as one pseudo-key (unreadable) by digest. config.Key got a GitKeys flag; a protected pattern from set is a config.UsageError (exit 2), from env or file a bare ProtectedPatternError (exit 1), and run checks it right after ResolveBudgets. Besides the tests the brief names, TestDefaultsEveryRow (internal/config) and TestConfigListDefaults/TestConfigListJSONOrderAndNull (internal/cli) also pin the key list and had to gain the new key. The brief's hooks/config subtests of TestWorkedExampleB8 do not mention .git/config, so none changed there. docs/guide/commands.md is unchanged: no help text changed. When running go test inside a vloop session, unset VLOOP_PLAN_SHA256 or TestTaskGate* fail.

## T1 — Make the .git/config guard compare parsed keys, ignore editor bookkeeping, and name what changed

- **Outcome:** done (review: PASS)
- **Summary:** Fixed the review's findings in the .git/config guard: a valueless key is now distinct from an empty value, and values of a repeated key are compared in file order.
- **Files:** internal/driver/gitguard.go, internal/driver/gitguard_test.go
- **Notes for next iteration:** Earlier attempt's work (internal/gitcfg, config key, run.go check, docs) is already in the tree and unchanged. configEntries keeps strings.Cut's found flag and records a NUL-prefixed sentinel for a valueless key (git config -z lists it without newline); values are no longer sorted. New tests: TestGitGuardValuelessKeyIsNotAnEmptyValue, TestGitGuardRepeatedKeyOrderMatters. When running go test inside a vloop session, unset VLOOP_PLAN_SHA256 or TestTaskGate* fail.

## T2 — Stamp the plan on load and resume acceptance after a halt instead of planning again

- **Outcome:** rejected (review: FAIL)
- **Summary:** The plan is stamped (run_id, brief, branch, status planning) as soon as it loads and becomes running only when the gate review passes it; vloop run now resumes acceptance, with no plan session, whenever a plan is planning or has no passed gate review.
- **Files:** internal/driver/plan.go, cmd/vloop/run_git_test.go
- **Notes for next iteration:** Planner.Plan takes the acceptance-resume path when needsAcceptance(plan): status planning, or status not blocked with gate_review not PASS and iteration 0. It opens a new run folder, logs 'resuming acceptance of <run id>: round <n>', skips plan session and base checks, and calls acceptPlan from round n = plan sessions recorded across the run id's folders (1..2). acceptPlan stamps via stampPlan(status) before the base gates; the 'planned:' line now prints after the base gates pass, as before. Blocked plans keep the iterator's resumeBaseGates path. New tests: TestRunBaseGateHaltResumesAcceptance and TestRunGateReviewHaltResumesAcceptance in cmd/vloop/run_git_test.go. Domain docs (plan.md lifecycle/acceptance, run.md) are not updated here; that is T10. When running go test inside a vloop session, unset VLOOP_PLAN_SHA256 or TestTaskGate* fail.

## T2 — Stamp the plan on load and resume acceptance after a halt instead of planning again

- **Outcome:** done (review: PASS)
- **Summary:** The plan is stamped (planning) as soon as it loads and becomes running only when the gate review passes; vloop run resumes acceptance with no plan session whenever a plan has not passed its gate review. A runs test now pins that the resumed acceptance folder is owned and its sessions counted.
- **Files:** internal/driver/plan.go, cmd/vloop/run_git_test.go, internal/runs/runs_test.go
- **Notes for next iteration:** Earlier attempt's plan.go work is unchanged in the tree. The review's finding that the resumed folder is not owned came from reading logOwns, which only serves shell-loop loop.log folders; vloop folders are owned by sitting under .vloop/state/runs/<run id>/, so no code change was needed, only the new test TestVloopResumedAcceptanceFolderIsOwned. Domain docs for the lifecycle and resume are T10's.

## T3 — Record ended and wall_ms, the driver's own measurement, on every session record

- **Outcome:** blocked (review: skipped)
- **Summary:** Implemented ended and wall_ms on every session record (driver-measured around the session process, duration_ms untouched) and declared both as optional in session.v1.json and session.v2.json; the gate fails only on a clause no correct implementation can pass.
- **Files:** internal/driver/session.go, internal/driver/session_test.go, schemas/session.v1.json, schemas/session.v2.json
- **Notes for next iteration:** The gate's for-loop over v1 and v2 validates $P, the plan record (schema session/v1), against session/v2; the v2 schema's const 'session/v2' rejects it regardless of ended/wall_ms. A correct fix needs a session/v2 record (the gate review one, $GR) for the v2 iteration. ended and wall_ms are taken from r.now() before and after state.RunGroup, rounded to strip the monotonic reading so sleep time counts; the test clock is fixed, so TestSessionFieldMapping expects wall_ms 0 and ended == started. TestRun15TelemetryContract lists only cost/turns/duration/is_error/denials keys, so it is unchanged.

## T6 — Print a line per session with its time and cost against the ceiling, and warn at 50% and 80%

- **Outcome:** done (review: PASS)
- **Summary:** Every plan, gate review, work and review session now prints '   <phase>[ <task>] · <m>m<ss>s · $<cost> · run $<spent> / $<ceiling>' on stdout and in run.log, and the 50% and 80% cost warnings go once each to stderr and run.log.
- **Files:** internal/driver/cost.go, internal/driver/cost_test.go, internal/driver/session.go, internal/driver/iterate.go, internal/driver/plan.go, internal/driver/gatereview.go, internal/cli/run.go
- **Notes for next iteration:** Session time is Result.Wall (the driver's ended-started, same measurement as wall_ms). Spend is the sum of cost_usd in the current run folder (spendIn), as the ceiling check uses; warnings are stateless: a threshold is warned in the session whose spend crossed it (before < threshold <= after), so a resume never repeats one. Planner got a CostCeiling field set from the resolved budget in internal/cli/run.go. No existing run-output test needed changing.

## T7 — Record the session in flight in the run lock and show the live run in vloop status

- **Outcome:** done (review: PASS)
- **Summary:** The run lock now records phase, task, session_started and spent as each session starts and ends, and vloop status prints a running line (and --json a running object) for a live run, including while it plans with no plan yet.
- **Files:** internal/driver/safety.go, internal/driver/session.go, internal/driver/session_test.go, internal/cli/status.go, internal/cli/status_test.go
- **Notes for next iteration:** Runner.Run updates the lock through updateLock (rename of a temp file) only when this process holds it, so tests without a lock create none; after a session phase, task and session_started are cleared and spent is re-summed from the run folder, so status between sessions prints 'running: between sessions'. The ceiling is not in the lock: status reads it from config (run.cost-ceiling), so a --cost-ceiling flag override is not reflected. Phase value for a gate review is gate-review in the lock and JSON, printed 'gate review'. internal/cli task-gate tests fail if VLOOP_PLAN_SHA256 is set in the environment (as inside a work session); run them with it unset. docs/guide/commands.md unchanged: no help text changed.

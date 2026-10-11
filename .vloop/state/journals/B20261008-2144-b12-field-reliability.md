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

## T8 — Name the commit from the build's VCS revision and the version from the module, when no stamp did

- **Outcome:** rejected (review: FAIL)
- **Summary:** cli.Resolve completes the Build from Go's build info: stamped commit, else the VCS revision (-dirty when modified), else unknown, plus the module version. version prints/emits module for an unknown commit; doctor's self-hosting passes a release module version.
- **Files:** internal/cli/buildinfo.go, internal/cli/root.go, internal/cli/version.go, internal/cli/doctor.go, cmd/vloop/main.go, internal/cli/root_test.go, internal/cli/doctor_test.go, cmd/vloop/e2e_test.go
- **Notes for next iteration:** Seam is the readBuildInfo var in internal/cli/buildinfo.go (TestResolveBuild injects it). The e2e test's go build now passes -buildvcs=false so its unstamped binary stays commit unknown without a module (otherwise it picks up the checkout's revision, which is the new behaviour). TestVersionJSON/TestVersionText/TestDoctorSelfHostingUnstamped needed no change. docs/guide/concepts.md still says go install carries no commit; left for the docs task. internal/cli task-gate tests fail if VLOOP_PLAN_SHA256 is set; run them with it unset.

## T8 — Name the commit from the build's VCS revision and the version from the module, when no stamp did

- **Outcome:** done (review: PASS)
- **Summary:** Added TestVersionModule covering vloop version text and --json with a module and unknown commit, with a module and a known commit, and with no module; the build-info implementation from the earlier attempt is unchanged.
- **Files:** internal/cli/root_test.go
- **Notes for next iteration:** The earlier attempt's implementation (internal/cli/buildinfo.go, version.go, doctor.go) was rejected only for lacking a version-output test; the new test is added alongside, existing tests untouched. internal/cli tests fail if VLOOP_PLAN_SHA256 is set; run them with it unset.

## Run ended — blocked

- **Run:** `B20261008-2144-b12-field-reliability` · 9 iteration(s) this run
- **Plan:** 5/11 done, 1 blocked

## T3 — Record ended and wall_ms, the driver's own measurement, on every session record

- **Outcome:** done (review: PASS)
- **Summary:** Session records carry ended and wall_ms, measured by the driver around the session process; duration_ms stays as the CLI reported it. Both session schemas declare them as optional with no version change.
- **Files:** internal/driver/session.go, internal/driver/session_test.go, schemas/session.v1.json, schemas/session.v2.json
- **Notes for next iteration:** The implementation was already in the tree from the earlier blocked attempt; this session verified it against the corrected gate and changed nothing. wall_ms is ended minus started from the driver clock (r.now) around the process.

## T4 — Take session time from wall_ms in the metrics, and start a brief's wall at its first plan session

- **Outcome:** blocked (review: skipped)
- **Summary:** Metrics take session time from wall_ms (duration_ms otherwise) in every phase figure and per-task time, and the wall starts at the first plan session's started; the gate's review_ms clause cannot pass for a correct implementation.
- **Files:** internal/runs/runs.go, internal/metrics/time.go, internal/metrics/metrics.go, internal/metrics/time_test.go
- **Notes for next iteration:** Session.WallMS (*int64) and Session.TimeMS() in internal/runs; Aggregate and the per-task rows use TimeMS; wall starts at the earliest plan session Started across folders, else the plan commit. No existing test needed changing, so no changed numbers to list; added TestTimeFromWallMSWhereRecorded and TestTimeWallStartsAtFirstPlanSession. The gate's review clause sums records with glob *-review.json, which also matches NNN-gate-review.json (phase gate-review), so its want folds gate review time into review_ms while the metric correctly keeps time.gate_review_ms apart (the plan, work and wall clauses pass). The gate needs glob [0-9]*-review.json or a filter on .phase.

## T5 — Keep a denied Bash call's command, masked, redacted and cut to 200 characters, and mask the dashed user form

- **Outcome:** done (review: PASS)
- **Summary:** A denied Bash call's permission_denials entry now keeps command, masked and redacted then cut to 200 characters with … appended; other tools are unchanged. Mask also replaces the dashed -Users-<user>- and -home-<user>- forms with USER.
- **Files:** internal/driver/session.go, internal/driver/session_test.go
- **Notes for next iteration:** Runner.cutCommand masks first (Mask), then cuts by runes, so a secret is never split by the cut. The dashed form is handled in Runner.Mask, so it applies to records, run.log and stderr captures alike. TestSessionDenialsKeepToolAndPath now expects the Bash command; added TestCutCommandCountsCharactersAndMasksFirst and TestMaskDashedUserForm. No past record under .vloop/ was rewritten.

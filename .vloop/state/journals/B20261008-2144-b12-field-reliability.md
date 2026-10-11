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

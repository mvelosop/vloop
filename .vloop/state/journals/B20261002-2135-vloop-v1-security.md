# Journal — B20261002-2135-vloop-v1-security

Append-only narrative of this plan. Rendered state lives in .vloop/state/plan.md.

## Plan — B20261002-2135-vloop-v1-security

- **Brief:** `docs/briefs/B20261002-2135-vloop-v1-security.loop-brief.md`
- **Tasks:** 18

## T1 — Give each session phase its own fence and deny the bypasses the review found (F10, F6 fence part)

- **Outcome:** done (review: PASS)
- **Summary:** The fence is now three embedded files (fence/plan.json, work.json, review.json), each extracted and passed to its own session, with the F10 bypass forms denied in all three. session.md says the fence is advisory.
- **Files:** fence/plan.json, fence/work.json, fence/review.json, embed.go, embed_test.go, internal/driver/extract.go, internal/driver/session.go, internal/driver/session_test.go, cmd/vloop/b6_e2e_test.go, internal/cli/fence_test.go, internal/cli/skills_test.go, docs/guide/evals.md, docs/domain/execution/session.md
- **Notes for next iteration:** fence/settings.json is gone; vloop.Fence(phase) replaces the vloop.Fence byte slice and ExtractFence takes the phase, writing .vloop/tmp/fence/<version>/<phase>.json (session_test.go and b6_e2e_test.go paths updated). The review fence differs from work's by allowing Edit/Write only under .vloop/tmp/**; the work fence's allow list is unchanged, so eval grants needed no change. Deny rules use glob forms such as Bash(find*-delete*); the skills_test matcher handles both :* prefixes and * globs. Dropped from my first draft: denying git add/mv/rm/restore/apply, since a work session may legitimately need them.

## T2 — Make the driver's git ignore hooks and fsmonitor, and halt when .git/config, the hooks or a gate's refs change (F4, F5)

- **Outcome:** done (review: PASS)
- **Summary:** Every driver git command now disables hooks and fsmonitor and reads paths unquoted. The run halts with exit 9, committing nothing, when a session or gate changes .git/config or the hooks, or when a gate moves refs. The preflight pre-commit line is reworded.
- **Files:** internal/driver/gitguard.go, internal/driver/plan.go, internal/driver/iterate.go, internal/driver/gateshape.go, internal/cli/run.go, cmd/vloop/run_git_test.go
- **Notes for next iteration:** gitCmd (internal/driver/gitguard.go) is the one constructor for driver git; it recreates .vloop/tmp/nohooks empty on each call because a session can write there. gitGuard hashes .git/config and the hooks dir from plain `git rev-parse --git-path`, without the overrides, so it sees the real hooks. Plan and work and review sessions, and each gate run (own and regression), are checked via it.gitChanged. Gate refs are compared inside runGate. Non-driver packages (metrics, defect, runs, cli doctor/close) still use plain git; they were out of this task's files.

## T3 — Keep the driver's inputs in memory, restore what a session changes among them, and run gates only from the plan the driver holds (F6 driver part, F7)

- **Outcome:** done (review: PASS)
- **Summary:** The driver now sums spend in memory, resolves work and review model and effort once, snapshots and restores its inputs around each work and review session (a review that changes one is failed), and hands sessions VLOOP_PLAN_SHA256; vloop task gate refuses a plan that does not match it.
- **Files:** internal/driver/iterate.go, internal/driver/safety.go, internal/driver/session.go, internal/driver/resolve.go, internal/cli/task_gate.go, internal/cli/task_gate_test.go, cmd/vloop/run_inputs_test.go
- **Notes for next iteration:** Spend is read from disk once at the start of the iterating phase (to pick up the plan session) and then summed from each session's Result; the old Iterator.spend is kept for that one read. Inputs are snapshotted as file bytes (config.toml, defects, interventions, run sessions/ and reports/) and compared after the session; the session's own record is excluded from the added-file check. A refused task gate leaves .vloop/tmp/gate-refused, which the driver logs and removes after the session. ResolveRun/Resolved.For in resolve.go replace ModelEffort for work and review; plan still uses ModelEffort.

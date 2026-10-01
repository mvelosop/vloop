# Journal — B20261001-0723-vloop-run-driver

Append-only narrative of this plan. Rendered state lives in .loop/state/plan.md.

## Plan — B20261001-0723-vloop-run-driver

- **Brief:** `docs/briefs/B20261001-0723-vloop-run-driver.loop-brief.md`
- **Tasks:** 11

The plan is written to `.loop/state/state.json` and passes `.loop/amend.sh check`.

- **Run id:** `B20261001-0723-vloop-run-driver`
- **Tasks:** 11 (the brief asked for 11 to 14)
- **First ready task:** T1, the test harness. T2 (budget keys) and T3 (sessions) have no dependencies either.

| Task | What it builds | Ported scenarios |
|---|---|---|
| T1 | Test harness with a stub `claude` | — |
| T2 | The six `run.*` config keys | — |
| T3 | Sessions: fence, invocation, `session/v1` record | — |
| T4 | `vloop run`, work branch, planning, plan checks | 08 25 27 29 35 36 38 43 |
| T5 | One iteration, plus the HEAD check before each commit | 01 02 09 16 17 21 30 37 40 |
| T6 | Gates across iterations, flaky gates, disputes | 03 33 39 41 |
| T7 | Halts and budgets | 04 05 06 07 11 19 42 44 |
| T8 | Safety and records | 10 13 14 15 18 20 22 23 24 |
| T9 | Snapshot, summary, flaky-gate defect | — |
| T10 | Docs, including the README Windows-shell fix (F1) | — |
| T11 | Close: worked example, real-data check, `go test ./...` | — |

**How far the gates are proven.** I ran all 11 gates against the current tree and each fails for the right reason. For T4–T9 and T11 that failure is `unknown command "run"`, so the checks that come after a run can't execute yet. I tested those checks against a throwaway mock driver for:
- the worked example, snapshot and summary;
- containment, the missing session record, and the gate environment variables.

The planted failures in T4 (other than plan-only), T6 and T7 were reviewed by eye only. If one of those gates is wrong, that's where it will be.

That testing caught one real defect: B3's metrics print `gates 0.0 · wall`, with no `min` after the gate figure, though the brief's example shows one. The gates accept both forms; I didn't change the metrics format.

**Names I pinned** (the brief is silent on these):
- Package `internal/driver`, and the embedded fence at `fence/settings.json`.
- Test names: `TestRunHarness`, `TestRun<NN><Name>` for the 38 ported scenarios, `TestRunWorkBranch`, `TestRunHeadCheck`, `TestRunFlakyGate`, `TestRunGateDispute`, `TestRunBudgetPrecedence`, `TestRunSessionRecordMissing`, `TestRunSnapshot`, `TestRunSummary`, `TestSession*`, a test named `*Flaky*` in `internal/metrics`, and `TestWorkedExampleB6*`. The test files are `cmd/vloop/run_*_test.go` and `b6_e2e_test.go`.
- Environment variables replace `-` with `_` as well as `.`, giving `VLOOP_RUN_MAX_ITERATIONS` and so on.
- The exit codes go in `docs/guide/concepts.md` under `## Exit codes`, and the README links to `#exit-codes`.

**Where I interpreted rather than read** — please correct any of these before the run:
1. **Scenarios 21 and 30 move from planning (T4) to T5**, because both run full iterations. T5 also enforces `--max-iterations`, which scenarios 39 and 40 need; its "resumable" test (05) stays in T7.
2. **A missing brief path exits 2**, as scenario 29 and B1's usage convention say, although R-3 lists usage errors under exit 1.
3. **`vloop run` refuses an untrusted workspace** (scenario 18), even though `vloop doctor` only warns about it. It also warns about an active pre-commit hook (scenario 23).
4. **The `run.*` keys are documented in T2, and the `vloop run` README row in T4.** The existing README and guide tests fail otherwise.
5. **Exit codes:** the README only lists the general codes 0–2. I read "move" as: the guide holds those plus `run`'s 0–9, and the README section becomes a pointer.
6. **Validation ranges:** `max-iterations` may be 0 (scenario 13 uses 0), while `max-attempts` and `stall-limit` must be at least 1.
7. **The final snapshot goes into the closing commit**, so the working tree is clean when a run ends.
8. **`.vloop/tmp/` is never committed**, even when `.gitignore` doesn't list it. This repository's `.gitignore` doesn't, which the real-data clone exercises.
9. **The brief's "no 004-review.json" doesn't match the session numbering** (001-plan, 002-work, 003-review). The gates check that no review record exists for iteration 1, not a filename.
10. **The fence's allow list is exactly** `.loop/settings.json` plus the four read-only commands. Its deny list must include `.loop/settings.json`'s denials plus every vloop write command. The `--settings` and `--plugin-dir` paths are repo-relative, as the worked example shows.
11. **Base SHA:** the brief says the plan pins it, but the shell loop's plan format has no field for it, so it isn't recorded. No gate compares against it, since a fixed baseline would trip the decaying-baseline rule.

The gates are long single-line shell commands (2–16 KB). `amend.sh check` raised one advisory: T2 and T4 edit `README.md`, which T10's gate reads. That's expected, because both run before T10 and the README test forces those edits.


## T1 — Build the run test harness: scratch repository, fake home, a stub claude scripted by prompt, and a smoke test

- **Outcome:** done (review: PASS)
- **Summary:** cmd/vloop/run_harness_test.go now provides newRunRepo (scratch repo on main, init, ready brief, fake trusted home, stub claude) and TestRunHarness.
- **Files:** cmd/vloop/run_harness_test.go
- **Notes for next iteration:** Stub is a POSIX sh script sourcing a per-test script.sh (r.script(body)) in the repo cwd; it sets PHASE, TASK, ARG, MODEL, ATTEMPT (per PHASE+ARG counter kept in the stub dir, outside the repo) and STUB_COST/EXIT/DURATION/TURNS/ERROR/SILENT, so tests branch with case "$PHASE:$TASK:$ATTEMPT". argv log is one space-joined line per invocation (r.argv()), including doctor's --version and plugin probes. Stub reports plugin version 0.0.0-dev to match binPath. Helpers: r.vloop(args), r.claude(args), r.git, r.write/read, r.env(extra...). Skips on windows via t.Skip. Gate's go test was run by me with the full verify only in parts (the verify's individual steps), not as one script.

## T2 — Add the six run.* budget keys to config, with VLOOP_RUN_* variables, validation and the config list update

- **Outcome:** done (review: PASS)
- **Summary:** Added the six run.* budget keys to config.Keys (after metrics.excluded) with VLOOP_RUN_* variables, integer/number validation, [run] TOML number storage and reading, plus tests, README and configuration guide rows.
- **Files:** internal/config/config.go, internal/config/config_test.go, internal/cli/config_test.go, internal/cli/readme_test.go, internal/cli/commands_doc_test.go, README.md, docs/guide/configuration.md
- **Notes for next iteration:** Key gained Int/Num/Min/MinOpen fields; config.EnvVar now turns '-' into '_' too. Values stay strings in Value/JSON (e.g. "30"); `config set` writes TOML integers/floats under [run], and the file reader accepts TOML numbers (a string in the file is a SourceError). Also edited internal/cli/commands_doc_test.go (not in the task's file list): its own copy of the env-name rule lacked '-' and failed on the new keys; it now calls config.EnvVar. README was at 199/200 lines, so I condensed the config prose paragraph to fit the six rows and avoided writing `vloop run` in it (README test rejects mentions of non-commands until T4). docs/guide/commands.md needed no change (no new command or flag).

## T3 — Start sessions: the embedded fence, plugin and fence extraction, the claude invocation, the session/v1 record with masking, and the missing-record warning

- **Outcome:** done (review: PASS)
- **Summary:** New internal/driver package: embedded fence (fence/settings.json via embed.go), shared plugin and fence extraction, the P-6 resolver, and Runner.Run, which starts claude under the fence, writes session/v1 records with masking, and logs SESSION RECORD MISSING when stdout is empty.
- **Files:** fence/settings.json, embed.go, embed_test.go, internal/driver/extract.go, internal/driver/resolve.go, internal/driver/session.go, internal/driver/session_test.go, internal/cli/plugin.go, internal/cli/task.go
- **Notes for next iteration:** ExtractPlugin/EnsureRealDir/withVersion moved from internal/cli/plugin.go to internal/driver/extract.go (cli imports driver, so driver cannot import cli); `vloop plugin path` calls driver.ExtractPlugin. Likewise the P-6 resolver now lives in driver.ResolveSession and cli.resolveSession delegates to it; driver.ModelEffort(root, task|nil, phase) is what T4/T5 call. Runner fields: Root, Version, RunID, RunDir (absolute run folder; records go in RunDir/sessions), Log, plus test seams Claude, Now, Home, User, Env. Run returns Result{ExitCode, Recorded, Path, Cost, Turns, IsError, Denials}; a non-zero claude exit is not an error. Stdout that is empty or not JSON counts as missing: no record, log line only. The session counter continues after the highest NNN already in sessions/, so a resumed run does not overwrite. stderr goes to RunDir/<phase>-<iter>.stderr (masked), as the shell loop did. Raw stdout is not persisted. Runner.Mask also masks the JSON-escaped form of a Windows home.

## T4 — Add `vloop run` with its flags and preflight, the work branch, the plan session, the plan checks and the plan commit

- **Outcome:** done (review: PASS)
- **Summary:** `vloop run [<brief>]` now exists with --plan-only, --replan and the four budget flags: usage checks, brief-not-found (exit 2), doctor preflight, the R-2 work branch, the plan session through T3's Runner, the plan checks (state/v1 + validate, the three gate-shape rules, the head-diff advisory), the stamped plan, plan.md, journal and the `[vloop] plan <run id>` commit.
- **Files:** internal/cli/run.go, internal/cli/root.go, internal/cli/doctor.go, internal/driver/plan.go, internal/driver/gateshape.go, internal/driver/budgets.go, internal/driver/plan_test.go, cmd/vloop/run_plan_test.go, README.md, docs/guide/commands.md
- **Notes for next iteration:** Without --plan-only, run plans (or finds the plan) and then exits 1 with 'running its tasks is not available in this build' — T5 must replace that branch in internal/cli/run.go with the iteration loop; Planner.Plan returns PlanResult{Plan, Planned} for it, but a resume does not yet create a work branch or run folder (T5). Budgets are resolved and validated (driver.ResolveBudgets: flag > VLOOP_RUN_* > file > default) but nothing consumes them yet. Decisions: the driver stamps run_id (= brief stem, which names the branch, run folder, journal and commit) as well as brief/branch, sets status running, and leaves base as the session wrote it; a rejected plan is deleted from .vloop/state/state.json (else a re-run with the same brief would resume it) and nothing is committed, the run folder stays; plan problems print on stderr one per line then one `vloop:` line, schema pointers /tasks/N are prefixed with the task id; run's preflight is runDoctor minus its 'branch' check, plus the untrusted-workspace warning promoted to a problem; the pre-commit hook only warns. Commit is `git add -A` then `git reset -- .vloop/tmp` (an exclude pathspec fails when .vloop/tmp is gitignored). The run.log is committed, so the plan-only summary lines go to stdout only (logging them after the commit left the tree dirty). Exit code 9 (plan session moved refs) goes through cli.ExitError; driver.Halt carries the code. README: trimmed two lines of prose to stay under the 200-line cap; no exit-code or F1 edits (T10). Scenarios 21 and 30 are T5's.

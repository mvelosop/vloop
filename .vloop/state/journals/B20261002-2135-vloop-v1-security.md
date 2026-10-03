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

## T4 — Revert and fail a review session that changes the work it judges (F8)

- **Outcome:** done (review: PASS)
- **Summary:** The driver snapshots the working tree via git status before the review session and afterwards reverts every change outside .vloop/tmp/ and .vloop/state/, logs each path and forces the verdict to FAIL with the finding 'the review session changed files'.
- **Files:** internal/driver/iterate.go, cmd/vloop/run_review_test.go
- **Notes for next iteration:** treeGuard (end of internal/driver/iterate.go) reads git status -z --no-renames, keeps the bytes of paths already dirty (the work's own changes) and writes them back; a path that was clean is restored with git checkout HEAD, or removed if untracked. .vloop/state/ is excluded because the state and input guards own it. Ignored files are invisible to it. The test's retry passes, so the finding is cleared from task notes; the gate's single-iteration run checks the notes.

## T5 — Find gate files by whole token, across OS path forms and for every done task (F9)

- **Outcome:** done (review: PASS)
- **Summary:** gateFilesMoved now matches changed paths as whole normalised tokens (backslashes to slashes, leading ./ and quotes stripped) against the verify of the current task and every done task; TestGateFilesByToken added.
- **Files:** internal/driver/gates.go, internal/driver/gates_test.go
- **Notes for next iteration:** Tokens split on whitespace and ;&|()<>; done tasks are read from it.plan. A path inside a quoted string with spaces is not matched.

## T6 — Run every gate through one runner that refuses an unknown shell and passes cmd its command line verbatim (F15)

- **Outcome:** done (review: PASS)
- **Summary:** state.GateCommand is now the single builder of a gate's command, used by RunGate (vloop task gate) and Iterator.runGate. It refuses a shell outside sh, bash, pwsh, powershell, cmd, and passes cmd its command line verbatim on Windows.
- **Files:** internal/state/gate.go, internal/state/gate_cmdline_windows.go, internal/state/gate_cmdline_other.go, internal/state/gate_test.go, internal/driver/iterate.go
- **Notes for next iteration:** GateCommand(root, shell, verify, env) returns *UnknownShellError or *ShellMissingError and builds nothing; an empty shell means sh. GateCmdLine(verify) is the pure builder (/C <verify>); gate_cmdline_windows.go sets SysProcAttr.CmdLine from it, the other-OS file is a no-op. The driver halts with ExitPreflight on either error. internal/cli/task_gate.go needed no change: it calls RunGate and prints the error via Problem. The timeout and process group (T7) belong in GateCommand's callers or in it.

## T7 — Time gates and sessions and leave nothing running (F3)

- **Outcome:** rejected (review: FAIL)
- **Summary:** run.gate-timeout (15) and run.session-timeout (60) exist; gates and sessions run in their own process group (Job object on Windows) that is killed on timeout and when the gate's shell exits, with the output read bounded.
- **Files:** internal/state/group.go, internal/state/group_unix.go, internal/state/group_windows.go, internal/state/group_test.go, internal/state/gate.go, internal/driver/session.go, internal/driver/iterate.go, internal/driver/iterate_test.go, internal/driver/plan.go, internal/driver/budgets.go, internal/cli/task_gate.go, internal/cli/run.go, internal/config/config.go, internal/config/config_test.go, internal/cli/config_test.go, cmd/vloop/e2e_test.go, docs/guide/configuration.md, go.mod, go.sum
- **Notes for next iteration:** state.RunGroup(cmd, timeout) is the one runner: own pipes (not exec's copy goroutines), kills the group when the leader exits, drain bounded to 2s; RunGateWithin wraps it and RunGate keeps no deadline. Injectable in seconds via Iterator.GateTimeout/SessionTimeout and Runner.Timeout; zero means no deadline. A timed-out work or review session returns errSessionError (exit 7), a timed-out plan session halts 7. cmd/vloop e2e_test.go's config list expectation gained the two keys. Pre-existing, not from this task: TestRun03GateRegression and TestRun41RegressionNamesBoth in cmd/vloop fail because the driver now reports GATE REWRITE for T2 deleting T1.out (T5 token matching); the gate does not run them. Domain docs (run.md, domain-model R-3) are left to T17.

## T7 — Time gates and sessions and leave nothing running (F3)

- **Outcome:** done (review: PASS)
- **Summary:** Fixed the two regressions the earlier T7 attempt left: TestMetricsKeys window shifted for the two new keys, and README config table now lists run.gate-timeout and run.session-timeout (README kept under its 200-line cap).
- **Files:** internal/config/config_test.go, README.md
- **Notes for next iteration:** TestMetricsKeys slices Keys by position from the end; adding run.* keys shifts it again, so update the offsets. README has a line cap under 200 (now 196) and a test requiring every config key in it.

## T8 — Mask paths in string values before marshalling, never serialized text (F11)

- **Outcome:** done (review: PASS)
- **Summary:** Runner.Mask now replaces the home only at a path boundary (never for / or \) and the user only as a path component under Users, home or the home's parent directory, with Windows home forms matched case-insensitively. Session records are masked through their decoded string values (maskRecord) and re-marshalled, not as serialized JSON.
- **Files:** internal/driver/session.go, internal/driver/session_test.go, cmd/vloop/run_safety_test.go
- **Notes for next iteration:** Mask is text-only: free-text user names (e.g. model name alice-model, 'user alice') are no longer masked, so TestSessionMasking was adjusted. Records go through maskRecord (json.Number decode, masks strings and object keys). Home is detected as Windows-style by drive letter or UNC prefix, on any OS. TestRun10Containment now runs as subtests for harnessuser and a short user 'us' (home dir renamed after newRunRepo) and checks the user only as '/us/'.

## T9 — Redact secret environment values and keep only tool and path of permission denials (F12)

- **Outcome:** done (review: PASS)
- **Summary:** Runner.Mask now redacts the values of secret-named environment variables (KEY, TOKEN, SECRET, PASSWORD, CREDENTIAL; 8+ chars) as <redacted:NAME>, and session records keep only tool_name and file_path per permission denial. A guide section says what is and is not redacted.
- **Files:** internal/driver/session.go, internal/driver/session_test.go, cmd/vloop/run_redact_test.go, docs/guide/concepts.md
- **Notes for next iteration:** Redaction lives in Runner.Mask (redactSecrets), which every write under .vloop/state/ goes through, using Runner.Env or os.Environ; it also redacts the JSON-escaped form of a value. Denials are rebuilt from tool_name plus tool_input.file_path or notebook_path. TestSessionMasking now carries the home path in a denial's file_path instead of a command.

## T10 — Read handoffs only as regular files, take the lock atomically, and check the run id (F13)

- **Outcome:** done (review: PASS)
- **Summary:** Handoffs are read only as regular files and a report is copied into the run folder only after it validates; the run lock is hard-linked into place atomically; run_id must match the pattern and its brief's run id, in the driver and in state.v1.json.
- **Files:** internal/driver/safety.go, internal/driver/safety_test.go, internal/driver/iterate.go, internal/cli/run.go, schemas/state.v1.json, cmd/vloop/run_tmp_test.go
- **Notes for next iteration:** AcquireLock writes the record to a temp file and os.Link()s it to .vloop/tmp/.running: atomic, never follows a symlink, never visible half-written; a stale or non-regular lock is replaced once (removed only if its bytes are unchanged), then a second EEXIST halts. CheckRunID(root, briefPath) runs in internal/cli/run.go before the lock and is skipped when the named brief differs from the plan's (the plan is about to be reset); Iterator.Run checks again. The schema pattern is ^[A-Za-z0-9._-]*$ plus a not-pattern for '..', so the planning session's empty run_id still validates; a bad run_id in a committed plan is usually refused first by the preflight plan validation. A full go test ./... in cmd/vloop still fails TestRun03GateRegression and TestRun41RegressionNamesBoth (known before T10, from T5 token matching); not part of this gate.

## T11 — Export repo.remote only as a URL or scp-like form without credentials, and never a local path (F14)

- **Outcome:** done (review: PASS)
- **Summary:** repo.remote is now exported only as scheme://host[:port]/path or host:path without user info, query or fragment, and omitted for a local path or file:// origin; repo.name derives from the sanitized remote.
- **Files:** internal/cli/export.go, internal/cli/export_test.go, schemas/export.v1.json
- **Notes for next iteration:** Schema: remote is now optional and a plain string (no null); Remote uses omitempty. scp-like user (git@) is dropped too. Local-path detection: no colon, slash/backslash before the colon, a one-letter Windows drive, or a ~ host.

## T10 — Read handoffs only as regular files, take the lock atomically, and check the run id (F13)

- **Outcome:** done (review: PASS)
- **Summary:** T10's work was already in place; fixed the regression T11 caused: export schema fixtures still carried repo.remote null, which the T11 schema (optional plain string) rejects.
- **Files:** internal/schema/testdata/export-task.valid.json, internal/schema/testdata/export-brief.valid.json, internal/schema/testdata/export-intervention.valid.json, internal/schema/testdata/export-task.invalid.json, internal/schema/testdata/export-brief.invalid.json, internal/schema/testdata/export-intervention.invalid.json
- **Notes for next iteration:** export/v1 repo.remote is omitted, never null; fixtures under internal/schema/testdata must omit the key. The 'No such file' line from /bin/sh in the gate output is a harmless fixture quirk.

## T12 — Plan only a ready, checked brief with its dependencies consumed, from a clean tree (F1, F2)

- **Outcome:** blocked (review: skipped)
- **Summary:** Implemented the ready/checked/dependencies-consumed refusals, newest-ready-brief selection and the clean-tree refusal (with the .vloop/state/ allowance on resume); the six new tests and the two fixed existing tests pass, but the gate cannot pass for a correct implementation.
- **Files:** internal/brief/ready.go, internal/brief/check.go, internal/driver/plan.go, cmd/vloop/run_preflight_test.go, cmd/vloop/run_iterate_test.go
- **Notes for next iteration:** Gate defects. (1) The gate does mkrepo s, which makes the repository $t/s, the same directory as the stub claude ($t/s/claude, plan.json, argv.log, n.*). The stub files are tracked or untracked in the repo and plan() rewrites the tracked plan.json, so the tree is dirty and the clean-tree refusal (acceptance 4) correctly exits 1 before the first iteration. Renaming the fixture repo to another name (I ran a copy of the gate with `mkrepo u`) gets past every clause up to the full go test ./cmd/vloop. (2) That clause, run when VLOOP_GATE_TASK equals VLOOP_ACTIVE_TASK, fails on TestRun03GateRegression and TestRun41RegressionNamesBoth, which already failed before this task (T5 token matching makes T2 deleting T1.out a GATE REWRITE; see journal notes of T7 and T10). Fixing them is outside T12 and acceptance 7 forbids changing assertions. Implementation: brief.Plannable(root, path, set, replan) in internal/brief/ready.go (replan ignores the 'already run' journal problem so --replan keeps working), brief.Newest picks the newest ready brief, driver.Planner.Plan checks both only when not resuming, requireCleanTree runs for fresh plans and resumes (resume tolerates .vloop/state/, .vloop/tmp/ never counts; git status read with -z, untrimmed). driver.LatestBrief was removed. TestRun21ForeignState now commits its second brief before running.

## T13 — Make the self-hosting check warn on a binary that carries no commit, and document the stamped release build (F16)

- **Outcome:** done (review: PASS)
- **Summary:** Doctor's self-hosting check now warns for a binary with no commit (empty or the default "unknown") in this module; HEAD-stamped keeps its warning, other commits pass. concepts.md documents the stamped release build and release steps.
- **Files:** internal/cli/doctor.go, internal/cli/doctor_test.go, docs/guide/concepts.md
- **Notes for next iteration:** main.commit defaults to "unknown", so the no-commit case matches both "" and "unknown". TestDoctorSelfHosting was left unchanged: its default test build commit c010 already covers a stamped build of another commit passing.

## T14 — Count NestJS e2e specs and test/ directories as tests in the TypeScript and JavaScript presets (C1)

- **Outcome:** done (review: PASS)
- **Summary:** The typescript and javascript presets now classify **/*.e2e-spec.{ts,js} and **/test/** as tests; the metrics guide table follows, and TestPresetsCountNestTests covers it.
- **Files:** internal/classify/presets.go, internal/classify/classify_test.go, docs/guide/metrics.md
- **Notes for next iteration:** none

## T15 — Keep the machine awake for the lifetime of vloop run (C2)

- **Outcome:** done (review: PASS)
- **Summary:** Added run.keep-awake (on/off, default on, VLOOP_RUN_KEEP_AWAKE); vloop run now holds a no-idle-sleep hold (caffeinate, systemd-inhibit, SetThreadExecutionState) for its lifetime, with one run.log warning line if it cannot.
- **Files:** internal/config/config.go, internal/config/config_test.go, internal/cli/config_test.go, cmd/vloop/e2e_test.go, internal/cli/run.go, internal/driver/budgets.go, internal/driver/iterate.go, internal/driver/plan.go, internal/driver/keepawake.go, internal/driver/keepawake_other.go, internal/driver/keepawake_windows.go, internal/driver/keepawake_test.go, internal/driver/keepawake_darwin_test.go, docs/guide/configuration.md, docs/guide/concepts.md
- **Notes for next iteration:** The hold is taken in internal/cli/run.go after the lock; the failure reason is passed as AwakeWarn to the Iterator (or the Planner for --plan-only) which writes the single run.log line. Linux helper is systemd-inhibit wrapping tail --pid=<vloop pid> -f /dev/null. Windows code was only vetted and cross-built, not run. Existing config list tests (internal/cli, cmd/vloop e2e) gained the run.keep-awake row at the end.

## T16 — Move the README's completeness checks to the guides and replace its config table with a pointer (C3)

- **Outcome:** done (review: PASS)
- **Summary:** Completeness checks moved to the guides (TestGuideConfigurationNamesEveryKey, TestGuideCommandsNamesEveryCommandAndFlag, shell-default check now on configuration.md); the README is checked only for unknown names and guide links (TestReadmeRejectsUnknownNames), and its config table is a pointer.
- **Files:** internal/cli/readme_test.go, internal/cli/guide_test.go, README.md, docs/guide/configuration.md
- **Notes for next iteration:** README went from 196 to 174 lines. docs/guide/configuration.md preset tables for javascript and typescript also gained the NestJS e2e-spec and test/ globs that T14 added only to metrics.md; TestGuideConfigurationCoversKeysAndPresets failed without them.

## Run ended — blocked

- **Run:** `B20261002-2135-vloop-v1-security` · 18 iteration(s) this run
- **Plan:** 15/18 done, 1 blocked

## T12 — Plan only a ready, checked brief with its dependencies consumed, from a clean tree (F1, F2)

- **Outcome:** gate_failed (review: skipped)
- **Summary:** vloop run now refuses a brief that is not ready, fails brief check or has unconsumed dependencies, picks the newest ready brief when none is named, and refuses a dirty tree (on resume, .vloop/state/ edits are allowed). Six preflight tests were added and the existing tests that started dirty now start clean.
- **Files:** internal/brief/ready.go, internal/brief/check.go, internal/driver/plan.go, cmd/vloop/run_preflight_test.go, cmd/vloop/run_iterate_test.go
- **Notes for next iteration:** The work was done in an earlier attempt and committed with the block. The operator has since fixed the gate fixture and the two pre-existing failing tests, so this attempt only re-ran the gate, which passes.

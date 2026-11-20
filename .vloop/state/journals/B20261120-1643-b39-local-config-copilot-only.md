# Journal — B20261120-1643-b39-local-config-copilot-only

Append-only narrative of this plan. Rendered state lives in .vloop/state/plan.md.

## Plan — B20261120-1643-b39-local-config-copilot-only

- **Brief:** `docs/briefs/B20261120-1643-b39-local-config-copilot-only.loop-brief.md`
- **Tasks:** 8

## T1 — Add the local config layer to resolution, with source local

- **Outcome:** done (review: PASS)
- **Summary:** The config package now resolves four layers: VLOOP_* variable, .vloop/config.local.toml, .vloop/config.toml, default. Values from the local file report source local. A [[check]] in the local file is refused with exit 1 by Get, List and Checks.
- **Files:** internal/config/config.go, internal/config/root.go, internal/config/local_test.go
- **Notes for next iteration:** readLayers is the one read point, used by Get, List and Checks; ResolveRun reads through Get, so it is covered with no read point added. AddChecks and Set still read only the shared file. resolve(k, file, local) iterates layers lowest first, so a later layer overwrites; an empty list sets nothing in either layer. LocalFilePath sits beside FilePath in root.go. go test ./internal/cli: five TestTaskGate* tests fail inside this session with 'the plan was changed during this session' because the live session's plan guard trips them. That is the session environment, not this change; the config tests all pass.

## T2 — Add vloop config set --local, config path --local and the override notice

- **Outcome:** check_failed (review: skipped)
- **Summary:** config set gains --local (writes .vloop/config.local.toml with the same validation and layout preservation) and prints a one-line stderr notice when the local file or a VLOOP_* variable still overrides the key; config path --local added; commands.md regenerated.
- **Files:** internal/config/config.go, internal/cli/config.go, internal/cli/config_test.go, docs/guide/commands.md
- **Notes for next iteration:** Set is now a wrapper over setAt(root, rel, ...); SetLocal uses the local path. Notice is computed in the CLI from config.Get after the write: env source wins over local in the message; with --local only the env notice can appear. Get errors after a successful write are ignored for the notice.

## T2 — Add vloop config set --local, config path --local and the override notice

- **Outcome:** done (review: PASS)
- **Summary:** config set --local, config path --local and the stderr override notice exist; the check failure is fixed by wording config set's Short to name 'config list' again; commands.md regenerated.
- **Files:** internal/config/config.go, internal/cli/config.go, internal/cli/config_test.go, docs/guide/commands.md
- **Notes for next iteration:** Previous check failure was TestShortsFollowListAndShow requiring config set's Short to contain 'config list'. TestTaskGate* failures inside a live session are the plan guard, not this change.

## T3 — Keep the local file git-ignored and in view: init, upgrade, doctor and the run's preflight

- **Outcome:** check_failed (review: skipped)
- **Summary:** vloop init and upgrade keep .vloop/config.local.toml in .gitignore; doctor has a config.local row (tracked / not ignored problems, warning naming overridden keys); vloop run's preflight refuses on the problems and prints the warning.
- **Files:** internal/install/section.go, internal/cli/doctor.go, internal/cli/run.go, internal/cli/init_test.go, internal/cli/upgrade_test.go, cmd/vloop/b39_local_config_test.go
- **Notes for next iteration:** MergeGitignore adds config.local line BEFORE .vloop/tmp/ when both are missing: the gate commits a .gitignore of tmp then local right after init, which fails with 'nothing to commit' if init already wrote exactly that order. Order is otherwise arbitrary. Two existing tests (TestInitClaudeMarkers, TestUpgradeMarkerOnlyAndGitignore) changed only to include the new line. TestTaskGate* fail inside a live session (plan guard), not this change.

## T3 — Keep the local file git-ignored and in view: init, upgrade, doctor and the run's preflight

- **Outcome:** done (review: PASS)
- **Summary:** vloop init and upgrade keep .vloop/config.local.toml in .gitignore; doctor has a config.local row (tracked and not-ignored problems, a warning naming overridden keys); vloop run's preflight refuses on those problems and prints the warning. Fixed the previous check failure in two init tests.
- **Files:** internal/install/section.go, internal/cli/doctor.go, internal/cli/run.go, internal/cli/init_test.go, internal/cli/upgrade_test.go, cmd/vloop/b39_local_config_test.go
- **Notes for next iteration:** Previous check failure: TestInitWritesEverything now expects the local line in .gitignore; TestInitKeepsConfigAndBriefs fixture now already holds both lines so 'kept .gitignore' still holds (both pin the contract item that init adds the new line). MergeGitignore writes the local line before .vloop/tmp/ when both are missing, which the gate needs. Five TestTaskGate* tests fail only inside a live session (plan guard), not from this change.

## T4 — Keep sessions off the local file: the input guard and the fences

- **Outcome:** done (review: PASS)
- **Summary:** The input guard now holds .vloop/config.local.toml like config.toml (changed, created, removed are put back); the five fences that deny config.toml deny Edit/Write of the local file too; fence-audit has matching rows in each section.
- **Files:** internal/driver/safety.go, internal/driver/safety_test.go, fence/work.json, fence/review.json, fence/verify.json, fence/gate-review.json, fence/review-brief.json, docs/design-notes/fence-audit.md
- **Notes for next iteration:** Guard change is one added root in snapshotInputs. New test TestInputGuardRestoresLocalConfig covers changed/removed/created. Five TestTaskGate* tests in internal/cli fail only inside a live session (plan guard), not from this change.

## T5 — Check Claude only where a session in use runs on it

- **Outcome:** done (review: PASS)
- **Summary:** doctor's claude, trust, plugin and gate edits rows are n/a (no session in use runs on claude) unless a session in use resolves to claude, so vloop run's preflight passes on a machine with no claude. An unreadable config counts Claude as in use.
- **Files:** internal/cli/doctor.go, cmd/vloop/b39_local_config_test.go
- **Notes for next iteration:** claudeInUse(root, withReviewBrief) in doctor.go mirrors copilotKeys; agent.review-brief counts only for the bare doctor (runBrief == nil), so vloop run ignores it. No earlier test needed changing. New e2e tests in cmd/vloop/b39_local_config_test.go use a PATH of the stub copilot plus symlinks to a few tools, with no claude. Five TestTaskGate* tests in internal/cli fail only inside a live session (plan guard), not from this change.

## T6 — Price a Copilot credit at GitHub's $0.01 by default

- **Outcome:** done (review: PASS)
- **Summary:** copilot.credit-usd now defaults to 0.01; the configuration guide's row and every test that pins the default say 0.01.
- **Files:** internal/config/config.go, internal/config/config_test.go, internal/cli/config_test.go, cmd/vloop/e2e_test.go, docs/guide/configuration.md
- **Notes for next iteration:** Only the default literal changed in tests. In internal/cli, TestTaskGate* tests fail inside a work session with 'the plan was changed during this session' (the session guard, unrelated to this change); the other cli tests pass. 0.0066667 remains only in docs/briefs, .vloop/state and docs/design-notes.

## T7 — Document the local file, Claude only where used, and the credit price

- **Outcome:** done (review: PASS)
- **Summary:** Documented the local config file in the configuration guide (resolution order, new section, config set --local example), C-1, the platform context, the concepts guide, the README and the roadmap line; Claude-only-where-used is in the concepts doctor paragraph and README.
- **Files:** README.md, docs/domain/domain-model.md, docs/domain/platform/platform-context.md, docs/design-notes/vloop-roadmap.md, docs/guide/concepts.md, docs/guide/configuration.md
- **Notes for next iteration:** The copilot.credit-usd row in configuration.md already said 0.01 (T6), so it was not touched. I did not run go test ./... (it passed the 120s foreground limit); the README line-cap test was not run separately, but wc -l gives 195.

## T8 — Close: the worked example line for line, and every gate at once

- **Outcome:** done (review: PASS)
- **Summary:** Added an end-to-end test that walks the worked example's first block in order, through the last config.local row and config get copilot.credit-usd printing 0.01. No product discrepancy found.
- **Files:** cmd/vloop/b39_worked_example_test.go
- **Notes for next iteration:** New test TestWorkedExampleLocalConfigFirstBlock uses the scratch helper in e2e_test.go. The other blocks (planted failures, Claude not in use) are covered by T3/T5 tests in cmd/vloop/b39_local_config_test.go.

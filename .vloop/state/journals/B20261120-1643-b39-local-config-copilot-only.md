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

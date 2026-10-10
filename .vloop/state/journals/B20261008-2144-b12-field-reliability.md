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

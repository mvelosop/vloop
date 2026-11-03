# Journal — B20261103-0528-b36-i96-i97-plugin-per-repo-and-agent

Append-only narrative of this plan. Rendered state lives in .vloop/state/plan.md.

## Plan — B20261103-0528-b36-i96-i97-plugin-per-repo-and-agent

- **Brief:** `docs/briefs/B20261103-0528-b36-i96-i97-plugin-per-repo-and-agent.loop-brief.md`
- **Tasks:** 4

## T1 — Declare the plugin in .claude/settings.json from vloop init and vloop upgrade (P1)

- **Outcome:** check_failed (review: skipped)
- **Summary:** vloop init and vloop upgrade now merge extraKnownMarketplaces.vloop and enabledPlugins["vloop@vloop"] into .claude/settings.json (order-preserving, existing values kept, refused with exit 1 when not an object), via the new install.MergeClaudeSettings.
- **Files:** internal/install/settings.go, internal/cli/init.go, internal/cli/upgrade.go, internal/cli/init_test.go, internal/cli/upgrade_test.go, internal/cli/doctor_test.go
- **Notes for next iteration:** install/settings.go reads the file with json.Decoder (UseNumber) into an ordered tree and writes it MarshalIndent-style with HTML escaping off; the file is rewritten only when an entry was added. Upgrade's fixtures run init, so the settings file already exists there and only tests that remove .claude/ see the 'updated' line. TestDoctorGateEditsNoClaudeAndWritesNothing changed per the brief's Constraints (it now pre-seeds the declaration and asserts .claude/ holds only settings.json unchanged). TestTaskGate* in internal/cli fail inside a work session with 'the plan was changed during this session' (environment of the session, not touched by this task); not investigated.

## T1 — Declare the plugin in .claude/settings.json from vloop init and vloop upgrade (P1)

- **Outcome:** done (review: PASS)
- **Summary:** vloop init and vloop upgrade merge extraKnownMarketplaces.vloop and enabledPlugins["vloop@vloop"] into .claude/settings.json via install.MergeClaudeSettings. The previous check failure (TestWorkedExampleB5Session pinning init's 17 output lines) is fixed.
- **Files:** internal/install/settings.go, internal/cli/init.go, internal/cli/upgrade.go, internal/cli/init_test.go, internal/cli/upgrade_test.go, internal/cli/doctor_test.go, cmd/vloop/b5_e2e_test.go
- **Notes for next iteration:** cmd/vloop/b5_e2e_test.go changed only where it pins init's output lines: 18 lines now, with 'wrote .claude/settings.json' after 'wrote CLAUDE.md' (brief Constraints). That test takes about 215s.

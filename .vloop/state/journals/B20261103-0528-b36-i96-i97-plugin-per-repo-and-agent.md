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

## T2 — Report the plugin per agent in vloop doctor: the plugin row for this repository, and a copilot plugin row (P3, P4)

- **Outcome:** check_failed (review: skipped)
- **Summary:** Doctor's plugin row now counts only this repository's vloop entries (user/absent scope or projectPath matching the root), honours the .claude/settings(.local).json opt-out, and uses P3's messages; a new copilot plugin row (P4) runs copilot plugin list --json in the root, only where the copilot row appears, and never in the run preflight.
- **Files:** internal/cli/doctor.go, internal/cli/doctor_test.go
- **Notes for next iteration:** Both lists run via runPluginList with cmd.Dir=root. Preflight is told apart by runBrief != nil in runDoctor (copilot plugin row skipped, so copilot plugin list never runs). TestDoctorPluginMessages changed only where P3 changes messages. New tests: TestDoctorPluginScopeP3, TestDoctorCopilotPluginP4, TestDoctorNoCopilotPluginRow, TestPreflightRunsNoCopilotPluginList; stubs record pwd -P. TestTaskGate* fail inside a work session with 'the plan was changed during this session' (same as T1's note), not touched.

## T2 — Report the plugin per agent in vloop doctor: the plugin row for this repository, and a copilot plugin row (P3, P4)

- **Outcome:** done (review: PASS)
- **Summary:** Doctor's plugin row counts only this repository's vloop entries, honours the .claude/settings(.local).json opt-out and uses P3's messages. A new copilot plugin row (P4) runs copilot plugin list --json in the root, only where the copilot row appears, and never in the run preflight. This attempt fixes the two cmd/vloop e2e tests that failed the go check.
- **Files:** internal/cli/doctor.go, internal/cli/doctor_test.go, cmd/vloop/b5_e2e_test.go, cmd/vloop/b13_e2e_test.go
- **Notes for next iteration:** cmd/vloop/b5_e2e_test.go (plugin_version_warning line) and b13_e2e_test.go (pluginWarn) changed only to P3's new message text, as the brief's Constraints allow. The rest of the go check was green in the previous failing run (only these two failed). Preflight is told apart from doctor by runBrief != nil in runDoctor.

# Journal — B20261102-0048-b28-i101-brief-skills-in-plugin

Append-only narrative of this plan. Rendered state lives in .vloop/state/plan.md.

## Plan — B20261102-0048-b28-i101-brief-skills-in-plugin

- **Brief:** `docs/briefs/B20261102-0048-b28-i101-brief-skills-in-plugin.loop-brief.md`
- **Tasks:** 6

## T1 — Move the two brief skills into the plugin as create-brief and review-brief, with their budgets

- **Outcome:** done (review: PASS)
- **Summary:** plugin/skills/ now holds create-brief and review-brief (copied from the project skills, renamed, experimental label dropped, pointers and step 5.1 updated), operate's description names /vloop:create-brief, and budget lists both skills with ceilings 1200 and 1000.
- **Files:** plugin/skills/create-brief/SKILL.md, plugin/skills/create-brief/guidelines.md, plugin/skills/create-brief/templates/brief-en.md, plugin/skills/create-brief/templates/brief-es.md, plugin/skills/review-brief/SKILL.md, plugin/skills/operate/SKILL.md, internal/budget/budget.go, internal/budget/budget_test.go, internal/cli/doctor_copilot_review_brief_test.go, cmd/vloop/b7_e2e_test.go
- **Notes for next iteration:** Only earlier test changed: TestWorkedExampleB7PluginPath in cmd/vloop/b7_e2e_test.go, to the eight skill names (K1, the eight folders). New tests: TestCheckCreateBriefOverCeiling (budget) and TestCopilotReviewSkillNamedReviewBriefClashes (internal/cli). TestTaskGate* tests in internal/cli fail when run inside a loop session (the plan-changed guard), not because of this change; the driver's gate pass runs outside it.

## T2 — Add the Go test that holds each plugin brief template to the binary's

- **Outcome:** done (review: PASS)
- **Summary:** internal/cli/plugin_templates_test.go adds checkPluginTemplates(pluginDir, binDir), a test on the repository's own files and fixture tests on temp copies covering each planted failure.
- **Files:** internal/cli/plugin_templates_test.go
- **Notes for next iteration:** The executive summary is cut from its heading up to the next line starting '## ' before the byte comparison. No template edited, no earlier test changed.

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

## T3 — Name /vloop:create-brief and /vloop:review-brief in the CLAUDE.md section vloop writes

- **Outcome:** done (review: PASS)
- **Summary:** install.Section's interactive-session paragraph now also sends a session to /vloop:create-brief to write a brief and /vloop:review-brief to review one in a separate session; the pinning test asserts the new paragraph.
- **Files:** internal/install/section.go, internal/install/install_test.go
- **Notes for next iteration:** internal/install/install_test.go line 108 changed only to the new paragraph (contract item K3). go test ./internal/cli had one failure, TestTaskGateUnknownTask ('plan was changed during this session'), which comes from running inside a driver session; not touched by this task.

## T4 — Add an eval case for each brief skill and update the evals README and guide

- **Outcome:** done (review: PASS)
- **Summary:** Added the eval cases review-brief-planted-faults and create-brief-drafts-checked (case.yaml, prompt.md, scaffold.sh, graders), their rows in plugin/evals/README.md, and updated docs/guide/evals.md (21 cases, suites table, operate line removed).
- **Files:** plugin/evals/review-brief-planted-faults/case.yaml, plugin/evals/review-brief-planted-faults/prompt.md, plugin/evals/review-brief-planted-faults/scaffold.sh, plugin/evals/review-brief-planted-faults/graders, plugin/evals/create-brief-drafts-checked/case.yaml, plugin/evals/create-brief-drafts-checked/prompt.md, plugin/evals/create-brief-drafts-checked/scaffold.sh, plugin/evals/create-brief-drafts-checked/graders, plugin/evals/README.md, docs/guide/evals.md
- **Notes for next iteration:** create-brief-drafts-checked grants Bash(vloop brief:*) beyond the fence; added 'Bash(vloop brief:*)' to the guide's --allow-tools list so TestEvalGrantsCoverFence holds. The planted brief passes vloop brief check once status is ready; its plant: GREET_NAME claim, usage message differing between contract and worked example, tests/golden.sh red after task 1 (usage.txt regenerated only in task 2). No eval was run.

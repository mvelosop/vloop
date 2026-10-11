# Journal — B20261011-2100-b13-docs-onboarding

Append-only narrative of this plan. Rendered state lives in .vloop/state/plan.md.

## Plan — B20261011-2100-b13-docs-onboarding

- **Brief:** `docs/briefs/B20261011-2100-b13-docs-onboarding.loop-brief.md`
- **Tasks:** 7

## T1 — Make doctor's plugin check warn when no vloop plugin is listed

- **Outcome:** done (review: PASS)
- **Summary:** doctorPlugin now returns a warning with the D1 message when no vloop@<any> entry is listed (empty list or other-marketplace-only); all other outcomes are unchanged.
- **Files:** internal/cli/doctor.go, internal/cli/doctor_test.go, cmd/vloop/b20261005_e2e_test.go
- **Notes for next iteration:** cmd/vloop/b20261005_e2e_test.go asserted the old pass message for [] (doctor with a fake claude listing none); changed only that expectation to the new warning, as it would otherwise fail. TestDoctorPlugin/TestDoctorPluginMessages gained an other@x-only case. The task_gate_test failures seen in a full go test run from inside a loop session (plan changed during this session) come from running within a live session, not from this change; TestDoctor* and cmd/vloop pass.

## T2 — Add doctor's `gate edits` check of the operator's permission to edit gate folders

- **Outcome:** done (review: PASS)
- **Summary:** doctor now has a `gate edits` check right after `plugin` that reads the user, project and local Claude settings and warns, printing the rule to add, unless an allow entry lets a session edit .vloop/state/gates/; it only reads.
- **Files:** internal/cli/doctor.go, internal/cli/doctor_test.go, cmd/vloop/b5_e2e_test.go
- **Notes for next iteration:** Home comes from the package var userHomeDir (defaults to os.UserHomeDir) in internal/cli/doctor.go; tests override it. Files are read in order user, project, local: the first unreadable file before a pass gives the `cannot read` warning, an earlier pass wins over a later unreadable file. Besides the tests the brief lists, TestWorkedExampleB5PlantedFailures/plugin_version_warning (cmd/vloop/b5_e2e_test.go) changed from 1 to 2 warnings, as the new line adds one there too.

## T3 — Make init's closing lines and the CLAUDE.md section say how to get /vloop:operate

- **Outcome:** done (review: PASS)
- **Summary:** vloop init now ends with D3's seven lines (install commands, --plugin-dir, ask rather than type), and the CLAUDE.md section's interactive paragraph carries D3's text about an unloaded plugin.
- **Files:** internal/cli/init.go, internal/install/section.go, internal/cli/init_test.go, cmd/vloop/b5_e2e_test.go, internal/install/install_test.go, internal/cli/b11_verify_test.go
- **Notes for next iteration:** TestInitDoesNotPointAtAMarketplace (internal/cli/b11_verify_test.go) asserted init does NOT mention marketplace, which D3 reverses; renamed to TestInitPointsAtBothWaysToLoadThePlugin and now asserts the three commands are present. b5_e2e_test.go line count for init output went 12 -> 16. No help text changed, so docs/guide/commands.md untouched; this repo's CLAUDE.md not rewritten. task_gate_test failures in a full internal/cli run come from running inside a live session (plan held by the driver), not from this change.

## T4 — Explain the name, PATH, and asking rather than typing in the README

- **Outcome:** done (review: PASS)
- **Summary:** README now explains the name (verification loop), the PATH note after go install with a link to the guide, a 'You ask; the skill drives' section before 'Everyday commands', and links 'driver' to the guide.
- **Files:** README.md
- **Notes for next iteration:** The 5 TestTaskGate* failures say 'the plan was changed during this session'; they fail because the tests run inside a driver-held session and do not touch the README. The PATH link targets docs/guide/concepts.md#setting-up-a-repository (existing heading, which T5 extends); the-driver anchor does not exist until T5 adds it.

## T5 — Define the driver, PATH, plugin loading, gate-folder lifecycle and retiring the shell loop in the guide

- **Outcome:** done (review: PASS)
- **Summary:** The concepts guide now has '## The driver' (with a who-does-what table), PATH per shell, plugin loading, gate-folder lifecycle, the doctor 'gate edits' and plugin warning, and '## Retiring the shell loop'; both guides say vloop reads .loop/state/ in any repository that has it.
- **Files:** docs/guide/concepts.md, docs/guide/metrics.md
- **Notes for next iteration:** Anchors added: #the-driver, #retiring-the-shell-loop; #setting-up-a-repository now carries the PATH text. No plugin skill was edited; their uses of 'the driver' (applies verdicts, runs gates, keeps logs) agree with the definition.

## T6 — Say in CLAUDE.md that this repository's operator playbook wins here, and link the glossary's driver

- **Outcome:** done (review: PASS)
- **Summary:** CLAUDE.md's 'Which rules bind you' now says this repository's operator playbook (.claude/skills/vloop-operator/SKILL.md) is the one to follow where the vloop section names /vloop:operate; the glossary's driver row links to ../guide/concepts.md#the-driver.
- **Files:** CLAUDE.md, docs/domain/domain-model.md
- **Notes for next iteration:** The link sits in the definition cell because the gate requires the row to start with '| **driver** |'.

# Journal — B20261126-1418-b45-docs-mixed-os-and-any-agent

Append-only narrative of this plan. Rendered state lives in .vloop/state/plan.md.

## Plan — B20261126-1418-b45-docs-mixed-os-and-any-agent

- **Brief:** `docs/briefs/B20261126-1418-b45-docs-mixed-os-and-any-agent.loop-brief.md`
- **Tasks:** 6

## T1 — Document the mai-code-2 evidence

- **Outcome:** done (review: PASS)
- **Summary:** Added the mai-code-2 loop evidence section with the external URL shortener measurement, B40-B44 committed metrics, session coverage, failure evidence, brief-review records, and an explicit non-qualification conclusion.
- **Files:** docs/design-notes/copilot-cli-experiments.md
- **Notes for next iteration:** B40-B44 figures use each committed .vloop/state/metrics/ snapshot's tasks.planned, tasks.first_pass and cost.total_usd rounded to cents; B42's $3.05 includes its halted first attempt. The URL shortener figures are explicitly identified as the operator's measurement in an acceptance repository outside this repository.

## T2 — Make the README agent-neutral and cite mai-code-2

- **Outcome:** check_failed (review: skipped)
- **Summary:** README now describes both Claude Code and GitHub Copilot CLI throughout onboarding and the loop, and links mai-code-2's run evidence while explicitly leaving it not yet qualified.
- **Files:** README.md
- **Notes for next iteration:** Preserved the gpt-6.1-terra limitation to work, review, second review and brief review, with plan and gate review kept on Claude and verify still untried; mai-code-2's evidence covers every session kind except second review.

## T2 — Make the README agent-neutral and cite mai-code-2

- **Outcome:** done (review: PASS)
- **Summary:** README onboarding and loop guidance now describe Claude Code and GitHub Copilot CLI consistently, with Copilot prerequisites and mai-code-2 run evidence linked without claiming qualification.
- **Files:** README.md
- **Notes for next iteration:** Retained the gpt-6.1-terra limits: plan and gate review stay on Claude and verify remains untried. The mai-code-2 paragraph covers every session kind except second review and links the unchanged eval qualification rule and D6 evidence.

## T3 — Explain mixed teams in the README

- **Outcome:** check_failed (review: skipped)
- **Summary:** README.md now has a concise mixed macOS and Windows team section, Copilot-specific Windows evidence, and accurate PowerShell install and update guidance.
- **Files:** README.md
- **Notes for next iteration:** The mixed-team section uses the shared platforms and pwsh declarations, summarizes the fixed portability lint and UTF-8 execution, and links to the full configuration guide; the Copilot bullet now retains only its PowerShell hook and B40 through B44 evidence.

## T3 — Explain mixed teams in the README

- **Outcome:** done (review: PASS)
- **Summary:** README.md now has a concise mixed macOS and Windows team entry point, Copilot-specific Windows evidence, and accurate PowerShell install and upgrade guidance while staying under the repository line cap.
- **Files:** README.md
- **Notes for next iteration:** The mixed-team prose remains consistent with the configuration guide and covers shared platforms and pwsh, rejected overrides, fixed-command linting, portable pwsh/toolchain commands, UTF-8, and this repository; prose was compacted without weakening T2's Copilot evidence to satisfy the README line cap.

## T4 — Update the guides for any agent

- **Outcome:** done (review: PASS)
- **Summary:** Updated the four guides to describe shared agent behavior while preserving the distinct Claude and Copilot adapters, installation paths, metrics, and model values. Added mixed-team portable gate guidance and agent-aware eval replay instructions.
- **Files:** docs/guide/concepts.md, docs/guide/metrics.md, docs/guide/configuration.md, docs/guide/evals.md
- **Notes for next iteration:** The guides now distinguish Claude's appended system prompt from Copilot's custom-instructions directory, document Copilot's credit-based estimated cost and token absence, and keep Claude-specific plugin evaluation and qualification wording scoped.

## T5 — List every session kind and portable gate shell

- **Outcome:** done (review: PASS)
- **Summary:** Updated the generated and repository CLAUDE.md sections to list all six loop session kinds, and aligned the operator and architect guidance with portable pwsh gates on this mixed macOS/Windows team.
- **Files:** internal/install/section.go, internal/install/install_test.go, CLAUDE.md, .claude/skills/vloop-operator/SKILL.md, .claude/skills/vloop-architect/SKILL.md
- **Notes for next iteration:** The only Go test change updates the existing generated-section pin to the full six-kind list; no gate fixtures or unrelated generated text changed. BSD-tool cautions now apply only to plans whose shell is sh.

## T6 — Run the complete documentation contract

- **Outcome:** done (review: PASS)
- **Summary:** The completed documentation now satisfies all five earlier gates and the brief's worked-example contract without additional source changes.
- **Files:** none
- **Notes for next iteration:** The combined gate ran T1 through T5 together, then rechecked the stale-string absences, mixed-team heading, mai-code-2 and self-update evidence, Copilot plugin-dir example, model-value wording, six-kind generated section, and operator pwsh shell.

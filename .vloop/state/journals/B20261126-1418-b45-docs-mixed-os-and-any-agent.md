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

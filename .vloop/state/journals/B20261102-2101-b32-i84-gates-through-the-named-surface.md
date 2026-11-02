# Journal — B20261102-2101-b32-i84-gates-through-the-named-surface

Append-only narrative of this plan. Rendered state lives in .vloop/state/plan.md.

## Plan — B20261102-2101-b32-i84-gates-through-the-named-surface

- **Brief:** `docs/briefs/B20261102-2101-b32-i84-gates-through-the-named-surface.loop-brief.md`
- **Tasks:** 4

## T1 — Make the plan skill judge through the surface the brief names, and act on every revision finding (G1, G2)

- **Outcome:** done (review: PASS)
- **Summary:** The plan skill's Naming paragraph now says to judge through the surface the brief names (G1), two dependent phrases are adjusted, and A revision round has the two G2 sentences.
- **Files:** plugin/skills/plan/SKILL.md
- **Notes for next iteration:** Only the four edits in plugin/skills/plan/SKILL.md; no other file changed.

## T2 — Say in the gate review when a pin is sound, and raise its word ceiling to 1600 (G3)

- **Outcome:** done (review: PASS)
- **Summary:** The gate-review skill's pins-unnamed bullet now has the sound-pin sentence and not-contract names the sound-pin exception. The gate-review ceiling in internal/budget/budget.go is 1600 with the dated comment.
- **Files:** plugin/skills/gate-review/SKILL.md, internal/budget/budget.go
- **Notes for next iteration:** Only the two brief edits were made; no test was changed. I started the full go test ./... once and stopped it before it finished, so the driver's gate pass is the full-suite check.

## T3 — Add the eval case plan-judges-through-named-surface and list it in both catalogues (G4)

- **Outcome:** done (review: PASS)
- **Summary:** Added the eval case plan-judges-through-named-surface (prompt, case.yaml, brief, scaffold, two graders) and listed it in plugin/evals/README.md and docs/guide/evals.md, whose case count is now 23.
- **Files:** plugin/evals/plan-judges-through-named-surface/prompt.md, plugin/evals/plan-judges-through-named-surface/case.yaml, plugin/evals/plan-judges-through-named-surface/brief.md, plugin/evals/plan-judges-through-named-surface/scaffold.sh, plugin/evals/plan-judges-through-named-surface/graders/state-validates.md, plugin/evals/plan-judges-through-named-surface/graders/judges-through-the-command.md, plugin/evals/README.md, docs/guide/evals.md
- **Notes for next iteration:** The brief names no library path or function and does not mention the bad gate. No committed case list needed updating: TestEvalGrantsCoverFence reads the case directories.

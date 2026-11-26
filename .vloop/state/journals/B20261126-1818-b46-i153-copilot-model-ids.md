# Journal — B20261126-1818-b46-i153-copilot-model-ids

Append-only narrative of this plan. Rendered state lives in .vloop/state/plan.md.

## Plan — B20261126-1818-b46-i153-copilot-model-ids

- **Brief:** `docs/briefs/B20261126-1818-b46-i153-copilot-model-ids.loop-brief.md`
- **Tasks:** 6

## T1 — List Copilot model ids and the account plan

- **Outcome:** check_failed (review: skipped)
- **Summary:** Added `vloop models copilot` with picker filtering, aligned text output, one `copilot-models/v1` JSON document, account-plan lookup, and a reusable standard-library Copilot API client.
- **Files:** docs/guide/commands.md, embed_test.go, internal/cli/models.go, internal/cli/models_test.go, internal/cli/root.go, internal/cli/schema_test.go, internal/copilot/client.go, internal/schema/schema_test.go, schemas/copilot-models.v1.json
- **Notes for next iteration:** The shared client resolves COPILOT_GITHUB_TOKEN, GH_TOKEN, GITHUB_TOKEN, then gh auth token; model and plan calls each use their own configurable base and 20-second client timeout. Models failures are fatal to the listing command, while plan failures intentionally produce an unknown plan.

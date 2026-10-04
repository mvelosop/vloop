# Journal — B20261004-1434-interventions-options

Append-only narrative of this plan. Rendered state lives in .vloop/state/plan.md.

## Plan — B20261004-1434-interventions-options

- **Brief:** `docs/briefs/B20261004-1434-interventions-options.loop-brief.md`
- **Tasks:** 11

## T1 — Version the intervention record to intervention/v2: the schema, reading v1 and v2, writing v2, the derived agreement

- **Outcome:** check_failed (review: skipped)
- **Summary:** intervention/v2 exists: schemas/intervention.v2.json, a parser that reads v1 and v2 (v1 Context, Suggested and Decided no longer fold into done), add writing v2, and one DeriveAgreement that list re-derives, failing on disagreeing records.
- **Files:** schemas/intervention.v2.json, internal/intervention/intervention.go, internal/intervention/intervention_test.go, internal/cli/schema_test.go, internal/schema/schema_test.go, cmd/vloop/b2_e2e_test.go
- **Notes for next iteration:** Intervention carries an unexported decision (stored decided: 1|2|3|other|"") that render writes back; Decided.Option is 0 for other or none. v1 records keep Schema intervention/v1 and ignore **Options.** and **Recommended.** markers. Empty sections are now omitted on write. Set does not touch options or agreement yet (T3). Some internal/cli TestTaskGate* tests fail when run inside this session ('plan was changed during this session'), an environment effect, not from this change; the narrower Schema/Intervention/Guide/Export tests pass.

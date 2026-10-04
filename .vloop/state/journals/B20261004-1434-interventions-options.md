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

## T1 — Version the intervention record to intervention/v2: the schema, reading v1 and v2, writing v2, the derived agreement

- **Outcome:** check_failed (review: skipped)
- **Summary:** The T1 work from the earlier attempt (intervention/v2 schema, parser, add writing v2, derived agreement) is kept. This attempt fixed the check failure: TestSchemasAreEmbedded now expects 16 embedded schemas.
- **Files:** embed_test.go
- **Notes for next iteration:** The check failure was embed_test.go counting 15 schema files; intervention.v2.json makes 16. TestWorkedExampleB6RealData and the internal/cli TestTaskGate* tests fail only when run inside a loop session: the clone carries the live plan's gate folders, or the plan counts as changed. They are not caused by this change.

## T1 — Version the intervention record to intervention/v2: the schema, reading v1 and v2, writing v2, the derived agreement

- **Outcome:** done (review: PASS)
- **Summary:** intervention/v2 (schema, parser reading v1 and v2, add writing v2, one derived agreement) is in place from earlier attempts; this attempt fixed the remaining check failure in TestWorkedExampleB6RealData.
- **Files:** cmd/vloop/b6_e2e_test.go
- **Notes for next iteration:** TestWorkedExampleB6RealData clones HEAD, which carries the live plan's committed .vloop/state/gates/T*; the plan check refused them as gates for tasks not in the fresh plan. The test now removes the clone's .vloop/state/gates beside state.json. The change adds one cleanup step and weakens no assertion.

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

## T2 — Record options with vloop intervention add: --context, --option, --recommended, --why, --decided-option, --decided-other, --adjusted, --decided

- **Outcome:** check_failed (review: skipped)
- **Summary:** vloop intervention add takes --context, --option (max three), --recommended, --why, --decided-option, --decided-other, --adjusted and --decided; it writes the v2 record with the derived agreement and refuses malformed combinations with the brief's messages (exit 2, nothing written). list text gains an agreement column.
- **Files:** internal/intervention/intervention.go, internal/cli/intervention.go, internal/cli/intervention_test.go
- **Notes for next iteration:** Validation lives in intervention.CheckOptions, called by the CLI before the root is resolved and again by Add. Flag presence uses RecommendedSet/DecidedOptionSet (cmd.Flags().Changed) so --recommended 0 is refused, not read as unset. list text is now phase  kind  automatable  agreement  id.

## T2 — Record options with vloop intervention add: --context, --option, --recommended, --why, --decided-option, --decided-other, --adjusted, --decided

- **Outcome:** done (review: PASS)
- **Summary:** vloop intervention add takes the options flags and refuses malformed combinations with the brief's messages; the previous check failure is fixed by regenerating docs/guide/commands.md, which now documents the new flags.
- **Files:** internal/intervention/intervention.go, internal/cli/intervention.go, internal/cli/intervention_test.go, docs/guide/commands.md
- **Notes for next iteration:** The earlier check failure was docs/guide/commands.md being stale: run go generate ./... after adding flags. TestTaskGate* in internal/cli fail only inside a loop session (plan changed during this session), not from this change.

## T3 — Correct a record with vloop intervention set: recommended, decided and adjusted, re-deriving agreement

- **Outcome:** done (review: PASS)
- **Summary:** vloop intervention set now accepts recommended, decided and adjusted, validated against the record's option count and rewriting agreement (T1's derivation) in the same write; agreement and options are refused with the brief's messages.
- **Files:** internal/intervention/intervention.go, internal/cli/intervention_test.go, docs/guide/commands.md
- **Notes for next iteration:** Set updates frontmatter lines in place (adjusted: false deletes the line; deciding other clears adjusted). The agreement/options refusals are plain errors from intervention.Set, so exit 2; value or no-options refusals on the new fields are Problem errors (exit 1) and leave the file untouched. docs/guide/commands.md regenerated with go generate ./... because the set help text lists SetFields.

## T4 — Add vloop intervention migrate: v1 records to intervention/v2, frontmatter only

- **Outcome:** done (review: PASS)
- **Summary:** vloop intervention migrate inserts the five v2 no-options lines after by: in every v1 record and changes no other byte; --dry-run lists, a second run prints nothing to migrate, and upgrade leaves records alone.
- **Files:** internal/intervention/intervention.go, internal/cli/intervention.go, internal/cli/intervention_test.go, internal/cli/upgrade_test.go, docs/guide/commands.md
- **Notes for next iteration:** A record counts as v1 when its frontmatter has no schema: line. migrateText splits with SplitAfter so line endings and the body are preserved verbatim, and the inserted lines use the by: line's own ending. docs/guide/commands.md regenerated with go generate ./...; upgrade needed no change since it never touched records.

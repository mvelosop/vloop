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

## T5 — Add vloop intervention show <id>, resolving the ids written in its Context

- **Outcome:** done (review: PASS)
- **Summary:** vloop intervention show <id> prints the record (id, summary, frontmatter, sections) then a links block resolving task, run, defect, intervention and commit ids named in Context, in order of first appearance; --json adds links [{ref, kind, title}]; an unknown id exits 1.
- **Files:** internal/intervention/show.go, internal/cli/intervention.go, internal/cli/intervention_test.go, docs/guide/commands.md
- **Notes for next iteration:** Only Context is scanned, word by word ([A-Za-z0-9-]+); hex words of 7-40 chars count as commits, so a plain word like 'defaced' prints (not found). Task titles come from the latest '[vloop] plan <run id>' commit via runs.PlanAt, run id being the record's brief minus .loop-brief. Unresolved links have an empty title in JSON. docs/guide/commands.md regenerated with go generate. TestTaskGate* in internal/cli fail only inside a loop session (plan changed), not from this change.

## T6 — Report agreement across briefs: vloop metrics --interventions, by kind and by phase, and across a workspace

- **Outcome:** done (review: PASS)
- **Summary:** vloop metrics --interventions prints the agreement table by kind (default) or phase with a total row, --json rows, and a repo-first table across --workspace; --by task is refused with exit 2.
- **Files:** internal/metrics/interventions.go, internal/cli/metrics.go, internal/cli/workspace.go, internal/cli/metrics_interventions_test.go, docs/guide/commands.md
- **Notes for next iteration:** Counts are recomputed from intervention.List each run. Share is recommended/(n-no-options) rounded half up, n/a when the divisor is 0. Every kind or phase gets a row, zeros included; the workspace table prints every kind per repo then one total row. JSON rows use snake_case keys and repo only in the workspace form. docs/guide/commands.md regenerated with go generate. TestTaskGate* in internal/cli fail only inside a loop session (plan changed), not from this change.

## T7 — Add the interventions line and object to per-brief metrics and the close snapshot, and agreement and options to the export

- **Outcome:** done (review: PASS)
- **Summary:** Per-brief metrics gain an interventions object (total, by_agreement) and a summary line after defects, which the close snapshot inherits; the export's intervention lines gain agreement and options (the count), never any text. metrics/v2 and export/v1 declare the keys as optional.
- **Files:** internal/metrics/metrics.go, internal/metrics/print.go, internal/cli/export.go, schemas/metrics.v2.json, schemas/export.v1.json, internal/cli/export_test.go, internal/cli/metrics_test.go, cmd/vloop/b3_e2e_test.go, cmd/vloop/b4_e2e_test.go
- **Notes for next iteration:** by_agreement keys are hyphenated (other-option, no-options), matching the agreement values. Interventions are counted from intervention.List(root, brief), so series-level records (brief empty) are excluded. TestTaskGate* in internal/cli fail only inside a loop session (plan changed during this session), not from this change.

## T8 — Teach the operator skills to propose options and record them, and add the operate-proposes-options eval case

- **Outcome:** done (review: PASS)
- **Summary:** Both operator skills now say to propose up to three real options, recommend one with a reason, and record with the intervention add flags after the decision; the operate-proposes-options eval case exists and is listed in the evals README.
- **Files:** plugin/skills/operate/SKILL.md, .claude/skills/vloop-operator/SKILL.md, plugin/evals/README.md, plugin/evals/operate-proposes-options/case.yaml, plugin/evals/operate-proposes-options/prompt.md, plugin/evals/operate-proposes-options/scaffold.sh, plugin/evals/operate-proposes-options/task.json, plugin/evals/operate-proposes-options/graders/options-with-recommendation.md, plugin/evals/operate-proposes-options/graders/gate-not-amended.md, plugin/evals/operate-proposes-options/graders/no-straw-option.md, plugin/evals/operate-proposes-options/graders/gate-file-untouched.md
- **Notes for next iteration:** The case reuses the work-disputes case's allowed_tools (same as the guide's --allow-tools), so TestEvalGrantsCoverFence needed no guide change. Its scaffold marks T1 blocked via jq after write_state. The skill text in plugin/skills/operate/SKILL.md is under a new subsection of step 4.

## T9 — Add the Agreement column to the interventions index, and document v2 in the guides and the domain

- **Outcome:** gate_failed (review: skipped)
- **Summary:** The interventions index script gains an Agreement column and validates by and agreement; defects.md, metrics.md, the interventions README, domain-model.md and measurement-context.md document intervention/v2 and its commands.
- **Files:** tools/interventions-index.sh, docs/guide/defects.md, docs/guide/metrics.md, .vloop/interventions/README-interventions.md, docs/domain/domain-model.md, docs/domain/measurement/measurement-context.md, internal/cli/guide_test.go, internal/cli/interventions_index_test.go
- **Notes for next iteration:** The committed index block in .vloop/interventions/README-interventions.md still has the old four columns, so tools/interventions-index.sh --check fails until T10 runs --write (after migrating records). TestGuideDefectsCoversInterventions now checks both intervention/v1 and intervention/v2 words; the index tests are internal/cli/interventions_index_test.go and skip without bash.

## T9 — Add the Agreement column to the interventions index, and document v2 in the guides and the domain

- **Outcome:** gate_failed (review: skipped)
- **Summary:** The interventions index gains an Agreement column and validates by and agreement; the guides, the interventions README and the domain docs document intervention/v2.
- **Files:** tools/interventions-index.sh, docs/guide/defects.md, docs/guide/metrics.md, .vloop/interventions/README-interventions.md, docs/domain/domain-model.md, docs/domain/measurement/measurement-context.md, internal/cli/guide_test.go, internal/cli/interventions_index_test.go
- **Notes for next iteration:** The previous attempt failed the gate only because README-interventions.md had no v2 text; added the v2 frontmatter rows and sections there. The committed index block still has the old four columns, so tools/interventions-index.sh --check fails until T10 runs --write after migrating records.

## T9 — Add the Agreement column to the interventions index, and document v2 in the guides and the domain

- **Outcome:** gate_failed (review: skipped)
- **Summary:** The interventions README now documents the intervention/v2 frontmatter and the Context, Options, Recommended, Suggested and Decided sections, the last piece the gate lacked; the index script, guides and domain docs were already committed.
- **Files:** .vloop/interventions/README-interventions.md
- **Notes for next iteration:** Earlier attempts' README edits were absent from the tree (the file was unchanged at HEAD), so this attempt re-added them; check the file is still modified after the driver's pass. The committed index block still has the old four columns until T10 runs tools/interventions-index.sh --write.

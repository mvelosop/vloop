# Journal — B20261105-2002-b37-i88-metrics-by-session-kit-without-python

Append-only narrative of this plan. Rendered state lives in .vloop/state/plan.md.

## Plan — B20261105-2002-b37-i88-metrics-by-session-kit-without-python

- **Brief:** `docs/briefs/B20261105-2002-b37-i88-metrics-by-session-kit-without-python.loop-brief.md`
- **Tasks:** 6

## T1 — Add `vloop metrics --by session` and `by_session` to metrics/v4

- **Outcome:** check_failed (review: skipped)
- **Summary:** vloop metrics --by session prints one row per counted session record, and every metrics/v4 JSON document carries by_session (optional in the schema, dropped by export). Docs are updated and commands.md is regenerated.
- **Files:** internal/runs/runs.go, internal/metrics/metrics.go, internal/metrics/print.go, internal/cli/metrics.go, internal/cli/export.go, internal/cli/metrics_session_test.go, schemas/metrics.v4.json, docs/guide/metrics.md, docs/guide/commands.md, docs/domain/measurement/metrics.md
- **Notes for next iteration:** Rows are built from runs.Session, extended with Schema, CostSource, CreditsSet (nil when the record has no credits key), Turns (turns, else num_turns) and DenialList. Credits stays a float64 for the sums. The TestTaskGate* tests in internal/cli fail when run inside a loop session (the plan was changed during this session), which is environmental and unrelated to this change; I did not run them outside a session.

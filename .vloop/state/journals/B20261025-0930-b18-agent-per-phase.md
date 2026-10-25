# Journal — B20261025-0930-b18-agent-per-phase

Append-only narrative of this plan. Rendered state lives in .vloop/state/plan.md.

## Plan — B20261025-0930-b18-agent-per-phase

- **Brief:** `docs/briefs/B20261025-0930-b18-agent-per-phase.loop-brief.md`
- **Tasks:** 4

## T1 — Add the five agent.<phase> config keys and resolve them as model and effort are

- **Outcome:** check_failed (review: skipped)
- **Summary:** Added agent.plan/work/review/gate-review/verify config keys (default claude, only valid value claude, VLOOP_AGENT_<PHASE>, stored under [agent]); ResolveRun resolves the agent for work, review and verify once, ResolveAgent resolves it per session for plan and gate-review. Docs rows added.
- **Files:** internal/config/config.go, internal/config/config_test.go, internal/driver/resolve.go, internal/driver/resolve_test.go, docs/guide/configuration.md
- **Notes for next iteration:** Resolved.Agent(phase) and driver.ResolveAgent(root, kind) are what T2 consumes; nothing a session receives changes yet. TestDefaultsEveryRow (existing) gained the five agent.* rows because the key list is the contract item that changed.

## T1 — Add the five agent.<phase> config keys and resolve them as model and effort are

- **Outcome:** done (review: PASS)
- **Summary:** Five agent.<phase> config keys (default claude, only valid value claude, VLOOP_AGENT_<PHASE>, stored under [agent]) with ResolveRun/ResolveAgent resolution and docs rows. This attempt fixed the three existing tests that pinned the old config list.
- **Files:** internal/config/config.go, internal/config/config_test.go, internal/driver/resolve.go, internal/driver/resolve_test.go, docs/guide/configuration.md, cmd/vloop/e2e_test.go, internal/cli/config_test.go
- **Notes for next iteration:** Previous attempt failed the go check because three existing tests listed config keys without agent.*: TestWorkedExampleEnglishSession (cmd/vloop/e2e_test.go), TestConfigListDefaults and TestConfigListJSONOrderAndNull (internal/cli/config_test.go). They changed only where the brief's contract adds the five agent keys to config list. Resolved.Agent(phase) and driver.ResolveAgent(root, kind) are what T2 consumes.

## T2 — Start every session through a seam with a Claude adapter, and write session/v3 records

- **Outcome:** done (review: PASS)
- **Summary:** Sessions now start through an Adapter seam (internal/driver/agent.go) with a Claude adapter owning preflight, supported phases, preparation, command/environment, and stdout-to-record mapping; records are session/v3 (agent, cost_source) with schemas/session.v3.json, and session.md is reworded.
- **Files:** internal/driver/agent.go, internal/driver/session.go, internal/driver/iterate.go, internal/driver/plan.go, internal/driver/gatereview.go, internal/cli/doctor.go, schemas/session.v3.json, docs/domain/execution/session.md, cmd/vloop/run_harness_test.go, cmd/vloop/run_agent_test.go, internal/driver/session_test.go, cmd/vloop/b2_e2e_test.go, cmd/vloop/b6_e2e_test.go, cmd/vloop/run_plan_test.go, cmd/vloop/run_safety_test.go, cmd/vloop/run_verify_test.go, internal/cli/schema_test.go, internal/schema/schema_test.go, embed_test.go
- **Notes for next iteration:** Spec.Agent selects the adapter (empty = claude); work/review/verify take it from Resolved.Agent, plan/gate-review from ResolveAgent per session. Runner.Claude is now the claude adapter's command. doctorClaude calls driver.AgentPreflight. Existing tests changed only where they pin session/v1 or session/v2 in a record or the schema list/count (contract item: records are session/v3): TestSessionFieldMapping, run_plan_test, run_safety_test, run_verify_test, b6_e2e_test, b2_e2e_test, schema_test, embed_test. The harness stub gained invocations.log and STUB_NOCOST. argv tests in session_test.go unchanged.

## T3 — Report agents per phase in metrics/v3, reading v1, v2 and v3 records and metrics/v2 snapshots

- **Outcome:** done (review: PASS)
- **Summary:** vloop metrics --json, workspace JSON and new snapshots are metrics/v3: metrics/v2 plus agents per phase (claude for session/v1 and v2 records, the agent field for session/v3). schemas/metrics.v3.json added; docs and tests updated.
- **Files:** schemas/metrics.v3.json, internal/metrics/metrics.go, internal/metrics/time.go, internal/metrics/time_test.go, internal/runs/runs.go, internal/schema/schema_test.go, internal/schema/testdata/metrics3.valid.json, internal/schema/testdata/metrics3.invalid.json, internal/closing/closing_test.go, internal/cli/close_test.go, internal/cli/metrics_test.go, internal/cli/schema_test.go, cmd/vloop/b2_e2e_test.go, cmd/vloop/b4_e2e_test.go, cmd/vloop/b14_e2e_test.go, cmd/vloop/run_snapshot_test.go, embed_test.go, docs/domain/measurement/metrics.md, docs/guide/metrics.md
- **Notes for next iteration:** Existing tests changed only where they pin metrics/v2 as the --json/snapshot schema or the schema count/list (contract item: --json is metrics/v3): b2/b4/b14 e2e, run_snapshot, cli close/metrics/schema tests, embed_test, schema_test, closing_test (fixtureReport now reads metrics3.valid.json). Report.Agents mirrors Models (verify only when the phase ran); by_task has no agents. In go test ./..., TestTaskGate* in internal/cli fail here with 'the plan was changed during this session' - they fail inside the running loop session, not from this change; nothing else fails.

## T4 — Close the run: every gate at once and the worked example line for line

- **Outcome:** done (review: PASS)
- **Summary:** Reconciled docs that still named session/v1 or metrics/v2 as what vloop writes; the finished tree passes all gates together.
- **Files:** docs/domain/measurement/metrics.md, docs/domain/bounded-contexts.md, docs/domain/platform/platform-context.md, docs/domain/execution/run.md, docs/domain/execution/session.md, docs/guide/metrics.md, docs/guide/concepts.md
- **Notes for next iteration:** Docs only, no code change. Remaining session/v1, session/v2 and metrics/v2 mentions describe older records and snapshots that still read. configuration.md already listed the five agent keys.

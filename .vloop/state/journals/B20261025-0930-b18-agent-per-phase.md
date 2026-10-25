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

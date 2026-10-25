# Journal — B20261025-1307-b19-copilot-review

Append-only narrative of this plan. Rendered state lives in .vloop/state/plan.md.

## Plan — B20261025-1307-b19-copilot-review

- **Brief:** `docs/briefs/B20261025-1307-b19-copilot-review.loop-brief.md`
- **Tasks:** 5

## T1 — Accept copilot for agent.review only, add copilot.credit-usd, and give the run harness a stub copilot

- **Outcome:** check_failed (review: skipped)
- **Summary:** agent.review now accepts claude and copilot (the other four agent keys stay claude only); copilot.credit-usd (default 0.0066667, positive Num) exists; the run harness has a stub copilot with TestRunHarnessCopilot.
- **Files:** internal/config/config.go, internal/config/config_test.go, cmd/vloop/run_harness_test.go, cmd/vloop/e2e_test.go, internal/driver/resolve_test.go
- **Notes for next iteration:** Stub copilot is runCopilotStubSource in cmd/vloop/run_harness_test.go, installed as <stub>/copilot; it logs to copilot.log (not argv.log/invocations.log), reads copilot-script.sh (r.copilotScript) with STUB_NANOAIU/EXIT/DURATION/TURNS/SILENT; helper r.copilot. Existing tests changed only where they pin the contract's new behaviour: internal/driver/resolve_test.go used VLOOP_AGENT_REVIEW=copilot as an invalid value, now VLOOP_AGENT_WORK (copilot valid for agent.review); cmd/vloop/e2e_test.go config list expectation gains the copilot.credit-usd line (contract: new key). The full go test ./... was not completed by me in one run; the driver runs it.

## T1 — Accept copilot for agent.review only, add copilot.credit-usd, and give the run harness a stub copilot

- **Outcome:** done (review: PASS)
- **Summary:** Fixed the check failure from the first attempt: configuration.md documents agent.review's copilot value and the new copilot.credit-usd key, and the config list tests expect it.
- **Files:** docs/guide/configuration.md, internal/cli/config_test.go
- **Notes for next iteration:** Earlier work (config.go, stub copilot) is unchanged. config_test.go list expectations changed only for the new copilot.credit-usd key (contract item: new key). TestTaskGate* failures are an artifact of running inside a driver session, not a defect.

## T2 — Start review sessions on copilot through a Copilot adapter with its own home, the embedded plugin, the guidelines and C2's denials

- **Outcome:** done (review: PASS)
- **Summary:** A copilot adapter (internal/driver/agent_copilot.go) behind adapterFor starts review sessions with its own per-session COPILOT_HOME and guidelines directory under .vloop/tmp/, the embedded plugin and the C2 deny patterns; tests added.
- **Files:** internal/driver/agent.go, internal/driver/agent_copilot.go, internal/driver/agent_copilot_test.go
- **Notes for next iteration:** Prepared gained Home and Guidelines fields. Guidelines are copied to AGENTS.md in a fresh copilot-guidelines-* dir; COPILOT_HOME is a fresh copilot-home-* dir. Parse returns not-a-record on purpose (T3). Runner has no copilot command override: copilot is looked up on PATH.

## T3 — Map Copilot's JSON Lines onto session/v3 with credits and an estimated cost, and add credits to metrics/v3

- **Outcome:** done (review: PASS)
- **Summary:** The Copilot adapter's Parse reads the JSON Lines into session/v3 (credits, estimated cost, turns, duration, is_error; no record without a result event); session/v3 and metrics/v3 gain credits and the estimated source; metrics carries credits and marks an estimated cost with ≈.
- **Files:** internal/driver/agent.go, internal/driver/agent_copilot.go, internal/driver/agent_copilot_parse_test.go, internal/driver/session.go, internal/runs/runs.go, internal/metrics/time.go, internal/metrics/metrics.go, internal/metrics/print.go, internal/metrics/print_test.go, internal/schema/session_credits_test.go, schemas/session.v3.json, schemas/metrics.v3.json, docs/domain/execution/session.md
- **Notes for next iteration:** Runner.withSession gives the copilot adapter the session model and copilot.credit-usd (read via config.Get, so file and VLOOP_COPILOT_CREDIT_USD both apply) before Prepare; Parse stays free of config. The ceiling needs no change: spendIn sums cost_usd. The ≈ shows when metrics credits.total > 0. credits.verify is omitempty so a brief without verify keeps no 'verify' text (TestPrintSummaryVerify). session/v3 refuses cost_source estimated for agent claude through an allOf if/then. TestTaskGate* fail only inside a driver session (plan-hash guard), as in T1.

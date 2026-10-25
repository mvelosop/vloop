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

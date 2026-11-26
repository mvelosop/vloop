# Journal — B20261126-1003-b41-i161-doctor-gate-shell-replaced-plan

Append-only narrative of this plan. Rendered state lives in .vloop/state/plan.md.

## Plan — B20261126-1003-b41-i161-doctor-gate-shell-replaced-plan

- **Brief:** `docs/briefs/B20261126-1003-b41-i161-doctor-gate-shell-replaced-plan.loop-brief.md`
- **Tasks:** 3

## T1 — Make replaced plans use the configured gate shell

- **Outcome:** rejected (review: FAIL)
- **Summary:** Doctor now uses one replaced-plan decision for both plan and gate shell checks, selecting the configured shell for replaced plans and the recorded shell for plans still in use. Added cross-platform table coverage for every G1 status and preserved configured-shell warnings in run preflight.
- **Files:** internal/cli/doctor.go, internal/cli/doctor_plan_shell_test.go, internal/cli/run.go
- **Notes for next iteration:** doctorPlanReplaced is shared by the plan and gate shell rows; complete state/v1 plans are always replaced, while complete state/v2 plans depend on whether doctor or run names a different brief. gate shell warnings are now included in preflight output without changing their result or message.

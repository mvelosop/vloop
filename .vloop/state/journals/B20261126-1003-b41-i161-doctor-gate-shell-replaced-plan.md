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

## T1 — Make replaced plans use the configured gate shell

- **Outcome:** done (review: PASS)
- **Summary:** Doctor now preserves the legacy complete state/v1 plan message while sharing the replaced-plan decision with the gate shell row. The Windows-safe table tests pin both the configured-shell behavior and the unchanged plan messages.
- **Files:** internal/cli/doctor.go, internal/cli/doctor_plan_shell_test.go
- **Notes for next iteration:** Complete state/v1 plans continue to use the configured gate shell, but the plan row retains `a complete state/v1 plan — the next brief's plan replaces it`; complete state/v2 plans retain the brief-specific replaced-plan message. The doctor table test remains cross-platform and now asserts both row messages.

## T2 — Document the configured shell for replaced plans

- **Outcome:** done (review: PASS)
- **Summary:** The Gates and the gate shell section now explains that doctor checks the configured shell for a plan the next brief replaces.
- **Files:** docs/guide/concepts.md
- **Notes for next iteration:** No other documentation was changed.

## T3 — Verify the complete doctor shell replacement contract

- **Outcome:** done (review: PASS)
- **Summary:** The final tree passes both earlier gates together, including the built-CLI worked example for doctor text and JSON, configured-shell warnings, and replaced versus in-use plans during run preflight.
- **Files:** none
- **Notes for next iteration:** No product or documentation changes were needed in the close task; the T3 gate reused the unchanged T1 and T2 gates and the T1 gate exercised the complete worked example in a temporary repository.

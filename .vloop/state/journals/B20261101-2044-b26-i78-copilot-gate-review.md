# Journal — B20261101-2044-b26-i78-copilot-gate-review

Append-only narrative of this plan. Rendered state lives in .vloop/state/plan.md.

## Plan — B20261101-2044-b26-i78-copilot-gate-review

- **Brief:** `docs/briefs/B20261101-2044-b26-i78-copilot-gate-review.loop-brief.md`
- **Tasks:** 6

## T1 — Add the gate-review phase to the fence judge, vloop fence check and vloop fence hook

- **Outcome:** rejected (review: FAIL)
- **Summary:** The fence judge, vloop fence check and vloop fence hook now accept a gate-review phase that reads fence/gate-review.json and words messages and the skill hint for gate-review.
- **Files:** internal/fence/judge.go, internal/fence/judge_test.go, internal/cli/fence.go, internal/driver/agent_copilot.go, docs/guide/commands.md, docs/guide/configuration.md
- **Notes for next iteration:** Contract item G1: judge_test.go's unknown-phase test changed only to the new message. Added TestGateReviewPhase covering both worked tables and the apply_patch lines. docs/guide/commands.md regenerated with go generate. Fences untouched.

## T1 — Add the gate-review phase to the fence judge, vloop fence check and vloop fence hook

- **Outcome:** done (review: PASS)
- **Summary:** The fence judge, fence check and fence hook accept a gate-review phase reading fence/gate-review.json; the hook now has committed gate-review tests too.
- **Files:** internal/fence/judge.go, internal/fence/judge_test.go, internal/cli/fence.go, internal/cli/fence_hook_test.go, internal/driver/agent_copilot.go, docs/guide/commands.md, docs/guide/configuration.md
- **Notes for next iteration:** Previous attempt was rejected because no hook test covered gate-review; fence_hook_test.go now loops review, second-review and gate-review over the shared cases and adds TestFenceHookGateReviewDecisions with exact reasons. Contract item G1: judge_test.go's unknown-phase test changed only to the new message. Other internal/cli task_gate tests fail inside a work session only because the plan guard sees a changed plan; not related.

## T2 — Run gate-review sessions on Copilot: the adapter, its record and agent.gate-review = copilot

- **Outcome:** done (review: PASS)
- **Summary:** The Copilot adapter supports the gate-review phase (same deny list incl. shell(rm:*), hooks file with --phase gate-review), its record is session/v4, and agent.gate-review accepts copilot while agent.plan still does not.
- **Files:** internal/config/config.go, internal/config/config_test.go, internal/driver/agent_copilot.go, internal/driver/agent_copilot_test.go, internal/driver/session.go
- **Notes for next iteration:** Contract item G2: TestCopilotSupportsWorkReviewAndVerify now expects gate-review supported (plan still not) and TestAgentKeys treats agent.gate-review like work/review/verify (accepts copilot). Added TestCopilotGateReviewArgvAndHooks and gate-review cases in TestCopilotVerifyRecordIsSessionV4. gatereview.go untouched. Docs (config guide wording of agent.gate-review) left to T4; preflight/doctor to T3.

## T3 — Check agent.gate-review = copilot in the run preflight and doctor

- **Outcome:** done (review: PASS)
- **Summary:** agent.gate-review = copilot now gets the experimental warning, the default-model refusal, K1, the login refusal and doctor's copilot rows, whatever run.verify says; claude behaviour is unchanged.
- **Files:** internal/cli/doctor.go, internal/cli/run.go, internal/cli/doctor_copilot_gatereview_test.go
- **Notes for next iteration:** Added copilotGateReview (no run.verify switch) to copilotKeys; run.go's warning filter lists agent.gate-review. The review look-alike warning is unchanged and now shows with only the gate review on copilot. Internal/cli task_gate tests fail inside a work session only because the plan guard sees a changed plan; unrelated.

## T4 — Document Copilot gate-review sessions in the fence audit, the configuration guide and the domain pages

- **Outcome:** done (review: PASS)
- **Summary:** The fence audit gains a gate-review part classifying every gate-review fence rule and adapter deny pattern; the configuration guide, session page and glossary now cover Copilot gate-review sessions.
- **Files:** docs/design-notes/fence-audit.md, docs/guide/configuration.md, docs/domain/execution/session.md, docs/domain/domain-model.md
- **Notes for next iteration:** The audit says the driver after a gate-review session checks only git (moved ref, exit 9; changed .git/config or hooks) and does not revert files outside .vloop/tmp/, so file-write rules are classified hook. Code comments (adapter, hook, doctor) already named gate-review from T1-T3. The model.gate-review refusal is quoted from internal/cli/run.go.

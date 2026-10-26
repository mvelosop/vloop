# Journal — B20261026-1002-b22-fence-second-review

Append-only narrative of this plan. Rendered state lives in .vloop/state/plan.md.

## Plan — B20261026-1002-b22-fence-second-review

- **Brief:** `docs/briefs/B20261026-1002-b22-fence-second-review.loop-brief.md`
- **Tasks:** 9

## T1 — Add vloop fence check: the judge that decides a tool call against a phase's fence, looking through shell wrappers

- **Outcome:** done (review: PASS)
- **Summary:** Added internal/fence, a pure judge (splitter, wrapper unwrapping, fence and Copilot-stem matching, path rules), and `vloop fence check` on top of it.
- **Files:** internal/fence/split.go, internal/fence/judge.go, internal/fence/judge_test.go, internal/cli/fence.go, internal/cli/root.go, internal/driver/agent_copilot.go, docs/guide/commands.md
- **Notes for next iteration:** fence.New(phase, root, home) reads fence/review.json (the Bash deny bodies) and driver.CopilotDenyStems() (new export of the shell(...) stems in Fence matching form); T2 should call Judge.Decide(tool, path, command). For apply_patch the patch text goes in the command argument. Beyond the contract, the judge also denies (fail closed): subshell parens, shell keywords as a command word (if/for/while/{ ...), and sh/bash/... with no -c string (script or stdin). Under `go test ./...` a VLOOP_PLAN_SHA256 inherited from a loop session makes the task_gate tests fail; unset it to run them.

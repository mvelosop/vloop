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

## T2 — Add vloop fence hook: read one Copilot preToolUse payload and answer allow or deny as JSON, failing closed

- **Outcome:** done (review: PASS)
- **Summary:** Added `vloop fence hook --phase --root`: reads a Copilot preToolUse payload on stdin, judges it with T1's judge and prints one allow/deny JSON line; unreadable payloads, unknown phases (including work) and internal errors print the deny line and exit 2.
- **Files:** internal/cli/fence.go, internal/cli/fence_hook_test.go, docs/guide/commands.md
- **Notes for next iteration:** Web access is denied, so Copilot's docs could not be read: the toolArgs form is unconfirmed. The experiments note shows toolArgs.command as an object; the hook accepts an object or a JSON string holding one. apply_patch's patch is read from a raw string toolArgs, or the input/patch/command keys; any other shape yields no paths and is denied. A shell tool with no readable command is denied by the hook itself (the judge would allow an empty command). Skill name read from toolArgs skill/name (unconfirmed key; a missing one is denied). Cobra flag-parse errors (e.g. an unknown flag) exit 2 without the deny line. Under go test ./..., unset VLOOP_PLAN_SHA256 or the task_gate tests fail.

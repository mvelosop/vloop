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

## T3 — Run every Copilot session under the hook, parse its denials into permission_denials, and keep Copilot on vloop's skills (K1)

- **Outcome:** done (review: PASS)
- **Summary:** Each Copilot session's home now holds hooks/vloop-fence.json whose preToolUse entry runs this binary's `fence hook --phase <phase> --root <root>` in bash and powershell forms; the Copilot parser fills permission_denials from denied tool.execution_complete events; vloop run and doctor refuse (copilot reviews only) a .claude/skills or .github/skills directory named like a plugin skill.
- **Files:** internal/driver/agent_copilot.go, internal/driver/agent_copilot_test.go, internal/driver/agent_copilot_parse_test.go, internal/cli/doctor.go, internal/cli/run.go, internal/cli/doctor_copilot_test.go
- **Notes for next iteration:** Web access is denied, so Copilot's hooks-file format could not be read from its docs: the file shape {version:1, hooks:{preToolUse:[{type:command,bash,powershell,timeoutSec}]}} is from memory and unconfirmed; T8's real-CLI probe should confirm it. The hook runs with the phase passed to Prepare (review now; second-review once T5 adds it). TestCopilotEnvAndHome pinned an empty COPILOT_HOME; changed to expect only hooks/ (brief H2). Plugin skill names come from the new driver.PluginSkillNames(). The K1 run refusal prints the doctor row as `vloop: <path> has the name of ...` and skips that row in the generic preflight list. K1 is checked for agent.review only; T4 adds agent.second-review once the key exists. Under go test ./..., unset VLOOP_PLAN_SHA256.

## T4 — Add the second reviewer's keys, its doctor rows, and the reworded Copilot experimental warning

- **Outcome:** check_failed (review: skipped)
- **Summary:** Added run/agent/model/effort.second-review keys, doctor's second-review row and no-independence warning, the reworded copilot experimental warning for both reviewers, and run preflight refusals (default model, login, K1 skill clash) for a copilot second reviewer in use.
- **Files:** internal/config/config.go, internal/config/config_test.go, internal/cli/doctor.go, internal/cli/run.go, internal/cli/doctor_copilot_test.go, internal/cli/doctor_second_review_test.go, internal/cli/config_test.go, docs/guide/configuration.md
- **Notes for next iteration:** Doctor rows for the second review are named 'run.second-review:' (with colon) so the printed line matches the brief's text. Copilot checks (on PATH, version, login, skill clashes) run once when either reviewer is copilot; the experimental warning is per key. Only the key table rows and the Copilot paragraph's first sentence of configuration.md were touched (tests require every key documented); T7 owns the rest. Existing tests changed: config list/JSON/defaults tests (new keys) and copilotWarning const (reworded warning, T4 contract). Under go test ./..., unset VLOOP_PLAN_SHA256.

## T4 — Add the second reviewer's keys, its doctor rows, and the reworded Copilot experimental warning

- **Outcome:** done (review: PASS)
- **Summary:** Second-review keys, doctor rows, no-independence warning, reworded Copilot warning and preflight refusals exist; fixed the e2e config-list test that failed the go check.
- **Files:** internal/config/config.go, internal/cli/doctor.go, internal/cli/run.go, docs/guide/configuration.md, cmd/vloop/e2e_test.go
- **Notes for next iteration:** Prior attempt failed the go check because cmd/vloop/e2e_test.go pins `config list` output; updated it for the four new keys (T4 acceptance 1). Under go test ./..., unset VLOOP_PLAN_SHA256.

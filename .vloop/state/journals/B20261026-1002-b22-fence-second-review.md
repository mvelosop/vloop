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

## T5 — Run the optional second review session after the primary, in each mode, with its records and the end-of-run dissent list

- **Outcome:** done (review: PASS)
- **Summary:** The optional second review session runs after the primary in record, feedback, on-pass and both-pass, with session/v4 and iteration/v3 records, a kept second verdict, run-log lines and the end-of-run dissent list.
- **Files:** internal/driver/iterate.go, internal/driver/session.go, internal/driver/agent.go, internal/driver/agent_copilot.go, internal/driver/cost.go, internal/driver/resolve.go, schemas/session.v4.json, schemas/iteration.v3.json, cmd/vloop/run_second_review_test.go
- **Notes for next iteration:** reviewSession is now reviewRun (shared guards, phase-parametrised) plus a thin wrapper; the second review is secondReviewStep. The primary verdict is copied to reports right after the primary review, then removed from .vloop/tmp/ before the second session. session/v4 is written only for second-review sessions and iteration/v3 only when a second review ran; every other record stays v3/v2 so existing readers and tests are unchanged. Existing tests changed: wantIterations accepts iteration/v3 for iterations with a second review; schema list/count tests (schema_test.go, cli/schema_test.go, b2_e2e_test.go, embed_test.go) list the two new schemas (T5 acceptance 6). metrics/v4 is T6's. Under go test ./..., unset VLOOP_PLAN_SHA256.

## T6 — Report the second review in metrics/v4

- **Outcome:** done (review: PASS)
- **Summary:** Metrics reports are now metrics/v4: agents gains second-review and a second_review object (mode, sessions, dissents, dissent_tasks); schemas/metrics.v4.json is embedded and validates; v3 and earlier stay readable.
- **Files:** schemas/metrics.v4.json, internal/metrics/metrics.go, internal/metrics/time.go, internal/metrics/time_test.go, internal/runs/runs.go, internal/schema/testdata/metrics4.valid.json, internal/schema/testdata/metrics4.invalid.json, docs/guide/metrics.md
- **Notes for next iteration:** second_review.mode is run.second-review read from config when the report is built, not recorded per run, so a later config change shows in a rebuilt report (the committed snapshot is frozen). Dissents come from iterations.jsonl second_review.dissent; dissent_tasks is distinct tasks in first-dissent order. Second-review sessions are counted and their agents listed, but their time and cost are not added to existing totals (cost/time meanings unchanged). Existing tests changed to metrics/v4: schema lists/counts (embed_test, cli/schema_test, b2_e2e), the schema-name assertions in close/metrics/closing/b4/b14/run_snapshot tests, and closing fixtureReport now reads metrics4.valid.json. Under go test ./..., unset VLOOP_PLAN_SHA256.

## T7 — Write the fence audit and update the configuration guide, the domain pages and both operator playbooks

- **Outcome:** done (review: PASS)
- **Summary:** Wrote docs/design-notes/fence-audit.md (every review-fence deny rule and adapter stem classified hook/driver/neither); the configuration guide now describes the hook, fence check, fail-closed rules, K1 and each second-review mode and dissent; domain pages and both operator playbooks updated.
- **Files:** docs/design-notes/fence-audit.md, docs/guide/configuration.md, docs/domain/domain-model.md, docs/domain/execution/session.md, docs/domain/execution/task.md, docs/domain/measurement/metrics.md, plugin/skills/operate/SKILL.md, .claude/skills/vloop-operator/SKILL.md
- **Notes for next iteration:** plugin/skills/operate/SKILL.md is word-budgeted (internal/budget ceiling 1750): the dissent paragraph is kept short for that reason. The configuration guide's new second-review mode table must not use run.second-review as a first-column cell, because the gate's row matcher takes the first match. The audit's 'neither' entries are shell programs that write or read outside redirections (tee, cp, cat ~/.claude/...) which the judge does not read; the driver's guards cover the writes.

## T8 — Extend the Copilot acceptance kit with the hook probe, the hook's failure modes and a run with a copilot second reviewer

- **Outcome:** done (review: PASS)
- **Summary:** The Copilot acceptance kit gains hook.sh/hook.py (the review session under vloop fence hook asked for each worked-example deny form and a non-vloop skill, plus the hook missing, exit 1, not JSON and timing out) and second.sh (a feedback-mode run with a copilot second reviewer, reported by report.py second); the README documents each part.
- **Files:** tools/copilot-acceptance/hook.sh, tools/copilot-acceptance/hook.py, tools/copilot-acceptance/second.sh, tools/copilot-acceptance/report.py, tools/copilot-acceptance/README.md
- **Notes for next iteration:** Not run against real Copilot (no network). hook.py edits the hooks file vloop wrote in the logged review session's COPILOT_HOME (shape from internal/driver/agent_copilot.go, unconfirmed against Copilot docs) and parses tool.execution_start/complete events (data.toolCallId, data.arguments, error.code denied) as written in copilot-cli-experiments.md; the start event's field names are unverified. hook.sh needs run.sh first (reuses its argv log and shims).

## T9 — Close the run: every gate at once, and the worked example line for line

- **Outcome:** done (review: PASS)
- **Summary:** Closing check: all gates T1–T8 pass together on the final tree; no code changes were needed.
- **Files:** none
- **Notes for next iteration:** No gaps found; working tree unchanged by this task.

## Run ended — complete

- **Run:** `B20261026-1002-b22-fence-second-review` · 10 iteration(s) this run
- **Plan:** 9/9 done, 0 blocked

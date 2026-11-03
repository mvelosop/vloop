# Journal — B20261102-2301-b33-i81-agent-eval-runner

Append-only narrative of this plan. Rendered state lives in .vloop/state/plan.md.

## Plan — B20261102-2301-b33-i81-agent-eval-runner

- **Brief:** `docs/briefs/B20261102-2301-b33-i81-agent-eval-runner.loop-brief.md`
- **Tasks:** 6

## T1 — Add tools/agent-eval's flags, its case reader and --list

- **Outcome:** done (review: PASS)
- **Summary:** tools/agent-eval now parses its flags with the brief's refusals in order, reads cases (prompt.md and graders/*.md frontmatter by hand) and implements --list.
- **Files:** tools/agent-eval/main.go, tools/agent-eval/cases.go, tools/agent-eval/frontmatter.go, tools/agent-eval/main_test.go
- **Notes for next iteration:** Without --list, after refusals and reading cases the tool prints 'agent-eval: running cases is not built yet' and exits 1; T2 replaces that. --list skips the agent/model/out/runs/timeout refusals but still refuses 'no case matches'. Double-quoted values are kept as written inside the quotes (no unescaping). A prompt.md with no frontmatter is read as all body; graders must have closed frontmatter.

## T2 — Run a case on Copilot end to end and score its regex, tool_used and file_exists graders

- **Outcome:** blocked (review: skipped)
- **Summary:** Implemented the Copilot run end to end (scaffold, session through driver.Runner, transcript, trace, regex/tool_used/file_exists graders, report.json, table); the gate fails at one clause that no unrounded score can satisfy.
- **Files:** internal/driver/session.go, internal/driver/agent.go, internal/driver/agent_copilot.go, internal/driver/agent_copilot_test.go, tools/agent-eval/main.go, tools/agent-eval/runner.go, tools/agent-eval/trace.go, tools/agent-eval/graders.go, tools/agent-eval/run_test.go
- **Notes for next iteration:** Driver options, both defaulting to today's behaviour: Runner.Transcript (writes the session's stdout as printed to that file) and Runner.HookBinary (the binary a Copilot fence hook runs; copilotAdapter.hook, os.Executable() when empty); tested in internal/driver/agent_copilot_test.go. The tool: runner.go (execute, runOnce, table), trace.go (Copilot events to the trace, tool-name map), graders.go. Not built yet, deliberately honest: an llm grader makes the run an error 'judge failed: <grader>: the llm grader is not built yet' (T3 replaces it), and --agent claude cases error with 'claude sessions are not built yet' (T4 replaces it). Gate dispute: the gate compares fixture-graders' run score to awk 'BEGIN { print 7 / 11 }', which prints 0.636364 (OFMT %.6g, checked on this machine), with a 1e-9 tolerance; the report carries the unrounded 0.6363636363636364 as the brief's 'unrounded mean' requires, so the difference is 3.6e-7. Only a score rounded to six digits passes, which I did not write. The gate clauses after that one (run folder, argv, plugin dir, hook, table, total, --runs 2, plan hash, o3, o4, no result, timeout) were not reached by the gate; my Go tests cover those behaviours with their own fake copilot. The operator should compare with tolerance 1e-6, or against jq's own 7/11.

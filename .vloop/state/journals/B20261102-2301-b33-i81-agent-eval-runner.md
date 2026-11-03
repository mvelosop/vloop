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

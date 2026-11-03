# agent-eval

A repository-owned program, not a vloop subcommand. It replays the plugin's eval
cases (`plugin/evals/`) on Claude or Copilot with any model, so an operator can
tell whether a model qualifies for a phase on an agent. Run it from the
repository root:

```
go run ./tools/agent-eval --agent copilot --model gpt-6.1-terra --out <scratch>/agent-eval --runs 3
```

Each run scaffolds the case's `scaffold.sh` into a fresh `repo/`, starts one
session there the way `vloop run` starts the case's phase, then scores the
case's graders (`regex`, `tool_used`, `file_exists`, `llm`) on what the session
did. `docs/design-notes/copilot-cli-experiments.md` (section `gpt-6.1-terra` on
every phase) is the hand qualification this tool automates.

## Flags

- `--agent` (required): `claude` or `copilot`.
- `--model` (required): the model the session runs on.
- `--out` (required): the output directory; it must not exist or must be empty.
- `--cases`: the directory of cases, default `plugin/evals`.
- `--case`: a case name or glob; repeatable. Default: every case.
- `--runs`: runs per case, default 1.
- `--judge-model`: the Claude model that judges the `llm` graders, default `sonnet`.
- `--timeout`: per-session limit, a duration, default `20m`.
- `--vloop`: the vloop binary for the fence hook (Copilot only); default the one on `PATH`.
- `--list`: print the selected cases, their phase and grader count, and exit.

Exit codes: 0 done, 1 failed, 2 refused or usage.

## Layout under --out

```
report.json
<case>/run-<k>/
  repo/             the scaffolded repository the session worked in
  transcript.jsonl  the agent's raw output
  trace.jsonl       the same events for both agents: tool calls and text
  sessions/         the session record the driver wrote
```

`report.json` holds, per case, its status (`scored`, `skipped` or `error`), the
mean score and each run's grader verdicts, cost and wall time. A table of the
same goes to stdout.

## What differs from `claude plugin eval`

- One arm: the case runs once per run, with no ablation arm.
- The phase's fence replaces `allowed_tools` and `max_turns`. The session runs
  under the same fence `vloop run` gives that phase, so a case's `allowed_tools`
  and `max_turns` are not read.
- A Claude judge scores the `llm` graders, asked three times with the majority
  deciding (`--judge-model`). A judge that cannot be asked makes the run an
  error, not a score.
- Cases are skipped when they name no phase vloop starts a session for; they
  appear in the report as `skipped` and do not count.
- `--vloop` names the binary that serves as the fence hook for Copilot.

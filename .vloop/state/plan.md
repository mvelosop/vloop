# Plan — B20261102-2301-b33-i81-agent-eval-runner

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **0/6 done** · iteration 0

**Brief:** `docs/briefs/B20261102-2301-b33-i81-agent-eval-runner.loop-brief.md` · **Updated:** 2026-11-02T23:55:30Z

## Progress

- [ ] **T1** — Add tools/agent-eval's flags, its case reader and --list
- [ ] **T2** — Run a case on Copilot end to end and score its regex, tool_used and file_exists graders
- [ ] **T3** — Score the llm grader with a Claude judge voting three times
- [ ] **T4** — Stream a Claude session's transcript into the same trace
- [ ] **T5** — Document the tool, the qualifying rule and the design note's answer
- [ ] **T6** — Close: the worked example as a committed test, every gate at once

## Tasks

### T1 — Add tools/agent-eval's flags, its case reader and --list

`pending` · depends on: none

Issue #81 needs a repository tool, `go run ./tools/agent-eval`, that replays the plugin's eval cases on Claude or Copilot with any model. This task builds its front door: the flags and their refusals (A1), the reading of a cases directory (each case's prompt.md and graders/*.md frontmatter, read by hand with no new module dependency), and `--list`, which prints the selected cases and runs nothing. Every later task runs cases through what this one reads, so the frontmatter subset and the refusal order must be exactly the brief's.

**Acceptance**

- tools/agent-eval is package main, run as `go run ./tools/agent-eval` from the repository root, with the flags --agent, --model, --out, --cases (default plugin/evals), --case (a path.Match glob, repeatable, a case runs when any matches; every case when absent), --runs (default 1), --judge-model (default sonnet), --timeout (a Go duration, default 20m), --vloop (default the vloop on PATH) and --list.
- The cases are the directories of --cases whose name does not start with `_`, in name order.
- Before running anything, the first condition that holds, in this order, prints one line on stderr, nothing on stdout, and exits 2: --agent, --model, --out missing (in that order) → `agent-eval: --<flag> is required`; an agent other than claude or copilot → `agent-eval: unknown agent <a>`; --runs below 1 or not a number → `agent-eval: --runs must be 1 or more`; --timeout not a Go duration → `agent-eval: --timeout <v> is not a duration`; --out exists and is not empty → `agent-eval: <dir> is not empty` (the dir as given); no case matches → `agent-eval: no case matches`; --agent copilot and no vloop binary found (--vloop naming no file, or no vloop on PATH) → `agent-eval: no vloop binary for the fence hook`. A refusal writes nothing into an existing --out.
- --list needs no --agent, --model or --out; it prints a header CASE, PHASE, GRADERS and one line per selected case (its name; its phase, `-` when it has none; its number of graders/*.md), columns aligned with text/tabwriter, padding 2, and exits 0. The phase is the first word of prompt.md's body, `/vloop:<skill>`, when the skill is plan, work, review, gate-review, verify or review-brief; any other skill has none.
- `go run ./tools/agent-eval --list` over plugin/evals exits 0 with a line for every case and nothing on stderr; no file under plugin/evals is changed to make it so (a violation the review rules against).
- A case without prompt.md, or a grader file whose frontmatter cannot be read, prints `agent-eval: <case>/<file>: <reason>` on stderr and exits 1, with or without --list, before anything runs.
- Frontmatter is the text between the first two `---` lines, read by hand, no module added to go.mod: top-level `key: value`; a key with no value followed by lines indented two spaces, each `key: value`, read as a one-level map; values plain, single-quoted (`''` an escaped quote) or double-quoted, otherwise kept as written, backslashes included; a `[...]` value read as a string and unused. Anything else, including a frontmatter that is never closed, is a reading error.
- Committed tests in tools/agent-eval (run by the configured go check under `go test ./...`) cover each refusal and its order, --list over a fixture cases directory built in a temporary directory, and the frontmatter reader's forms and errors; a task that adds this behaviour without such tests fails review.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T1/gate.sh
```

</details>

### T2 — Run a case on Copilot end to end and score its regex, tool_used and file_exists graders

`pending` · depends on: T1

This is the core of the runner: one run of a case scaffolds a workspace, starts one session exactly as vloop starts that phase on that agent, keeps the transcript, normalises it into a trace with Claude's tool names, scores the case's graders and writes report.json and the table. It is built and judged on Copilot first, with every grader type except llm (the next task's). The driver may gain options for this (keeping the transcript, naming the fence hook's binary), but every vloop session's argv, record and output must stay exactly as they are: what is being qualified is how vloop really runs the agent.

**Acceptance**

- A case whose skill is plan, work, review, gate-review or verify starts as vloop run starts that phase, and review-brief as vloop brief review starts it, with the case's argument (the rest of prompt.md's first body line) as the session's, the --model given, no effort, killed at --timeout; any other skill skips the case with reason `no phase` (phase "" in report.json, `-` in the table, 0 runs), starting nothing.
- Run k of a case uses `<out>/<case>/run-<k>/repo/`, created empty, where the case's scaffold.sh runs as `bash <absolute path of the script>`; a non-zero exit makes the run an error with reason `scaffold exited <code>: <the scaffold's last stderr line>` and starts no session.
- The session's run folder is `<out>/<case>/run-<k>/` (its record under sessions/, its stderr file), outside repo/; the plugin is the tool's embedded copy, extracted into the workspace's .vloop/tmp/ as vloop extracts it under the version name `agent-eval`; the tool takes no lock and does not clear the workspace's .vloop/tmp/.
- A Copilot session's fence hook runs --vloop's binary (`'<vloop path>' fence hook --phase <phase> --root ...`); vloop's own Copilot sessions still run os.Executable()'s.
- A work, review or verify session gets VLOOP_PLAN_SHA256 computed as vloop run computes it (state.PlanDigest over the scaffold's .vloop/state/state.json and its gate files) when the scaffold left a plan, none otherwise; plan, gate-review and review-brief sessions get none.
- `<out>/<case>/run-<k>/transcript.jsonl` holds Copilot's JSON Lines exactly as printed; `trace.jsonl` holds one object per event in order: `{"kind":"text","text":...}` for an assistant.message with non-empty content, `{"kind":"tool","tool":...,"input":...,"output":...}` for a tool.execution_start paired by toolCallId with its tool.execution_complete — input the arguments as compact JSON without HTML escaping, output the result's content ("" when none), the tool name mapped by the brief's table, any other name unchanged.
- The trace text is each event on its own line: a text event's text; a tool event's `<tool> <input>` and, on the next line, its output.
- regex searches the trace text (target: trace) or a workspace file (target source: file, path) with a Go regular expression, match contains or not_contains, a missing file scoring 0 with reason `no <path>`; tool_used counts tool events whose tool equals `tool` and, when input_match is given, whose input matches it as a Go regular expression, scoring 1 when the count is at least min (default 1) and at most max (unbounded when absent); file_exists scores 1 when path exists in the workspace; any other type except llm scores 0 with reason `unsupported grader <type>` — never as passing. A run's score is the weighted mean of its graders' scores.
- A session that leaves no result is an error with reason `no session result`; one killed at --timeout is `timed out after <d>` as Go prints the duration (`2s`, `20m0s`), the timeout's reason winning when both hold.
- report.json is `{agent, model, judge_model, runs, cases: [{name, phase, status scored|skipped|error, reason, mean, runs: [{status, reason, score, graders: [{name, weight, score, reason}], cost_usd, credits, judge_cost_usd, wall_ms}]}]}`: a grader's name is its file name without .md; an errored run has score null and no graders; a case is error when any run errored, skipped when skipped, scored otherwise; mean is the unrounded mean of its scored runs' scores, null when none; credits is the session's Copilot credits.
- stdout is a header CASE PHASE RUNS MEAN COST NOTES, one line per case in name order (text/tabwriter, padding 2; RUNS the runs started, MEAN two decimals or `-`, COST the sessions' and judge's dollars with two decimals, NOTES a skipped or errored case's reason, else the names of graders below 1 in any run, each once, in name order, joined by ", "), then `<c> case(s) · <s> skipped · <e> error(s) · $<total>`. Exit 0 when no case is error, 1 when any is.
- The argv, record and output of every vloop session are unchanged, and TestSessionArgs (internal/driver/session_test.go) and TestRunClaudeInvocations (cmd/vloop/run_agent_test.go) pass unmodified — a change to either test file, or to any file under plugin/evals, is a violation the review rules against.
- Committed tests in tools/agent-eval cover a Copilot run end to end with a fake copilot and fake vloop scripts on a temporary PATH and fixture cases in a temporary directory (scored, scaffold error, skipped, no result, timeout), the trace read from testdata/observed/copilot-events.jsonl, each grader type and the table; any driver option added has a test in internal/driver; no test runs a real claude or copilot.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T2/gate.sh
```

</details>

### T3 — Score the llm grader with a Claude judge voting three times

`pending` · depends on: T2

Most of the plugin's graders are llm graders, so without them no case scores fully. The judge is always Claude, whatever agent is under test, so scores across agents compare; it is asked three times and the majority decides. A judge that cannot be asked is a run error, never a vote, so a broken judge can never pass or fail a model silently.

**Acceptance**

- An llm grader's focus is the trace's last text event (default), the trace text (`focus: trace`) or a workspace file (focus source: file, path); a missing file scores 0 with reason `no <path>` and calls no judge.
- The judge is `claude -p --model <--judge-model> --output-format json --strict-mcp-config --setting-sources project`, started in an empty temporary directory outside the workspace and --out, with no --plugin-dir, its prompt on stdin holding the grader's body and the focus and asking for PASS or FAIL on the last line; the answer is the JSON's `result`, its cost `total_cost_usd`.
- It is asked three times and the majority decides 1 or 0; an answer whose last line is neither PASS nor FAIL counts as FAIL; the three calls' cost is the run's judge_cost_usd, and COST in the table adds it to the session's.
- A judge call that does not run, exits non-zero or prints no JSON with a `result` is not a vote: the run is an error with reason `judge failed: <grader>: <cause>`, and the tool exits 1.
- The judge is Claude whatever --agent is; report.json's judge_model is the --judge-model given (sonnet by default).
- Committed tests in tools/agent-eval cover the judge with a fake claude on a temporary PATH: the majority, a non-answer, a failing judge, each focus and a missing focus file; no test runs a real claude.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T3/gate.sh
```

</details>

### T4 — Stream a Claude session's transcript into the same trace

`pending` · depends on: T3

A Claude session normally prints one JSON result, which says nothing of the tools it called, so the graders cannot be applied. Only this tool runs Claude with `--output-format stream-json --verbose`, whose last line carries the result the record is read from; the stream becomes the same normalised trace a Copilot session gives, so one case's graders score both agents alike. vloop's own Claude sessions keep `--output-format json`.

**Acceptance**

- A Claude session of the tool is started as vloop starts that phase (fence, plugin, adapter command and environment) with `--output-format stream-json --verbose` in place of `--output-format json`; no vloop session's argv, record or output changes, and TestSessionArgs and TestRunClaudeInvocations pass unmodified (a change to either test file is a violation the review rules against).
- transcript.jsonl is the stream as printed; the session's record is read from its last line (type result), so cost_usd, turns and duration_ms are the result's; a stream without that line is an error with reason `no session result`.
- trace.jsonl: each `type: assistant` line's message.content items — `text` a text event, `tool_use` a tool call paired by id with the `type: user` line's `tool_result` of the same tool_use_id, whose content is a string or a list of `{type: text, text}` items; thinking items and system, rate_limit_event and result lines are not events. The same session on Claude and on Copilot gives the same trace events.
- credits is null in report.json for a Claude run.
- Committed tests in tools/agent-eval cover a Claude run with a fake claude streaming on a temporary PATH and the trace read from testdata/observed/claude-stream.jsonl; any driver option added has a test in internal/driver; no test runs a real claude.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T4/gate.sh
```

</details>

### T5 — Document the tool, the qualifying rule and the design note's answer

`pending` · depends on: T4

The tool is only useful if an operator knows how to run it and what its scores mean: when a model qualifies for a phase on an agent. The README says what the tool is and how it differs from `claude plugin eval`; the evals guide gains the qualifying rule; the design note that deferred the runner says it now exists.

**Acceptance**

- tools/agent-eval/README.md says what the tool is, lists every flag (--agent, --model, --out, --cases, --case, --runs, --judge-model, --timeout, --vloop, --list), gives the layout under --out (report.json; &lt;case&gt;/run-&lt;k&gt;/ with repo/, transcript.jsonl, trace.jsonl and the session record), and what differs from `claude plugin eval`: one arm, the phase's fence instead of allowed_tools and max_turns, a Claude judge, skipped cases, --vloop.
- docs/guide/evals.md has a section `## Other agents` right after `## Running`'s content and right before `## vloop's suites`, giving the command `go run ./tools/agent-eval`, a link to tools/agent-eval/README.md, and the rule: a model qualifies for a phase on an agent when, at 3 runs, every case of that phase scores a mean no lower than the phase's default Claude model scores on the same tool with the same runs.
- docs/design-notes/multi-agent-backends.md reads `It is tools/agent-eval since B33.` right after the sentence ending `another agent needs it first.` and before `The fence audit is worth doing`; nothing else in it changes.
- Docs only: no Go file changes; the configured go check stays green.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T5/gate.sh
```

</details>

### T6 — Close: the worked example as a committed test, every gate at once

`pending` · depends on: T1, T2, T3, T4, T5

Each earlier task was judged on its own slice; the close shows the whole tool does what the brief's worked example says, line for line, with every earlier gate still passing together. It leaves the worked example behind as a committed test, so the configured go check keeps holding the tool to it after the run's gates are archived.

**Acceptance**

- A committed test in tools/agent-eval runs every line of the brief's worked example (o1, o2, o3, o4, --list over the fixture cases, --list over plugin/evals, --agent nope, a non-empty --out, o6) against fixture cases fixture-review (graders a, b, c, d as the brief gives them) and fixture-operate under tools/agent-eval/testdata, with fake copilot, claude and vloop scripts on a temporary PATH, asserting each line's exit code, report fields and messages as the brief states them.
- Every earlier task's gate passes in the same tree.
- No test runs a real claude or copilot, or the network; no file under plugin/evals and neither testdata/observed sample changes; the configured go check is green.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T6/gate.sh
```

</details>


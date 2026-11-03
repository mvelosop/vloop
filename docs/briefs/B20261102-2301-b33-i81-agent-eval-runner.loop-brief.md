---
name: B20261102-2301-b33-i81-agent-eval-runner.loop-brief
description: Add tools/agent-eval, which replays the plugin's eval cases on Claude or Copilot with any model, starting each session as vloop starts that phase and scoring it with the cases' own graders, so a model can be qualified per phase
kind: brief
status: consumed
created: 2026-11-02
issue: 81
seeds: Issue 81 under 69; B32's acceptance, which qualified gpt-6.1-terra by hand on whole runs; decided by the assistant under the operator's delegation of 2026-11-02 for the #81–#88 chain
depends-on: [B20261102-2101-b32-i84-gates-through-the-named-surface.loop-brief]
---
# Brief — vloop B33: an eval runner for any agent and model

- **Issue:** #81 (under #69). The PR closes it.
- **Starting point:** extends `main` at `24c0533`, plus this brief's ready
  merge. The planner pins that SHA as the **base** every gate compares
  against — never `HEAD~n`.
- **Produced by:** the operator's delegation of 2026-11-02 for a chain of
  briefs on #84, #81, #82, #83, #96–#97 and #88. Written with
  `plugin/skills/create-brief/`; every decision below was taken by the
  assistant under that delegation.
- **Review:** `vloop brief review` (Claude Opus). Round 1: FIX, 9 findings, all fixed (the fence hook's binary, `--cases` for fixtures, literal messages and the table, a run's status, the driver seam, the observed event shapes, review-brief's command, the design note's paragraph, the S invariants). Rounds 2 and 3 (the
  third reviewed the unchanged brief: an edit had failed): FIX, 6 and 9
  findings, all fixed (`--list` and a shape gated through the command; the
  judge's call, directory, stdin and fields; `input_match` a regular
  expression; the literals; the plan hash and effort; the argv line's
  tests; the design note's sentence; frontmatter read by hand; the
  acceptance's environment; the observed samples committed). Round 4: FIX,
  5 findings, all fixed (the plan hash is `state.PlanDigest` and only for
  work, review and verify; the scaffold's absolute path; no lock or clear;
  where the guide's section goes; `--list` in the summary). Round 5: FIX,
  4 findings, all fixed (a failing judge is a run error, the timeout's
  reason wins, the table's placeholders, no HTML escaping in `input`).
  Round 6: FIX, 1 finding, fixed (the run folder outside repo/, the
  extraction's version name); not re-reviewed, the fix only places two
  paths. No `ASK` in any round.

## Executive summary

The plugin's eval cases (`plugin/evals/`) hold the skills to their
behaviour, but they run only through `claude plugin eval`, so only on
Claude. Qualifying another agent's model for a phase, as B32 did for
`gpt-6.1-terra`, takes whole runs of a real brief by hand: slow, costly,
and unable to say which skill failed where.

**The goal:** one command replays the eval cases of the phases vloop
starts, on Claude or Copilot with any model; it starts each session as
vloop starts that phase on that agent, scores it with the case's own
graders, and writes a report; the guide says how a model qualifies for a
phase.

- **A1 — the command.** `go run ./tools/agent-eval --agent <a> --model <m>
  --out <dir>`, with `--cases`, `--case`, `--runs`, `--judge-model`,
  `--timeout`, `--vloop` and `--list`; a repository tool, like
  `tools/release`.
- **A2 — one run of a case.** A scaffolded workspace, one session started
  as vloop starts the case's phase, the transcript kept. `internal/driver`
  gains the options this needs, each defaulting to today's behaviour.
- **A3 — the trace.** Both agents' transcripts become one normalised trace
  with Claude's tool names, so the cases' graders apply unchanged.
- **A4 — the graders.** `regex`, `tool_used`, `file_exists` and `llm`, as
  the cases write them; the `llm` judge is always Claude.
- **A5 — the report.** report.json, a table on stdout, exit codes.
- **A6 — the docs.** The tool's README, a section in
  `docs/guide/evals.md` with the qualifying rule, a sentence in the design
  note.
- Unchanged: every `vloop` session's argv, record and output; the cases and
  their graders; `claude plugin eval`'s use.

## What it is

A repository tool that turns the eval suite into a qualification bench for
any agent vloop can run, with the cases and graders the Claude suite
already has. It is what `docs/design-notes/multi-agent-backends.md` calls
"a backend-agnostic eval runner replaying the same cases and graders", and
what moves a phase on an agent from experimental to supported.

## Why this shape, and what was rejected

Decided by the assistant on 2026-11-02 under the operator's delegation.
**Do not re-litigate these** — the planner inherits them.

- **A repository tool, not a vloop subcommand.** The cases live only here:
  `vloop plugin path` never extracts `plugin/evals/`, so a user's
  repository has none to run.
  *Rejected:* `vloop eval` — it would ship a command with nothing to run.
- **Sessions start as vloop starts them**: the phase's fence, the plugin,
  the adapter's command and environment. That is what is being qualified.
  *Rejected:* honouring a case's `allowed_tools` and `max_turns` — they are
  `claude plugin eval`'s sandbox grant, and Copilot has no equivalent; the
  phase's fence is the real grant.
- **The fence hook runs an installed `vloop`.** A Copilot session's hook
  runs a binary's `fence hook`; under `go run` the tool's own executable has
  none, so the tool names a `vloop` binary for the hook (`--vloop`, default
  the one on `PATH`).
  *Rejected:* building `fence hook` into the tool — two copies of the
  judge to keep in step.
- **Claude streams its transcript only in this tool.** The tool needs the
  whole transcript; vloop's own sessions need only the result.
  *Rejected:* streaming in every session — it changes every Claude
  session's argv for nothing a run reads.
- **The judge is always Claude.** Scores across agents compare only if the
  same judge reads them. *Rejected:* the agent under test judging itself.
- **One arm.** No without-plugin baseline: the question is how the agent
  does with the skills; `claude plugin eval` already measures what the
  skills add.
- **Cases of skills vloop does not start are skipped**, not failed:
  `/vloop:operate` and `/vloop:create-brief` run in the operator's
  interactive session.

## Binding references

- `plugin/evals/README.md` — the case layout: `prompt.md`, `case.yaml`,
  `graders/*.md`, `scaffold.sh`, and the `_` files that are not cases.
- `docs/guide/evals.md` — the grader types and their keys, and the `llm`
  grader's `focus`.
- `internal/driver/session.go` — `Runner`, `Spec`, `Prompt` and `Args`: how
  a session of a phase is started and recorded; `Run` parses stdout and
  keeps only the record.
- `internal/driver/agent.go` — the `Adapter` interface and the agents.
- `internal/driver/agent_copilot.go` — Copilot's command, its JSON Lines
  parsing, and `writeCopilotHooks`, whose hook runs `os.Executable()`'s
  `fence hook`.
- `internal/fence/judge.go` — the Copilot tool names the fence knows, which
  A3's mapping covers.
- `internal/driver/briefreview.go` — how `vloop brief review` starts a
  review-brief session.
- `tools/release/main.go` — the shape of a repository tool: `run(args, dir,
  stdout, stderr) int`, exit codes 0, 1 and 2.
- `docs/domain/domain-model.md` — S-1 (a session is a fresh process), S-2
  (sessions never commit or move refs) and S-4 (where a session writes):
  the tool's sessions are vloop sessions and keep them.
- `docs/design-notes/multi-agent-backends.md` — "The two hard parts", which
  this brief answers, and the paragraph that defers the runner.
- `docs/design-notes/copilot-cli-experiments.md` — section "`gpt-6.1-terra`
  on every phase": the hand qualification this tool replaces.

## Behaviour contract

### A1 — the command

`go run ./tools/agent-eval [flags]`, from the repository root.

| Flag | Meaning | Default |
| --- | --- | --- |
| `--agent <a>` | `claude` or `copilot` | required |
| `--model <m>` | the model id passed to the agent | required |
| `--out <dir>` | where runs and the report go; must not exist or be empty | required |
| `--cases <dir>` | the directory of cases | `plugin/evals` |
| `--case <glob>` | a case name pattern (`path.Match`); repeatable, a case runs when any matches | every case |
| `--runs <n>` | runs per case, 1 or more | `1` |
| `--judge-model <m>` | the Claude model of the `llm` judge | `sonnet` |
| `--timeout <d>` | one session's limit, a Go duration | `20m` |
| `--vloop <path>` | the `vloop` binary a Copilot session's fence hook runs | `vloop` on `PATH` |
| `--list` | print the selected cases and exit, running nothing; `--agent`, `--model` and `--out` are then not required | off |

The cases are the directories of `--cases` whose name does not start with
`_`, in name order. Before running anything, the tool refuses with one line
on stderr and exit 2, for the first condition that holds in the table's
order (the required flags in the order `--agent`, `--model`, `--out`):

| Condition | stderr |
| --- | --- |
| `--agent`, `--model` or `--out` missing | `agent-eval: --<flag> is required` |
| an agent other than `claude` or `copilot` | `agent-eval: unknown agent <a>` |
| `--runs` below 1 or not a number | `agent-eval: --runs must be 1 or more` |
| `--timeout` not a Go duration | `agent-eval: --timeout <v> is not a duration` |
| `--out` exists and is not empty | `agent-eval: <dir> is not empty` |
| no case matches | `agent-eval: no case matches` |
| `--agent copilot` and no `vloop` binary found | `agent-eval: no vloop binary for the fence hook` |

`--list` prints a header and one line per selected case, aligned as A5's
table is: `CASE`, `PHASE`, `GRADERS` — the case's name, its phase (`-`
when it has none) and its number of graders; exit 0. A grader file whose
frontmatter cannot be read, or a case without `prompt.md`, prints
`agent-eval: <case>/<file>: <reason>` on stderr and exits 1, with or
without `--list`, before anything runs.

**Frontmatter** (`prompt.md`'s and each grader's) is the part between the
first two `---` lines, read by hand — no module dependency is added. The
subset the cases use: `key: value` at the top level; a key with no value
followed by lines indented two spaces, each `key: value`, a one-level map
(`target:` and `focus:` with `source` and `path`); a value plain, or in
single quotes (`''` an escaped quote) or double quotes, kept as written
otherwise (backslashes included); a value in `[...]` (only `prompt.md`'s
`allowed_tools`) read as a string and not used. Anything else is a
reading error.

### A2 — one run of a case

- The case's skill is the first word of `prompt.md`'s body, `/vloop:<skill>`,
  and the rest of that line is its argument. `plan`, `work`, `review`,
  `gate-review` and `verify` start as `vloop run` starts that phase;
  `review-brief` as `vloop brief review` starts it. Any other skill skips
  the case with the reason `no phase`; its phase is `""` in report.json
  and `-` in the tables.
- The workspace is `<out>/<case>/run-<k>/repo/`, `k` from 1: created empty;
  the case's `scaffold.sh` runs in it as `bash <absolute path of the
  script>` (the scaffolds find `_lib.sh` through their own path); a
  non-zero exit makes the
  run an error with the reason `scaffold exited <code>: <the scaffold's
  last stderr line>`.
- One session runs in the workspace, started as vloop starts that phase on
  that agent (the same fence, plugin, adapter command and environment), with
  the model given and the case's argument as the session's, killed at
  `--timeout`. A Copilot session's fence hook runs `--vloop`'s binary.
  Unlike `vloop run`, no effort is set. A work, review or verify session
  gets the plan hash (`VLOOP_PLAN_SHA256`) as `vloop run` computes it
  (`state.PlanDigest` over the scaffold's `.vloop/state/state.json` and
  its gate files) when the scaffold left a plan, none otherwise; a plan,
  gate-review or review-brief session gets none, as in vloop. The tool
  takes no lock and does not clear the workspace's `.vloop/tmp/`: what the
  scaffold staged there is the case's. The session's run folder (its
  record under `sessions/`, its stderr file) is `<out>/<case>/run-<k>/`,
  outside repo/; the plugin and fence are the tool's own embedded copies,
  extracted into the workspace's `.vloop/tmp/` as vloop extracts them,
  under the version name `agent-eval`.
- Its full transcript is kept as `<out>/<case>/run-<k>/transcript.jsonl`:
  Copilot's JSON Lines as printed; for Claude, `--output-format stream-json
  --verbose` in place of `--output-format json`, a form used only by this
  tool, whose last line (type `result`) carries the fields the JSON form
  does and is what the session's record is read from.
- A session that leaves no result is an error with the reason `no session
  result`; one killed at the timeout, `timed out after <d>`, `<d>` as Go
  prints a duration (`20m0s`) — the timeout's reason wins when both hold.
- `internal/driver` may gain what this needs (keeping a transcript,
  streaming Claude's, naming the hook's binary), each option defaulting to
  today's behaviour: every `vloop` session's argv, record and output stay
  as they are, and the existing driver and session tests pass unchanged
  (among them `TestSessionArgs` in `internal/driver/session_test.go` and
  `TestRunClaudeInvocations` in `cmd/vloop/run_agent_test.go`, which pin
  Claude's `--output-format json`).

### A3 — the trace

`<out>/<case>/run-<k>/trace.jsonl` holds one JSON object per event, in
order:

- `{"kind":"text","text":"..."}` — an assistant message's text.
- `{"kind":"tool","tool":"<name>","input":"<input>","output":"<output>"}` —
  one tool call: `input` is the call's input as compact JSON without HTML
  escaping (`<`, `>` and `&` as written), `output` the
  result's text (empty when there is none).

The shapes read, as observed on 2026-11-02 (Claude Code's `stream-json`;
Copilot CLI 1.0.105's events, which its `--output-format json` stdout
carries, ending in a `result` line). Two samples with exactly these shapes
are committed with this brief: `tools/agent-eval/testdata/observed/claude-stream.jsonl`
(a real session, its paths made relative) and
`tools/agent-eval/testdata/observed/copilot-events.jsonl` (real events' structure,
their text shortened); the tool's tests read both, and neither changes.

- **Claude**: lines of `"type":"assistant"` hold `message.content[]` items;
  an item `{"type":"text","text":…}` is a text event, an item
  `{"type":"tool_use","id":…,"name":…,"input":{…}}` a tool call. Lines of
  `"type":"user"` hold `{"type":"tool_result","tool_use_id":…,"content":…}`
  items, `content` a string or a list of `{"type":"text","text":…}`. A
  `tool_use` pairs with the `tool_result` of the same id. Other item types
  (`thinking`) and line types (`system`, `rate_limit_event`, `result`) are
  not events.
- **Copilot**: `{"type":"assistant.message","data":{"content":…}}` is a text
  event when `content` is not empty;
  `{"type":"tool.execution_start","data":{"toolCallId":…,"toolName":…,"arguments":{…}}}`
  a tool call, paired with
  `{"type":"tool.execution_complete","data":{"toolCallId":…,"result":{"content":…}}}`
  of the same `toolCallId`. Other types are not events.

Copilot's tool names become Claude's:

| Copilot | Trace |
| --- | --- |
| `bash`, `powershell` | `Bash` |
| `create` | `Write` |
| `edit`, `apply_patch` | `Edit` |
| `view`, `read_file` | `Read` |
| `grep`, `rg`, `grep_search` | `Grep` |
| `glob`, `file_search` | `Glob` |
| `skill` | `Skill` |
| any other | unchanged |

The **trace text** a grader reads is each event on its own line: a `text`
event's text; a `tool` event's `<tool> <input>` followed, on the next line,
by its output.

### A4 — the graders

Each grader file's frontmatter is read as the cases write it; a run's score
is the weighted mean of its graders' scores, 0 to 1.

- **`regex`** — `target: trace` searches the trace text; a file target
  (`source: file` and `path`) searches that file in the workspace after the
  session. `pattern` is a Go regular expression. `match: contains` scores 1
  when it matches, `not_contains` when it does not. A file that does not
  exist scores 0 with the reason `no <path>`.
- **`tool_used`** — counts the trace's `tool` events whose `tool` equals
  `tool` and, when `input_match` is given, whose `input` matches it as a Go
  regular expression (not a substring: the cases write `\s+` and `|`); scores
  1 when the count is at least `min` (default 1) and at most `max` (no
  bound when absent).
- **`file_exists`** — scores 1 when `path` exists in the workspace.
- **`llm`** — its focus is the trace's last `text` event (the default), the
  trace text (`focus: trace`) or a file of the workspace (`source: file`,
  `path`; a missing file scores 0 with the reason `no <path>`). The judge is
  `claude -p --model <judge model> --output-format json
  --strict-mcp-config --setting-sources project`, started in an empty
  temporary directory (it sees nothing of the workspace), with no
  `--plugin-dir` (which is how a fake `claude` tells a judge call from a
  session), and its prompt on stdin: the grader's body and the focus,
  asking for `PASS` or `FAIL` on the last line. The answer is the JSON's
  `result` field, its cost `total_cost_usd`. It is asked three times and
  the majority decides, 1 or 0; an answer whose last line is neither counts
  as `FAIL`. A judge call that does not run, exits non-zero or prints no
  JSON with a `result` is not a vote: the run is an error with the reason
  `judge failed: <grader>: <cause>`.
- Any other grader type scores 0 with the reason `unsupported grader
  <type>`.

### A5 — the report

`<out>/report.json`:

```
{"agent": "...", "model": "...", "judge_model": "...", "runs": <n>,
 "cases": [{"name": "...", "phase": "...", "status": "scored|skipped|error",
            "reason": "...", "mean": <number or null>,
            "runs": [{"status": "scored|error", "reason": "...", "score": <number or null>,
                      "graders": [{"name": "...", "weight": <n>, "score": <0..1>, "reason": "..."}],
                      "cost_usd": <number>, "credits": <number or null>,
                      "judge_cost_usd": <number>, "wall_ms": <n>}]}]}
```

A grader's `name` is its file name without `.md`. A run that errored has
`score` null and no graders. A case is `error` when any run errored,
`skipped` when skipped, `scored` otherwise; `mean` is the unrounded mean of
its scored runs' scores, null when none. `credits` is null on Claude.

On stdout, a header and one line per case in name order, columns aligned
with two spaces between them (`text/tabwriter`, padding 2), then a total:

```
CASE                 PHASE   RUNS  MEAN  COST   NOTES
<name>               <phase> <r>   0.67  $0.12  <reason, or graders below 1 joined by ", ">
<c> case(s) · <s> skipped · <e> error(s) · $<total>
```

`NOTES` holds a skipped or errored case's reason, else the names of the
graders that scored below 1 in any run, each once, in name order.
`MEAN` has two decimals, `-` when null; `COST` and the total are the
sessions' and the judge's cost in dollars with two decimals; `<r>` is the
runs started for the case (0 for a skipped case, whose `MEAN` is `-` and
`COST` `$0.00`); `<c>` counts every selected case.

Exit 0 when no case is `error`, 1 when any is; 2 as in A1.

### A6 — the docs

- tools/agent-eval/README.md: what the tool is, the flags, the layout under
  `--out`, and what differs from `claude plugin eval` (one arm; the phase's
  fence instead of `allowed_tools` and `max_turns`; a Claude judge; skipped
  cases; `--vloop`).
- `docs/guide/evals.md` gains a section `## Other agents`, after
  `## Running` and before `## vloop's suites`: the command, a
  pointer to the README, and the qualifying rule — **a model qualifies for
  a phase on an agent when, at 3 runs, every case of that phase scores a
  mean no lower than the phase's default Claude model scores on the same
  tool with the same runs.**
- `docs/design-notes/multi-agent-backends.md`: after the sentence ending
  `another agent needs it first.`, and before `The fence audit is worth
  doing`, the sentence `It is tools/agent-eval since B33.`

### Decided here, because the cited documents leave it underdetermined

- Three judge votes, as `claude plugin eval`'s reports show its judge
  voting.
- `path.Match` for `--case`, the glob `claude plugin eval --help` offers.
- The rule compares against Claude on the same tool, not against `claude
  plugin eval`'s scores: different grants and judges make those
  incomparable.
- No adapter refuses a phase today, so there is no "unsupported phase"
  skip.

### Violations the review must rule on

- Any change to the argv, record or output of a `vloop` session.
- A case or grader file under `plugin/evals/` changed to suit the tool.
- A grader type scored as passing when the tool cannot evaluate it.
- A test that runs a real `claude` or `copilot`.

## Worked example

Fixture cases live in the tool's test data, read with `--cases`. Case
`fixture-review`: `prompt.md` body `/vloop:review T1`; a scaffold that
creates `.vloop/tmp/`; graders `a` (regex, trace, `contains`,
`verdict\.json: ok`, weight 1), `b` (regex, file `.vloop/tmp/verdict.json`,
`"verdict": *"FAIL"`, weight 2), `c` (tool_used, `Bash`, `input_match:
vloop schema validate`, weight 1) and `d` (llm, focus the file, weight 2).
Case `fixture-operate`: body `/vloop:operate`. A fake `copilot` on `PATH`
prints a session that calls `bash` with `vloop schema validate verdict/v1
.vloop/tmp/verdict.json` (result `.vloop/tmp/verdict.json: ok`) and writes
`.vloop/tmp/verdict.json` with `"verdict": "FAIL"`; a fake `claude` judge
answers `PASS`; a fake `vloop` stands for `--vloop`.

```
agent-eval --agent copilot --model m --cases <td> --case 'fixture-review' --out <tmp>/o1
  -> exit 0; <tmp>/o1/fixture-review/run-1/trace.jsonl has a tool event "Bash" whose
     input contains "vloop schema validate"; report.json: status scored, mean 1,
     graders a, b, c, d each 1
same, the fake copilot writing "verdict": "PASS", --out <tmp>/o2
  -> exit 0; b 0, the others 1, mean 0.6666…; the table's MEAN 0.67, NOTES "b"
same, the fixture's scaffold exiting 3, --out <tmp>/o3
  -> exit 1; the run's status error, reason beginning "scaffold exited 3:"
agent-eval --agent copilot --model m --cases <td> --case 'fixture-operate' --out <tmp>/o4
  -> exit 0; status skipped, reason "no phase"
agent-eval --list --cases <td>
  -> exit 0; a line for fixture-operate with "-" and 0, one for fixture-review with "review" and 4
agent-eval --list
  -> exit 0; one line per case of plugin/evals/, none an error
agent-eval --agent nope --model m --out <tmp>/o5
  -> stderr "agent-eval: unknown agent nope", exit 2
agent-eval --agent claude --model m --cases <td> --out <tmp>/o1   (not empty)
  -> stderr "agent-eval: <tmp>/o1 is not empty", exit 2
agent-eval --agent claude --model m --cases <td> --case 'fixture-review' --out <tmp>/o6,
  a fake claude streaming the same session in stream-json
  -> exit 0; the same trace events as o1's, mean 1
a vloop run session's Claude argv                             -> ends in "--output-format" "json", as before
```

The last line is not run by hand: `TestSessionArgs` and
`TestRunClaudeInvocations` cover it and must pass unchanged. In the o6
line the fake `claude` answers a session (its argv has `--plugin-dir`)
with the stream, and a judge call (no `--plugin-dir`) with
`{"result":"PASS","total_cost_usd":0}`.

## Acceptance, at close (the operator's)

No session can run a real agent. On macOS (`gate-review-real-misses` is
macOS-only), with `gh` logged in (a Copilot session signs in through `gh
auth token`), and with `vloop` built from the branch installed on `PATH`,
which the sessions and the hook both run, `go run ./tools/agent-eval --agent claude --model sonnet --case
'gate-review-*' --runs 1 --out <scratch>/claude` and the same with `--agent
copilot --model gpt-6.1-terra --out <scratch>/copilot`. Expected: both exit
0 with every gate-review case scored. The scores go in
`docs/design-notes/copilot-cli-experiments.md` and on #84.

## Out of scope

- A without-plugin arm, the `tool_order` and `baseline` graders, a
  concurrency flag.
- Qualifying any model in the run itself; changing a default agent or
  model.
- New eval cases, or changes to existing cases and graders.
- A `vloop` subcommand, or shipping the tool.
- Other agents than Claude and Copilot (#85).

## Constraints

- Gates run on macOS's BSD tools: no empty alternatives in `grep -E`, no
  `sed -i` without a suffix, no `\+` or `\|` in basic regexes.
- Sessions never move refs.
- Tests build their fixtures in temporary directories and address them with
  `git -C`; fake `claude`, `copilot` and `vloop` are scripts on a temporary
  `PATH`.
- No test or gate runs a real `claude` or `copilot`, or the network.
- The tool's tests run under `go test ./...`; every configured check stays
  green after every task.
- Repo-relative paths everywhere.

## Shape

5–6 tasks; every configured check green after each.

Every gate judges through the command, `go run ./tools/agent-eval`, with
fake `copilot`, `claude` and `vloop` scripts on a temporary `PATH` and
fixture cases the gate builds in a temporary directory; no gate runs the
repository's tests (P-8), which the configured check runs.

1. **The flags, the cases read and `--list` (A1).** Gated on A1's refusals
   and on `--list` over `plugin/evals/` and over a fixture directory.
2. **A Copilot run end to end, without the judge (A2, A3, A5; A4's
   `regex`, `tool_used` and `file_exists`).** Gated on the o1, o3 and o4
   lines with the fixture's grader `d` left out.
3. **The `llm` grader (A4).** Gated on o1 and o2 in full.
4. **Claude streaming (A2, A3).** Gated on the o6 line.
5. **The docs (A6).**
6. **Close:** every gate at once, and the worked example line for line.
<!-- vloop:run-record:begin -->
## Run record

Generated by vloop brief close on 2026-11-03. Recompute with: vloop metrics B20261102-2301-b33-i81-agent-eval-runner.loop-brief

```
B20261102-2301-b33-i81-agent-eval-runner  consumed · not merged
 tasks     6 planned · 6 done · 0 blocked · first-pass 5/6
 size      delivered  code 1,161 · test 898 · docs 57 · test:code 0.77
           churn      code 1,172 · test 898 · docs 57 · rework 1.01
 time      agent 16.9 min (work 14.5 · review 2.5) · plan 12.5 min · gate review 2.0 min · verify 1.7 min · gates 0.3 · wall 60.9 min
 rate      68.5 code lines/min · 121.5 incl. tests
 cost      $7.70 · plan 3.79 · work 2.27 · review 0.75 · gate review 0.57 · verify 0.31 · $6.63/kloc
 models    plan claude-opus-5-5 · work claude-sonnet-5-5 · review claude-sonnet-5-5 · gate review claude-sonnet-5-5 · verify claude-sonnet-5-5
 defects   in-loop 0 · operator 2 · escaped 0 · removal efficiency 100%
 interventions  3 · recommended 0 · other-option 0 · adjusted 0 · different 0 · no-options 3
 instructions  plan 4,693 · gate review 2,338 · work 2,435 · review 2,456 · verify 2,109 words · guidelines 2 rules
```

```
id  area  kind     att  code+  test+  docs+  other+  test:code  $/kloc  agent  cost   model
T1  -     feature  1    373    132    0      0       0.35       1.19    2m03s  $0.45  claude-sonnet-5-5
T2  -     feature  2    557    225    0      0       0.40       1.99    6m37s  $1.11  claude-sonnet-5-5
T3  -     feature  1    116    108    0      0       0.93       3.64    3m44s  $0.42  claude-sonnet-5-5
T4  -     feature  1    126    81     0      0       0.64       3.35    1m53s  $0.42  claude-sonnet-5-5
T5  -     docs     1    0      0      57     0       -          -       0m56s  $0.25  claude-sonnet-5-5
T6  -     test     1    0      352    0      0       -          -       1m43s  $0.37  claude-sonnet-5-5
```

- I20261103-005740-t2-s-gate-compared-an-unrounded-score-wi — verification-finding, plan, found by assistant: T2's gate compared an unrounded score with awk's six-digit print; fixed in T2, T3 and T6 and resumed
- I20261103-005742-the-design-note-s-two-hard-parts-still-c — verification-finding, work, found by operator: the design note's 'two hard parts' still called the runner needed after A6's sentence; found by the verify session, fixed by the operator's assistant
<!-- vloop:run-record:end -->

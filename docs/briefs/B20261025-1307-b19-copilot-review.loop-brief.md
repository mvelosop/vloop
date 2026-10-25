---
name: B20261025-1307-b19-copilot-review.loop-brief
description: Let review sessions run on GitHub Copilot CLI through a second adapter behind B18's seam — the embedded plugin, an isolated Copilot home, a deny list that keeps a session from acting outward, credits recorded and priced — marked experimental until B20 fences it
kind: brief
status: draft
created: 2026-10-25
seeds: The Copilot series' second brief (docs/design-notes/multi-agent-backends.md → "The sequence, as briefs"), settled with the operator on 2026-10-25
depends-on: [B20261025-0930-b18-agent-per-phase.loop-brief]
---
# Brief — vloop B19: GitHub Copilot CLI for review sessions

- **Starting point:** extends `main` at `a7ec05e` (B18 merged, the repository
  upgraded to v0.12.0). The planner pins the base it plans from as the
  **base** every gate compares against — never `HEAD~n`.
- **Produced by:** the Copilot series (B18–B20) and the design act of
  2026-10-25, after B18's run; the CLI facts are
  `docs/design-notes/copilot-cli-experiments.md` and GitHub's CLI command
  reference as checked on 2026-10-25.

## Executive summary

B18 gave vloop a seam with one adapter per agent, and Claude is the only one.
The experiments showed Copilot CLI runs vloop's review skill unchanged and that
a reviewer from another family finds what Claude misses, but vloop cannot start
it.

**The goal:** `agent.review = copilot` makes every review session of a run a
GitHub Copilot CLI session, with the model of your choice, while nothing it can
do reaches outside the repository through GitHub, git or the network — and the
operator is told plainly that it is experimental until B20 fences it.

- **C1 — the Copilot adapter.** Starts `copilot -p "/vloop:review <task>"` with
  the plugin embedded in the binary, the review's model and effort, a Copilot
  home of its own, and the repository's guidelines; maps Copilot's JSON Lines
  onto `session/v3`.
- **C2 — a deny list for what a review must never do.** No `gh`, no push or
  remote changes, no network commands, no global git config, no web fetch, no
  built-in GitHub MCP server; file tools keep Copilot's path verification.
- **C3 — credits recorded and priced.** `credits` on the session record and in
  metrics; `cost_usd` estimated from a configured credit price, so the cost
  ceiling still counts review spend.
- **C4 — experimental, said where it is chosen.** `vloop doctor`, the run
  preflight and the configuration guide say a Copilot review session is not
  fenced until B20.
- **C5 — proven on the real CLI.** At close the operator runs a real `vloop run`
  with Copilot reviews and a deny probe, and records both in the findings note.

Unchanged: Claude sessions byte for byte, every other phase on Claude, the
fence, gates, guards and budgets' logic.

## Why this shape, and what was rejected

Decided with the operator on 2026-10-25. **Do not re-litigate these** — the
planner inherits them.

- **Review only.** `copilot` is a valid value of `agent.review` alone; every
  other `agent.<phase>` stays `claude` only. The replays qualified Copilot for
  review; work, plan, gate review and verify were not tried.
  *Rejected:* opening all phases now (unqualified, and a work session edits the
  code the fence protects).
- **A denylist until B20, with cheap guards, not a fence.** The driver's B8
  checks already revert or catch anything a review session does inside the
  repository; what they cannot see is outward or outside the repository. B19
  denies those by name and B20 follows right away with the fence.
  *Rejected:* an allowlist (the experiments' pilot denied `printf` and the
  reviewer never saw the diff); waiting for B20 before any Copilot session.
- **The plugin embedded in the binary, never an installed one.** `--plugin-dir`
  points at the plugin the driver already extracts, so a Copilot session runs
  the same skills as a Claude one, at the binary's version.
  *Rejected:* `copilot plugin install vloop@vloop` (the marketplace's version,
  not the binary's, and a global install).
- **A Copilot home per session, inside the repository's scratch.** The adapter
  prepares it under `.vloop/tmp/` before each session; nothing is read from or
  written to `~/.copilot`. Login survives: Copilot keeps its token in the
  system credential store, or reads `COPILOT_GITHUB_TOKEN`, `GH_TOKEN` or
  `GITHUB_TOKEN`.
  *Rejected:* the user's `~/.copilot` (global state, rule 1 of `CLAUDE.md`).
- **The model is explicit.** `model.review` defaults to the Claude alias
  `sonnet`, which is not a Copilot model id; with `agent.review = copilot` the
  run refuses a `model.review` that comes from the default.
  *Rejected:* mapping Claude aliases to Copilot ids (a table that goes stale
  with every model release).
- **Cost is estimated, and says so.** Copilot reports credits, not dollars.
  `cost_usd` is credits × `copilot.credit-usd`, and `cost_source` is
  `"estimated"`; the run's cost ceiling counts it.
  *Rejected:* leaving Copilot sessions at `cost_usd: 0` (the ceiling would not
  see them); budgets in credits (a second unit for every budget).
- **`session/v3` and `metrics/v3` gain `credits` before their first release.**
  Neither has been released (B18 merged after v0.12.0), so they are amended in
  place rather than bumped.
- **The real CLI proves what a stub cannot.** Gates run a stub `copilot` and
  stay offline. Whether Copilot's matching denies the forms C2 names, and how
  it reports a denial, is proven at close by the operator on the real CLI.

## Binding references

- `docs/design-notes/copilot-cli-experiments.md` — the command line, the JSON Lines events and what maps onto a record (`result`, `session.usage_checkpoint`, `assistant.turn_start`), credits, the permission findings, the skill-name collision
- `docs/design-notes/multi-agent-backends.md` — "The two hard parts" and B19's open questions in "The sequence, as briefs", which this brief settles
- `internal/driver/agent.go` — the `Adapter` interface and the Claude adapter the Copilot adapter sits beside; `Output`, `CostReported`, `CostNone`
- `internal/driver/session.go` — the runner: masking, stderr capture, timeouts, the record it writes from an adapter's `Output`
- `internal/driver/cleartmp.go` — when `.vloop/tmp/` is emptied during a run, which is why the Copilot home is prepared per session
- `internal/config/config.go` — the `agent.<phase>` keys and their `Valid` lists, and how a new numeric key is declared
- `internal/cli/doctor.go` — the agent preflight and where a warning is printed
- `schemas/session.v3.json` — the record that gains `credits`
- `schemas/metrics.v3.json` — the metrics document that gains `credits`
- `docs/domain/execution/session.md` — S-1 to S-4, which a Copilot session must keep; the record section
- `docs/guide/configuration.md` — the key table, where `agent.review`'s `copilot` and `copilot.credit-usd` are documented
- `cmd/vloop/run_harness_test.go` — the stub claude; a stub copilot follows its pattern

## Behaviour contract

### Configuration

- `agent.review` accepts `claude` (default) and `copilot`. `agent.plan`,
  `agent.work`, `agent.gate-review` and `agent.verify` accept `claude` only;
  `vloop config set agent.work copilot` refuses with
  `vloop: invalid value "copilot" for agent.work: want one of claude`, exit 2.
- A new key `copilot.credit-usd`: a positive number, default `0.0066667`
  ($10 for 1,500 credits, the Copilot Pro plan). Environment variable
  `VLOOP_COPILOT_CREDIT_USD`.

### C1 — the session

With `agent.review = copilot`, every review session starts as

```
copilot -p "/vloop:review <task>" --model <model.review> [--effort <effort.review>]
  --plugin-dir <the extracted embedded plugin> --no-ask-user --output-format json
  --allow-all-tools <the C2 denials> --disable-builtin-mcps
```

- **Working directory:** the repository root, as for Claude.
- **Environment:** the runner's, with the variables a Claude session drops also
  dropped, the plan hash added as for Claude, and `COPILOT_HOME` set to a
  directory the adapter creates empty under `.vloop/tmp/` for that session.
- **Guidelines:** when `.vloop/repo-guidelines.md` exists and is not empty, the
  session receives it as custom instructions, through
  `COPILOT_CUSTOM_INSTRUCTIONS_DIRS` pointing at a directory the adapter
  prepares in that session's scratch. Copilot also loads `CLAUDE.md` itself.
- Never passed: `--allow-all-paths`, `--add-dir`, `--sandbox` or `--no-sandbox`
  (B20 decides the sandbox).

The record, from the JSON Lines on stdout:

- **Is there a record?** Exactly when a `result` event is present. Otherwise
  the runner logs `SESSION RECORD MISSING`, as for Claude.
- `duration_ms` is `result.usage.sessionDurationMs`; `turns` is the number of
  `assistant.turn_start` events; `is_error` is true when `result.exitCode` is
  not 0.
- `credits` is `session.usage_checkpoint.data.totalNanoAiu` / 1e9, the last
  checkpoint's; `cost_usd` is `credits × copilot.credit-usd`; `cost_source` is
  `"estimated"`, or `"none"` with both 0 when no checkpoint was printed.
- `models_used` has one entry for the session's model, with zero tokens and the
  estimated cost; Copilot reports no tokens per model.
- `permission_denials` is `[]`: Copilot's denial event is not yet known (C5
  records it for B20).
- Nothing from `result.usage.codeChanges` is recorded (it carries absolute
  paths).
- `agent` is `"copilot"`.

### C2 — the denials

Every Copilot session denies, so that none of these runs whatever form it is
written in:

- `gh`, in any form;
- `git push`, `git remote` and any `git config` that writes outside the
  repository (`--global`, `--system`, `--file`);
- `curl`, `wget`, `ssh`, `scp`, `nc`;
- `rm`, and the git commands the review skill already forbids (`commit`,
  `reset`, `checkout`, `switch`, `stash`, `branch`, `tag`);
- the URL tool (Copilot's `url` kind), so no web fetch.

Copilot's pattern syntax is `Kind(argument)`, with `shell(<stem>:*)` matching a
command stem followed by a space; the exact patterns are the implementer's,
and C5 proves them on the real CLI against the forms a model would actually
write: the plain command, a chain (`git diff && gh pr view`), `sh -c`, `env`
and `git -C . push`. A form C5 finds unblocked is recorded, not patched around
in this run, unless a deny pattern blocks it.

### C3 — credits in metrics

- `session/v3` gains an optional `credits` (number); a Claude record has none.
  `cost_source` gains `"estimated"`.
- `metrics/v3` gains `credits`: the run's total, and per phase like `cost`, 0
  where none.
- The run's cost ceiling, the per-task costs and the metrics text use
  `cost_usd`, estimated or reported, unchanged in logic. The metrics text marks
  an estimated total with `≈` (`cost ≈$6.91`).

### C4 — experimental

- `vloop doctor` and the run preflight, when any phase's agent is `copilot`:
  check `copilot` is on `PATH` and `copilot --version` runs, as they do for
  `claude`; and print
  `! agent.review is copilot: experimental — review sessions are not fenced until B20 (docs/guide/configuration.md)`
  as a warning, not a problem.
- The run preflight refuses, exit 1, when `agent.review` is `copilot` and
  `model.review` comes from the default:
  `vloop: agent.review is copilot but model.review is the default "sonnet", a Claude alias: set model.review to a Copilot model id`.
- `vloop doctor` warns when the repository has a project skill in
  `.claude/skills/` whose name ends in `review`
  (`! .claude/skills/loop-review may be loaded instead of /vloop:review by a copilot session`),
  the collision the experiments recorded.
- `docs/guide/configuration.md` documents `copilot` for `agent.review`,
  `copilot.credit-usd`, the experimental status and what C2 does and does not
  stop.

### C5 — the acceptance, by the operator, on the real CLI

At close, with Copilot CLI installed and logged in:

1. a real `vloop run` of a small fixture brief in a scratch repository with
   `agent.review = copilot` and an explicit `model.review`, which completes,
   writes `session/v3` review records with `agent: "copilot"`, credits and an
   estimated cost, and whose verdicts validate;
2. a deny probe: a real Copilot session started with exactly the argv and
   environment the adapter builds (taken from the adapter, not retyped), asked
   to try each C2 form, whose results — denied or not, and the event Copilot
   printed for a denial — are recorded in
   `docs/design-notes/copilot-cli-experiments.md`.

### Decided here, because the cited documents leave it underdetermined

- `effort.review` is passed as `--effort`; Copilot accepts the same five levels
  vloop does.
- The Copilot home is not reused across sessions; nothing Copilot writes there
  survives the session's scratch.
- A Copilot review session's verdict, proposal handling and tree revert are
  the driver's, exactly as for Claude: nothing in the review path knows which
  agent ran.

### Violations the review must rule on

- Any change to a Claude session's argv, working directory, environment or
  record (B18's comparison gate must still pass unchanged).
- A Copilot session that reads or writes `~/.copilot`, uses an installed
  plugin, or is started with `--allow-all-paths`.
- A denial in C2 missing from the argv, or a gate that asserts deny patterns
  only by string without the stub checking each one is passed.
- Copilot's output parsed in the driver instead of the adapter, or
  `codeChanges` paths reaching a record.
- `copilot` accepted for any phase but review.

## Worked example

With the run harness and a stub `copilot` that logs its argv one argument per
line, its working directory and environment, and prints JSON Lines with a
`session.usage_checkpoint` of `totalNanoAiu: 14200000000` and a `result` with
`exitCode: 0`, `sessionDurationMs: 18000`, and three `assistant.turn_start`
events:

```
vloop config set agent.review copilot && vloop config set model.review gpt-5.6-terra
  -> both exit 0

vloop config set agent.work copilot
  -> vloop: invalid value "copilot" for agent.work: want one of claude    exit 2

VLOOP_AGENT_REVIEW=copilot vloop run <brief>          (model.review from the default)
  -> vloop: agent.review is copilot but model.review is the default "sonnet", a Claude alias: set model.review to a Copilot model id
                                                                          exit 1

vloop doctor                                          (agent.review copilot, model set)
  -> ! agent.review is copilot: experimental — review sessions are not fenced until B20 (docs/guide/configuration.md)
                                                                          exit 0

vloop run <brief>                                     (one task)
  -> the review session: copilot -p "/vloop:review T1" --model gpt-5.6-terra ... with
     every C2 denial and --disable-builtin-mcps, COPILOT_HOME under .vloop/tmp/,
     cwd the repository root                                              exit 0
     sessions/<n>-review.json: "agent": "copilot", "credits": 14.2,
       "cost_usd": 0.0946671 (14.2 × 0.0066667), "cost_source": "estimated",
       "turns": 3, "duration_ms": 18000, "permission_denials": []
     plan, work and verify sessions: claude, byte-identical to B18's

the same with a stub that prints no result event
  -> SESSION RECORD MISSING review T1 (iteration 1)

vloop metrics --json
  -> "credits": {"total": 14.2, ...}, "agents": {"review": ["copilot"], ...}
```

The gate threads the cost it computes from the stub's credits and the
configured price; it does not hardcode `0.0946671`. B18's comparison gate,
re-run unchanged with every phase on `claude`, still finds no difference.

## Out of scope

- The fence for non-Claude agents, hooks, `vloop fence check`, Copilot's
  sandbox (B20).
- A second reviewer, or any way to combine verdicts (B20).
- Copilot for plan, work, gate review or verify.
- Recording Copilot's permission denials (their event is unknown until C5).
- Mapping Claude model aliases to Copilot model ids.
- Budgets in credits, or a premium-request budget.
- Grok, Codex, Gemini or any third agent.
- Renaming or moving this repository's `.claude/skills/loop-*` skills (the
  doctor warning names the risk; the operator decides).
- A release between B19 and B20 (a roadmap decision, not a task).

## Constraints

- Go, at the version in `go.mod`; no new dependencies.
- Must keep building and passing on macOS, Linux and Windows.
- **No gate runs the real Copilot CLI or touches the network**; gates use a
  stub `copilot`, as they use a stub `claude`. C5 is the operator's.
- **Claude sessions stay byte-identical:** B18's comparison gate runs unchanged
  and passes.
- **Gates follow the diff surface.** Each task gates on what it touches; no
  task runs the whole test suite.
- **No task may weaken a check to pass.** An earlier test changes only where it
  pins behaviour the contract changes, and only to the new behaviour; the work
  session names the contract item in its notes and the review cites it.
- **New tooling is proven by fixtures**, including where it must fail: a stub
  that prints no `result`, one with no usage checkpoint, a default
  `model.review`, `copilot` for another phase.
- Tests build fixtures in temporary directories and address them with
  `git -C`; sessions never move refs.
- Repo-relative paths everywhere. No absolute paths in any file or commit
  message.

## Shape

5 to 6 tasks, each verifiable by a single command. Order is load-bearing:

1. **The stub `copilot` and the configuration** — `copilot` valid for
   `agent.review` only, `copilot.credit-usd`. Gated on the config lines of the
   worked example.
2. **The Copilot adapter: command, environment, Copilot home, guidelines and
   C2's denials.** Gated on the stub's argv, cwd and environment for a review
   session, every denial checked, and B18's comparison gate unchanged.
3. **The record: parsing the JSON Lines, credits and the estimated cost**, and
   `credits` in `session/v3` and `metrics/v3`. Gated on the worked example's
   record and metrics lines, the missing-record case and the no-checkpoint case.
4. **C4 — doctor, preflight and the guide.** Gated on the doctor and preflight
   lines, the skill-collision warning on a fixture, and the guide's rows.
5. **Close:** every gate at once and the worked example line for line. The
   operator then runs C5 and records it before merging.

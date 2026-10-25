---
name: B20261025-0930-b18-agent-per-phase.loop-brief
description: Choose the agent CLI per session phase through a seam the driver starts every session through, with Claude Code as the only agent and its sessions unchanged, and record the agent in sessions and metrics
kind: brief
status: consumed
created: 2026-10-25
seeds: The operator's draft of 2026-10-23, written after the GitHub Copilot CLI experiments, revised on 2026-10-25 against the code; the first of the Copilot series B18–B20
depends-on: [B20261015-1651-b17-release.loop-brief]
---
# Brief — vloop B18: agent per phase, Claude only, no change in behaviour

- **Starting point:** extends `main` after the design notes of the Copilot
  series merge (`main` at `bacb262`, v0.12.0, plus
  `docs/design-notes/multi-agent-backends.md` and
  `docs/design-notes/copilot-cli-experiments.md`). The planner pins the base it
  plans from as the **base** every gate compares against — never `HEAD~n`.
- **Produced by:** operator decision, 2026-10-23, after the GitHub Copilot CLI
  experiments (`docs/design-notes/copilot-cli-experiments.md`): review and work
  sessions replayed through Copilot with Claude and GPT models; revised in the
  design act of 2026-10-25.

## Executive summary

Every session vloop starts is a `claude -p` session, built inline in the
driver. The experiments showed that GitHub Copilot CLI runs vloop's skills
unchanged and that a reviewer from another model family adds real findings —
but vloop has no way to start any agent other than Claude.

**The goal:** which agent runs each phase becomes a setting, and the driver
starts every session through one seam with one adapter per agent, so the next
brief can add Copilot by adding an adapter. Nothing a Claude session receives
changes.

- **A1 — `agent.<phase>` settings.** Five keys beside `model.<phase>` and
  `effort.<phase>`; `claude` is the only valid value in this run.
- **A2 — the seam.** Everything agent-specific moves into a Claude adapter:
  preflight, per-run preparation, the command and environment, deciding whether
  the output is a record, and mapping it. Claude's argv, working directory and
  environment stay byte-identical, proven against the base binary.
- **A3 — the agent recorded.** Session records move to `session/v3` with
  `agent` and `cost_source`; `vloop metrics --json` moves to `metrics/v3` with
  `agents` per phase. Older records and snapshots read as before.

Unchanged: the fence, budgets, gates, checks, guards, and everything Claude
sessions see.

This is the first of three briefs. B19 adds a GitHub Copilot CLI adapter for
review sessions; B20 adds the fence for non-Claude agents and a second
reviewer. This run leaves the seam ready for those without building any of
them.

## Why this shape, and what was rejected

Decided with the operator on 2026-10-23, revised 2026-10-25. **Do not
re-litigate these** — the planner inherits them.

- **The agent is chosen per phase, in config.** Mixing agents is the point (for
  example Claude for plan and work, another agent for review), so the choice
  sits next to `model.<phase>` and `effort.<phase>` with the same precedence:
  environment, then file, then default.
  *Rejected:* one global `agent` key; it cannot express the mix the next briefs
  need.
- **The agent is resolved when and how the phase's model is.** Work, review and
  verify resolve once at run start (`ResolveRun`); plan and gate-review resolve
  per session, as their model and effort do today. The driver's input guards
  already keep a session from changing the config under a run.
  *Rejected:* freezing all five at run start, which would make `agent` and
  `model` of the same phase resolve at different times.
- **No per-task agent override in this run.** Tasks keep their `model`/`effort`
  overrides for their own phase.
  *Rejected:* adding it now "while we are here"; nothing needs it and it widens
  the state schema.
- **One adapter per agent behind a single seam.** An adapter owns everything
  agent-specific: its preflight checks, any per-run preparation, the command
  and environment of a session, whether its output is a record, and the mapping
  of that output onto the session record. The driver keeps everything
  agent-independent: gates, checks, the input guards and git guards around a
  session, the lock, budgets, masking, stderr capture, timeouts, record
  numbering, and how a session's result decides the iteration.
  *Rejected:* an argv template in config that users edit; it would let any
  configuration bypass the fence and the isolation flags, and nothing could gate
  it.
- **"Is there a record" belongs to the adapter.** Today the runner treats
  stdout that is not one JSON document as no record; Copilot prints JSON Lines
  (`docs/design-notes/copilot-cli-experiments.md`). The adapter says whether
  there is a record; the runner keeps the `SESSION RECORD MISSING` log line.
- **Unsupported agents are refused by config validation**, not at session
  start. The valid values are the registered adapters; in this run, `claude`.
- **Session records move to `session/v3` for every phase**, replacing
  `session/v1` and `session/v2`, and add `agent` and `cost_source`. Readers keep
  accepting v1 and v2, which mean `agent: "claude"`.
  *Rejected:* adding the fields to v1 and v2 in place; a record's schema must say
  what it contains.
- **Metrics move to `metrics/v3`**, adding `agents` per phase beside `models`.
  `metrics/v2` documents, saved snapshots included, stay readable.
- **The word is `agent`.** It names the CLI that ran a session. It is not the
  *agent* clock (`agent_ms`, the sum of session durations); the domain pages
  say so where both appear.

## Binding references

- `docs/domain/execution/session.md` — S-1 to S-4, what a session is and may not do, and the record section this run moves to `session/v3`; the page is reworded so a session is no longer defined as a `claude -p` process
- `docs/domain/measurement/metrics.md` — the per-phase `models` the new `agents` sits beside, and the agent clock the word must not be confused with
- `docs/design-notes/copilot-cli-experiments.md` — the Copilot facts the seam must leave room for: JSON Lines output, credits instead of dollars, `COPILOT_HOME`, the plugin loading unchanged
- `internal/driver/session.go` — the claude argv (`Args`, `WithGuidelines`), environment (`env`, `dropEnv`), the record decision and mapping (`Run`, `record`) the Claude adapter must reproduce byte for byte
- `internal/driver/session_test.go` — the existing argv expectations; they must pass unchanged
- `internal/driver/resolve.go` — how `model` and `effort` are resolved per phase (`ResolveRun` for work, review and verify; `ModelEffort` per session for plan and gate-review); `agent` follows the same rules
- `internal/config/config.go` — key definitions, `Valid` lists and the `invalid value … want one of …` error the new keys reuse
- `docs/guide/configuration.md` — the key table the new keys are added to
- `schemas/session.v1.json` — the record work, review and plan sessions write today
- `schemas/session.v2.json` — the record gate-review and verify sessions write today; identical fields to v1
- `schemas/metrics.v2.json` — the metrics format that gains `agents`
- `cmd/vloop/run_harness_test.go` — the stub claude, which logs its argv today; the comparison gate extends what it logs

## Behaviour contract

### Configuration

- Five keys: `agent.plan`, `agent.work`, `agent.review`, `agent.gate-review`,
  `agent.verify`. Default `claude`. Valid values: the registered agents, in
  this run exactly `claude`. Environment variables `VLOOP_AGENT_PLAN`,
  `VLOOP_AGENT_WORK`, `VLOOP_AGENT_REVIEW`, `VLOOP_AGENT_GATE_REVIEW`,
  `VLOOP_AGENT_VERIFY`. Stored under `[agent]` in `.vloop/config.toml`.
- `vloop config set agent.review copilot` refuses with the existing
  invalid-value message and exit 2, and writes nothing.
- An invalid value in the environment or the file is a source error (exit 1)
  for `vloop config get`, `vloop config list`, `vloop doctor` and `vloop run`,
  as for any other key.
- `vloop config list` shows the five keys with value and source like every
  other key.

### Sessions

- Each session is started through the adapter of its phase's agent, resolved as
  "Why this shape" says.
- **With every phase on `claude`, a session's argv, working directory and
  environment are byte-identical to the base binary's**, for every phase, with
  and without `.vloop/repo-guidelines.md`, with and without a plan hash, with
  and without an effort. Variables the base passes through (for example
  `CLAUDE_CODE_USE_FOUNDRY`, `ANTHROPIC_FOUNDRY_RESOURCE`,
  `ANTHROPIC_DEFAULT_SONNET_MODEL`) still reach claude unchanged; the ones the
  base drops (`VLOOP_ACTIVE_TASK`, `VLOOP_GATE_TASK`, the plan-hash variable
  before it is set) are still dropped.
- Stderr capture, masking, timeouts, the lock record, the
  `SESSION RECORD MISSING` log line (same cases: empty stdout, or stdout that
  is not claude's JSON) and every exit code are unchanged.

### Records and metrics

- Every session record is `session/v3`: every field of `session/v2`, with
  `ended` and `wall_ms` now required (every record already carries them), plus:
  - `agent`: the adapter that ran the session (`"claude"`);
  - `cost_source`: `"reported"` when the agent's output had a cost field (for
    claude, the key `total_cost_usd` is present), else `"none"`, with
    `cost_usd` 0.
- schemas/session.v3.json validates them; `vloop schema validate session/v3 <file>`
  works.
- `vloop metrics` reads v1, v2 and v3 records alike. `vloop metrics --json` is
  `metrics/v3`: everything `metrics/v2` has, plus `agents`, per phase like
  `models`: the sorted distinct agents that ran that phase's sessions
  (`["claude"]` for every phase that ran; for a phase that did not run, the
  same empty or null value `models` uses). v1 and v2 session records count as
  `claude`.
- Saved snapshots (`.vloop/state/metrics/<run-id>.json`) written as
  `metrics/v2` are read by `vloop brief close`, `vloop metrics --workspace` and
  the cross-brief table as before; new snapshots are `metrics/v3`.

### Preflight

- Each adapter declares its preflight checks and the phases it supports.
  Before the first session, `vloop run` runs the preflight of every agent
  configured for a phase the run will start. With only `claude` this is today's
  claude check, so `vloop doctor` and the run preflight print exactly what the
  base prints, and fail exactly when the base fails.

### Decided here, because the cited documents leave it underdetermined

- The seam's names, packages and file layout are the implementer's. What is
  pinned is the behaviour above and the responsibilities split in "Why this
  shape".
- `cost_source` has two values in this run. B19 adds the value for an agent
  whose cost vloop estimates.
- Fields a future agent may report (credits, peak context tokens) are not added
  here. B19 adds the ones Copilot reports.
- `docs/domain/execution/session.md` defines a session as a fresh process of
  its phase's agent, Claude Code by default.

### Violations the review must rule on

- Any claude flag, environment variable or record field that differs from the
  base in a default configuration, even if the tests still pass.
- Agent-specific logic left in the driver (for example the driver building
  `--plugin-dir`, parsing stdout as one JSON document, or reading claude's JSON
  fields itself) instead of in the Claude adapter.
- Agent-independent logic moved into the adapter (gates, checks, input or git
  guards, budget decisions, masking).
- A per-task agent override, a second adapter, a fence change, or more than one
  session per phase.
- A `session/v1` or `session/v2` record, or a `metrics/v2` snapshot, that
  `vloop metrics`, `brief close` or the workspace table no longer reads.

## Worked example

In a repository initialised by the changed binary, with no `[agent]` table in
the config:

```
vloop config get agent.review
  -> claude                                                              exit 0

vloop config get agent.review --json
  -> {"key":"agent.review","value":"claude","source":"default"}          exit 0
     (field order as the existing --json output prints it)

vloop config set agent.review copilot
  -> vloop: invalid value "copilot" for agent.review: want one of claude    exit 2
     (.vloop/config.toml unchanged)

VLOOP_AGENT_WORK=copilot vloop config list
  -> vloop: VLOOP_AGENT_WORK: invalid value "copilot" for agent.work: want one of claude
                                                                         exit 1

vloop config set agent.review claude && vloop config get agent.review --json
  -> {"key":"agent.review","value":"claude","source":"file"}             exit 0
```

With the run harness's stub claude, the same brief and the same stub scripts,
run once with the base binary and once with the changed binary:

```
diff <what the stub logged, base> <the same, changed>
  -> no differences, for plan, gate-review, work, review and verify      exit 0
     (argv one argument per line, the working directory, the environment)

the same diff with a planted extra flag in the changed binary
  -> the flag shows in the diff                                          exit 1

each .vloop/state/runs/<brief>/<run>/sessions/*.json
  -> "schema": "session/v3", "agent": "claude", "cost_source": "reported"
     vloop schema validate session/v3 <file>                            exit 0

a stub session whose JSON has no total_cost_usd
  -> "cost_source": "none", "cost_usd": 0

vloop metrics --json
  -> "schema": "metrics/v3", "agents": {"plan":["claude"], "work":["claude"], ...}
     every value of "models" equal to the base binary's for the same run
```

A run directory written by the base binary (`session/v1` and `session/v2`
records) and a `metrics/v2` snapshot still give the same `vloop metrics`,
`brief close --dry-run` and workspace numbers with the changed binary, with
`agents` all `claude`.

## Out of scope

- The GitHub Copilot CLI adapter (B19), and any other agent (Grok, Codex,
  Gemini).
- `COPILOT_HOME` isolation, local plugin marketplaces, custom instruction
  directories, the guidelines route for agents without
  `--append-system-prompt-file`.
- `vloop fence check`, hooks, any change to `fence/*.json`, shell parsing
  (B20).
- More than one session per phase (a second reviewer) and any policy to combine
  verdicts (B20).
- A per-task agent override.
- New record fields beyond `agent` and `cost_source` (credits, context tokens).
- Budgets in a unit other than dollars.
- `export/v1` and `vloop metrics export`: unchanged in this run.
- Microsoft Foundry or any provider setting: it already works through the
  environment claude inherits, and must keep working.
- Renaming `Runner.Claude` or other public names for its own sake, beyond what
  the seam requires.

## Constraints

- Go, at the version in `go.mod`; no new dependencies.
- Must keep building and passing on macOS, Linux and Windows.
- **Gates follow the diff surface.** Each task gates on what it touches; no
  task runs the whole test suite.
- **No task may weaken a check to pass.** An earlier test changes only where it
  pins behaviour the contract changes, and only to the new behaviour (here:
  where a record now reads `session/v3`); the work session names the contract
  item in its notes and the review cites it.
- **Byte-identity is proven, not asserted.** The comparison gate:
  - builds the base binary from the base SHA with `git archive <base> | tar -x`
    into a temporary directory — never `git worktree add`, which writes under
    `.git/` while the driver guards it;
  - extends the stub claude to log each argument on its own line, its working
    directory and its full environment, in a form the base binary's sessions
    produce the same way;
  - runs both binaries in the same fixture path, one after the other, so
    nothing needs normalizing; anything it does normalize is named in the gate
    and justified;
  - must be seen to fail on a planted change (an extra flag).
- Tests build fixtures in temporary directories and address them with
  `git -C`; sessions never move refs.
- Repo-relative paths everywhere. No absolute paths in any file or commit
  message.

## Shape

4 to 5 tasks, each verifiable by a single command. Order is load-bearing:

1. **The base-versus-changed comparison gate, plus its fixtures.** Builds both
   binaries, runs the stub through one session of each phase in each
   configuration the Sessions contract names, and diffs argv, working directory
   and environment. Proven on a planted difference.
2. **The `agent.<phase>` keys**, their validation, environment variables,
   config docs row and resolution. Gated on the config tests and the worked
   example's `config` lines.
3. **The seam and the Claude adapter**, including the record decision. The
   driver starts every session through it. Gated on task 1's comparison and the
   existing session tests.
4. **`session/v3`, `cost_source`, `metrics/v3` with `agents`**, readers for
   v1, v2 and v3 and for `metrics/v2` snapshots, and the domain pages. Gated on
   schema validation of a harness run's records and on metrics over a
   base-written run directory.
5. **Close:** every gate at once, and the worked example line for line.
<!-- vloop:run-record:begin -->
## Run record

Generated by vloop brief close on 2026-10-25. Recompute with: vloop metrics B20261025-0930-b18-agent-per-phase.loop-brief

```
B20261025-0930-b18-agent-per-phase  consumed · not merged
 tasks     4 planned (brief said 4–5) · 4 done · 0 blocked · first-pass 3/4
 size      delivered  code 666 · test 412 · docs 47 · test:code 0.62
           churn      code 666 · test 417 · docs 47 · rework 1.00
 time      agent 45.0 min (work 42.9 · review 2.1) · plan 16.8 min · gate review 1.4 min · verify 1.7 min · gates 3.1 · wall 92.7 min
 rate      14.8 code lines/min · 23.9 incl. tests
 cost      $6.85 · plan 3.77 · work 1.88 · review 0.63 · gate review 0.26 · verify 0.32 · $10.28/kloc
 models    plan claude-opus-5-5 · work claude-sonnet-5-5 · review claude-sonnet-5-5 · gate review claude-sonnet-5-5 · verify claude-sonnet-5-5
 defects   in-loop 1 · operator 2 · escaped 0 · removal efficiency 100%
 interventions  3 · recommended 1 · other-option 0 · adjusted 0 · different 0 · no-options 2
 instructions  plan 4,569 · gate review 2,088 · work 2,436 · review 2,291 · verify 1,806 words · guidelines 2 rules
```

```
id  area  kind      att  code+  test+  docs+  other+  test:code  $/kloc  agent   cost   model
T1  -     feature   2    24     80     6      0       3.33       23.91   6m18s   $0.57  claude-sonnet-5-5
T2  -     refactor  1    367    151    18     0       0.41       2.80    20m10s  $1.03  claude-sonnet-5-5
T3  -     feature   1    275    186    12     0       0.68       2.52    13m21s  $0.69  claude-sonnet-5-5
T4  -     chore     1    0      0      11     0       -          -       5m13s   $0.22  claude-sonnet-5-5
```

- B20261025-0930-b18-agent-per-phase/i1-gate — regression, work, found by gate: check failed
- I20261025-123516-the-domain-glossary-docs-domain-domain-m — verification-finding, brief, found by operator: the domain glossary (docs/domain/domain-model.md) still defined a session as a claude -p process with three phases; the brief named only session.md to reword; fixed by hand
- I20261025-123517-a-metrics-comment-still-named-only-metri — verification-finding, work, found by operator: a metrics comment still named only metrics/v2 after the move to metrics/v3; fixed by hand
<!-- vloop:run-record:end -->

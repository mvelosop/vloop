# Plan — B20261025-0930-b18-agent-per-phase

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **2/4 done** · iteration 3

**Brief:** `docs/briefs/B20261025-0930-b18-agent-per-phase.loop-brief.md` · **Updated:** 2026-10-25T11:43:28Z

## Progress

- [x] **T1** — Add the five agent.&lt;phase&gt; config keys and resolve them as model and effort are · 1 attempt(s)
- [x] **T2** — Start every session through a seam with a Claude adapter, and write session/v3 records
- [ ] **T3** — Report agents per phase in metrics/v3, reading v1, v2 and v3 records and metrics/v2 snapshots
- [ ] **T4** — Close the run: every gate at once and the worked example line for line

## Tasks

### T1 — Add the five agent.&lt;phase&gt; config keys and resolve them as model and effort are

`done` · 1 attempt(s) · depends on: none

Which agent CLI runs each phase becomes a setting beside model.&lt;phase&gt; and effort.&lt;phase&gt;, so later briefs can mix agents per phase. In this run claude is the only registered agent, so the keys exist, validate, list and resolve, but nothing a session receives changes. The seam task (T2) consumes the resolved agent; this task only defines and resolves it.

**Acceptance**

- Five keys exist: agent.plan, agent.work, agent.review, agent.gate-review, agent.verify; default claude; the only valid value is claude; environment variables VLOOP_AGENT_PLAN, VLOOP_AGENT_WORK, VLOOP_AGENT_REVIEW, VLOOP_AGENT_GATE_REVIEW, VLOOP_AGENT_VERIFY; `vloop config set` stores them under [agent] in .vloop/config.toml.
- In a repository initialised by the changed binary with no [agent] table, `vloop config get agent.review` prints `claude` (exit 0) and `--json` prints {"key":"agent.review","value":"claude","source":"default"}.
- `vloop config set agent.review copilot` prints `vloop: invalid value "copilot" for agent.review: want one of claude` on stderr, exits 2 and leaves .vloop/config.toml byte-identical.
- `VLOOP_AGENT_WORK=copilot vloop config list` prints `vloop: VLOOP_AGENT_WORK: invalid value "copilot" for agent.work: want one of claude` and exits 1; an invalid agent value in the environment or the file makes `vloop config get`, `vloop config list`, `vloop doctor` and `vloop run` exit 1, and `vloop run` starts no session.
- `vloop config list` shows each of the five keys as `agent.<phase>=claude (default)`, like every other key.
- The agent of work, review and verify is resolved once in ResolveRun at run start; the agent of plan and gate-review is resolved per session, when their model and effort are. There is no per-task agent override and the state schema is unchanged.
- docs/guide/configuration.md has a table row for each of the five keys with default `claude` and its `VLOOP_AGENT_<PHASE>` variable.
- With every VLOOP_AGENT_* set to claude, a run hands claude exactly what the base binary's run hands it (the gate's comparison).
- Committed tests the checks run cover the new keys (default, environment, file, invalid value refused by set, source error) and the agent's resolution; a key or resolver shipped with no test that exercises it is a violation.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T1/gate.sh
```

</details>

### T2 — Start every session through a seam with a Claude adapter, and write session/v3 records

`done` · depends on: T1

Every session vloop starts is built inline as `claude -p` in the driver; this task moves everything agent-specific behind one seam with a Claude adapter (preflight and the phases it supports, per-run preparation, command and environment, deciding whether stdout is a record, mapping it onto the record), so B19 can add Copilot by adding an adapter. Nothing a Claude session receives may change: the gate builds the base binary and diffs argv, working directory and environment for every phase. Records gain the adapter's identity: session/v3 with agent and cost_source.

**Acceptance**

- Every session (plan, gate-review, work, review, verify) is started through the adapter of its phase's resolved agent; claude is the only adapter registered, and it owns its preflight checks and the phases it supports, any per-run preparation, the command and environment, whether stdout is a record, and the mapping onto the record.
- With every phase on claude, the argv, working directory and environment claude receives are byte-identical to the base binary's for every phase, with and without .vloop/repo-guidelines.md, with and without a plan hash, with and without an effort; CLAUDE_CODE_USE_FOUNDRY, ANTHROPIC_FOUNDRY_RESOURCE and ANTHROPIC_DEFAULT_SONNET_MODEL still reach claude; VLOOP_ACTIVE_TASK, VLOOP_GATE_TASK and a stale VLOOP_PLAN_SHA256 are still dropped.
- Violation: any claude flag, environment variable or record field (beyond schema, agent and cost_source) that differs from the base in a default configuration, even if the tests pass.
- Violation: agent-specific logic left in the driver — the driver building --plugin-dir, --settings or --output-format, parsing stdout as one JSON document, or reading claude's JSON fields (total_cost_usd, modelUsage, permission_denials) itself — instead of in the Claude adapter.
- Violation: agent-independent logic moved into the adapter — gates, checks, input or git guards, the lock, budget decisions, masking, stderr capture, timeouts or record numbering.
- Violation: a per-task agent override, a second adapter, a fence change, or more than one session per phase.
- Every record is session/v3: every field of session/v2, ended and wall_ms required, plus agent ("claude") and cost_source — "reported" when claude's output has the key total_cost_usd, else "none" with cost_usd 0. schemas/session.v3.json validates them and `vloop schema validate session/v3 <file>` works; schemas/session.v1.json and session.v2.json stay and still validate old records.
- Stderr capture, masking, timeouts, the lock record, the SESSION RECORD MISSING log line (empty stdout, or stdout that is not claude's JSON) and every exit code are unchanged; `vloop doctor` and the run preflight print exactly what the base prints and fail exactly when it fails.
- Readers that consume records during a run (spend for budgets, the lock) read session/v3 as they read v1 and v2.
- docs/domain/execution/session.md defines a session as a fresh process of its phase's agent, Claude Code by default (no longer as a `claude -p` process), and its record section documents session/v3 and says v1 and v2 records mean agent claude; where `agent` and the agent clock (agent_ms) both appear, the page says they differ.
- The run harness stub in cmd/vloop/run_harness_test.go also logs each argument on its own line, its working directory and its environment, and a committed harness test pins, for a default run, each phase's argv and that the pass-through variables reach claude and the dropped ones do not; committed tests cover session/v3 records with cost_source reported and none. An earlier test changes only where it pins session/v1 or session/v2 in a record, the work notes name that contract item, and the argv tests in internal/driver/session_test.go are unchanged.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T2/gate.sh
```

</details>

### T3 — Report agents per phase in metrics/v3, reading v1, v2 and v3 records and metrics/v2 snapshots

`pending` · depends on: T2

`vloop metrics --json` moves to metrics/v3, which adds `agents` per phase beside `models`, so a mixed-agent run is visible in its metrics once B19 lands. Everything already recorded must keep reading: session/v1 and v2 records count as claude, and metrics/v2 snapshots stay readable by brief close, the workspace table and the cross-brief table.

**Acceptance**

- `vloop metrics --json` (and --workspace --json, and new snapshots in .vloop/state/metrics/) is metrics/v3: every field metrics/v2 has, unchanged, plus `agents`, per phase like `models`: the sorted distinct agents that ran that phase's sessions, ["claude"] for every phase that ran, and for a phase that did not run the same empty or absent value models uses.
- schemas/metrics.v3.json validates the document with agents required; `vloop schema validate metrics/v3 <file>` works and metrics.v2.json still validates a base-written snapshot.
- session/v1 and session/v2 records count as agent claude; session/v3 records count as their agent field.
- Over a run directory the base binary wrote (session/v1 and v2 records, a metrics/v2 snapshot), `vloop metrics`, `vloop metrics --by task`, `vloop metrics --workspace` and `vloop brief close --dry-run` give the same numbers as the base, and --json differs only by schema and agents.
- Over a run the changed binary wrote, every value of models equals the base binary's for the same run.
- Violation: a session/v1 or session/v2 record, or a metrics/v2 snapshot, that vloop metrics, brief close or the workspace table no longer reads; export/v1 and `vloop metrics export` change in any way.
- docs/domain/measurement/metrics.md documents agents beside models and metrics/v3, and says agents (the CLI that ran a session) is not the agent clock (agent_ms).
- Committed tests the checks run cover agents in metrics/v3 for v1, v2 and v3 records and the reading of a metrics/v2 snapshot; an earlier test changes only where it pins metrics/v2 as the --json schema, and the work notes name that contract item.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T3/gate.sh
```

</details>

### T4 — Close the run: every gate at once and the worked example line for line

`pending` · depends on: T1, T2, T3

The earlier gates each ran while their task was the newest; this task checks the finished tree against all of them together and against the brief's worked example, and reconciles whatever drifted between tasks — docs that still name session/v1, session/v2 or metrics/v2 as what vloop writes, or a config page row out of step. It adds no behaviour of its own.

**Acceptance**

- The gates of T1, T2 and T3 all pass on the finished tree, run together.
- Every line of the brief's worked example holds against the changed binary: the config lines with their exact output and exit codes, the stub comparison with no differences for all five phases and the planted flag caught, session/v3 records with agent claude and cost_source reported (none with cost_usd 0 when total_cost_usd is absent), metrics/v3 with agents, and the base-written run directory reading as before.
- No page under docs/domain/ or docs/guide/ still says that vloop writes session/v1 or session/v2 records or that `vloop metrics --json` prints metrics/v2, except where it describes what older records and snapshots are and that they still read.
- The task adds no behaviour: any code change it makes fixes a divergence from the brief's contract that an earlier task left, and its work notes name it; such a fix ships with the committed test that would have caught it.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T4/gate.sh
```

</details>


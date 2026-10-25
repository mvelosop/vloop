# Plan — B20261025-1307-b19-copilot-review

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **3/5 done** · iteration 4

**Brief:** `docs/briefs/B20261025-1307-b19-copilot-review.loop-brief.md` · **Updated:** 2026-10-25T15:19:01Z

## Progress

- [x] **T1** — Accept copilot for agent.review only, add copilot.credit-usd, and give the run harness a stub copilot · 1 attempt(s)
- [x] **T2** — Start review sessions on copilot through a Copilot adapter with its own home, the embedded plugin, the guidelines and C2's denials
- [x] **T3** — Map Copilot's JSON Lines onto session/v3 with credits and an estimated cost, and add credits to metrics/v3
- [ ] **T4** — Check copilot in doctor and the run preflight, say it is experimental, refuse the default review model, warn of skill collisions, and document it
- [ ] **T5** — Close the run: every gate at once and the worked example line for line

## Tasks

### T1 — Accept copilot for agent.review only, add copilot.credit-usd, and give the run harness a stub copilot

`done` · 1 attempt(s) · depends on: none

Opening review sessions to GitHub Copilot CLI starts with configuration: agent.review must accept copilot while every other phase stays claude only, and the credit price that turns Copilot's credits into an estimated dollar cost needs a key. The committed run harness also needs a stub copilot, following the stub claude's pattern, so the adapter (T2), the record (T3) and doctor (T4) can be tested offline. Nothing a session receives changes in this task.

**Acceptance**

- `vloop config set agent.review copilot` and `vloop config set model.review gpt-5.6-terra` both exit 0; `vloop config get agent.review` then prints `copilot`, and with `--json` `{"key":"agent.review","value":"copilot","source":"file"}`.
- `vloop config set agent.work copilot` prints `vloop: invalid value "copilot" for agent.work: want one of claude` on stderr, exits 2 and leaves .vloop/config.toml byte-identical; the same holds for agent.plan, agent.gate-review and agent.verify, and copilot for any of them in the environment or the file is a source error (exit 1) for config get, config list and vloop run, which starts no session.
- agent.review accepts exactly claude and copilot: another value is refused by config set (exit 2, naming copilot among the valid values) and is a source error from the environment (exit 1); VLOOP_AGENT_REVIEW=copilot reads with source env.
- A new key copilot.credit-usd: default 0.0066667 (`vloop config get` prints `0.0066667`, source default, and `vloop config list` shows `copilot.credit-usd=0.0066667 (default)`), a number greater than 0 (0, negative and non-numeric values refused by config set with exit 2 and by the environment as a source error naming VLOOP_COPILOT_CREDIT_USD, exit 1), overridden by VLOOP_COPILOT_CREDIT_USD.
- cmd/vloop/run_harness_test.go gains a stub copilot beside the stub claude, following its pattern: it logs each argument on its own line, its working directory and environment, answers --version, and for `-p "/vloop:review <task>"` plays a per-test script and prints JSON Lines (assistant.turn_start, session.usage_checkpoint with data.totalNanoAiu, a result event with exitCode and usage.sessionDurationMs). The stub claude, argv.log and every existing harness test are unchanged.
- Violation: copilot accepted for agent.plan, agent.work, agent.gate-review or agent.verify; a new error message or exit code for an invalid agent value instead of the existing one.
- Committed tests the checks run cover the new values and key: agent.review copilot accepted and agent.&lt;other&gt; copilot refused (set, environment, file), copilot.credit-usd default, environment, file and refusal of non-positive values. A key shipped with no test that exercises it is a violation.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T1/gate.sh
```

</details>

### T2 — Start review sessions on copilot through a Copilot adapter with its own home, the embedded plugin, the guidelines and C2's denials

`done` · depends on: T1

With agent.review = copilot every review session must become a GitHub Copilot CLI session started by a second adapter behind B18's seam, running the plugin embedded in the binary, with a Copilot home of its own under .vloop/tmp/ (never ~/.copilot) and a deny list that stops what a review must never do outward: gh, push and remote changes, network commands, global git config, rm and the git commands the review skill forbids, web fetch, the built-in GitHub MCP server. The deny list is a denylist until B20 fences Copilot, so it must also leave the review able to read the diff and run vloop and the checks. Claude sessions must not change by a byte. Parsing Copilot's output into the record is T3's; this task may leave the record missing.

**Acceptance**

- With agent.review = copilot every review session is started as `copilot` with `-p "/vloop:review <task>"`, `--model <model.review>`, `--effort <effort.review>` only when effort.review is set, `--plugin-dir` the plugin the driver already extracts for claude sessions (the binary's embedded plugin under .vloop/tmp/plugin/&lt;version&gt;), `--no-ask-user`, `--output-format json`, `--allow-all-tools`, the C2 `--deny-tool` patterns and `--disable-builtin-mcps`; the working directory is the repository root. Every other session of the run is a claude session.
- The copilot session's environment is the one a claude review session gets — the runner's, with VLOOP_ACTIVE_TASK, VLOOP_GATE_TASK and a stale VLOOP_PLAN_SHA256 dropped and the plan hash added — plus COPILOT_HOME set to an empty directory the adapter creates under .vloop/tmp/ for that session (replacing any COPILOT_HOME the runner had) and, when .vloop/repo-guidelines.md exists and is not empty, COPILOT_CUSTOM_INSTRUCTIONS_DIRS naming a directory the adapter prepares under .vloop/tmp/ that holds the guidelines. Variables such as GH_TOKEN pass through so login survives.
- The deny patterns stop, as Copilot's Kind(argument) patterns with shell(&lt;stem&gt;:*) matching a command stem alone or followed by a space: gh in any form; git push; git remote; git config --global, --system and --file; curl, wget, ssh, scp, nc; rm; git commit, reset, checkout, switch, stash, branch and tag; and the url tool. They do not stop what a review needs: git diff, status, log and show, vloop, printf, cat, jq, go test, ls, grep, and the file tools writing .vloop/tmp/verdict.json.
- Violation: a copilot session started with --allow-all-paths, --add-dir, --sandbox, --no-sandbox, --allow-all or --yolo; a COPILOT_HOME that is ~/.copilot, outside .vloop/tmp/ or reused across sessions; an installed plugin (`copilot plugin install`) instead of --plugin-dir; any read or write of ~/.copilot.
- Violation: any change to a claude session's argv, working directory, environment or record — B18's comparison gate (run unchanged inside this task's gate) must pass, and with agent.review = copilot the plan, gate-review, work and verify sessions receive exactly what they receive with every phase on claude.
- Violation: Copilot specifics (argv, denials, COPILOT_HOME, guidelines directory) built in the driver or the runner instead of the Copilot adapter; the review path (verdict reading, proposal handling, tree revert) learning which agent ran.
- Committed tests the checks run, using the harness's stub copilot, cover: the argv of a copilot review session (with and without effort), its working directory and environment (COPILOT_HOME under .vloop/tmp/ and empty, the dropped and passed variables), the guidelines directory with and without .vloop/repo-guidelines.md, and the presence of every C2 denial with --allow-all-tools and --disable-builtin-mcps; a behaviour shipped with no such test is a violation. The argv tests of the claude adapter are unchanged.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T2/gate.sh
```

</details>

### T3 — Map Copilot's JSON Lines onto session/v3 with credits and an estimated cost, and add credits to metrics/v3

`done` · depends on: T2

Copilot reports credits, not dollars, and prints JSON Lines, not one document, so the Copilot adapter must decide whether there is a record and map the events onto session/v3, with cost_usd estimated from copilot.credit-usd so the run's cost ceiling, per-task costs and metrics still count review spend. session/v3 and metrics/v3 are unreleased, so they gain credits in place. Claude records must stay exactly as they are.

**Acceptance**

- A copilot review session has a record exactly when its stdout holds a result event; otherwise the runner logs `SESSION RECORD MISSING review <task> (iteration <n>)` as for claude and writes no record.
- The record is session/v3 with agent "copilot", the session's model, duration_ms from result.usage.sessionDurationMs, turns the number of assistant.turn_start events, is_error true exactly when result.exitCode is not 0, credits the last session.usage_checkpoint's data.totalNanoAiu / 1e9, cost_usd credits × copilot.credit-usd (from the default, the file or VLOOP_COPILOT_CREDIT_USD), cost_source "estimated" — or "none" with credits and cost_usd 0 when no checkpoint was printed —, models_used one entry for the session's model with zero tokens and the estimated cost, and permission_denials [].
- Nothing from result.usage.codeChanges reaches a record, and no absolute path does.
- session/v3 gains an optional numeric credits and cost_source "estimated"; claude records are unchanged (no credits, reported or none) and the schema still rejects a claude record with cost_source "estimated"; a base-written session/v3 record still validates.
- `vloop metrics --json` and the committed snapshot (metrics/v3, amended in place) carry credits: the run's total and per phase like cost, 0 where none; agents.review is ["copilot"] for a run with copilot reviews; cost totals, per-phase and per-task costs use cost_usd, estimated or reported, unchanged in logic; the metrics text marks an estimated total with ≈ (`cost ≈$6.91` in the brief), and a run with no estimated cost shows no ≈.
- The run's cost ceiling counts a copilot review's estimated cost exactly as it counts a claude session's reported cost.
- Violation: Copilot's JSON Lines parsed in the driver instead of the Copilot adapter; credits or an estimated cost on a claude record; codeChanges or an absolute path in a record; a new schema version instead of amending session/v3 and metrics/v3.
- docs/domain/execution/session.md's record section documents credits, the estimated cost source and agent copilot.
- Committed tests the checks run cover the parse (result present and absent, no checkpoint, several checkpoints with the last counting, exitCode not 0, codeChanges dropped), the estimated cost at a configured price, the schemas accepting and refusing as above, and credits and ≈ in metrics; a behaviour shipped with no such test is a violation.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T3/gate.sh
```

</details>

### T4 — Check copilot in doctor and the run preflight, say it is experimental, refuse the default review model, warn of skill collisions, and document it

`pending` · depends on: T2

A Copilot review session is not fenced until B20, so the operator must be told so wherever the choice is made, and the run must not start with a review model that cannot work: model.review defaults to the Claude alias sonnet, which is not a Copilot model id. Copilot also loads a repository's .claude/skills/, where a skill named like the plugin's review can be picked instead of /vloop:review, so doctor names that risk. The configuration guide documents the new value, key and status.

**Acceptance**

- When agent.review is copilot, `vloop doctor` and the run preflight check that copilot is on PATH and that `copilot --version` runs, as they do for claude: a missing or broken copilot is a problem (doctor exits 1; vloop run exits 1, says why and starts no session).
- When agent.review is copilot, `vloop doctor` prints `! agent.review is copilot: experimental — review sessions are not fenced until B20 (docs/guide/configuration.md)` and the run preflight prints the same `! …` line; it is a warning, not a problem: doctor exits 0 and the run goes on.
- When agent.review is copilot and model.review comes from the default, `vloop run` prints `vloop: agent.review is copilot but model.review is the default "sonnet", a Claude alias: set model.review to a Copilot model id`, exits 1 and starts no session; a model.review set in the file or the environment is not refused.
- When agent.review is copilot, `vloop doctor` warns once for each directory in .claude/skills/ whose name ends in `review`, as `! .claude/skills/<name> may be loaded instead of /vloop:review by a copilot session` (exit 0); a skill whose name does not end in review gets no warning.
- With every agent claude, doctor and the run need no copilot on PATH, print nothing about copilot (no skill-collision warning either) and behave exactly as before.
- docs/guide/configuration.md lists `copilot` among agent.review's valid values (and only claude for the other agent keys), has a row for copilot.credit-usd with default `0.0066667` and variable `VLOOP_COPILOT_CREDIT_USD`, and says that copilot review sessions are experimental until B20 and what C2's deny list stops (gh, push and remote changes, network commands, global git config, rm and forbidden git commands, web fetch, built-in MCP servers) and does not (anything not named, and other forms of the named commands).
- Committed tests the checks run cover the copilot preflight (found, missing, --version failing), the experimental warning in doctor and the run, the default-model refusal (and its absence for a set model), the skill-collision warning, and their absence with every agent claude; a behaviour shipped with no such test is a violation.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T4/gate.sh
```

</details>

### T5 — Close the run: every gate at once and the worked example line for line

`pending` · depends on: T1, T2, T3, T4

Each earlier gate ran while its task was the newest; this task checks the finished tree against all of them together, including B18's comparison gate, and against the brief's worked example, and reconciles whatever drifted between tasks — a doc that still says claude is the only agent or that a record has no credits, a config row out of step. It adds no behaviour of its own; C5, the proof on the real Copilot CLI, is the operator's at close.

**Acceptance**

- The gates of T1, T2 (with B18's comparison gate inside it), T3 and T4 all pass on the finished tree, run together.
- Every line of the brief's worked example holds against the changed binary: the config lines with their output and exit codes, the default-model refusal, doctor's experimental warning, a one-task run whose review is a copilot session with every C2 denial and --disable-builtin-mcps, COPILOT_HOME under .vloop/tmp/ and the repository root as working directory, its record (agent copilot, credits 14.2, cost_usd 14.2 × the default price, cost_source estimated, turns 3, duration_ms 18000, permission_denials []) while plan, work and verify run on claude, SESSION RECORD MISSING for a stub that prints no result event, and metrics with credits total 14.2 and agents review ["copilot"].
- No page under docs/domain/ or docs/guide/ still says claude is the only agent a review may run on, or that a session record never carries credits, except where it describes older records.
- The task adds no behaviour: any code change it makes fixes a divergence from the brief's contract that an earlier task left, its work notes name it, and it ships with the committed test that would have caught it.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T5/gate.sh
```

</details>


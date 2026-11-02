# Plan — B20261101-2305-b27-i77-copilot-plan

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** complete · **6/6 done** · iteration 7

**Brief:** `docs/briefs/B20261101-2305-b27-i77-copilot-plan.loop-brief.md` · **Updated:** 2026-11-02T00:28:28Z

## Progress

- [x] **T1** — Add the plan phase to the fence judge, vloop fence check and vloop fence hook · 1 attempt(s)
- [x] **T2** — Run plan sessions on Copilot: the adapter, its record, agent.plan = copilot and the exit line
- [x] **T3** — Check agent.plan = copilot in the run preflight and doctor
- [x] **T4** — Document Copilot plan sessions in the fence audit, the configuration guide and the domain pages
- [x] **T5** — Extend the Copilot acceptance kit with a plan run on the real CLI
- [x] **T6** — Close the run: every gate and the worked example on the final tree

## Tasks

### T1 — Add the plan phase to the fence judge, vloop fence check and vloop fence hook

`done` · 1 attempt(s) · depends on: none

A Copilot plan session needs a hook that judges each tool call against the plan fence, and the hook and `vloop fence check` share one judge. This task gives the judge a `plan` phase that decides exactly as review does, with the plan fence's rules, the plan's own write rule (the plan file, its gate folders and .vloop/tmp/), and the phase's own name and skill hint, and makes both commands accept it. Everything T2 builds runs under this phase, so the other phases must not move by a word.

**Acceptance**

- `vloop fence check --phase plan --root <dir> --tool <name> [--path <p>] ["<command>"]` gives every decision of the brief's two worked-example tables (with `--root .`) and its two apply_patch lines, with exactly the reasons shown there (`deny: ` followed by the message), exit 0 for allow and 1 for deny.
- In plan a call is decided exactly as in review, with four differences only: shell denials by a fence rule read `<command> is denied by the plan fence (Bash(<rule>))`; a write is allowed when its path, resolved against `--root`, is `.vloop/state/state.json` or lies under `.vloop/state/gates/` or `.vloop/tmp/` (the two directories themselves are not writable targets; `/dev/null` stays allowed as in review), any other write being denied with `write to <path> is not allowed in a plan session: only .vloop/state/state.json, .vloop/state/gates/ and .vloop/tmp/ may be written`; the tool denial is `tool <name> is not allowed in a plan session`; the skill hint is `skill <name> is not vloop's — use /vloop:plan`. Deny-list denials keep `<command> is denied by the Copilot deny list (<stem>)`, `rm:*` included; `sh <script>` stays denied. The rules are read from the embedded `fence/plan.json`; reading `fence/review.json` for this phase, or a second hand-copied list or path list, is a violation.
- `vloop fence hook --phase plan --root <dir>` answers each Copilot payload with the same decision, the reason being `vloop fence: ` and the message, exit 0. An unknown phase is, for `fence check`, exit 2 with `vloop: unknown phase "<p>": want review, second-review, work, verify, gate-review or plan` on stderr, and for `fence hook` the deny line holding that message and exit 2.
- `vloop fence check --help` and `vloop fence hook --help` name `plan` among the phases in their usage line, their `--phase` flag and their long help; `docs/guide/commands.md` is regenerated with `go generate ./...` in this task, so `TestCommandsReferenceIsCurrent` stays green in the `go` check.
- The review, second-review, work, verify and gate-review phases decide and word every call exactly as before, bar the unknown-phase message (the brief's unchanged lines: gate-review denies a write to `.vloop/state/state.json` with its own message; review's skill hint for `vloop-plan` is `/vloop:review`; work allows `rm src/old.go`). Existing tests that pin `plan` as an unknown phase or the old unknown-phase message (`TestUnknownPhase`, `TestFenceHookFailsClosed`'s `plan phase` case) change only to the new behaviour — the unknown name moves to another — and the work notes name the contract item (P1).
- The plan phase is covered by committed Go table tests in `internal/fence` (and the hook's in its own test) that run on every OS in the `go` check, covering every line of the worked example's two tables, the apply_patch lines, the write rule's edges (the two directories themselves, a path after `..`, an absolute path under the root) and the unknown-phase message; a plan phase that leaves no committed test behind is a violation.
- `fence/*.json` is unchanged, and no Claude session's argv, fence or settings changes.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T1/gate.sh
```

</details>

### T2 — Run plan sessions on Copilot: the adapter, its record, agent.plan = copilot and the exit line

`done` · depends on: T1

With the judge's plan phase in place (T1), the Copilot adapter can run the plan session exactly as it runs a gate-review session, under a hooks file whose phase is plan. This task makes the adapter support the phase, gives the session its session/v4 record, lets `agent.plan` take `copilot` and makes the driver's failure line name the agent; everything else the driver does with a plan must stay the same whatever the agent. The preflight and doctor checks are T3's.

**Acceptance**

- `agent.plan` accepts `claude` (default) and `copilot`, from the file and from `VLOOP_AGENT_PLAN`, and refuses any other value with exit 2.
- With `agent.plan = copilot`, each plan session — the first and the revision a failed gate-review round asks for — runs `copilot -p "/vloop:plan <brief path>"` with the model and effort P-6 resolves, the plugin by `--plugin-dir`, a fresh home under `.vloop/tmp/` holding a hooks file that runs `fence hook --phase plan` in both command forms, the guidelines as custom instructions, and the same `--deny-tool` list as a Copilot review session — `--deny-tool=shell(rm:*)` and `--deny-tool=shell(git push:*)` included. A plan session started without the hooks file, with another phase in it, or without `shell(rm:*)` is a violation.
- A Copilot plan session's JSON Lines become a `session/v4` record with phase `plan`, agent `copilot`, the model, credits, estimated cost, `cost_source` and `permission_denials`; it validates as session/v4. A Claude plan record keeps its schema as at the base.
- Everything around the plan session is unchanged whatever the agent: `planning from <brief> using <model>`; a moved ref halts with exit 9 (`vloop: REFS MOVED plan — the planning session changed git refs; nothing was committed`); a timeout halts with exit 7; no plan halts with exit 1 (`vloop: planning produced no .vloop/state/state.json`), as do a plan failing `state/v2` or the gate-shape checks; the base gates, the gate review and its two rounds; the plan commit. The one change: a plan session that exits non-zero halts with `vloop: planning session failed (<agent> exited <n>) — see <run folder>`, `copilot` for a Copilot session and `claude`, as at the base, for a Claude one. Planner code that otherwise branches on the agent, or a fallback to Claude when a Copilot plan session fails, is a violation.
- A Copilot work, review, second-review, verify or gate-review session's argv and hooks file are unchanged (work without `shell(rm:*)` and `--phase work`; the others with it and their own phase); a Claude session's argv, fence and settings are unchanged — a Claude plan session still runs under the plan fence.
- Committed Go tests cover the adapter's plan argv (the whole deny list, `shell(rm:*)` included), its hooks file's `--phase plan`, `Supports(plan)`, the plan record's schema, `agent.plan` accepting copilot and the exit line for each agent; tests that pinned plan as unsupported or `agent.plan` as claude-only (`TestCopilotSupportsWorkReviewAndVerify` among them) or the old exit line change only to the new behaviour, and the work notes name the contract item (P2).

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T2/gate.sh
```

</details>

### T3 — Check agent.plan = copilot in the run preflight and doctor

`done` · depends on: T2

An operator who puts the plan session on Copilot must be warned it is experimental and stopped before any session when the setup cannot work, as for the other Copilot phases. This task adds agent.plan to the keys the preflight and doctor check; planning has no switch, so the key is in use whenever it is copilot, also on a resume that will not plan, as B26 decided for the gate review.

**Acceptance**

- With `agent.plan = copilot`: `vloop run` prints `  ! agent.plan is copilot: experimental (docs/guide/configuration.md)` and goes ahead; `vloop doctor` prints `! agent.plan is copilot: experimental (docs/guide/configuration.md)` and the copilot and copilot login rows. This holds also on a resume that will not plan (after `vloop run --plan-only`).
- `vloop run` with `agent.plan = copilot` (from the file or `VLOOP_AGENT_PLAN`) and `model.plan` at its default refuses before any session, exit 1: `vloop: agent.plan is copilot but model.plan is the default "opus", a Claude alias: set model.plan to a Copilot model id` — on a resume too.
- With `agent.plan = copilot`, a project skill directory under `.claude/skills/` or `.github/skills/` named like one of vloop's skills stops `vloop run` with K1's message (`vloop: .claude/skills/plan has the name of vloop's plan skill; a copilot session may load it instead — rename it`), exit 1, no session, and is an error row in doctor.
- With `agent.plan = copilot`, no GitHub login reaching the sessions refuses the run naming `agent.plan` (`vloop: agent.plan is copilot but no GitHub login reaches its sessions: …`), exit 1, and is doctor's copilot login error row; copilot missing or `copilot --version` failing is doctor's copilot error row and stops the run.
- With `agent.plan = claude` none of this appears, and a project skill named `plan` stops nothing. The checks for agent.work, agent.review, agent.second-review, agent.verify and agent.gate-review keep their words and behaviour.
- Committed tests cover the warning, the model.plan refusal, K1, the login refusal, the resume and the claude case for agent.plan (extending the existing doctor and run preflight tests); a code comment that lists the Copilot phases or keys (doctor's experimental warning and copilotKeys included) names plan too.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T3/gate.sh
```

</details>

### T4 — Document Copilot plan sessions in the fence audit, the configuration guide and the domain pages

`done` · depends on: T1, T2, T3

Operators and later briefs learn what a Copilot plan session may do from the fence audit, the configuration guide and the domain pages, which today keep the plan on Claude. This task brings every surface that names the Copilot phases up to what T1–T3 built, in the code's own words. It is prose: if it fails review twice on wording, it is blocked and the operator finishes it.

**Acceptance**

- `docs/design-notes/fence-audit.md` gains a part for a Copilot plan session, under a heading naming the plan session or the plan fence, that classifies every deny rule of `fence/plan.json`, its allowed writes (`.vloop/state/state.json`, `.vloop/state/gates/`, `.vloop/tmp/`) and every deny-list pattern (the whole review list, `shell(rm:*)` included) by who enforces it in a Copilot plan session — the driver after the session, the hook, or neither — and says, for each `neither`, what covers it. It must not claim the driver reverts a file a plan session changed outside the files it may write: the plan commit stages the whole tree, so the hook is the line (out of scope to change).
- `docs/guide/configuration.md`: the `agent.plan` row offers `claude` and `copilot` and says `copilot` is experimental; the Copilot paragraph covers plan sessions, the `agent.plan` warning and the `model.plan` refusal (default `"opus"`), and no longer says `agent.plan` stays on `claude`; the fence hook section's `--phase` lists name plan, it says a plan session is judged against the plan fence and may write only `.vloop/state/state.json`, under `.vloop/state/gates/` and under `.vloop/tmp/`, and K1's hint names `/vloop:plan` in a plan session.
- `docs/domain/execution/session.md` says copilot runs every session kind, plan included; its Copilot hook paragraph says the hook serves the plan phase with the plan fence and allows writes only to the plan, its gate folders and `.vloop/tmp/` in it. The glossary's hook entry in `docs/domain/domain-model.md` covers the plan phase.
- Every other surface that names the Copilot phases names plan too, the code comments that list them (the adapter's `AgentCopilot` comment, the hook's, doctor's experimental warning) included; a surface still listing only work, review, second-review, verify and gate-review is a violation.
- Every sentence about behaviour matches what T1–T3 built (messages quoted as the code prints them); `fence/*.json` is unchanged; the `go` check still passes.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T4/gate.sh
```

</details>

### T5 — Extend the Copilot acceptance kit with a plan run on the real CLI

`done` · depends on: T2, T3

Gates use a stub copilot; only a run on the real CLI proves a Copilot plan session works and leaves a plan the driver accepts. This task gives the operator a script for that acceptance at close and documents it; no gate or check runs it.

**Acceptance**

- `tools/copilot-acceptance/` gains a script that runs `greet.loop-brief.md` in a scratch repository with a vloop built from this tree, `agent.plan = copilot` and `model.plan` the model given, gate review, work, review and verify on Claude, and writes a summary of each plan session's record (agent, model, credits, estimated cost, `session/v4` validity, its `permission_denials`), of the plan (its tasks, `state/v2` validity) and of the gate review's verdict per round.
- The README gains its section, under a heading naming plan sessions: how to run the script, what it needs (copilot, a logged-in gh, claude), what it spends, what the summary holds, and that the outcome is recorded in `docs/design-notes/copilot-cli-experiments.md` before close.
- The kit's scripts pass `sh -n` and its Python parses; flags, keys and messages of existing vloop commands the script uses are taken from running them, not memory; no gate or check runs the kit, copilot, gh or the network.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T5/gate.sh
```

</details>

### T6 — Close the run: every gate and the worked example on the final tree

`done` · depends on: T1, T2, T3, T4, T5

The run is done only when every gate holds together on the final tree and the brief's worked example holds line for line. This task changes nothing unless a gate or a line of the example fails, in which case it fixes the cause within the brief's scope.

**Acceptance**

- Every gate T1–T5 passes on the final tree.
- Every line of the brief's worked example holds against the built binary: the two `fence check --phase plan` tables and the apply_patch lines, the three unchanged lines, the changed unknown-phase line, the stub run's line, record, argv and hooks file, and each planted failure (model.plan unset, `.claude/skills/plan`, exit 3 for each agent, no plan, moved HEAD).
- The `go` check passes on the final tree.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T6/gate.sh
```

</details>


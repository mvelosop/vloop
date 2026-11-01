# Plan — B20261101-2044-b26-i78-copilot-gate-review

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **5/6 done** · iteration 6

**Brief:** `docs/briefs/B20261101-2044-b26-i78-copilot-gate-review.loop-brief.md` · **Updated:** 2026-11-01T21:51:09Z

## Progress

- [x] **T1** — Add the gate-review phase to the fence judge, vloop fence check and vloop fence hook · 1 attempt(s)
- [x] **T2** — Run gate-review sessions on Copilot: the adapter, its record and agent.gate-review = copilot
- [x] **T3** — Check agent.gate-review = copilot in the run preflight and doctor
- [x] **T4** — Document Copilot gate-review sessions in the fence audit, the configuration guide and the domain pages
- [x] **T5** — Extend the Copilot acceptance kit with a gate-review run on the real CLI
- [ ] **T6** — Close the run: every gate at once, and the worked example line for line

## Tasks

### T1 — Add the gate-review phase to the fence judge, vloop fence check and vloop fence hook

`done` · 1 attempt(s) · depends on: none

A Copilot gate-review session needs a hook that judges each tool call against the gate-review fence, and the hook and `vloop fence check` share one judge. This task gives the judge a `gate-review` phase that decides exactly as review does, with the gate-review fence's rules and the phase's own name and skill hint, and makes both commands accept it. Everything T2 builds runs under this phase, so the other phases must not move by a word.

**Acceptance**

- `vloop fence check --phase gate-review --root <dir> --tool <name> [--path <p>] ["<command>"]` gives every decision of the brief's two worked-example tables and its two apply_patch lines, with exactly the reasons shown there (`deny: ` followed by the message), exit 0 for allow and 1 for deny.
- In gate-review a call is decided exactly as in review, with three differences only: shell denials by a fence rule read `<command> is denied by the gate-review fence (Bash(<rule>))`; write and tool denials name the phase (`write to <path> is not allowed in a gate-review session: only .vloop/tmp/ may be written`, `tool <name> is not allowed in a gate-review session`); the skill hint is `skill <name> is not vloop's — use /vloop:gate-review`. Deny-list denials keep `<command> is denied by the Copilot deny list (<stem>)`, `rm:*` included. The rules are read from the embedded `fence/gate-review.json`; reading `fence/review.json` for this phase, or a second hand-copied list, is a violation.
- `vloop fence hook --phase gate-review --root <dir>` answers each Copilot payload with the same decision, the reason being `vloop fence: ` and the message, exit 0. An unknown phase (`plan`, anything else) is, for `fence check`, exit 2 with `vloop: unknown phase "<p>": want review, second-review, work, verify or gate-review` on stderr, and for `fence hook` the deny line holding that message and exit 2.
- `vloop fence check --help` and `vloop fence hook --help` name `gate-review` among the phases in their usage line, their `--phase` flag and their long help; `docs/guide/commands.md` is regenerated with `go generate ./...` in this task, so `TestCommandsReferenceIsCurrent` stays green in the `go` check.
- The review, second-review, work and verify phases decide and word every call exactly as before, bar the unknown-phase message (the unchanged lines: `--phase review --tool skill --path vloop-gate-review` is `deny: skill vloop-gate-review is not vloop's — use /vloop:review`; `--phase verify` denies `rm x` with `rm x is denied by the Copilot deny list (rm:*)`; `--phase work` allows `rm src/old.go`; `vloop config set agent.plan copilot` exits 2 with `vloop: invalid value "copilot" for agent.plan: want one of claude`). Existing tests that pin `gate-review` as an unknown phase or the old unknown-phase message change only to the new behaviour, and the work notes name the contract item (G1).
- The gate-review phase is covered by committed Go table tests in `internal/fence` (and the hook's in its own test) that run on every OS in the `go` check, covering every line of the worked example's two tables, the apply_patch lines and the unknown-phase message; a gate-review phase that leaves no committed test behind is a violation.
- `fence/*.json` is unchanged, and no Claude session's argv, fence or settings changes.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T1/gate.sh
```

</details>

### T2 — Run gate-review sessions on Copilot: the adapter, its record and agent.gate-review = copilot

`done` · depends on: T1

With the judge's gate-review phase in place (T1), the Copilot adapter can run the gate-review session exactly as it runs a verify session, under a hooks file whose phase is gate-review. This task makes the adapter support the phase, gives the session its session/v4 record and lets `agent.gate-review` take `copilot`; it does not touch what the driver does with the verdict, which must stay the same whatever the agent. The preflight and doctor checks are T3's.

**Acceptance**

- `agent.gate-review` accepts `claude` (default) and `copilot`, from the file and from `VLOOP_AGENT_GATE_REVIEW`, and refuses any other value with exit 2; `agent.plan` still refuses `copilot` (`vloop: invalid value "copilot" for agent.plan: want one of claude`, exit 2; a run with `VLOOP_AGENT_PLAN=copilot` starts no session).
- With `agent.gate-review = copilot`, each gate-review round runs `copilot -p "/vloop:gate-review"` with the model and effort P-6 resolves, the plugin by `--plugin-dir`, a fresh home under `.vloop/tmp/` holding a hooks file that runs `fence hook --phase gate-review` in both command forms, the guidelines as custom instructions, and the same `--deny-tool` list as a Copilot review session — `--deny-tool=shell(rm:*)` and `--deny-tool=shell(git push:*)` included. A gate-review session started without the hooks file, with another phase in it, or without `shell(rm:*)` is a violation.
- A Copilot gate-review session's JSON Lines become a `session/v4` record with phase `gate-review`, agent `copilot`, the model, credits, estimated cost, `cost_source` and `permission_denials`; it validates as session/v4. A Claude gate-review record keeps its schema as at the base.
- Everything around the gate-review session is unchanged whatever the agent: `reviewing the gates using <model> (round <n>)`; a moved ref halts with exit 9 (`REFS MOVED gate-review — the gate review changed git refs; nothing was committed`); a timeout halts with exit 7; a missing or invalid `.vloop/tmp/gate-verdict.json` is a `FAIL` (kept as the driver's FAIL with its note); the verdict is kept as `reports/gate-review-<round>.json`; a `FAIL` goes back to the planner once, and a plan failed twice is saved blocked with exit 2; a `check-gap` finding halts with exit 2 and keeps nothing. Code in `gatereview.go` or the planner that branches on the agent is a violation.
- A Copilot work, review, second-review or verify session's argv and hooks file are unchanged (work without `shell(rm:*)` and `--phase work`; review, second-review and verify with it and their own phase); a Claude session's argv, fence and settings are unchanged — a Claude gate-review session still runs under the gate-review fence.
- Committed Go tests cover the adapter's gate-review argv (the whole deny list, `shell(rm:*)` included), its hooks file's `--phase gate-review`, `Supports(gate-review)` (with `Supports(plan)` still false), the gate-review record's schema and `agent.gate-review` accepting copilot while `agent.plan` does not; tests that pinned gate-review as unsupported or `agent.gate-review` as claude-only (`TestCopilotSupportsWorkReviewAndVerify` among them) change only to the new behaviour, and the work notes name the contract item (G2).

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T2/gate.sh
```

</details>

### T3 — Check agent.gate-review = copilot in the run preflight and doctor

`done` · depends on: T2

An operator who puts the gate review on Copilot must be warned it is experimental and stopped before any session when the setup cannot work, as for the other Copilot phases. This task adds agent.gate-review to the keys the preflight and doctor check; the gate review has no switch, so the key is in use whenever it is copilot, also on a resume past the gate review.

**Acceptance**

- With `agent.gate-review = copilot`: `vloop run` prints `  ! agent.gate-review is copilot: experimental (docs/guide/configuration.md)` and goes ahead; `vloop doctor` prints `! agent.gate-review is copilot: experimental (docs/guide/configuration.md)` and the copilot and copilot login rows. This holds whatever `run.verify` says.
- `vloop run` with `agent.gate-review = copilot` (from the file or `VLOOP_AGENT_GATE_REVIEW`) and `model.gate-review` at its default refuses before any session, exit 1: `vloop: agent.gate-review is copilot but model.gate-review is the default "sonnet", a Claude alias: set model.gate-review to a Copilot model id`.
- With `agent.gate-review = copilot`, a project skill directory under `.claude/skills/` or `.github/skills/` named like one of vloop's skills stops `vloop run` with K1's message (`vloop: .claude/skills/gate-review has the name of vloop's gate-review skill; a copilot session may load it instead — rename it`), exit 1, no session, and is an error row in doctor. No new warning is added for look-alike gate-review names; the `review` warning keeps its condition and message (`may be loaded instead of /vloop:review by a copilot session`), and so now shows with only the gate review on copilot.
- With `agent.gate-review = copilot`, no GitHub login reaching the sessions refuses the run naming `agent.gate-review` (`vloop: agent.gate-review is copilot but no GitHub login reaches its sessions: …`), exit 1, and is doctor's copilot login error row; copilot missing or `copilot --version` failing is doctor's copilot error row and stops the run.
- With `agent.gate-review = claude` none of this appears, and a project skill named `gate-review` stops nothing. The checks for agent.work, agent.review, agent.second-review and agent.verify keep their words and behaviour.
- Committed tests cover the warning, the model.gate-review refusal, K1, the login refusal and the claude case for agent.gate-review (extending the existing doctor and run preflight tests); a code comment that lists the Copilot phases or keys (doctor's experimental warning and copilotKeys included) names gate-review too.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T3/gate.sh
```

</details>

### T4 — Document Copilot gate-review sessions in the fence audit, the configuration guide and the domain pages

`done` · depends on: T1, T2, T3

Operators and later briefs learn what a Copilot gate-review session may do from the fence audit, the configuration guide and the domain pages, which today keep the gate review on Claude. This task brings every surface that names the Copilot phases up to what T1–T3 built, in the code's own words. It is prose: if it fails review twice on wording, it is blocked and the operator finishes it.

**Acceptance**

- `docs/design-notes/fence-audit.md` gains a part for a Copilot gate-review session, under a heading naming gate review, that classifies every deny rule of `fence/gate-review.json` and every deny-list pattern (the whole review list, `shell(rm:*)` included) by who enforces it in a Copilot gate-review session — the driver after the session, the hook, or neither — and says, for each `neither`, what covers it. It must not claim the driver reverts a file a gate-review session changed outside `.vloop/tmp/`: the plan commit stages the whole tree, so the hook is the line (out of scope to change).
- `docs/guide/configuration.md`: the `agent.gate-review` row offers `claude` and `copilot` and says `copilot` is experimental; `agent.plan` still offers `claude` alone; the Copilot paragraph covers gate-review sessions, the `agent.gate-review` warning and the `model.gate-review` refusal, and no longer says `agent.gate-review` stays on `claude`; the fence hook section's `--phase` lists name gate-review, it says a gate-review session is judged against the gate-review fence and may write only under `.vloop/tmp/`, and K1's hint names `/vloop:gate-review` in a gate-review session.
- `docs/domain/execution/session.md` says copilot runs work, review, second-review, verify and gate-review sessions; its Copilot hook paragraph says the hook serves the gate-review phase with the gate-review fence and allows writes only under `.vloop/tmp/` in it. The glossary's hook entry in `docs/domain/domain-model.md` covers the gate-review phase.
- Every other surface that names the Copilot phases names gate-review too, the code comments that list them (the adapter's `AgentCopilot` comment, the hook's, doctor's experimental warning) included; a surface still listing only work, review, second-review and verify is a violation.
- Every sentence about behaviour matches what T1–T3 built (messages quoted as the code prints them); `fence/*.json` is unchanged; the `go` check still passes.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T4/gate.sh
```

</details>

### T5 — Extend the Copilot acceptance kit with a gate-review run on the real CLI

`done` · depends on: T2, T3

Gates use a stub copilot; only a run on the real CLI proves a Copilot gate review works and leaves a valid verdict. This task gives the operator a script for that acceptance at close and documents it; no gate or check runs it.

**Acceptance**

- `tools/copilot-acceptance/` gains a script that runs `greet.loop-brief.md` in a scratch repository with a vloop built from this tree, `agent.gate-review = copilot` and `model.gate-review` the model given, plan, work, review and verify on Claude, and writes a summary of each gate-review session's record (round, agent, model, credits, estimated cost, `session/v4` validity, its `permission_denials`) and of each kept verdict (round, present or not, `gate-verdict/v2` validity, the verdict, its number of findings by kind).
- The README gains its section, under a heading naming gate review: how to run the script, what it needs (copilot, a logged-in gh, claude), what it spends, what the summary holds, and that the outcome is recorded in `docs/design-notes/copilot-cli-experiments.md` before close.
- The kit's scripts pass `sh -n` and its Python parses; flags, keys and messages of existing vloop commands the script uses are taken from running them, not memory; no gate or check runs the kit, copilot, gh or the network.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T5/gate.sh
```

</details>

### T6 — Close the run: every gate at once, and the worked example line for line

`pending` · depends on: T1, T2, T3, T4, T5

The run is done only when every gate holds together on the final tree and the brief's worked example holds line for line. This task changes nothing unless a gate or a line of the example fails, in which case it fixes the cause within the brief's scope.

**Acceptance**

- Every gate T1–T5 passes on the final tree.
- Every line of the brief's worked example holds against the built binary: the two `fence check --phase gate-review` tables and the apply_patch lines, the four unchanged lines, the changed unknown-phase line, the stub run's warning, round line, record, argv, hooks file and kept verdict, and each planted failure.
- The `go` check passes on the final tree.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T6/gate.sh
```

</details>


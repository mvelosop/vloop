# Plan — B20261031-1151-b25-i79-copilot-verify

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **0/6 done** · iteration 0

**Brief:** `docs/briefs/B20261031-1151-b25-i79-copilot-verify.loop-brief.md` · **Updated:** 2026-10-31T13:21:38Z

## Progress

- [ ] **T1** — Add the verify phase to the fence judge and to vloop fence check
- [ ] **T2** — Run verify sessions on Copilot: the hook's verify phase, the adapter and agent.verify = copilot
- [ ] **T3** — Check agent.verify = copilot in the run preflight and doctor
- [ ] **T4** — Document Copilot verify sessions in the fence audit, the configuration guide and the domain pages
- [ ] **T5** — Extend the Copilot acceptance kit with a verify run on the real CLI
- [ ] **T6** — Close the run: every gate at once, and the worked example line for line

## Tasks

### T1 — Add the verify phase to the fence judge and to vloop fence check

`pending` · depends on: none

V1's core. The judge behind `vloop fence check` and the Copilot hook knows review, second-review and work, so a Copilot verify session cannot be fenced. This task gives it a `verify` phase decided exactly as review — writes only under `.vloop/tmp/`, the adapter's whole deny list with `rm:*` — but reading `fence/verify.json`'s rules and naming verify in its messages, so that a later change to either fence moves only its own phase. T2 puts it behind the hook; the other phases must not change by a byte, bar the unknown-phase message.

**Acceptance**

- `vloop fence check --phase verify --root <dir> --tool <name> [--path <p>] ["<command>"]` gives every decision of the brief's two worked-example tables and its two apply_patch lines, with exactly the reasons shown there (`deny: ` followed by the message), exit 0 for allow and 1 for deny.
- In verify a call is decided exactly as in review, with three differences only: shell denials by a fence rule read `<command> is denied by the verify fence (Bash(<rule>))`; write and tool denials name the phase `verify` (`write to <path> is not allowed in a verify session: only .vloop/tmp/ may be written`, `tool <name> is not allowed in a verify session`); the skill hint is `skill <name> is not vloop's — use /vloop:verify`. Deny-list denials keep `<command> is denied by the Copilot deny list (<stem>)` with `rm:*` included. The rules are read from the embedded `fence/verify.json` (not from `fence/review.json`, and not from a second hand-copied list).
- An unknown phase (`plan`, `gate-review`, anything else) exits 2 with `vloop: unknown phase "<p>": want review, second-review, work or verify` on stderr.
- `vloop fence check --help` and `vloop fence hook --help` name `verify` among the phases in their usage line, their `--phase` flag and their long help; `docs/guide/commands.md` is regenerated with `go generate ./...` in this task, so `TestCommandsReferenceIsCurrent` stays green in the `go` check.
- The review, second-review and work phases decide and word every call exactly as before (the unchanged lines: `--phase review --tool skill --path vloop-verify` is `deny: skill vloop-verify is not vloop's — use /vloop:review`; `--phase work` allows `rm src/old.go`; `vloop config set agent.gate-review copilot` exits 2 with `vloop: invalid value "copilot" for agent.gate-review: want one of claude`). Existing tests that pin `verify` as an unknown phase or the old unknown-phase message change only to the new behaviour, and the work notes name the contract item (V1).
- The verify phase is covered by committed Go table tests in `internal/fence` that run on every OS in the `go` check, covering every line of the worked example's two tables, the apply_patch lines and the unknown-phase message; a verify phase with no committed tests is a violation.
- `fence/verify.json`, `fence/review.json` and `fence/work.json` are unchanged, and no Claude session's argv, fence or settings changes.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T1/gate.sh
```

</details>

### T2 — Run verify sessions on Copilot: the hook's verify phase, the adapter and agent.verify = copilot

`pending` · depends on: T1

V1's hook half and V2's core. With T1's judge, `vloop fence hook --phase verify` can fence a Copilot verify session; this task lets the Copilot adapter run `/vloop:verify` exactly as it runs a review — same plugin, fresh home, guidelines, the whole deny list on the argv, JSON Lines parsed into a session/v4 record — with one difference: its hooks file runs the hook with `--phase verify`. `agent.verify` then accepts `copilot`. Everything the driver does around the verify session stays as it is for any agent: it is a report, not a gate, so a finding, a failed session or a missing report never halts the run.

**Acceptance**

- `vloop fence hook --phase verify --root <dir>` answers each Copilot payload with T1's verify decision, the reason being `vloop fence: ` and T1's message, exit 0; an unknown phase is the deny line with `unknown phase "<p>": want review, second-review, work or verify` and exit 2; the review, second-review and work phases answer exactly as before.
- `agent.verify` accepts `claude` (default) and `copilot`, from the file and from `VLOOP_AGENT_VERIFY`; `agent.plan` and `agent.gate-review` still refuse `copilot` (config set exits 2 with `want one of claude`; a run with their variable set to copilot starts no session).
- With `agent.verify = copilot`, the verify session runs `copilot -p "/vloop:verify"` with the model and effort P-6 resolves, the plugin by `--plugin-dir`, a fresh home under `.vloop/tmp/` holding a hooks file that runs `fence hook --phase verify` in both command forms, the guidelines as custom instructions, and the same `--deny-tool` list as a Copilot review session — `--deny-tool=shell(rm:*)` and `--deny-tool=shell(git push:*)` included.
- A Copilot verify session's JSON Lines become a `session/v4` record (schema `session/v4`, as a Copilot work record is — at the base a Copilot review record is `session/v3`, and that stays) with phase `verify`, agent `copilot`, the model, credits, estimated cost, `cost_source` and `permission_denials`; it validates as session/v4.
- Everything around the verify session is unchanged whatever the agent: it runs only when `run.verify` is not `false`; `verify: no findings` or `verify: <n> finding(s) — see …reports/verify.json`; a valid `.vloop/tmp/verify.json` is kept as the run folder's `reports/verify.json`; a changed file is reverted with `vloop: the verify session changed <path> — reverted` and the report still kept; a moved ref halts with exit 9 and `REFS MOVED verify — a verify session changed git refs; nothing was committed`; a finding, a failed session or a missing report is a warning and the run exits 0. A Copilot verify session that changes the run's ending is a violation.
- A Copilot work, review or second-review session's argv and hooks file are unchanged (work without `shell(rm:*)` and `--phase work`; review with it and `--phase review`); a Claude session's argv, fence and settings are unchanged — a Claude verify session still runs under the verify fence.
- Committed Go tests cover the adapter's verify-phase argv (the whole deny list, `shell(rm:*)` included), its hooks file's `--phase verify`, `Supports(verify)`, the verify record's schema, the hook's verify phase and `agent.verify` accepting copilot while `agent.plan` and `agent.gate-review` do not; tests that pinned verify as unsupported or `agent.verify` as claude-only change only to the new behaviour, and the work notes name the contract item (V1, V2).

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T2/gate.sh
```

</details>

### T3 — Check agent.verify = copilot in the run preflight and doctor

`pending` · depends on: T2

V2's guard rails. An operator who sets `agent.verify = copilot` gets what the other Copilot phases already give: the experimental warning, a refusal while `model.verify` is still the Claude default, and the copilot checks — copilot on PATH and runnable, a GitHub login reaching its sessions, K1's refusal of a project skill named like one of vloop's — in `vloop run` and as `vloop doctor` rows. None of it applies when `run.verify = false`, since no verify session will run, just as the second review's key adds nothing while `run.second-review` is off.

**Acceptance**

- With `agent.verify = copilot` and `run.verify` not `false`: `vloop run` prints `  ! agent.verify is copilot: experimental (docs/guide/configuration.md)` and goes ahead; `vloop doctor` prints `! agent.verify is copilot: experimental (docs/guide/configuration.md)` and the copilot and copilot login rows.
- `vloop run` with `agent.verify = copilot` (from the file or `VLOOP_AGENT_VERIFY`) and `model.verify` at its default refuses before any session, exit 1: `vloop: agent.verify is copilot but model.verify is the default "sonnet", a Claude alias: set model.verify to a Copilot model id`.
- With `agent.verify = copilot`, a project skill directory under `.claude/skills/` or `.github/skills/` named like one of vloop's skills stops `vloop run` with K1's message (`vloop: .claude/skills/verify has the name of vloop's verify skill; a copilot session may load it instead — rename it`), exit 1, no session, and is an error row in doctor; no new warning is added for skill names ending in `verify`, and the `review` warning keeps its condition and message.
- With `agent.verify = copilot`, no GitHub login reaching the sessions refuses the run naming `agent.verify` (`vloop: agent.verify is copilot but no GitHub login reaches its sessions: …`), exit 1, and is doctor's copilot login error row; copilot missing or `copilot --version` failing is doctor's copilot error row and stops the run.
- With `run.verify = false`, `agent.verify = copilot` adds no row, warning or refusal — not even with `model.verify` unset or `.claude/skills/verify` present — and the run does T1 with no verify session, exit 0. With `agent.verify = claude` none of this appears. The checks for agent.work, agent.review and agent.second-review keep their words and behaviour.
- Committed tests cover the warning, the model.verify refusal, K1, the login refusal and the `run.verify = false` case for agent.verify (extending the existing doctor and run preflight tests); a code comment that lists the Copilot phases or keys (doctor's experimental warning included) names verify too.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T3/gate.sh
```

</details>

### T4 — Document Copilot verify sessions in the fence audit, the configuration guide and the domain pages

`pending` · depends on: T1, T2, T3

A1's prose. The fence audit, the guide and the domain pages say what a Copilot verify session may do and who enforces each rule: the operator reads these to decide whether to trust `agent.verify = copilot`, so every sentence must match what T1–T3 built. Verify's pass 3 looks for every surface that names the Copilot phases, so none may be left saying work, review and second-review only. Prose task: if it fails review twice on wording, it is blocked and the operator finishes it.

**Acceptance**

- `docs/design-notes/fence-audit.md` gains a part for a Copilot verify session, under a heading naming verify, that classifies every deny rule of `fence/verify.json` and every deny-list pattern (the whole review list, `shell(rm:*)` included) by who enforces it in a Copilot verify session — the driver after the session, the hook, or neither — and says, for each `neither`, what covers it.
- `docs/guide/configuration.md`: the `agent.verify` row offers `claude` and `copilot` and says `copilot` is experimental; `agent.plan` and `agent.gate-review` still offer `claude` alone; the Copilot paragraph covers verify sessions, the `agent.verify` warning and the `model.verify` refusal, and no longer says `agent.verify` stays on `claude`; the fence hook section's `--phase` lists name verify, it says a verify session is judged against the verify fence and may write only under `.vloop/tmp/`, and K1's hint names `/vloop:verify` in a verify session.
- `docs/domain/execution/session.md` says copilot runs work, review, second-review and verify sessions; its list of the fence files (one file per phase) names the verify fence; its Copilot hook paragraph says the hook serves the verify phase with the verify fence and allows writes only under `.vloop/tmp/` in it. The glossary's hook entry in `docs/domain/domain-model.md` covers the verify phase.
- Every other surface that names the Copilot phases names verify too, the code comments that list them (the adapter's, the hook's, doctor's experimental warning) included; a surface still listing only work, review and second-review is a violation.
- Every sentence about behaviour matches what T1–T3 built (messages quoted as the code prints them); `fence/*.json` is unchanged; the `go` check still passes.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T4/gate.sh
```

</details>

### T5 — Extend the Copilot acceptance kit with a verify run on the real CLI

`pending` · depends on: T2, T3

A1's operator tool. Gates use a stub copilot; only a run on the real CLI proves a Copilot verify session works and keeps a valid report. This task gives the operator a script that runs the fixture brief with `agent.verify = copilot` and a given model, plan, work and review on Claude, and summarises the verify session's record and the kept report, so the acceptance at close can be recorded. No gate or check runs it.

**Acceptance**

- `tools/copilot-acceptance/` gains a script that runs `greet.loop-brief.md` in a scratch repository with a vloop built from this tree, `agent.verify = copilot` and `model.verify` the model given, plan, work and review on Claude and `run.verify` on, and writes a summary of the verify session's record (agent, model, credits, estimated cost, `session/v4` validity, its `permission_denials`) and of the kept report (present or not, `verify/v1` validity, its number of findings).
- The README gains its section, under a heading naming verify: how to run the script, what it needs (copilot, a logged-in gh, claude), what it spends, what the summary holds, and that the outcome is recorded in `docs/design-notes/copilot-cli-experiments.md` before close.
- The kit's scripts pass `sh -n` and its Python parses; flags, keys and messages of existing vloop commands the script uses are taken from running them, not memory; no gate or check runs the kit, copilot, gh or the network.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T5/gate.sh
```

</details>

### T6 — Close the run: every gate at once, and the worked example line for line

`pending` · depends on: T1, T2, T3, T4, T5

The brief's close. Every earlier gate re-runs on the final tree, and the review holds the brief's worked example against the built binary line for line, so nothing a later task changed has broken an earlier one's contract. It builds nothing new unless a gate or a worked-example line fails, and then fixes only that.

**Acceptance**

- Every gate T1–T5 passes on the final tree.
- Every line of the brief's worked example holds against the built binary: the two `fence check --phase verify` tables and the apply_patch lines, the three unchanged lines, the changed unknown-phase line, the stub run's warning, record, argv, hooks file and kept report, and each planted failure.
- The `go` check passes on the final tree.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T6/gate.sh
```

</details>


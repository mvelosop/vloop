# Plan — B20261027-2111-b24-copilot-work

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **6/6 done** · iteration 7

**Brief:** `docs/briefs/B20261027-2111-b24-copilot-work.loop-brief.md` · **Updated:** 2026-10-27T22:32:53Z

## Progress

- [x] **T1** — Add the work phase to the fence judge and to vloop fence check
- [x] **T2** — Run work sessions on Copilot: the hook's work phase, the adapter and agent.work = copilot
- [x] **T3** — Check agent.work = copilot in the run preflight and doctor
- [x] **T4** — Document Copilot work sessions in the fence audit, the configuration guide and the domain pages
- [x] **T5** — Extend the Copilot acceptance kit with a work run on the real CLI · 1 attempt(s)
- [x] **T6** — Close the run: every gate at once, and the worked example line for line

## Tasks

### T1 — Add the work phase to the fence judge and to vloop fence check

`done` · depends on: none

W1's core. The judge behind `vloop fence check` and the Copilot hook knows only the review phases, so a Copilot work session cannot be fenced. This task gives it a `work` phase that enforces `fence/work.json` — the same rule set Claude's work sessions run under — plus the adapter's work deny-list stems (the review stems without `rm:*`), and lets work sessions write anywhere in the repository except where S-4 forbids. T2 puts it behind the hook; the review phases must not change by a byte.

**Acceptance**

- `vloop fence check --phase work --root <dir> --tool <name> [--path <p>] ["<command>"]` gives every decision of the brief's two worked-example tables and its apply_patch lines; where the brief shows a message, the deny line is `deny: ` followed by that message without its `vloop fence: ` prefix (the prefix is the hook's, as for review).
- Shell in work: the same splitting, unwrapping and outright denials as review (backtick, `$(`, `<(`, `>(`, `eval`, `source`, `.`, unterminated quote, non-literal `-c` string, command word starting with `$`, the shell keywords, subshells); a simple command is denied by a `Bash(…)` deny rule of the embedded `fence/work.json` (`<command> is denied by the work fence (Bash(<rule>))`) or a work deny-list stem (`<command> is denied by the Copilot deny list (<stem>)`). The rules and stems are read from the embedded fence and the adapter's list; a second hand-copied rule list is a violation.
- Writes in work (`create`, `edit`, `apply_patch`, a redirection) resolve against --root as in review and are allowed anywhere under the root except under `.git/` or under `.vloop/` outside `.vloop/tmp/` (`write to <path> is not allowed in a work session: nothing under .git/ or .vloop/ but .vloop/tmp/`) and outside the root (`write to <path> is not allowed in a work session: it is outside the repository`); `/dev/null` is allowed.
- Reads in work as in review; `skill` only for one of the embedded plugin's skills, else `skill <name> is not vloop's — use /vloop:work`; every other tool, `web_fetch`, `task` and names ending in `_subagent`: `tool <name> is not allowed in a work session`.
- `rm` of a file is allowed in work; `rm -rf`, `rm -r`, `rm -R`, `rm -fr` and the rest of the work fence's recursive rules deny; `pnpm add`, `npm install` and `node x.js` are allowed and `pip install` and `sh script.sh` are denied, as the brief decides.
- An unknown phase is a usage error, exit 2, with `unknown phase "<p>": want review, second-review or work`.
- The review and second-review phases decide and word every call exactly as before — `rm src/old.ts` with `--phase review` is still `deny: rm src/old.ts is denied by the Copilot deny list (rm:*)`, exit 1. The existing judge tests that pin `work` as an unknown phase change only to the new behaviour, and the work notes name the contract item (W1).
- The work phase's rules are covered by committed Go table tests in `internal/fence` that run on every OS in the `go` check, covering every line of the worked example's two tables and the path rules; a work phase with no committed tests is a violation.
- No Claude session's argv, fence file or settings changes, and `fence/work.json` and `fence/review.json` are unchanged.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T1/gate.sh
```

</details>

### T2 — Run work sessions on Copilot: the hook's work phase, the adapter and agent.work = copilot

`done` · depends on: T1

W1's hook half and W2's core. With T1's judge, `vloop fence hook --phase work` can fence a Copilot work session; this task lets the Copilot adapter run `/vloop:work <task>` exactly as it runs a review — same plugin, fresh home, guidelines, JSON Lines parsed into session/v4 — with only two differences: its hooks file runs the hook with `--phase work`, and its argv omits `--deny-tool=shell(rm:*)`. `agent.work` then accepts `copilot`. Everything the driver does around the work session stays as it is: the checks after the session are the boundary (S-2, S-4).

**Acceptance**

- `vloop fence hook --phase work --root <dir>` answers each payload with T1's work decision, the reason being `vloop fence: ` and T1's message, exit 0; an unknown phase is the deny line with `unknown phase "<p>": want review, second-review or work` and exit 2; the review and second-review phases answer exactly as before.
- `agent.work` accepts `claude` (default) and `copilot`, from the file and from `VLOOP_AGENT_WORK`; `agent.plan`, `agent.gate-review` and `agent.verify` still refuse `copilot` (config set exits 2; a run with their variable set to copilot starts no session).
- With `agent.work = copilot`, the work session runs `copilot -p "/vloop:work <task>"` with the model and effort P-6 resolves, the plugin by `--plugin-dir`, a fresh home under `.vloop/tmp/` holding a hooks file that runs `fence hook --phase work` in both command forms, the guidelines as custom instructions, and every review deny-list entry on the argv except `--deny-tool=shell(rm:*)`.
- A Copilot work session's JSON Lines become a `session/v4` record with phase `work`, agent `copilot`, the model, credits, estimated cost, `cost_source` and `permission_denials`.
- A Copilot review or second-review session's argv and hooks file are unchanged (`shell(rm:*)` on the argv, `--phase review`/`second-review` in the hooks file); a Claude session's argv, fence and settings are unchanged.
- Everything around the work session is unchanged whatever the agent: the gate, the checks, the guards (a moved ref halts with exit 9; inputs restored), the proposal read from `.vloop/tmp/`, the review and the one commit.
- Committed Go tests cover the adapter's work-phase argv (no `shell(rm:*)`, the rest of the list present), its hooks file's `--phase work`, `Supports(work)`, the hook's work phase and `agent.work` accepting copilot; tests that pinned work as unsupported or `work` as an unknown hook phase change only to the new behaviour, and the work notes name the contract item (W1, W2).

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T2/gate.sh
```

</details>

### T3 — Check agent.work = copilot in the run preflight and doctor

`done` · depends on: T2

W2's guard rails. An operator who sets `agent.work = copilot` gets what Copilot review already gives: the experimental warning, a refusal while `model.work` is still the Claude default, and the copilot checks — copilot on PATH and runnable, a GitHub login reaching its sessions, K1's refusal of a project skill named like one of vloop's — in `vloop run` and as `vloop doctor` rows. Without these, a run would hand a Claude alias to Copilot or let a project skill shadow `/vloop:work`.

**Acceptance**

- `vloop run` with `agent.work = copilot` prints `  ! agent.work is copilot: experimental (docs/guide/configuration.md)` and goes ahead; `vloop doctor` prints `! agent.work is copilot: experimental (docs/guide/configuration.md)` and the copilot and copilot login rows.
- `vloop run` with `agent.work = copilot` (from the file or `VLOOP_AGENT_WORK`) and `model.work` at its default refuses before any session, exit 1: `vloop: agent.work is copilot but model.work is the default "sonnet", a Claude alias: set model.work to a Copilot model id`.
- With `agent.work = copilot`, a project skill directory under `.claude/skills/` or `.github/skills/` named like one of vloop's skills stops `vloop run` with K1's message (`vloop: .claude/skills/work has the name of vloop's work skill; a copilot session may load it instead — rename it`), exit 1, no session, and is an error row in doctor.
- With `agent.work = copilot`, no GitHub login reaching the sessions refuses the run naming `agent.work` (`vloop: agent.work is copilot but no GitHub login reaches its sessions: …`), exit 1, and is doctor's copilot login error row; copilot missing or `copilot --version` failing is doctor's copilot error row and stops the run.
- With `agent.work = claude`, none of this appears and a project skill named `work` stops nothing; the checks for agent.review and agent.second-review keep their words and behaviour.
- Committed tests cover the warning, the model.work refusal, K1 and the login refusal for agent.work (extending the existing doctor and run preflight tests).

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T3/gate.sh
```

</details>

### T4 — Document Copilot work sessions in the fence audit, the configuration guide and the domain pages

`done` · depends on: T1, T2, T3

A1's prose. The fence audit, the guide and the domain pages say what a Copilot work session may do and who enforces each rule: the operator reads these to decide whether to trust `agent.work = copilot`, so every sentence must match what T1–T3 built. Prose task: if it fails review twice on wording, it is blocked and the operator finishes it.

**Acceptance**

- `docs/design-notes/fence-audit.md` gains a part for the work fence, under a heading naming it, that classifies every deny rule of `fence/work.json` and every work deny-list pattern (the review list without `shell(rm:*)`) by who enforces it in a Copilot work session — the driver after the session, the hook, or neither — and says, for each `neither`, what covers it; it says `rm` of a file is allowed and why.
- `docs/guide/configuration.md`: the `agent.work` row offers `claude` and `copilot` (experimental); `agent.plan`, `agent.gate-review` and `agent.verify` still offer `claude` alone; the Copilot section covers work sessions — the hook judges them against the work fence, what it allows them, that shell syntax it cannot judge (`if`, `for`, `while`, `$(…)`, subshells) is denied, the model.work refusal and the experimental warning.
- `docs/domain/execution/session.md` and the glossary of `docs/domain/domain-model.md` say copilot runs work, review and second-review sessions and the hook serves the work phase with the work fence; nothing still says copilot runs review sessions only.
- Every sentence about behaviour matches what T1–T3 built (messages quoted as the code prints them); the `go` check still passes.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T4/gate.sh
```

</details>

### T5 — Extend the Copilot acceptance kit with a work run on the real CLI

`done` · 1 attempt(s) · depends on: T2, T3

A1's operator tool. Gates use a stub copilot; only a run on the real CLI proves a Copilot work session works. This task gives the operator a script that runs the fixture brief with `agent.work = copilot` and a given model, plan and review on Claude, and summarises each work session and each task's outcome, so the acceptance at close can be recorded. No gate or check runs it.

**Acceptance**

- `tools/copilot-acceptance/` gains a script that runs `greet.loop-brief.md` in a scratch repository with a vloop built from this tree, `agent.work = copilot` and `model.work` the model given, plan and review on Claude, and writes a summary listing each work session record (agent, model, credits, estimated cost, `session/v4` validity, its `permission_denials`) and each task's outcome.
- The README gains its section: how to run it, what it needs (copilot, a logged-in gh, claude), what it spends, what the summary holds, and that the outcome is recorded in `docs/design-notes/copilot-cli-experiments.md` before close.
- The kit's scripts pass `sh -n` and its Python parses; flags, keys and messages of existing vloop commands the script uses are taken from running them, not memory; no gate or check runs the kit, copilot, gh or the network.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T5/gate.sh
```

</details>

### T6 — Close the run: every gate at once, and the worked example line for line

`done` · depends on: T1, T2, T3, T4, T5

The brief's close. Every earlier gate re-runs on the final tree, and the review holds the brief's worked example against the built binary line for line, so nothing a later task changed has broken an earlier one's contract. It builds nothing new unless a gate or a worked-example line fails, and then fixes only that.

**Acceptance**

- Every gate T1–T5 passes on the final tree.
- Every line of the brief's worked example holds against the built binary: the two `fence check --phase work` tables and the apply_patch lines, the review phase's `rm` line, the unknown phase, the stub run's warning, record, argv and hooks file, and each planted failure.
- The `go` check passes on the final tree.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T6/gate.sh
```

</details>


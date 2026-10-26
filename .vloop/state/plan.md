# Plan — B20261026-1002-b22-fence-second-review

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **2/9 done** · iteration 2

**Brief:** `docs/briefs/B20261026-1002-b22-fence-second-review.loop-brief.md` · **Updated:** 2026-10-26T22:09:31Z

## Progress

- [x] **T1** — Add vloop fence check: the judge that decides a tool call against a phase's fence, looking through shell wrappers
- [x] **T2** — Add vloop fence hook: read one Copilot preToolUse payload and answer allow or deny as JSON, failing closed
- [ ] **T3** — Run every Copilot session under the hook, parse its denials into permission_denials, and keep Copilot on vloop's skills (K1)
- [ ] **T4** — Add the second reviewer's keys, its doctor rows, and the reworded Copilot experimental warning
- [ ] **T5** — Run the optional second review session after the primary, in each mode, with its records and the end-of-run dissent list
- [ ] **T6** — Report the second review in metrics/v4
- [ ] **T7** — Write the fence audit and update the configuration guide, the domain pages and both operator playbooks
- [ ] **T8** — Extend the Copilot acceptance kit with the hook probe, the hook's failure modes and a run with a copilot second reviewer
- [ ] **T9** — Close the run: every gate at once, and the worked example line for line

## Tasks

### T1 — Add vloop fence check: the judge that decides a tool call against a phase's fence, looking through shell wrappers

`done` · depends on: none

H1's core. Copilot review sessions today run with every tool allowed and a deny list that `sh -c` and `env` walk past (B19's C5.2). This task builds the one judge — a pure splitter, unwrapper, matcher and path rule set over the embedded review fence plus the adapter's deny-list stems — exposed as `vloop fence check` for the operator and the gates; T2 puts the same judge behind the Copilot hook. Fail closed: what the judge cannot parse it denies.

**Acceptance**

- `vloop fence check --phase <review|second-review> --root <dir> --tool <name> [--path <p>] ["<command>"]` prints `allow` and exits 0, or prints `deny: <reason>` and exits 1; bad arguments (unknown phase, no --tool) exit 2. A file or read tool's path, and a skill's name, are given with `--path`, as the worked example's second table (`--tool`, `--path`) shows.
- Every line of the worked example's two tables gives the decision the brief shows, `skill vloop-review` with the reason `skill vloop-review is not vloop's — use /vloop:review` and `frobnicate` with `tool frobnicate is not allowed in a review session`.
- Shell (`bash`, `powershell`): the command string is split at `;`, `&&`, `||`, `|`, `&` and newlines outside quotes; each simple command loses leading NAME=value assignments and is unwrapped repeatedly through env (-i, -u NAME, -, assignments), command, exec, nohup, time, nice (-n N), timeout (options and duration), stdbuf, xargs (options) and sh/bash/zsh/dash/ksh -c (options before -c allowed), whose literal string is judged in turn; a command word with a directory is judged by its base name.
- A simple command is denied when it matches a `Bash(…)` deny rule of the embedded review fence with the fence's matching (`:*` a prefix, `*` any run) or one of the adapter's deny-list stems; the rules are read from the embedded fence and the adapter's list, not a hand-copied second list — a duplicated rule list in the judge is a violation.
- Denied before any matching: a backtick, `$(`, `<(`, `>(`, `eval`, `source`, `.` as a command, an unterminated quote, a -c string that is missing or holds `$`, and a command word starting with `$`. A redirection (`>`, `>>`, `>|`, `&>`) to a file is a write judged by the path rules; `/dev/null` is allowed.
- Writes (`create`, `edit`, a redirection) resolve against --root (absolute taken as given, then cleaned) and are allowed only under `.vloop/tmp/`; reads (`view`, `read_file`, `grep`, `rg`, `grep_search`, `glob`, `file_search`) are allowed except under `~/.claude`, `~/.copilot` or `~/.config/gh` (a leading `~` expands to HOME); `skill` is allowed only for the name of one of the embedded plugin's skills; every other tool, `web_fetch`, `task` and names ending in `_subagent` are denied.
- The splitter, the unwrapping and the path rules are pure functions with committed Go table tests that run on every OS in the `go` check, covering every deny line of the worked example and each wrapper; a judge with no committed tests is a violation.
- No Claude session's argv, fence file or settings change.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T1/gate.sh
```

</details>

### T2 — Add vloop fence hook: read one Copilot preToolUse payload and answer allow or deny as JSON, failing closed

`done` · depends on: T1

H1's second half: the command Copilot runs before every tool call. It wraps T1's judge around Copilot's payload so a Copilot review session can do only what the review fence allows. It must never answer allow on an error — that is the violation the review is told to rule on.

**Acceptance**

- `vloop fence hook --phase <phase> --root <dir>` reads one payload (camelCase `toolName`, `toolArgs`, `cwd`) on stdin and prints exactly one line: `{"permissionDecision":"allow"}` or `{"permissionDecision":"deny","permissionDecisionReason":"vloop fence: <reason>"}`, exit 0; phases are `review` and `second-review`, sharing the review fence.
- A payload that is not JSON or has no `toolName`, an unknown phase (including `work`), or any internal error prints the deny line and exits 2 — never allow.
- The hook judges with T1's judge: a shell tool's `toolArgs.command`, a file tool's `toolArgs.path`, and for `apply_patch` the paths of its `*** Add File:`, `*** Update File:`, `*** Delete File:` and `*** Move to:` lines (a patch with none is denied, and so is an `apply_patch` whose `toolArgs` hold the patch in a shape the hook does not read — the brief does not say whether Copilot sends it as a raw string or an object field, so the hook fails closed on the one it cannot read); the worked example's `apply_patch` with `*** Update File: .git/config` is denied.
- `toolArgs` is accepted both as an object and as a JSON string holding one, so the hook does not deny every call if Copilot sends the string form; which form Copilot sends is taken from its docs, and the work session says in its notes what it found.
- The hook reads nothing but the payload and its embedded fences, writes nothing and never runs a command.
- Committed Go tests cover the payload forms, the JSON answer and every exit-2 case, and run in the `go` check.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T2/gate.sh
```

</details>

### T3 — Run every Copilot session under the hook, parse its denials into permission_denials, and keep Copilot on vloop's skills (K1)

`pending` · depends on: T2

H2 and K1. The judge only fences what it is asked about: this task installs it as the preToolUse hook in each Copilot session's own home, with the running binary and the repository root, so every Copilot tool call passes through it; records Copilot's denials in the session record; and stops a run where a project skill named like one of vloop's could be loaded instead of /vloop:review (the operator saw that happen).

**Acceptance**

- Each Copilot session's own `COPILOT_HOME` holds, under `hooks/`, a hooks file in the format Copilot's docs give, whose `preToolUse` entry runs `<the running binary> fence hook --phase <phase> --root <the repository root>` in both the `bash` and `powershell` command forms; the `--deny-tool` list stays on the argv unchanged.
- A Copilot session record's `permission_denials` lists every `tool.execution_complete` event whose `error.code` is `denied`, in order, as `{"tool_name", "tool_use_id", "message"}` — the tool name from the `tool.execution_start` with the same `toolCallId`, the message masked as the driver masks text; `[]` when there are none.
- The hook allows `skill` only for the embedded plugin's skill names (from T1's judge); `vloop run`, when `agent.review` is `copilot`, refuses to start while `.claude/skills/<name>/` or `.github/skills/<name>/` has exactly a plugin skill's name: `vloop: <path> has the name of vloop's <name> skill; a copilot session may load it instead — rename it`, exit 1, before any session; `vloop doctor` shows the same as an error row; the warning for other names ending in `review` stays.
- With `agent.review = claude`, such a skill stops nothing, and Claude sessions' argv, fence and settings are unchanged.
- Committed tests cover the hooks file, the denial parsing (interleaved calls, a non-denied failure, none) and the K1 refusal, extending the existing Copilot adapter and doctor tests, and run in the `go` check.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T3/gate.sh
```

</details>

### T4 — Add the second reviewer's keys, its doctor rows, and the reworded Copilot experimental warning

`pending` · depends on: T3

S1's configuration, before the session exists: `run.second-review` says what a second reviewer's verdict does, `agent/model/effort.second-review` say who it is. Doctor tells the operator what is configured and warns when the second reviewer adds no independence. The Copilot warning drops "not fenced": after T3, Copilot reviews run under the hook.

**Acceptance**

- Keys: `run.second-review` (`off` default, `record`, `feedback`, `on-pass`, `both-pass`), `agent.second-review` (`claude` default, `copilot`), `model.second-review` (default `sonnet`), `effort.second-review` (the effort values, unset by default), each with its `VLOOP_` variable; an invalid value exits 2.
- `vloop doctor` with `run.second-review` off says nothing of the second review; otherwise a row names the mode, agent and model, and `! run.second-review: the second reviewer is the primary's agent and model — it adds no independence` appears when both match the primary's configured agent and model.
- The Copilot warning reads `! agent.review is copilot: experimental (docs/guide/configuration.md)` in doctor and the run preflight, and `! agent.second-review is copilot: experimental (docs/guide/configuration.md)` for a copilot second reviewer, whose doctor rows include the copilot checks (on PATH, --version, login).
- A copilot second review in use with the default `model.second-review` is refused before any session, naming `model.second-review`, as `agent.review` is; K1's skill refusal also applies when `agent.second-review` is copilot and the second review is in use, and not when `run.second-review` is off.
- Committed tests cover the keys, the doctor rows and the refusals, and run in the `go` check.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T4/gate.sh
```

</details>

### T5 — Run the optional second review session after the primary, in each mode, with its records and the end-of-run dissent list

`pending` · depends on: T4

S1's behaviour. A second review session from any agent and model judges the same task as a fresh /vloop:review, blind to the primary's verdict, under the same guards; the mode decides whether its verdict is only recorded, fed to the retry, asked only after a pass, or a veto (both-pass, the only mode that can fail a task). Dissents are what the operator reads at close.

**Acceptance**

- When the mode is not off, a second review runs after the primary review of an iteration — after every primary review for `record` and `feedback`, only after a primary PASS for `on-pass` and `both-pass`, never when the primary did not run or ended the run — as `/vloop:review <task>` in a fresh session of phase `second-review`, under the review fence (the hook's `second-review` phase for copilot), with its model and effort from `model/effort.second-review`, never from the task's own model or effort.
- Before it starts the driver removes the primary's verdict from `.vloop/tmp/`; the guards are the primary's: inputs restored, a moved ref or git setting halts with exit 9, changed files reverted; no valid verdict, a changed file or tampering make it FAIL with the finding.
- In `record`, `feedback` and `on-pass` the second review never changes the task's outcome, status or attempts, and its session error or timeout is a warning; in `both-pass` a task is done only when both PASS, a second FAIL rejects the iteration as a primary FAIL does with its findings as notes prefixed `second review: `, and its session error or timeout is treated as the primary's would be.
- In `feedback`, when the primary did not PASS, the second's findings join the notes after the primary's, each prefixed `second review: `; in `record` and `on-pass` nothing reaches the notes.
- The run log says `   second review: <PASS|FAIL>`, with ` (dissent)` when it differs from the primary's verdict; the verdict is kept as `reports/<iter>-second-verdict.json`; the run's end lists `second review: <n> dissent(s): <task> (primary <v>, second <v>), …` when there are any; with the mode off none of this appears.
- Records: `session/v4` (session/v3 with `second-review` among the phases), `iteration/v3` adding `second_review` (`verdict`, `dissent`), absent when none ran; both schemas embedded and known to `vloop schema validate`; readers keep accepting v3, v2 and earlier; the second session's cost counts toward the cost ceiling.
- Committed tests with a stub agent cover each mode, the guards and the records, extending the existing run tests, and run in the `go` check.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T5/gate.sh
```

</details>

### T6 — Report the second review in metrics/v4

`pending` · depends on: T5

S1's measurement: the operator compares runs with and without a second reviewer, so the metrics a run commits must say which agents ran it, how many second sessions there were and where they dissented.

**Acceptance**

- The metrics a run reports are `metrics/v4`: `metrics/v3` with `second-review` in `agents` (the sorted distinct agents of its sessions, `[]` when none ran) and a `second_review` object holding `mode`, `sessions`, `dissents` and `dissent_tasks`; the schema is embedded and `vloop schema validate metrics/v4` knows it.
- Readers keep accepting metrics/v3 and earlier snapshots; nothing else in the metrics changes meaning.
- Committed tests cover the new fields for a run with and without a second review, and run in the `go` check.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T6/gate.sh
```

</details>

### T7 — Write the fence audit and update the configuration guide, the domain pages and both operator playbooks

`pending` · depends on: T3, T5, T6

A1. The operator needs to know, rule by rule, who enforces the fence for a Copilot review session; the guide must describe the hook, its fail-closed rules, K1 and what each second-review mode does; the domain gains the second-review kind, dissent and the hook; and the close reads the dissents. Prose task: if it fails review twice on wording, block it and the operator finishes it.

**Acceptance**

- `docs/design-notes/fence-audit.md` classifies every deny rule of `fence/review.json` and every deny-list pattern of the adapter by who enforces it for a Copilot review session — the driver after the session, the hook, or neither — and says, for each `neither`, what covers it.
- `docs/guide/configuration.md` has a row for each S1 key (default, valid values, variable), describes `vloop fence check` and `vloop fence hook`, the fail-closed rules, K1, what each `run.second-review` mode does and what a dissent is, keeps Copilot review experimental, and no longer says it is unfenced.
- `docs/domain/domain-model.md` gains second review and dissent in the glossary; `docs/domain/execution/session.md` the second-review kind, the hook and the Copilot denials; `docs/domain/execution/task.md` what the second review's findings do to notes; `docs/domain/measurement/metrics.md` the second review in metrics/v4.
- `plugin/skills/operate/SKILL.md` and `.claude/skills/vloop-operator/SKILL.md` read the dissents at close and record each one that holds as a `verification-finding`.
- Every sentence about behaviour matches what T1–T6 built; a doc test the repository already has still passes (the `go` check).

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T7/gate.sh
```

</details>

### T8 — Extend the Copilot acceptance kit with the hook probe, the hook's failure modes and a run with a copilot second reviewer

`pending` · depends on: T2, T5

The binary never calls the network, so acceptance on the real Copilot CLI is the operator's, at close. This kit is how the operator measures what no gate can: that Copilot denies each worked-example form under the hook, refuses a skill other than vloop's, what Copilot does when the hook itself fails, and a real feedback-mode run with a copilot second reviewer.

**Acceptance**

- `tools/copilot-acceptance/` gains a probe of the review session under the hook asking for each deny form of the worked example's shell table (verbatim) and a skill other than vloop's, recording per form whether Copilot ran it.
- It probes the hook's failure modes — its command missing, exit 1, output not JSON, a timeout — and records whether Copilot then runs the call.
- It runs a fixture brief with `run.second-review = feedback` and `agent.second-review = copilot`, and its summary reports the second-review records and dissents.
- The README says how to run each part and that the outcome is recorded in `docs/design-notes/copilot-cli-experiments.md` before close; no gate or check runs the kit, copilot, gh or the network.
- The kit's scripts pass `sh -n` and its Python parses; flags and messages of existing commands are taken from running them.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T8/gate.sh
```

</details>

### T9 — Close the run: every gate at once, and the worked example line for line

`pending` · depends on: T1, T2, T3, T4, T5, T6, T7, T8

The last check before the operator's acceptance: every gate of this plan passes on the same tree, which together hold the worked example line for line. Any gap found here is fixed in the code it belongs to, not in a gate.

**Acceptance**

- Every gate T1–T8 passes on the final tree.
- Every line of the brief's worked example holds against the built binary: the fence check tables, the hook's unreadable payload and apply_patch, the second-review lines for each mode, and the K1 refusal and rename.
- The `go` check passes on the final tree.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T9/gate.sh
```

</details>


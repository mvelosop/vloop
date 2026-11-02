# Plan — B20261102-1332-b30-i113-verify-unrunnable-lines

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **4/5 done** · iteration 6

**Brief:** `docs/briefs/B20261102-1332-b30-i113-verify-unrunnable-lines.loop-brief.md` · **Updated:** 2026-11-02T14:53:33Z

## Progress

- [x] **T1** — Add the verify/v2 report schema, embedded and listed · 1 attempt(s)
- [x] **T2** — Read verify/v2 in the driver's verify step and print not_run and complete at close · 1 attempt(s)
- [x] **T3** — Make the verify skill write verify/v2 from the start and list the lines it does not run
- [x] **T4** — Teach the brief skills and operate about unrunnable lines; name verify/v2 in the docs, the kit and the evals
- [ ] **T5** — Close the run: every gate and the worked example on the final tree

## Tasks

### T1 — Add the verify/v2 report schema, embedded and listed

`done` · 1 attempt(s) · depends on: none

The verify session needs a report that says whether it finished its three passes and which worked-example lines it did not run, and a published schema never changes, so this adds verify/v2 beside verify/v1. Everything after it — the driver reading the report, the skill writing it, the docs naming it — depends on this schema existing and on its exact shape.

**Acceptance**

- schemas/verify.v2.json exists and is embedded like the others; `vloop schema list` names both verify/v1 and verify/v2.
- verify/v2 has verify/v1's properties and constraints, `schema`'s const is `verify/v2`, and it requires two more properties: `complete`, a boolean, and `not_run`, an array (empty allowed) of objects with required `line` and `reason`, both non-empty strings.
- `vloop schema validate verify/v2` exits 0 on each worked-example stub report and 1 on a verify/v1 report, on a report without `complete`, and on a `not_run` entry missing `reason`.
- schemas/verify.v1.json is byte for byte as at the base.
- The schema lists pinned in internal/cli/schema_test.go, internal/schema/schema_test.go and cmd/vloop/b2_e2e_test.go gain verify/v2 and change in nothing else; the work session names the contract item (V1, the new schema listed) in its notes.
- Committed tests run by the go check cover verify/v2: at least one valid report and the three planted invalid ones (a verify/v1 report, no `complete`, a `not_run` entry without `reason`). A task that adds the schema and leaves no such test fails this criterion.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T1/gate.sh
```

</details>

### T2 — Read verify/v2 in the driver's verify step and print not_run and complete at close

`done` · 1 attempt(s) · depends on: T1

B29's verify session stopped early and left no report, so nothing it checked was kept. The driver now reads the report as verify/v2 even after a session that failed, timed out or left no record, prints how many example lines were not run, and warns when the session did not finish; `vloop brief close` shows the lines not run and the unfinished session to the operator. None of this may change how a run ends.

**Acceptance**

- The verify step reads .vloop/tmp/verify.json as verify/v2 only, after a session that exited 0 and also after one that timed out, exited non-zero or left no record; a session that failed to start leaves nothing to read, as at the base. Accepting a verify/v1 report from a session is a violation.
- A valid report is kept as `<run dir>reports/verify.json` (masked, trailing newline) and the driver prints on stdout exactly one of `verify: no findings`, `verify: <N> finding(s) — see <run dir>reports/verify.json` (when not_run is empty), `verify: no findings; <M> example line(s) not run — see <run dir>reports/verify.json` or `verify: <N> finding(s); <M> example line(s) not run — see <run dir>reports/verify.json` (when not_run has M ≥ 1 entries); `<run dir>` is repo-relative with a trailing `/`; `example line(s)` is used even for one.
- When `complete` is false and the session exited 0, the driver then warns on the warning stream `vloop: the verify session stopped before its last pass — verify the rest by hand; see <run dir>reports/verify.json`. After a session that timed out, exited non-zero or left no record, the line is followed by the base's `vloop: the verify session failed to run — verify by hand; see <run dir>run.log` and not by the stopped-early warning.
- A report that does not validate as verify/v2 (a verify/v1 report included) is kept by no one: after a session that exited 0 the driver prints the base's `vloop: the verify session left no valid .vloop/tmp/verify.json — verify by hand`; after a failed session only the base's failed-to-run warning.
- Neither `complete` nor `not_run` changes the run's ending, its exit code or any task.
- A bare `vloop brief close` on a kept verify/v2 report prints the header `verify of <run id>/<stamp>: ` followed by `no findings` or `<N> finding(s)`, then `; <M> example line(s) not run` when not_run is not empty; the findings as at the base; one line `  not run: <line> — <reason>` per entry; and `  the session stopped before its last pass` when complete is false. A kept verify/v1 report prints exactly the base's lines.
- The base's stub verify reports in cmd/vloop/run_verify_test.go, cmd/vloop/run_iterate_test.go and cmd/vloop/b16_e2e_test.go become verify/v2 and nothing else in those tests weakens; internal/cli/close_test.go keeps its verify/v1 reports and gains verify/v2 siblings; the work session names the contract item in its notes.
- Committed tests run by the go check cover every worked-example line and planted failure: in cmd/vloop/run_verify_test.go for the run lines, in internal/cli/close_test.go for the close lines, and in internal/driver/ with a stub runner whose result says the session timed out and one that left no record (the same two lines, the report kept). A case of the worked example with no test behind it fails this criterion.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T2/gate.sh
```

</details>

### T3 — Make the verify skill write verify/v2 from the start and list the lines it does not run

`done` · depends on: T1

The verify skill had only two outcomes, a full report or none, so a session that hit a denied command or ran out of time left nothing. It now writes and validates the report as soon as it has read the brief and after each pass, lists each worked-example line it cannot or does not run instead of stopping, keeps command output short, and makes its scratch repositories in a form no fence denies.

**Acceptance**

- The skill writes .vloop/tmp/verify.json as verify/v2 and validates it with `vloop schema validate verify/v2 .vloop/tmp/verify.json` as soon as it has read the brief and again after pass 1, pass 2 and pass 3, each time with every finding, count and line not run so far; `complete` is false in every write but the last, which sets it true, and `notes` says which passes remain while it is false.
- A worked-example line is not run when the brief keeps it for the operator's acceptance, when its command is one the verify fence denies (the skill names at least `vloop init`, `vloop run`, `vloop brief review`, `vloop brief close` and `claude`), when it needs a live model (Claude or Copilot), the network or credentials, or when a command it needs was denied. Each is one `not_run` entry: `line` quoted as the brief writes it, `reason` in a few words (the denied rule, `needs a live model`, `the operator's, at close`, `covered by <test>`).
- The skill does not retry a denied command in another form (the denial goes into the reason), does not count a line not run in `example_lines`, and does not make it a finding — any of these is a violation.
- Scratch set-up is one command per tool call; a scratch repository is made with `git init -q <dir>`, never `git -C <dir> init` and never chained after `mkdir`. Each command's output is cut to what the session needs (through `head` or `tail`).
- The skill's report example is a valid verify/v2 report with one `not_run` entry; every mention of verify/v1 in the skill (its description included) becomes verify/v2; the skill cites no `fence/` path.
- The skill is at most 1500 words and the verify ceiling in internal/budget/budget.go stays 1500; no ceiling changes in this task.
- The skills test (go check: `go test ./internal/cli/`) and the budget test (`go test ./internal/budget/`) pass on the changed skill; they are the committed tests that keep the example valid and the skill under its ceiling.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T3/gate.sh
```

</details>

### T4 — Teach the brief skills and operate about unrunnable lines; name verify/v2 in the docs, the kit and the evals

`done` · depends on: T2, T3

The knowledge of which worked-example lines a verify session cannot run moves upstream to the brief's author and reviewer, and the operator is told to run what the session left; the docs, the Copilot acceptance kit and the verify evals follow verify/v2. This is mostly prose: if the review fails it twice on wording, the task is blocked and the operator finishes it.

**Acceptance**

- plugin/skills/create-brief/guidelines.md carries, in the section on what sessions cannot do, the brief's rule: **Say which worked-example lines the verify session cannot run.** — no `vloop init`, `vloop run`, `vloop brief review` or `claude`, no live model (Claude or Copilot), no network, no credentials; cover such a line with an end-to-end test the plan writes with stub agents, named beside the line, or move it to `## Acceptance, at close (the operator's)`.
- Both plugin templates and both binary templates gain, between the Worked example and Out of scope, the optional section `## Acceptance, at close (the operator's)` (English pair) / `## Aceptación, al cierre (del operador)` (Spanish pair), with a placeholder in the template's language saying what goes there, identical within each pair so internal/cli/plugin_templates_test.go stays green; `vloop brief new` writes it.
- review-brief's check 4 (Runnability) asks whether a worked-example line needs what the verify session cannot do (a command the verify fence denies, a live model, the network, credentials) with no named test covering it and outside the acceptance section — such a line is a `runnability` finding; the commands are named in plain words, not as code, and no fence/ path is cited.
- operate's close step 2 gains one sentence: run each line the verify report lists as not run, and the brief's acceptance section, before stating findings. internal/budget/budget.go sets operate's ceiling to 1800 with a comment naming B30 (#113); every other ceiling is unchanged, the verify ceiling included.
- docs/domain/execution/session.md's verify report section becomes verify/v2's, naming `complete`, `not_run` and the driver's stopped-early warning; docs/domain/execution/run.md and docs/guide/concepts.md name verify/v2 where they name the report the session writes.
- tools/copilot-acceptance/report.py validates each kept verify report against the schema its own `schema` field names (verify/v1 or verify/v2) and its report line says which; tools/copilot-acceptance/README.md and verify.sh's header comment say so.
- plugin/evals/verify-planted-mismatches/graders/report-validates.md's pattern is `vloop schema validate verify/v2`.
- A new eval case plugin/evals/verify-unrunnable-line/, built like verify-planted-mismatches (prompt.md, case.yaml, a scaffold.sh that exits non-zero when its built result does not hold), whose brief's worked example holds one sound line and one line running `claude -p "say hi"`; its graders: a `regex` on the trace for `vloop schema validate verify/v2`, a `regex` on the file .vloop/tmp/verify.json for `"complete": *true`, and an `llm` grader focused on that file asking that the claude line is the one not_run entry and that no finding names it. It joins the catalogue in plugin/evals/README.md.
- The skills test's eval and fenced-skill checks (go check: `go test ./internal/cli/`, among them TestPluginSkills and TestEvalGrantsCoverFence) and the budget test pass: the new case's grants cover the work fence and appear in docs/guide/evals.md's --allow-tools, as verify-planted-mismatches' do. These and the eval case itself are the committed tests this task leaves.
- No fence file and not docs/design-notes/fence-audit.md change.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T4/gate.sh
```

</details>

### T5 — Close the run: every gate and the worked example on the final tree

`pending` · depends on: T1, T2, T3, T4

The last task proves the whole brief holds at once on the final tree: every earlier gate passes together, no fence or the fence audit changed, and the worked example holds line for line. It exists so a regression one task caused in another's surface is found before the run's verify session, not after.

**Acceptance**

- Gates T1 to T4 all pass on the final tree, run together by this task's gate.
- fence/*.json and docs/design-notes/fence-audit.md are byte for byte as at the base, with no fence file added or removed.
- The work session's notes walk the brief's worked example line for line (each stub report, each planted failure, the close lines, the schema lines, the skills lines) and name the committed test or gate that covers each; a line no test covers is fixed in this task or reported, not skipped.
- Any fix this task makes is to the contract as the brief states it, leaves a committed test where it changes behaviour, and weakens no check.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T5/gate.sh
```

</details>


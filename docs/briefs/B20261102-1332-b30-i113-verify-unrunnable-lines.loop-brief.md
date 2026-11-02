---
name: B20261102-1332-b30-i113-verify-unrunnable-lines.loop-brief
description: Make the verify session leave a report even when it cannot run every worked-example line — verify/v2 with the lines not run and whether every pass finished — and make the brief skills say which lines it cannot run
kind: brief
status: draft
created: 2026-11-02
issue: 113
seeds: Issue 113 under 75, from B29's verify session, which left no report; the operator's go-ahead of 2026-11-02 for one brief on the verify skill and the brief skills, the rest decided by the assistant under the operator's delegation
depends-on: [B20261102-0920-b29-i94-copilot-review-brief.loop-brief]
---
# Brief — vloop B30: verify reports what it could not run

- **Issue:** #113 (under #75).
- **Starting point:** extends `main` at `75bba52`, plus this brief's ready
  merge. The planner pins that SHA as the **base** every gate compares
  against — never `HEAD~n`.
- **Produced by:** the operator's go-ahead of 2026-11-02 for one brief on
  the verify skill and the brief skills, after B29's close; written with
  `plugin/skills/create-brief/`, reading B29's verify session record
  (`.vloop/state/runs/B20261102-0920-b29-i94-copilot-review-brief/20261102-101453/sessions/013-verify.json`)
  and its intervention. The design decisions were taken by the assistant
  under the operator's delegation.
- **Review:** pending.

## Executive summary

B29's verify session stopped after 11 turns and 56 s and left no report, so
the driver printed "verify by hand" and nothing the session had checked was
kept. Its record shows why: three commands denied, among them
`mkdir -p .vloop/tmp/vs/ex && git -C .vloop/tmp/vs/ex init -q`, the
skill's own scratch-repository step written with `git -C`, which every
fence denies; and one read denied, of a long command output that Claude
Code had saved under the user's Claude directory, which the fence keeps
sessions out of. And B29's worked example starts with `vloop init` and runs
`vloop brief review`, which the verify fence also denies: no verify session
could run those lines. The verify skill has no way to say so, and no brief
skill tells an author which lines it cannot run.

**The goal:** a verify session always ends with a valid report, listing
each worked-example line it did not run and why, and saying whether it
finished its three passes; and a brief says, for each line the verify
session cannot run, which test covers it or that the operator runs it at
close.

- **V1 — `verify/v2`.** `verify/v1` plus `complete` (did the session finish
  its three passes) and `not_run` (each line not run, with its reason). The
  driver reads only `verify/v2`, also after a session that failed or timed
  out, prints the count of lines not run and warns when the session did not
  finish; `vloop brief close` prints them.
- **V2 — the verify skill.** It writes and validates the report as soon as
  it has read the brief and again after each pass, lists each line it does
  not run instead of stopping, keeps command output short, and sets up its
  scratch directories with one command per call and `git init -q <dir>`.
- **V3 — the brief skills.** `/vloop:create-brief`'s guidelines and both
  templates (with the binary's copies, which `vloop brief new` writes) say how a brief keeps a line the verify session cannot run;
  `/vloop:review-brief`'s runnability check flags a line that does neither.
- **V4 — the docs and the evals.** The domain pages, the concepts guide and
  the Copilot acceptance kit (its README and its report script) name
  `verify/v2`; the verify eval's grader
  follows it, and a new eval plants a line the session must not run.

Unchanged: when the verify session runs, its fence and its Copilot hook;
`verify/v1`'s file, which stays published; every other session, report and
exit code; a run's ending, which a verify report never changes.

## What it is

After the final pass of a complete run, the verify session holds the built
result against the brief. From this brief on, it writes its `verify/v2`
report as soon as it has read the brief and again after each of its three
passes, so a session that stops early, fails or times out still leaves what
it checked, marked `complete: false`, and the driver keeps it and says so.
A worked-example line the session cannot run (a command its fence denies,
a live model, the network, credentials) is listed in `not_run` with its
reason rather than ending the session. The brief skills move that knowledge
upstream: the author covers such a line with an end-to-end test the plan
writes, or moves it to the operator's acceptance section, and the brief
review flags a line that does neither.

**Why the defect got through.** The verify skill was written (B16) and moved
to Copilot (B25) against briefs whose worked examples ran in a scratch
repository with the built binary. B29's was the first worked example a
verify session could not run at all, and the skill's only outcomes were a
full report or none. Nothing between the brief and the session (the brief
check, the review, the plan) knows what the verify fence denies.

## Why this shape, and what was rejected

**Do not re-litigate these** — the planner inherits them.

Decided by the operator on 2026-11-02:

- **One brief for the verify skill and the brief skills**, issue #113;
  issues #110, #111 and #112 stay out (see Out of scope).

Decided by the assistant under the operator's delegation:

- **A new schema, `verify/v2`**, `verify/v1`'s shape plus two required
  fields. *Rejected:* listing the lines in `notes` of `verify/v1` — the
  driver cannot count prose, and the operator would read "no findings" over
  lines that were never run; and adding the fields to `verify/v1` — a
  published schema does not change (B29's decision on `session/v4`).
- **The driver reads `verify/v2` only.** Sessions load the binary's own
  plugin (`--plugin-dir`, the Claude and Copilot adapters alike), so the
  skill and the driver always match. *Rejected:* accepting `verify/v1` too —
  no session can write it once the skill changes. `vloop brief close` still
  prints a kept `verify/v1` report, as today.
- **`complete` is the session's own statement**, written `false` in the
  first report (before pass 1) and after the first and second passes, and
  `true` with the third. The driver warns on `false` and keeps the report.
  *Rejected:* a pass counter — the operator needs only "finished or not", and
  `checked` already counts what each pass covered.
- **A valid report is read even after a session that failed or timed out**,
  so what it checked before it stopped is kept; the base's "failed to run"
  warning still follows. *Rejected:* reading it only after a clean exit —
  B29's session is the case this brief exists for, and a timeout is the
  likeliest way a long verify ends early.
- **A line not run is not a finding** and is not counted in
  `checked.example_lines`. It is one `not_run` entry, `{line, reason}`, the
  line quoted as the brief writes it. *Rejected:* a finding of a new kind —
  findings are disagreements between the brief and the result, and a line
  not run is neither.
- **The skill does not retry a denied command in another form**: the denial
  goes into the line's reason. *Rejected:* teaching the skill ways around a
  fence — the fence is the boundary (S-4).
- **No fence changes.** The skill's scratch repositories are made with
  `git init -q <dir>`, which no fence denies, instead of
  `git -C <dir> init`, which every fence denies (`Bash(git -C:*)`).
  *Rejected:* an allow rule for `git init` in the verify fence — a prefix
  rule cannot exclude `git init --separate-git-dir`, which moves the
  repository's `.git`.
- **A brief keeps an unrunnable line one of two ways**, and the guideline
  names both: covered by an end-to-end test the plan writes with stub
  agents (the test named beside the line), or moved to a section
  `## Acceptance, at close (the operator's)` (Spanish:
  `## Aceptación, al cierre (del operador)`), which the operator runs on the
  real tools at close, as B26–B29 did. *Rejected:* a marker inside the
  Worked example — the verify session would still have to judge each line,
  and a section is what this repository's briefs already use.
- **The verify skill lists a line the brief keeps for the operator** in
  `not_run` with the reason `the operator's, at close` and does not try it;
  a line the brief covers with a named test, whose command the session
  cannot run, is listed with the reason `covered by <test>`. Both stay
  visible in the count, so the operator sees what the session left to them.
- **The verify skill's word ceiling stays 1500.** It is 1061 words at the
  base; this brief adds about 250.

## Binding references

- `docs/domain/domain-model.md` — S-4: what a session may write (the
  scratch builds stay under `.vloop/tmp/`); M-6: the close shows the newest
  verify report; C-2: a scratch repository is its own root
- `docs/domain/execution/session.md` — the "Verify report — `verify/v1`"
  section, which becomes `verify/v2`'s
- `docs/domain/execution/run.md` — the run folder's `reports/verify.json` line
- `schemas/verify.v1.json` — the shape `verify/v2` extends
- `internal/driver/iterate.go` — the verify step: the report read as
  `verify/v1` and the lines it prints
- `internal/cli/close.go` — `printVerifyReport`, what the close prints
- `plugin/skills/verify/SKILL.md` — the skill this brief changes
- `plugin/skills/create-brief/guidelines.md` — the guideline this brief adds
- `plugin/skills/create-brief/templates/brief-en.md` — the English template
- `plugin/skills/create-brief/templates/brief-es.md` — the Spanish template
- `internal/brief/templates/en.md` — the binary's English template, which
  `internal/cli/plugin_templates_test.go` holds equal to the plugin's outside
  the executive summary
- `internal/brief/templates/es.md` — the binary's Spanish template, held
  equal the same way
- `internal/cli/skills_test.go` — the skills test: each skill's JSON examples
  against their schemas, and the code in a fenced skill (review-brief among
  them) against its phase's fence
- `docs/guide/evals.md` — the grader types and the documented invocation
- `tools/copilot-acceptance/report.py` — the kit's report, which validates a
  kept verify report
- `plugin/skills/review-brief/SKILL.md` — check 4, runnability
- `fence/verify.json` — what the verify session may run: the deny rules the
  guideline names, unchanged
- `internal/budget/budget.go` — the skills' word ceilings, unchanged
- `plugin/evals/verify-planted-mismatches/graders/report-validates.md` — the
  grader that names `verify/v1`
- `plugin/evals/README.md` — the eval catalogue a new case joins

## Behaviour contract

### V1 — `verify/v2`

schemas/verify.v2.json, embedded like the others and listed by
`vloop schema list`: `verify/v1`'s properties and constraints, `schema`'s
const `verify/v2`, and two more required properties:

- `complete`: a boolean;
- `not_run`: an array (empty allowed) of objects with required `line` and
  `reason`, both non-empty strings.

`schemas/verify.v1.json` stays as at the base.

The driver's verify step reads `.vloop/tmp/verify.json` as `verify/v2`,
after a session that exited 0 as at the base, and also after one that timed
out, exited non-zero or left no record (a session that failed to start
leaves nothing to read, as at the base). A report that does not validate as
`verify/v2` (a `verify/v1` report included) is kept by no one: after a
session that exited 0 it is the base's "left no valid" case, with the
base's message; after a failed one, the base's "failed to run" warning
alone. A valid report is kept as `reports/verify.json`, as the base keeps
it, and the driver prints one line:

```
verify: no findings
verify: <N> finding(s) — see <run dir>reports/verify.json
verify: no findings; <M> example line(s) not run — see <run dir>reports/verify.json
verify: <N> finding(s); <M> example line(s) not run — see <run dir>reports/verify.json
```

the first two when `not_run` is empty (the base's lines, unchanged), the
last two when it holds `M` ≥ 1 entries. `<run dir>` is the run folder,
repo-relative with a trailing `/`, as at the base. When `complete` is
`false` and the session exited 0, the driver then warns, on the warning
stream like the base's "verify by hand" lines:

```
vloop: the verify session stopped before its last pass — verify the rest by hand; see <run dir>reports/verify.json
```

After a session that timed out, exited non-zero or left no record, the
line is followed by the base's warning, unchanged:
`vloop: the verify session failed to run — verify by hand; see <run dir>run.log`,
and not by the "stopped before its last pass" warning.

None of this changes the run's ending, its exit code or a task.

`vloop brief close`, when it prints the newest kept verify report (M-6),
prints for a `verify/v2` report:

- the header `verify of <run id>/<stamp>: ` followed by `no findings` or
  `<N> finding(s)`, then, when `not_run` is not empty,
  `; <M> example line(s) not run`;
- the findings as at the base;
- one line per `not_run` entry: `  not run: <line> — <reason>`;
- when `complete` is `false`, the line
  `  the session stopped before its last pass`.

A kept `verify/v1` report prints exactly as at the base.

### V2 — the verify skill

`plugin/skills/verify/SKILL.md`:

- **The report from the start.** As soon as it has read the brief (step
  1), and again after pass 1, after pass 2 and after pass 3, the session
  writes `.vloop/tmp/verify.json` as `verify/v2`, with every finding, count
  and line not run so far, and validates it with
  `vloop schema validate verify/v2 .vloop/tmp/verify.json`. `complete` is
  `false` in every write but the last, which sets it `true`; `notes` says
  which passes remain while it is `false`.
- **A line it does not run.** A worked-example line is not run when the
  brief keeps it for the operator's acceptance, when its command is one the
  verify fence denies (`vloop init`, `vloop run`, `vloop brief review`,
  `claude`, and the rest of `fence/verify.json`'s deny list), when it needs
  a live model (Claude or Copilot), the network or credentials, or when a command it
  needs was denied. The session does not retry a denied command in another
  form. Each such line is one `not_run` entry: `line` quotes the line as the
  brief writes it, `reason` says why in a few words (the denied rule, `needs
  a live model`, `the operator's, at close`, `covered by <test>` when the
  brief names one). It is not counted in `example_lines` and is not a
  finding.
- **Scratch set-up**: one command per tool call; a scratch repository is
  made with `git init -q <dir>`, never `git -C <dir> init` and never
  chained after `mkdir`.
- **Short output**: each command's output is cut to what the session needs
  (through `head` or `tail`); a long output may be saved where the session
  cannot read it back.
- The report example in the skill is a valid `verify/v2` report with one
  `not_run` entry, and every mention of `verify/v1` in the skill becomes
  `verify/v2`.

### V3 — the brief skills

`plugin/skills/create-brief/guidelines.md` gains one rule, in the section
that already says what sessions cannot do:

> **Say which worked-example lines the verify session cannot run.** It runs
> under the verify fence: no `vloop init`, `vloop run`, `vloop brief review`
> or `claude`, no live model (Claude or Copilot), no network, no credentials. Cover
> such a line with an end-to-end test the plan writes with stub agents, and
> name the test beside the line; or move it to the section
> `## Acceptance, at close (the operator's)`, which the operator runs on the
> real tools at close.

Both plugin templates and the binary's two templates
(`internal/brief/templates/en.md`, `internal/brief/templates/es.md`) gain,
between the Worked example and Out of scope, the optional section, with a
placeholder that says what goes there: `## Acceptance, at close (the
operator's)` in the English pair, `## Aceptación, al cierre (del operador)`
in the Spanish pair, the placeholder in each template's language, identical
within each pair so `internal/cli/plugin_templates_test.go` stays green.

`plugin/skills/review-brief/SKILL.md`, check 4 (Runnability), gains one
question: does a worked-example line need what the verify session cannot
do (a command the verify fence denies, a live model, the network,
credentials), with no named test covering it and outside the acceptance
section? Such a line is a `runnability` finding. review-brief is a fenced
skill in the skills test, so the question names no command the
review-brief fence denies as code: it points at fence/verify.json's deny
list instead.

### V4 — the docs and the evals

- `docs/domain/execution/session.md`: the verify report section becomes
  `verify/v2`'s, with `complete` and `not_run` and the driver's warning.
- `docs/domain/execution/run.md` and `docs/guide/concepts.md`: `verify/v1`
  becomes `verify/v2` where they name the report the session writes.
- `tools/copilot-acceptance/README.md`, `verify.sh`'s header comment and
  `report.py`: a kept verify report is validated against the schema its own
  `schema` field names (`verify/v1` or `verify/v2`), and the README and the
  report line say which.
- `plugin/evals/verify-planted-mismatches/graders/report-validates.md`:
  its pattern becomes `vloop schema validate verify/v2`.
- A new eval case, plugin/evals/verify-unrunnable-line/, built like
  `verify-planted-mismatches` (prompt, `case.yaml`, a scaffold that exits
  non-zero when its built result does not hold): a small built result whose
  brief's worked example holds one sound line and one line that runs
  `claude -p "say hi"`. Its graders, of the types `docs/guide/evals.md`
  lists: a `regex` on the trace, `vloop schema validate verify/v2`; a
  `regex` on the file `.vloop/tmp/verify.json`, `"complete": *true`; and an
  `llm` grader whose focus is that file, asking that the `claude` line is
  the one `not_run` entry and that no finding names it. It joins the
  catalogue in `plugin/evals/README.md`.

### Decided here, because the cited documents leave it underdetermined

- `M` counts `not_run` entries, whatever their reasons.
- The driver's lines use `example line(s)` even for one, as the base uses
  `finding(s)`.
- A `not_run` entry whose `line` is not in the brief is not the driver's
  concern; the operator reads it at close.

### Violations the review must rule on

- The driver accepting `verify/v1` from a session, or `verify/v1`'s file
  changed.
- A run's exit code, ending or a task changed by `complete` or `not_run`.
- The skill retrying a denied command in another form, or counting a line
  not run in `example_lines`.
- A fence file or the fence audit changed.
- The verify skill over its ceiling of 1500 words, or the ceiling raised.

## Worked example

Gated with a stub Claude verify session in a fixture repository whose run
completes, as the base's verify tests do. The verify session cannot run
these lines (`vloop run` and `vloop brief close` are in its fence's deny
list): the run lines are covered by the tests in
`cmd/vloop/run_verify_test.go`, the close lines by those in
`internal/cli/close_test.go`, and the schema lines run as written.
`<run dir>` is the run folder the run printed. The stub writes, then exits
0 unless the line says otherwise:

```
verify/v2, complete true, findings [], not_run []
  -> verify: no findings                                                   run exit 0

verify/v2, complete true, one finding, not_run [] 
  -> verify: 1 finding(s) — see <run dir>reports/verify.json               run exit 0

verify/v2, complete true, findings [], not_run [{"line": "vloop init", "reason": "denied by the verify fence (Bash(vloop init:*))"}]
  -> verify: no findings; 1 example line(s) not run — see <run dir>reports/verify.json      run exit 0

verify/v2, complete true, two findings, two not_run entries
  -> verify: 2 finding(s); 2 example line(s) not run — see <run dir>reports/verify.json     run exit 0

verify/v2, complete false, findings [], not_run []
  -> verify: no findings
     vloop: the verify session stopped before its last pass — verify the rest by hand; see <run dir>reports/verify.json      run exit 0

verify/v2, complete false, one finding, not_run [], then the stub exits 3
  -> verify: 1 finding(s) — see <run dir>reports/verify.json
     vloop: the verify session failed to run — verify by hand; see <run dir>run.log      run exit 0
```

Each valid report is kept as `<run dir>reports/verify.json` as the base
keeps it (masked, with a trailing newline). Planted failures:

```
a verify/v1 report (valid as v1)            -> vloop: the verify session left no valid .vloop/tmp/verify.json — verify by hand      run exit 0, nothing kept
verify/v2 without complete                  -> the same line                                                                        run exit 0, nothing kept
verify/v2 with a not_run entry missing reason -> the same line                                                                      run exit 0, nothing kept
no report, then the stub exits 3            -> vloop: the verify session failed to run — verify by hand; see <run dir>run.log        run exit 0, nothing kept
```

`vloop brief close` without `--no-findings` or `--finding`, on the run
whose kept report is the third case above, prints and refuses as at the base,
with:

```
verify of <run id>/<stamp>: no findings; 1 example line(s) not run
  not run: vloop init — denied by the verify fence (Bash(vloop init:*))
```

and on the fifth case: `verify of <run id>/<stamp>: no findings`, then
`  the session stopped before its last pass`. On a kept `verify/v1` report
it prints exactly the base's lines.

Schema fixtures: `vloop schema validate verify/v2 <file>` exits 0 on the
skill's example and on each valid stub report above, and 1 on each planted
report; `vloop schema list` names `verify/v1` and `verify/v2`.

The skills, checked by `go test ./internal/cli/` (the skills test, which
validates each skill's JSON examples against their schemas) and
`go test ./internal/budget/`: the verify skill's example validates as
`verify/v2`; the verify skill is at most 1500 words; the guideline's rule,
both templates' section headings and review-brief's question are present
(a grep each).

## Acceptance, at close (the operator's)

The operator installs a binary built from the branch and runs the two
verify evals with the invocation `docs/guide/evals.md` gives, narrowed with
`--case verify-planted-mismatches` and `--case verify-unrunnable-line`
(`--case` from `claude plugin eval --help`), one run each, and records in
`plugin/evals/README.md` the scores and the cost, as B28 did. The run's own
verify session uses the installed vloop's skill, not this one, and is not
evidence for it.

## Out of scope

- #110: a finished plan of an earlier brief blocking the next run's
  preflight.
- #111: the plan commit keeping what a plan or gate-review session changed
  outside the plan's files.
- #112: patterns in the report schemas for fields the driver checks.
- Any fence change, including an allow rule for `git init` or for the
  toolchain's build command in the verify fence.
- A way for the verify session to run a command that starts an agent
  session (stubs on `PATH`, a fence exception): the brief now covers such a
  line with a test or the operator's acceptance.
- `vloop brief check` warning on a worked-example line the verify session
  cannot run; the review's question is the only check here.
- Re-running B29's verify; the gate-review eval case for B29's T2 gate.

## Constraints

- Go, at the version in `go.mod`; no new dependencies.
- Must keep building and passing on macOS, Linux and Windows.
- **Gates run on macOS's BSD tools**, and **no gate runs `claude`,
  `copilot`, `gh` or touches the network**: stub agents stand in, as in
  B16 and B24–B29.
- Tests build their fixtures in temporary directories and address them with
  `git -C`; sessions never move refs.
- **Gates follow the diff surface; a gate never runs the suite (P-8).**
- **No task may weaken a check to pass.** An earlier test changes only where
  it pins behaviour the contract changes (the stub verify reports of
  `cmd/vloop/run_verify_test.go`, `cmd/vloop/run_iterate_test.go` and
  `cmd/vloop/b16_e2e_test.go`, which become `verify/v2`; the schema lists of
  `internal/cli/schema_test.go`, `internal/schema/schema_test.go` and
  `cmd/vloop/b2_e2e_test.go`, which gain `verify/v2`; the kept `verify/v1`
  reports of `internal/cli/close_test.go` stay and gain `verify/v2`
  siblings), and only to the new behaviour; the work session names
  the contract item in its notes and the review cites it.
- **New tooling is proven by fixtures, including where it must fail**: every
  line of the worked example and each planted failure.
- Repo-relative paths everywhere. No absolute paths in any file or commit
  message.

## Shape

4 to 5 tasks, each verifiable by a single command. Order is load-bearing:

1. **V1 — the schema** `verify/v2`, embedded and listed, with valid and
   invalid fixtures. Gated on the fixtures.
2. **V1 — the driver and the close**: the verify step reads `verify/v2` and
   prints its lines; `vloop brief close` prints `not_run` and `complete`;
   the base's verify tests' stub reports become `verify/v2`. Gated with the
   stub verify session, every worked-example line and planted failure.
3. **V2 — the verify skill**, its example a valid `verify/v2` report, under
   its ceiling. Gated on the skills test and the ceiling.
4. **V3 and V4 — the brief skills, the docs and the evals.** Prose: if it
   fails review twice on wording, block it and the operator finishes it.
   Gated on the skills test, the eval-shape test and a grep per item.
5. **Close:** every gate at once, and the worked example line for line.

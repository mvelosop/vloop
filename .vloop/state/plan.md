# Plan — B20261014-1138-b15-code-health-escapes

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **13/14 done** · iteration 23

**Brief:** `docs/briefs/B20261014-1138-b15-code-health-escapes.loop-brief.md` · **Updated:** 2026-10-14T16:52:41Z

## Progress

- [x] **T1** — Make cmd/vloop's TestMain fail cleanly when the binary does not build
- [x] **T2** — Run every git command through one package
- [x] **T3** — Filter the environment for gates, checks and sessions with one function
- [x] **T4** — One plan-at-the-last-commit function, one task count, one table of budget keys
- [x] **T5** — Split the five long functions to at most 100 lines each
- [x] **T6** — Compare run.git-config-ignore patterns as git reports keys
- [x] **T7** — Give new defects and interventions second-stamped IDs that never collide · 2 attempt(s)
- [x] **T8** — Show test:code and $/kloc per task, and name the cost per 1,000 code lines $/kloc everywhere · 1 attempt(s)
- [x] **T9** — Refuse vloop defect add for a brief that is not merged
- [x] **T10** — Count a recorded defect as escaped when it was created after its brief merged
- [x] **T11** — Record operator findings as verification-finding interventions with an origin · 1 attempt(s)
- [x] **T12** — Retire the defect triage: rule, rule-outcome and close's defect report
- [x] **T13** — Describe the new model in the docs, the domain and both operate playbooks · 1 attempt(s)
- [ ] **T14** — Close: commit the worked example as an end-to-end test and pass the real-data check

## Tasks

### T1 — Make cmd/vloop's TestMain fail cleanly when the binary does not build

`done` · depends on: none

TestMain in cmd/vloop builds the vloop binary the end-to-end tests run; when that build fails it panics today, burying the compiler's message under a goroutine trace. This task comes first so that any build break later in this run reads as the compiler wrote it. It changes no product behaviour.

**Acceptance**

- When the build fails, TestMain prints the go build output, the compiler's error included, to standard error, removes its temporary directory and exits 1; it never calls panic (a panic exits 2 with a goroutine trace, which is the violation).
- When the temporary directory cannot be created, TestMain prints the error to standard error and exits 1, again without panic.
- When the build succeeds, the tests run exactly as before and the temporary directory is removed after m.Run; no Test function changes.
- Only TestMain (and any small helper it needs) in cmd/vloop changes; no product code changes. This task ships no product behaviour, so it adds no test of its own; the gate plants the build break through a Go overlay and is the check.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T1/gate.sh
```

</details>

### T2 — Run every git command through one package

`done` · depends on: T1

Fourteen call sites run git with their own exec.Command, each with its own trimming and error text, so a change to how vloop runs git has to be made many times. One package becomes the only place that runs git; every caller goes through it and keeps exactly the output and errors it has today. The driver's hardening is security-relevant: it must apply exactly where it does today, no more and no less.

**Acceptance**

- Exactly one package calls exec.Command or exec.CommandContext with "git"; no other non-test Go file under internal/ or cmd/ does (internal/cli, internal/defect, internal/driver, internal/intervention, internal/metrics and internal/runs all go through it).
- The driver's hardening — -c core.hooksPath=<the recreated empty hooks dir>, -c core.fsmonitor=false, -c core.quotepath=false — applies to exactly the commands gitCmd builds today and to no other; gitPath's `rev-parse --git-path` and configEntries' `git config --file <path> --no-includes --list -z` stay plain. A hardening flag added to or dropped from any call is a violation.
- Each caller keeps its trimming and its error text exactly: runGit's `git <subcommand>: <stderr>` and `git <subcommand>: <err>`, blame's `git <args>: <err>: <stderr>`, the callers that return raw bytes, TrimSpace or TrimRight of newlines, and those that only test the exit status.
- No command's output, exit code, files written or session arguments change, and no earlier test changes (H1-H7: the unchanged tests are the behaviour check).
- The runner package carries unit tests of its own for what it centralises (at least: the hardened form's arguments and the plain form's absence of them, and an error's text), run by the checks.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T2/gate.sh
```

</details>

### T3 — Filter the environment for gates, checks and sessions with one function

`done` · depends on: T2

Four copies of the same loop drop vloop's own variables from an environment before a gate, a base gate, a check or a session starts. One function that removes a given list of variable names replaces them, so the set of dropped variables is stated at each call and implemented once.

**Acceptance**

- One function removes a given list of variable names from an environment ([]string of NAME=value), matching NAME= exactly so that a longer name sharing the prefix is kept; the gate run (runGate), the base gates (baseGates), the checks (runCheck) and the session launch (Session.Run) call it, and no other function filters an environment by matching "<NAME>=" prefixes.
- Each caller keeps its base environment and the set it drops today: the gate run its injected base else os.Environ, the base gates and the checks os.Environ, all three dropping VLOOP_ACTIVE_TASK and VLOOP_GATE_TASK; the session launch also drops VLOOP_PLAN_SHA256 (PlanHashEnv).
- The difference that the base gates and the checks start from the process environment while the gate run honours an injected one is kept as is and reported in the task notes, not fixed.
- No command's output, exit code, files written or session arguments change, and no earlier test changes (H1-H7: the unchanged tests are the behaviour check).
- The function has a committed unit test: the named variables are gone, every other entry is kept in order, and a variable whose name merely starts with a dropped name is kept.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T3/gate.sh
```

</details>

### T4 — One plan-at-the-last-commit function, one task count, one table of budget keys

`done` · depends on: T3

Three small duplications from the v1 review: the rule for which commit a brief's plan is read at, the driver's own copy of the plan's task counts, and the run.* budget keys listed in several places. Each becomes one definition the other places call, so the next change to any of them is made once.

**Acceptance**

- H3: one function returns the commit a brief's plan is read at — the run commit, else the last task commit, else the plan commit; defect.CheckTask, the metrics (every current caller of metrics.planSHA) and internal/cli/defect.go's defectMatrix all call it, and no other function picks among .Plan.SHA, the last task's SHA and .Run.SHA itself.
- H4: the driver's counts function is removed; the driver uses state.Plan's Counts(), deriving pending as the task count minus done and blocked, with identical results.
- H5: the run.* budget keys (run.max-iterations, run.cost-ceiling, run.max-attempts, run.stall-limit, run.convergence-max, run.convergence-min, run.gate-timeout, run.session-timeout, run.keep-awake) are named in one table besides config.Keys; ResolveBudgets, the `vloop run` budget flags and the two stray run.gate-timeout reads (internal/driver/plan.go, internal/cli/task_gate.go) read it. No key's string literal appears more than twice in non-test Go.
- The `vloop run` flags keep their names, help text, order and precedence (flag, then VLOOP_RUN_*, then the file, then the default).
- No command's output, exit code, files written or session arguments change, and no earlier test changes (H1-H7: the unchanged tests are the behaviour check).
- Each new shared function or table has a committed unit test of its own (the commit choice for each of the three cases; the table naming every key ResolveBudgets resolves).

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T4/gate.sh
```

</details>

### T5 — Split the five long functions to at most 100 lines each

`done` · depends on: T4

iterate, Plan, runInit, runDoctor and runBriefClose are 145 to 296 lines and every later brief touches them. Each is split into named steps so that none exceeds 100 lines, after H1-H5 have already shrunk some of them. Later tasks in this run change runBriefClose and must keep it within the bound, which this task's gate keeps checking.

**Acceptance**

- Iterator.iterate (internal/driver/iterate.go), Planner.Plan (internal/driver/plan.go), runInit (internal/cli/init.go), runDoctor (internal/cli/doctor.go) and runBriefClose (internal/cli/close.go) each exist once under that name and receiver and are at most 100 lines, counted from the func line to its closing brace.
- The split extracts cohesive steps into named helpers; it does not pad, inline-collapse statements onto one line, or move logic into closures declared elsewhere just to pass the count.
- The functions the brief leaves out (Iterator.Run, intervention.parse, newRun, metrics.Aggregate, runBriefCheck) are not touched beyond what an extraction strictly needs.
- No command's output, exit code, files written or session arguments change, and no earlier test changes (H1-H7: the unchanged tests are the behaviour check).

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T5/gate.sh
```

</details>

### T6 — Compare run.git-config-ignore patterns as git reports keys

`done` · depends on: T5

git reports a config key's section and variable in lower case and keeps a subsection's case, but vloop compares patterns literally, so a pattern like Core.* is accepted and silently matches nothing (D20261011-1458). Patterns are compared the way git spells keys, and the protected-section refusal sees the lower-cased form, so Core.* is refused like core.*.

**Acceptance**

- A run.git-config-ignore pattern is compared with keys after lower-casing its text before the first "." and after the last ".", keeping a subsection between them as given (a pattern with no "." is lower-cased whole); lower-case patterns behave exactly as today.
- The protected-section refusal applies to that lower-cased form: `vloop config set run.git-config-ignore 'Core.*'` exits 2 with exactly `vloop: run.git-config-ignore: Core.* would ignore core.* — the guard must see it`, naming the pattern as given; `core.*` gives the message it gives today.
- `Branch.Main.Foo-Bar` is accepted, and during `vloop run` a session setting branch.Main.foo-bar does not halt the run, while one setting branch.main.foo-bar halts it with exit 9 as today.
- The same comparison applies wherever the patterns come from (config set, the file, VLOOP_RUN_GIT_CONFIG_IGNORE) and to the run's start-up refusal of a bad source; the built-in pattern and the guard's halt message are unchanged.
- Committed unit tests (internal/gitcfg, and the config refusal) cover a mixed-case protected pattern refused, a mixed-case subsection pattern matching only its own subsection's case, and lower-case patterns unchanged.
- No record under .vloop/ is edited; the defect's status is the operator's to set.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T6/gate.sh
```

</details>

### T7 — Give new defects and interventions second-stamped IDs that never collide

`done` · 2 attempt(s) · depends on: T6

Ten records share D20261004-1142 today because IDs carry minutes and collisions get -2, -3 suffixes. New recorded defects and interventions get a stamp with seconds that no other record of their kind has, made by one generator that `brief close --dry-run` also predicts with. Every command that reads an ID keeps accepting the old form, because existing IDs are never renamed.

**Acceptance**

- A new recorded defect is D<YYYYMMDD-HHMMSS>-<slug> and a new intervention I<YYYYMMDD-HHMMSS>-<slug>, in local time as today.
- When any record of the kind already uses the stamp, whatever its slug, the stamp moves on one second until it is free; the -2, -3 suffixes are no longer made; creating the file stays atomic (exclusive create, moving on when the file exists).
- One generator makes both kinds of ID; `vloop brief close --dry-run` predicts the IDs of its pending findings with it, the IDs taken by earlier findings of the same dry run included, replacing the copy of the ID code in closeDryRun.
- Every command that reads or names an ID accepts both forms, old and new: intervention show's ID patterns (its links), doctor's guidelines provenance check, defect set and intervention set.
- No existing record is renamed or edited.
- internal/defect/defect_test.go and internal/intervention/intervention_test.go change only for the second-stamped IDs and their allocation; committed tests cover a collision across different slugs, the one-second move, and both ID forms resolved by intervention show and accepted by the provenance check.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T7/gate.sh
```

</details>

### T8 — Show test:code and $/kloc per task, and name the cost per 1,000 code lines $/kloc everywhere

`done` · 1 attempt(s) · depends on: T7

The per-task table shows each task's lines and cost but not the two ratios the operator reads them for. It gains test:code and $/kloc, the JSON gains the same as optional fields, and the summary and all-briefs table use the same two names so one name means one thing.

**Acceptance**

- `vloop metrics <brief> --by task` gains two columns between other+ and agent: test:code (the task's test+ / code+) and $/kloc (the task's cost per 1,000 of its code+ lines, no $ sign), both with two decimals and both `-` where code+ is 0. A task with code+ 200, test+ 300 and cost 0.50 shows 1.50 and 2.50.
- `--json` by_task entries gain test_code_ratio and per_1000_code_lines_usd, null where the task's code lines are 0; schemas/metrics.v2.json gains both as optional properties of a by_task item, and the output still validates against metrics/v2.
- The summary's cost line ends ` · $<c>/kloc` instead of ` · $<c> per 1,000 code lines`; its figure still uses delivered code lines.
- The all-briefs table's headers t:c and $/1k become test:code and $/kloc; the values are unchanged.
- The earlier tests that change are only internal/metrics/print_test.go, internal/metrics/lines_test.go, internal/cli/metrics_test.go, cmd/vloop/b3_e2e_test.go and cmd/vloop/run_iterate_test.go, and only by the two per-task columns, the cost line's `$<c>/kloc` and the all-briefs headers; committed tests cover both columns, the `-` and null cases and the new summary ending.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T8/gate.sh
```

</details>

### T9 — Refuse vloop defect add for a brief that is not merged

`done` · depends on: T8

A recorded defect now means an escape: a problem that got past the brief's merge. What the operator finds before merge is an intervention instead, so `defect add` refuses a brief that is not merged (M-5) and says how to record the intervention. This is enforced where the record is made, because a convention in the docs is how 28 escapes went uncounted.

**Acceptance**

- `vloop defect add` refuses, before writing anything, when the brief (from --brief or attributed by --blame) is not merged (M-5), exit 1, stdout empty, stderr exactly `vloop: <brief name> is not merged: record what you found before merge as an intervention — vloop intervention add "<summary>" --brief <brief name> --kind verification-finding --origin <brief|plan|work|env> --phase verify --automatable partly --by operator`, where <brief name> is the brief's name (as B20260101-0900-a.loop-brief) and the rest is printed as quoted.
- When there is no default branch, it refuses with the error M-5's lookup gives today (`vloop: no default branch: origin/HEAD, main and master are all missing`), exit 1, writing nothing.
- A merged brief is recorded exactly as before; found-by keeps its four values and stays required.
- The release lookup runs git through the one runner (T2).
- The tests that run defect add against a brief that is not merged — cmd/vloop/b14_e2e_test.go, cmd/vloop/b20261005_e2e_test.go, cmd/vloop/b3_e2e_test.go, internal/cli/defect_test.go, internal/cli/export_test.go, internal/cli/intervention_test.go, internal/cli/matrix_test.go, internal/cli/metrics_test.go — change only by merging the brief in their fixture first, or by asserting the refusal; committed tests cover the refusal by --brief and by --blame, the no-default-branch refusal and a merged brief accepted.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T9/gate.sh
```

</details>

### T10 — Count a recorded defect as escaped when it was created after its brief merged

`done` · depends on: T9

found-by alone decided whether a defect escaped, so an operator who found a defect a week after the merge was counted as catching it: 28 escapes were counted as caught and removal efficiency overstated the loop. A recorded defect is escaped when created after its brief's release commit, whoever found it; the history is recomputed, not edited (M-1).

**Acceptance**

- A recorded defect is escaped when its created time is after the committer time of its brief's release commit (M-5), whatever its found-by; a record created in the same second is not escaped. Otherwise it counts by found-by as today: gate and review in-loop, operator operator, user escaped. A brief that is not merged has no release, so its records count by found-by.
- The rule applies to every record, old and new, everywhere the counts are computed (vloop metrics' summary and JSON, the all-briefs table, close's summary and snapshot); no record is edited or migrated.
- Removal efficiency stays (in-loop + operator) / all; `defect list --matrix` is unchanged (origin by catcher).
- The release commit's time is read through the one git runner (T2).
- Real-data check, on this repository: the task notes give, for each consumed brief, the `vloop metrics` defects line before and after; only operator and escaped move, only for briefs with recorded defects created after their merge, and the total escaped across them rises from 3.
- internal/metrics/defects_test.go changes only by the time rule; committed tests cover a defect before, at the same second as, and after the release, by each catcher, and an unmerged brief.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T10/gate.sh
```

</details>

### T11 — Record operator findings as verification-finding interventions with an origin

`done` · 1 attempt(s) · depends on: T10

What the operator finds before merge is no longer a defect but an intervention that says whose defect it is (brief, plan, work or env). `intervention add` gains --origin for verification findings, `brief close --finding "<origin>: <summary>"` records those interventions, and the metrics count them as caught by the operator, alongside the operator defects that did not escape, without counting anything twice.

**Acceptance**

- `vloop intervention add` gains --origin brief|plan|work|env: required with --kind verification-finding (else exit 2, `vloop: --kind verification-finding needs --origin brief, plan, work or env`) and refused with any other kind (exit 2, `vloop: --origin is only for --kind verification-finding`), before anything is written.
- The record gains `origin: <v>` right after `by:`; `intervention show --json` (and list --json) carry it as origin; schemas/intervention.v2.json gains origin as an optional property; `vloop intervention set <id> origin <v>` sets it. Records without it stay valid and unchanged.
- `vloop brief close --finding "<origin>: <summary>"` records one intervention per finding: --brief the closed brief, phase verify, kind verification-finding, automatable partly, by operator, the origin, and the text after `: ` as its summary; close prints `recorded .vloop/interventions/<id>.md` for each where it printed defect files, and commits them; no defect file is written for a finding.
- A finding not starting with one of the four origins (lower case only) followed by `: ` is refused before anything is written, exit 2: `vloop: a finding starts with whose defect it is: --finding "<brief|plan|work|env>: <summary>"`.
- The metrics' operator count is the brief's recorded operator defects not escaped (T10) plus its verification-finding interventions that have an origin; an intervention without an origin, or of another kind, is not counted, and a defect and an intervention are each counted once. Removal efficiency stays (in-loop + operator) / all.
- The run record lists those interventions after the defects, one line each: `- <id> — verification-finding, <origin>, found by operator: <summary>`; `--dry-run` counts them in its summary and lists them as `would record .vloop/interventions/<id>.md`, with the IDs the real close would give.
- runBriefClose stays within 100 lines (T5's gate keeps checking); git runs only through the runner.
- The earlier tests that change are only internal/cli/close_test.go, internal/cli/close_flags_test.go, cmd/vloop/b4_e2e_test.go and cmd/vloop/b14_e2e_test.go (findings taking an origin, recorded as interventions), internal/metrics/defects_test.go (the interventions in the operator count) and the commands.md staleness test (regenerate docs/guide/commands.md); committed tests cover --origin's two refusals, the record's origin line, close's refusal and recording, the operator count's two sources without overlap and the run record line.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T11/gate.sh
```

</details>

### T12 — Retire the defect triage: rule, rule-outcome and close's defect report

`done` · depends on: T11

B14 added a triage of each defect (the rule that caught it and what it did) and a defect report at close; with defects narrowed to escapes it costs more than it tells. The setters and the report go; records that already carry the fields stay valid and are shown as they are.

**Acceptance**

- `vloop defect set` no longer accepts rule or rule-outcome: each is refused as any unknown field is, exit 2, with `vloop: cannot set "<field>": want one of status, fixed-by, case, severity, origin, kind, task`, leaving the record byte-identical; defect set's help lists the same fields.
- Records that have rule and rule-outcome stay valid and are shown as they are (defect list and its --json keep both fields); schemas/defect.v1.json keeps both properties.
- `vloop brief close` and its --dry-run no longer print `defects of <run id>`, the per-defect report lines or the untriaged hint; `squash-merge with the trailer: Vloop-Brief: <name>` stays close's last line.
- doctor's guidelines provenance check is unchanged: a rule still cites a defect or intervention ID.
- runBriefClose stays within 100 lines (T5's gate keeps checking).
- The earlier tests that change are only internal/cli/close_test.go, internal/cli/close_flags_test.go, cmd/vloop/b4_e2e_test.go, cmd/vloop/b14_e2e_test.go (the defect report going), internal/defect/defect_test.go (the retired setters) and the commands.md staleness test; a committed test asserts that rule and rule-outcome are refused and close prints no report.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T12/gate.sh
```

</details>

### T13 — Describe the new model in the docs, the domain and both operate playbooks

`done` · 1 attempt(s) · depends on: T12

The code now records findings before merge as interventions, counts escapes by time, stamps IDs with seconds and has no triage; the docs, the domain model and the two operate playbooks still describe the old model. They are brought in line, and docs/guide/commands.md is regenerated. The operate skill must stay within its word ceiling.

**Acceptance**

- docs/domain/domain-model.md: M-4 decides escaped by time against M-5's release and cites M-5 (the sentence "every other catcher is before it" goes); M-6's commit carries intervention files, not defect files; the Identifiers table shows `D<YYYYMMDD-HHMMSS>-<slug>` and `I<YYYYMMDD-HHMMSS>-<slug>` and says old IDs stay.
- docs/domain/measurement/defect.md: a recorded defect is an escape; the time rule; the "The rule that caught it" section is removed, leaving one line saying the fields were retired in B15 and stay valid in old records; no mention of untriaged.
- docs/domain/measurement/metrics.md and docs/guide/metrics.md: the two per-task columns, `$/kloc`, the renamed all-briefs headers (no t:c, $/1k or "per 1,000 code lines" left), the operator count's two sources (verification-finding interventions with an origin named), and escaped by time.
- docs/guide/defects.md: what to record as a defect (escapes, with --blame) and what as an intervention (anything found before merge, `--kind verification-finding --origin …`); no untriaged.
- plugin/skills/operate/SKILL.md: the close step's --finding takes its origin (`--finding "<brief|plan|work|env>: …"`); the triage step is removed (no rule-outcome, untriaged or `defects of` left); it stays within 1,750 words and is not changed beyond D1.
- .claude/skills/vloop-operator/SKILL.md, section 5 (Closing and merging): its "Record defects" item says to record findings before merge as verification-finding interventions with --origin, and defects only for escapes.
- docs/guide/commands.md is regenerated and equals what cmd/gendocs makes now.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T13/gate.sh
```

</details>

### T14 — Close: commit the worked example as an end-to-end test and pass the real-data check

`pending` · depends on: T13

The last task proves the brief as written: the worked example line for line, against a fresh build in temporary repositories, committed as an end-to-end test the checks keep running, and the real-data check on this repository. Anything the worked example exposes is fixed here.

**Acceptance**

- A new end-to-end test in cmd/vloop, in the style of the earlier b<n>_e2e_test.go files, runs the brief's worked example with a fresh build in temporary repositories and a temporary HOME addressed with git -C: the git-config-ignore lines, defect add on an unmerged and a merged brief (with the same-second stamp), the metrics defects line, the intervention add refusals and origin, the defect set rule refusal, the close with a refused and an accepted finding and its run record line, and the per-task test:code and $/kloc values.
- Real-data check on this repository, in the task notes: `vloop metrics` for each consumed brief, the defects lines before and after (only operator and escaped move); every other summary line unchanged but the cost line's ending; `vloop defect list` and `vloop intervention list`/`show` list every existing record.
- No historical record is edited, renamed or migrated, and nothing is written under .vloop/ but .vloop/tmp/.
- Fixes needed to make the worked example hold are made here, each with its committed test, and named in the notes.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T14/gate.sh
```

</details>


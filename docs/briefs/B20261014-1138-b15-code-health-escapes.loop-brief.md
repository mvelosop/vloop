---
name: B20261014-1138-b15-code-health-escapes.loop-brief
description: Pay down the v1 review's code-health findings with no behaviour change, and narrow recorded defects to escapes — operator findings become interventions saying whose defect it is, escaped is counted by time against the merge, the defect triage retires — plus unique second-stamped record IDs, a case-correct git-config-ignore, and test:code and $/kloc per task
kind: brief
status: ready
created: 2026-10-14
seeds: The roadmap's B15 row as rewritten after B14's close; the verification phase follows as B16
depends-on: [B20261011-2251-b14-learning-loop.loop-brief]
---
# Brief — vloop B15: code health, and defects narrowed to escapes

- **Starting point:** extends `main` after this design merges (`main` at
  `49dc17c`, B14 merged, the roadmap rewritten). The planner pins the base it
  plans from as the **base** every gate compares against — never `HEAD~n`.
- **Produced by:** the roadmap's B15 row
  (`docs/design-notes/vloop-roadmap-b12-b17.md`), the v1 review's code-health
  findings, and the design act of 2026-10-14, with the forks settled with the
  operator (the decision to narrow defects is recorded as
  `I20261014-1110-defects-narrowed-to-escapes-operator-fin`).
- **Run by:** `vloop run` from the latest released vloop.

## Executive summary

The code the loop runs on has grown duplicated and long: 14 places run `git`,
four copies filter the environment, `iterate` is 296 lines. Separately, the
defect records cost more than they tell: B14's three operator defects each
duplicated an intervention, their triage needed a re-rendered run record, and
28 of the 62 operator defects were in fact recorded after their brief merged —
escapes the metrics count as caught.

**The goal:** the code is easier to change safely, and a recorded defect means
one thing — a problem that got past the merge — so the metrics measure what
matters: the defects that reach users.

- **H1–H7 — code health, no behaviour change.** One way to run `git`, one
  environment filter, one "plan at the last commit", one task count, one table
  of budget keys, the five long functions split, and a test harness that fails
  cleanly. Every later brief touches this code.
- **G1 — `run.git-config-ignore` reads patterns as git does.** A pattern like
  `Core.*` is accepted today and silently matches nothing (`D20261011-1458`);
  after, it is refused as `core.*` is.
- **S1 — record IDs carry seconds and never collide.** Ten records share
  `D20261004-1142` today; new defects and interventions get a stamp with
  seconds that no other record of their kind has.
- **M1 — `test:code` and `$/kloc` per task.** The per-task table shows each
  task's test-to-code ratio and cost per 1,000 code lines, next to its lines.
- **E1 — a recorded defect is an escape.** `vloop defect add` refuses a brief
  that is not merged yet and says to record an intervention instead.
- **E2 — escaped is counted by time.** A recorded defect created after its
  brief merged is escaped, whoever found it; the history is corrected.
- **E3 — operator findings are interventions.** `brief close --finding`
  records a `verification-finding` intervention that says whose defect it is
  (brief, plan, work or env); the metrics count it as caught by the operator.
- **E4 — the triage retires.** The `rule` and `rule-outcome` fields, close's
  defect report and the operate skill's triage step go; old records stay valid.
- **D1 — the docs, the domain and both operate playbooks** describe the new
  model.

Unchanged: no exit code meaning, no schema version (two schemas gain optional
properties), the derived in-loop defects, every existing record's validity.
The Windows run harness and the newer long functions wait.

## What it is

vloop's internals lose the duplication the v1 review found, with every test
passing unchanged except where this brief names it. A recorded defect becomes
what escaped a brief's merge; what the operator finds before merge is an
intervention that names its origin, counted as caught by the operator; and
escaped is decided by when a defect was recorded, not by who found it.

The blind spot: `found-by` was the only thing deciding whether a defect
escaped, but the operator who finds a defect a week after the merge is still
`operator`. Twenty-eight escapes were counted as caught that way, and the
removal efficiency of B1–B10 overstated the loop.

## Why this shape, and what was rejected

Decided with the operator on 2026-10-14. **Do not re-litigate these** — the
planner inherits them.

- **Defects only for escapes.** Escapes are the end-user defects worth
  optimizing, and only a defect record captures them; in-loop defects are
  derived for free, and the operator's findings duplicated interventions.
  *Rejected:* dropping defects entirely (an escape found months later is not
  an intervention, and found-before versus escaped is lost); keeping
  everything until B16 (B15 and B16 would build more on the triage).
- **The operator count reads both sources, without overlap.** `operator` is
  the recorded operator defects created before their brief merged plus the
  `verification-finding` interventions that carry an origin. No existing
  intervention has an origin, so nothing is counted twice. *Rejected:*
  interventions only (every consumed brief's operator count would drop to 0);
  dropping the column (removal efficiency would ignore what verification
  caught).
- **`defect add` refuses before merge.** The rule is enforced where the record
  is made. *Rejected:* a convention in the docs with `found-by: user` as the
  only escape (it is how 28 escapes went uncounted).
- **Escaped by time, history included.** One rule for old and new records; the
  numbers that move are the ones that were wrong. *Rejected:* applying it only
  to records made after B15 (a cut-over date in the rule, and the history
  keeps undercounting).
- **The five long functions named at v1, at most 100 lines each.** Bounded.
  *Rejected:* a ceiling for every function with a test (larger, and one more
  rule); dropping them.
- **The Windows run harness waits** for a runner that can execute it: no one
  runs vloop on Windows yet, and a harness nobody here can run cannot be
  verified. *Rejected:* a Go stub verified only on macOS.
- **`test:code` and `$/kloc`**, the names the operator chose; the summary and
  the all-briefs table use them too, so one name means one thing.

## Binding references

- `docs/design-notes/vloop-v1-review.md` — "Code health" and "Tests": the findings H1–H7 implement, and the TestMain finding H7 fixes
- `docs/domain/domain-model.md` — M-4, M-5 and M-6 (the catcher, the release E1 and E2 test against, closing's commit, which E2–E4 change), the Identifiers table (S1), S-2 (the git-config guard G1 changes), M-1 (metrics recomputed from raw sources)
- `docs/domain/measurement/defect.md` — the defect record, its axes and "The rule that caught it", which E1–E4 narrow and retire
- `docs/domain/measurement/metrics.md` — the summary, the per-task table and the metrics/v2 JSON, which M1 and E2 extend
- `docs/guide/defects.md` — the defect workflow D1 rewrites
- `.vloop/defects/D20261011-1458-run-git-config-ignore-accepts-a-mixed-ca.md` — G1's defect
- `.vloop/interventions/I20261014-1110-defects-narrowed-to-escapes-operator-fin.md` — the decision E1–E4 implement, with the options weighed
- `plugin/skills/operate/SKILL.md` — the close and triage steps E3 and E4 change

## Behaviour contract

Messages are exact where quoted; `<…>` marks a value.

### H1–H7 — code health, no behaviour change

**Required.** Every command's output, exit codes, files written and session
arguments stay as they are. No earlier test changes for H1–H7.

- **H1 — one git runner.** One package runs `git`; no other non-test file
  calls `exec.Command` or `exec.CommandContext` with `"git"`. The driver's
  hardening (`core.hooksPath`, `core.fsmonitor=false`, `core.quotepath=false`)
  stays exactly where it applies today, and the plain call `gitPath` makes
  stays plain. Each caller keeps the trimming and error text it has today.
- **H2 — one environment filter.** One function removes a given list of
  variable names from an environment; the four copies (the gate run, the base
  gates, the checks, the session launch) call it. Each keeps its base
  environment and the set it drops today: the three gate and check callers
  drop `VLOOP_ACTIVE_TASK` and `VLOOP_GATE_TASK`; the session launch also
  drops `VLOOP_PLAN_SHA256`.
- **H3 — the plan at the last commit, once.** `defect.CheckTask` and the
  metrics use one function for "the run commit, else the last task commit,
  else the plan commit".
- **H4 — one task count.** The driver's `counts` goes; it uses the plan's
  `Counts`, with pending derived from it.
- **H5 — one table of budget keys.** The `run.*` budget keys are named once;
  `ResolveBudgets`, the `vloop run` flags and the two stray `run.gate-timeout`
  literals read it.
- **H6 — five long functions.** `iterate`, `Plan`, `runInit`, `runDoctor` and
  `runBriefClose` are at most 100 lines each, counted from the `func` line to
  its closing brace.
- **H7 — the test binary fails cleanly.** `TestMain` in `cmd/vloop` prints the
  build failure to standard error, removes its temporary directory and exits 1;
  it never panics.

### G1 — `run.git-config-ignore` matches as git does

**Required.**
- git reports section and variable names in lower case and keeps a
  subsection's case. A pattern is compared that way: its text before the first
  `.` and after the last `.` is lower-cased; a subsection between keeps its
  case.
- The protected-section refusal applies to the lower-cased form, so
  `vloop config set run.git-config-ignore 'Core.*'` exits 2 with
  `vloop: run.git-config-ignore: Core.* would ignore core.* — the guard must see it`,
  the message `core.*` gets today, naming the pattern as given.
- `Branch.Main.Foo-Bar` ignores the key `branch.Main.foo-bar`;
  `branch.main.foo-bar` it does not (the subsection's case differs).
- Lower-case patterns behave exactly as today.

### S1 — unique record IDs with seconds

**Required.**
- A new recorded defect is `D<YYYYMMDD-HHMMSS>-<slug>`, a new intervention
  `I<YYYYMMDD-HHMMSS>-<slug>`, in local time as today.
- No two records of a kind share a stamp: when one already uses it, whatever
  its slug, the stamp moves on one second until it is free. The `-2`, `-3`
  suffixes are no longer made. Creating the file stays atomic.
- One generator makes both; `brief close --dry-run` predicts with it.
- Every command that reads or names an ID accepts both forms, old and new:
  `intervention show`'s ID patterns, the guidelines provenance check, `defect
  set` and `intervention set` lookups.
- Existing IDs are never renamed.

### M1 — `test:code` and `$/kloc`

**Required.**
- The per-task table gains two columns between `other+` and `agent`:
  `test:code` (the task's `test+` / `code+`, two decimals) and `$/kloc` (the
  task's cost per 1,000 of its `code+` lines, two decimals, no `$`). Both are
  `-` where `code+` is 0.
- `--json`'s `by_task` entries gain `test_code_ratio` and
  `per_1000_code_lines_usd`, `null` where the task's code lines are 0;
  `schemas/metrics.v2.json` gains them as optional properties.
- The summary's cost line ends `$<c>/kloc` instead of
  `$<c> per 1,000 code lines`.
- The all-briefs table's headers `t:c` and `$/1k` become `test:code` and
  `$/kloc`; the values are unchanged.

### E1 — a recorded defect is an escape

**Required.**
- `vloop defect add` refuses, before writing anything, when the brief (from
  `--brief` or `--blame`) is not merged (M-5), exit 1:
  `vloop: <brief name> is not merged: record what you found before merge as an intervention — vloop intervention add "<summary>" --brief <brief name> --kind verification-finding --origin <brief|plan|work|env> --phase verify --automatable partly --by operator`
- When there is no default branch it refuses with the error M-5's lookup gives
  today, exit 1.
- `found-by` keeps its four values and stays required.

### E2 — escaped is counted by time

**Required.**
- A recorded defect is **escaped** when its `created` time is after the
  committer time of its brief's release commit (M-5), whatever its
  `found-by`. Otherwise it counts by `found-by` as today: `gate` and `review`
  in-loop, `operator` operator, `user` escaped.
- This applies to every record, old and new; no record changes.
- `defect list --matrix` is unchanged: it shows origin by catcher.

### E3 — operator findings are interventions

**Required.**
- `vloop intervention add` gains `--origin brief|plan|work|env`. It is
  required with `--kind verification-finding` and refused with any other
  kind, exit 2:
  `vloop: --kind verification-finding needs --origin brief, plan, work or env` and
  `vloop: --origin is only for --kind verification-finding`.
  The record gains `origin:` after `by:`; `schemas/intervention.v2.json` gains
  it as an optional property, and `intervention set <id> origin <v>` sets it.
- `vloop brief close --finding "<origin>: <summary>"` records one intervention
  per finding: `--brief` the closed brief, `phase: verify`,
  `kind: verification-finding`, `automatable: partly`, `by: operator`, the
  origin, and the text after `: ` as its summary. A finding without one of the four origins before
  `: ` is refused before anything is written, exit 2:
  `vloop: a finding starts with whose defect it is: --finding "<brief|plan|work|env>: <summary>"`.
  `close` prints `recorded .vloop/interventions/<id>.md` for each, where it
  prints the defect files today.
- The metrics' `operator` count is the brief's recorded `operator` defects not
  escaped (E2) plus its `verification-finding` interventions that have an
  origin. Removal efficiency stays (in-loop + operator) / all.
- The run record lists those interventions after the defects, one line each:
  `- <id> — verification-finding, <origin>, found by operator: <summary>`.
- `--dry-run` counts and lists the pending findings the same way.

### E4 — the triage retires

**Required.**
- `vloop defect set` no longer accepts `rule` or `rule-outcome`; they are
  refused as any unknown field is (exit 2), and the list of fields its
  message gives drops them. Records that have them stay valid and are shown
  as they are; `schemas/defect.v1.json` keeps both properties.
- `vloop brief close` no longer prints `defects of <run id>` or the untriaged
  hint; `squash-merge with the trailer:` stays its last line.
- The guidelines provenance check (`doctor`'s `guidelines` line) is unchanged:
  a rule still cites a defect or intervention ID.

### D1 — docs, domain and playbooks

**Required.**
- `docs/domain/domain-model.md`: M-4 says escaped is decided by time against
  M-5's release; M-6's commit carries intervention files, not defect files;
  the Identifiers table shows the new stamps and says old IDs stay.
- `docs/domain/measurement/defect.md`: recorded means escaped; the time rule;
  "The rule that caught it" is removed, with one line saying the fields were
  retired in B15 and stay valid in old records.
- `docs/domain/measurement/metrics.md` and `docs/guide/metrics.md`: the two
  per-task columns, `$/kloc`, the renamed headers, the operator count's two
  sources, escaped by time.
- `docs/guide/defects.md`: what to record as a defect (escapes, with
  `--blame`) and what as an intervention (anything found before merge).
- `plugin/skills/operate/SKILL.md`: the close step's findings take their
  origin; the triage step is removed. It stays within its word ceiling.
- `.claude/skills/vloop-operator/SKILL.md`, this repository's playbook: step 5's
  "Record defects" says to record findings before merge as interventions with
  `--origin`, and defects only for escapes.
- `docs/guide/commands.md` is regenerated.

### Decided here, because the cited documents leave it underdetermined

- Per-task `$/kloc` uses the task's `code+` (its churned lines, the only lines
  a task has); the summary's uses delivered lines, as today.
- The release commit's committer time is the merge time; a record created in
  the same second counts as not escaped.
- `automatable: partly` for a close finding: B16's verification phase is meant
  to find some of them.
- A finding's origin is matched case-sensitively, lower case only.
- H2 keeps the base gates and checks starting from the process environment, as
  today, even though the gate run honours an injected one: a difference to
  report in the task notes, not to fix here.

### Violations the review must rule on

- Any H1–H7 change that alters output, exit codes or session arguments, or
  changes an earlier test.
- A git call left outside the runner, or the driver's hardening applied where
  it is not today, or dropped where it is.
- A historical record edited, renamed or migrated.
- An intervention counted as operator without an origin, or counted twice
  with a defect.
- `found-by` given a new meaning or value.
- A skill grown past its ceiling, or changed beyond D1.

## Worked example

With a fresh build, in temporary repositories and a temporary `HOME`; "merged"
means the brief's consumed commit is on `main` (M-5):

```
vloop config set run.git-config-ignore 'Core.*'
  -> vloop: run.git-config-ignore: Core.* would ignore core.* — the guard must see it   exit 2
vloop config set run.git-config-ignore 'core.*'
  -> the same message for core.*, as today                                exit 2
vloop config set run.git-config-ignore 'Branch.Main.Foo-Bar'            exit 0
  vloop run, a stub session sets branch.Main.foo-bar -> the run continues
  vloop run, a stub session sets branch.main.foo-bar -> exit 9, as today

vloop defect add "x" --brief <unmerged brief> --found-by user
  -> vloop: <brief> is not merged: record what you found before merge as an intervention — vloop intervention add "<summary>" --brief <brief> --kind verification-finding --origin <brief|plan|work|env> --phase verify --automatable partly --by operator
                                                                           exit 1; no file
vloop defect add "x" --brief <merged brief, no other defects> --found-by operator
  -> .vloop/defects/D<YYYYMMDD-HHMMSS>-x.md                                 exit 0
the same again, in the same second
  -> .vloop/defects/D<the stamp one second later>-x.md                      exit 0
vloop metrics <merged brief>   -> defects   in-loop 0 · operator 0 · escaped 2 · removal efficiency 0%

vloop intervention add "y" --brief <b> --phase verify --kind verification-finding --automatable no --by operator
  -> vloop: --kind verification-finding needs --origin brief, plan, work or env   exit 2
… --kind repair --origin work
  -> vloop: --origin is only for --kind verification-finding                exit 2
… --kind verification-finding --origin plan
  -> .vloop/interventions/I<YYYYMMDD-HHMMSS>-y.md, with "origin: plan"     exit 0

vloop defect set <id> rule 'gate: x'
  -> vloop: cannot set "rule": want one of status, fixed-by, case, severity, origin, kind, task   exit 2

a brief run to completion with stub sessions, one gate failure:
vloop brief close <brief> --finding "spec gap: z"
  -> vloop: a finding starts with whose defect it is: --finding "<brief|plan|work|env>: <summary>"   exit 2
vloop brief close <brief> --finding "brief: z"
  -> recorded .vloop/interventions/I<stamp>-z.md
     no "defects of" lines; "squash-merge with the trailer: …" last         exit 0
  -> defects   in-loop 1 · operator 1 · escaped 0 · removal efficiency 100%
  -> the run record: "- I<stamp>-z — verification-finding, brief, found by operator: z"

vloop metrics <brief>, a task with code+ 200, test+ 300, cost 0.50
  -> its row: … other+ <n>  1.50  2.50  <agent> …
a task with code+ 0          -> its row: …  -  -  …
summary                      -> " · $<c>/kloc" ends the cost line
--json by_task[i]            -> "test_code_ratio": 1.5, "per_1000_code_lines_usd": 2.5 (null where code+ is 0)
```

Planted failures, each in a scratch copy of the tree:

```
a file outside the runner calling exec.Command("git", …)  -> H1's gate fails
iterate padded to 101 lines                               -> H6's gate fails
cmd/vloop made not to build                               -> go test ./cmd/vloop prints the build error and exits 1, no panic trace
```

**Real-data check**, on a clone of this repository:

- `vloop metrics` for each consumed brief: the `defects` lines before and after,
  in the task notes. Only `operator` and `escaped` move, and only for briefs
  with recorded defects created after their merge; the total escaped across
  them rises from 3. The operator compares each count with an independent one.
- Every brief's other summary lines are unchanged, but the cost line's ending.
- `vloop intervention show` and `vloop defect list` list every existing record.

## Out of scope

- **B16, the verification phase**; **B17, `vloop release`**.
- The Windows run harness; the long functions not named in H6 (`Iterator.Run`,
  `intervention.parse`, `newRun`, `metrics.Aggregate`, `runBriefCheck`).
- The v1 review's other findings: metrics correctness, ignored errors,
  cross-platform paths, `t.Parallel()`, `TestWorkedExampleB6RealData`.
- The two frontmatter readers that still interpret values on their own.
- Migrating, editing or backfilling any existing defect or intervention;
  removing `rule` or `rule-outcome` from the defect schema.
- An `origin` on intervention kinds other than `verification-finding`.
- Refreshing `.vloop/interventions/README-interventions.md` from `close`
  (this repository's tool, run by the operator).
- A new catcher value, or a found-by for escapes.

## Constraints

- Go as pinned in `go.mod`; no new dependencies. Every gate builds for
  `windows` and `linux` as well as the host.
- Gates run on macOS's BSD tools: no empty alternatives in `grep -E`, no
  `sed -i` without a suffix, no `\+` or `\|` in basic regexes, no `date -d`.
  A scratch directory is `mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX"`.
- Sessions never move refs and write nothing under `.vloop/` but .vloop/tmp/
  (S-4): no task's work, and no gate, may require a session to change this
  repository's records, config or state. Tests build their repositories in
  temporary directories, with a temporary `HOME`, and address them with
  `git -C`.
- **The gate model applies** (`plugin/skills/plan/SKILL.md`): each gate judges
  this brief's contract by running the built binary or reading the changed
  document or code, with its judge in the verify or its gate folder; no gate
  runs the repository's tests — this repository's `[[check]]` does, after
  every iteration. H1–H7's gates judge the code (a count of git call sites, a
  function's length), and the unchanged tests are their behaviour check.
- **No task may weaken a check to pass.** These earlier tests **must** change,
  and only as stated:
  - `internal/metrics/print_test.go`, `internal/metrics/lines_test.go`,
    `internal/cli/metrics_test.go`, `cmd/vloop/b3_e2e_test.go`,
    `cmd/vloop/run_iterate_test.go` — only by the two per-task columns, the
    cost line's `$<c>/kloc`, and the all-briefs headers (M1);
  - `internal/cli/close_test.go`, `internal/cli/close_flags_test.go`,
    `cmd/vloop/b4_e2e_test.go`, `cmd/vloop/b14_e2e_test.go` — only by findings
    taking an origin, being recorded as interventions, and the defect report
    going (E3, E4);
  - `internal/defect/defect_test.go` and `internal/intervention/intervention_test.go`
    — only by the second-stamped IDs and their allocation (S1), and the
    retired `rule` and `rule-outcome` setters (E4);
  - the tests that run `defect add` against a brief that is not merged —
    `cmd/vloop/b14_e2e_test.go`, `cmd/vloop/b20261005_e2e_test.go`,
    `cmd/vloop/b3_e2e_test.go`, `internal/cli/defect_test.go`,
    `internal/cli/export_test.go`, `internal/cli/intervention_test.go`,
    `internal/cli/matrix_test.go`, `internal/cli/metrics_test.go` — only by
    merging the brief in their fixture first, or by asserting the refusal (E1);
  - `internal/metrics/defects_test.go` — only by the time rule and the
    interventions in the operator count (E2, E3);
  - `docs/guide/commands.md`'s staleness test (regenerate, as always).
- Repo-relative paths everywhere. No absolute paths in any file or commit
  message.

## Shape

12 to 14 tasks, each independently verifiable by a single command. Order is
load-bearing:

1. **H7, the test binary fails cleanly** — first, so a later build break is
   readable.
2. **H1, one git runner.**
3. **H2, one environment filter.**
4. **H3, H4 and H5** — the plan at the last commit, the task count, the
   budget keys.
5. **H6, the five long functions** — after H1–H5, which shrink some of them.
6. **G1, git-config-ignore's case.**
7. **S1, unique stamps with seconds**, and one generator.
8. **M1, `test:code` and `$/kloc`.**
9. **E1, `defect add` refuses before merge.**
10. **E2, escaped by time**, with the real-data check's before and after.
11. **E3, `--origin` and close's findings as interventions**, and the operator
    count.
12. **E4, the triage retires.**
13. **D1, the docs, the domain and the playbooks.**
14. **Close:** the worked example line for line, and the real-data check.

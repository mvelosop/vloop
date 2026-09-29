---
name: B20260929-2325-vloop-metrics-defects.loop-brief
description: Give vloop its metrics and defect records — line classification with stack presets, per-brief and per-task metrics from the loop's own telemetry and git, defect files with blame attribution, release detection, and their guide docs
kind: brief
status: ready
created: 2026-09-29
seeds: A `.loop/run.sh` plan + run that builds slice B3 in docs/design-notes/vloop-roadmap.md
depends-on: [B20260929-2148-vloop-state-tasks-status.loop-brief]
---
# Brief — vloop B3: metrics and defects

- **Status:** ready to plan
- **Starting point:** extends `main` at the commit that adds this brief. The
  planner pins that SHA as the base every gate compares against.
- **Produced by:** operator decision, 2026-09-29, from the B3 row of
  `docs/design-notes/vloop-roadmap.md`, split in two as that row anticipated.

---

## What it is

The third slice of `vloop`. At the end of this run `vloop metrics` reports, for
any brief the loop has run, what it produced (lines of code, tests and docs —
delivered and churned — classified by built-in stack presets and the repo's own
globs), what it cost, how long it took, which models did it, and how many defects
were caught where; per brief, per task, and across briefs. It reads the shell
loop's run folders and commits, so B1 and B2 have metrics from their first
invocation. Defects found by the operator or by users are recorded as files with
`vloop defect add`, attributed to the brief that introduced them with
`--blame`, and counted with the ones the loop caught itself. A guide explains
every metric and the defect model.

## Why this shape, and what was rejected

Decided with the operator on 2026-09-29. **Do not re-litigate these** — the
roadmap's `## Metrics and defects` section records them in full.

- **This run is B3 of a split.** The roadmap's B3 row was ~16 tasks; it is cut
  into B3 (this: metrics, classification, defects, release detection, their
  guide) and B4 (`brief close`, the saved snapshot, export, `--workspace`, the
  rest of the guide). `init`/`doctor` and `run` move to B5 and B6.
  *Rejected:* one brief — twice the size of B1 or B2, and `close` needs every
  number this slice computes before it can write one.
- **Metrics are keyed by brief, not by run folder.** A brief spans several runs
  (B1: a refused resume and a resume; B2: a stalled run and a resume) and its
  plan session can sit in another branch's folder (B1's is under `main`).
- **Metrics are derived, never hand-entered.** Everything is recomputed from the
  run folders, git, the plan in git and the defect files. The only hand-entered
  records are defects the loop could not see.
- **Lines are classified by path globs** in layers: the repo's own globs first,
  then the built-in presets its `metrics.stacks` names. File extension alone is
  not enough: vloop's brief templates are `.md` files that are product code.
- **Delivered and churn are both reported**; rework = churn / delivered.
- **Defects have an origin and a catcher.** Origin `brief`, `plan`, `work`,
  `env`; caught by `gate`, `review`, `operator`, `user`. Gate failures and review
  findings are defects the loop derives itself; the rest are files. A gate can be
  the defect: a failure before an operator's `task verify` on that task is
  `origin: plan`.
- **Release is the brief's merge to the default branch**, detected, not
  declared. `found-by: user` is after release; the other catchers are before it.
- **This is vloop's, not the shell loop's.** vloop *reads* `.loop/state/runs/`
  and the `[loop]` commits; it never writes under `.loop/`.
- **Libraries:** B2's, plus `github.com/bmatcuk/doublestar/v4` for `**` globs.
  Git is invoked as the `git` command, not a Go reimplementation.

## Binding references

- `docs/design-notes/vloop-roadmap.md` — `## Metrics and defects`: every
  definition, the defect axes, release, the stack presets. Where it and this
  brief differ, this brief wins.
- `docs/additional-context-files/20260929-2143-vloop-metrics-sample-outputs.md` —
  the shape of the output this brief pins exactly.
- `docs/briefs/B20260929-2148-vloop-state-tasks-status.loop-brief.md` — the
  conventions this slice extends: the schemas and `vloop schema`, `session/v1`
  and `iteration/v1`, `gate_history`, the config key table.
- `docs/briefs/B20260929-1804-vloop-skeleton-config-briefs.loop-brief.md` — global
  flags, exit codes 0/1/2, `--json`, the repo root, the README test.
- `.loop/run.sh` — the shell loop's telemetry and commits this slice reads:
  `run_session` (session records), the iteration outcomes `done`, `blocked`,
  `gate_fail`, `review_fail`, and the commit subjects.

## Behaviour contract

B1's and B2's conventions hold: global flags, `--json` (one document on stdout,
on exit 0 and 1), exit codes `0` ok / `1` problems or failure / `2` usage, errors
as one stderr line starting `vloop: `, paths repo-relative with `/`.

### Where the numbers come from

A brief's **run id** is its name minus `.loop-brief`. vloop reads two layouts.

**The shell loop's** (read-only):

- Run folders `.loop/state/runs/<any>/<folder>/`. A folder belongs to a brief if
  its `loop.log` (ANSI colour stripped) contains `planning from <brief path>` or
  `resuming <run id>`. A folder with neither is ignored.
- `sessions/NNN-<phase>.json`: the `claude --output-format json` result plus
  `phase` and `iteration`. Cost `total_cost_usd`, time `duration_ms`, models and
  tokens from `modelUsage`, `permission_denials`.
- `iterations.jsonl`: `iteration`, `task`, `outcome` (`done`, `blocked`,
  `gate_fail`, `review_fail`).
- `reports/NNN-verdict.json`: the review's verdict for iteration NNN; `findings`
  is a list of strings.
- Commits, from every local ref: `[loop] plan <run id>`, `[loop] <task>:
  <outcome>`, `[loop] run <run id>/<folder>: <status>`. The plan at any commit is
  `.loop/state/state.json` in that commit.

**vloop's own** (the contract B6's driver will write; this slice only reads it,
from fixtures):

- Run folders `.vloop/state/runs/<run id>/<YYYYMMDD-HHMMSS>/`, holding
  `sessions/NNN-<phase>.json` (`session/v1`) and `iterations.jsonl`
  (`iteration/v1`). Verdicts: `reports/NNN-verdict.json` (`verdict/v1`).
- Commits `[vloop] plan <run id>`, `[vloop] <task>: <outcome>`,
  `[vloop] run <run id>/<folder>: <status>`; the plan is
  `.vloop/state/state.json` in that commit.

**Which commits a brief owns.** The latest `plan` commit for the run id, the last
`run <run id>/…` commit after it on the same history, and the `<task>:` commits
between them. When several plan commits exist (a re-plan), only the latest
counts.

### Line classification — `vloop metrics stacks` and `vloop metrics classify`

Five categories: `code`, `test`, `docs`, `other`, `excluded`. A path is
classified by the first **layer** that matches it:

1. **The repo's globs**, config keys `metrics.excluded`, `metrics.test`,
   `metrics.docs`, `metrics.code`, checked in that order.
2. **The presets** named in `metrics.stacks`, merged, checked in the same order:
   every preset's `excluded`, then every `test`, then `docs`, then `code`.
3. Nothing matched: `other`.

`.vloop/**` and `.loop/**` are always `excluded`, before any layer. Globs are
doublestar patterns matched against the repo-relative `/` path.

The presets are product data:

| Preset | code | test | excluded |
| --- | --- | --- | --- |
| `go` | `**/*.go` | `**/*_test.go`, `**/testdata/**` | `go.sum`, `vendor/**` |
| `typescript` | `**/*.ts` | `**/*.test.ts`, `**/*.spec.ts`, `**/__tests__/**` | package-lock.json, `yarn.lock`, pnpm-lock.yaml, `**/dist/**`, `**/node_modules/**` |
| `javascript` | `**/*.js`, `**/*.mjs`, `**/*.cjs` | `**/*.test.js`, `**/*.spec.js`, `**/__tests__/**` | as `typescript` |
| `react` | `**/*.tsx`, `**/*.jsx`, `**/*.css`, `**/*.scss` | `**/*.test.tsx`, `**/*.spec.tsx`, `**/*.test.jsx`, `**/*.spec.jsx`, `**/*.stories.*` | as `typescript` |
| `csharp` | `**/*.cs`, `**/*.razor`, `**/*.cshtml` | `**/*.Tests/**`, `**/*Tests.cs` | `**/bin/**`, `**/obj/**`, `**/*.Designer.cs`, `**/packages.lock.json` |
| `python` | `**/*.py` | `**/test_*.py`, `**/*_test.py`, `**/tests/**` | `**/__pycache__/**`, `poetry.lock`, `uv.lock`, `Pipfile.lock` |
| `java` | `**/*.java` | `**/src/test/**` | `**/build/**`, `**/target/**` |
| `kotlin` | `**/*.kt`, `**/*.kts` | `**/src/test/**` | `**/build/**`, `**/target/**` |
| `rust` | `**/*.rs` | `**/tests/**`, `**/benches/**` | `**/target/**`, `Cargo.lock` |

Every preset's `docs` is `**/*.md`.

- `metrics stacks` prints the preset names, one per line, sorted.
  `metrics stacks <name>` prints `code:`, `test:`, `docs:`, `excluded:` lines,
  each followed by its globs indented two spaces. An unknown name is
  `vloop: unknown stack "<name>"`, exit 2. `--json`: `{name: {code, test, docs, excluded}}`.
- `metrics classify <path>…` prints `<path>  <category>  (<layer>: <glob>)` per
  path — layer `repo`, a preset name, or `always` — or `<path>  other  (none)`.
- Config keys, after `areas`, in this order: `metrics.stacks` (preset names),
  `metrics.code`, `metrics.test`, `metrics.docs`, `metrics.excluded` (globs). All
  lists, default empty, set and printed like `areas`. An unknown stack is exit 2
  with B1's invalid-value message.

### What is counted

- **Lines** are **added non-blank lines** in a diff (a line holding only
  whitespace is blank), per category; deleted lines are reported separately.
  Binary files are not counted. Renames count as their content change only.
- **Delivered**: the diff from the plan commit's parent to the brief's last `run`
  commit.
- **Churn**: the sum over the brief's `<task>:` commits, whatever their outcome.
  **Rework** = churn code / delivered code.
- **Test:code** = delivered test / delivered code.
- **Time.** Agent = work + review session durations. Plan is reported apart.
  Gates: the sum of `iteration/v1` gate durations, `n/a` for the shell loop.
  Wall: the plan commit's committer time to the last `run` commit's.
- **Rates** use agent time: delivered code lines per minute, and code + test.
- **Cost** per phase and total: sums of session costs, rounded only for display.
  Per 1,000 delivered code lines. Tokens (input, output, cache read, cache
  creation) and cache-hit ratio = cache read / (input + cache read + cache
  creation), in `--json`.
- **Models**: per phase, the models in the sessions' model usage, sorted,
  comma-joined. Effort, per phase, when the records carry it (`session/v1`).
- **Tasks**: planned = tasks in the plan at the last `run` commit; done, blocked
  likewise; **first-pass** = tasks whose only iteration outcome is `done`; the
  brief's estimate is its `<n> to <m> tasks` phrase. Iterations per closed task.
- **Missing records**: an iteration whose outcome means a review ran (`done`,
  `review_fail`, `rejected`) with no review session record, or any iteration
  with no work session record. Their time and cost are not counted, and the
  summary says so.
- **Release**: the default branch is `origin/HEAD`'s target, else `main`, else
  `master`. The brief is **merged** at the first commit on it where the brief's
  frontmatter says `status: consumed`.

### Defects

**Derived by vloop** from the run, never written to disk:

- each `gate_fail` / `gate_failed` iteration: one defect, `origin: work`,
  `kind: bug`, `found-by: gate`;
- each `review_fail` / `rejected` iteration: one defect per verdict finding
  (`kind` from the finding in `verdict/v1`, `bug` for the shell loop's strings;
  one defect if the list is empty), `origin: work` (`brief` for a `spec-gap`
  finding, `plan` for a `gate-gap`), `found-by: review`;
- a gate failure on a task **before** an entry in that task's `gate_history`
  with `by: operator` becomes `origin: plan`, `kind: gate`.

`blocked` outcomes are not defects by themselves.

**Recorded by the operator**, one file each:
`.vloop/defects/D<YYYYMMDD-HHMM>-<slug>.md` (local time; `-2`, `-3`… on
collision), YAML frontmatter then the description:

```
---
id: D20260101-1000-util-drops-a-line
brief: B20260101-0900-a.loop-brief
task: T2                  # optional
origin: work              # brief | plan | work | env
found-by: user            # gate | review | operator | user
kind: bug                 # bug | spec-gap | gate | gate-gap | regression
severity: medium          # low | medium | high | critical
status: open              # open | fixed | wontfix
fixed-by: ""              # the loop brief that fixed it
case: ""                  # the failing test written first
created: 2026-01-01T10:00:00Z
---
util drops a line
```

Its schema is `defect/v1` (the frontmatter as JSON), shown by `vloop schema`.

- `vloop defect add "<summary>" --found-by <c> [--brief <name>] [--task <id>]
  [--origin work] [--kind bug] [--severity medium] [--case <path>]
  [--blame <file>:<line>]` writes the file and prints its path. The slug is the
  summary lower-cased, non-alphanumerics runs as `-`, at most 40 characters.
- `--blame <file>:<line>` attributes the defect when `--brief` is absent: `git
  blame` the line on the default branch; the commit's `Vloop-Brief: <name>`
  trailer names the brief, else the loop brief whose frontmatter became
  `status: consumed` in that commit. stderr gets
  `attributed to <name> (trailer on <sha7>)` or `(consumed in <sha7>)`.
  Unattributable: `vloop: cannot attribute <file>:<line> to a loop brief — pass
  --brief`, exit 1, nothing written.
- A `--brief` that is not a loop brief in the briefs directory, or a `--task`
  not in that brief's plan, is exit 1. Invalid enum values are exit 2 with B1's
  invalid-value message.
- `vloop defect list [--brief <name>] [--json]` prints recorded defects:
  `<id>  <brief>  <task|->  <kind>  <origin>  <found-by>  <status>  <summary>`,
  sorted by id. `--matrix` prints origin × catcher counts for the selected
  briefs, **derived defects included**:

```
       gate  review  operator  user
brief     0       0         0     0
plan      0       0         0     0
work      1       0         0     1
env       0       0         0     0
```

- `vloop defect set <id> <field> <value>` sets `status`, `fixed-by`, `case`,
  `severity`, `origin`, `kind` or `task`, validated as `add` validates.
- **Counting.** In-loop = found by `gate` or `review` (derived and recorded);
  operator; escaped = `user`. **Removal efficiency** = (in-loop + operator) /
  all, shown as a whole percentage, or `n/a` with no defects.

### `vloop metrics [<brief>…] [--by task] [--json]`

A `<brief>` is a loop brief's path or name. With one brief:

```
<run id>  <status> · merged <sha7>            (or: · not merged)
 tasks     <p> planned (brief said <n>–<m>) · <d> done · <b> blocked · first-pass <f>/<p>
 size      delivered  code <n> · test <n> · docs <n> · test:code <r>
           churn      code <n> · test <n> · docs <n> · rework <r>
 time      agent <m> min (work <m> · review <m>) · plan <m> min · gates <m|n/a> · wall <m> min
 rate      <r> code lines/min · <r> incl. tests
 cost      $<t> · plan <c> · work <c> · review <c> · $<c> per 1,000 code lines
 models    plan <models> · work <models> · review <models>
 defects   in-loop <n> · operator <n> · escaped <n> · removal efficiency <p>%
 records   <n> session record(s) missing: <task> <phase> (iteration <i>)[, …] — their time and cost are not counted
```

- The `records` line appears only when records are missing. `(brief said …)`
  is omitted when the brief states no estimate. Minutes and ratios to one or two
  decimals as shown in the worked example; money to two; counts with `,`
  thousands separators. Rows are compared with runs of spaces collapsed.
- `--by task`: one row per task in plan order —
  `id  area  kind  att  code+  test+  docs+  other+  agent  cost  model` — where
  `att` is its iterations, the line columns are its churn, `agent` is `XmYYs`,
  and `-` stands for a missing area or kind.
- With no `<brief>`: one row per brief that has runs, in name order —
  `brief  tasks  first-pass  code  test  t:c  agent  $/1k  in-loop  operator  escaped  efficiency`
  (`brief` is the run id, `agent` in minutes, `$/1k` the cost per 1,000
  delivered code lines).
- A brief with no runs: `vloop: no runs for <name>`, exit 1.
- `--json` prints one `metrics/v1` document per brief (an array when several):
  top-level keys `schema`, `brief`, `run_id`, `status`, `merged`, `tasks`,
  `size`, `time`, `rate`, `cost`, `tokens`, `models`, `effort`, `defects`,
  `records`, `lead_time`, `permission_denials`, `iterations`,
  `iterations_per_closed`, `by_task`. Durations in milliseconds, money unrounded,
  unknowns `null`. `metrics/v1` joins `vloop schema list`.

### Decided here, because the cited documents leave it underdetermined

- Which commits and folders a brief owns, and that only the latest plan counts.
- "Added non-blank lines" as the line measure, and churn per task as the sum of
  its commits.
- Gate time is `n/a` for the shell loop, which never recorded it; wall time comes
  from commit timestamps.
- Derived defects are computed on every call, not stored, so reclassification by
  a later `task verify` needs no migration.
- `found-by` alone decides before/after release; the merge commit only labels
  the brief.
- B1's run record said $7.41 and 15.6 min; the telemetry says $7.40 and
  15.8 min. Summing unrounded values, as specified above, is the rule; the run
  records were hand-rounded.

### Violations the review must rule on

- Writing anything under `.loop/`, or anywhere but .vloop/defects/ and the
  config file. `vloop metrics` and `defect list` write nothing.
- A number the worked example or the real-data check pins, computed a different
  way that happens to match on one fixture.
- Line counts that include `.loop/` or .vloop/ paths, blank lines, or binary
  files.
- A preset or heading table that exists only in a test file.
- B4's work appearing here: `brief close`, writing a run record or a metrics
  snapshot, `metrics export`, `--workspace`.
- Tests reading this repo's `.loop/` or git history. Only the close task's
  real-data check reads them, read-only.

## Worked example

A fixture repository, built by the test with fixed committer dates:

- `main` holds docs/briefs/B20260101-0900-a.loop-brief.md (`status: ready`,
  Shape `2 to 3 tasks`); the config sets `metrics.stacks = ["go"]`.
- Branch `B20260101-0900-a`: `[loop] plan B20260101-0900-a` at 09:00:00 (a plan
  with T1 and T2, no area or kind); `[loop] T1: done` adding `main.go` (10
  non-blank and 2 blank lines), `main_test.go` (6), `README.md` (3);
  `[loop] T2: gate_fail` adding `util.go` (4); `[loop] T2: done` deleting 2 of
  those and adding 3; `[loop] run B20260101-0900-a/20260101-090000: complete`
  at 09:10:00.
- Its run folder, .loop/state/runs/B20260101-0900-a/20260101-090000/:
  `loop.log` says `planning from docs/briefs/B20260101-0900-a.loop-brief.md`;
  `iterations.jsonl` holds T1 `done`, T2 `gate_fail`, T2 `done` (iterations
  1–3); sessions: plan $1.00 / 120 s on `claude-opus-5-5`; T1 work $0.40 / 60 s,
  review $0.10 / 20 s; T2 work $0.30 / 40 s; T2 work $0.20 / 30 s, review
  $0.10 / 10 s — all on `claude-sonnet-5-5`.
- `main` then gets one squash commit of the branch with the brief set to
  `status: consumed` and the trailer `Vloop-Brief: B20260101-0900-a.loop-brief`.
  `<sha7>` is threaded from `git rev-parse --short` on it.

```
$ vloop metrics B20260101-0900-a.loop-brief
B20260101-0900-a  consumed · merged <sha7>
 tasks     2 planned (brief said 2–3) · 2 done · 0 blocked · first-pass 1/2
 size      delivered  code 15 · test 6 · docs 3 · test:code 0.40
           churn      code 17 · test 6 · docs 3 · rework 1.13
 time      agent 2.7 min (work 2.2 · review 0.5) · plan 2.0 min · gates n/a · wall 10.0 min
 rate      5.6 code lines/min · 7.9 incl. tests
 cost      $2.10 · plan 1.00 · work 0.90 · review 0.20 · $140.00 per 1,000 code lines
 models    plan claude-opus-5-5 · work claude-sonnet-5-5 · review claude-sonnet-5-5
 defects   in-loop 1 · operator 0 · escaped 0 · removal efficiency 100%
                                                               exit 0
$ vloop metrics B20260101-0900-a.loop-brief --by task
id  area  kind  att  code+  test+  docs+  other+  agent  cost   model
T1  -     -     1    10     6      3      0       1m20s  $0.50  claude-sonnet-5-5
T2  -     -     2    7      0      0      0       1m20s  $0.60  claude-sonnet-5-5
                                                               exit 0
$ vloop metrics classify main.go main_test.go README.md go.sum .loop/x notes.txt
main.go  code  (go: **/*.go)
main_test.go  test  (go: **/*_test.go)
README.md  docs  (go: **/*.md)
go.sum  excluded  (go: go.sum)
.loop/x  excluded  (always: .loop/**)
notes.txt  other  (none)                                       exit 0

$ vloop defect add "util drops a line" --found-by user --blame util.go:1
.vloop/defects/D<stamp>-util-drops-a-line.md                   exit 0
  (stderr) attributed to B20260101-0900-a.loop-brief (trailer on <sha7>)
$ vloop metrics B20260101-0900-a.loop-brief | grep defects
 defects   in-loop 1 · operator 0 · escaped 1 · removal efficiency 50%
$ vloop defect list --matrix
       gate  review  operator  user
brief     0       0         0     0
plan      0       0         0     0
work      1       0         0     1
env       0       0         0     0                              exit 0
```

The failures, each planted in a copy of the fixture:

```
the trailer removed from the squash commit
  -> attributed to B20260101-0900-a.loop-brief (consumed in <sha7>)   exit 0
--blame on a line of a commit that is neither
  -> vloop: cannot attribute notes.txt:1 to a loop brief — pass --brief
                                                                   exit 1
the review session record for iteration 3 deleted
  -> records   1 session record(s) missing: T2 review (iteration 3) — their time and cost are not counted
     and review 0.10, cost $2.00                                   exit 0
the run folder's loop.log names another brief
  -> vloop: no runs for B20260101-0900-a.loop-brief                exit 1
the plan's T2 given gate_history [{by: operator, …}] at the done commit
  -> the matrix's gate failure moves from work to plan             exit 0
vloop metrics stacks cobol
  -> vloop: unknown stack "cobol"                                  exit 2
```

**Real data.** On this repository, read-only, with the work branches of B1 and
B2 present as local refs:

```
$ vloop metrics B20260929-1804-vloop-skeleton-config-briefs.loop-brief   (among its lines)
B20260929-1804-vloop-skeleton-config-briefs  consumed · merged 8da6c95
 tasks     9 planned (brief said 8–10) · 9 done · 0 blocked · first-pass 9/9
 time      agent 15.8 min (work 12.9 · review 2.9) · plan 12.3 min · gates n/a · wall <threaded>
 cost      $7.40 · plan 2.94 · work 3.37 · review 1.10 · $<threaded> per 1,000 code lines
 models    plan claude-opus-5-5 · work claude-sonnet-5-5 · review claude-sonnet-5-5

$ vloop metrics B20260929-2148-vloop-state-tasks-status.loop-brief     (among its lines)
B20260929-2148-vloop-state-tasks-status  consumed · merged 6de9cfc
 tasks     9 planned (brief said 8–10) · 9 done · 0 blocked · first-pass 8/9
 time      agent 17.0 min (work 14.0 · review 2.9) · plan 14.3 min · gates n/a · wall <threaded>
 cost      $8.36 · plan 3.76 · work 3.54 · review 1.06 · $<threaded> per 1,000 code lines
 records   1 session record(s) missing: T2 review (iteration 4) — their time and cost are not counted
```

## Out of scope

B4, the other half of the split, not smaller versions of it:

- `vloop brief close`, writing a brief's `## Run record`, saving a metrics
  snapshot under .vloop/state/metrics/.
- `vloop metrics export` (JSON Lines) and `vloop metrics --workspace`.
- The rest of the guide: concepts, the configuration reference, the command
  reference generated from the binary.

The roadmap's later rows:

- `init`, `upgrade`, `doctor`, skills (B5); `run` and anything that writes run
  folders or `[vloop]` commits (B6).

Also not in this run:

- Task-level blame (which task wrote a line), re-pricing old runs, dashboards.
- Gate time or effort for shell-loop runs, which never recorded them.
- Writing anything under `.loop/`, or fixing the shell loop's telemetry gaps.
- Deleting defect files, or editing their description (it is a file; edit it).

## Constraints

- Go as pinned in `go.mod`. Third-party modules: B2's, plus
  `github.com/bmatcuk/doublestar/v4`; anything else needs its reason in the task
  notes. No cgo. vloop never uses the network. Git is run as the `git` command.
- Must build for `darwin`, `linux` and `windows`. Nothing may assume `/` as the
  on-disk separator or a case-sensitive filesystem.
- **Gates run on macOS with its BSD tools.** No GNU-only syntax in a gate: no
  empty alternatives in `grep -E` patterns, no `sed -i` without a suffix, no
  `\+` or `\|` in basic regexes, no `date -d`. B2's first run stalled on exactly
  this.
- **Gates follow the diff surface**: each task gates on the packages it touches,
  plus `go vet ./...`, `gofmt -l .` printing nothing, `go mod tidy -diff`
  printing nothing, and the linux, windows and host builds. Only the close runs
  `go test ./...`.
- A gate runs in under a minute. Tests build fixture repositories in temporary
  directories and never read this repo's `.loop/`, `.claude/`, git history, or
  the home directory.
- **No task may weaken a check to pass.** Two earlier expectations must change,
  and changing them is required, not weakening: `config list` gains the five
  `metrics.*` lines, and `schema list` gains `defect/v1` and `metrics/v1`.
  Update exactly those expectations.
- **New checks are proven by fixtures**: every preset, every layer, every
  counting rule and every defect rule has a fixture where it applies and one
  where it must not.
- Repo-relative paths everywhere. No absolute paths in any file or commit message.

## Shape

9 to 11 tasks, each independently verifiable by a single command. Order is
load-bearing:

1. **`defect/v1` and `metrics/v1`** in the schemas, and the five `metrics.*`
   config keys, with the two expectation updates above.
2. **Classification**: the presets, the layers, `metrics stacks`,
   `metrics classify`, with a fixture per preset and per layer.
3. **Reading runs**: the shell-loop and vloop layouts into one model — folders to
   brief, sessions, iterations, verdicts, owned commits, the plan in git — with
   fixture run folders for both.
4. **Lines**: delivered, churn per task, non-blank counting, deletions, rework,
   test:code, on fixture repositories.
5. **Time, cost, tokens, models, tasks, first-pass, missing records.**
6. **Release detection and `defect add|list|set`**, with `--blame` through a
   trailer and through a consumed-in commit.
7. **Derived defects**, `gate_history` reclassification, the matrix, removal
   efficiency.
8. **`vloop metrics`**: the summary, `--by task`, the cross-brief table,
   `--json` validating against `metrics/v1`.
9. **Docs**: the README's new commands and keys (B1's README test passes
   unmodified); docs/guide/metrics.md — every `metrics/v1` key with its
   definition and formula, the classification layers and the preset table;
   docs/guide/defects.md — the axes, derived vs recorded, gate-as-defect,
   blame, the file format. Gated by tests that every `metrics/v1` leaf key,
   every preset name and every defect field appears in the guide.
10. **Close**: an end-to-end test that builds the binary and plays the worked
    example line for line on the fixture; the real-data check on this
    repository (read-only); `go test ./...`, `go vet ./...`, `gofmt -l .`,
    `go mod tidy -diff` and the three builds.

The worked example and the real-data check need their own gate (task 10): unit
tests on fixtures can all pass while the adapter misreads the real shell-loop
layout, which only the real run folders exercise.

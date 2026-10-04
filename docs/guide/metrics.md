# Metrics guide

`vloop metrics` reports what a brief's loop produced, what it cost and how long
it took. Nothing is hand-entered: every number is recomputed on each call from
the run folders, git, the plan stored in git and the defect files. Defects are
explained in [defects.md](defects.md).

```
vloop metrics [<brief>…] [--by task] [--json]
vloop metrics stacks [name]
vloop metrics classify <path>…
```

A `<brief>` is a loop brief's path or name. With none, one row per brief that
has runs is printed. A brief with no runs is `vloop: no runs for <name>`, exit 1.
`--by task` prints one row per task in plan order; `--json` prints one
`metrics/v2` document per brief (an array when several). `vloop schema show
metrics/v2` prints the schema; every key below is a property of it.

## Where the numbers come from

A brief's **run id** is its file name without `.loop-brief` (`run_id`). Metrics
are keyed by brief, not by run folder: a brief can span several runs, and its
plan session can sit in another branch's folder.

- **Run folders**: the shell loop's `.loop/state/runs/<any>/<folder>/` and
  vloop's `.vloop/state/runs/<run id>/<folder>/`. A shell-loop folder belongs to
  a brief if its `loop.log` says `planning from <brief path>` or `resuming <run
  id>`. Sessions give cost, duration, models and tokens; `iterations.jsonl`
  gives each iteration's task and outcome; `reports/NNN-verdict.json` the
  review's findings.
- **Commits**: `[loop]` or `[vloop]` `plan <run id>`, `<task>: <outcome>` and
  `run <run id>/<folder>: <status>`, from every local ref. A brief owns the
  latest plan commit, the last run commit after it and the task commits between
  them. With several plan commits (a re-plan) only the latest counts.

vloop only reads `.loop/`; it never writes there. `metrics` and `defect list`
write nothing at all.

## Line classification

Every changed path falls into one of five categories: `code`, `test`, `docs`,
`other` and `excluded`. `excluded` lines are not counted anywhere. A path is
classified by the first **layer** that matches it, and globs are doublestar
patterns matched against the repo-relative `/` path:

1. **always**: `.loop/**` and `.vloop/**` are `excluded`, before any other layer.
2. **repo**: the repository's own globs, from the config keys `metrics.excluded`,
   `metrics.test`, `metrics.docs` and `metrics.code`, checked in that order.
3. **presets**: the built-in presets named in `metrics.stacks`, merged and
   checked in the same order: every preset's `excluded`, then every `test`, then
   `docs`, then `code`. Only the presets of the path's **scope** apply (below).
4. **other**: nothing matched.

File extension alone is not enough, which is why the repo layer wins: a
project's brief templates can be `.md` files that are product code, so set
`metrics.code = ["internal/brief/templates/**"]` and they count as code.

**Scopes.** A `metrics.stacks` entry may be scoped to a directory:
`csharp@services/api`. A path's scope is the longest scoped path containing it.
Inside a scope only that scope's stacks apply, so another stack's globs never
classify its files; outside every scope the unscoped stacks apply. A scoped
preset's globs match the path relative to the scope's directory:
`services/api/Api.Tests/CalcTests.cs` is tested against `**/*.Tests/**` as
`Api.Tests/CalcTests.cs`, and `services/api/go.sum` against `go.sum`. The repo's
own globs stay repo-relative and unscoped.

The config keys are lists of doublestar globs (`metrics.stacks` a list of preset
names), set with `vloop config set metrics.code a,b`, or from the environment:
`VLOOP_METRICS_STACKS`, `VLOOP_METRICS_CODE`, `VLOOP_METRICS_TEST`,
`VLOOP_METRICS_DOCS`, `VLOOP_METRICS_EXCLUDED`. An unknown stack is exit 2.

`vloop metrics classify <path>…` prints `<path>  <category>  (<layer>: <glob>)`,
the layer being `repo`, a preset name or `always`, or `<path>  other  (none)`.
A match by a scoped preset is labelled `(<stack>@<path>: <glob>)`, for example
`(csharp@services/api: **/*.Tests/**)`.

### The presets

`vloop metrics stacks` lists them and `vloop metrics stacks <name>` prints one.
The tables are the binary's own data.

| Preset | code | test | excluded |
| --- | --- | --- | --- |
| `csharp` | `**/*.cs`, `**/*.razor`, `**/*.cshtml` | `**/*.Tests/**`, `**/*Tests.cs` | `**/bin/**`, `**/obj/**`, `**/*.Designer.cs`, `**/packages.lock.json` |
| `go` | `**/*.go` | `**/*_test.go`, `**/testdata/**` | `go.sum`, `vendor/**` |
| `java` | `**/*.java` | `**/src/test/**` | `**/build/**`, `**/target/**` |
| `javascript` | `**/*.js`, `**/*.mjs`, `**/*.cjs` | `**/*.test.js`, `**/*.spec.js`, `**/__tests__/**`, `**/*.e2e-spec.js`, `**/test/**` | `package-lock.json`, `yarn.lock`, `pnpm-lock.yaml`, `**/dist/**`, `**/node_modules/**` |
| `kotlin` | `**/*.kt`, `**/*.kts` | `**/src/test/**` | `**/build/**`, `**/target/**` |
| `python` | `**/*.py` | `**/test_*.py`, `**/*_test.py`, `**/tests/**` | `**/__pycache__/**`, `poetry.lock`, `uv.lock`, `Pipfile.lock` |
| `react` | `**/*.tsx`, `**/*.jsx`, `**/*.css`, `**/*.scss` | `**/*.test.tsx`, `**/*.spec.tsx`, `**/*.test.jsx`, `**/*.spec.jsx`, `**/*.stories.*` | `package-lock.json`, `yarn.lock`, `pnpm-lock.yaml`, `**/dist/**`, `**/node_modules/**` |
| `rust` | `**/*.rs` | `**/tests/**`, `**/benches/**` | `**/target/**`, `Cargo.lock` |
| `typescript` | `**/*.ts` | `**/*.test.ts`, `**/*.spec.ts`, `**/__tests__/**`, `**/*.e2e-spec.ts`, `**/test/**` | `package-lock.json`, `yarn.lock`, `pnpm-lock.yaml`, `**/dist/**`, `**/node_modules/**` |

Every preset's `docs` is `**/*.md`.

## What is counted

**Lines** are *added non-blank lines* in a diff, per category; a line holding
only whitespace is blank. Deleted lines are reported apart, per category, as
`deleted`. Binary files are not counted, and a rename counts only its content
change.

- **Delivered** (`size.delivered`): the diff from the plan commit's parent to the
  brief's last run commit. It is what the brief left behind.
- **Churn** (`size.churn`): the sum over the brief's task commits, whatever their
  outcome, so failed attempts count. `by_task[].churn` is the same per task.
- **Rework** (`size.rework`) = churn `code` / delivered `code`. 1.00 would mean
  every line was written once; 1.13 means 13% more code was written than kept.
- **Test:code** (`size.test_code_ratio`) = delivered `test` / delivered `code`.

Ratios with a zero denominator are `null` in JSON and `n/a` in text.

Each of `delivered` and `churn` holds `code`, `test`, `docs` and `other`, and a
`deleted` object with the same four counts.

## Time

- **agent** (`time.agent_ms`) = `work_ms` + `review_ms`: the durations of the work
  and review sessions. **plan** (`plan_ms`) is reported apart.
- **gates** (`gates_ms`): the sum of the gate durations in the iteration records;
  `null` (`n/a`) when none was recorded, as always for the shell loop.
- **checks** (`checks_ms`): the sum of the check durations in the iteration
  records, the checks the driver ran after each gate; `null` (`n/a`) when none
  was recorded. The final pass writes no iteration record, so it is not counted.
- **gate review** (`gate_review_ms`): the duration of the gate-review sessions,
  the gate-review phase before any work. It is not part of `agent_ms`.
- **wall** (`wall_ms`): the plan commit's committer time to the last run
  commit's. `null` without a run commit.
- **lead_time** (`lead_time.plan_to_merge_ms`): the plan commit to the merge
  commit; `null` when not merged.

Durations in JSON are milliseconds; the text shows minutes.

## Rate

Rates use agent time and delivered lines: `rate.code_per_min` is delivered code
lines per agent minute, and `rate.code_test_per_min` counts code plus test.

## Cost and tokens

- `cost.total_usd` (gate-review sessions included), `cost.plan_usd`, `cost.work_usd` and `cost.review_usd` are
  sums of the sessions' costs, unrounded in JSON and rounded to cents only for
  display. `cost.per_1000_code_lines_usd` is the total per 1,000 delivered code
  lines.
- `tokens` holds `input`, `output`, `cache_read` and `cache_creation` totals and
  `cache_hit_ratio` = `cache_read` / (`input` + `cache_read` + `cache_creation`).
- `permission_denials` counts the tool calls the sessions were denied.

## Models and effort

`models` lists, for each of `plan`, `gate-review`, `work` and `review`, the models seen in the
sessions' model usage, sorted. `effort` gives each phase's effort level when the
records carry it (`session/v1` does, the shell loop's do not), else `null`.

## Tasks and iterations

- `tasks.planned`, `tasks.done` and `tasks.blocked` count the tasks in the plan
  at the last run commit.
- `tasks.first_pass` counts the tasks whose only iteration outcome is `done`:
  nothing failed a gate or a review on the way.
- `tasks.estimate` is the `<n> to <m> tasks` phrase (`<n> a <m> tareas` in
  Spanish) in the brief's `## Shape` (`## Forma`) section, as `min` and `max`;
  `null` when that section states none. Only that section counts: a worked
  example may describe a fixture brief with an estimate of its own.
- `iterations` counts all iterations and `iterations_per_closed` = `iterations`
  / `tasks.done`, `null` when none is done.

`--by task` and `by_task` give one row per task in plan order: `id`, `area`,
`kind` (`-` or `null` when missing), `attempts` (its iterations), `churn`,
`agent_ms`, `cost_usd` and `models`.

## Missing records

An iteration whose outcome means a review ran (`done`, `review_fail`,
`rejected`) with no review session record, or any iteration with no work
session record, is listed in `records.missing` as `task`, `phase` (`work` or
`review`) and `iteration`. Its time and cost are **not counted**, and the text
summary says so on a `records` line.

## Defects, release and status

- `defects` holds `in_loop`, `operator`, `escaped`, `total` and
  `removal_efficiency`; see [defects.md](defects.md).
  Each iteration that ended `check_failed` derives a defect of kind
  `regression` (origin `work`) found by the gate: a check caught what the
  work broke.
- `status` is the brief's frontmatter status. `merged` is the full SHA of the
  first commit on the default branch (`origin/HEAD`'s target, else `main`, else
  `master`) where the brief says `status: consumed`; `null` when not merged.
- `schema` is `metrics/v2` (snapshots and records of earlier briefs may be `metrics/v1`, still read), and `brief` is the brief's file name.

## The summary

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
```

The cross-brief table has one row per brief in name order: `brief`, `tasks`,
`first-pass`, `code`, `test`, `t:c`, `agent` (minutes), `$/1k`, `in-loop`,
`operator`, `escaped` and `efficiency`.

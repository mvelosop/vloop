# Defects guide

A defect is something wrong that a brief's loop produced or let through. vloop
counts them on two axes, so you can see not only how many there were but where
they came from and who caught them.

## Origin and catcher

**Origin** is where the defect entered:

- `brief`: the brief was wrong or incomplete (a spec gap).
- `plan`: the plan or its gate was wrong: a task split badly, or a verify
  command that could not catch what it should have.
- `work`: the work session got it wrong.
- `env`: the environment: tools, dependencies, infrastructure.

**Catcher** (the `found-by` field) is who found it: `gate`, `review`, `operator`
or `user`. The first two are the loop itself, `operator` is you before release
and `user` is after release. Release is the brief's merge to the default branch,
detected (`merged` in `vloop metrics`), not declared; `found-by` alone decides
which side of it a defect is on.

## Derived and recorded defects

**Derived** defects come from the run and are never written to disk. They are
recomputed on every call, so a later correction needs no migration:

- each `gate_fail` / `gate_failed` iteration is one defect: `origin: work`,
  `kind: bug`, `found-by: gate`;
- each `review_fail` / `rejected` iteration is one defect per finding in the
  verdict, `found-by: review`, `origin: work`. The `kind` comes from the finding;
  it is `bug` for the shell loop's plain-string findings. A `spec-gap` finding is
  `origin: brief` and a `gate-gap` finding `origin: plan`. An empty findings
  list is one defect;
- an iteration whose gate failed and then passed on its immediate re-run
  (`gate.flaky` in the iteration record) is one defect: `origin: env`,
  `kind: bug`, `found-by: gate`. The work is not at fault, so it yields no work
  defect and charges no attempt;
- a `blocked` outcome is not a defect by itself.

**Recorded** defects are the ones the loop could not see, one file each, written
by `vloop defect add`.

### The gate as a defect

A gate can be the defect. When a task's `gate_history` (written by `vloop task
verify --reason`) has an entry with `by: operator`, every gate failure on that
task *before* that entry is reclassified: `origin: plan`, `kind: gate`. The gate
was wrong, not the work. A failure whose time is unknown is never reclassified.

## The `.vloop/defects/` files

`.vloop/defects/D<YYYYMMDD-HHMM>-<slug>.md` (local time; `-2`, `-3`… on
collision): YAML frontmatter, then the description.

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

The frontmatter is the `defect/v1` schema (`vloop schema show defect/v1`). Every
field:

- `schema`: always `defect/v1` in the JSON form.
- `id`: the file name without `.md`.
- `brief`: the loop brief that introduced the defect.
- `task`: optional; a task of that brief's plan.
- `origin`: `brief`, `plan`, `work` or `env`.
- `found-by`: `gate`, `review`, `operator` or `user`.
- `kind`: `bug`, `spec-gap`, `gate`, `gate-gap` or `regression`.
- `severity`: `low`, `medium`, `high` or `critical`.
- `status`: `open`, `fixed` or `wontfix`.
- `fixed-by`: the loop brief that fixed it, empty until then.
- `case`: the failing test written first, empty until then.
- `created`: an RFC 3339 date-time.

## Commands

`vloop defect add "<summary>" --found-by <c>` writes the file and prints its
path. Flags: `--brief <name>`, `--task <id>`, `--origin` (default `work`),
`--kind` (default `bug`), `--severity` (default `medium`), `--case <path>` and
`--blame <file>:<line>`. The slug is the summary lower-cased, runs of
non-alphanumerics as `-`, at most 40 characters. A `--brief` that is not a loop
brief, or a `--task` not in its plan, is exit 1; an invalid value is exit 2.

`vloop defect list [--brief <name>]` prints
`<id>  <brief>  <task|->  <kind>  <origin>  <found-by>  <status>  <summary>`,
sorted by id.

`vloop defect set <id> <field> <value>` changes `status`, `fixed-by`, `case`,
`severity`, `origin`, `kind` or `task`, validated as `add` validates.

### Attribution with `--blame`

When `--brief` is absent, `--blame <file>:<line>` finds the brief for you. vloop
runs `git blame` on that line on the default branch, then reads the commit:

1. a `Vloop-Brief: <name>` trailer names the brief; else
2. the loop brief whose frontmatter became `status: consumed` in that commit.

stderr says `attributed to <name> (trailer on <sha7>)` or `(consumed in <sha7>)`.
If neither works: `vloop: cannot attribute <file>:<line> to a loop brief — pass
--brief`, exit 1, and nothing is written.

## The matrix and removal efficiency

`vloop defect list --matrix [--brief <name>]` counts origin × catcher for the
selected briefs, derived defects included:

```
       gate  review  operator  user
brief     0       0         0     0
plan      0       0         0     0
work      1       0         0     1
env       0       0         0     0
```

Without `--brief` it covers every brief that has runs, plus all recorded defects.

The counts roll up as **in-loop** (found by `gate` or `review`), **operator** and
**escaped** (found by `user`). **Removal efficiency** = (in-loop + operator) /
all defects, shown as a whole percentage, or `n/a` with no defects. It is the
share of defects removed before users saw them. In `vloop metrics --json` these
are `defects.in_loop`, `operator`, `escaped`, `total` and `removal_efficiency`
(a fraction, or `null`).

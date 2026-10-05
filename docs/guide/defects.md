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

## Interventions

An intervention is what the operator (or the assistant as their hands) did around
a run besides testing it: a decision, a halt handled, a repair. Recording them
shows what a driver could one day do itself, and, where the assistant proposed
options, how often the operator took its recommendation. One file each:
`.vloop/interventions/I<YYYYMMDD-HHMM>-<slug>.md`, YAML frontmatter, then the
summary line and the sections **Trigger.**, **Done.** and **What would automate
it.**, plus, in a v2 record, **Context.**, **Options.**, **Recommended.**,
**Suggested.** and **Decided.**

```
---
id: I20260101-1000-gate-was-wrong
brief: B20260101-0900-a.loop-brief   # "" for series-level
phase: halt          # setup | design | run | halt | verify | close | next
kind: repair         # direction | decision | context-supply | halt | verification-finding | repair | carry-forward | ceremony
automatable: partly  # yes | partly | no
by: operator         # operator | assistant | both
schema: intervention/v2
options: 2           # 0 to 3
recommended: 1       # the option the assistant recommended; 0 with no options
decided: 2           # 1 to 3, other, or "" with no options
agreement: other-option   # derived: recommended | other-option | adjusted | different | no-options
occurred: 2026-01-01
recorded: 2026-01-01T10:00:00Z
---
```

The frontmatter is the `intervention/v2` schema (`vloop schema show
intervention/v2`). A record written before options existed is `intervention/v1`
(`vloop schema show intervention/v1`): it has no `schema:` line, and is still
read, as a record with no options. Every field:

- `schema`: `intervention/v2` (the JSON form says `intervention/v1` or
  `intervention/v2`).
- `id`: the file name without `.md`.
- `brief`: the loop brief it belongs to, empty for a series-level one.
- `phase`: `setup`, `design`, `run`, `halt`, `verify`, `close` or `next`.
- `kind`: `direction`, `decision`, `context-supply`, `halt`,
  `verification-finding`, `repair`, `carry-forward` or `ceremony`.
- `automatable`: `yes`, `partly` or `no`: could a driver do it.
- `by`: `operator`, `assistant` or `both`.
- `options`: how many options the assistant proposed, 0 to 3.
- `recommended`: the number of the option the assistant recommended.
- `decided`: the option the operator chose, 1 to 3, or `other` when they chose
  something that was not among the options. The number, not the option's text:
  the text is in the Options section once.
- `adjusted`: `true` when the operator took a numbered option but changed it.
  Only with a numbered `decided`; otherwise the line is absent.
- `agreement`: derived from the three above, never written by hand:
  `recommended` (decided the recommended option as it stood), `other-option`
  (decided another numbered option), `adjusted` (decided a numbered option,
  changed), `different` (decided `other`: something none of the options
  offered) or `no-options` (the record has no options, so it cannot agree or
  disagree; every v1 record).
- `occurred`: the date it happened; `recorded`: an RFC 3339 date-time.
- `backfilled`: `true` on a record written after the fact.
- `summary`, `trigger`, `done`, `automation`: the summary line and the three
  sections, in the JSON form. `context`, `options` (the option texts),
  `recommended` (`option` and `why`), `suggested`, `decided` (`option`, `text`
  and `adjusted`) and `agreement` are the v2 additions.

The body sections of a v2 record: **Context.** names the ids it concerns (a
task, run, defect, intervention or commit), **Options.** lists the numbered
options, **Recommended.** gives the reason for the recommendation,
**Suggested.** holds anything the assistant suggested beyond the options, and
**Decided.** says what was decided, in the operator's words.

`vloop intervention add "<summary>" --phase <p> --kind <k> --automatable <a>
--by <b>` writes the file and prints its path; it also takes `--brief <name>`,
`--trigger`, `--done` and `--automation`, and the flags that record options:

- `--context <text>`: the Context section.
- `--option <text>`: one option; repeat it, up to three times.
- `--recommended <n>` and `--why <text>`: the recommended option's number and
  the reason for it.
- `--decided-option <n>`: the option the operator chose.
- `--decided-other`: the operator chose something none of the options offered;
  `--decided <text>` says what, in the Decided section.
- `--adjusted`: the chosen option was changed; it goes with `--decided-option`.

Without `--option` the record has no options and is `no-options`. A malformed
combination is refused, exit 2, and nothing is written:

- `at most three options`
- `options need --recommended and --why`
- `--recommended must name an option, 1 to <n>`
- `options need --decided-option or --decided-other`
- `--decided-option must name an option, 1 to <n>`
- `--decided-option and --decided-other exclude each other`
- `--adjusted needs --decided-option`
- `--recommended, --decided-option and --decided-other need options`

`vloop intervention list [--brief <name>]` prints `<phase>  <kind>
<automatable>  <agreement>  <id>`, by phase, kind and id.
`vloop intervention set <id> <field> <value>` changes `brief`, `phase`, `kind`,
`automatable`, `by`, `occurred`, and, on a record with options, `recommended`
(1 to the option count), `decided` (1 to the option count, or `other`) and
`adjusted` (`true` or `false`); the derived agreement is rewritten in the same
write, and deciding `other` clears `adjusted`. Two refusals, exit 2:

- `agreement is derived — set decided, adjusted or recommended instead`
- `options are recorded with the intervention, not set`

Setting `recommended`, `decided` or `adjusted` on a record with no options is
refused (`cannot set <field>: the record has no options`), exit 1.

`vloop intervention migrate [--dry-run]` moves every v1 record to
`intervention/v2`: it inserts the `schema`, `options`, `recommended`, `decided`
and `agreement: no-options` lines after `by:` and changes no other byte, so
`recorded` and `backfilled` stay as they were. `--dry-run` lists the records
it would migrate and writes nothing; a second run finds nothing to migrate.

`vloop intervention show <id>` prints the record, then a links block resolving
every task, run, defect, intervention and commit id named in its Context, in
order of first appearance, each with its title or `(not found)`; `--json` adds
`links` as `{ref, kind, title}`. An unknown id exits 1. Only Context is read,
so a link is something the recorder chose to name.

The per-brief `interventions` line of `vloop metrics` and the `--interventions`
table are in [metrics.md](metrics.md).

`vloop metrics export` emits one intervention record per intervention, after the
briefs (with `agreement` and the count of `options`, never their text), and every export record's `repo` carries `stacks`: the repository's
`metrics.stacks`, empty when unset.

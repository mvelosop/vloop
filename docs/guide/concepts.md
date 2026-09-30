# Concepts guide

The ideas behind vloop, in the order you meet them. Every command is in
[commands.md](commands.md), every setting in [configuration.md](configuration.md).

## The loop

A loop starts with a **plan**: a session reads a brief and splits it into
tasks. Then, for each task in turn, a **work** session does the task, a **gate**
runs its verify command, a **review** session judges the result independently,
and the driver makes one **commit**. Every session is a fresh Claude session
with no memory; the sessions share only files in the repository. A task that
fails its gate or its review is tried again, and a run that cannot go on halts
and says why.

## Briefs and their lifecycle

A brief is a Markdown file `B<YYYYMMDD-HHMM>-<slug>.loop-brief.md` in
`docs/briefs/`, written with `vloop brief new` and checked with `vloop brief
check`. Its `status` moves `draft` → `ready` → `consumed`, or to `abandoned`
if it is dropped. Only `ready` briefs are checked. `consumed` means a loop built
it and it was closed; `abandoned` means it never will be, and it blocks the
briefs that depend on it.

## depends-on

`depends-on` lists briefs, by name, that must be `consumed` before this one can
run. `vloop brief list` shows the briefs in dependency order, ready or blocked.
A name matching no brief, or a cycle, is a problem.

## Binding references

A brief's `## Binding references` lists the files the work is bound by, each
with the reason it binds. The plan attaches the relevant ones to each task as
`references`, and both the work and the review session read them: a convention
nobody points at is a convention that gets broken.

## Plans and tasks

The plan is `.vloop/state/state.json` (schema `state/v1`): a list of tasks, each
with a goal, acceptance criteria, a `verify` command, its dependencies, its
references and a status. The work session only proposes an outcome; the gate and
the review decide. `vloop status` and `vloop task list` read the plan, and the
`vloop task` commands that change it refuse a result that fails `vloop task
validate`.

## Gates and the gate shell

A task's `verify` command is its **gate**: the task is done only when it exits 0.
`vloop task gate <id>` runs it in the shell the `shell` key names (`sh` by
default, `pwsh` on Windows). A gate is written before the work exists, and the
work session may not change it; only the operator does, with `vloop task verify
<id> <command> --reason <text>`, which records the change in the task's history.

## Metrics and defects

`vloop metrics` reports what a brief's loop produced, what it cost and how long
it took, recomputed from run folders, git and the plan on every call; nothing
is entered by hand. See [metrics.md](metrics.md). A **defect** is something wrong
that the loop produced or let through, counted by where it entered and who caught
it. Some are derived from the run and some are recorded by hand with `vloop
defect add`. See [defects.md](defects.md).

## Closing and release

When a run is finished and you have verified it, `vloop brief close <brief>` on
the work branch records your findings as defects (`--finding`, or
`--no-findings` to say there were none), freezes the metrics as a snapshot, writes
the run record into the brief, marks it `consumed` and makes one commit. It
prints the `Vloop-Brief: <name>` trailer the squash-merge commit must carry.
`--abandon "<reason>"` closes an unfinished plan as `abandoned`, and `--dry-run`
shows what would happen and writes nothing. Close never merges and never pushes.

Release is not declared: vloop detects it as the brief's merge to the default
branch (`origin/HEAD`'s target, else `main`, else `master`), through that trailer
or the commit that marked the brief `consumed`. That is what lets `vloop defect
add --blame <file>:<line>` attribute a later bug to the brief that wrote the line.

## The .vloop/ layout

vloop keeps its files under `.vloop/` in the repository root:

- `.vloop/config.toml`: the settings, written by `vloop config set`.
- `.vloop/state/state.json`: the plan.
- `.vloop/state/runs/<run id>/`: a run's folders, sessions and iteration records.
- `.vloop/state/metrics/<run id>.json`: the metrics snapshot `brief close` freezes.
- `.vloop/defects/`: one Markdown file per recorded defect.

The shell loop's own state lives in `.loop/`; vloop only reads it.

## The flow

```
vloop brief new <slug>       write a draft brief, edit it
vloop brief check <path>     fix what it reports, set status: ready
(run the loop)               plan, then work → gate → review → commit per task
vloop defect add …           record what you find while verifying the run
vloop brief close <brief>    findings, snapshot, run record, consumed, one commit
(squash-merge)               with the trailer Vloop-Brief: <brief>
vloop defect add --blame …   after release, attribute a bug found later
vloop metrics [<brief>]      what it cost and delivered
vloop metrics export         the same as JSON Lines, numbers and titles only
vloop metrics --workspace <file>   one table across several repositories
```

`vloop metrics export` writes `export/v1` records with the repository's identity,
and `--workspace` reads the repositories a workspace file lists (see
[configuration.md](configuration.md)) and only reads them: each repository
owns its data.

---
name: metrics
description: Binds the Metrics entity — what is measured per brief and per task, how lines are classified into code, test, docs, other and excluded through layers and stack presets, delivered versus churn, and how release and the snapshot relate to it
---
# Metrics

A brief's numbers, recomputed on every call from raw sources (M-1), frozen once
at close as a snapshot (`metrics/v2`; v1 records and snapshots stay readable). *Part of
[measurement](measurement-context.md).* Every field and formula:
[`docs/guide/metrics.md`](../../guide/metrics.md).

## What is measured

| Group | Per brief | Per task |
| --- | --- | --- |
| tasks | planned, done, blocked, first-pass, the brief's estimate (B-4), iterations per closed task | attempts |
| size | delivered and churned lines per category, deletions, rework, test:code | churned lines per category, test:code |
| time | agent (work + review), plan, gates, checks (`checks_ms`), gate review (`gate_review_ms`), wall | agent |
| rate | delivered code lines per agent minute, with and without tests | — |
| cost | total and per phase (gate review included), `$/kloc` (per 1,000 delivered code lines); tokens and cache-hit ratio | cost, `$/kloc` |
| models | the models actually used per phase; the configured effort | the work model |
| defects | in-loop, operator (recorded operator defects before release plus verification-finding interventions with an origin), escaped (by time against release), removal efficiency; a `check_failed` iteration derives a `work` defect found by the gate, counted apart as check failures | — |
| instructions | words each session kind starts with (skill + `.vloop/repo-guidelines.md` + `CLAUDE.md`) and the guidelines' rule count; `n/a` for a kind with no skill | — |
| records | session records that should exist and do not | — |
| lead time | created → planned → completed → merged | — |

## Lines

- A line counts when it is **added and non-blank** (M-2); deletions are
  reported apart; binary files are not counted.
- Session time comes from a record's `wall_ms` when it has one and from
  `duration_ms` otherwise; the brief's wall runs from the first plan session's
  `started` (the plan commit's time when there is none) to the last run commit.
- A brief that delivered no line classified as code or test shows `n/a` in the
  cross-brief table's code and test columns, and the summary adds a line giving
  how many delivered lines match no stack in `metrics.stacks`; JSON is unchanged.
- **Delivered**: the diff from the plan commit's parent to the brief's last run
  commit. **Churn**: every task commit, whatever its outcome, summed.
  **Rework** = churned code / delivered code.

## Instructions

The summary's `instructions` line, after `interventions`, gives the words each
session kind (plan, gate review, work, review) starts with, counted at the
commit the brief's delivered lines are measured to, and the number of rules in
the guidelines file. `--json` carries the same as an optional `instructions`
object (`plan`, `gate_review`, `work`, `review`, `guidelines_rules`; `null` for
a kind with no skill), absent in snapshots written before it existed. Computed
from git (M-1); nothing is recorded during a run.

## Classification

Five categories — `code`, `test`, `docs`, `other`, `excluded` — decided by the
first layer that matches a path (M-3):

1. **Always excluded**: `.vloop/**`, `.loop/**`.
2. **The repo's own globs**: `metrics.excluded`, `metrics.test`,
   `metrics.docs`, `metrics.code`, in that order.
3. **The stack presets** the repo names in `metrics.stacks` — `go`,
   `typescript`, `javascript`, `react`, `csharp`, `python`, `java`, `kotlin`,
   `rust` — merged, in the same order.
4. Nothing matched: `other`.

**Scopes.** An entry of `metrics.stacks` is `<stack>` or `<stack>@<path>`, with
a repo-relative path naming an existing directory (`csharp@services/api`). A
path's scope is the longest scoped path that contains it; with none, the
unscoped stacks apply. Inside a scope only that scope's stacks apply, and their
globs match relative to the scope's directory. The repo's own globs (layer 2)
stay repo-relative and unscoped. `vloop metrics classify` labels a scoped match
`(<stack>@<path>: <glob>)`.

The repo's globs come first because extension alone lies: vloop's brief
templates and its schemas are product code in `.md` and `.json` files.
`vloop metrics classify <path>` shows which layer and glob decided a path.

## Release and the snapshot

- A brief is **merged** at the first commit on the default branch where it is
  `consumed` (M-5). Metrics show `merged <sha>` or `not merged`.
- The **snapshot** at `.vloop/state/metrics/<run id>.json` is what `close`
  froze. Live metrics may later differ — a config change reclassifies lines,
  a defect recorded after release lowers removal efficiency — and that difference
  is the point: the snapshot is what the operator saw at close.

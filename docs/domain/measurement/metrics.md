---
name: metrics
description: Binds the Metrics entity — what is measured per brief and per task, how lines are classified into code, test, docs, other and excluded through layers and stack presets, delivered versus churn, and how release and the snapshot relate to it
---
# Metrics

A brief's numbers, recomputed on every call from raw sources (M-1), frozen once
at close as a snapshot (`metrics/v1`). *Part of
[measurement](measurement-context.md).* Every field and formula:
[`docs/guide/metrics.md`](../../guide/metrics.md).

## What is measured

| Group | Per brief | Per task |
| --- | --- | --- |
| tasks | planned, done, blocked, first-pass, the brief's estimate (B-4), iterations per closed task | attempts |
| size | delivered and churned lines per category, deletions, rework, test:code | churned lines per category |
| time | agent (work + review), plan, gates, wall | agent |
| rate | delivered code lines per agent minute, with and without tests | — |
| cost | total and per phase, per 1,000 delivered code lines; tokens and cache-hit ratio | cost |
| models | the models actually used per phase; the configured effort | the work model |
| defects | in-loop, operator, escaped, removal efficiency | — |
| records | session records that should exist and do not | — |
| lead time | created → planned → completed → merged | — |

## Lines

- A line counts when it is **added and non-blank** (M-2); deletions are
  reported apart; binary files are not counted.
- **Delivered**: the diff from the plan commit's parent to the brief's last run
  commit. **Churn**: every task commit, whatever its outcome, summed.
  **Rework** = churned code / delivered code.

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

The repo's globs come first because extension alone lies: vloop's brief
templates and its schemas are product code in `.md` and `.json` files.
`vloop metrics classify <path>` shows which layer and glob decided a path.

## Release and the snapshot

- A brief is **merged** at the first commit on the default branch where it is
  `consumed` (M-5). Metrics show `merged <sha>` or `not merged`.
- The **snapshot** at `.vloop/state/metrics/<run id>.json` is what `close`
  froze. Live metrics may later differ — a config change reclassifies lines,
  a defect found after release lowers removal efficiency — and that difference
  is the point: the snapshot is what the operator saw at close.

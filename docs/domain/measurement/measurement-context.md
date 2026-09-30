---
name: measurement-context
description: Binds the measurement context — Metrics and line classification, Defects, release, closing and its snapshot, export and workspaces — as a reader of execution that writes only its own records, and the invariants M-1..M-7 it enforces
---
# Measurement

What a brief produced, what it cost, what went wrong and where it was caught —
per brief, per task, and across briefs and repositories.

| Entity | Page |
| --- | --- |
| Metrics, line classification, release | [metrics.md](metrics.md) |
| Defect, derived and recorded | [defect.md](defect.md) |

The how-to lives in the guide: [`docs/guide/metrics.md`](../../guide/metrics.md)
defines every metric and formula; [`docs/guide/defects.md`](../../guide/defects.md)
the defect workflow. This page and its entities bind the concepts those pages
use.

## A reader, then a closer

- **Reading.** Everything measurement reports is computed from what execution
  left: run folders, the brief's commits, the plan at its last commit, and the
  defect files. `vloop metrics` writes nothing (M-1).
- **Closing.** `vloop brief close` is the one command that turns measurement into
  records: it writes the operator's findings as defect files, freezes the
  metrics as a snapshot, writes the brief's run record, marks it consumed, and
  makes one commit with the trailer the squash-merge must carry (M-6).
- **Releasing** is detected, not declared (M-5); it moves no file.
- **Consolidating** only reads: `metrics export` prints JSON Lines carrying
  numbers and titles (M-7); `--workspace` reads several repositories' metrics
  into one table. Each repository keeps owning its data.

## Invariants it enforces

M-1..M-7 in [`../domain-model.md`](../domain-model.md#measurement--m).

## Gaps

- Gate time and configured effort are `n/a` for shell-loop runs.
- The snapshot is written at close only; `vloop run` will also write it after
  every iteration *(planned, B6)*.
- Interventions (`.vloop/interventions/`) are measurement's likely next entity:
  recorded as data, analysed by hand, no command or schema yet.

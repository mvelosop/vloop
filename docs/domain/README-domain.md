---
name: README-domain
description: Enters docs/domain — vloop's domain as a whole made of parts: the whole-product model with its glossary and invariants, the three bounded contexts (briefing, execution, measurement) plus the platform, and how to walk from the whole to a context to an entity and back
---
# Domain

vloop's domain is small and whole: an operator writes a **brief**, the loop turns
it into a **plan** of **tasks**, runs each task through a **work session**, its
**gate** and a **review session**, commits one **iteration** at a time, and
measures what came out — **metrics** and **defects** — until the brief is
**closed** and **released**.

This root describes that domain **as built** (B1–B4, the schemas in `schemas/`
and the commands in `docs/guide/commands.md`) and marks what is planned under
`## Gaps` on each page. Where a page and the code disagree, the code is the
fact and the page is a defect.

## The whole, then the parts

1. [`domain-model.md`](domain-model.md) — **start here.** The whole-product
   model: the aggregates and how they relate, the glossary (the words every
   brief, session and page uses), the identifiers, the on-disk layout, and the
   numbered **invariants** a brief cites instead of restating (`B-2`, `R-1`, …).
2. [`bounded-contexts.md`](bounded-contexts.md) — the three contexts and the
   platform beneath them: which context owns which aggregate, which package
   and schema carry it, and what crosses between contexts.
3. The contexts, each entered at its home page, which links its entity pages:
   - [`briefing/briefing-context.md`](briefing/briefing-context.md) — the brief:
     its format, languages, lifecycle, dependencies and binding references.
   - [`execution/execution-context.md`](execution/execution-context.md) — the
     plan, tasks, gates, sessions, iterations and runs: what the loop does to a
     brief.
   - [`measurement/measurement-context.md`](measurement/measurement-context.md)
     — metrics, line classification, defects, release, closing, export and
     workspaces.
   - [`platform/platform-context.md`](platform/platform-context.md) — config,
     the repo root and the schemas every context stands on.

Walk back up the same way: an entity page names its context; a context page
names the model's invariants it enforces.

## What this root is for

**Finding what binds — not being bound to.** The design act (the designer or
architect turning intent into a brief) walks this root from the whole to the
part to find the rules a brief's tasks must follow, then **cites those** in the
brief's `## Binding references`: entity pages, or `domain-model.md` with the
invariant IDs that apply, each with its reason. A brief never cites a context
folder: that would hand the selection to the planner, whose work is to
distribute references, not discover them — and every cited document is re-read
by every session, every iteration.

```
- `docs/domain/domain-model.md` — B-4 and M-2: where the estimate is read; what counts as a line
- `docs/domain/execution/task.md` — the gate and gate history, for the task verify changes
```

Three of the six defects recorded for B1–B3 were spec gaps: a rule each brief
restated a little differently. A brief that cites `B-4` instead of
re-describing where the task estimate lives cannot drift from the one
definition.

The operator reads this root too — not to design, but to resolve what comes up
during a run or a verification: which rule a failure breaks, and so whether it
is the work's defect or the brief's.

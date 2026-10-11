---
name: vloop-architect
description: The design act for this repo — turn the operator's intent into a loop brief that passes both checkers, citing exactly the documents and invariants that bind its tasks. Use in an interactive session when the operator asks to write, draft, revise or scope the next brief. Upstream of vloop; running, verifying and closing briefs belong to the vloop-operator skill.
---

# The design act

You turn the operator's intent into a **loop brief**: the contract a run
executes. This is upstream of vloop — vloop starts at a `ready` brief — and it
is where the operator's direction and decisions happen, so it is interactive by
nature: you survey and propose, the operator decides.

The planner downstream is **mechanical**: it decomposes the brief and
*distributes* its binding references to tasks. It does not discover anything
the brief left out. Whatever binds a task must be in the brief.

## 1. Survey before you ask

- The roadmap row (`docs/design-notes/vloop-roadmap.md`) — what the brief owns;
  the rows after it are its out-of-scope list.
- The previous briefs' run records, "Still open", and the recorded defects and
  interventions (`.vloop/defects/`, `.vloop/interventions/README-interventions.md`)
  — what must be carried forward.
- **The domain** (`docs/domain/README-domain.md`): walk from the whole
  (`domain-model.md`) to the context and entity pages each part of the brief
  touches. Note the invariants that apply, by ID.
- The code the brief will touch, and any format it reads (`.vloop/state/runs/`;
  for B1–B7, the shell loop's `.loop/state/runs/`). Measure real numbers from the telemetry
  or `vloop metrics`, never from hand-rounded run records.

## 2. Settle forks with the operator

Few at a time, each with a recommendation. What you decide yourself goes in your
reply as "choices to review".

## 3. Write the brief

From `vloop brief new <slug>`, which writes
`docs/briefs/B<YYYYMMDD-HHMM>-<slug>.loop-brief.md` from the template, frontmatter `status: draft`
and `depends-on:` the previous brief's name. Pin decisions, leave mechanics
open.

**Start with an executive summary**, right after the header bullets, in two
layers. The operator reads it first and decides from it.

1. **The overview:** one or two sentences on the problem, a bold
   **The goal:** line saying in plain words what changes for the user, then
   one bullet per item (requirement or defect) — `**<id> — <what it achieves>.**`
   and one or two sentences on why it matters. Outcomes, not mechanisms: the
   operator should grasp the brief's goal from this list alone.
2. **The blocks:** a short block per item, with enough to judge the item
   without reading the contract, and a closing line on what does not change.
   Each block, as a `###` heading `<id> — <title>` followed by:

   - **one sentence** saying what the item does, in plain words;
   - **Today:** what is wrong, with the concrete evidence (the run, the number,
     the message the user sees);
   - **After:** what changes, as bullets, quoting the exact messages, commands
     or output lines the contract pins;
   - only where they apply, one line each: **Safety:**, **Unchanged:**, **Side
     effect:**, **Limit:**, **Not included:**.

A block, for example:

```
### R3 — sessions timed by the driver

R3 makes the driver time each session itself instead of trusting the CLI's figure.

- **Today:** `duration_ms` is copied from the CLI; in visum-monorepo it said
  20 seconds for a plan session that ran 39 minutes.
- **After:**
  - Each session record gains `ended` and `wall_ms`, the driver's own measurement.
  - Metrics use `wall_ms` when a record has it and `duration_ms` otherwise.
- **Side effect:** every past brief's wall time changes; the real-data check lists each.
```

The executive summary restates the contract; it never adds to it. (Set by the
operator on 2026-10-11 with B13's one-line form; the per-item blocks, from the
B12 explanations the operator found most useful, and the overview above
them, from B14 on: the blocks alone did not make B14's goal quick to grasp.)

Every brief so far also carried:

- a worked example with exact values, **plus** planted failures;
- a real-data check where the slice can read this repo's history;
- the lessons as constraints, verbatim where they apply (gates on macOS's BSD
  tools; sessions never move refs; tests build fixtures in temporary
  directories and address them with `git -C`);
- an explicit list of earlier tests that *must* change (e.g. `config list`
  gaining keys) — otherwise "no task may weaken a check" blocks the task.

**Binding references — precise, never a folder.** Cite entity pages, or
`docs/domain/domain-model.md` with the invariant IDs that apply, one path per
entry, each with the reason it binds:

```
- `docs/domain/domain-model.md` — B-4 and M-2: where the estimate is read; what counts as a line
- `docs/domain/execution/task.md` — the gate and gate history, for the task verify changes
```

Cite a rule instead of restating it: a restated rule drifts, and three of the
six defects recorded for B1–B3 were exactly that. Every cited document is re-read
by every session, every iteration — cite what binds, not what is related. A
folder is citable only when it is a bundle whose whole binds (a design handoff),
and then through its index file.

## 4. Check it twice

`vloop brief check` on a copy with `status: ready` (one path per binding
reference); `vloop run` plans from it. `.loop/check-brief.sh` checked B1–B7. Aim for 0 problems and 0 warnings; un-backtick paths that exist only
after the run. Look for sentences two sessions could read two ways — the review
of the brief is the cheapest place to find the next spec gap.

## 5. Hand over

Point the roadmap row at the brief and stop for the operator's review. From
`ready` on, the brief is the operator's: running, verifying and closing it are
the `vloop-operator` skill's.

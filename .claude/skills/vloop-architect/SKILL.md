---
name: vloop-architect
description: The design act for this repo — the vloop-create-brief skill, plus what this repository adds to it (the roadmap row, the domain's invariants cited by ID, the constraints every brief here carries). Use in an interactive session when the operator asks to write, draft, revise or scope the next brief. Running, verifying and closing briefs belong to the vloop-operator skill.
---

# The design act here

**Follow the `vloop-create-brief` skill** for the whole act: find the
documentation, survey, settle forks with the operator, write from the template,
check it, have it reviewed with `vloop-review-brief`, hand over at `draft`. This
skill only adds what is particular to this repository.

## Scope comes from the roadmap

The brief's row in `docs/design-notes/vloop-roadmap.md` says what it owns; the
rows after it are its out-of-scope list. At hand-over, point the row at the
brief.

## Bind by invariant

Walk the domain (`docs/domain/README-domain.md` → `domain-model.md` → the
context and entity pages each part of the brief touches) and cite the
invariants that apply by ID, one path per entry:

```
- `docs/domain/domain-model.md` — B-4 and M-2: where the estimate is read; what counts as a line
- `docs/domain/execution/task.md` — the gate and gate history, for the task verify changes
```

## What every brief here carries

- Real numbers from the telemetry (`.vloop/state/runs/`, `vloop metrics`), never
  from hand-rounded run records. B1–B7's history is the shell loop's, under
  `.loop/state/runs/`.
- A real-data check where the slice can read this repository's own history.
- These lessons as constraints, verbatim where they apply: gates run on macOS's
  BSD tools; sessions never move refs; tests build their fixtures in temporary
  directories and address them with `git -C`.
- In `vloop brief check`, un-backtick paths that exist only after the run.

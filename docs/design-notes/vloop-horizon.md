---
name: vloop-horizon
description: Ideas beyond the current roadmap, kept apart so they cannot leak into it. First entry — a project-level driver over a project-state.json that schedules briefs by their dependencies, several at once. Not a plan; nothing in the roadmap depends on it.
kind: design-note
status: idea
created: 2026-09-30
---
# vloop — horizon

The roadmap (`docs/design-notes/vloop-roadmap.md`) is the plan: B1–B6, each
brief owning a slice, built in order. This note holds what comes **after** it,
or beside it, as ideas. Nothing here binds a brief; a brief may cite an entry
only once the operator has promoted it into the roadmap.

## H1 — a project-level driver over `project-state.json`

*Noted 2026-09-30, by the operator.*

### The observation

The roadmap has turned into a runbook. Each row names a brief, what it owns and
what it depends on, and the notes around it carry the context needed to seed
the next brief. That is the same shape one level up as `state.json` is to a
brief: a list of units of work with dependencies, a status each, and a driver
that picks the next ready one. At the task level the driver is `run.sh` (later
`vloop run`); at the project level it is, today, the operator.

### The idea

A `project-state.json` — the roadmap as data — and a driver over it:

- **Entries are briefs**: name, status (`idea`, `draft`, `ready`, `running`,
  `verifying`, `consumed`, `abandoned`), `depends-on`, and the context that
  seeds the brief (what it owns, the decisions it inherits, the lessons in its
  constraints).
- **The driver schedules**: every brief whose dependencies are consumed is
  ready, and ready briefs can run **concurrently**, each on its own work branch
  in its own worktree, each a `vloop run` with its own budget.
- **The roadmap becomes a view** rendered from `project-state.json`, the way
  `plan.md` is rendered from `state.json` — read by humans, never parsed back.

### Why it fits what exists

- `depends-on`, `brief list` (dependency order, ready vs blocked) and the
  `abandoned` status already model the graph.
- Runs are already isolated per work branch, and metrics and defects are keyed
  by brief, so concurrent runs aggregate without collision; `--workspace`
  already reads across repositories.
- The operator skill already describes the per-brief loop the project driver
  would repeat.

### Evidence so far

`docs/design-notes/vloop-interventions-b1-b4.md`: 38 operator interventions
across B1–B4, by kind and by whether a driver could do them. In short: ceremony
and halts are automatable, verification mostly is (by gates on real data and
log checks), direction and decisions cluster in the design act and stay human,
and carrying context forward is the gap.

### What it has to answer first

- **The human gates between briefs.** A dependency is complete when a brief is
  *consumed and merged* — which today means the operator verified it, recorded
  findings and merged. Concurrency multiplies runs, not operators. Which of those
  steps can be automated, and which must stay a human decision, decides how much
  parallelism is real.
- **Briefs that do not exist yet.** Roadmap rows say "not written". Writing a
  brief is the design act, which is upstream of vloop by decision. Either the
  project driver only schedules briefs a human made `ready`, or it gains a
  design step — a much bigger claim.
- **Conflicts between concurrent briefs.** Two briefs touching the same files
  merge badly. The roadmap's "Owns" column is informal; the driver would need
  ownership it can check (files or areas per brief), or serialise overlapping
  briefs.
- **Budgets and pacing**: a project-wide cost ceiling, how many runs at once,
  and what one halted brief does to its dependents.
- **Rebasing**: a brief planned against a base that moves while a sibling
  merges. Re-plan, rebase, or pin and merge in order?
- **Lessons flowing forward**: today the operator carries a run's lessons into
  the next brief's constraints by hand (BSD tools, refs). A project driver needs
  somewhere to put them that the next brief's design step reads.

### Not before

`vloop run` (B6) replacing the shell loop, and `brief close` (B4) making a
brief's completion a single recorded event. A project driver over the shell loop
would be scheduling around the gaps B3's run exposed.

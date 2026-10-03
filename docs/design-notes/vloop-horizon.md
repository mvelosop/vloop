---
name: vloop-horizon
description: What comes after the current roadmap, as Now / Next / Later — configurable merge and branch strategy (now), the documentation layer for the design act (next), a project-level driver over a project-state.json (later). Kept apart from the roadmap so ideas cannot leak into it; nothing in the roadmap depends on it.
kind: design-note
status: idea
created: 2026-09-30
---
# vloop — horizon

The roadmap (`docs/design-notes/vloop-roadmap.md`) is the plan to v1.0: one
brief per slice, built in order. This note holds what comes **after** it, as
**Now / Next / Later**:

- **Now** — the first thing after v1.0; decided enough to brief.
- **Next** — wanted, with open questions.
- **Later** — an idea worth keeping, not yet worth designing.

Nothing here binds a brief; a brief may cite an entry only once the operator has
promoted it into the roadmap.

## Now

### The gate model: what a gate is, who reviews it, and the checks the driver runs

*Noted 2026-10-03, by the operator, from B8's run. The first brief after v1.0.*

B8 showed the gate concept is underdefined: two of its 18 gates could never
pass (a fixture sharing the stub's directory; a brief check of the running
brief), two regressions went past per-task gates (T5, T15), its gates ran to
3,000–10,000 characters, and F9's gate-file rule froze a task's own product
(I20261003-1147). Direction set with the operator on 2026-10-03:

- **A gate judges a contract** from the brief, by exercising the product. Its
  judge — expected values, assertions — is in the verify itself, or in fixtures
  the planner writes under a plan-owned directory (such as
  `.vloop/state/gates/<task id>/`) that sessions cannot write. **A gate never
  judges by a file a task writes, and never runs the tests.**
- **The task's `files` property goes**, and with it gate-file detection: there
  is nothing in a task's reach to protect.
- **The driver runs an authoritative check command** after every iteration —
  the repository's tests and health checks (vet, brief check…), written by the
  planner or configured, required to pass at the plan's base, and responsible
  for its own flaky tests. Its result is the reviewer's evidence; reviewers stop
  re-running tests.
- **A gate review step**: an independent session confirms every gate is a
  contract check, fails on the base for the right reason, and keeps its judge out
  of tasks' reach, before any work.
- An earlier brief's test that a task must change is a contract change: the
  brief lists it, and the check command judges it.
- From B8's verification: a timed-out gate costs twice the timeout, because a
  failed gate is re-run once to detect flakiness.

### The quality pass (was B9)

*Moved out of the v1.0 series by the operator on 2026-10-03.*

The quality findings of the v1.0 review, in full in
`docs/design-notes/vloop-v1-review.md` → "B9 — quality": the guides against the
binary, error messages and exit codes, help text, code health, the test suite's
speed. **The README rewrite comes after the gate model**, so it describes gates as
they will be.


### Configurable merge strategy and work-branch retention

*Noted 2026-10-01, by the operator.*

Today a brief's work branch is squash-merged and kept, by convention the
operator (or `/vloop:operate`) follows: the squash keeps the default branch one
commit per brief, and the kept branch is the evidence of how it was built — one
commit per iteration, each by a fresh session. That suits this repository; it
will not suit every team.

- **Merge strategy**: `squash` (today), `merge` (keep the iteration commits on
  the default branch), `rebase`. A config key, with the trailer `Vloop-Brief:`
  carried by whichever commit the strategy produces, so `--blame` and release
  detection keep working.
- **Branch retention**: `keep` (today), `delete` after merge, or `archive` (a
  tag such as `vloop/<run id>` before deleting), so the evidence survives a team
  that prunes branches.
- `brief close` prints the merge command for the configured strategy;
  `/vloop:operate` follows it. vloop still never merges by itself.

Not for v1; ready to brief once v1.0 is tagged.

### Interventions that measure the model against the operator, and explain themselves later

*Noted 2026-10-01, by the operator.*

B7 makes interventions a vloop record (`intervention/v1`). Two things it does
not yet capture, both needed to learn from the data rather than only count it:

- **The model's suggestion, beside the operator's decision.** Each record gains
  `suggested` — what the model (the assistant acting as the operator's hands, or
  a session) proposed, in a sentence, or `none` when it proposed nothing — and
  `decided` — what the operator chose. A derived `agreement` (`same`,
  `adjusted`, `different`, `no-suggestion`) lets `vloop metrics` report, per kind
  and phase and across a workspace, how far the model's judgement is from the
  operator's. That is the evidence for which interventions a driver could take
  over: the kinds where the model already agrees are the candidates.
- **Context for review after the fact.** A record read weeks later has to stand
  on its own. Each gains, where they apply: the run, iteration and task it
  happened in; the commit or files involved; the triggering message or output,
  quoted briefly; the alternatives considered; and what changed because of it
  (a defect recorded, a gate amended, a brief rescoped). `vloop intervention show
  <id>` prints the record with those links resolved.

Settled with the operator, 2026-10-01: `suggested` is **written at the time** —
honest, and only what the model actually proposed; a record with no suggestion
says `none`. The context is **one succinct paragraph**, so records stay cheap
enough to keep being written. The 61 records to date were given their context,
suggestion and decision on 2026-10-01, in their bodies, while the session that
made them still held the history.

## Next

### The documentation layer for the design act

*Parked 2026-09-30, by the operator, to keep the focus on v1.*

The domain root (`docs/domain/`) exists and is what the design act walks to cite
precise rules. Still parked: a use-cases root (the operator and the session
kinds as actors, the use cases a project driver would automate, each linked to
the interventions seen in it), an `architecture-decisions/` root (starting with
"the brief carries precise references; planners distribute, never discover"),
and a top-level `docs/README-docs.md`.

### Interventions across project types

Interventions and defects are recorded per repository. Their value is in the
patterns across projects — which kinds of intervention a NestJS service needs
that a Go CLI does not — so the export carries each repository's stacks, and
analysis across a workspace can group by them. What remains open is the
analysis itself: which questions to ask of the data, and what would change in
the loop because of the answers.

## Later

### A project-level driver over `project-state.json`

*Noted 2026-09-30, by the operator.*

#### The observation

The roadmap has turned into a runbook. Each row names a brief, what it owns and
what it depends on, and the notes around it carry the context needed to seed
the next brief. That is the same shape one level up as `state.json` is to a
brief: a list of units of work with dependencies, a status each, and a driver
that picks the next ready one. At the task level the driver is `vloop run`; at
the project level it is, today, the operator.

#### The idea

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

#### Why it fits what exists

- `depends-on`, `brief list` (dependency order, ready vs blocked) and the
  `abandoned` status already model the graph.
- Runs are isolated per work branch, and metrics and defects are keyed by brief,
  so concurrent runs aggregate without collision; `--workspace` already reads
  across repositories.
- `/vloop:operate` describes the per-brief loop the project driver would repeat.

#### Evidence so far

`docs/design-notes/vloop-interventions-b1-b4.md`: 38 operator interventions
across B1–B4, by kind and by whether a driver could do them. In short: ceremony
and halts are automatable, verification mostly is (by gates on real data and
log checks), direction and decisions cluster in the design act and stay human,
and carrying context forward is the gap. The records since B5 are in
`.vloop/interventions/`.

#### What it has to answer first

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

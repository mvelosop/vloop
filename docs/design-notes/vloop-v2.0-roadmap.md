---
name: vloop-v2.0-roadmap
description: The briefs that took vloop from v1.0 to v0.8.0 (planned as v2.0; renumbered before going public) — the gate model, interventions that weigh the model's options against the operator's decision, and the quality pass — what each owns, depends on, and the references each brief must cite. Read before writing any of them; each section is that brief's design input.
kind: design-note
status: draft
created: 2026-10-03
---
# vloop — the v2.0 roadmap, released as 0.8.0

**Renumbered.** This series was planned and built as v2.0 and released as
**v0.8.0**: before going public, the operator renumbered it (`docs/guide/concepts.md`
→ "What changed in 0.8"; `.vloop/interventions/I20261006-1237-renumbered-to-0-8-0-before-going-public.md`).
The `v1.0.0` and `v2.*` tags were deleted then (`docs/guide/concepts.md` lists
them and their commits, which stay). "v1.0"
and "v2.0" in the sections below are that history and stay as written; the
file keeps its name because consumed briefs cite it.

v1.0 (tag `v1.0.0`, `fc0b67a`) is B1–B8, as `docs/design-notes/vloop-roadmap.md`
records. This note is the plan to the next release, built from the horizon's
"Now" entries the operator promoted on 2026-10-03; this note now holds them, and
`docs/design-notes/vloop-horizon.md` points here. The configurable merge strategy stays on
the horizon.

**Why 2.0, not 1.1.** B9 changes vloop's public contracts. The task loses
`files` and the plan gains `check` (`state/v1` → `state/v2`), a gate gains a judge it may not share with a
task, and a gate review can reject a plan that v1 accepted. Plans and briefs
written for v1 may not plan under v2. That is a breaking change, so the next
version is major. B10 and B11 alone would be minor.

Every brief here is run by `vloop run` from the latest released vloop, never
from a binary built from the tree it changes (design session, section 6). In
practice that is v1.0.0 for B9; the release after a brief may be cut before the
next one if the operator wants the driver it adds. Each brief names its
predecessor in `depends-on:`.

| # | Brief | Owns | Depends on |
| --- | --- | --- | --- |
| B9 | `docs/briefs/B20261003-2049-gate-model.loop-brief.md` — **consumed** | **The gate model.** A gate judges a contract, by exercising the product, with its judge — the verify and planner-owned gate fixtures — out of every task's reach. A gate may be the planner's oracle test; it never runs the repository's suite. The task's `files` goes, and so does gate-file detection. The plan gains named **checks**, scoped by path, that the driver runs, authoritatively, after every iteration and once in full at the end. The driver runs every gate on the base; an independent **gate review** session precedes the work. `state/v2` | B8 |
| B10 | `docs/briefs/B20261004-1434-interventions-options.loop-brief.md` — **consumed** | **Interventions that weigh the model against the operator.** Each record carries the model's **three options**, the one it recommends **and why**, and the operator's decision. A derived `agreement` per record; `vloop metrics` reports it by kind and phase; a self-contained context paragraph; `vloop intervention show`. `intervention/v2` | B9 |
| B11 | `docs/briefs/B20261005-0933-quality-pass.loop-brief.md` — **consumed** | **The quality pass**, for everything a user meets: exit 2 for usage only, actionable messages, help, the guides, `doctor`'s plugin check, metrics ordering and counts, one CRLF- and BOM-safe frontmatter reader and rewriter, the evals' graders; the suite faster first and **the README rewritten last** | B10 |
| — | **v0.8.0**, tagged after B11 (planned as v2.0.0) | | |
| T1 | `docs/design-notes/vloop-t1-triage.md` — **done** (2026-10-08), a **design act**, not a loop brief | **Triage of defects and interventions** (B1–B11, with B10's options and agreement): what they say about the skills and about repo-specific guidelines that could automate interventions; a design note whose proposals seed later briefs | v0.8.0 |

After B11: v0.8.0 was tagged by the operator (planned as v2.0). Then the triage (T1), set by the operator on 2026-10-05; the series it set, B12–B17, is planned in `docs/design-notes/vloop-roadmap-b12-b17.md`. The release steps are in
`docs/guide/concepts.md` → "The stamped release build".

## Decisions every brief in this series inherits

- Everything in `docs/design-notes/vloop-roadmap.md` → "Decisions every brief in
  the series inherits" still holds, except where a brief below changes it on
  purpose and says so.
- **The driver is the boundary** (S-2, S-4, R-4, from B8). No brief here may
  move enforcement back into the fence or into a skill's instructions alone.
- **Schemas version on a breaking change.** `state/v2` and `intervention/v2` are
  new schemas beside the v1 ones. vloop reads v1 records (closed briefs keep
  their metrics) and writes v2. `vloop upgrade` does not rewrite past records.
- **Evals are the skills' gates.** A brief that changes a skill's behaviour
  changes or adds the eval cases that pin it (`plugin/evals/`), and its run
  record states the scores and spend (`docs/guide/evals.md`).

---

## B9 — the gate model

### What it owns

**Amended at the design act, 2026-10-03** (the brief is binding where the two
differ): "never runs the tests" means the repository's suite, not the planner's
oracle tests, which may run from a git-ignored scratch folder; the plan's
`check` is several named checks from config, scoped by path, plus a full pass
at the end; the driver runs every gate on the base and the gate review judges
why it fails. Background: exploring-claude's `.loop/oracles/` and the lesson of
its SPA-208 run (gate fixtures locked against sessions, amendable by the
operator).

Direction set with the operator on 2026-10-03, from B8's run (promoted from the
horizon; this section is now its record):

- **A gate judges a contract** from the brief, by exercising the product. Its
  judge — expected values, assertions — is in the verify command itself, or in
  fixtures the planner writes with the plan under a plan-owned directory (such as
  `.vloop/state/gates/<task id>/`), which sessions cannot write (S-4).
- **A gate never judges by a file a task writes, and never runs the tests.** The
  tests are the work session's tool for its own quality.
- **The task's `files` property goes**, and with it gate-file detection
  (`internal/driver/gates.go`): nothing in a task's reach needs protecting.
- **The driver runs an authoritative check command** after every iteration: the
  repository's tests and health checks (vet, `vloop brief check`…). It is the
  plan's **`check`** property in `state/v2`, beside the tasks — one command for
  the whole plan, as each task's `verify` is one command for its gate. It must pass at the plan's base. It
  is responsible for its own flaky tests (for example, re-running failures
  individually); vloop treats its exit code as the truth. Its result is the
  reviewer's evidence, so reviewers stop re-running tests.
- **A gate review step**: an independent session, after planning and before any
  work, confirms that every gate judges a contract, fails on the base for the
  right reason, and keeps its judge out of tasks' reach. A failing gate review
  sends the plan back to the planner, or to the operator.
- **An earlier brief's test that a task must change is a contract change.** The
  brief lists it under the tests that must change; the check command judges it.

### Evidence

- B8: two of 18 gates could never pass — a fixture sharing the stub claude's
  directory (T12), a brief check of the running brief (T17). Two regressions
  passed per-task gates (T5, T15) and surfaced only in the one gate that ran the
  whole suite. Gates of 3,000–10,000 characters, one line each. F9's gate-file
  rule froze a task's own product.
- The plan skill prescribes the pattern this replaces: the task ships its tests
  in `files`, and the gate runs them "alongside whatever independent checking it
  does" (`plugin/skills/plan/SKILL.md` → "What the gate does not outlive").
- vloop's planner costs about twice the shell loop's on the same models (the
  url-shortener benchmark), plausibly from checking gates on the base. The
  benchmark is the place to see whether a gate review moves that cost or adds to
  it.
- From B8's verification: a timed-out gate costs twice the timeout, because a
  failed gate is re-run once to detect flakiness.

### Forks for its design act

- **Where the plan's `check` comes from:** the planner writes it from the brief
  and the repository; whether a config key gives a default the planner starts
  from (and whether a brief may override it).
- **What the iteration records of it:** exit code and duration as for a gate,
  and whether its output is kept (as gate logs are) for the reviewer.
- **The gate review's outcome:** back to the planner automatically (how many
  rounds?), or always to the operator. Is it a new session kind (`gate-review`),
  with its own model and effort keys, verdict schema and fence?
- **Judge fixtures:** a directory the planner writes, or fixtures written inline
  in the verify and nowhere else; how they are committed with the plan.
- **Flaky gates:** whether the re-run that doubles a timed-out gate's cost stays.
- **A plan the gate review rejects twice:** exit code and resumability (R-3).
- **Migration:** a repository with a v1 plan in flight, and a v1 brief whose
  gates would fail the review.

### Binding references its brief must cite

- `docs/domain/domain-model.md` — P-3, P-4, P-5, S-2, S-3, S-4, R-1, R-3, R-4, M-4: the task's fields, gates before work, the plan's shell, the boundary, the review's independence, the exit codes, gate failures blamed on the plan
- `docs/domain/execution/task.md` — the task's fields (`files` goes) and "The gate", which this brief redefines
- `docs/domain/execution/plan.md` — how a plan is accepted, which gains the gate review and judge fixtures
- `docs/domain/execution/session.md` — the session kinds and their fences, which gain the gate review
- `docs/domain/execution/run.md` — the iteration and its record, which gain the check command
- `docs/domain/measurement/metrics.md` — how gate failures and their origin are counted (M-4), now beside check failures
- `schemas/state.v1.json` — the plan this brief versions to `state/v2`
- `schemas/iteration.v1.json` — the iteration record, which gains the check command's result
- `schemas/verdict.v1.json` — the shape a gate review's verdict starts from
- `plugin/skills/plan/SKILL.md` — "The one rule…", "What the gate cannot carry", "What the gate does not outlive", "The gates, together, are the regression net", "Check every gate on the base — three ways": the guidance that changes
- `plugin/skills/review/SKILL.md` — the reviewer, who stops re-running tests and reads the check command's result
- `plugin/skills/work/SKILL.md` — the work session, whose `gate_dispute` stays the way a contract change reaches the operator
- `docs/guide/concepts.md` — "Gates and the gate shell": the operator-facing account of gates
- `docs/briefs/B20261002-2135-vloop-v1-security.loop-brief.md` — F9 and the operator notes: what the gate-file rule got wrong, and why

Code and evidence it touches (for the brief's constraints and "tests that must
change", not as binding references): `internal/driver/gates.go`,
`internal/driver/gateshape.go`, `internal/driver/plan.go` (plan acceptance),
`internal/driver/iterate.go`, `internal/state/`, `fence/plan.json`,
`fence/work.json`, `fence/review.json`, `plugin/evals/plan-checks-its-gates/`,
the interventions `I20261003-1147` and `I20261003-1217`, and the defects
`D20261003-1147` (F9) and `D20261003-1217` (T12's and T17's gates).

### Out of scope

The interventions (B10) and the quality pass (B11); the configurable merge
strategy (horizon); containers or sandboxes for gates.

---

## B10 — interventions that weigh the model against the operator

### What it owns

Promoted from the horizon (noted 2026-10-01), with one change set by the
operator on 2026-10-03: **the
operator skill proposes three options, not one, and says which it recommends and
why.**

- **The model's options, the recommendation and its rationale, beside the
  operator's decision.** Each record gains:
  - `options` — three options the model proposed at the time, a sentence each,
    or `none` when it proposed nothing;
  - `recommended` — which of the three, with its rationale in a sentence or
    two;
  - `decided` — what the operator chose: one of the options, a variation of
    one, or something else, in a sentence.
- **A derived `agreement`** per record: `recommended` (the operator chose the
  recommended option), `other-option` (one of the other two), `adjusted` (a
  variation of an option), `different` (none of them), or `no-options`.
  `vloop metrics` reports it by kind and phase, and across a workspace. Read
  together with `automatable`, it is the evidence for which interventions a
  driver could take over: the kinds where the operator already takes the
  model's recommendation.
- **Context for review after the fact**, in one succinct paragraph: the run,
  iteration and task it happened in; the commit or files involved; the
  triggering message or output, quoted briefly; and what changed because of it
  (a defect recorded, a gate amended, a brief rescoped). `vloop intervention
  show <id>` prints the record with those links resolved.
- **Written at the time, honestly.** Options are only what the model actually
  proposed, before the operator decided; never reconstructed afterwards. The
  operator skill (`/vloop:operate`, and this repository's
  `.claude/skills/vloop-operator`) proposes three options and a recommendation
  whenever it brings the operator a decision, and records them.

### Evidence

- The 61 records up to 2026-10-01 were given context, a suggestion and a decision
  by hand in their bodies (commit `e6fc579`), not as fields; later records use
  the CLI's Trigger, Done and What would automate it, which have none of them.
- The design act and the operator sessions of B7 and B8 already brought most
  decisions as several options with a recommendation (forks with a recommended
  option; "options for these two cases") — the practice this makes a record.

### Forks for its design act

- **Fields or sections:** frontmatter fields validated by `intervention/v2`, or
  body sections the CLI writes and parses.
- **How `adjusted` and `different` are decided:** by the operator when
  recording, by the model, or derived from text.
- **The 61 back-filled records:** migrated to `intervention/v2` as
  `no-options`, or left as v1 and reported apart.
- **Exactly three options, always**, or "up to three" when fewer real
  alternatives exist (never padding with a straw option).
- **Workspace analysis:** which cut of `agreement` the metrics show first.

### Binding references its brief must cite

- `docs/domain/domain-model.md` — the intervention's vocabulary row and id, the "On disk" layout, and M-1 and M-7: metrics recomputed from raw records, and what an export may carry
- `docs/domain/measurement/measurement-context.md` — interventions as measurement's next entity, which this brief makes them
- `schemas/intervention.v1.json` — the record this brief versions to `intervention/v2`
- `schemas/metrics.v1.json` — the metrics document that gains the agreement report
- `schemas/export.v1.json` — the export that carries interventions, which must stay numbers and titles (M-7)
- `docs/guide/defects.md` — the guide that documents interventions beside defects
- `plugin/skills/operate/SKILL.md` — the shipped operator playbook, which gains the three-options rule and how to record it
- `.claude/skills/vloop-operator/SKILL.md` — this repository's operator playbook, step 6 (recording interventions), which follows it
- `docs/design-notes/vloop-interventions-b1-b4.md` — what the first 38 records showed, and why a driver would want these fields

Code and evidence it touches: `internal/intervention/`,
`internal/cli/intervention.go`, `internal/metrics/`, `internal/cli/export.go`,
`tools/interventions-index.sh`, `.vloop/interventions/README-interventions.md`,
and the records back-filled in `e6fc579`.

### Out of scope

The gate model (B9); the quality pass (B11); a project-level driver that acts on
the agreement data (the horizon's H1).

---

## B11 — the quality pass

### What it owns

The v1.0 review's quality findings, unchanged in substance, in full in
`docs/design-notes/vloop-v1-review.md` → "B9 — quality" (written when this pass
was numbered B9): the guides against the binary; error messages and exit codes
(exit 2 for usage only); help text; code health (one frontmatter parser, one git
wrapper, long functions, metrics correctness, cross-platform paths, `gendocs`
and the tracked `--help` file); the test suite's speed (`t.Parallel()`, the
real-repository test). Two additions:

- **`vloop doctor`'s plugin check** reports the plugin as supplied when vloop
  passes it itself (`vloop run`'s `--plugin-dir`), and suggests
  `claude --plugin-dir "$(vloop plugin path)"` for interactive sessions, instead
  of a marketplace install that does not exist while the repository is private.
- **The README, last**, rewritten for a newcomer under the 200-line cap
  (prerequisites, install, quickstart, the loop with v2 gates, guides), now that
  B8 moved the completeness checks to the guides.

### Forks for its design act

- Whether code health (refactors with no behaviour change) shares a run with the
  user-facing fixes, or splits off; the review counted about 25 findings.
- The exit-code mapping: which of today's exit-2 errors become exit 1, and how
  the change is announced (it changes scripts' behaviour).
- The README's outline (the review proposes one of about 150 lines).

### Binding references its brief must cite

- `docs/design-notes/vloop-v1-review.md` — "B9 — quality": every finding this pass owns, with its evidence and location
- `docs/domain/platform/platform-context.md` — the CLI conventions every message and exit code must follow
- `docs/domain/domain-model.md` — R-3 and C-1 to C-4: the exit codes and the configuration and language rules the guides document
- `docs/guide/commands.md` — the generated reference the help text and README must agree with
- `docs/guide/concepts.md` — the concepts guide, whose exit-code table and gates section must match v2
- `docs/guide/configuration.md` — every config key, now the place a key must be documented (B8's C3)
- `README.md` — the newcomer's page this pass rewrites, last

Code it touches: `internal/cli/` (`root.go` exit mapping, every `Short` and
`Long`, `doctor.go`), `cmd/gendocs/`, the five frontmatter parsers and git
wrappers the review lists, `internal/metrics/`, `cmd/vloop/` (test speed).

### Out of scope

New behaviour beyond the findings and the two additions; anything B9 or B10
changes, which this pass only documents.

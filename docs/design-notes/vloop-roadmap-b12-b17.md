---
name: vloop-roadmap-b12-b17
description: The briefs the T1 triage set after v0.8.0 — field reliability, docs and onboarding, the learning loop, code health, a verification phase, a release command — what each owns and depends on. Read before writing any of them.
kind: design-note
status: draft
created: 2026-10-08
---
# vloop — the roadmap after the T1 triage (B12–B17)

v0.8.0 (tag `v0.8.0`) is B1–B11: `docs/design-notes/vloop-roadmap.md` (B1–B8)
and `docs/design-notes/vloop-v2.0-roadmap.md` (B9–B11, planned as v2.0). This
series is the T1 triage's: `docs/design-notes/vloop-t1-triage.md` → "Proposals"
holds each brief's design input and evidence, and "Decisions" the order and
shape the operator settled on 2026-10-08.

Every brief here is run by `vloop run` from the latest released vloop, never
from a binary built from the tree it changes. A release may be cut after any
brief, as the operator chooses. Each brief names its predecessor in
`depends-on:`.

| # | Brief | Owns | Depends on |
| --- | --- | --- | --- |
| B12 | `docs/briefs/B20261008-2144-b12-field-reliability.loop-brief.md` — **consumed** | **Field reliability** (T1 P1): the `.git/config` guard ignores editor keys; an exit 9 during plan acceptance keeps the plan and its run folder; session wall-clock time; denied Bash commands recorded, masked (with the dashed `-Users-<name>-` form); running cost, ceiling warnings and `vloop status` during a run; the installed version identifies itself; metrics say "not measured" | T1 |
| B13 | `docs/briefs/B20261011-2100-b13-docs-onboarding.loop-brief.md` — **consumed** | **Docs and onboarding** (T1 P2): the CLI is driven by asking `/vloop:operate`, and how to get it; the driver defined; the name explained; the install's `PATH`; gate-folder lifecycle; retiring the shell loop; `doctor` checks that the operator's session may edit `.vloop/state/gates/` (`Edit(./.vloop/state/gates/**)` in `.claude/settings.json`), so a gate repair does not stall on a permission prompt (B12's I20261011-1027) | B12 |
| B14 | `docs/briefs/B20261011-2251-b14-learning-loop.loop-brief.md` — **consumed** | **The learning loop** (T1 P3–P5): `.vloop/repo-guidelines.md` handed to every session by the driver and read by `vloop brief check`; every gate-review miss an eval case, with the open eval defects; `vloop brief check` rules for the repeating spec gaps | B13 |
| B15 | not written | **Code health, and defects narrowed to escapes**: one git wrapper, the duplicated env filters, `planSHA`, `driver.counts`, budget keys, the id generators, long functions, `TestMain`, a Windows run harness — the review's findings B11 left out; B12's `D20261011-1458` (a mixed-case `run.git-config-ignore` pattern matches nothing, unwarned); new defects and interventions get a stamp with seconds, allocated so no two records share one (`D20261004-1142` alone matches ten); the per-task metrics table gains `test:code` and `$/kloc` (cost per 1,000 delivered code lines, the summary's basis; the summary says `$/kloc` too), between `other+` and `agent`, `-` where a task delivered no code, in the text and `--json` output and the docs; and **defects narrowed to escapes** (the operator, 2026-10-14): a recorded defect is one found after merge, by a user or a later brief — the end-user defects worth optimizing; defects found in the loop stay derived from the run's telemetry, as now; what the operator finds while verifying is an intervention only (`verification-finding`, saying whose defect it is: brief, plan, work or env), so `brief close --finding` records interventions, not defects; the triage retires — the `rule` and `rule-outcome` fields, close's defect report and its recurrences (B14's L12), the operate skill's triage step; the metrics `defects` line keeps in-loop, escaped and removal efficiency, the operator's findings counted in the interventions line; existing defect records stay as read-only history, still valid; guidelines keep citing a defect or an intervention as provenance (B14's L9), unresolved — `doctor` resolving every cited ID, set at B14's close, is dropped with the triage. Only the code-health list leaves behaviour unchanged | B14 |
| B16 | not written | **A brief-level verification phase** (T1 P6), the brief's main item, since finding functional errors sooner is the point: `/vloop:verify` after the last task, a findings table the operator reads before close; **pruning rules** (B14 decision): eval runs with and without a rule, on a model change, for the repo guidelines and the skills; **contract-licensed test changes** (set at B14's close): must-change lists name tests, an implementation detail, and their five recorded misses (B8–B10, B12, B13) were bookkeeping or gate-model problems — an earlier test may change only where it pins behaviour the contract changes, and only to the new behaviour, the review citing the item for each changed test; the architect, work and review skills change to match, and L3's pinning-test warning in `brief check` retires (the repository's `## Pinning tests` section was removed at B14's close); **the guidelines file holds only lessons a session acts on** (set at B14's close): what a tool enforces is not sent to every session — L4's external-format mapping moves out of `.vloop/repo-guidelines.md` (this repository's line was removed at B14's close), to `config.toml` or to a repository-owned checker, a counterpart of the guidelines file for `brief check` and gates (the repository's own rules as scripts or declarations the tools run); **filter at design time** (the operator, at B14's close): the items above were added as candidates — adopt only the strictly necessary ones, since every rule adds complexity with its own defects, and weigh them by architecture and end-user defects, the ones worth optimizing. Provenance in the defect records and retired rules, set at B14's close, were dropped on 2026-10-14 with the defect triage (B15) | B15 |
| B17 | not written | **`vloop release`** (T1 P7): bump, tag, stamped build, module-path and public-install checks | B16 |

## Naming: the series number in the slug

From B12 on, a brief's slug starts with its series number in lower case:
`vloop brief new b13-docs-onboarding` writes
`docs/briefs/B<YYYYMMDD-HHMM>-b13-docs-onboarding.loop-brief.md`, so the run id,
work branch, journal and PR all show where the brief sits in the series. Set by
the operator on 2026-10-10, who works in several repositories at once.

- The number goes in when the file is created, and only then: until a brief is
  written its number may still move (code health went from B12 to B15); once it
  exists, its name is fixed.
- B1–B11 keep their names; the convention starts at B12.
- It is this repository's convention, not vloop's: the slug is free text.
  Carrying the number in a frontmatter field that `brief list` shows is a
  possible vloop feature, not planned.

## Decisions every brief in this series inherits

- Everything in `docs/design-notes/vloop-v2.0-roadmap.md` → "Decisions every
  brief in this series inherits" still holds, except where a brief changes it on
  purpose and says so.
- The triage's findings are the evidence: a brief cites
  `docs/design-notes/vloop-t1-triage.md` for the proposal it implements rather
  than restating it.

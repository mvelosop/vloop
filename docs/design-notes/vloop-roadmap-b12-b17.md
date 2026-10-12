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
| B14 | `docs/briefs/B20261011-2251-b14-learning-loop.loop-brief.md` — **ready** | **The learning loop** (T1 P3–P5): `.vloop/repo-guidelines.md` handed to every session by the driver and read by `vloop brief check`; every gate-review miss an eval case, with the open eval defects; `vloop brief check` rules for the repeating spec gaps | B13 |
| B15 | not written | **Code health, no behaviour change**: one git wrapper, the duplicated env filters, `planSHA`, `driver.counts`, budget keys, the id generators, long functions, `TestMain`, a Windows run harness — the review's findings B11 left out; B12's `D20261011-1458` (a mixed-case `run.git-config-ignore` pattern matches nothing, unwarned); and, set at B14's close as two behaviour changes: `doctor` resolves every ID `.vloop/repo-guidelines.md` cites and warns on one that matches no record or more than one (`D20261004-1142` alone matches ten), and new defects and interventions get a stamp with seconds, allocated so no two records share one | B14 |
| B16 | not written | **A brief-level verification phase** (T1 P6): `/vloop:verify` after the last task, a findings table the operator reads before close; **pruning rules** (B14 decision): eval runs with and without a rule, on a model change, for the repo guidelines and the skills | B15 |
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

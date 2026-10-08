---
name: vloop-t1-triage
description: The T1 triage — what B1–B11's 73 defects and 90 interventions, and 0.8.0's first field use, say about the skills, the driver and repo-specific guidelines; the proposals that seed the next briefs, and the forks the operator must settle.
kind: design-note
status: draft
created: 2026-10-07
---
# T1 — triage of defects and interventions

The inputs are `docs/design-notes/vloop-t1-triage-inputs.md`: the open defects,
findings 1–10 and field notes 1–9. This note reads the records as a whole
(`vloop defect list`, `vloop intervention list`, `vloop metrics --interventions`)
and proposes. Nothing here is decided until the operator settles the forks at
the end.

## What the defects say

73 defects, B1–B11 (5 open). By where the fault was made (`origin`) and who
caught it (`found-by`):

```
origin   operator  user  gate  review  total
brief          24     1     4       1     30
work           28     2     2       0     32
plan            4     0     7       0     11
total          56     3    13       1     73
```

1. **The per-task review caught one defect in eleven briefs** (`D20261005-2129`,
   a brief that asked for two incompatible things). Gates caught 13. The operator
   caught 56, most of them after the loop had said done. The review is not
   failing at its job as written: its job is "what a command cannot check" *in one
   task's diff*. The escapes are not in one task's diff.
2. **What escaped the work (28, operator-caught) falls in three groups:**
   - **Statements that disagree across surfaces** (about 12): help text,
     `docs/guide/commands.md`, the domain docs, the README and error messages
     stating a behaviour differently from the code or the brief (B11's T6, T7,
     T12, T13 and T16 findings; B1's README shell default; B6's `plan.md`
     wording). Each task changed one surface; nobody held them side by side.
   - **The security pass** (6 work, 9 brief): B8's F-findings, found by a
     dedicated five-pass review the operator ran (`I20261002-2140`), not by the
     loop.
   - **External formats written from memory** (3, B7): eval cases and graders
     the eval tool rejected.
3. **What the operator found was found by an after-run verification pass**: an
   independent session running every worked-example line on a fresh build,
   checking the brief's sentences, and tabulating findings
   (`I20261005-0757`, `I20261006-1028`). Its "what would automate it" is the
   same both times: *the findings table is mechanical; deciding what to fix
   before merge is the operator's*.
4. **Brief spec gaps (30) repeat in a few shapes:**
   - a **must-change list missing tests that pin counts and lists** (schema
     list, embedded-schema count, `init`'s output) — three times
     (`D20261004-1142-the-brief-s-must-change…`, `D20261004-2304-the-brief-s-must-change…`,
     finding 6);
   - a brief that **asks a session to write under `.vloop/`**, which S-4 forbids
     (`D20261004-2304-the-brief-asked-sessions…`);
   - an **external format pinned from memory** (`D20261001-1338`);
   - **half a requirement**: tolerance without the detection it must keep
     (`D20261005-2129`);
   - B8's security spec gaps, which a design act alone did not find.
5. **Gate defects are caught by gates and gate reviews, mostly** (7 of 11), but
   the gate review **missed twice with its rule already written**: "every
   matcher is exercised once against the exact output the brief pins" is in
   `plugin/skills/gate-review/SKILL.md` since B9, and B11's T12 matcher still
   could not match a message-less pass; B10's T9/T10 needed a session to write
   under `.vloop/`, a rule the skill does not state. A rule in a skill is not
   a check.

## What the interventions say

90 interventions, B1–B11 and the release. By phase, with how many could be
automated (`yes` / `partly` / `no`):

```
phase    total  yes  partly  no
design      29    6      11  12
verify      21    5      15   1
next        15    1       5   9
close        9    7       2   0
run          8    2       4   2
halt         7    2       5   0
setup        1    1       0   0
total       90   24      42  24
```

1. **Design is the largest and the least automatable**, as B1–B4's note found:
   direction and decisions are the operator's. What can be automated there is
   already a check (`vloop brief check`) or a survey.
2. **Verify is the largest automatable block** (20 of 21 yes or partly), and
   it is the same pass as defect point 3: a verification session whose table
   is mechanical.
3. **Halts are repairs of gates** (5 of 7), and every "what would automate it"
   names the gate review or the planner running the gate against the brief's
   pinned output — the same misses as defect point 5.
4. **Ceremony is nearly solved** (close: 7 of 9 yes). What recurs is the
   release (`I20261004-1337`, `I20261005-0906`, `I20261006-1503`,
   `I20261006-1748`): four records, one command (finding 8).
5. **Agreement since B10 is high but thin**: of 11 records with options, the
   operator took the recommendation 9 times (82%), chose another option once and
   adjusted once. Too few to change how far the operator skill may go alone.

## What the field says (0.8.0 in another repo)

One brief in visum-monorepo, 9 tasks, 9 first-pass, and five operator defects —
all **about vloop, not about the work**: the exit-9 guard again, a lost plan's
cost, session time 20 s for a 39-minute plan, Bash denials without their
commands, zero delivered size for a docs brief. The loop did the work; the
**instruments around it** were wrong. A user outside this repository has no
way to tell a wrong number from a right one.

## Proposals

Numbered in the proposed order of work (see "Order, and B12").

### P1 — Field reliability: instruments you can trust

Everything the field found, with the two open driver defects, as one brief:
- the `.git/config` guard ignores editor bookkeeping keys
  (`branch.*.vscode-merge-base`, and the like) and keeps guarding hooks,
  fsmonitor, filters and `core.*` (`D20261005-1125-the-driver-s-git-config…`,
  `D20261007-1243`);
- an exit 9 during plan acceptance **keeps** the plan and its run folder: the
  re-run resumes acceptance (`D20261005-1125-an-exit-9…`; $8.23 and $5.51 lost);
- session time is the driver's own wall clock (finding 9);
- a denied Bash call keeps its command, masked and truncated (finding 10), with
  the dashed `-Users-<name>-` mask (finding 3);
- the running cost and elapsed time per session in the log, warnings at 50% and
  80% of the ceiling, and `vloop status` showing the run (field note 4, the
  carried options);
- the installed version identifies itself (finding 2);
- metrics say "not measured" instead of 0 beside a percentage (field note 9).

### P2 — Docs and onboarding: the CLI is driven by asking

One docs brief: the skill is the interface (field note 5) and how to get it
(finding 1: `doctor` warns when no interactive plugin is enabled, `init`'s next
steps and the `CLAUDE.md` section say how); the driver defined (field note 6);
the name explained (field note 2); the install's `PATH` (field note 7); the gate
folders' lifecycle (field note 8); retiring the shell loop (field note 1, with
the "this repository only" correction); the two rule sets side by side (finding
4); not combining an installed plugin with `--plugin-dir` (finding 5, after the
cheap confirmation). P3's repo-guidelines file is documented here if it lands
first.

### P3 — Repo guidelines: a place where lessons become inputs

Many fixes so far went into a brief, a skill, or the operator's memory. A
lesson specific to *this* repository ("a new schema changes `vloop schema list`,
`embed_test.go`'s count and `TestWorkedExampleB6RealData`") has no home: it is
not vloop's rule, so it does not belong in the plugin's skills, and a brief
only carries it if the architect remembers.

**Proposal**: `.vloop/repo-guidelines.md`, a short file the repository owns,
which the plan, gate-review, work and review skills read when present (like
`references`, but standing). `vloop defect` gains a way to point a fixed defect
at the guideline that prevents it (`fixed-by: guideline`). `vloop init` writes
an empty one; this repository seeds it from the defects above (the pinned-count
tests, S-4's `.vloop/` rule for briefs and gates, external formats pinned from
the tool, the shell loop's rule 1 on `~/.claude`). This is the "lessons store"
of `I20260929-2325`, and it is the triage's answer to "repo-specific guidelines
that could automate interventions".

### P4 — Turn every gate-review miss into an eval case

The gate review missed twice with its rule already written. **Proposal**: the
real misses become `gate-review` eval cases, scaffolded from the plans that
fooled it (B10 T9's `.vloop/` write, B11 T12's message-less pass, B2 T1's
GNU-only grep, B5 T8's TOML key after a table). Add S-4's `.vloop/` rule to the
skill's `unpassable` kind. From then on, a gate defect recorded with
`found-by: operator` or a halt carries a "case" link (`case:`, already in the
defect schema) or a reason it has none. The same for the review skill: the
statements-disagree escapes become review eval cases, to measure P6 and the
review against. The three open eval defects (`D20261004-1309`,
`D20261005-0757`, `D20261006-1028`) belong here: the fixture grader and the
operate scaffold are the same eval infrastructure.

### P5 — `vloop brief check` learns the repeating spec gaps

Deterministic, cheap, in the design act where most spec gaps are made:
- **error**: the brief asks a session (work or gate) to write under `.vloop/`
  outside `.vloop/tmp/`;
- **warning**: the brief adds or removes something enumerated (a schema, a
  command, a config key, an `init` file) and its must-change list names no test
  — with the repo-guidelines file (P3) supplying this repository's list of
  pinning tests;
- **warning**: the brief names an external tool's format (a manifest, an eval
  case, a schema) with no reference to the tool's own definition.

### P6 — A brief-level verification phase

The operator's after-run verification found what the loop could not, and its
table is mechanical. **Proposal**: an optional final phase, `verify`, after the
last task: one independent session, with the whole diff against the base,
the brief, and a skill (`/vloop:verify`) whose checks are those passes:
every worked-example line on a fresh build; every sentence of the brief that
states an observable behaviour; **every surface that states a changed
behaviour** (help, guides, domain docs, README, messages) held side by side.
Its output is a findings table in the run folder; the driver does not act on it
and the brief does not close without the operator reading it. A cost line in
the config (`run.verify = true|false`) keeps it optional.

The per-task review stays as it is. It costs a session per task and catches
little; whether to cut it (review only tasks whose gate is weak, or a cheaper
model) is a fork, not a proposal, until P6 shows what it catches.

### P7 — Release command

`vloop release` (finding 8): bump, tag, stamped build, the module-path major
check (`I20261006-1503`) and the public install check (`I20261006-1748`). The
merge stays the operator's.

### Housekeeping, no brief needed

- Close `D20261006-1102` (moot since 0.8.0).
- Rename the roadmap's "v2.0" rows and file to the 0.x numbering.
- Finding 6's operator lessons (watch the driver by pid; close the editor until
  P1) go into `.claude/skills/vloop-operator/SKILL.md` now.

## Order, and B12

The roadmap's B12 (code health) is unchanged in content; the triage only asks
where it goes. The proposed order, by what the field pays for today:

1. **P1 field reliability** — the only proposals a user outside this repository
   meets every run; two of them cost money on each recurrence.
2. **P2 docs and onboarding** — cheap, no behaviour change, and what a public
   repository needs first.
3. **P3 + P4 + P5 the learning loop** — repo guidelines, eval cases from
   misses, brief-check rules; one brief, since each feeds the others.
4. **B12 code health.**
5. **P6 verification phase** — after P4 gives the evals to measure it with.
6. **P7 release** — whenever the release ceremony next hurts.

## Decisions

Settled by the operator on 2026-10-08:

1. **P3's shape**: `.vloop/repo-guidelines.md`, a file the skills read; the
   name says the repository owns it.
2. **P6**: a verification phase in the loop, not the operator's step.
3. **The per-task review**: kept as is until P6 measures what it catches.
4. **Order**: P1 first, as above; the roadmap's B12 (code health) moves after
   the learning loop.
5. **P1's size**: one brief, not a patch release for the exit-9 defects alone.

The briefs: B12 is P1, B13 P2, B14 P3–P5, B15 code health, B16 P6, B17 P7
(`docs/design-notes/vloop-v2.0-roadmap.md`).

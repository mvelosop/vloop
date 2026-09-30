---
name: vloop-interventions-b1-b4
description: What the operator (and the assistant as the operator's hands) had to do between and around the loop's runs for B1-B4, besides testing — 38 interventions by kind, phase and whether a driver could do them — and what that says about a project-level driver (horizon H1)
kind: design-note
status: draft
created: 2026-09-30
---
# Operator interventions, B1–B4

The records are data, one file each in `.vloop/interventions/` (the shape a
future `vloop intervention add` would write). This note is what they say.
All 38 were backfilled on 2026-09-30 from the session that ran B1–B4, so the
classification is a first reading, not a measurement; `by` is who acted —
the operator, the assistant as the operator's hands, or both.

## The kinds

| Kind | Meaning |
| --- | --- |
| direction | the operator changed what the product or the plan is |
| decision | a fork settled — in a design act, or accepting a session's reading |
| context-supply | something the loop needed that only a human could provide |
| halt | the driver stopped and someone had to act |
| verification-finding | something wrong found after the loop said done |
| repair | undoing damage or a wrong state, including the loop's own |
| carry-forward | moving what was learned into the next brief, the docs or the tooling |
| ceremony | branches, status, run records, PRs, merges |

## The numbers

```
kind                  yes  partly  no   total
direction               0       0   6       6
decision                0       5   1       6
context-supply          0       2   0       2
halt                    1       1   0       2
verification-finding    4       3   0       7
repair                  3       1   0       4
carry-forward           1       3   1       5
ceremony                6       0   0       6
total                  15      15   8      38
```

By phase: design 10, verify 10, close 7, next 6, run 4, setup 1. By brief:
B1 10, B2 8, B3 7, B4 7, series-level 6.

## What it says

- **Ceremony is solved or about to be.** All six are automatable, B4's
  `close` already took the run record and status, and B6's `run` takes the
  branch. What remains is the merge — an outward-facing step the operator
  wants to authorise, so it stays a gate, but a one-word one.
- **The run itself barely needed anyone.** Four interventions in four runs, two
  of them one mistake (planning before branching). The loop is not where the
  operator's time goes.
- **Verification is where the second-largest share goes, and it is mostly
  automatable** — 7 of 10 verify-phase items. The findings came from running
  the product on **real data** (B3's estimate, found by `vloop metrics` on B3
  itself), **reading the log** (denials, a missing record) and **running one
  more check** (tidy). None came from replaying the worked example by hand,
  which the close task's gate already does. Each is a gate or a log check that
  could run before the operator looks.
- **Repairs were the loop damaging itself** — B3's planning session moving refs,
  a stale fetch, a rehearsal clone's `origin/HEAD`. Detection is automated now
  (exit 9); the repair stays human, and should be rare.
- **Decisions and direction are the operator's, and they cluster in the design
  act**: 12 of 38, none automatable, all but one before a run. Every brief took
  a round of forks (4–11) and every review of a draft moved something. Three of
  the six defects recorded for B1–B3 were spec gaps in the briefs themselves.
- **Carrying forward is the gap a project driver would have to fill**: open
  items into the next brief, lessons into constraints, the operator's role into
  a skill. Today it depends on the operator (or the assistant) remembering.

## For H1

A project-level driver is plausible if it stops for the operator at two points
and runs the rest:

1. **The design act** — direction and decisions. It can prepare the survey, the
   carried-forward items and the draft; the forks are the operator's.
2. **The merge** — after automated verification (real-data gates, log checks)
   has run and its findings are in front of the operator.

Between the two, the evidence says the loop needs no one, provided the
carry-forward has a home it reads from: a domain model and use cases that pin
rules once rather than per brief, architecture decisions, and the lessons. That
is the documentation layer — the next step, and the first thing B5's design act
should use.

## Keeping the data

From B5 on, each intervention is recorded when it happens, as the operator
skill's closing step says, not backfilled.

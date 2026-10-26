---
name: B20261025-2302-b20-gates-checks-review-scope.loop-brief
description: Make "a gate never runs the suite" a domain invariant every session obeys, have the gate review check that every acceptance criterion is judged by a gate, a check or the review — halting for the operator on a missing check — and keep vloop's own records out of the review's product constraints
kind: brief
status: consumed
created: 2026-10-25
seeds: The first of two hardening briefs before the fence (docs/design-notes/vloop-roadmap-b18-b21.md → "Before the fence: hardening"), from B19's acceptance run on the URL-shortener reference brief, settled with the operator on 2026-10-25
depends-on: [B20261025-1307-b19-copilot-review.loop-brief]
---
# Brief — vloop B20: gates, checks and the review's scope

- **Starting point:** extends `main` at `303884f` (B19 merged). The planner
  pins the base it plans from as the **base** every gate compares against —
  never `HEAD~n`.
- **Produced by:** the design act of 2026-10-25, after B19's acceptance run
  (`docs/design-notes/copilot-cli-experiments.md` → "Acceptance on the real
  CLI") and B19's close findings.

## Executive summary

In B19's acceptance run, the URL-shortener's close task blocked twice for
reasons no work session could fix: its criterion asked for the test suites in a
fresh copy, which no gate may run and no check covered, and the review ruled
"no absolute paths in any committed file" against vloop's own records. Neither
was caught before work started.

**The goal:** a plan whose criteria nobody can judge stops before any work, and
says what is missing — a check the operator must add, or a criterion the
planner must reword — and a review never fails a task for something only vloop
writes.

- **J1 — a gate never runs the suite, as a domain invariant.** One rule, P-8,
  cited by the plan, gate-review, review and operate skills, so no session —
  and no operator repair — puts the repository's tests into a gate.
- **J2 — the gate review checks coverage.** Every acceptance criterion is
  judged by its gate, by a check the plan holds, or by the review; one that is
  not is a finding, and one no work session could meet is `unpassable`.
- **J3 — a missing check halts for the operator.** The planner cannot add a
  check, so a `check-gap` ends the run before any work, naming what to add.
- **J4 — the review's scope.** vloop's records under `.vloop/` are the
  driver's: an absolute path or a stale text there is a vloop defect in `notes`,
  never a finding against the task; a gate that leaves the suite to the checks
  is not a gate gap.
- **J5 — proven by evals.** Two eval cases plant the B19 misses; the operator
  runs them, and the existing gate-review and review calibrations, at close.

Unchanged: the driver's loop, attempts and exit codes other than J3's halt,
every other finding kind, the checks themselves and when they run.

## Why this shape, and what was rejected

Decided with the operator on 2026-10-25. **Do not re-litigate these** — the
planner inherits them.

- **Gates are contract checks; tests are the checks'.** A gate judges what the
  brief promises from outside; the repository's tests are internal and run as
  `[[check]]` entries (P-7). The plan skill already says so; it binds nobody
  else, and in B19's acceptance the reviewer called a gate without the suite a
  gap and the operator's repair put the suite into it.
  *Rejected:* letting a close task's gate run the suite in a clean copy (a
  second, slower regression net that judges the work by the work's own tests).
- **A missing check goes to the operator at once.** Only the operator edits
  `.vloop/config.toml`; sending a `check-gap` back to the planner spends a
  round on what it cannot fix.
  *Rejected:* treating it as any other finding (a second FAIL, then the same
  halt, a planning round later).
- **A criterion nobody can judge goes back to the planner** (`uncovered`), who
  rewords it to what the gate, a check or the review can judge.
- **No operator override for a blocked task.** With J2 and J4 a task blocks on
  a real problem; marking one done by hand is not added.
  *Rejected:* `vloop task accept` (a shortcut that hides the cause).
- **The review's scope is the product.** Everything under `.vloop/` is
  written by the driver or the planner (S-4); a work session cannot change it,
  so no criterion can be met or failed by it.

## Binding references

- `docs/domain/domain-model.md` — P-4 and P-7, beside which J1's P-8 goes; S-4, what a work session may write, which J2's `unpassable` criterion rests on; R-3, the exit codes J3 keeps
- `docs/domain/execution/task.md` — "The gate": the judge, gate disputes and `task verify`, which P-8 constrains
- `docs/domain/execution/plan.md` — the plan's `checks`, copied from the config, which J2's coverage reads
- `plugin/skills/plan/SKILL.md` — the gate rule P-8 replaces with a citation, the checks section and "Before you finish", where the planner's side of J2 goes
- `plugin/skills/gate-review/SKILL.md` — the evidence it reads, the three checks and the finding kinds J2 extends; its verdict section
- `plugin/skills/review/SKILL.md` — the questions and finding kinds J4 scopes, and the absolute-path question that read vloop's records
- `plugin/skills/operate/SKILL.md` — the halt table and the gate-dispute step J1 and J3 change
- `.claude/skills/vloop-operator/SKILL.md` — this repository's playbook: the same halt table and gate repair
- `internal/driver/gatereview.go` — reads the gate verdict and turns findings into problems; where J3's halt is decided
- `internal/driver/plan.go` — the acceptance rounds, and what a plan that is not fit leaves behind
- `schemas/gate-verdict.v1.json` — the verdict whose kinds J2 extends, as `gate-verdict/v2`
- `internal/budget/budget.go` — the skills' word ceilings
- `plugin/evals/README.md` — the eval layout and the case tables J5's cases join
- `docs/guide/concepts.md` — "The gate review" and the exit codes, which say what the operator sees

## Behaviour contract

### J1 — P-8

- `docs/domain/domain-model.md` gains, under "Plans and tasks":
  **P-8** *A gate judges the brief's contract from outside the work: it never
  runs the repository's suite, a check, or a test a task writes. The checks
  run the tests (P-7). Every acceptance criterion is judged by its task's gate,
  by a check the plan holds, or by the review.*
- The plan, gate-review, review and operate skills, and this repository's
  operator playbook, cite P-8 by id where they speak of a gate's scope; the plan
  skill's own statement of the rule becomes that citation, not a second copy.
- The operate skill and the playbook say: a gate repair keeps the gate's
  intent and never adds the suite, a check or a task's test to it (P-8); a
  criterion that needs a test run is a check the operator adds to
  `.vloop/config.toml`.

### J2 — coverage, in the gate review

- The gate review reads the plan's `checks` and, for every task, rules on
  every acceptance criterion: judged by the gate, by a named check, or by the
  review (from the diff, or from a check's log).
- Two new finding kinds:
  - `uncovered` — no gate, check or review could judge the criterion as
    written (it names an outcome no command in the plan produces and the diff
    cannot show);
  - `check-gap` — only a check could judge it, and no check the plan holds runs
    what it names (a script, a suite, a clean install).
- `unpassable` also covers a criterion no work session can meet: one that
  requires a change under `.vloop/` outside `.vloop/tmp/` (S-4), or rules on
  files only the driver or the planner writes.
- The gate-review skill names, among its findings, a gate that runs the suite,
  a check or a task's test, as `task-reach` (P-8) — as it does today for the
  suite.
- The plan skill's "Before you finish" gains: every acceptance criterion is
  one the gate, a check the plan holds, or the review can judge; a criterion
  only a check can judge names that check.
- The verdict is `gate-verdict/v2`: v1's shape, with the kinds `uncovered` and
  `check-gap` added. `gate-verdict/v1` stays embedded, unchanged, so verdicts
  already recorded still validate; the driver reads v2.

### J3 — a check gap halts

- When a gate verdict has any `check-gap` finding, in any round, the driver
  prints every finding of the verdict, as it does today, then halts with exit
  **2** before any work session and without sending the plan back:

  ```
  vloop: the gate review found 1 check gap(s): the plan needs a check .vloop/config.toml does not have — add it, commit it, then re-run (see <run folder>reports/gate-review-<round>.json)
  ```

  (`<run folder>` is the run folder relative to the repo root, as in the
  existing "failed the plan twice" message; the count is the number of
  `check-gap` findings.)
- Nothing is committed and the plan is not kept: the halt leaves the
  repository as a plan that is not fit does today, so the re-run plans afresh
  with the new check in its `checks`.
- A verdict with no `check-gap` behaves exactly as today: `uncovered` and every
  other kind go back to the planner, and a second FAIL halts as today.

### J4 — the review's scope

The review skill says:

- The product is every file outside `.vloop/`. vloop's records under
  `.vloop/` — state, plan, notes, journals, run reports, gate folders — are
  written by the driver and the planner, and a work session cannot change them
  (S-4). A criterion, a constraint or a question that speaks of "any file" or
  "every committed file" is judged on the product.
- An absolute path, or any other defect, found only in vloop's records is a
  vloop defect: one line in `notes`, never a finding and never a reason for
  `met: false`.
- A gate that leaves the suite to the checks keeps P-8 and is not a
  `gate-gap`. A criterion a check judges is judged from that check's log; a
  criterion nothing could judge is a `spec-gap`, with a line in `notes` naming
  the gap in the plan.
- The absolute-path question asks about the files the task changed and the
  output it produced, and the skill's home-directory placeholder example is
  replaced by a description that is not itself a path-shaped string.

### J5 — the evals

- plugin/evals/gate-review-coverage/ — a small plan whose gates are all sound
  by the existing three checks, and whose tasks carry: a criterion only a check
  could judge, naming a script no check runs (expect `check-gap`); a criterion
  nothing could judge (expect `uncovered`); a criterion requiring a change to
  `.vloop/state/plan.md` (expect `unpassable`); and one task whose every
  criterion is covered (expect `PASS`, no finding). Its scaffold exits non-zero
  if any planted gate fails the existing checks (it must fail on the base at a
  behavioural assertion), so the case measures coverage alone.
- plugin/evals/08-vloop-records-out-of-scope/ — correct work that passes its
  gate; the repository's `.vloop/state/plan.md` and a run report hold an
  absolute home-directory path and the placeholder text of an earlier finding;
  the task's gate leaves the suite to a check that ran and passed; the criteria
  include "no absolute paths in any committed file". Expect `PASS`, no finding,
  and a `notes` line about the records.
- Both join the tables of `plugin/evals/README.md`. The bar: both pass; the
  review calibration keeps its bar (every case caught, `06` allowed to miss);
  `gate-review-planted-gates` and `gate-review-real-misses` still pass.

### Decided here, because the cited documents leave it underdetermined

- A `check-gap` and other findings in the same verdict: the halt wins; the
  other findings are printed and reach the planner on the re-run's own review.
- J3's halt is exit 2 (blocked, R-3), not exit 1: the plan was well-formed and
  the gates held; what is missing is the operator's.
- The word ceilings may rise by what J1, J2 and J4 add, each new ceiling with
  its date and reason in its comment, as `internal/budget/budget.go` requires.
- `docs/guide/concepts.md` says what a `check-gap` halt is and what the
  operator does, in "The gate review" and the exit-code table's row for 2.

### Violations the review must rule on

- The suite, a check or a test a task writes, run by any gate of this plan.
- P-8 restated in a skill instead of cited, or cited by a different id.
- A `check-gap` sent back to the planner, or a halt that commits anything or
  leaves a plan behind.
- `gate-verdict/v1` changed, or a recorded v1 verdict that no longer
  validates.
- A review-skill rule that lets a task's own files out of the absolute-path
  question, rather than only vloop's records.
- An eval case whose scaffold does not refuse a planted gate that fails the
  existing checks.

## Worked example

With the run harness, a stub `claude` whose gate-review session writes a fixed
`gate-verdict/v2`, and a two-task plan:

```
gate verdict: T1 PASS; T2 FAIL [check-gap: "pnpm smoke in a fresh copy: no check runs pnpm smoke"]
vloop run <brief>
  -> gate T2 (check-gap): pnpm smoke in a fresh copy: no check runs pnpm smoke
     vloop: the gate review found 1 check gap(s): the plan needs a check .vloop/config.toml does not have — add it, commit it, then re-run (see .vloop/state/runs/<run id>/<stamp>/reports/gate-review-1.json)
                                                                     exit 2
     no work session ran; git log unchanged; no .vloop/state/state.json

gate verdict: T2 FAIL [uncovered: "..."], round 1; then PASS, round 2
vloop run <brief>
  -> sending 1 problem(s) back to the plan session ... committing the plan   (as today)

gate verdict: T2 FAIL [check-gap: "..."] and T1 FAIL [wrong-reason: "..."], round 1
  -> both findings printed; the check-gap halt, exit 2; no revision round

gate verdict: T2 FAIL [check-gap: "..."], round 2 (round 1 sent back for wrong-reason)
  -> the check-gap halt, exit 2 — not "the gate review failed the plan twice"

vloop schema validate gate-verdict/v2 <a verdict with uncovered and check-gap>   -> exit 0
vloop schema validate gate-verdict/v1 <the same verdict>                         -> exit 1
vloop schema validate gate-verdict/v1 internal/schema/testdata/gate-verdict.valid.json -> exit 0
```

Planted failures the gates must be seen to catch: a halt that commits the plan;
a `check-gap` sent back as a revision round; the round-2 case reporting
"failed the plan twice"; a v2 schema that rejects a v1 verdict's kinds; a skill
that restates P-8 instead of citing it. The evals are the behaviour's proof:
gates never run `claude`.

## Out of scope

- The fence for non-Claude agents and a second reviewer (the fence brief).
- B21's items: the repo root for nested repositories, the metrics' bare brief
  name, #53's temp-dir masking, Copilot's login through `gh`.
- A command that marks a blocked task done.
- Changing what the driver copies from a verdict into a task's `notes`.
- Running checks in a clean copy, or a new kind of check.
- Any change to when checks run, or to `[[check]]`'s fields.
- Re-planning the URL-shortener reference brief, or editing it.

## Constraints

- Go, at the version in `go.mod`; no new dependencies.
- Must keep building and passing on macOS, Linux and Windows.
- **Gates never run `claude` or Copilot**; skill behaviour is proven by J5's
  evals, run by the operator at close with `claude plugin eval plugin`.
- **Gates follow the diff surface.** Each task gates on what it touches; no
  task runs the whole test suite (P-8, this brief's own rule).
- **No task may weaken a check to pass.** An earlier test changes only where it
  pins behaviour the contract changes, and only to the new behaviour; the work
  session names the contract item in its notes and the review cites it.
- **New tooling is proven by fixtures**, including where it must fail: the
  planted failures above.
- Gates run on macOS's BSD tools: no `grep -P`, no `sed -i` without a suffix, no
  `\+` or `\|` in basic regexes, no `date -d`.
- Tests build fixtures in temporary directories and address them with
  `git -C`; sessions never move refs.
- Repo-relative paths everywhere. No absolute paths in any file or commit
  message.

## Shape

5 to 6 tasks, each verifiable by a single command. Order is load-bearing:

1. **P-8 and `gate-verdict/v2`.** The invariant in the domain model; the v2
   schema with v1 unchanged. Gated on the schema lines of the worked example and
   P-8's text.
2. **J3 — the check-gap halt.** Gated on the worked example's run lines with the
   stub gate review, every planted failure included.
3. **J2 — coverage in the gate review and the planner's side**, the
   `gate-review-coverage` eval case, P-8 cited from both skills. Gated on the
   case's scaffold (it refuses a planted unsound gate), the skills' citations of
   P-8, the kinds named, and the word ceilings.
4. **J4 — the review's scope** and the `08-vloop-records-out-of-scope` case.
   Gated on the scaffold, the skill's citation of P-8, and its word ceiling.
5. **J1's playbooks and the guide**: the operate skill, this repository's
   playbook, `docs/guide/concepts.md`. Gated on their citations and the halt
   rows.
6. **Close:** every gate at once and the worked example line for line. The
   operator then runs J5's evals and the existing calibrations, and records the
   results in the run record before merging.

A prose-heavy task that fails review twice on wording, not on a gate, is
blocked; the operator finishes it by hand.
<!-- vloop:run-record:begin -->
## Run record

Generated by vloop brief close on 2026-10-26. Recompute with: vloop metrics B20261025-2302-b20-gates-checks-review-scope.loop-brief

```
B20261025-2302-b20-gates-checks-review-scope  consumed · not merged
 tasks     6 planned (brief said 5–6) · 6 done · 0 blocked · first-pass 5/6
 size      delivered  code 74 · test 504 · docs 61 · test:code 6.81
           churn      code 74 · test 504 · docs 61 · rework 1.00
 time      agent 24.3 min (work 21.7 · review 2.6) · plan 14.7 min · gate review 1.1 min · verify 1.0 min · gates 1.2 · wall 73.0 min
 rate      3.0 code lines/min · 23.8 incl. tests
 cost      $6.88 · plan 3.82 · work 1.76 · review 0.77 · gate review 0.29 · verify 0.24 · $92.91/kloc
 models    plan claude-opus-5-5 · work claude-sonnet-5-5 · review claude-sonnet-5-5 · gate review claude-sonnet-5-5 · verify claude-sonnet-5-5
 defects   in-loop 3 · operator 5 · escaped 0 · removal efficiency 100%
 interventions  8 · recommended 2 · other-option 0 · adjusted 0 · different 1 · no-options 5
 instructions  plan 4,581 · gate review 2,304 · work 2,436 · review 2,457 · verify 1,806 words · guidelines 2 rules
```

```
id  area  kind     att  code+  test+  docs+  other+  test:code  $/kloc  agent   cost   model
T1  -     feature  1    35     15     4      0       0.43       9.21    8m17s   $0.32  claude-sonnet-5-5
T2  -     feature  3    38     170    0      0       4.47       23.33   10m48s  $0.89  claude-sonnet-5-5
T3  -     feature  1    0      205    31     0       -          -       2m16s   $0.59  claude-sonnet-5-5
T4  -     feature  1    1      114    16     0       114.00     343.12  1m19s   $0.34  claude-sonnet-5-5
T5  -     docs     1    0      0      10     0       -          -       0m48s   $0.23  claude-sonnet-5-5
T6  -     chore    1    0      0      0      0       -          -       0m48s   $0.15  claude-sonnet-5-5
```

- B20261025-2302-b20-gates-checks-review-scope/i2-gate — regression, work, found by gate: check failed
- B20261025-2302-b20-gates-checks-review-scope/i3-review-1 — bug, work, found by review: No committed test runs the check-gap halt (exit 2, line, nothing kept, no revision round, no commit, round 1 and 2, alone and with another finding, re-run plans afresh) or shows an uncovered finding is still sent back; the task's last acceptance criterion is unmet
- B20261025-2302-b20-gates-checks-review-scope/i3-review-2 — gate-gap, plan, found by review: The gate (gate.sh) passes without these behaviours being exercised, so it does not enforce the acceptance list
- I20261026-012945-the-run-left-the-domain-s-session-and-pl — verification-finding, work, found by operator: the run left the domain's session and plan pages, the evals guide and one cross-reference in the review skill stale (the verify session's four findings); the operator fixed them on the work branch
- I20261026-012946-in-j5-s-eval-08-the-review-passed-the-wo — verification-finding, work, found by operator: in J5's eval 08 the review passed the work with no finding in 3 of 3 runs but wrote the notes line about vloop's records in 0 of 3: it no longer looks into the records it leaves out of scope
- I20261026-012947-in-j5-s-eval-gate-review-coverage-the-fu — verification-finding, work, found by operator: in J5's eval gate-review-coverage the fully covered task was flagged check-gap in 1 of 3 runs (it doubted the tests check runs a new test file); a false check-gap halts a run before any work
- I20261026-012948-eval-sessions-run-in-don-t-ask-mode-were — verification-finding, env, found by operator: eval sessions run in don't-ask mode were denied the verdict write-and-validate compound command, so verdict-validates failed across cases while the verdicts were right
- I20261026-012949-calibration-04-unchecked-criterion-named — verification-finding, env, found by operator: calibration 04-unchecked-criterion named its defect in 1 of 3 runs with B20's plugin and 2 of 3 with main's; within run-to-run noise, to watch
<!-- vloop:run-record:end -->

### J5's evals, run by the operator at close (2026-10-26)

`claude plugin eval plugin`, sonnet, no baseline arm, the B20 build first on
`PATH`; three runs per case unless said. `verdict-validates` is left out of
the counts (don't-ask mode denied the session's write-and-validate command;
the verdicts themselves were right).

| Case | Result |
| --- | --- |
| `gate-review-coverage` | `check-gap`, `uncovered` and `unpassable` found 3/3; the covered task passed 2/3 (one false `check-gap`) |
| `08-vloop-records-out-of-scope` | `PASS` with no finding 3/3; the `notes` line about the records 0/3 |
| `gate-review-planted-gates` | every planted gate found, the sound gate passed, 3/3 |
| `gate-review-real-misses` | every real miss found as a finding 3/3 (the v1-as-v2 one twice as `unpassable`, once as `pins-unnamed`); the sound T7 `PASS` with no finding 3/3, though the judge scored it 1/3 |
| review calibration `01`–`07b` (one run) | `FAIL` with the defect named in all but `03`, `04` and `06` (`06` allowed to miss) |
| `03`, `04` again, B20 vs `main` | `03` named 3/3 vs 2/3; `04` named 1/3 vs 2/3 |

# Plan — B20261025-2302-b20-gates-checks-review-scope

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **1/6 done** · iteration 3

**Brief:** `docs/briefs/B20261025-2302-b20-gates-checks-review-scope.loop-brief.md` · **Updated:** 2026-10-26T00:32:18Z

## Progress

- [x] **T1** — State P-8 in the domain model and add gate-verdict/v2 beside an unchanged v1
- [ ] **T2** — Halt the run with exit 2 when the gate verdict holds a check-gap finding, keeping nothing · 2 attempt(s)
- [ ] **T3** — Have the gate review rule on every criterion's coverage, and the planner write criteria something can judge
- [ ] **T4** — Keep vloop's records under .vloop/ out of the review's product constraints
- [ ] **T5** — Cite P-8 in the operator's playbooks and say what a check-gap halt is in them and the guide
- [ ] **T6** — Close the run: every gate at once and the worked example line for line

## Tasks

### T1 — State P-8 in the domain model and add gate-verdict/v2 beside an unchanged v1

`done` · depends on: none

Every later task cites P-8 and reads gate-verdict/v2, so both must exist first. P-8 makes 'a gate never runs the suite' a domain invariant every session obeys, and v2 is the gate verdict with the two new finding kinds, uncovered and check-gap, that the gate review will use. v1 must stay byte for byte as it is, because verdicts already recorded under .vloop/state/runs/ are validated against it.

**Acceptance**

- docs/domain/domain-model.md has, under "Plans and tasks", P-8 with the brief's wording: A gate judges the brief's contract from outside the work: it never runs the repository's suite, a check, or a test a task writes. The checks run the tests (P-7). Every acceptance criterion is judged by its task's gate, by a check the plan holds, or by the review.
- schemas/gate-verdict.v2.json exists and is embedded: vloop schema list shows gate-verdict/v2 and still shows gate-verdict/v1
- gate-verdict/v2 has v1's shape with the finding kinds uncovered and check-gap added and no other change of kinds: a verdict using every v1 kind validates as v2, an unknown kind does not
- vloop schema validate gate-verdict/v2 on a verdict with uncovered and check-gap exits 0; gate-verdict/v1 on the same verdict exits 1; gate-verdict/v1 on internal/schema/testdata/gate-verdict.valid.json exits 0
- schemas/gate-verdict.v1.json is unchanged, and every gate-review report committed under .vloop/state/runs/ still validates as gate-verdict/v1
- The task leaves committed tests the go check runs: valid and invalid v2 testdata under internal/schema/testdata/ exercised by the schema package's tests (a v2 verdict with the new kinds valid, an unknown kind invalid)

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T1/gate.sh
```

</details>

### T2 — Halt the run with exit 2 when the gate verdict holds a check-gap finding, keeping nothing

`pending` · 2 attempt(s) · depends on: T1

Only the operator can add a check to .vloop/config.toml, so sending a check-gap back to the planner spends a round on something it cannot fix. When any round's gate verdict has a check-gap, the driver prints every finding as today and halts with exit 2 before any work, committing nothing and leaving no plan, so the re-run plans afresh with the new check in its checks. A verdict without a check-gap must behave exactly as today.

**Acceptance**

- The driver validates the session's .vloop/tmp/gate-verdict.json as gate-verdict/v2
- When a gate verdict has any check-gap finding, in any round, the run prints every finding of the verdict as it does today (gate &lt;task&gt; (&lt;kind&gt;): &lt;summary&gt;), then ends with exit 2 and the line: vloop: the gate review found &lt;n&gt; check gap(s): the plan needs a check .vloop/config.toml does not have — add it, commit it, then re-run (see &lt;run folder&gt;reports/gate-review-&lt;round&gt;.json), where &lt;n&gt; counts the check-gap findings and &lt;run folder&gt; is repo-relative
- The check-gap halt sends nothing back to the plan session (no revision round), runs no work session, commits nothing on any branch and leaves no .vloop/state/state.json, gate folders or plan.md behind; the report it names exists
- A check-gap in round 2 halts with the check-gap line naming gate-review-2.json, not with "the gate review failed the plan twice"
- A verdict with no check-gap behaves exactly as today: uncovered and every other kind are sent back to the planner, and a second FAIL blocks the plan with "the gate review failed the plan twice"
- A re-run after the halt plans afresh (a new plan session, not "resuming acceptance")
- The task leaves committed tests the go check runs, covering the check-gap halt (round 1 and round 2, alone and beside another finding) and an uncovered finding being sent back

**From the last attempt:** No committed test runs the check-gap halt (exit 2, line, nothing kept, no revision round, no commit, round 1 and 2, alone and with another finding, re-run plans afresh) or shows an uncovered finding is still sent back; the task's last acceptance criterion is unmet; The gate (gate.sh) passes without these behaviours being exercised, so it does not enforce the acceptance list

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T2/gate.sh
```

</details>

### T3 — Have the gate review rule on every criterion's coverage, and the planner write criteria something can judge

`pending` · depends on: T1

In B19's acceptance run a criterion asked for the suites in a fresh copy, which no gate may run and no check covered, and nobody noticed before work began. The gate review must now rule, for every acceptance criterion, whether the gate, a named check the plan holds, or the review judges it, with two new finding kinds; and the plan skill must ask the planner the same question before it finishes. P-8 is cited, not restated, and the gate-review-coverage eval case is the behaviour's proof, since gates never run claude.

**Acceptance**

- plugin/skills/gate-review/SKILL.md reads the plan's checks and, for every task, rules on every acceptance criterion: judged by the gate, by a named check, or by the review (from the diff, or from a check's log)
- The gate-review skill names the finding kinds uncovered (no gate, check or review could judge the criterion as written) and check-gap (only a check could judge it, and no check the plan holds runs what it names), and its verdict section writes and validates gate-verdict/v2
- The gate-review skill's unpassable also covers a criterion no work session can meet: one requiring a change under .vloop/ outside .vloop/tmp/ (S-4) or ruling on files only the driver or the planner writes; and it names a gate that runs the suite, a check or a task's test as task-reach, citing P-8
- plugin/skills/plan/SKILL.md's own statement that a gate never runs the suite becomes a citation of P-8, not a second copy; its "Before you finish" gains: every acceptance criterion is one the gate, a check the plan holds, or the review can judge, and a criterion only a check can judge names that check
- Neither skill restates P-8's text or cites it by another id
- plugin/evals/gate-review-coverage/ follows the eval layout (prompt.md invoking /vloop:gate-review, case.yaml, graders/*.md, scaffold.sh); its plan's gates are all sound by the existing three checks, and its tasks carry a criterion only a check could judge, naming a script no check runs (graded check-gap), a criterion nothing could judge (graded uncovered), a criterion requiring a change to .vloop/state/plan.md (graded unpassable), and one task whose every criterion is covered (graded PASS, no finding)
- The case's scaffold exits non-zero if any planted gate fails the existing checks — at least, if a planted gate passes on the base or fails it other than at a behavioural assertion — so the case measures coverage alone; a scaffold that does not refuse such a gate violates the brief
- plugin/evals/README.md lists gate-review-coverage in its tables
- Word ceilings in internal/budget/budget.go rise, if at all, only by what this task adds, each with its date and reason in the comment; the go check's budget test passes

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T3/gate.sh
```

</details>

### T4 — Keep vloop's records under .vloop/ out of the review's product constraints

`pending` · depends on: T1

In B19's acceptance run the review ruled "no absolute paths in any committed file" against vloop's own records, which a work session cannot change, and blocked a task no session could fix. The review skill must judge criteria about files on the product, everything outside .vloop/, and report a defect found only in vloop's records as one line in notes. The 08-vloop-records-out-of-scope eval case proves the behaviour, since gates never run claude.

**Acceptance**

- plugin/skills/review/SKILL.md says the product is every file outside .vloop/; vloop's records under .vloop/ (state, plan, notes, journals, run reports, gate folders) are written by the driver and the planner and a work session cannot change them (S-4); a criterion, constraint or question about "any file" or "every committed file" is judged on the product
- The review skill says an absolute path, or any other defect, found only in vloop's records is a vloop defect: one line in notes, never a finding and never a reason for met: false
- The review skill says a gate that leaves the suite to the checks keeps P-8 and is not a gate-gap; a criterion a check judges is judged from that check's log; a criterion nothing could judge is a spec-gap with a line in notes naming the gap in the plan
- The absolute-path question asks about the files the task changed and the output it produced, and does not let a task's own files out of it — only vloop's records; the skill holds no path-shaped placeholder (no home directory, ~/ or drive path), only a description
- The review skill cites P-8 by id and does not restate it
- plugin/evals/08-vloop-records-out-of-scope/ follows the eval layout (prompt.md invoking /vloop:review, case.yaml, graders/*.md, scaffold.sh): correct work that passes its gate, .vloop/state/plan.md and a run report holding an absolute home-directory path and the placeholder text of an earlier finding, a gate that leaves the suite to a check that ran and passed (recorded in .vloop/tmp/checks.json with its log), and criteria including "no absolute paths in any committed file"; its graders expect PASS, no finding, and a notes line about the records
- The case's committed files hold no literal absolute home path: the scaffold composes the records' path when it runs; and its scaffold exits non-zero if the work fails its own gate
- plugin/evals/README.md's reviewer-calibration table lists 08-vloop-records-out-of-scope
- The review skill's word ceiling rises, if at all, only by what this task adds, with its date and reason in the comment; the go check's budget test passes

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T4/gate.sh
```

</details>

### T5 — Cite P-8 in the operator's playbooks and say what a check-gap halt is in them and the guide

`pending` · depends on: T2

The operator's repair of a gate is where B19's suite crept into a gate, and the check-gap halt is new behaviour the operator has to act on: add the missing check to .vloop/config.toml, commit it, re-run. Both the plugin's operate skill and this repository's own playbook must cite P-8 where they speak of a gate repair, and their halt tables and the concepts guide must say what the new exit-2 halt means and what to do.

**Acceptance**

- plugin/skills/operate/SKILL.md and .claude/skills/vloop-operator/SKILL.md both say: a gate repair keeps the gate's intent and never adds the suite, a check or a task's test to it (P-8); a criterion that needs a test run is a check the operator adds to .vloop/config.toml
- Both halt tables' rows for exit 2 cover the check-gap halt: what it is and that the operator adds the named check to .vloop/config.toml, commits it and re-runs
- Both cite P-8 by id where they speak of a gate's scope and neither restates it
- docs/guide/concepts.md's "The gate review" says what a check-gap halt is and what the operator does, and its exit-code row for 2 covers it
- The operate skill's word ceiling rises, if at all, only by what this task adds, with its date and reason in the comment; the go check's budget test passes

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T5/gate.sh
```

</details>

### T6 — Close the run: every gate at once and the worked example line for line

`pending` · depends on: T1, T2, T3, T4, T5

The tasks were judged one at a time; this proves them together on the finished tree, with the brief's worked example (the schema lines and the run lines, every planted failure included). After it the operator runs the two new eval cases and the existing gate-review and review calibrations, and records the results in the run record before merging.

**Acceptance**

- Every gate of T1 to T5 passes on the finished tree, run together
- Any fix this task needs is made in the product, not in a gate, and every check passes
- No gate of this plan runs the suite, a check or a test a task writes (P-8)

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T6/gate.sh
```

</details>


# Plan — B20261102-1844-b31-i110-i112-plan-phase-edges

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **3/5 done** · iteration 3

**Brief:** `docs/briefs/B20261102-1844-b31-i110-i112-plan-phase-edges.loop-brief.md` · **Updated:** 2026-11-02T19:59:05Z

## Progress

- [x] **T1** — Tighten brief-review/v1's brief to a brief's name, in place
- [x] **T2** — Let a complete plan of another brief pass vloop run's preflight and the bare vloop doctor
- [x] **T3** — Revert and warn what each plan and gate-review session changes outside .vloop/state/ and .vloop/tmp/
- [ ] **T4** — Write V1, V2 and V3 into the configuration guide, the plan and session pages and the fence audit
- [ ] **T5** — Close the run: every gate and the worked example on the final tree

## Tasks

### T1 — Tighten brief-review/v1's brief to a brief's name, in place

`done` · depends on: none

The driver accepts a brief review's report only when `brief` is the brief's file name without `.md`, but the schema accepts any string, so a skill example saying "path" drifted without any test noticing and a Copilot report was rejected in B29 (#112). This task makes the schema reject every `brief` the driver rejects — no `/` or `\`, ending `.loop-brief` — so a drifting example fails the existing skills test. It is a stated exception to the rule that a published schema does not change: no report the driver accepts today becomes invalid.

**Acceptance**

- schemas/brief-review.v1.json's `brief` is `{"type": "string", "pattern": "^[^/\\\\]*\\.loop-brief$"}` as the brief gives it (no `/` or `\`, ending `.loop-brief`); nothing else in the schema changes and its title stays brief-review/v1; no brief-review/v2 is added.
- `vloop schema validate brief-review/v1` exits 0 on the worked example's `B1-x.loop-brief` and `.loop-brief` reports and 1 on `docs/briefs/B1-x.loop-brief`, `docs\\briefs\\B1-x.loop-brief`, `B1-x.loop-brief.md` and `b`.
- Every kept .vloop/state/brief-reviews/*/*/brief-review.json still validates, and the review-brief skill's report example still validates.
- The driver's check in internal/driver/briefreview.go and its refusal line are unchanged, and no skill file changes.
- In internal/schema/schema_test.go only the brief-review/v1 cases built from the two literals whose `brief` is `b` change, and only that value, to `B1-x.loop-brief`; the work session names the contract item (V3) in its notes. A changed expectation anywhere else is a violation.
- Committed tests run by the go check gain the planted cases of the worked example (a `/` path, a `\` path, the `.md` file name and `b` refused; `B1-x.loop-brief` and `.loop-brief` accepted). A task that tightens the schema and leaves no such test fails this criterion.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T1/gate.sh
```

</details>

### T2 — Let a complete plan of another brief pass vloop run's preflight and the bare vloop doctor

`done` · depends on: none

B29 could not start because B28's finished plan referenced files B28 had moved, and the preflight validates whatever plan is on disk, so the operator deleted it by hand (#110). The planner resets another brief's plan anyway, so validating a complete one protects nothing; a resume, or a plan not complete, is the plan the run goes on with and stays validated. The doctor's plan check has to tell three cases apart — the bare `vloop doctor`, a run naming a brief, a run naming none — so `runPreflight` and `runDoctor` gain what that takes, in a form this task chooses.

**Acceptance**

- `vloop run <brief>`, when the plan on disk is `state/v2` with status `complete` and its `brief` differs from the brief named, does not validate the plan: the preflight prints nothing for the `plan` check and the run goes on to the planner's `the plan <run id> was for <plan's brief> — you asked for <brief>: resetting and planning fresh`.
- The bare `vloop doctor` does not validate a complete `state/v2` plan and prints `✓ plan a complete plan of <plan's brief> — the next brief's plan replaces it`, `<plan's brief>` the plan's `brief` field as stored; `vloop doctor --json` carries the same message with result `pass`.
- Validated as at the base: a plan whose status is not `complete`; a complete plan when `vloop run` names no brief or names the plan's own brief (`  ✗ plan <n> problem(s) — run vloop task validate`, then `vloop: preflight failed — 1 problem(s), nothing has run`, exit 1, nothing run); both state/v1 cases; `vloop task validate`, which validates whatever plan is there. A complete plan of another brief validated, or a resumed plan not, is a violation.
- The existing callers of `runDoctor` and `runPreflight` (internal/cli/doctor_copilot_test.go, internal/cli/doctor_second_review_test.go) change only to pass the new parameter, with no brief, and keep every assertion; the work session names the contract item (V1) in its notes.
- Committed tests run by the go check cover every V1 line of the worked example with stub sessions in a fixture repository, as cmd/vloop's run tests do — the bare doctor's pass (text and --json), `vloop run <b>` resetting and completing, `vloop run <a>` and a bare `vloop run` refused, and the planted running plan refused — plus the doctor's three cases in internal/cli. A task that changes this behaviour and leaves no such test fails this criterion.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T2/gate.sh
```

</details>

### T3 — Revert and warn what each plan and gate-review session changes outside .vloop/state/ and .vloop/tmp/

`done` · depends on: none

The plan commit stages the whole tree, so a file a plan or gate-review session changes elsewhere is committed with the plan, and only the fence stands in the way — for Copilot its hook fails open (#111). The operator chose reverting with a warning, as the review and verify sessions' guards do, around each session rather than once before the commit: a plan session's stray change would otherwise be in the tree while the gates are checked on the base and could make a gate pass there for the wrong reason, and per session the warning names who made it. This reuses the review's tree guard; it adds no new halt.

**Acceptance**

- Before each plan session (the first and each revision) and each gate-review session the driver snapshots the tree as the review's guard does, leaving out .vloop/state/ and .vloop/tmp/; after the session it reverts every path changed since: a tracked file restored, a new file removed, a deleted file put back.
- Each reverted path is one warning on stderr and in the run's run.log, in path order: `vloop: the plan session changed <path> — reverted` or `vloop: the gate-review session changed <path> — reverted`.
- Order after a session: the moved-ref and git-config checks first, halting as at the base without reverting; then the revert; then the session's ending judged as at the base — a plan session that timed out or exited non-zero, and a gate-review session that timed out, halt as at the base after their changes are reverted and warned; a gate-review session's exit code is treated as at the base. No new halt besides the could-not-revert one: `the plan session changed files the driver could not revert: <path>: <error>` (`gate-review` for that session), exit 1, before the plan commit.
- No plan or gate-review session's change outside .vloop/state/ and .vloop/tmp/ reaches the plan commit or is in the tree when the gates are checked on the base; that it does is a violation. The plan commit still stages the tree and leaves out .vloop/tmp/; a run whose sessions changed nothing outside those folders prints what it printed at the base.
- No existing test's expectation changes; if one pins a plan or gate-review session's stray change reaching the commit, the work session stops and reports it rather than editing it.
- Committed tests run by the go check, in cmd/vloop with stub plan and gate-review sessions as the base's run tests do, cover every V2 line of the worked example: the four warnings in order with the plan commit and tree checked, the plan session that exits 3 after a change, the plan-line gate that fails on the base after the revert, and a clean plan session with no warning and its gate folder committed. A task that ships the guard and leaves no such test fails this criterion.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T3/gate.sh
```

</details>

### T4 — Write V1, V2 and V3 into the configuration guide, the plan and session pages and the fence audit

`pending` · depends on: T1, T2, T3

The docs describe the behaviour T1–T3 change: the configuration guide still says the driver does not revert a plan session's changes, the plan page names only the state/v1 exemption, the session page's guards paragraph and review-brief paragraph do not say what V2 and V3 add, and the fence audit's two rows predate V2. Each text is literal as the brief gives it, wrapped in each file's own style. This is prose: if the review fails it twice on wording, the task is blocked and the operator finishes it.

**Acceptance**

- docs/guide/configuration.md, the plan fence bullet: "The driver does not revert a file a plan session changed elsewhere (the plan commit stages the whole tree), so the hook is the line." is gone and in its place: "The driver also reverts a file a plan or gate-review session changed outside `.vloop/state/` and `.vloop/tmp/`, with a warning (`vloop: the plan session changed <path> — reverted`), so it never reaches the plan commit."
- docs/domain/execution/plan.md: right after the first paragraph (ending "*Part of [execution](execution-context.md).*"), the paragraph "A complete `state/v2` plan of another brief does not block the next brief either: `vloop run` naming another brief does not validate it, and `vloop doctor` run on its own reports every complete `state/v2` plan as a pass; a resumed plan, or one not complete, is validated."
- docs/domain/execution/session.md: in the guards paragraph, right after "what a session changed among them is restored (…).", the sentence "A plan or gate-review session's change outside `.vloop/state/` and `.vloop/tmp/` is reverted after the session, with a warning."; in the paragraph beginning "A **review-brief** session is started by", right after the sentence ending "with no task.", the sentence "Its report's `brief` is the brief's name, its file name without `.md`; the schema rejects a path."
- docs/design-notes/fence-audit.md: under "# Fence audit: who enforces each rule in a Copilot gate-review session" and under "# Fence audit: who enforces each rule in a Copilot plan session", as the first paragraph, the note "Since B31 the driver reverts a file a gate-review or plan session changed outside `.vloop/state/` and `.vloop/tmp/`, with a warning, before the plan commit: where this audit says the driver does not revert files after such a session, or that the plan commit would commit the change, that predates it."; the rest of the audit, prose and rows, is unchanged.
- Every text is word for word as above; only its line wrapping (and, for the audit's note, italics matching the audit's own first paragraphs) follows the file's style. No other doc changes; docs carry no behaviour, so no test is owed beyond the gate.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T4/gate.sh
```

</details>

### T5 — Close the run: every gate and the worked example on the final tree

`pending` · depends on: T1, T2, T3, T4

The last task proves the whole brief holds at once on the final tree: every earlier gate passes together, no fence changed, and the worked example holds line for line. It exists so a regression one task caused in another's surface is found before the run's verify session, which cannot run `vloop run` itself and relies on the committed tests named here.

**Acceptance**

- Gates T1 to T4 all pass on the final tree, run together by this task's gate.
- fence/*.json are byte for byte as at the base, with no fence file added or removed; the Copilot hook is unchanged.
- The work session's notes walk the brief's worked example line for line (V1's doctor and run lines and the planted running plan; V2's four warnings, the plan commit, the exit-3 plan session, the plan-line gate and the clean session; V3's six schema lines, the kept brief reviews and the skill's example) and name the committed test or gate that covers each; a line no committed test covers is fixed in this task or reported, not skipped.
- Any fix this task makes is to the contract as the brief states it, leaves a committed test where it changes behaviour, and weakens no check.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T5/gate.sh
```

</details>


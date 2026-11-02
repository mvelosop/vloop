# Plan — B20261102-2101-b32-i84-gates-through-the-named-surface

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **3/4 done** · iteration 3

**Brief:** `docs/briefs/B20261102-2101-b32-i84-gates-through-the-named-surface.loop-brief.md` · **Updated:** 2026-11-02T22:03:21Z

## Progress

- [x] **T1** — Make the plan skill judge through the surface the brief names, and act on every revision finding (G1, G2)
- [x] **T2** — Say in the gate review when a pin is sound, and raise its word ceiling to 1600 (G3)
- [x] **T3** — Add the eval case plan-judges-through-named-surface and list it in both catalogues (G4)
- [ ] **T4** — Close the run: every gate and the worked example on the final tree

## Tasks

### T1 — Make the plan skill judge through the surface the brief names, and act on every revision finding (G1, G2)

`done` · depends on: none

The plan skill lets a planner pin any name a gate needs where the brief is silent, while the gate review flags such pins as `pins-unnamed`; a model that follows the skill literally (gpt-6.1-terra on 2026-11-02) fails the gate review by construction, and its revision round changed nothing it was told to. This task replaces the "Naming" section's second paragraph with the brief's fixed text (gates judge through the named surface, a pin only where no named surface reaches a criterion, stated in the task's acceptance), adjusts two phrases that depend on it, and adds two sentences to "A revision round" so every finding is acted on. The text is fixed so that gates and the gate review can find it; nothing else in the skill moves.

**Acceptance**

- In plugin/skills/plan/SKILL.md the paragraph beginning `**Where the brief is silent, you may pin what a gate needs.**` is replaced, whole, by the brief's G1 paragraph beginning `**Where the brief is silent, judge through what it names.**` and ending `a cost the operator should see rather than discover.`, word for word, wrapped as the file wraps; the section's first paragraph is unchanged.
- The "Known trap" parenthesis reads `(a file's contents, a `--help` exit code, an import of a module the brief names)`, and the "Gates are judged, not self-checked" sentence ends `pins no name the brief does not, beyond a pin "Naming" allows.`; the rest of both paragraphs is unchanged.
- In "A revision round", the two G2 sentences (`Each finding names a task: change that task's gate, or the task itself when the finding is about a criterion, until the finding no longer holds. A gate you leave as it was gets the same finding again.`) sit right after `not patched until it fails.` and before `You get one revision round;`.
- Violation to rule against: any other sentence of the plan skill changed, moved or reworded, or any other file changed (the gate compares the whole file's joined text to the base with exactly these edits).
- The skill stays within its 4100-word ceiling and the configured go check (TestPluginSkills and the budget test among it) is green; the task changes behaviour of no program, so it leaves no new test, but it may not weaken an existing one.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T1/gate.sh
```

</details>

### T2 — Say in the gate review when a pin is sound, and raise its word ceiling to 1600 (G3)

`done` · depends on: none

The gate review's `pins-unnamed` and `not-contract` must apply the same rule the plan skill now states, or a planner following the plan skill is still sent back for a sound pin. This task adds the brief's one sentence to the `pins-unnamed` bullet, one parenthesis to `not-contract`, and, because the gate-review skill had only 23 words of headroom, raises its ceiling in internal/budget/budget.go from 1550 to 1600 with a dated comment. Finding kinds and every other sentence stay as they are.

**Acceptance**

- plugin/skills/gate-review/SKILL.md's `pins-unnamed` bullet keeps its sentence and gains, right after it in the same bullet, `A pin is sound only where no surface the brief names (a command, a route, a file) could judge the same criterion, and its task's acceptance states the name.`
- The `not-contract` bullet reads `or internals the brief does not name (beyond a sound pin, as below),` where it read `or internals the brief does not name,`; the rest of that bullet is unchanged.
- Violation to rule against: any other sentence of the gate-review skill changed, moved or reworded, or a finding kind added, removed or renamed (the gate compares the whole file's joined text to the base with exactly these two edits).
- `wc -w < plugin/skills/gate-review/SKILL.md` is at most 1600, and internal/budget/budget.go's `Ceilings["gate-review"]` is `1600` with the comment `2026-11-02, B32 (#84): 1550 (B14) plus the sound-pin sentence of pins-unnamed and its reference in not-contract`; no other ceiling changes.
- The configured go check (the budget test and TestPluginSkills among it) is green; the existing budget tests, which derive from `Ceilings`, cover the new value, and no test is weakened.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T2/gate.sh
```

</details>

### T3 — Add the eval case plan-judges-through-named-surface and list it in both catalogues (G4)

`done` · depends on: none

The skill edits of T1 need a test that holds a planner to them: an eval case whose brief names a command (`bin/tally`) and leaves its shell library's file and function names to the implementer, graded on the plan validating and on no gate naming the library. The case is built like plan-checks-its-gates (prompt, case.yaml, graders, brief, scaffold) and joins plugin/evals/README.md and docs/guide/evals.md, whose case count is corrected to 23. The case must leave the temptation to pin the library, not describe it; no session runs it, the operator does at close.

**Acceptance**

- plugin/evals/plan-judges-through-named-surface/ holds prompt.md (the same frontmatter as plan-checks-its-gates's, body `/vloop:plan docs/briefs/tally.loop-brief.md`), case.yaml (`schema_version: "1.1"`, `name: plan-judges-through-named-surface`, `context.scaffold_script: scaffold.sh`), brief.md, scaffold.sh, graders/state-validates.md (a copy of plan-checks-its-gates's) and graders/judges-through-the-command.md (`type: llm`, `weight: 3`, focus the file `.vloop/state/state.json`).
- scaffold.sh, run with bash in an empty directory, exits 0 after `require_empty`, `git_init`, copying brief.md to docs/briefs/tally.loop-brief.md, writing a one-line CLAUDE.md and a mode-755 `bin/tally` (`#!/bin/sh`, prints `tally: not implemented` on stderr, exits 1), and `commit_all "baseline"`; the tree is clean, exactly those three files are tracked, no .vloop/ is written, and a second run in the same directory exits non-zero.
- brief.md has frontmatter `name: tally`, `kind: brief`, `status: ready` and a description, and states exactly the brief's contract: `bin/tally add <n>` adds to the total in `tally.txt` in the directory it is run from (created when missing), prints the new total, exits 0; `bin/tally total` prints the total, `0` when `tally.txt` is missing, exits 0; `bin/tally add x` prints `tally: not a number: x` on stderr, exits 2 and leaves `tally.txt` as it was; the arithmetic and file handling may live in a shell library `bin/tally` sources, whose existence, file name and function names `are the implementer's`; plan shape one or two tasks, each judged by its `verify` command alone, with no file in a gate folder.
- Violations to rule against: brief.md pins a library file or function name (it contains no `lib/`) or names the bad gate it guards against; anything in the case or the catalogues mentions the URL shortener template or its brief.
- judges-through-the-command.md scores 1 when every task's `verify` judges by running `bin/tally` and reading only its stdout, stderr, exit code and `tally.txt`, and no `verify` or acceptance criterion sources, runs or names a library file or function; 0 otherwise.
- plugin/evals/README.md's "The other cases" table has the brief's row for `plan-judges-through-named-surface` word for word; docs/guide/evals.md's `/vloop:plan` row lists `plan-judges-through-named-surface` among its cases and its last cell ends `, gates through the surface the brief names`; its cost sentence says `is 23 cases × 3 runs × 2 arms`.
- The case is itself the skill's test; the configured go check (which holds eval cases to the fence and the guide's list to the cases, e.g. TestEvalGrantsCoverFence) is green with it, and any committed list of cases those tests read is updated so they cover the new case rather than skip it.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T3/gate.sh
```

</details>

### T4 — Close the run: every gate and the worked example on the final tree

`pending` · depends on: T1, T2, T3

The last task proves the whole brief holds at once on the final tree: gates T1 to T3 pass together, no schema or other skill changed, and the brief's worked example holds line for line. It exists so a regression one task caused in another's surface is found before the run's verify session; the operator's acceptance at close (the evals and gpt-6.1-terra on every phase) cannot run in a session and is not this task's.

**Acceptance**

- Gates T1 to T3 all pass on the final tree, run together by this task's gate.
- schemas/*.json are byte for byte as at the base with none added or removed, and the work, review, verify, operate, create-brief and review-brief skills are unchanged.
- The work session's notes walk the brief's worked example line for line (the plan skill's five phrases and the absent one, the gate review's two phrases, its word count, the ceiling, the configured go check, the scaffold run, `bin/tally total`, the clean status, the three tracked files, the refused second run, brief.md's strings and the catalogues) and name the gate or committed test that covers each; a line nothing covers is fixed in this task or reported, not skipped.
- No task runs `claude`, `copilot` or any eval; any fix this task makes is to the contract as the brief states it and weakens no check.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T4/gate.sh
```

</details>


---
name: B20261102-2101-b32-i84-gates-through-the-named-surface.loop-brief
description: Make the plan skill judge through the surface a brief names and act on every gate-review finding, align the gate review's pins-unnamed with it, and add an eval case, so gpt-6.1-terra can be qualified to replace gpt-5.6-terra on every Copilot phase
kind: brief
status: consumed
created: 2026-11-02
issue: 84
seeds: Issue 84 (gpt-5.6-terra retires on 2026-11-16), and the gpt-6.1-terra planning test of 2026-11-02, whose plan the gate review sent back twice for pinning internal names; decided by the assistant under the operator's delegation of 2026-11-02 for the #81–#88 chain
depends-on: [B20261102-1844-b31-i110-i112-plan-phase-edges.loop-brief]
---
# Brief — vloop B32: gates judge through the surface the brief names

- **Issue:** #84 (under #69). The PR closes it only if the acceptance at
  close qualifies `gpt-6.1-terra` on every phase; otherwise it names the
  phases still open.
- **Starting point:** extends `main` at `7de2855` (the v0.15.0-rc.1
  release), plus this brief's ready merge. The planner pins that SHA as the
  **base** every gate compares against — never `HEAD~n`.
- **Produced by:** the operator's delegation of 2026-11-02 for a chain of
  briefs on #84, #81, #82, #83, #96 and #88 ("run all of them as we did
  yesterday"). Written with `plugin/skills/create-brief/`; every decision
  below was taken by the assistant under that delegation.
- **Review:** `vloop brief review` (Claude Opus). Round 1: FIX, 3
  findings, all fixed (the eval's library made a criterion no named
  surface reaches, `tally.txt`'s directory, the evals' baseline). Round 2:
  FIX, 4 findings, all fixed (the baseline's skills, `language-es`, the
  "Known trap" parenthesis, the catalogue line). Round 3: FIX, 2
  findings, both fixed (the "Gates are judged" sentence, `--case`'s glob
  from the tool's help). Round 4: FIX, 1 finding, fixed (G1 forbids the
  implementation's names, not the gate's own folder). Round 5: FIX, 2
  findings, both fixed (the guide's case count, `--runs 1`). Round 6: FIX,
  2 findings, both fixed (the count is 23, `not-contract` also defers to a
  sound pin). Round 7: FIX, 3 findings, all fixed (no gate runs the tests,
  the out-of-scope list and references, a shell-safe expected string).
  Round 8: FIX, 1 finding, fixed (the summary names every edit). Round 9:
  FIX, 1 finding, fixed by deleting a wrong parenthesis; not re-reviewed,
  the fix only removes words. No `ASK` in any round; every decision is the
  assistant's under the operator's delegation.

## Executive summary

`gpt-5.6-terra` retires from Copilot on 2026-11-16, and every phase vloop
runs on Copilot was accepted on it. Its cheaper successor, `gpt-6.1-terra`,
planned the URL shortener brief on 2026-11-02 with gates that were real
judges, but eight of nine reached into the implementation: they imported
modules and functions the brief leaves open. The gate review flagged them
`pins-unnamed` in both rounds, and the revision round changed nothing it
was told to. The plan skill allows exactly that ("Where the brief is
silent, you may pin what a gate needs"), while the gate review's
`pins-unnamed` forbids it, and the revision round never says that each
finding must change something.

**The goal:** the plan skill and the gate review agree on when a gate may
pin a name — only where no surface the brief names reaches the criterion —
and a revision round acts on every finding; an eval case holds the plan
skill to it; and `gpt-6.1-terra` is tried on every phase, at close, as
`gpt-5.6-terra` was.

- **G1 — the plan skill judges through what the brief names.** Its
  "Naming" section's second paragraph is replaced: a gate judges through
  the surface the brief names (a command, a route, a file), fixtures
  included, and pins a name only where no named surface reaches the
  criterion, stating it in the task's acceptance. Two phrases elsewhere in
  the skill follow it: the "Known trap" example and "Gates are judged".
- **G2 — a revision round acts on every finding.** Two sentences join "A
  revision round": each finding's task changes until the finding no longer
  holds; a gate left as it was gets the same finding again.
- **G3 — the gate review's `pins-unnamed` says when a pin is sound.** One
  sentence joins its bullet, and `not-contract` defers to a sound pin; the
  gate-review skill's word ceiling rises from 1550 to 1600.
- **G4 — an eval case.** `plan-judges-through-named-surface`: a brief that
  names a command and leaves its library's names open; graded on the plan
  validating and on no gate naming the library.
- Unchanged: every schema, the driver, the finding kinds, the other
  skills, every part of these two not named above, every default model
  and agent.

## What it is

Two skill edits, one ceiling and one eval case, so that a planner on any
model is told the same rule the gate review applies. The gates of the
2026-11-02 attempt shared one blind spot: nothing told the planner that a
name it was allowed to pin was one the gate review would reject, and
nothing told the revision round that a finding left unanswered comes back.
A model that follows the plan skill literally, as `gpt-6.1-terra` did,
fails the gate review by construction.

The qualification of `gpt-6.1-terra` itself is the operator's, at close: a
whole run of the URL shortener brief with every phase on Copilot.

## Why this shape, and what was rejected

Decided by the assistant on 2026-11-02 under the operator's delegation.
**Do not re-litigate these** — the planner inherits them.

- **Change the plan skill, not the gate review's standard.** The gate
  review is right: a gate that imports slug.js's generateSlug fails a
  correct implementation that named it otherwise. The plan skill's
  permission is what is too wide.
  *Rejected:* dropping `pins-unnamed` or accepting any pin the plan
  reports — it would pass gates that fail correct work.
- **A pin stays possible where no named surface reaches the criterion.**
  Some briefs name only a library's behaviour; forbidding every pin would
  leave those criteria unjudgeable.
  *Rejected:* "never pin a name": it turns such criteria into `uncovered`
  findings.
- **The pinned name goes into the task's acceptance.** The work session
  and the review then hold the same name the gate does, and the gate
  review can see it is stated.
  *Rejected:* a separate list in the plan's report only — the gate review
  does not read the report.
- **`gpt-6.1-terra` is the candidate.** Same family as 5.6, about a third
  of its credits, and the one already tried as planner.
  *Rejected for this brief:* `gpt-6-luna`, `claude-sonnet-5.5` through
  Copilot, `kimi-k3`: untried; they can follow through #81's runner.
- **Qualification is the operator's acceptance, not a task.** A session
  cannot run Copilot (no network, no credentials), so no gate can.
- **The ceiling rises for the gate review only.** The plan skill has room
  (3,804 of 4,100 words); the gate review has 23 words of headroom.

## Binding references

- `plugin/skills/plan/SKILL.md` — the "Naming" section (its second
  paragraph is replaced), "A revision round" (two sentences join it), the
  "Known trap" paragraph and "Gates are judged, not self-checked" (one
  phrase each).
- `plugin/skills/gate-review/SKILL.md` — section 3's `pins-unnamed` bullet,
  which gains one sentence, and its `not-contract` bullet, one phrase.
- `internal/budget/budget.go` — `Ceilings`, where `gate-review` rises to
  1600 with a dated comment, as the other entries have.
- `plugin/evals/plan-checks-its-gates/scaffold.sh` — the model for the new
  case: an empty directory, `_lib.sh`'s `require_empty`, `git_init` and
  `commit_all`, a brief under `docs/briefs/`.
- `plugin/evals/plan-checks-its-gates/graders/state-validates.md` — the
  regex grader the new case copies.
- `plugin/evals/README.md` — the catalogue of cases the new case joins.
- `docs/guide/evals.md` — the guide's table of cases per skill, which the
  new case joins.
- `docs/domain/domain-model.md` — P-4 and P-8: a gate's judge is the
  planner's and judges the brief's contract from outside the work; this
  brief narrows how, it does not change them.
- `docs/design-notes/copilot-cli-experiments.md` — section "`gpt-6.1-terra`:
  planning and the URL probe": the attempt this brief answers, and where
  the acceptance at close is recorded.

## Behaviour contract

### G1 — the plan skill judges through what the brief names

In `plugin/skills/plan/SKILL.md`, section "Naming", the paragraph that
begins `**Where the brief is silent, you may pin what a gate needs.**` is
replaced, whole, by exactly this paragraph (wrapped as the file wraps):

> **Where the brief is silent, judge through what it names.** When the
> brief names a surface the behaviour shows through (a command and its
> output, an HTTP route and its responses, a file and its contents), every
> gate judges through that surface, fixtures included: it runs the
> command, calls the route or reads the file, and imports, calls or reads
> no module, function, file or label of the implementation that the brief
> leaves open (the gate's own folder and temporary files are yours, not
> the implementation's). Only where no
> surface the brief names reaches a criterion may you pin the minimum a
> gate needs, and no more; state each pinned name in its task's acceptance
> criteria and report it. A pin constrains structure rather than just
> behaviour, a cost the operator should see rather than discover.

The section's first paragraph is unchanged.

In the same file, the "Known trap" paragraph's parenthesis `(an import, a
file's contents, a `--help` exit code)` becomes `(a file's contents, a
`--help` exit code, an import of a module the brief names)`; the rest of
that paragraph is unchanged. And in section "Gates are judged, not
self-checked", `pins no name the brief does not.` becomes `pins no name
the brief does not, beyond a pin "Naming" allows.`

### G2 — a revision round acts on every finding

In the same file, section "A revision round", these two sentences are
inserted after the sentence ending `not patched until it fails.` and
before `You get one revision round;`:

> Each finding names a task: change that task's gate, or the task itself
> when the finding is about a criterion, until the finding no longer
> holds. A gate you leave as it was gets the same finding again.

### G3 — the gate review's `pins-unnamed`

In `plugin/skills/gate-review/SKILL.md`, the `pins-unnamed` bullet of
section 3 keeps its sentence and gains this one after it:

> A pin is sound only where no surface the brief names (a command, a
> route, a file) could judge the same criterion, and its task's acceptance
> states the name.

The `not-contract` bullet's `or internals the brief does not name,` becomes
`or internals the brief does not name (beyond a sound pin, as below),`;
the rest of that bullet is unchanged.

In `internal/budget/budget.go`, `Ceilings["gate-review"]` becomes `1600`,
its comment `2026-11-02, B32 (#84): 1550 (B14) plus the sound-pin sentence
of pins-unnamed and its reference in not-contract`.

### G4 — the eval case

A new case, plugin/evals/plan-judges-through-named-surface/, built like
`plan-checks-its-gates`: `prompt.md`, `case.yaml`, `graders/`, a `brief.md`
and a `scaffold.sh`.

- **`prompt.md`**: the same frontmatter as `plan-checks-its-gates`'s
  (`max_turns: 80` and its `allowed_tools`), and the body
  /vloop:plan docs/briefs/tally.loop-brief.md.
- **`case.yaml`**: `schema_version: "1.1"`, `name:
  plan-judges-through-named-surface`, `context.scaffold_script:
  scaffold.sh`.
- **`scaffold.sh`**, a bash script like the other scaffolds, run in an
  empty directory: `require_empty`,
  `git_init`; it copies `brief.md` to docs/briefs/tally.loop-brief.md,
  writes a `CLAUDE.md` of one line, and a `bin/tally` (mode 755) that is
  `#!/bin/sh`, prints `tally: not implemented` on stderr and exits 1; then
  `commit_all "baseline"`. It writes no `.vloop/state/`.
- **`brief.md`**: frontmatter `name: tally`, `kind: brief`, `status:
  ready`, and a description; then a brief whose contract is exactly:
  - `bin/tally add <n>`, `n` an integer, adds `n` to the total kept in
    `tally.txt` in the directory it is run from (created when missing),
    prints the new total and exits 0.
  - `bin/tally total` prints the total, `0` when `tally.txt` is missing,
    and exits 0.
  - `bin/tally add x` prints `tally: not a number: x` on stderr, exits 2
    and leaves `tally.txt` as it was.
  - The arithmetic and the file handling may live in a shell library that
    `bin/tally` sources, so other scripts can source it too; whether there
    is one, its file name and its function names are the implementer's.
    Only `bin/tally`'s behaviour above is the contract.
  - Its plan shape: one or two tasks, and each task's judge is its
    `verify` command alone, with no file in a gate folder.
- **`graders/state-validates.md`**: a copy of
  `plan-checks-its-gates/graders/state-validates.md`.
- **graders/judges-through-the-command.md**: `type: llm`, `weight: 3`,
  `focus` the file `.vloop/state/state.json`; it scores 1 when every
  task's `verify` judges by running `bin/tally` and reading only what the
  brief names (its stdout, stderr, exit code and `tally.txt`), and no
  `verify` or acceptance criterion sources, runs or names a library file
  or function; 0 otherwise.

The case joins the catalogue table of `plugin/evals/README.md` ("The other
cases") with the row

`| plan-judges-through-named-surface | /vloop:plan | a brief that names a command and leaves its library's file and function names open: the plan validates, and every gate judges through the command, naming no library file or function |`

(each of the three cells' names in backticks, as the table's other rows
have them), and the `/vloop:plan` row of `docs/guide/evals.md`'s table
gains `plan-judges-through-named-surface` in its cases cell and `, gates
through the surface the brief names` at the end of its last cell; the
guide's cost sentence `is 21 cases × 3 runs × 2 arms` becomes `is 23
cases × 3 runs × 2 arms` (22 cases before this one: the count has been
stale since B30).

### Decided here, because the cited documents leave it underdetermined

- The new sentences are fixed text so that a gate can find them; the
  skills' other wording does not change.
- The eval's library is a shell library so the case needs nothing but a
  shell.
- No domain invariant changes: P-4 and P-8 already hold the principle; the
  rule on pins is a planning practice, so it lives in the skills.

### Violations the review must rule on

- Any other sentence of the plan or gate-review skill changed, moved or
  reworded, beyond G1's "Known trap" parenthesis and "Gates are judged"
  sentence, and G3's `not-contract` parenthesis.
- A finding kind added, removed or renamed, or a schema changed.
- The eval's brief pinning the library's file or function names, or
  naming the bad gate it guards against: the case must leave the
  temptation, not describe it.
- Any reference, in the case, the skills or the catalogues, to the URL
  shortener template or its brief.

## Worked example

The joined form of a file is its text with newlines turned into spaces and
runs of whitespace collapsed to one; `grep -F` runs on it.

```
joined plugin/skills/plan/SKILL.md contains
  "**Where the brief is silent, judge through what it names.**"
  and "state each pinned name in its task's acceptance criteria and report it."
  and "A gate you leave as it was gets the same finding again."
  and "an import of a module the brief names)"
  and 'pins no name the brief does not, beyond a pin "Naming" allows.'
  and not "you may pin what a gate needs"
joined plugin/skills/gate-review/SKILL.md contains
  "or internals the brief does not name (beyond a sound pin, as below),"
  and
  "A pin is sound only where no surface the brief names (a command, a route, a file) could judge the same criterion, and its task's acceptance states the name."
wc -w < plugin/skills/gate-review/SKILL.md                  -> at most 1600
internal/budget/budget.go                                   -> "gate-review":  1600
the configured go check (go test ./... and the rest)       -> green

R=$(git rev-parse --show-toplevel); d=$(mktemp -d)
(cd "$d" && bash "$R"/plugin/evals/plan-judges-through-named-surface/scaffold.sh)
                                                            -> exit 0
"$d"/bin/tally total                                        -> stderr "tally: not implemented", exit 1
git -C "$d" status --porcelain                              -> empty
git -C "$d" ls-files                                        -> CLAUDE.md, bin/tally, docs/briefs/tally.loop-brief.md
running the scaffold again in "$d"                          -> exit non-zero (not empty)

plugin/evals/plan-judges-through-named-surface/brief.md     -> contains "bin/tally add x", "tally: not a number: x",
                                                               "are the implementer's"; does not contain "lib/"
plugin/evals/README.md, docs/guide/evals.md                 -> each contains "plan-judges-through-named-surface"
```

## Acceptance, at close (the operator's)

Neither line can run in a session (they call models over the network).

1. **The evals.** With a binary built from the branch first on `PATH`,
   the invocation `docs/guide/evals.md` gives, narrowed with `--case
   'plan-*' --case language-es --case 'gate-review-*'` (`claude plugin eval --help`:
   `--case <glob>`, "Filter cases by name glob"), one run each with
   `--runs 1` (`--runs <n>`, "Override per-case runs"). Expected:
   the new case scores 1.00 with the plugin. Any other case below 1.00 is
   run again from a worktree of `main` at `7de2855`, with a binary built
   there, so its skills are main's; the branch's score is not lower. The scores, the reruns and the cost go in
   `plugin/evals/README.md`, as the first record of these cases.
2. **`gpt-6.1-terra` on every phase.** A fresh clone of
   `mvelosop/url-shortener-template` under
   `~/source/personal/url-shortener-tests/b32-terra61`, the URL shortener
   brief, the `app` and `clean` checks, `agent.<phase> = "copilot"` and
   `model.<phase> = "gpt-6.1-terra"` for plan, gate-review, work, review
   and verify (second review off), and a binary built from the branch:
   `vloop run` the brief; then `vloop brief review` of the same brief with
   `agent.review-brief = "copilot"` and `model.review-brief =
   "gpt-6.1-terra"`. A phase qualifies when its sessions end with a valid
   report and the loop's outcome is right (the gate review's findings are
   real, the reviews' verdicts hold up, verify's findings are real). The
   result is recorded in a new section of
   `docs/design-notes/copilot-cli-experiments.md` and in a comment on #84.

## Out of scope

- The model-qualifying eval runner (#81), the hook's fail-open (#82), the
  fence's compound commands (#83).
- Any other candidate model, or changing a default agent or model.
- The URL shortener template and its brief: untouched.
- Any part of the plan or gate-review skill beyond what G1, G2 and G3
  name; the work, review or verify skills; the gate-review report's
  finding kinds.
- A driver check on pinned names.

## Constraints

- Gates run on macOS's BSD tools: no empty alternatives in `grep -E`, no
  `sed -i` without a suffix, no `\+` or `\|` in basic regexes.
- Sessions never move refs.
- Tests and gates build their fixtures in temporary directories and
  address them with `git -C`.
- A docs or skill gate joins lines and collapses whitespace before it
  greps, so a correct edit wrapped in the file's style passes.
- The skills test (`TestPluginSkills`) and the budget test stay green after
  every task; the configured check runs them, no gate does.
- No task runs `claude` or `copilot`, or any eval.
- Repo-relative paths everywhere.

## Shape

3–4 tasks; every configured check green after each.

No gate runs the repository's tests: the configured `go` check runs them
after every task (P-7, P-8).

1. **The plan skill: G1 and G2.** Gated on its joined-text lines.
2. **The gate review: G3 and its ceiling.** Gated on its joined-text
   lines, the word count and the ceiling's value in
   `internal/budget/budget.go`.
3. **The eval case and the catalogues (G4).** Gated on the scaffold and
   catalogue lines of the worked example, the scaffold run in a temporary
   directory.
4. **Close:** every gate at once, and the worked example line for line.
<!-- vloop:run-record:begin -->
## Run record

Generated by vloop brief close on 2026-11-02. Recompute with: vloop metrics B20261102-2101-b32-i84-gates-through-the-named-surface.loop-brief

```
B20261102-2101-b32-i84-gates-through-the-named-surface  consumed · not merged
 tasks     4 planned · 4 done · 0 blocked · first-pass 4/4
 size      delivered  code 1 · test 80 · docs 23 · test:code 80.00
           churn      code 1 · test 80 · docs 23 · rework 1.00
 time      agent 17.2 min (work 15.1 · review 2.1) · plan 4.3 min · gate review 1.3 min · verify 1.1 min · gates 0.0 · wall 42.3 min
 rate      0.1 code lines/min · 4.7 incl. tests
 cost      $2.79 · plan 1.37 · work 0.58 · review 0.38 · gate review 0.27 · verify 0.20 · $2794.06/kloc
 models    plan claude-opus-5-5 · work claude-sonnet-5-5 · review claude-sonnet-5-5 · gate review claude-sonnet-5-5 · verify claude-sonnet-5-5
 defects   in-loop 0 · operator 1 · escaped 0 · removal efficiency 100%
 interventions  3 · recommended 0 · other-option 0 · adjusted 0 · different 0 · no-options 3
 instructions  plan 4,693 · gate review 2,338 · work 2,435 · review 2,456 · verify 2,109 words · guidelines 2 rules
```

```
id  area  kind  att  code+  test+  docs+  other+  test:code  $/kloc  agent  cost   model
T1  -     docs  1    0      0      18     0       -          -       4m58s  $0.23  claude-sonnet-5-5
T2  -     docs  1    1      0      3      0       0.00       182.88  3m13s  $0.18  claude-sonnet-5-5
T3  -     test  1    0      80     2      0       -          -       1m39s  $0.31  claude-sonnet-5-5
T4  -     test  1    0      0      0      0       -          -       7m22s  $0.24  claude-sonnet-5-5
```

- I20261102-230010-the-eval-case-s-brief-md-left-out-two-ph — verification-finding, work, found by operator: the eval case's brief.md left out two phrases G4 requires (that other scripts can source the library, and that only bin/tally's behaviour is the contract); found by the verify session, fixed by the operator's assistant
<!-- vloop:run-record:end -->

---
name: vloop-operator
description: The operator's playbook for this repo — turn the operator's intent into a loop brief, run it with the shell loop on a work branch, handle halts, verify the result independently, record defects, close the brief, and merge. Use in an interactive session whenever the operator asks to write, check, run, resume, verify, close or merge a brief, or asks what the next brief is.
---

# Operating the loop

You are the operator's hands in an interactive session — not a loop session.
The loop's sessions do one task and never commit; you drive whole briefs, and
you commit on work branches. Merging, pushing, opening PRs and anything else
outward-facing wait for the operator's explicit go-ahead, every time.

The brief is the contract; the roadmap (`docs/design-notes/vloop-roadmap.md`) is
the runbook: its rows say what each brief owns, and the next row's out-of-scope
list is the rows after it. Read the roadmap and the last consumed brief's
`## Run record` before anything else.

## 1. The design act — intent to brief

1. **Survey before you ask.** The roadmap row, the previous briefs' run records
   ("Still open"), the code the brief will touch, and the shell loop's own
   formats if the brief reads them (`.loop/run.sh`, `.loop/state/runs/`).
   Measure real numbers from the telemetry rather than quoting old run records —
   those were hand-rounded.
2. **Settle forks with the operator**, few at a time, each with a recommendation.
   Record what you decided yourself as "choices to review" in your reply.
3. **Write the brief** from `.loop/loop-brief.template.md`, named
   `docs/briefs/B<YYYYMMDD-HHMM>-<slug>.loop-brief.md`, frontmatter
   `status: draft` and `depends-on:` the previous brief's name. Pin decisions,
   leave mechanics open. Every brief so far also carried:
   - a worked example with exact values, **plus** planted failures;
   - a real-data check where the slice can read this repo's history;
   - the constraints section's lessons (below), verbatim where they apply;
   - an explicit list of earlier tests that *must* change (e.g. `config list`
     gaining keys) — otherwise "no task may weaken a check" blocks the task.
4. **Check it twice**: `.loop/check-brief.sh <brief>` (the shell loop plans from
   it) and `vloop brief check` on a copy with `status: ready` (vloop's rules are
   stricter: one path per binding reference). Aim for 0 problems, 0 warnings;
   un-backtick paths that exist only after the run.
5. Point the roadmap row at the brief. Stop for the operator's review.

**Where defects come from.** Of the six recorded for B1–B3, three are spec gaps
in the briefs themselves. The review of the brief is the cheapest place to find
the next one: look for sentences two sessions could read two ways.

## 2. Running

1. Operator says go: set `status: ready`, then **create the work branch before
   anything else** — named for the run id (the brief name minus `.loop-brief`):
   `git switch -c <run id>`, commit the brief there. B1's plan commit landed on
   `main` because the branch came after the planning preview.
2. Run in the background and wait for the notification; do not poll:
   `.loop/run.sh <brief path>`, output to a log in your scratchpad.
   `--plan-only` first when the operator wants to see the task split.
3. Resuming on another branch than the one that planned needs the brief path
   again; the driver refuses without it (exit 1) and says so.

## 3. When it halts

Read the log's end, the journal (`.loop/state/journals/<run id>.md`) and the
blocked task's `notes` in `.loop/state/state.json`.

| Exit | Meaning | What you do |
| --- | --- | --- |
| 0 | complete | verify (step 4) |
| 2 | blocked | read the notes; usually a design question → the operator |
| 3 | stalled | read the notes. A **gate defect** the work session diagnosed, whose fix keeps the gate's intent, you may fix: `.loop/amend.sh verify <id> '<cmd>'`, then `.loop/amend.sh reset <id>`, run the new gate once by hand, resume. Record it later as a `plan`/`gate` defect |
| 4, 6 | budget | resumable; tell the operator |
| 5, 7, 8 | needs a human | report, don't retry blindly |
| 9 | a session moved git refs | **do not re-run.** Compare `git branch -vv`, `git reflog`, `git show-ref` against what the log lists; restore refs (nothing was committed); tell the operator |

Anything that is a design decision rather than a mechanical defect goes back to
the operator, even when you could guess.

## 4. Verifying — never trust a clean run

B1 closed 9/9 first-pass and still had an untidy `go.mod`; B3 closed 10/10 and
reported its own estimate wrong and its branches renamed. So, every time:

1. **Refs first**: `git status -sb`, `git branch -vv`, `git symbolic-ref
   refs/remotes/origin/HEAD`. The run's commits must be on the work branch, and
   `main` must still be `main`.
2. The full suite and checks (`CLAUDE.md` → Toolchain).
3. **The worked example by hand**, in a scratch repo, with a freshly built
   binary — every line and exit code, and the planted failures.
4. The real-data check if the brief has one, and `vloop metrics <brief>` — read
   it critically: the estimate, the records line, the merged state.
5. Which earlier tests changed (`git diff <base> HEAD -- <test files>`) and
   whether each change is one the brief required.
6. Permission denials and missing session records in the log — explain each.

Report to the operator: what passed, what you found, and what you recommend,
before closing anything.

## 5. Closing and merging

Until B4 ships `vloop brief close`, by hand:

1. Fix what the operator agrees to fix, on the work branch, test first.
2. Record defects: `vloop defect add "<summary>" --brief <name> --task <id>
   --origin <brief|plan|work|env> --found-by <gate|review|operator|user>
   --kind <…>`, then `vloop defect set <id> status fixed` (and `fixed-by`).
3. Mark the brief consumed: frontmatter `status: consumed`, the body
   `**Status:**` line to "consumed — … **Do not re-plan from this brief.**",
   and a `## Run record` — outcome, spend (from `vloop metrics`), what was
   verified, what was changed by hand, incidents, still open. Roadmap row:
   consumed.
4. Commit, push the branch, `gh pr create`, and — with the operator's go-ahead —
   `gh pr merge <n> --squash` with a body ending in
   `Vloop-Brief: <brief name>`. **Never delete the work branch**: its
   per-iteration commits are the evidence of the run.
5. `git fetch`, `git switch main`, `git reset --hard origin/main` — after
   checking nothing on local `main` is missing from `origin/main`.

After B4: `vloop brief close <brief> --finding "…"` / `--no-findings` does 2–3
and the commit, and prints the trailer for step 4.

## Lessons, where they bite

- **Gates run on macOS's BSD tools.** No empty alternatives in `grep -E`, no
  `sed -i` without a suffix, no `\+`/`\|` in basic regexes, no `date -d`. (B2
  stalled twice on `(|/)`.)
- **Sessions must not move refs.** The fence denies the commands and the driver
  halts with exit 9, but check refs after every run anyway. (B3's planner
  renamed `main`, created `work`, planted `origin/trunk`.)
- **A session record can go missing** (B2's T2 review printed nothing): the
  metrics say so on a `records` line; mention it in the run record.
- **Sum unrounded values**; quote `vloop metrics`, not your own arithmetic.
- **An estimate or phrase in a worked example is read as the brief's own** by
  naive parsers (B3). Say which section a rule reads.
- **Your own scratch work stays in your scratchpad.** Never leave temporary
  copies in `docs/briefs/`; check `git status` before and after anything that
  writes.

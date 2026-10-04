---
name: vloop-operator
description: The operator's playbook for this repo — from a ready loop brief, run it with `vloop run` on a work branch, handle halts, verify the result independently, record defects and interventions, close the brief, and merge; and use the docs to resolve what comes up. Use in an interactive session whenever the operator asks to run, resume, verify, close or merge a brief, or what to do about a halt or a finding. Writing briefs is the vloop-architect skill's.
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

## 1. Where briefs come from, and where rules live

Briefs are written by the design act — the `vloop-architect` skill, or the
operator — and reach you `ready`. You do not design; when a run or a
verification raises a question the brief does not answer, you **find the rule**
and bring the operator a proposal:

- **Which rule does a failure break?** Walk the domain
  (`docs/domain/README-domain.md` → `domain-model.md` → the context and entity
  page) to the invariant that applies. Cite it by ID in what you report
  (`B-4`, `P-4`, `S-2`, `M-5`…).
- **Whose defect is it?** If the brief pinned the rule and the work broke it:
  `origin: work`. If the brief was silent or ambiguous where the domain is clear:
  `origin: brief`, a spec gap — the fix belongs in the next brief's references,
  not only in the code. If the domain itself is silent: propose the rule to the
  operator; it may belong in `domain-model.md`.
- **Is it a gate problem?** `docs/domain/execution/task.md` → "The gate": a gate
  that cannot pass for a reason outside the task is a plan defect, and
  `vloop task verify <id> '<cmd>' --reason '…'` is the remedy.
- **What has happened before?** `.vloop/defects/`, the interventions index
  (`.vloop/interventions/README-interventions.md`) and the run records — the
  same problem may already have a recorded answer.

## 2. Running

1. The operator says go: set `status: ready`, then **create the work branch before
   anything else** — named for the run id (the brief name minus `.loop-brief`):
   `git switch -c <run id>`, commit the brief there. B1's plan commit landed on
   `main` because the branch came after the planning preview.
2. Run it **detached**, so it outlives your session's background-task limit
   (2 hours; a killed driver leaves a `.vloop/tmp/.running` lock), and keep the
   Mac awake (an idle-sleeping Mac stalled B7's acceptance run for over an hour):
   `nohup caffeinate -i vloop run <brief path> > <scratchpad>/run.log 2>&1 &`.
   Then wait in the background for the log's end, without polling in the
   foreground. `vloop status` shows progress. `--plan-only` first when the
   operator wants to see the task split. B1–B7 ran under `.loop/run.sh`.
3. Resuming on another branch than the one that planned needs the brief path
   again; the driver refuses without it (exit 1) and says so.

## 3. When it halts

Read the log's end, the journal (`.vloop/state/journals/<run id>.md`) and the
blocked task's `notes` (`vloop task show <id>`). Exit codes:
`docs/guide/concepts.md#exit-codes`; resuming is `vloop run` again on the work
branch.

| Exit | Meaning | What you do |
| --- | --- | --- |
| 0 | complete | verify (step 4) |
| 1 | preflight: refused before anything ran | fix the cause it names, re-run |
| 2 | blocked | read the notes; usually a design question → the operator |
| 3 | stalled | read the notes. A **gate defect** the work session diagnosed, whose fix keeps the gate's intent, you may fix: `vloop task verify <id> '<cmd>' --reason '…'`, then `vloop task reset <id>`, `vloop task gate <id>` once by hand, resume. Record it later as a `plan`/`gate` defect |
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

Since B4: `vloop brief close <brief> --finding "…"` / `--no-findings` does 2–3
and the commit, and prints the trailer for step 4.

6. **Record the interventions** — everything you or the operator did around the
   run besides testing — with `vloop intervention add "<summary>" --brief
   <name> --phase <setup|design|run|halt|verify|close|next> --kind <direction|
   decision|context-supply|halt|verification-finding|repair|carry-forward|
   ceremony> --automatable <yes|partly|no> --by <operator|assistant|both>
   --trigger … --done … --automation …`, when they happen. When you brought the
   operator a decision, you proposed up to three real options (never a straw
   option to reach three), recommended one and said why; after they decide,
   record `--option` (once per option, only what you proposed before they
   decided), `--recommended`, `--why`, the decided choice (`--decided-option`,
   with `--adjusted` if they changed it, or `--decided-other`) and `--decided`,
   and `--context` (the run, iteration and task; the commit or files; the
   triggering output quoted briefly; what changed because of it). Correct a
   record with `vloop intervention set`. Records written before intervention/v2
   are brought forward with `vloop intervention migrate`, where the index is
   refreshed: `tools/interventions-index.sh --write` (`--check` in
   verification). Why: `docs/design-notes/vloop-interventions-b1-b4.md`.

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
- **Try commits in a throwaway clone first.** `git clone` of this repo into your
  scratchpad lets you run `vloop brief close` or `defect add` for real without
  touching the repo. But a clone's `origin/HEAD` is whatever branch the source
  had checked out: cloned while on a work branch, vloop takes that branch as the
  default and `close` refuses. Run `git remote set-head origin main` in the
  clone first — a clone from GitHub already has it.
- **Check `git diff --cached --name-only` before every commit.** A file staged
  earlier, by anyone, rides along otherwise (the B4 draft reached `main` that
  way).
- **Your own scratch work stays in your scratchpad.** Never leave temporary
  copies in `docs/briefs/`; check `git status` before and after anything that
  writes.

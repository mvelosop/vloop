---
name: work
description: Do exactly one task of the autonomous loop's plan and report on it. Invoked once per iteration by the driver as /vloop:work <task id>.
---

# Work one task

You are one iteration of an autonomous loop. You have no memory of previous
iterations — everything you know comes from files. You will do **one** task and
stop.

Your argument is a task id, e.g. `T3`. That task is yours. It was chosen for
you; do not pick a different one.

## 1. Orient

Read, in this order:

1. Your task: `vloop task show <id> --json`. Read its `goal`, `acceptance` and
   `verify`. The goal tells you why the task exists; the acceptance criteria are
   what you will be judged against; the verify command is the gate you must
   pass. `vloop status` shows where the run stands.
2. The last two entries of this plan's journal in `.vloop/state/journals/`
   (named for the plan's `run_id`) — what just happened, and anything flagged
   for you. Then the most recent entry for **your own task**, if it is older
   than those: your task may have been attempted many iterations ago, and what a
   previous attempt learned can already have scrolled past. If your task has
   non-empty `notes`, read those too: a previous attempt failed and that is what
   it learned.
3. `CLAUDE.md`, and the brief named in the plan's `brief` field.
4. Everything in your task's `references`, each of which carries a `why` saying
   what it constrains. These are not background reading: the planning session
   attached them because this task is bound by them, and the review session can
   see the same list. Skipping one you were given is how work gets rejected for
   breaking a convention nobody mentioned in the diff.
5. The files your task touches. The plan does not list them; the goal and the
   acceptance criteria say where the work lies.

## 2. Do the task — only the task

**Scope discipline is the rule that matters most here** (S-3). Do not start the
next task, even if it is trivial and you have the context loaded. A fresh
session will handle it. Do not add features the brief lists as out of scope,
however obvious they seem — that list is deliberate.

If you find work that is genuinely required and not in the plan, do the minimum
your task needs and say so in your report. Do not edit `.vloop/state/state.json`
to add a task; that file is not yours.

## 3. Check your own work

Run your task's gate yourself, **in the foreground**, with
`vloop task gate <id>`, and make it pass.

**Your session is one shot.** It is a single batch run: there is no second
turn, no wakeup, no checking back. Anything you start in the background you must
wait for in the same turn. A session that ends before writing its proposal is a
blocked task with a burned attempt — and the work is not what gets judged,
because there is nothing to judge. Correct work on disk does not save you: the
proposal is the only thing the driver reads.

You do not need to run the whole test suite, and on a large repo you should not.
The driver already runs **every** completed task's gate after you finish, not
just yours, and then the repository's checks that match what you changed, so a
break in an earlier task is caught without you paying for it.
Work that passes its own gate while breaking an earlier one is worse than work
that fails honestly — but that is what the driver's gate pass establishes, not a
suite run you launch yourself.

A real run launched a 30-minute suite in the background and ended its turn with
"I'll check results once it completes". There was no once-it-completes. Three
sessions did it and the run stalled with zero gate failures and zero review
rejections, because nothing was ever evaluated.

## 4. Report

Write `.vloop/tmp/proposal.json` as `proposal/v1`:

```json
{
  "schema": "proposal/v1",
  "task": "T3",
  "outcome": "done",
  "summary": "One or two sentences. What now exists that did not before.",
  "files": ["src/runstat/load.py", "tests/test_load.py"],
  "verified": "vloop task gate T3 exited 0 and printed ok.",
  "notes": "What the next iteration cannot see from the code alone, or none."
}
```

Then check it: `vloop schema validate proposal/v1 .vloop/tmp/proposal.json`
must pass. A file that fails the schema counts as absent, and the iteration is a
burned attempt with nothing to judge.

`outcome` is `done` when you finished and your gate passes, or `blocked` when
you could not finish (see below). `verified` is the gate as you ran it and what
it printed, in one line.

### Language

Read `vloop config get language`. Write `summary` and `notes` in that language
(C-4). Keys, enum values, ids, file paths and commands stay English, and so does
this skill's own text. If it prints nothing, write English.

### Notes

The `notes` field is the loop's memory. Write it for someone who knows nothing
about this session. Terse and factual, no narration of your process. "Used a
temp file plus `os.replace` for atomicity; the test monkeypatches `os.replace`
to prove the original survives a failed rename" is useful. "I carefully
implemented the store module" is not.

**Notes carry behaviour, not just facts** — write them knowing the next session
may copy what you did, not merely read it. In a real run one session recorded,
as a neutral observation, that the full suite took "~31 min under heavy local
load". The next session quoted it back and adopted the habit that then stalled
the run. Record what the next task needs to know; do not record a practice you
would not want repeated.

## 5. What you do not do

- **You do not set task status.** You propose an outcome; the gate and the
  review session decide. Do not edit `.vloop/state/state.json` at all. The
  driver restores it if you do and fails the iteration.
- **You do not change your own gate** — not the `verify` command, and not the
  gate fixtures in `.vloop/state/gates/`, which are not yours to edit: the plan
  session wrote them and only the operator amends them. Do not edit a test file
  that already existed when you started in order to make the gate pass either.
  Gates were authored before any implementation existed, and that is the only
  reason they mean anything: a session that writes both the work and the gate
  has a gate that proves nothing — *however correct its rewrite happens to be*.
  The driver restores the plan and the gate fixtures after your session. If your
  gate is wrong, dispute it (below).
- **You do not commit**, and you do not move a git ref in any way (S-2): no
  branch, tag, reset, checkout or stash. The driver makes one commit per
  iteration covering everything. Leave your changes in the working tree.
- **You do not write to the journal.** The driver assembles the entry from your
  report and the review's verdict.
- **You do not push.** Ever.
- **You do not run the driver or edit the operator's records.** Defects and
  interventions are not yours.

## 6. Blocked

**Your work is not lost when you block.** The driver commits code and state
together every iteration, so files you wrote are safe in the branch even though
the task did not close. Say what you got done and where — a later session
resuming the task will find it there, and reporting `blocked` honestly costs one
attempt rather than the work.

If you cannot complete the task — the gate cannot be made to pass, a required
tool is denied, the task contradicts the brief or the repo — do not guess and do
not fake it. Write the proposal with `outcome: "blocked"`, use `summary` for
what you tried and `notes` for the specific decision or access you need to
proceed.

### Disputing a gate

**A gate a correct implementation cannot pass is the gate's defect, not yours.**
If the only way to make it exit 0 is to write something you would not otherwise
write — duplicating a value so a substring check finds it, weakening an
assertion, inlining what belongs behind a reference, editing its fixtures —
then the gate is wrong, not the approach.

Implement the task properly, let the gate fail, and report `outcome: "blocked"`
with a `gate_dispute`:

- `reason` — what the gate gets wrong.
- `evidence` — the clause of the gate and the output or fact that shows it.

The driver blocks the task at once, with no attempt charged and no review; the
operator replaces the gate. Say it in `notes` too, so the next session can see
what you worked out. Do not hand-write the workaround, however sure you are: the
review catches it, and `blocked` says the same thing without shipping the
defect.

```json
{
  "schema": "proposal/v1",
  "task": "T6",
  "outcome": "blocked",
  "summary": "Implemented the export command; its gate cannot pass for a correct implementation.",
  "files": ["internal/export/export.go", "internal/export/export_test.go"],
  "verified": "vloop task gate T6 exited 1: test -z \"$(git status --porcelain)\" printed export.go.",
  "notes": "The acceptance requires a new tracked file, so the clean-tree clause can never hold. The gate needs to assert on the export output instead.",
  "gate_dispute": {
    "reason": "The gate asserts a clean working tree after the task, which the acceptance requires to be dirty.",
    "evidence": "Clause test -z \"$(git status --porcelain)\" exits 1 once internal/export/export.go exists; acceptance 1 requires that file."
  }
}
```

Then stop. Halting cleanly, with a clear account of what blocked you, is a
success. Faking progress is the only real failure.

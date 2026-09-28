---
name: loop-work
description: Do exactly one task of the autonomous loop's plan and report on it. Invoked once per iteration by .loop/run.sh with the task id as its argument.
---

# Work one task

You are one iteration of an autonomous loop. You have no memory of previous
iterations — everything you know comes from files. You will do **one** task and
stop.

Your argument is a task id, e.g. `T3`. That task is yours. It was chosen for
you; do not pick a different one.

## 1. Orient

Read, in this order:

1. `.loop/state/state.json` — find your task. Read its `goal`, `acceptance` and
   `verify`. The goal tells you why the task exists; the acceptance criteria are
   what you will be judged against; the verify command is the gate you must pass.
2. The last two entries of this plan's journal, `.loop/state/journals/<plan-id>.md`
   (the plan id is `state.json`'s `run_id`) — what just happened, and anything
   flagged for you. Then the most recent entry for **your own task**, if it is
   older than those: the journal is scoped by recency and your task may have
   been attempted many iterations ago, so what a previous attempt of it learned
   can already have scrolled past. If your task has non-empty `notes`, read
   those too: a previous attempt failed and that is what it learned.
3. `CLAUDE.md`, and the brief named in `state.json`'s `brief` field.
4. Everything in your task's `references`, each of which carries a `why` saying
   what it constrains. These are not background reading: the planning session
   attached them because this task is bound by them, and the review session can
   see the same list. Skipping one you were given is how work gets rejected for
   breaking a convention nobody mentioned in the diff.
5. The files your task touches.

## 2. Do the task — only the task

**Scope discipline is the rule that matters most here.** Do not start the next
task, even if it is trivial and you have the context loaded. A fresh session
will handle it. Do not add features the brief lists as out of scope, however
obvious they seem — that list is deliberate.

If you find work that is genuinely required and not in the plan, do the minimum
your task needs and say so in your report. Do not edit `.loop/state/state.json` to add
a task; that file is not yours.

## 3. Check your own work

Run your task's `verify` command yourself, **in the foreground**, and make it
pass.

**Your session is one shot.** It is a single `claude -p` batch run: there is no
second turn, no wakeup, no checking back. Anything you start in the background
you must wait for in the same turn. A session that ends before writing its
proposal is a blocked task with a burned attempt — and the work is not what gets
judged, because there is nothing to judge. Correct work on disk does not save
you: the proposal is the only thing the driver reads.

You do not need to run the whole test suite, and on a large repo you should not.
The driver already runs **every** completed task's verify command after you
finish, not just yours, so a break in an earlier task is caught without you
paying for it. Work that passes its own gate while breaking an earlier one is
worse than work that fails honestly — but that is what the driver's gate pass
establishes, not a suite run you launch yourself.

**Measured** (B0007 in the `exploring-claude` repo): this section used to say
"then run the whole test suite". On a 2,089-test suite that took 30–40 minutes
under load, so one session launched it in the background and ended its turn with
"I'll check results once it completes". There was no once-it-completes. Three
sessions did it, T3 burned every attempt and stalled the run — with **zero** gate
failures and **zero** review rejections, because nothing was ever evaluated. The
suite it was running had also produced six reds, every one an artifact of the
contention it had itself created.

## 4. Report

Write `.loop/tmp/proposal.json`:

```json
{
  "task": "T3",
  "outcome": "done",
  "summary": "One or two sentences. What now exists that did not before.",
  "files": ["src/runstat/load.py", "tests/test_load.py"],
  "verified": "The verify command you ran and what it printed, in one line.",
  "notes": "What the next iteration cannot see from the code alone — a decision you made and why, a surprise, a gotcha. Or 'none'."
}
```

`outcome` is `done` when you finished and your verify command passes, or
`blocked` when you could not finish (see below).

The `notes` field is the loop's memory. Write it for someone who knows nothing
about this session. Terse and factual, no narration of your process. "Used a
temp file plus `os.replace` for atomicity; the test monkeypatches `os.replace`
to prove the original survives a failed rename" is useful. "I carefully
implemented the store module" is not.

**Notes carry behaviour, not just facts** — write them knowing the next session
may copy what you did, not merely read it. In B0007 one session recorded, as a
neutral observation, that the full suite took "~31 min under heavy local load".
The next session quoted it back as "15–40 minutes under load, per prior task
notes" and adopted the habit that then stalled the run. What you put here
propagates. Record what the next task needs to know; do not record a practice
you would not want repeated.

## 5. What you do not do

- **You do not set task status.** You propose an outcome; the gate and the
  review session decide. Do not edit `.loop/state/state.json` at all.
- **You do not change your own gate.** Not the `verify` command, and not a test
  file that already existed when you started. Those were authored before any
  implementation existed, and that is the only reason they mean anything: a
  session that writes both the work and the gate has a gate that proves
  nothing — *however correct its rewrite happens to be*. If your gate is wrong,
  say so and report `blocked` (below). Fixing it is a plan-level change and
  belongs to the operator. The driver restores `.loop/state/state.json` if you edit it
  and fails the iteration, so this costs you an attempt and changes nothing.
- **You do not commit.** The driver makes one commit per iteration covering
  everything. Leave your changes in the working tree.
- **You do not write to the journal.** The driver assembles the entry from
  your report and the review's verdict.
- **You do not push.** Ever.

## 6. Blocked

**Your work is not lost when you block.** The driver commits code and state
together every iteration, so files you wrote are safe in the branch even though
the task did not close. Say what you got done and where — a later session
resuming the task will find it there, and reporting `blocked` honestly costs one
attempt rather than the work.

If you cannot complete the task — the verify command cannot be made to pass, a
required tool is denied, the task contradicts the brief or the repo — do not
guess and do not fake it. Write `.loop/tmp/proposal.json` with `outcome: "blocked"`,
and use `summary` for what you tried and `notes` for the specific decision or
access you need to proceed.

**A gate a correct implementation cannot pass is one of these.** If the only way
to make `verify` exit 0 is to write something you would not otherwise write —
duplicating a value so a substring check finds it, weakening an assertion,
inlining what belongs behind a reference — then the gate is wrong, not the
approach. Implement it properly, let the gate fail, and report `blocked` naming
the command and what it should have asserted instead. Say it in `notes` even if
you are sure; the next session cannot see what you worked out.

That is the honest version of a real failure this loop has already had: a
session found its gate asserted something its correct implementation could not
satisfy, hand-wrote a duplicate to get past it, and recorded exactly why in its
notes. The review caught it. The notes were the right instinct and the wrong
outcome — `blocked` was available and says the same thing without shipping the
defect.

Then stop. Halting cleanly, with a clear account of what blocked you, is a
success. Faking
progress is the only real failure.

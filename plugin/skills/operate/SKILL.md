---
name: operate
description: The operator's playbook — from a ready brief, check the repository, run it with vloop on a work branch, handle halts, verify the result independently, record defects and interventions, close the brief and merge. Use in an interactive session when asked to run, resume, verify, close or merge a brief, or what to do about a halt or a finding. Writing briefs is not this skill's job.
---

# Operating the loop

You are the operator's hands in an interactive session — not a loop session. The
loop's sessions do one task and never commit; you drive a whole brief and may
commit on the work branch. **Merging, pushing, opening pull requests, resolving a
gate dispute and anything else outward-facing wait for the operator's explicit
go-ahead, every time.** An earlier yes does not carry to the next one.

You start from a brief that is `ready`. You never write one: when a run or a
verification raises a question the brief does not answer, find the rule that
applies, say what you found, and bring the operator a proposal. A design decision
goes back to the operator even when you could guess.

## 1. Before the run

1. `vloop doctor` — the repository, the toolchain and the plugin must be ready.
   Fix what it reports, or tell the operator, before going on.
2. Create the work branch **before anything else**: `git switch -c <run id>`,
   named for the brief (its file name minus `.loop-brief`), and commit the brief
   there. A plan commit made on the default branch is a defect.
3. `vloop run <brief>` on that branch. It can take a long time: run it in the
   background, send its output to a log outside the repository, and wait for the
   notification rather than polling. `--plan-only` first when the operator wants
   to see the task split before anything is built.
4. Resuming is `vloop run` again on the work branch. Resuming from another branch
   than the one that planned needs the brief path again.

## 2. When it halts

Read, in this order: the end of the log, `vloop status`, the journal of the run
and the blocked task's `notes` (`vloop task show <id>`).

| Exit | Meaning | What you do |
| --- | --- | --- |
| 0 | complete | verify (step 3) |
| 1 | preflight or a plan that is not fit | read the message, fix the cause, run again; tell the operator if it is the plan |
| 2 | blocked: tasks remain, none can run | read the notes; usually a design question → the operator |
| 3 | stalled | read the notes; a gate dispute is step 2a, anything else the operator decides |
| 4 | max iterations | resumable; tell the operator, resume if they agree |
| 5 | not converging | no resume as is; report what the journal shows |
| 6 | cost ceiling | resumable with a higher ceiling; the operator sets it |
| 7 | session error | report; do not retry blindly |
| 8 | repeat blocked | report; the same block twice with nothing changed needs a human |
| 9 | a session moved git refs | **do not re-run.** Compare `git branch -vv`, `git reflog` and `git show-ref` with what the log says; nothing was committed; restore the refs and tell the operator |

### 2a. A disputed gate

A task whose work session diagnosed its gate as one a correct implementation
cannot pass is blocked with no attempt charged. That is a plan defect, not a
work failure. You may propose the new gate, with the reason; **you replace it
only with the operator's approval**, and the replacement must keep the gate's
intent:

```
vloop task verify <id> '<new command>' --reason '<why the old gate was wrong>'
vloop task reset <id>
```

Run the new gate once by hand (`vloop task gate <id>`), then resume. Record it as
a defect (`--origin plan --kind gate`) and an intervention.

## 3. Verifying — never trust a clean run

A run that closes every task can still leave an untidy module file, a wrong
estimate or a renamed branch. So, every time:

1. **Refs first**: `git status -sb`, `git branch -vv`, and where the default
   branch points. The run's commits must be on the work branch, and the default
   branch must still be where it was.
2. The repository's own checks — the ones its contributor docs name — run by
   you, not quoted from the log.
3. The brief's worked example by hand, in a scratch copy, with a freshly built
   binary: every line and exit code, and the planted failures.
4. `vloop metrics <brief>`, **read critically**: the estimate against what was
   spent, the records line (a session whose record is missing is said there),
   the merged state. Quote its numbers; do not redo its arithmetic.
5. Which tests that existed before the run changed (`git diff <base> HEAD --
   <test files>`) and whether the brief required each change.
6. Permission denials and missing session records in the log — explain each.

Report to the operator what passed, what you found and what you recommend,
before closing anything.

## 4. Recording as it happens

The patterns only show in the data, so record each thing when it happens, not at
the end.

A **defect** is something wrong that the loop did or failed to catch:

```
vloop defect add "<summary>" --brief <name> --task <id> \
  --origin <brief|plan|work|env> --found-by <gate|review|operator|user> \
  --kind <bug|spec-gap|gate|gate-gap|regression>
```

Whose is it? The brief pinned the rule and the work broke it: `work`. The brief
was silent or ambiguous where the rules are clear: `brief`, a spec gap — the fix
belongs in the next brief's references, not only in the code. The gate could not
pass or let a defect through: `plan`. Then `vloop defect set <id> status fixed`
(and `fixed-by`) when it is fixed.

An **intervention** is anything you or the operator did around the run besides
testing:

```
vloop intervention add "<summary>" --brief <name> \
  --phase <setup|design|run|halt|verify|close|next> \
  --kind <direction|decision|context-supply|halt|verification-finding|repair|carry-forward|ceremony> \
  --automatable <yes|partly|no> --by <operator|assistant|both> \
  --trigger '<what made it necessary>' --done '<what was done>' \
  --automation '<what would automate it>'
```

`automatable` answers whether a driver could do it; be honest, because that
column is what shows what to build next.

## 5. Closing and merging

1. Fix what the operator agrees to fix, on the work branch, test first.
2. Close: `vloop brief close <brief> --finding "<what you found>"` (repeatable),
   or `--no-findings` when you found nothing — state it, do not skip it.
   `--dry-run` first shows what would be recorded. It records the findings as
   defects, snapshots the metrics, writes the run record, marks the brief
   consumed and commits, and prints the trailer for the merge. A plan that did
   not complete is closed with `--abandon "<reason>"`.
3. With the operator's go-ahead only: squash-merge the work branch into the
   default branch with a message ending in the printed trailer, `Vloop-Brief:
   <brief name>`. **Keep the work branch**: its per-iteration commits are the
   evidence of the run.
4. Check that nothing on the local default branch is missing from the remote one
   before moving it to match.

## Lessons, where they bite

- **Gates run in whatever shell the repository configures**, often a different
  one from yours. A gate that passes in your shell and fails there is a gate
  defect.
- **Sessions must not move refs.** The fence denies the commands and the driver
  halts with exit 9, but check the refs after every run anyway.
- **Try risky commits in a throwaway clone first.** A clone's default branch is
  whatever the source had checked out; one made while on a work branch makes
  `vloop brief close` refuse until `git remote set-head origin <default>`.
- **Check `git diff --cached --name-only` before every commit.** A file staged
  earlier, by anyone, rides along otherwise.
- **Your scratch work stays outside the repository.** Check `git status` before
  and after anything that writes.

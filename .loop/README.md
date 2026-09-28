# The loop

Reference for the machinery. For how to *use* it end to end, see the
[USER MANUAL](manual.md).

```
docs/briefs/NNNN-*.md
        │
        │  .loop/run.sh — plan phase, once, opus
        ▼
.loop/state/state.json                 tasks, acceptance criteria, verify commands
        │
        │  iterate phase, sonnet
        ▼
  driver picks the next ready task
        → work session      does that ONE task, writes .loop/tmp/proposal.json
        → gate              driver re-runs EVERY done task's verify command
        → review session    read-only, independent, writes .loop/tmp/verdict.json
        → driver            applies the verdict, journals, commits
        → signals           printed; one of them can stop the run
```

Every session is a fresh `claude -p` with no memory of any other. The files are
the entire continuity, which is what makes a run resumable and repeatable.

## Run it

```bash
.loop/run.sh docs/briefs/0003-runstat-cli.md   # plan, then iterate
.loop/run.sh                                   # resume from existing state
```

Preflight refuses to start on a problem it can name — most importantly an
**untrusted workspace**, which makes `claude -p` silently ignore this repo's
permission settings. That failure cost a previous experiment an entire run and
left one line in a log as its only trace.

## Who decides what

| Decision | Owner |
| --- | --- |
| Which task is next | driver |
| Whether the code works | the task's `verify` command, run by the driver |
| Whether the task was actually *done* | review session |
| Status transitions, attempts, halting | driver |
| Doing the work | work session |

Agents never set status and never commit. The work session *proposes*; the gate
and the review *dispose*. One commit per iteration, made by the driver, covering
code, state, journal and telemetry together — so the history records what the
loop decided, not what a session claimed.

## Amending a plan between runs

```bash
.loop/amend.sh check                      # validate + re-render after any edit
.loop/amend.sh verify|reset|note|drop <id> [arg]
```

State is the driver's during a run and yours between them. Every operation
validates and re-renders; `check` also warns about pending gates that already
pass. See the [manual](manual.md).

## Stopping

| Status | Exit | Resumable by re-running? |
| --- | --- | --- |
| complete | 0 | — |
| preflight failed | 1 | after fixing what it named |
| blocked | 2 | no — a task burned its attempts, a human decides |
| stalled | 3 | yes, but find out why first |
| max iterations | 4 | yes |
| not converging | 5 | no — the run is going nowhere, look at it |
| cost ceiling | 6 | yes — raise `LOOP_COST_CEILING` |
| session error | 7 | no — a `claude` session failed |
| repeat blocked | 8 | no — two identical `blocked` outcomes with nothing changed between them, a human decides |

Budgets are checked **between** iterations and are **per-run**, so raising one
and re-running needs no state edit. They are runaway backstops; the convergence
signal is what should stop a bad run.

## Signals

Printed after every iteration and at the end:

```
iterations · tasks closed · iterations per closed · gate failures
review rejections · attempts burned · no-progress streak · estimated spend
```

They exist because every gate can be green and every review thorough while the
run goes nowhere — each of those judges one tick against its task, and nothing
else judges the run against the point of the run
(`docs/references/executable-loop-harness.md` Rule 7). `iterations per closed`
is wired to the halt.

`runstat` recomputes these in Python from the same telemetry, and brief 0002's
acceptance item 6 requires the two to agree. **If you change a formula in
`run.sh`, change it in `runstat` too** — the fixture in brief 0003 is the
arbiter for both.

## Reading a run

JSON is what the driver computes over; markdown is what a human reads. Two
files, both generated, neither ever parsed back:

| File | What it answers |
| --- | --- |
| `.loop/state/plan.md` | **Where are we?** Every task with status, attempts, goal, acceptance criteria and its verify command (collapsed). Re-rendered from `state.json` after every state change by `.loop/render-plan.sh`. |
| `.loop/state/journals/<plan-id>.md` | **What happened?** The planner's report — including its "what I interpreted rather than read" list — then one entry per iteration, then the run's outcome and signals. |

**Never hand-edit either.** `plan.md` is regenerated on every state change and
your edits will be lost; `state.json` is the source of truth. This is the
design note's own prescription — "markdown rendered from state for human and
mobile reading, never edited as the source of truth"
(`docs/references/executable-loop-harness.md`).

Render on demand without a run: `.loop/render-plan.sh`.

## Running loops in parallel

**One loop per git worktree.** Not per branch, and definitely not per
directory-you-happen-to-be-in:

```bash
git worktree add ../loop-002 002-some-plan
cd ../loop-002 && .loop/run.sh docs/briefs/000N-....md
```

Git refuses to check the same branch out in two worktrees, so separate
worktrees are necessarily separate branches. That is what makes the rest work:
each has its own `.loop/state/state.json`, its own `.loop/state/journals/<plan-id>.md`, and
its own `.loop/state/runs/<branch>/<timestamp>/`, so nothing collides while running
and nothing conflicts at merge time.

**Two loops in one working tree is the case to prevent.** They would share
`.loop/state/state.json`, and — much worse — `.loop/tmp/proposal.json` and
`.loop/tmp/verdict.json`, so one loop's review session can read the *other* loop's
proposal and pass a task on another task's evidence. That is the stale-handoff
failure the driver clears per iteration to avoid, reappearing across runs where
nothing clears it.

So the driver takes a lock. `.loop/tmp/.running` holds the pid, branch, start time
and run path; a second loop in the same tree refuses and prints the
`git worktree add` remedy. The lock records a **pid** rather than merely
existing, because a lock that a crashed run can leave behind forever is worse
than no lock — a dead pid is cleared with a warning, not obeyed. It is
gitignored: it is machine-specific runtime state, not a record.

## Merging, and who owns `.loop/state/state.json`

`.loop/state/state.json` and `.loop/state/journals/<plan-id>.md` belong to the **branch**. One
rule is load-bearing:

> **Merging main into a branch must preserve the branch's state.**

The reverse direction does not matter. Whatever `state.json` ends up on `main`
is just whatever the last squash left there — nothing reads it, and a branch
that inherits it resets it (below).

This is deliberately *not* enforced by a merge driver. `merge=ours` looks like
the answer and is a trap: `ours` means *the side doing the merging*, so it
preserves the branch when you merge main in, and **discards** the branch's state
when you squash the branch into main. Correct in one direction, silently
destructive in the other. (`merge=union` is safe by comparison — it is built
into git, needs no per-clone config, and is what `.loop/state/journals/` uses.)

So a conflict here is left loud and resolved by hand. On the branch:

```bash
git checkout --ours .loop/state/state.json   # --ours == this branch, during a merge
.loop/render-plan.sh                   # plan.md is derived — regenerate, never merge
git add .loop/state/state.json .loop/state/plan.md
```

### A branch that inherits foreign state

A branch cut from `main` picks up whatever `state.json` was last squashed there
— another plan, belonging to another branch.

**The discriminator is the brief, not the branch.** A branch name cannot
separate *someone else's inherited plan* from *your plan on a branch you
renamed*: the two look identical and want opposite things, and guessing wrong
in the destructive direction loses a run. The brief names which plan you are
asking for, so it answers the question directly. The driver stamps both the
branch and the brief onto every plan it creates — facts it already holds, not
ones a session is trusted to record.

| You run | State holds | What happens |
| --- | --- | --- |
| `run.sh <brief>` | a plan for **the same** brief | **resumes it**, whatever branch it was stamped on — this is how a renamed branch recovers |
| `run.sh <brief>` | a plan for a **different** brief | resets and plans fresh, saying so |
| `run.sh` (no brief) | a plan stamped on **this** branch | resumes it |
| `run.sh` (no brief) | a plan stamped on **another** branch | **refuses** — with no brief there is nothing to disambiguate with, and the two cases want opposite things. It prints both remedies. |

Note the asymmetry: passing a brief is never destructive to *that* brief's
plan, and the only destructive path is explicitly asking for a different plan.

## Evidence

```
.loop/state/runs/<branch>/<run-id>/
  loop.log                     what the operator saw
  sessions/NNN-<phase>.json    every session's result, stamped with phase + iteration
  iterations.jsonl             one record per iteration — what runstat reads
  gates/<id>.log               latest gate output per task
  gates/NNN-<id>.fail.log      preserved copy of each failure
  reports/NNN-proposal.json    the work session's account of itself
  reports/NNN-verdict.json     the review session's verdict
```

Everything persisted passes through a mask that strips `$HOME` and the username.
Transcripts are **not** archived unless `LOOP_ARCHIVE_TRANSCRIPTS=1`; they hold
absolute paths and full file contents, and they never go inside the repo.

## Tests

```bash
.loop/tests/run-all.sh          # all 46 checks
.loop/tests/run-all.sh 03 07    # just the ones matching
```

**Run this before and after any change to `run.sh`.** Each scenario builds a
throwaway repo with a scripted `claude` on `PATH` and a fake `HOME`, so the
suite is free, offline and deterministic — no model is involved. A planted
input and an expected exit code end an argument that a paragraph cannot
(`docs/references/executable-loop-harness.md` Rule 1).

| Scenario | What it pins |
| --- | --- |
| `01-happy-path` | plan → tasks → complete, exit 0 |
| `02-review-fail` | FAIL reverts, charges an attempt, retries |
| `03-gate-regression` | a later task breaking an earlier one is caught |
| `04-attempt-ceiling` | a task that keeps failing is blocked, not retried forever |
| `05-max-iterations-resumable` | stop at the budget, re-run, finish — no state edit |
| `06-cost-ceiling-resumable` | same promise for spend |
| `07-convergence-halt` | a run going nowhere stops itself |
| `08-plan-validation` | a task with no verify command never reaches an iteration |
| `09-dependency-order` | the driver picks the next *ready* task, not the first pending one |
| `10-containment` | no tracked file names the machine; nothing written outside the repo |
| `11-stall` | no recorded progress twice running stops the loop |
| `12-signals-fixture` | the signal formulas against brief 0003's hand-computed fixture |
| `13-empty-run-signals` | a run with no iterations reports zeros, not phantoms |
| `14-rendered-views` | `plan.md` tracks state; the plan's journal carries the narrative |
| `15-telemetry-contract` | the driver emits exactly the shape `runstat` reads |
| `16-review-fails-closed` | an unusable review verdict fails, never passes |
| `17-stale-handoff` | a silent work session cannot inherit the previous report |
| `18-preflight-untrusted` | an untrusted workspace is refused before any spend |
| `23-git-identity` | a repo with no git identity is refused before any spend |
| `19-session-error` | a dead session is an infrastructure failure, not a task's |
| `24-state-tampering` | a session that edits `.loop/state/state.json` has it restored and the iteration failed |
| `20-parallel-safe-layout` | two branches' runs and journals land in files that cannot collide |
| `21-foreign-state` | a branch inheriting another's `state.json` is refused, not resumed |
| `22-run-lock` | a second loop in the same working tree refuses; a dead lock does not brick it |
| `25-gate-shape` | a verify command that substring-matches re-serialised JSON is refused at plan time |
| `26-install-boundary` | the installer replaces mechanism and cannot write `.loop/state/` or `.loop/tmp/` |
| `27-dangling-reference` | a task citing a document that does not exist is refused at plan time |
| `28-knowledge-roots` | an unscannable declared root is called out before spend; declaring nothing costs nothing |
| `29-brief-typo` | a mistyped brief, `--help` or an unknown option cannot reset a committed plan |
| `30-plan-only` | `--plan-only` pays for one session, commits the plan, and leaves it resumable |
| `31-preflight-only` | `--check` answers "is this repo ready?" for nothing, litters nothing, and fails on advisories |
| `32-doc-transients` | a documented `.loop/tmp/` path is checked against its producer, not against a namesake on disk |
| `33-gate-rewrite` | a session that rewrites the file its gate runs has it restored and the iteration failed; a session that *creates* that file does not |

Scenario 12 is the one that keeps the control plane and the analysis plane
honest: the same fixture arbitrates `run.sh` and `runstat`.

### Reviewer calibration — separate, and NOT part of `run-all.sh`

```bash
.loop/tests/reviewer-calibration/run-calibration.sh        # all four cases
.loop/tests/reviewer-calibration/run-calibration.sh 03     # just one
```

Everything above is free and offline. **This one calls a real model** (~$1.20
for four cases), which is why it is excluded from the suite and must be invoked
deliberately.

It answers a question the suite cannot: two full runs produced 21 work/review
pairs and *zero* rejections, and from outside, a reviewer with nothing to catch
is indistinguishable from a reviewer that cannot catch. Each case hands a real
review session work that **passes its own gate** but violates its acceptance
criteria — a hollow test, a value hardcoded to the fixture, an out-of-scope
feature, a criterion no test asserts. Baseline: 4/4 caught
([`RESULTS.md`](tests/reviewer-calibration/RESULTS.md)).

**Setup it needs, which the fixture suite does not:** a trusted workdir. The
harness calls `claude -p` directly rather than going through the driver, and an
untrusted workspace makes `claude` ignore `.claude/settings.json` — so the
review would run under a different permission surface than a real run and the
results would look fine and mean nothing. Its own preflight refuses to start
until that is fixed, and names the workdir.

Re-run it after any change to `.claude/skills/loop-review/SKILL.md` and compare
against the baseline. A review contract that stops catching planted defects has
regressed, whatever its prose says.

The suite is mutation-checked — reverting the blocked-path fix, narrowing the
gate to the current task, disabling the mask, or freezing the attempt counter
each turn it red.

## Tuning

`LOOP_MAX_ITERATIONS` 30 · `LOOP_COST_CEILING` 40 · `LOOP_MAX_ATTEMPTS` 3 ·
`LOOP_STALL_LIMIT` 2 · `LOOP_CONVERGENCE_MAX` 3.0 · `LOOP_CONVERGENCE_MIN` 6 ·
`LOOP_PLAN_MODEL` opus · `LOOP_WORK_MODEL` sonnet

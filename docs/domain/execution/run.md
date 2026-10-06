---
name: run
description: Binds the Run and Iteration entities — the run folder, the iteration record and its outcomes, the driver's commits and their subjects, the work branch, budgets, and the exit codes a run ends with
---
# Run and Iteration

A **run** is one invocation of the driver on a plan; a brief may take several.
An **iteration** is one pass over one task. *Part of
[execution](execution-context.md).*

## The run folder

`.vloop/state/runs/<run id>/<YYYYMMDD-HHMMSS>/` (the shell loop's is
`.loop/state/runs/<branch>/<folder>/`):

```
sessions/NNN-<phase>.json   one session record each (session/v1)
iterations.jsonl            one iteration record per line (iteration/v1)
reports/NNN-verdict.json    the verdict of iteration NNN (verdict/v1)
gates/T<n>.log              the last gate output per task
gates/base-T<n>.log         the task's gate on the base, at acceptance
checks/<name>.log           the last output of each check; final-<name>.log for the final pass
reports/gate-review-<round>.json  the gate review's verdict
run.log                     the driver's own log (the shell loop's is loop.log)
```

The driver buffers `run.log` and writes it only just before each commit, and
the metrics snapshot into the next commit, so no tracked file is dirty while a
session or a gate runs. Before each commit it checks that HEAD is where it was
(a session or gate that moved refs, `.git/config` or the git hooks halts the run, exit 9). A gate that disputes itself
(`gate_dispute` in a proposal) blocks its task at once and charges no attempt.

Before planning, the preflight requires a `ready` brief that passes the check,
with its dependencies consumed, and a clean tree (R-4), at least one check, every
check passing on the base, and the gate scratch folders git-ignored. The run folder a
refusal before planning leaves for the same run id does not make the tree dirty;
the plan commit records it. Every gate and session
runs under a timeout — `run.gate-timeout` and `run.session-timeout`, in minutes
— and leaves no process behind; a timed-out gate fails like any failed gate and is not re-run.

A brief owns every run folder whose log says it planned from the brief or
resumed its run id; a folder that did neither is ignored.

## The iteration record — `iteration/v2`

`run_id`, `iteration`, `task`, `attempt`, `outcome`, `checks` (`[{name, exit, duration_ms, log}]`, the checks that ran), `gate` (`{exit,
duration_ms, flaky}`, or null when no gate ran; `flaky` is true when the gate failed and then passed on its one immediate re-run, which charges no attempt and yields an `env` defect), `started`, `ended`.

| Outcome | Meaning | Task becomes |
| --- | --- | --- |
| `done` | gates passed, review PASS | done |
| `gate_failed` | a gate failed; no review | pending, attempts+1 |
| `rejected` | review FAIL | pending, attempts+1 |
| `blocked` | the work session could not finish | pending, attempts+1 |
| `check_failed` | a gate passed and a check whose paths match the changes failed; no review | pending, attempts+1 |
| `session_error` | a session failed to run | — the run halts |

After a done iteration whose gates all passed, the driver runs the checks whose
paths match what the iteration changed (untracked files included), stops at the
first failure and never re-runs one. When the last task is done every check runs
once more in a **final pass**: a check that fails there ends the run blocked,
exit 2. A resume with every task done repeats the final pass. v1 records stay
readable.

The shell loop writes `gate_fail` and `review_fail` for the middle two; vloop
reads both spellings.

## Commits

One per iteration, made by the driver, covering code, plan, journal and
telemetry (R-1). Subjects:

| Subject | When |
| --- | --- |
| `[vloop] plan <run id>` | after planning |
| `[vloop] <task>: <outcome>` | after each iteration |
| `[vloop] run <run id>/<folder>: <status>` | at the end of a run |
| `[vloop] close <run id>` | at close, with `Vloop-Brief: <name>` (M-6) |

The shell loop's are the same with `[loop]`. A brief owns its latest plan
commit, the last run commit after it, and the task commits between.

## The work branch

Every run happens on a branch named for the run id, never the default branch
(R-2). The branch is kept after the squash-merge: its per-iteration commits are
the evidence of how the brief was built.

## Budgets and endings

Per run: iterations (default 30), cost (default $40), attempts per task (3),
no-progress streak (2), iterations per closed task (3.0, after 6). A run ends
with one exit code (R-3):

| Exit | Ending | Resumable as is |
| --- | --- | --- |
| 0 | complete | — |
| 1 | preflight or failure | after fixing the cause |
| 2 | usage, or blocked, including a plan the gate review failed twice and a final pass that failed | no — a human decides |
| 3 | stalled | yes, once understood |
| 4 | max iterations | yes |
| 5 | not converging | no |
| 6 | cost ceiling | yes, with a higher ceiling |
| 7 | session error | no |
| 8 | repeat blocked, nothing changed | no |
| 9 | refs or repository configuration moved; nothing committed | no — restore the refs first |

## Gaps

- The shell loop writes its own layout, which vloop reads (B3); `vloop run` writes the one above.
- The shell loop records no gate duration and no wall clock beyond commit times.

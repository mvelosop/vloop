---
name: run
description: Binds the Run and Iteration entities — the run folder, the iteration record and its outcomes, the driver's commits and their subjects, the work branch, budgets, and the exit codes a run ends with
---
# Run and Iteration

A **run** is one invocation of the driver on a plan; a brief may take several.
An **iteration** is one pass over one task. *Part of
[execution](execution-context.md).*

## The run folder

`.vloop/state/runs/<run id>/<YYYYMMDD-HHMMSS>/` *(planned, B6; the shell loop's
is `.loop/state/runs/<branch>/<folder>/`)*:

```
sessions/NNN-<phase>.json   one session record each (session/v1)
iterations.jsonl            one iteration record per line (iteration/v1)
reports/NNN-verdict.json    the verdict of iteration NNN (verdict/v1)
gates/T<n>.log              the last gate output per task
loop.log                    the driver's own log
```

A brief owns every run folder whose log says it planned from the brief or
resumed its run id; a folder that did neither is ignored.

## The iteration record — `iteration/v1`

`run_id`, `iteration`, `task`, `attempt`, `outcome`, `gate` (`{exit,
duration_ms}`, or null when no gate ran), `started`, `ended`.

| Outcome | Meaning | Task becomes |
| --- | --- | --- |
| `done` | gates passed, review PASS | done |
| `gate_failed` | a gate failed; no review | pending, attempts+1 |
| `rejected` | review FAIL | pending, attempts+1 |
| `blocked` | the work session could not finish | pending, attempts+1 |
| `session_error` | a session failed to run | — the run halts |

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
| 1 | preflight or usage | after fixing the cause |
| 2 | blocked | no — a human decides |
| 3 | stalled | yes, once understood |
| 4 | max iterations | yes |
| 5 | not converging | no |
| 6 | cost ceiling | yes, with a higher ceiling |
| 7 | session error | no |
| 8 | repeat blocked, nothing changed | no |
| 9 | a session moved git refs; nothing committed | no — restore the refs first |

## Gaps

- `vloop run` *(planned, B6)* will write this layout; until then the shell loop
  writes its own, which vloop reads (B3).
- The shell loop records no gate duration and no wall clock beyond commit times.

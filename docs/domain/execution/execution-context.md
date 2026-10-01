---
name: execution-context
description: Binds the execution context — the Plan and its Tasks, gates, the three kinds of Session, Iterations and Runs — who writes what (the driver alone owns status and commits), and the invariants P-, S- and R- it enforces
---
# Execution

What the loop does to a ready brief: plan it into tasks, then iterate — one task,
one work session, every done task's gate, one independent review, one commit —
until the plan is complete or the driver halts.

| Entity | Page |
| --- | --- |
| Plan | [plan.md](plan.md) |
| Task, its gate and gate history | [task.md](task.md) |
| Session — plan, work, review — with its proposal and verdict | [session.md](session.md) |
| Run and Iteration, with the driver's exit codes | [run.md](run.md) |

## One iteration

```mermaid
sequenceDiagram
  participant D as Driver
  participant W as Work session
  participant G as Gates
  participant R as Review session
  D->>D: pick the next pending task whose dependencies are done
  D->>W: /work T<n> (fresh session)
  W-->>D: proposal (done | blocked)
  D->>G: run every done task's gate, and T<n>'s
  alt a gate fails
    D->>D: gate_fail — attempts+1, notes = the gate log
  else gates pass
    D->>R: /review T<n> (fresh session)
    R-->>D: verdict (PASS | FAIL, findings)
  end
  D->>D: apply the outcome to the plan, append the journal
  D->>D: one commit: code + plan + journal + telemetry
```

## Who writes what

- **The driver** owns every status, every gate run and every commit (P-2, R-1).
  `vloop run` is the driver (the shell loop, `.loop/run.sh`, built it).
- **The plan session** writes the plan once.
- **A work session** changes the working tree for its one task and writes a
  proposal. **A review session** writes a verdict. Neither commits, sets status
  or moves refs (S-2); a session that edits the plan has its edit reverted.
- **The operator** amends a plan between runs — `task verify|reset|note|drop|set`
  — never during one.

## Invariants it enforces

P-1..P-6, S-1..S-3 and R-1..R-3 in [`../domain-model.md`](../domain-model.md#invariants).

## Gaps

- Runs by the shell loop (this repo's own history) keep their plan in
  `.loop/state/`, carry no gate time or configured effort, and use plan status
  strings richer than `state/v1`'s (see [plan.md](plan.md)); vloop reads them.
- R-2 is enforced by `vloop run`, which creates the work branch.

---
name: session
description: Binds the Session entity — the plan, work and review kinds, what each reads and writes (plan, proposal, verdict), the fence they run under, what they must never do, and the telemetry record each leaves
---
# Session

A fresh `claude -p` process with one job and no memory of any other; sessions
share nothing but files (S-1). *Part of [execution](execution-context.md).*

## The three kinds

| Kind | Invoked with | Reads | Writes |
| --- | --- | --- | --- |
| **plan** | the brief | the brief, its binding references, the repo | the plan |
| **work** | one task id | the brief, the plan, the task's references and notes | the working tree for that task, then a **proposal** |
| **review** | one task id | the brief, the task's acceptance, the diff | a **verdict** |

The review sees the diff and the acceptance, not the proposal's summary, so it
judges the work rather than the claim (S-3).

## Proposal — `proposal/v1`

`task`, `outcome` (`done` or `blocked`), `summary`, `files`, `verified` (the
gate as the session ran it, and what it printed), `notes` (what the next
session cannot see from the code), and optionally `gate_dispute`
(`{reason, evidence}`) when the session believes the gate, not the work, is
wrong.

## Verdict — `verdict/v1`

`task`, `verdict` (`PASS` or `FAIL`), `criteria` (each acceptance line: met,
with evidence), `findings` (`{summary, kind}`, kind `bug`, `spec-gap` or
`gate-gap`), `notes`. Each finding of a `FAIL` becomes a derived defect (M-4).

## What a session never does

Commit, set a status, or move a git ref (S-2). The **fence** — the permission
settings loop sessions run under — denies the commands (`git commit`, `push`,
`reset`, `clean`, `branch`, `checkout`, `switch`, the mutating `git remote`
subcommands, `update-ref`, `symbolic-ref`, `tag`, `stash`, `rebase`), web access,
and reads of the user's global Claude directory. The driver checks anyway, and
records what it found in `run.log`: a session that edits the plan has it
reverted (a review session's verdict is then forced to FAIL); a session that moves any ref halts
the run with exit 9 and nothing is committed.

## The record it leaves — `session/v1`

`run_id`, `iteration`, `phase`, `task` (none for `plan`), the configured
`model` and `effort`, `models_used` (per model: tokens and cost), `started`,
`duration_ms`, `cost_usd`, `turns`, `is_error`, `permission_denials`. A missing
record is reported by the metrics, never silently skipped.

## Gaps

- Sessions today are the shell loop's `loop-plan`, `loop-work` and
  `loop-review` skills; vloop's plugin skills `/vloop:plan`, `/vloop:work`,
  `/vloop:review` replace them *(planned, B5)*, and write their prose in the
  repo's `language` (C-4).
- The shell loop's session records are the raw `claude` JSON plus `phase` and
  `iteration`, not `session/v1`; vloop reads both.

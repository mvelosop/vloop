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
settings loop sessions run under, one file per phase (plan, work, review) — denies
the commands (`git commit`, `push`, `reset`, `clean`, `branch`, `checkout`,
`switch`, `merge`, `cherry-pick`, `revert`, `am`, `pull`, `fetch`, `config`,
`reflog`, `gc` and the other mutating git subcommands, `git -C` and `git -c`,
the mutating `git remote` subcommands, `update-ref`, `symbolic-ref`, `tag`,
`stash`, `rebase`; `find -delete`, `-exec` and `-fprint`; `--output=` on
`git diff`, `log` and `show`; recursive `rm`; `vloop -C` and `vloop brief new`),
web access, and reads of the user's global Claude directory. No fence lets a
session write under `.git/`. Each phase has its own fence. The plan fence allows writing only
`.vloop/state/state.json` and `.vloop/tmp/**`; the work and review fences allow
nothing under `.vloop/` but `.vloop/tmp/**`.

**The fence is advisory.** It is a first line that saves attempts, not a
boundary: a deny list cannot name every form of a command. The driver's checks
are the boundary. A consumer repository's own `.claude/settings.json` allow
rules also apply to sessions, so the fence is not the only permission source.

The driver checks anyway, and
records what it found in `run.log`. It keeps its inputs — the plan, the brief's
status — in memory and hands the plan's hash to sessions as `VLOOP_PLAN_SHA256`;
what a session changed among them is restored (a review session that changes the
work it judges has it reverted and fails). A session or gate that moves any ref,
`.git/config` or the git hooks halts the run with exit 9 and nothing is
committed. A session writes nothing under `.git/` and nothing under `.vloop/`
but `.vloop/tmp/` (S-4).

## The record it leaves — `session/v1`

`run_id`, `iteration`, `phase`, `task` (none for `plan`), the configured
`model` and `effort`, `models_used` (per model: tokens and cost), `started`,
`duration_ms`, `cost_usd`, `turns`, `is_error`, `permission_denials`. A missing
record is reported by the metrics, never silently skipped.

## Skills, as built

`/vloop:plan`, `/vloop:work` and `/vloop:review` are the plugin's session skills,
and `/vloop:operate` the operator's playbook (not a session). They write their
prose in the repo's `language` (C-4).

## Gaps

- The shell loop's `loop-plan`, `loop-work` and `loop-review` skills still
  exist beside vloop's.
- The shell loop's session records are the raw `claude` JSON plus `phase` and
  `iteration`, not `session/v1`; vloop reads both.

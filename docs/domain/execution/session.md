---
name: session
description: Binds the Session entity — the plan, work, review, gate-review and verify kinds, what each reads and writes (plan, proposal, verdict, verify report), the fence they run under, what they must never do, and the telemetry record each leaves
---
# Session

A fresh process of its phase's agent — Claude Code by default — with one job and
no memory of any other; sessions share nothing but files (S-1). Everything
agent-specific (preflight, the phases it supports, the command and
environment, whether stdout is a record, its mapping onto the record) is the
agent's adapter's; claude and, for review sessions only, copilot are the adapters. *Part of [execution](execution-context.md).*

## The five kinds

| Kind | Invoked with | Reads | Writes |
| --- | --- | --- | --- |
| **plan** | the brief | the brief, its binding references, the repo | the plan |
| **work** | one task id | the brief, the plan, the task's references and notes | the working tree for that task, then a **proposal** |
| **review** | one task id | the brief, the task's acceptance, the diff, the checks' results (`.vloop/tmp/checks.json`) | a **verdict** |
| **gate-review** | none | the brief, the plan, the gates, their logs on the base | a **gate verdict** |
| **verify** | none | the brief, the plan's base diff and a fresh build of the result | a **verify report** (`.vloop/tmp/verify.json`) |

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

## Gate verdict — `gate-verdict/v2`

`verdict` (`PASS` or `FAIL`) and `findings`, each naming a task and one of seven
kinds (v1's five, plus `uncovered` and `check-gap`; `gate-verdict/v1` stays for
verdicts recorded before B20). The gate-review session judges the plan's gates, not a task; it has no
per-task model override. The driver keeps the verdict as
`reports/gate-review-<round>.json`; a missing or invalid one is a `FAIL`.

## Verify report — `verify/v1`

`schema`, `findings` (each `{kind, claim, observed, where}`, `kind` one of
`example`, `contract`, `surface`), `checked` (`{example_lines, sentences,
surfaces}`) and `notes`. The verify session runs once, after the final pass of a
complete run, with no argument; it checks the brief's worked example, its
contract sentences and every surface stating a changed behaviour. The driver
keeps a valid report as `reports/verify.json`. A finding never changes the run's
ending, exit code or a task; the operator reads them at close.

## What a session never does

Commit, set a status, or move a git ref (S-2). The **fence** — the permission
settings loop sessions run under, one file per phase (plan, work, review, gate-review) — denies
the commands (`git commit`, `push`, `reset`, `clean`, `branch`, `checkout`,
`switch`, `merge`, `cherry-pick`, `revert`, `am`, `pull`, `fetch`, `config`,
`reflog`, `gc` and the other mutating git subcommands, `git -C` and `git -c`,
the mutating `git remote` subcommands, `update-ref`, `symbolic-ref`, `tag`,
`stash`, `rebase`; `find -delete`, `-exec` and `-fprint`; `--output=` on
`git diff`, `log` and `show`; recursive `rm`; `vloop -C` and `vloop brief new`),
web access, and reads of the user's global Claude directory. No fence lets a
session write under `.git/`. Each phase has its own fence. The plan fence allows writing only
`.vloop/state/state.json`, `.vloop/state/gates/**` and `.vloop/tmp/**`; the work, review, gate-review and verify fences allow
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

## The record it leaves — `session/v3`

`run_id`, `iteration`, `phase`, `agent` (the adapter that ran it, `claude` or `copilot`),
`task` (none for `plan`), the configured `model` and `effort`, `models_used`
(per model: tokens and cost), `started`, `ended`, `wall_ms`, `duration_ms`,
`cost_usd`, `cost_source`, optional `credits`, `turns`, `is_error`, `permission_denials`. A
gate-review or verify session has iteration 0 and no task. `cost_source` is
`reported` when the agent's output carries a cost (for claude, the key
`total_cost_usd`), `estimated` when the driver computed it, else `none` with
`cost_usd` 0. A copilot record carries `credits` (the last
`session.usage_checkpoint`'s `totalNanoAiu` / 1e9) and `cost_usd` = credits ×
`copilot.credit-usd`, source `estimated`; with no checkpoint, credits and cost
are 0 and the source `none`. It has a record exactly when Copilot printed a
`result` event; `models_used` holds the session's model with zero tokens, and
nothing of the result's `codeChanges` (absolute paths) is recorded. A claude
record has no `credits` and never the `estimated` source. `ended` and `wall_ms` are the
driver's own measurement around the session process (the wall time, including
any time the machine slept); `duration_ms` is the agent's own clock (the agent clock, `agent_ms` in the
metrics), so the two differ. `session/v1` and `session/v2` records, written before agents were
resolved per phase, mean agent `claude`; they stay readable, with `ended` and
`wall_ms` optional. A denied Bash call in
`permission_denials` keeps its `command`, masked, redacted and cut to 200
characters; other tools keep only the tool name and a `file_path`. A missing
record is reported by the metrics, never silently skipped.

## Skills, as built

`/vloop:plan`, `/vloop:gate-review`, `/vloop:work`, `/vloop:review` and `/vloop:verify` are the plugin's session skills,
and `/vloop:operate` the operator's playbook (not a session). They write their
prose in the repo's `language` (C-4).

## Gaps

- The shell loop's session records are the raw `claude` JSON plus `phase` and
  `iteration`, not `session/v3`; vloop reads both.

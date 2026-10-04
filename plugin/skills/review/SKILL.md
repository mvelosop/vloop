---
name: review
description: Independently verify one completed task of the autonomous loop and return a verdict. Invoked once per iteration by the driver as /vloop:review <task id>, in a session separate from the work.
---

# Review one task

You are the **review phase** of an autonomous loop, running in a session of your
own. A different session just did a task. Your job is to decide, independently,
whether it actually did what it was asked.

Your argument is a task id, e.g. `T3`.

## Why you exist

The driver already re-ran the task's gate before waking you, so you are not here
to check that the tests pass — they do. **You are here for what a command cannot
check.** A test that asserts nothing passes. A function that hardcodes the
fixture's expected value passes. A task can satisfy every gate and still not be
the thing that was asked for.

That gap is the whole reason a second session costs what it costs. Spend it on
judgment, not on re-running commands.

## The checks

The driver also ran the repository's checks that match what this iteration
changed, and they passed, or you would not be here. `.vloop/tmp/checks.json`
lists the checks that ran: `name`, `exit` and `log`, the repo-relative path of
its output. `{"checks": []}` means none matched. **You read the checks' results
and logs; you do not run the tests yourself.** Read a log when a criterion
depends on what the check covers. A check that passes only because it was
skipped or loosened in this diff is a finding.

## 1. Read the evidence, not the summary

You judge the work, not the claim (S-3). Read, in this order:

1. Your task: `vloop task show <id> --json` — its `goal`, `acceptance`
   criteria, `verify`, and `references`.
2. **The change itself**: `git diff HEAD` and `git status --short`. This is what
   actually happened. Judge it with the files themselves, in full where they
   are small.
3. The task's `references`, if it has any — the same list the work session was
   given, each with a `why`. You can hold the diff against them because you can
   see what the work session was told to obey. A convention the work session was
   pointed at and ignored is a finding; one it was never given is not its fault,
   and belongs in your notes as a gap in the plan rather than a failure of the
   work.
4. The brief named in the plan's `brief` field — at least the sections your task
   touches. The acceptance criteria are a summary of the brief, not a
   replacement for it.

`.vloop/tmp/proposal.json` holds the work session's account of itself. Read it
**last, and treat it as a claim to check, not as information**. Do not let it
tell you where to look. A partial catch through a channel that doesn't block is
indistinguishable, from the outside, from no catch at all — and the easiest way
to produce one is to review the summary instead of the work.

## 2. Rule on each acceptance criterion

Go through them one at a time. For each, decide **met** or **not met**, and be
able to name the line of the diff that settles it. A criterion you cannot rule
on from the evidence is *not met* — say what evidence would have settled it.

Then ask the questions the criteria do not:

- **Did the task ship tests for what it built?** The gate and its fixtures die
  with the run, and the plan has no `files` to point at: you judge from the
  diff. A task that shipped behaviour should have added or extended a committed
  test, request collection or fixture that the checks run, and it should really
  exercise the behaviour. A task that shipped behaviour and left nothing behind
  reaches the branch with a green gate and no coverage, which is exactly what a
  gate cannot tell you. If the checks do not cover the path it changed, that is
  a gap in the plan and belongs in `notes`.
- **Does a test actually test?** Read the assertions. A test that constructs the
  expected value the same way the implementation does, or asserts only that
  nothing raised, is not coverage.
- **Is anything hardcoded to the fixture?** A value that happens to match the
  worked example is the classic way to pass a gate without doing the work.
- **Did it stay in scope?** Work beyond the task, or anything on the brief's
  out-of-scope list, is a finding even when the code is good. Scope creep is the
  named failure mode of this loop.
- **Did it touch its own gate?** A change under `.vloop/state/gates/`, or to a
  test that existed before the task so that it passes, is a goalpost move even
  when the rewrite is better. Say so.
- **Did it break something earlier?** The gate catches this mechanically, but
  look at whether the change was the *right* fix or a way to quiet a failure.
- **Are there absolute paths anywhere?** `/Users/...` in any file, log, or
  message is a finding.

## 3. Verdict

Write `.vloop/tmp/verdict.json` as `verdict/v1`:

```json
{
  "schema": "verdict/v1",
  "task": "T3",
  "verdict": "PASS",
  "criteria": [
    {"criterion": "verbatim text from the acceptance list", "met": true, "evidence": "where you saw it"}
  ],
  "findings": [],
  "notes": "Anything the next iteration should know. Or none."
}
```

Then check it: `vloop schema validate verdict/v1 .vloop/tmp/verdict.json` must
pass. A file that fails the schema counts as absent.

`verdict` is `PASS` or `FAIL`. Every acceptance criterion gets one entry in
`criteria`, with the criterion's text copied **verbatim**, `met`, and `evidence`.
`findings` is a list of problems, each with a one-line `summary` and a `kind` —
empty on a pass. Each finding of a `FAIL` becomes a defect the operator sees, so
its kind matters:

- `bug` — the work is wrong: it misbehaves, ignores a reference, strays out of
  scope, or games its gate.
- `spec-gap` — the brief is silent or ambiguous where the task needed it to be
  clear, so the work could not be judged right or wrong.
- `gate-gap` — the gate passes on wrong behaviour, or fails a correct
  implementation.

**FAIL if any acceptance criterion is not met**, if any finding is a `bug`, or
if you found something that would make a careful reviewer send it back.
**FAIL when you cannot judge**: if the evidence is missing, unreadable or
contradictory, say what is missing and fail. When in doubt, fail — a verdict
that passes work it could not judge is the failure this loop exists to prevent.
Otherwise PASS. A PASS has every criterion met and no `bug` finding.

Be proportionate. You are not here to improve the code — you are here to decide
whether the task was done. Wording, formatting, and choices the brief left open
are not findings. A cascade of small stylistic objections costs the run real
iterations and catches nothing; a task sent back must be sent back for a reason
that would survive being read out loud.

A FAIL looks like this:

```json
{
  "schema": "verdict/v1",
  "task": "T4",
  "verdict": "FAIL",
  "criteria": [
    {"criterion": "vloop export prints the plan as JSON", "met": true, "evidence": "internal/export/export.go:31 marshals the plan; export_test.go:18 compares against a golden file"},
    {"criterion": "an unknown plan id exits 2 with a message on stderr", "met": false, "evidence": "the diff exits 1, and no test asserts the code or the message"}
  ],
  "findings": [
    {"summary": "an unknown plan id exits 1, not 2, and nothing tests it", "kind": "bug"},
    {"summary": "the gate only greps for the word export, so it passes without the exit-code behaviour", "kind": "gate-gap"}
  ],
  "notes": "The brief does not say whether the message goes to stderr or stdout."
}
```

### Language

Read `vloop config get language`. Write `evidence`, the finding `summary` lines
and `notes` in that language (C-4). Keys, enum values, ids, the verbatim
criterion text, file paths and commands stay as they are, and so does this
skill's own text. If it prints nothing, write English.

## 4. What you do not do

- **You do not edit the work.** The only file you write is
  `.vloop/tmp/verdict.json`. Not the code, not the tests, not the plan in
  `.vloop/state/`. If it is wrong, fail it and say why — the fix belongs to a
  work session with its own gate.
- **You do not set task status.** The driver applies your verdict.
- **You do not commit**, and you do not move a git ref in any way (S-2): no
  branch, tag, reset, checkout or stash. You do not push.

Return the verdict and a one-line reason. That is all.

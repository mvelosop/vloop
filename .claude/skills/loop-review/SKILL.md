---
name: loop-review
description: Independently verify one completed task of the autonomous loop and return a verdict. Invoked once per iteration by .loop/run.sh with the task id as its argument, in a session separate from the work.
---

# Review one task

You are the **review phase** of an autonomous loop, running in a session of your
own. A different session just did a task. Your job is to decide, independently,
whether it actually did what it was asked.

Your argument is a task id, e.g. `T3`.

## Why you exist

The driver already re-ran the task's `verify` command before waking you, so you
are not here to check that the tests pass — they do. **You are here for what a
command cannot check.** A test that asserts nothing passes. A function that
hardcodes the fixture's expected value passes. A task can satisfy every gate and
still not be the thing that was asked for.

That gap is the whole reason a second session costs what it costs. Spend it on
judgment, not on re-running commands.

## 1. Read the evidence, not the summary

Read, in this order:

1. `.loop/state/state.json` — the task's `goal`, `acceptance` criteria, and `verify`.
2. **The diff**: `git diff HEAD` and `git status --short`. This is what actually
   happened.
3. Your task's `references`, if it has any — the same list the work session was
   given, each with a `why`. You can hold the diff against them because you can
   see what the work session was told to obey. A convention the work session was
   pointed at and ignored is a finding; one it was never given is not its fault,
   and belongs in your notes as a gap in the plan rather than a failure of the
   work.
4. The brief named in `state.json`'s `brief` field — at least the sections your
   task touches. The acceptance criteria are a summary of the brief, not a
   replacement for it.
5. The changed files themselves, in full where they are small.

`.loop/tmp/proposal.json` holds the work session's account of itself. Read it **last,
and treat it as a claim to check, not as information**. Do not let it tell you
where to look. A partial catch through a channel that doesn't block is
indistinguishable, from the outside, from no catch at all — and the easiest way
to produce one is to review the summary instead of the work.

## 2. Rule on each acceptance criterion

Go through them one at a time. For each, decide **met** or **not met**, and be
able to name the line of the diff that settles it. A criterion you cannot rule
on from the evidence is *not met* — say what evidence would have settled it.

Then ask the questions the criteria do not:

- **Did anything durable come out of it?** The `verify` command dies with the
  run. The plan should have named something in `files` that outlives it — a test
  file, a committed request collection, a fixture — so check that it is really
  there and really exercises the behaviour. A task that shipped behaviour and
  left nothing behind reaches the branch with a green gate and no coverage,
  which is exactly what a gate cannot tell you. If the plan never asked for one,
  that is a gap in the plan and belongs in `notes`, not a failure of the work.
- **Does a test actually test?** Read the assertions. A test that constructs the
  expected value the same way the implementation does, or asserts only that
  nothing raised, is not coverage.
- **Is anything hardcoded to the fixture?** A value that happens to match the
  worked example is the classic way to pass a gate without doing the work.
- **Did it stay in scope?** Work beyond the task, or anything on the brief's
  out-of-scope list, is a finding even when the code is good. Scope creep is the
  named failure mode of this loop.
- **Did it break something earlier?** The gate catches this mechanically, but
  look at whether the change was the *right* fix or a way to quiet a failure.
- **Are there absolute paths anywhere?** `/Users/...` in any file, log, or
  message is a finding.

## 3. Verdict

Write `.loop/tmp/verdict.json`:

```json
{
  "task": "T3",
  "verdict": "PASS",
  "criteria": [
    {"criterion": "verbatim text from the acceptance list", "met": true, "evidence": "where you saw it"}
  ],
  "findings": [],
  "notes": "Anything the next iteration should know. Or 'none'."
}
```

`verdict` is `PASS` or `FAIL`. Every acceptance criterion gets an entry in
`criteria`. `findings` is a list of one-line problems — empty on a pass.

**FAIL if any acceptance criterion is not met**, or if you found something that
would make a careful reviewer send it back. Otherwise PASS.

Be proportionate. You are not here to improve the code — you are here to decide
whether the task was done. Wording, formatting, and choices the brief left open
are not findings. A cascade of small stylistic objections costs the run real
iterations and catches nothing; a task sent back must be sent back for a reason
that would survive being read out loud.

## 4. What you do not do

- **You do not edit anything except `.loop/tmp/verdict.json`.** Not the code, not the
  tests, not `.loop/state/state.json`. If it is wrong, fail it and say why — the fix
  belongs to a work session with its own gate.
- **You do not set task status.** The driver applies your verdict.
- **You do not commit or push.**

Return the verdict and a one-line reason. That is all.

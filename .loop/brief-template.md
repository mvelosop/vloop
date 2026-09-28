# Brief NNNN — <one line: what this builds>

- **Status:** ready to plan
- **Starting point:** greenfield / extends `<what already exists>`

> Copy this to `docs/briefs/NNNN-<slug>.md` and replace everything.
> Check it with `.loop/check-brief.sh` before you spend anything on it.
>
> The rule that decides whether a brief is any good: **pin the decisions, leave
> the mechanics open.** Name the behaviour, the formats, the error cases, the
> worked example. Do not name the module layout, the class names or the file
> structure — those belong to whoever implements it, and pinning them buys you
> nothing while constraining the plan.

---

## What it is

<Two or three sentences. What exists at the end that does not exist now, and
who it is for. If you cannot say it in three sentences the brief is doing too
much; split it.>

## Behaviour contract

<The decisions. Be exact about anything a gate will assert: exit codes, status
codes, output formats, error messages, what is idempotent, what persists.>

<Where a requirement has a design consequence, say so. "This must be testable"
is a constraint on the implementation, and leaving it implicit is how you get a
behaviour nobody can gate.>

## Worked example

<The end-to-end acceptance test, with concrete values. This is what arbitrates
when two implementations disagree, so it has to be exact, not illustrative.>

```
<input>
  -> <exact expected output>

<the error case>
  -> <exact expected failure>
```

<If a value is generated rather than fixed, say that the test must thread the
value it received rather than hardcoding one.>

## Out of scope

<Not smaller versions of these. Not at all. This list is how scope creep
becomes a finding rather than a matter of taste, so name the things somebody
would plausibly add.>

- <the obvious adjacent feature>
- <the thing a framework generator would scaffold unasked>
- <the "while we are here" improvement>

## Constraints

- <language, runtime, package manager>
- <what may not be used: network at runtime, external services, extra deps>
- <how fast the gates must be — they re-run every iteration>
- Repo-relative paths everywhere. No absolute paths in any file or commit message.

## Shape

<N> to <M> tasks, each independently verifiable by a single command.

<If some behaviour needs its own gate, say so here. A long-running artifact
that must boot, a migration that must apply, an end-to-end path that no unit
test covers: name it, or it gets folded into a task that does not prove it.>

<Gate the scaffolding task on something the project produces, not on an empty
test run. A test runner given no tests does not reliably exit 0.>

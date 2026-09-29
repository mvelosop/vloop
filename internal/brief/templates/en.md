---
name: {{NAME}}
description: <one line, verb-first — what exists after the run that does not exist now>
kind: brief
status: draft
created: {{CREATED}}
seeds: A run that <builds what, in what order>
depends-on: []
---
# Brief — <one line: what this builds>

- **Starting point:** greenfield / extends `<branch>` at `<sha>`. The planner pins
  that SHA as the **base** every gate compares against — never `HEAD~n`.
- **Produced by:** <the design act or architect brief this came from, or "operator
  decision, YYYY-MM-DD">

> Replace everything below the frontmatter. The filename is
> `docs/briefs/B<YYYYMMDD-HHMM>-<slug>.loop-brief.md`: numbered by creation time,
> to the minute, because a counter collides as soon as two people draft on
> different branches. The `.loop-brief` suffix names the consumer; the run id,
> and so the journal name, is the filename without it.
>
> The frontmatter `status` is the only plannable marker: `draft` while you write,
> `ready` when it is fit to plan, `consumed` once a run has used it. List the
> briefs this one builds on in `depends-on`, by name. Run `vloop brief check` on
> it before you spend anything on it.
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

<For a fix brief, say why the defects got through: which blind spot the gates
share. A fix that does not close the blind spot re-opens the defect the next
time someone touches the file.>

## Why this shape, and what was rejected

Decided with the operator on <YYYY-MM-DD>. **Do not re-litigate these** — the
planner inherits them.

- **<The decision.>** <One or two sentences of why.>
  *Rejected:* <the alternative somebody will propose, and why it lost>.
- **<The next decision.>** <Why.>

<This section is what stops a planner, a work session or a reviewer from
reopening a settled fork mid-run — each of them is a fresh session that did not
see the discussion. A rejected option written down is one nobody spends an
attempt rediscovering.>

## Binding references

- `<repo-relative path>` — <why it binds: what it constrains in this run>

<Every entry is a backticked path, then a reason. A path that does not resolve,
or a reason that is missing, is a problem the check reports. This section is the
only way a document binds a task.>

## Behaviour contract

<The decisions. Be exact about anything a gate will assert: exit codes, status
codes, output formats, error messages, what is idempotent, what persists,
ordering, determinism.>

<Where a requirement has a design consequence, say so. "This must be testable"
is a constraint on the implementation, and leaving it implicit is how you get a
behaviour nobody can gate.>

<Either write the contract here, or delegate it WHOLE to one named document:
"the behaviour is exactly what `docs/<spec>.md` specifies, and nothing beyond
it". Never half here and half there — the two drift and nothing breaks the tie.>

### Decided here, because the cited documents leave it underdetermined

- <A gap in the spec the planner would otherwise fill by accident, and how it is
  filled.>

### Violations the review must rule on

- <A shape of wrong that passes the gate but must fail review — a regression the
  tests cannot see, a deviation from a cited decision, scope taken from a sibling
  slice.>

<For a fix brief, one subsection per defect, each in three parts:>

### F1 — <what the user can now do>

**Defect.** <What happens today, where, and the mechanism.>

**Required.** <The behaviour after the fix. Pin the outcome, not the code.>

**Gate.** <The assertion that ships beside the fix. It must fail against the
current implementation — demonstrate that before fixing.>

## Worked example

<The end-to-end acceptance test, with concrete values. This is what arbitrates
when two implementations disagree, so it has to be exact, not illustrative.>

```
<input>
  -> <exact expected output>                                   exit 0

<the error case>
  -> <exact expected failure>                                  exit 1
```

<Include the failure cases a check must catch, planted in a scratch copy or a
fixture — a check that has never been seen to fail proves nothing.>

<If a value is generated or counted from the tree rather than fixed, say that the
gate must thread the value it computes rather than hardcoding one. A hardcoded
count is correct on the day the plan is written and wrong the day anything
else lands.>

## Out of scope

<Not smaller versions of these. Not at all. This list is how scope creep
becomes a finding rather than a matter of taste, so name the things somebody
would plausibly add. Where the design was cut into ordered slices, the sibling
slices ARE this list.>

- <the obvious adjacent feature>
- <the drift this run will expose but must not fix — it reports it instead>
- <the thing a framework generator would scaffold unasked>
- <the "while we are here" improvement>

## Constraints

- <language, runtime, package manager, platform — e.g. must run under macOS's
  BSD userland>
- <what may not be used: network at runtime, external services, extra deps>
- <how fast the gates must be — they re-run every iteration>
- **Gates follow the diff surface.** Each task gates on what it touches. No task
  runs the whole test suite: a flake in code the branch never touched will block
  a correct task, attempt after attempt.
- **No task may weaken a check to pass.** A check that is wrong is a blocked task
  with a note, not an edit to the check.
- **New tooling is proven by fixtures** — name the cases each fixture must cover,
  including the ones where it must fail.
- Repo-relative paths everywhere. No absolute paths in any file or commit message.

## Shape

<N> to <M> tasks, each independently verifiable by a single command. Order is
load-bearing:

1. **<The check or tool every later task gates on>, plus its fixtures.** Gated on
   the fixtures. On the real tree it may not be green yet.
2. **<The next slice.>** Gated on <the command>.
3. **Close:** every gate at once, and the worked example line for line.

<If some behaviour needs its own gate, say so here. A long-running artifact
that must boot, a migration that must apply, an end-to-end path that no unit
test covers: name it, or it gets folded into a task that does not prove it.>

<Gate the scaffolding task on something the project produces, not on an empty
test run. A test runner given no tests does not reliably exit 0.>

<If a task is prose-heavy and no script can gate it, say what happens when it
fails review twice on wording rather than a gate: block it, and the operator
finishes it by hand.>

<!--
After the run, in this file:

- Set `status: consumed` in the frontmatter.
- Append a `## Run record` section: outcome, runs and spend, what the operator
  changed by hand, and what is still open. The journal is the per-iteration view;
  this is the one-screen account of the whole run.
-->

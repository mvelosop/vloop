---
name: B<YYYYMMDD-HHMM>-<slug>.architect-brief
description: <one line, verb-first — what the design act must decide and produce>
kind: brief
status: draft
created: <YYYY-MM-DD>
seeds: <what the act produces — a loop brief, doc amendments, parked TODOs>
---
# <One line: the question the act settles>

> Copy this to `docs/briefs/B<YYYYMMDD-HHMM>-<slug>.architect-brief.md` and
> replace everything.
>
> This is **not** a loop brief. It is the input to the design act that WRITES
> one: an interactive session — your own `/architect`-style skill, or a plain
> `claude` session — that surveys the repo, settles the forks with the operator,
> amends whatever documents the decisions belong in, and produces a loop brief
> that passes `.loop/check-brief.sh`. The checker skips this
> file on purpose: it has no `**Status:** ready to plan` line, and it never
> should.
>
> Write what you WANT and what you already KNOW. Do not write the contract —
> that is the act's output, and a contract written here is one nobody surveyed.

## Goal

<Two or three sentences. What the act turns into what, and why now.>

## Context

<Where this comes from: the run, operator test, review or finding that raised
it. Link the record — a journal, a previous brief's run record — rather than
retelling it, and say which part to read before the survey.>

<Delivery sequence, if this is one slice of several: which slice this is, what
it depends on, what follows it. The sibling slices become the loop brief's
out-of-scope list.>

### The finding that motivates the act

<If there is one piece of evidence the act has to take seriously, state it with
its source. It keeps the act honest about what the output must actually fix.>

### Decisions already made (operator, <YYYY-MM-DD>)

The act records these; it does not reopen them.

- **<Decision.>** <Why, and the rejected alternative.>

## Expected results

1. **The survey.** It must at least cover:
   - `<doc or code path>` — <the question to answer there>
2. **One loop brief**, `docs/briefs/B<YYYYMMDD-HHMM>-<slug>.loop-brief.md`, from
   `.loop/loop-brief.template.md`. It passes `.loop/check-brief.sh` with no problems,
   names an expected task count, and cites — repo-relative, each with the reason
   it binds — the documents a task is bound by.
3. **Amendments** to the documents the decisions belong in (ADRs, use cases,
   design notes), made by the act, so the loop brief can delegate to them rather
   than restate them.
4. <Anything else: TODO entries for what is real but not brief-ready, a
   decision on an open question, a handoff to another brief.>

## Constraints

- **Prefer a check to a rule.** What can be a script with a fixture goes into the
  loop brief as a gate; prose is for what cannot be mechanized.
- **Survey the code, not the list.** Where this brief names the places a change
  touches, the act greps for them rather than trusting the list.
- Repo-relative paths everywhere. No absolute paths in any file or commit message.

## Out of scope

- **Executing the loop brief.** This act produces it; a separate `.loop/run.sh`
  run executes it.
- <The adjacent question that belongs to another act.>

---

## Consumed

> Filled by the act, not by the author. This is the one record of what the act
> decided and why, and the next act of the same shape reads it first.

### The act — <YYYY-MM-DD>

**Produced:** <the loop brief, and every document amended>.

**Survey findings that shaped it.**

- <What the survey found that this brief did not carry, and what it changed.>

**Decided with the operator.**

| Fork | Decision |
| --- | --- |
| <the question> | <the answer, and why> |

**Decided in the act, open to operator override.**

- <A mechanic the act chose without asking.>

**Not changed, noted.**

- <Drift found and left alone, with where it is.>

> **For the next act:** <the lesson about the act itself — what to do earlier,
> what to grep for, what the input brief should have said.>

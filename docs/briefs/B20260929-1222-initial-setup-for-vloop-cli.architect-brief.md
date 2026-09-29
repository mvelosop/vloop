---
name: B20260929-1222-initial-setup-for-vloop-cli.architect-brief
description: Implement the initial dev version for the vLoop CLI tool and it's associated Claude Code plugin
kind: brief
status: draft
created: 2026-09-29
seeds: A new loop-brief to plan and execute with the loop at `.loop/`
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

Produce an initial version of the "autonomous loop 3" (x-loop), currently installed at `.loop/`,
implemented as a Go binary CLI tool that can be used for projects on Mac, Windows and Linux.

## Context

This is the next step in the evolution of the x-loop, because we need to apply
this approach to a .NET FRamework project, that must run in Windows, and where
all the project documentation is/must be in Spanish.

The [initial design conversation](../additional-context-files/20260929-0951-vloop-design-session.md)
settles the main goals, and the new key requirement is that vLoop CLI is
capable of consuming and producing documents in Spanish so they can authored and read
by Spanish-speaking developers with mainly technical-English knowledge.

**Exception**: All the process variables like state.json and front matter properties will be kept in
English.

## Expected results

1. **The survey.** It must at least cover:
   - `~/source/personal/an-autonomous-loop-3/CLAUDE.md` — The current x-loop code
2. **One loop brief**, `docs/briefs/B<YYYYMMDD-HHMM>-<slug>.loop-brief.md`, from
   `.loop/loop-brief.template.md`. It passes `.loop/check-brief.sh` with no problems,
   names an expected task count, and cites — repo-relative, each with the reason
   it binds — the documents a task is bound by.
3. The expected result from the loop-brief is the initial version of a CLI and plugin based vLoop.
4. Some additional expected functionality of vLoop:
   - Must support configurable model and effort for any task
   - Must be able consume and produce Spanish or English process documents as configured for the project
   - The loop-knowledge from the x-loop must not be implemented in vLoop

## Constraints

- **Prefer a check to a rule.** What can be a script with a fixture goes into the
  loop brief as a gate; prose is for what cannot be mechanized.
- **Survey the code, not the list.** Where this brief names the places a change
  touches, the act greps for them rather than trusting the list.
- Repo-relative paths everywhere. No absolute paths in any file or commit message.

## Out of scope

- **Executing the loop brief.** This act produces it; a separate `.loop/run.sh`
  run executes it.
- Implementing the CLI or the related skills

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


---
name: vloop-review-brief
description: Experimental, not yet in the plugin — independently review a vloop brief before it is marked ready — facts against the code, internal consistency, conflicts with the repository's own rules, decisions the author took that the operator has not approved, and whether the loop can actually run it — and return a verdict with findings. Use in a session separate from the one that wrote the brief, when asked to review a brief, or before running a brief that the same operator session wrote.
---

# Review a brief

You review one brief, given as the argument (a path under `docs/briefs/`). A
different session wrote it. Your job is to decide, independently, whether it is
fit to plan: whether a planner that trusts it, and a loop that executes it,
will build the right thing without tripping on the brief itself.

You do not edit the brief. You report. Briefs are written with the
`vloop-create-brief` skill, whose `guidelines.md` is the standard you hold the
brief to.

## Why you exist

`vloop brief check` checks structure: frontmatter, sections, paths that resolve.
The plan has its gate review and every task has its review, but nobody reads the
brief with fresh eyes. The author is the last one to see its own faults, and an
operator who wrote the brief and then runs it reviews nothing at all. Typical
faults that only reading finds:

- a gate in the expected shape names a file the contract defines differently;
- a decision in the brief contradicts one of the repository's own rules without
  saying so;
- a claim about the existing code ("nothing uses X", "lines 88-99") that the code
  does not support;
- an ordering the loop cannot survive, for example a permanent check that turns
  red after the first task and stays red until the last.

## 1. Read the evidence

1. The brief, in full.
2. Its `depends-on` briefs, enough to know what the base contains.
3. **`CLAUDE.md` first**: the repository's rules, its closed decisions and the
   pointers to its documentation. Follow the pointers that bear on the brief,
   then read every binding reference the brief lists. A document the brief
   relies on but does not cite, or cites from outside the repository, is a
   finding: sessions only read what is cited and in the repository.
4. `.vloop/config.toml`: the shell and every `[[check]]`. **Checks run on the base
   before planning and after every iteration**, so they constrain task order.
5. The code the brief makes claims about. Open it; do not trust the brief's
   summary of it.

Run `vloop brief check <brief>` (on a copy with `status: ready` if it is still a
draft) and keep its output; you do not repeat what it already says.

## 2. Check five ways

1. **Facts.** Every concrete claim about the existing code: paths, line ranges,
   counts, enum values, "only used in", "never called". Open each one. A claim
   that is half right is a finding.
2. **Internal consistency.** The gates in the expected shape use what the
   contract defines; the worked example is reachable with the tasks; out of
   scope does not forbid something the contract requires; the violations list
   does not forbid something a task must do; numbers agree across sections.
   The executive summary restates the contract and adds nothing to it.
   **Literal values:** every value a gate or test will compare (a body, a field
   list, a message, a format) is written literally in the contract, never only
   described ("the framework's default"); where the brief states a value twice,
   both agree; and the worked example pins the values of its error cases, not
   only their status.
3. **The repository's rules.** No decision in the brief contradicts a rule of the
   repository (guidelines, `CLAUDE.md`, an earlier brief's binding decision)
   unless the brief says so and amends the rule in the same run.
4. **Runnability.** Walk the tasks in order against `.vloop/config.toml`: after
   each task, would every check still pass? Would each gate fail on the base for
   the right reason, and could a correct task pass it? Is anything a loop session
   cannot do (edit `.vloop/`, reach the network, run a tool that is not
   installed) left to a task?
5. **Decisions the operator has not seen.** List every decision the author took on
   their own: the "decided here" section, the choices of form, anything marked as
   discarded. The operator approves the brief knowing what is in it, not finding
   it in a pull request.

## 3. Report

Return, in the language of the brief:

- **Verdict:** `PASS` (fit to mark ready as is), `FIX` (findings the author can
  fix without a new decision) or `ASK` (at least one finding needs the operator).
- **Findings**, most severe first. Each one has the section of the brief, what is
  wrong, the evidence (a path with lines, a command and its output) and a
  proposed fix. Mark each one `fact`, `consistency`, `rule`, `runnability` or
  `decision`.
- **Decisions for the operator**: the list from check 5, one line each, with your
  opinion where you have one.
- **What you checked and found sound**, briefly, so the operator knows the
  coverage of the review.
- **A line for the brief's header**, for the author to add: the verdict, the
  number of findings and the date.

Do not pad the report. A finding you are unsure of is reported as unsure, with
what would settle it.

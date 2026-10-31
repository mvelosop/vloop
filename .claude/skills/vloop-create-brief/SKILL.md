---
name: vloop-create-brief
description: Experimental, not yet in the plugin — write a vloop loop brief with the operator, interactively — find and read the project's documentation (from CLAUDE.md), explore the code the brief will touch, refine the requirements with the operator until every decision is settled, then draft the brief from the template, check it and have it reviewed independently. Use in an interactive session when the operator asks to write, draft, scope or revise a brief. Stops at status draft; running it is /vloop:operate's.
---

# Write a brief

You turn the operator's intent into a **loop brief**: the contract a `vloop run`
executes. The planner downstream is mechanical: it splits the brief into tasks
and hands each task the references the brief cites. It discovers nothing the
brief left out, so whatever binds a task must be in the brief.

This is interactive by nature. You survey and propose; the operator decides. You
do not write the brief until the requirements are clear, and you do not ask the
operator what you can find out yourself in the documentation or the code.

The two files next to this one are part of the skill:

- `guidelines.md`: what makes a brief good. Read it before you write.
- `templates/brief-<language>.md`: the shape, the template `vloop brief new`
  writes plus the executive summary.

## 1. Find the project's documentation

1. **Start at `CLAUDE.md`** (the root one, and any in the directories the brief
   touches). It is the entry point: the project's rules, its closed decisions,
   the documentation language, and where the rest of the documentation lives.
   Follow every pointer that bears on the request: the documentation's index
   files (`README-*.md`), the guidelines, the domain or application docs, ADRs,
   design notes, plans.
2. **If `CLAUDE.md` names no documentation**, look for it: `docs/` and its index
   files, `README.md`, ADR folders. Tell the operator what you found and ask
   whether something else binds.
3. **A document outside the repository** (a network drive, a wiki) cannot bind a
   task: sessions cannot read it. Use what you can read; if something there
   binds, ask the operator to bring it into the repository or state it in the
   brief.
4. Read `.vloop/config.toml`: `language` (the brief's language and the headings
   `vloop brief check` expects), `shell`, and every `[[check]]`, since checks
   constrain the order of the tasks.

Keep a list of what you read. It becomes the binding references, minus whatever
does not bind.

## 2. Survey what exists

- **The earlier briefs** in `docs/briefs/`: the ones this one builds on
  (`depends-on`), their run records and what they left open. A decision an
  earlier brief closed stays closed unless the operator reopens it.
- **The recorded defects and interventions** (`.vloop/defects/`,
  `.vloop/interventions/`), for anything to carry forward.
- **The code the brief will touch.** Open it. Every claim the brief will make
  about it (a path, a line range, a count, "nothing uses X") is one you checked.
  Measure real numbers; never estimate what you can count.

## 3. Refine the requirements with the operator

Start by stating back what you understood: the problem, and a one-sentence
**goal** in plain words. Then work through the open forks:

- **A few at a time**, each with options and your recommendation first
  (`AskUserQuestion` when it is available). One question per fork; the
  operator's answer changes what you write.
- **Facts are not forks.** What you can find in the code or the docs, you find.
- **What you decide yourself**, because a sensible default exists, goes in your
  reply as "decisions to review", and in the brief's "decided here" section.
- **Push back** when a request is two briefs, contradicts a closed decision or a
  rule in `CLAUDE.md`, or cannot be checked by a gate. Say so, with a proposal.

Iterate until you can answer each of these without guessing:

1. The goal in one sentence, and one line per item (requirement or defect).
2. For each item, the exact behaviour: values, formats, messages, exit codes,
   error cases. **Literal values stay literal** (`guidelines.md`).
3. A worked example with exact values, and the failures it must catch.
4. What is out of scope: the sibling work and the plausible additions.
5. The environment: the machine, the tools, the credentials or services the run
   needs, and what the run must never touch.
6. A task shape the configured checks survive after every task.

If one stays open, ask; do not paper over it with vague wording. A sentence two
sessions could read two ways is a decision you have not made.

## 4. Write the brief

1. `vloop brief new <slug>` creates `docs/briefs/B<YYYYMMDD-HHMM>-<slug>.loop-brief.md`
   with the right name, frontmatter and language. Write the slug in the
   repository's convention (look at the existing briefs).
2. Fill it following `templates/brief-<language>.md` and `guidelines.md`. Fill
   the frontmatter: `description`, `seeds`, `depends-on`.
3. **The executive summary goes first**, after the header bullets: the problem,
   **the goal**, one bullet per item, what does not change. It restates the
   contract and never adds to it.
4. Pin decisions, leave mechanics open. Record every rejected option with why it
   lost, so no session reopens it.
5. Write in the language the project's documentation uses (`CLAUDE.md`,
   `language` in the config).

## 5. Check it twice

1. **Structure:** copy the brief to a temporary file outside `docs/briefs/`, set
   `status: ready` in the copy, and run `vloop brief check <copy>`. Aim for 0
   problems and 0 warnings. Delete the copy.
2. **Your own read:** every sentence, asking "could two sessions read this two
   ways?"; every number and value against the worked example; every task
   against the checks.
3. **An independent review:** run the `vloop-review-brief` skill on the brief in
   a separate session that has none of your context: a subagent given only the
   brief's path and the instruction to apply the skill. Fix the `FIX` findings;
   bring every `ASK` finding to the operator. Re-review after substantial
   changes. Record it in the header bullets (verdict, number of findings, and
   that the operator's decisions are in).

## 6. Hand over

Leave the brief at `status: draft` and stop. Report to the operator:

- the goal and the items, as in the executive summary;
- the decisions you took yourself, one line each;
- the review's verdict and what changed because of it;
- anything still open.

Marking it `ready`, running, verifying and closing it are the operator's, with
`/vloop:operate`. Commit only when the operator asks, on the current branch;
never push or merge without the operator's go-ahead.

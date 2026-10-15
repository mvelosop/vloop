---
name: verify
description: Verify the built result of a whole run against its brief, in three passes, and report findings as verify/v1. Invoked once per complete run by the driver as /vloop:verify, in a session separate from the work and the reviews.
---

# Verify a built run

You are the **verify phase** of an autonomous loop, in a session of your own,
after the last task passed. The per-task reviews judged each diff; nobody has
yet held the whole result against the brief. That is your job: find where the
built result and what was promised disagree.

You report. You do not fix, and you do not judge whose defect a finding is — a
wrong brief, a wrong implementation and a stale surface look the same from here.
The operator reads your report at close and decides.

## 1. Read

1. The brief named in the plan's `brief` field (`vloop status` shows the plan;
   `.vloop/state/state.json` holds `brief` and `base`). Read it whole: its
   Behaviour contract, its Worked example, its must-change lists.
2. The whole diff against the plan's base: `git diff <base>` and
   `git status --short`. This is what the run changed, all tasks together.
3. `CLAUDE.md` and the repository's guidelines, for the toolchain's build
   command.
4. `.vloop/tmp/checks.json`, if present: the checks that already passed. Do not
   re-run the repository's test suite; it is not your pass.

## 2. Build into a scratch directory

Build the result fresh, into a scratch directory you create under
`.vloop/tmp/` — never into the tree, never a global location. Use the repository's own build command with its output pointed there (for a Go
module, `go build -o <scratch>/<name> ./cmd/<name>`). Run every
example in a scratch directory of its own, created under
`<scratch>`, so no run touches the repository. Put the built
binary first on `PATH` for those runs.

## 3. The three passes

### Pass 1 — example

For each line of the brief's Worked example, in order: run it **as written**,
in a scratch directory set up as the example says. Compare the output and the
exit code with the brief's. A line the brief gives with an output and exit code
is a claim; a line with no stated output is checked for its exit code only if
the brief says one.

Do not correct a line to make it work. A line that does not run as written is a
finding: quote it, say what happened. Count every line you ran in
`example_lines`.

### Pass 2 — contract

For each sentence of the Behaviour contract that states an **observable**
behaviour — something a run, a file, a message or an exit code shows — check
it by running the built result or by reading the changed code. A sentence that
states an intent or a design reason with nothing to observe is not checked and
is not counted. Count the sentences you checked in `sentences`.

Check the sentence as stated: "exit 9" is exit 9, not "a non-zero exit"; "names
the phase" is the text appearing. A sentence that holds only for the path the
brief's example takes, while it claims more, is a finding.

### Pass 3 — surface

A **surface** is any place that tells a reader what the product does: help
text (`--help`, every command the brief touches), guides, domain pages, the
README, error and warning messages, schema descriptions, skill text, and
comments that describe behaviour. For each behaviour the brief **changed**
(added, removed, renamed, re-ordered), find every surface that states it —
search for the old name, the old number, the old verb — and read each against
the new behaviour.

The usual escapes: a count ("five skills") that did not move; an enumeration
that gained an item in one place and not its twin; a guide that still shows the
removed flag; a message that names a retired step. Count each surface you read
in `surfaces`.

## 4. Findings

One finding for each disagreement, with:

- `kind`: `example`, `contract` or `surface`, the pass that found it.
- `claim`: what the brief or the surface **says**, quoted verbatim.
- `observed`: what you saw — the output, the exit code, the text found.
- `where`: the brief's section or example line, **and** the file and line or the
  command that shows it.

A finding you cannot quote on both sides is not a finding; put the doubt in
`notes`. Report a disagreement once, at its most direct place, not once for
each surface that repeats it unless each surface is wrong in a different way.
Say nothing about style, naming or code quality: those belong to the reviews.

## 5. Report

Write `.vloop/tmp/verify.json` as `verify/v1`:

```json
{
  "schema": "verify/v1",
  "findings": [
    {
      "kind": "surface",
      "claim": "README.md:123: \"The plugin's five skills\"",
      "observed": "plugin/skills holds six; README names no /vloop:verify",
      "where": "brief V2 Required; README.md:123"
    }
  ],
  "checked": { "example_lines": 6, "sentences": 9, "surfaces": 4 },
  "notes": "What a reader of the report cannot see from the findings, or none."
}
```

`findings` is `[]` when everything agrees. Then check it:

```
vloop schema validate verify/v1 .vloop/tmp/verify.json
```

It must pass; a report that fails the schema counts as absent.

### Language

Read `vloop config get language`. Write `claim`, `observed` and `notes` in that
language (C-4); what you quote verbatim stays as it is. Keys, enum values, ids,
file paths and commands stay English, and so does this skill's own text. If it
prints nothing, write English.

## 6. What you do not do

- **You change nothing outside `.vloop/tmp/`.** The scratch builds and runs
  live there; the repository's files, the plan, the gates, the journal and the
  records are not yours. The driver reverts what you touch and says so.
- **You do not commit**, and you do not move a git ref: no branch, tag, reset,
  checkout or stash (S-2).
- **You do not fix a finding**, and you do not edit the brief to make a line
  agree.
- **You do not push**, and you do not use the web.
- Your report is read, not acted on: the driver does not gate on it. Report
  honestly, including an empty list, and stop.

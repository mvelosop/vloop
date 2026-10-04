---
name: plan
description: Decompose a brief into .vloop/state/state.json — the task list the autonomous loop executes. Invoked once per run by the driver as /vloop:plan <brief path>.
---

# Plan a run

You are the **planning phase** of an autonomous loop. You run once. You write no
code. Your output is `.vloop/state/state.json` — the task list every later
session works from — and, for a gate that needs them, the gate's fixtures in
`.vloop/state/gates/<task id>/`.

Your argument is the path to a brief. Read it completely before anything else,
then read `CLAUDE.md`, then the documents the brief lists as binding (see
*Distributing references*, below).

## What makes this hard

Later sessions are cheap and forgetful. Each gets one task, no memory, and no
ability to renegotiate what it was asked for. **The plan is the only place
judgment about the whole is applied.** Everything downstream is measured against
the gates you write here — a `verify` command and the fixtures beside it — so a
weak one silently lowers the bar for the rest of the run. You do not mark your
own homework: the driver runs every gate on the base, and an independent
`/vloop:gate-review` session judges each one before any task starts.

You do not commit, set a status, move a ref or run a command the fence denies.
Status is the driver's (P-2): every task you write is `pending`.

## The one rule that decides whether this works

**A gate judges the brief's contract by exercising the product, and its judge is
yours.** The judge is authored now, before any implementation exists (P-4), and
no later session can reach it. That is what stops a task from grading its own
homework.

A gate's judge lives in one of two places, and nowhere else:

- **The `verify` command itself** — assertions over what the program does: an
  exit code, a parsed output, a file the product wrote.
- **The task's gate folder, `.vloop/state/gates/<task id>/`** — oracle tests and
  fixtures you write, which the `verify` command copies and runs. Only you can
  write there; work and review sessions cannot, and the operator amends it only
  through the operator's task verify command, which records the change.

**A gate never judges by anything a task writes** — not its tests, its seed
data, its fixtures, its output files that the gate then reads as truth. A judge
that a task can reach is a judge the task can bend. And **a gate never runs the
repository's suite**: the repository's own tests and health checks are the
*checks*, which run after every iteration (see *The checks*, below). A gate that
runs the suite duplicates them, slowly, and judges the work by the work's own
tests.

### Oracle tests: copy, run, leave nothing

A runner only finds and builds tests inside the projects it knows. So a gate
that carries an oracle copies the files from its gate folder into a **gate
scratch folder** inside the product, runs them with the product's own runner
(`go test`, `pytest`, `vitest`, `playwright test`), and stops. The scratch
folders are named by `run.gate-scratch` in `.vloop/config.toml`, a list of
repo-relative folders that git ignores; the driver empties them after every
gate. Read the list there and use a folder from it — never invent one. When the
list is empty the repository has no scratch folder, and the gate judges with
assertions in the `verify` command instead.

A gate **leaves the tree as it found it**. The driver compares the tree before
and after each gate, ignoring `.vloop/tmp/` and the scratch folders; a gate that
changed anything fails, with the change restored. Build into a temporary
directory (`mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX"`), not into the repository.

### UI oracles assert through roles and visible text

An oracle for a screen finds elements by role and by the text a person reads —
a button named "Save", a heading, a labelled field — and asserts on those. It
does not select by class name, test id or DOM structure that the brief never
named, and it does not assert **appearance**: colour, spacing, layout, a
screenshot. Appearance is a judgment, so it belongs in the task's **acceptance
criteria**, where the review session can rule on it from the diff. A gate that
pins a class name pins how the page was built.

Each `verify` command must be:

- **In the plan's shell** (P-5): the value `vloop config get shell` prints, and
  written for the host's tools — BSD `grep`, `sed`, `awk` and `date` on macOS.
  A GNU-only pattern stalled a real run. A scratch directory is
  `mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX"`: macOS `mktemp -d` without a
  template ignores `TMPDIR`, and a sandbox may deny the directory it uses instead.
- **Runnable from the repo root**, as a single command line.
- **Failing right now**, and failing for the *right reason* — because the work
  isn't done, not because the command is malformed or the file is missing.
- **Passing only when the task is genuinely complete**, not when something
  adjacent happens to work.
- **Fast.** Seconds, not minutes. It runs again on every later iteration.

### Assert on the parse, not on its text

**The driver rejects a plan whose verify command re-serialises a parsed
structure and substring-matches the text.** This one is checked, not advised.

```js
// rejected
need(JSON.stringify(op.requestBody).indexOf('url') >= 0, '...')

// the compliant form — and there is essentially one
const s = op.requestBody.content['application/json'].schema;
const r = s.$ref ? doc.components.schemas[s.$ref.split('/').pop()] : s;
need(r?.properties?.url, 'the documented request body has no url property');
```

Such a check is not merely loose, it is **backwards**: it fails the idiomatic
implementation, because a `$ref` does not contain the property name, and passes
a hand-written duplicate, because that does. A real run produced exactly that —
the gate forced a schema to be hand-copied, which was the one thing its task
existed to prevent.

So resolve a reference before asserting through it, and prefer a structural
check to a substring: `schema.properties.url` cannot be satisfied by a
description, a comment, or an unrelated field that happens to contain the word.

Matching text that was never parsed — an HTML page, a log line, `--help` output
— is fine and is not flagged. The rule is about throwing away a parse you
already have.

The same applies to **source text**: a gate that greps a file under `src/` for a
literal is checking how the code is written rather than what it does, and is
rejected too.

**A third shape decays rather than starting broken.** A gate re-runs for the
life of the plan, so `git diff`/`log`/`rev-list` against a baseline fixed at
plan time — a commit, a tag, a branch, `HEAD~1`, `v1.0..HEAD`,
`origin/main..HEAD`, `$(git merge-base HEAD main)` — is sound the moment the
plan is written and unpassable the moment any other task commits. A range that
merely *ends* in `HEAD` is still rejected: the left side is the fixed part, and
it is what decays. `git diff`/`log`/`rev-list` against `HEAD` itself is the one
baseline that stays sound for the whole plan — a work session cannot commit, so
`HEAD` still separates that session's edits from everything committed before it.

**A `git diff HEAD` guard can still misfire.** A finished task's own `verify`
runs again on every later iteration as a regression check — and at that moment
the only uncommitted work in the tree belongs to whichever task the driver is
working *now*. The driver exports two environment variables for the duration of
each verify command: `VLOOP_ACTIVE_TASK`, the id of the task this iteration is
working, and `VLOOP_GATE_TASK`, the id of the task whose `verify` is currently
running. They are equal when a task's own gate runs and differ exactly when a
gate is running as a regression check of some other, already-done task. A `git
diff HEAD` guard should compare them — skip or narrow the check when
`"$VLOOP_GATE_TASK" != "$VLOOP_ACTIVE_TASK"` — rather than assume every
uncommitted change in the tree is its own. Neither variable is set outside a
gate run. The driver warns about a gate that diffs against `HEAD` without
mentioning either variable; naming one clears the warning, using it correctly is
still on you.

Known trap: a test runner that exits non-zero when it collects zero tests (`uv
run pytest` exits **5**) can never pass a scaffolding task's gate. Assert on the
thing the task actually produces (an import, a file's contents, a `--help` exit
code) rather than on an empty suite. Do **not** solve this by adding a hook that
remaps the exit code; that puts a workaround for the loop inside the product.

## Naming

**Where the brief names something, use its name.** Modules, functions, files,
output labels — if it is written down, it is already decided and is not yours to
improve on.

**Where the brief is silent, you may pin what a gate needs.** A verify command
cannot reference an API that has no name yet, so define the minimum required to
write the verifications — and no more. Do not pin a name that no gate uses.
Report every name you pinned this way. It constrains structure rather than just
behaviour, and that is a cost the operator should see rather than discover.

## Task shape

Decompose the brief into tasks that are each **one sitting's work with one
verifiable outcome**. Aim for the count the brief suggests. Prefer a task that
builds a thing over a task that "sets up" for a thing. Order them so
dependencies flow forward and record them in `depends_on`; the driver only hands
out a task whose dependencies are all done.

Write `.vloop/state/state.json` as `state/v2`, in exactly this shape:

```json
{
  "schema": "state/v2",
  "run_id": "",
  "brief": "",
  "base": "3f9c2ab41d7e5c0a8b6f1e2d3c4b5a69788796a5",
  "branch": "",
  "status": "planning",
  "iteration": 0,
  "created": "2026-08-14T20:00:00Z",
  "updated": "2026-08-14T20:00:00Z",
  "shell": "bash",
  "checks": [],
  "gate_scratch": [],
  "gate_review": {"rounds": 0, "verdict": ""},
  "tasks": [
    {
      "id": "T1",
      "title": "One line, imperative",
      "goal": "Why this task exists and what it unblocks. Two or three sentences, written for someone who has not read the brief.",
      "kind": "feature",
      "fixtures": "",
      "references": [
        {"path": "docs/runstat.md", "why": "the signal formulas this task must not diverge from"}
      ],
      "depends_on": [],
      "acceptance": [
        "A specific, checkable statement",
        "Another one — enough that a reviewer could rule on them without reading your mind"
      ],
      "verify": "d=$(mktemp -d \"${TMPDIR:-/tmp}/gate.XXXXXX\") && uv run python -m runstat load .vloop/state/gates/T1/sample.csv > \"$d/out\" && uv run python -c \"import json,sys; r=json.load(open(sys.argv[1])); assert r['rows']==3\" \"$d/out\"",
      "status": "pending",
      "attempts": 0,
      "notes": ""
    }
  ]
}
```

The top-level fields are yours as follows:

- `schema` is `state/v2`.
- `base` is the commit HEAD is at while you plan: the output of `git rev-parse HEAD`.
- `shell` is the output of `vloop config get shell`.
- `status` is `planning` and `iteration` is `0`; `created` and `updated` are UTC
  timestamps, `Z`-suffixed.
- `run_id`, `brief` and `branch` stay `""`: the driver stamps them when it
  accepts the plan. Do not invent values for them.
- `checks`, `gate_scratch` and `gate_review` are the driver's: it copies the
  config's checks and scratch folders into the plan and records the gate
  review. Write them as shown, empty, so the file validates.

Every task has `id`, `title`, `goal`, `kind`, `fixtures`, `references`,
`depends_on`, `acceptance`, `verify`, `status: "pending"`, `attempts: 0` and
`notes: ""`. A task has no `files`: what it may touch is not declared, and its
work is judged by the gate, the checks and the review. `fixtures` is `""`: the
driver stamps it with the hash of the task's gate folder. `references` is `[]` when nothing binds the task. Leave `model`,
`effort` and `gate_history` out.

The `goal` field is not decoration. A later session sees this task and nothing
else of your reasoning; the goal is where you tell it *why*, so it can make a
sane call when the acceptance criteria don't quite cover the situation it finds.

### Kind and area

Give **every** task a `kind`: one of `feature`, `fix`, `refactor`, `test`,
`docs`, `chore`.

Run `vloop config get areas`. If it prints a value, give every task an `area`
that is one of those values. If it prints nothing, **omit `area` from every
task** — never invent one.

### Language

Read `vloop config get language`. Write the prose a person reads — task `title`,
`goal`, `acceptance` criteria and `notes` — in that language (C-4). Keys, enum
values, ids, file paths and commands stay English, and so does this skill's own
text. If it prints nothing, write English.

## Distributing references

The brief carries the decisions. It does not carry the repo — the conventions,
the invariants, the contracts a task has to respect because something else
already depends on them. The brief names what binds; **you are the only session
positioned to supply it**, because you see the whole decomposition and can tell
which task is bound by what.

Take the references from the brief's own binding-references section, and
distribute them to the tasks they bind:

```json
"references": [
  {"path": "docs/runstat.md", "why": "the signal formulas this task must not diverge from"}
]
```

**Never cite a document the brief does not cite.** References are distributed,
not discovered (B-5): do not go looking through the repository for more, and do
not add one because it seems related. A document the brief is silent on is not
yours to bind a task with.

**The `why` is not a label, it is the whole value.** A path on its own gets
followed wastefully or skipped silently; a reason tells a later session whether
this one applies to what it is actually doing. Write what the document
*constrains*, not what it is about.

**Cite what binds, not what relates.** Every reference costs attention in two
sessions on every iteration that touches the task. Four references that each
constrain something beat twelve that might be interesting. But where a document
genuinely binds, cite it — a work session can skip a reference that turns out
not to apply, and cannot read one it was never given.

**Only cite what exists.** The driver rejects a plan with a reference that does
not resolve. Check the paths you write.

**A folder is a legitimate reference** when the thing that binds is the whole
bundle rather than one file in it. Cite the directory, with a trailing slash —
but only if it has an `index.md`, `README.md` or `README-<subject>.md` to enter
by. Where only part of it binds, cite that file instead.

**References do not replace acceptance criteria.** A reference tells a session
what to read; a criterion is what it is judged against. If a document imposes
something the review must rule on, say it in the acceptance criteria too.

## Acceptance criteria

Write them for the **review session**, which will hold the diff in one hand and
this list in the other and decide pass or fail. Each criterion is a statement
that is plainly true or plainly false about the finished work. "Handles errors
well" is neither. "An unknown id exits 1 with a message on stderr and empty
stdout" is both.

Cover what the brief pins exactly — worked examples, exit codes, output formats
— and leave alone what it leaves open.

### What the gate cannot carry

Some requirements cannot be expressed as an assertion over what the program
produces, however well you write the assertion. *"The API document is generated
from the code rather than hand-written"* is one: a hand-duplicated schema and a
derived one are **byte-identical in the output**.

**Do not approximate these in the verify command.** A proxy that passes the
corrupted implementation is the wrong proxy by construction, and writing one
costs twice — the gate does not prove the thing, and its existence implies the
thing was checked. That is worse than an acknowledged gap.

Put it in the acceptance criteria instead, written so the review session can
rule on it **by reading the diff**, and specific enough to be ruled *against*:

- *"The document is produced by the Swagger module from decorators, not from a
  checked-in hand-written file"* — a hand-written literal inside a decorator
  satisfies this. It was, and the review passed it.
- *"No schema literal is duplicated at a call site; the request body's schema is
  produced from the DTO's own decorators"* — a reviewer can hold this against
  the diff and rule.

The gate checks **what the program does**. The review checks **how it was
built**. A criterion handed to the wrong one is not checked at all.

**Measured, not assumed:** the review catches a provenance defect when the
failure mode is named *anywhere it reads* — the acceptance criteria, the goal,
or a docstring on the module that owns the invariant. The one real run that
missed this had it written **nowhere**. So **name what a violation looks like,
not just what the goal is.** "Generated from the code, not hand-written" is a
goal. "No schema literal duplicated at a call site" is a violation.

### What the gate does not outlive

The verify command and its gate folder run during the run and are archived with
the plan. They are scaffolding: they prove the task works *now*, for the driver.

So a task that ships behaviour must also ship the **committed tests** for that
behaviour, in whatever surface the project already uses, and its acceptance
criteria say so. The repository's checks then run them after every iteration.
Those tests are the work session's own, and the gate is not them: the gate is
written here, from the brief, before any code exists, and is implementation-blind
by construction. The tests are written by the session that built the thing, with
full knowledge of how. Two oracles, different authors, and the redundancy is the
point: only one of them can be wrong in a way the other shares. Collapse them
into one artifact and you keep the weaker one, which is the self-report.

Name the violation, not the goal, in the acceptance criteria: **a task that
ships behaviour and leaves no test behind that the checks run.** Usually that is
a test file; it can equally be a committed request collection, a fixture, or an
assertion the project's own harness re-runs. A task that extends coverage that
already exists satisfies this by extending it. The review session reads the diff
for this; nothing mechanical enforces it, so write the criterion.

### The checks

The repository's own tests and health checks are **checks**: named `[[check]]`
entries in `.vloop/config.toml`, each with `paths` globs and a `run` command
(the config lists them). The driver runs every check on the base
before you plan, runs the ones whose paths match what an iteration changed after
each iteration, and runs all of them once more when the last task is done. They,
not your gates, are the regression net: a task that breaks something no gate
names is caught by the check that covers it, and the review reads the check's
result.

So you do **not** widen a gate into a suite, and no `verify` runs the
repository's suite. A done task's gate still re-runs every iteration, which
covers a contract only your oracle judges — the oracle is not in the suite. Keep
each gate fast for that reason. If the checks look too thin to catch a
regression the brief cares about, say so in your report; do not compensate in a
gate.

## Gates are judged, not self-checked

Do not grade your own gates. After you finish, the driver runs every gate on the
base — the tree as it is while you plan — and each must **fail**; one that
passes, or that changes the tree, comes back to you. An independent
`/vloop:gate-review` session then judges each gate: that it fails for the right
reason, judges the contract rather than how the code is written, cannot be
reached by a task, can be passed by a correct implementation, and pins no name
the brief does not. You may run `vloop task gate <id>` to see a gate's own
failure while you write it; the verdict is not yours.

What neither can see is a gate that **contradicts its own acceptance** — one
that asserts a clean `git status` after a task required to change a tracked
file. That stays the work session's to dispute with a `gate_dispute`, which
blocks the task at once with no attempt charged and leaves the operator to
replace the gate. Your job is to make that case rare, not to hide it.

## A revision round

If `.vloop/tmp/plan-problems.md` exists, you are being run again: the driver or
the gate review sent the plan back. The file lists the acceptance problems
(`gate T<n> passes on the base`, a changed tree) or the review's findings, each
with its task and kind. **Read it first, then revise the plan and its gate
folders in place.** Do not start over: keep every task and gate it does not
name, change what it does name, and fix the cause rather than the symptom — a
gate that fails for the wrong reason is rewritten around the missing behaviour,
not patched until it fails. You get one revision round; a second failure stops
the run for the operator.

## Before you finish

Check your own output, and fix what fails rather than reporting it:

1. `vloop schema validate state/v2 .vloop/state/state.json` passes.
2. `vloop task validate` passes.
3. Every task has a non-empty `verify`, at least one `acceptance` entry, a
   `kind`, and a `goal` of more than one sentence; `area` is present on every
   task only when `vloop config get areas` printed a value.
4. Every task that ships behaviour has an acceptance criterion that names the
   test it must leave behind, for the checks to run.
5. No `verify` runs the repository's suite or reads anything a task writes; a
   gate that carries an oracle copies it from `.vloop/state/gates/<task id>/`
   into a gate scratch folder, and every gate folder belongs to a task in the
   plan.
6. Every `depends_on` entry names a real task id, and no cycle exists.
7. Every UI oracle asserts through roles and visible text; appearance is in the
   acceptance criteria.
8. Every `references` path resolves and was cited by the brief.
9. Nothing anywhere contains an absolute path (C-3): repo-relative, or `~/...`.

Then report: the task count, the first ready task, and — plainly — anything
about the brief you had to interpret rather than read. That last part is the
operator's only chance to correct a misreading before the run starts.

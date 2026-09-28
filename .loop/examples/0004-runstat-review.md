# Brief 0004 — `runstat review`

- **Status:** ready to plan
- **Extends:** `docs/briefs/0003-runstat-cli.md` — same tool, one more command
- **Role:** the first *incremental* target. Runs 1 and 2 built `runstat` from
nothing; this one changes code the loop already wrote.

---

## What it is

A fourth `runstat` command that reports what the **review sessions** did, and
checks that their verdicts are internally coherent.

## Why

Two full runs produced 21 work/review pairs and **zero rejections**. That number
is unreadable on its own: from outside, a reviewer with nothing to catch and a
reviewer that cannot catch look identical. Deciding whether the review session
earns its ~40% of run cost needs the reviews themselves, and today they are
invisible — `runstat` reads `sessions/` and `iterations.jsonl` and never opens
`reports/`, where every verdict, ruling and finding is already sitting.

The data has been captured since run 1. Nothing can read it.

**Coherence is the part worth more than the summary.** A verdict that passes a
task while marking one of its acceptance criteria *not met* is self-
contradictory, and nothing anywhere checks for that today. A reviewer's own
inconsistency is exactly the failure that would let a caught defect through
while every count still looks healthy.

## Input format — extending brief 0003

Brief 0003's *Input format* section gains one directory. Everything there still
holds; this adds to it.

```
<run-id>/
  sessions/*.json           (0003)
  iterations.jsonl          (0003)
  reports/NNN-verdict.json  ← this brief
  reports/NNN-proposal.json ← present, deliberately NOT read (see Out of scope)
```

**`reports/NNN-verdict.json`** — one per reviewed iteration, `NNN` matching the
iteration number:

```json
{
  "task": "T3",
  "verdict": "PASS",
  "criteria": [
    {"criterion": "verbatim text from the task's acceptance list",
     "met": true,
     "evidence": "where the reviewer saw it"}
  ],
  "findings": [],
  "notes": "none"
}
```

`verdict` is `PASS` or `FAIL`. `findings` is a list of one-line strings, empty
on a pass. Not every iteration has a verdict: a gate failure skips review
entirely, so `reports/` is legitimately sparser than `iterations.jsonl`.

## The command

### `runstat review <run-dir>`

A per-iteration table and a totals row:

| column | meaning |
| --- | --- |
| iteration | from the filename |
| task | `task` |
| verdict | `PASS` / `FAIL` |
| criteria | how many acceptance criteria were ruled on |
| not met | how many were ruled `met: false` |
| findings | length of `findings` |
| evidence | criteria carrying non-empty `evidence`, over the total |

Then **every finding's text**, grouped by iteration. The findings are the point;
a count of them is not.

Then the coherence checks below, and — when a run recorded no findings at all —
a line saying so plainly, because that is the result a reader most needs to not
miss.

## Coherence checks

Each is a statement about one verdict that is either true or false. Report every
violation, with the iteration and task:

1. **A `PASS` with any criterion `met: false`** — the verdict contradicts its
   own rulings.
2. **A `PASS` with a non-empty `findings` list** — findings are what a `FAIL`
   is made of; on a pass they are a defect routed through a non-blocking
   channel, which is indistinguishable from no catch at all.
3. **A `FAIL` with no criterion `met: false` and no findings** — a rejection
   with no stated reason.
4. **Any criterion with empty `evidence`** — a ruling with nothing behind it.
5. **A verdict whose `task` disagrees with the iteration's task** in
   `iterations.jsonl`, where that iteration exists.

Violations do **not** change the exit code. They are reported, not enforced:
this command describes a finished run, and a run cannot be un-run.

## Behaviour contract

Inherits brief 0003 exactly. **Exit codes:** `0` success · `1` the run directory
is valid but has no verdicts · `2` usage error, missing directory, or malformed
input. Errors on stderr, stdout empty on failure, no traceback.

**A malformed verdict is a hard failure, never a silent skip** — same reasoning
as brief 0003, and it matters more here: a summary of reviews that quietly
omitted one would understate exactly the thing being measured.

**A missing `reports/` directory is not malformed** — it is a run with no
verdicts, so exit 1.

## Worked example

**Fixture run `20260817-120000`** — `reports/` holds three verdicts:

| file | task | verdict | criteria | not met | evidence present | findings |
| --- | --- | --- | --- | --- | --- | --- |
| `001-verdict.json` | T1 | PASS | 3 | 0 | 3 | 0 |
| `002-verdict.json` | T2 | FAIL | 3 | 1 | 3 | 2 |
| `003-verdict.json` | T2 | PASS | 3 | 0 | 3 | 0 |

`runstat review` reports these totals, exactly:

```
reviews:            3
passed:             2
failed:             1
criteria ruled:     9
criteria not met:   1
findings:           2
evidence cited:     9/9
coherence:          ok
```

and prints both of `002`'s findings verbatim under a heading naming iteration 2
and task T2.

**A second fixture** carries one incoherent verdict — a `PASS` whose second
criterion is `met: false` — and its report names check 1, the iteration and the
task, while still exiting 0.

Assert on content, not on column alignment: the totals block is `key: value`
lines, so assert its values exactly; the per-iteration table's layout is the
implementation's choice.

## Out of scope

- Reading `reports/NNN-proposal.json`. This command is about the reviewer, and
  the work session's self-report is the thing a reviewer is explicitly told to
  read last and trust least.
- Judging whether a finding was *correct* — that is not mechanizable here.
- Changing any existing command's output. `summary`, `signals` and `compare`
  must behave exactly as they do now.
- Cross-run review comparison. `compare` stays about signals.

## Constraints

- Standard library only at runtime; `pytest` the sole dev dependency.
- Reuse the existing loader and error types rather than adding a parallel way to
  read a run. The whole point is one tool with one contract.
- Every existing test must still pass, unchanged.
- `docs/runstat.md` and `docs/runstat-cli.md` both describe the telemetry
  contract and the commands; both must cover `review` and the `reports/`
  directory when this is done, and their examples must match real output.

## Shape

Four to six tasks. Smaller than brief 0003 because the loader, the error
contract, the CLI skeleton and the fixture machinery all already exist — the
decomposition should extend them, not rebuild them.

The verify commands still have to be authored before the implementation, still
have to fail now for the right reason, and now have one extra obligation: at
least one must prove the existing commands are unchanged.

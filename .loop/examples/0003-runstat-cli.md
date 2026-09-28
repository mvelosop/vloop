# Brief 0003 — `runstat`, a run-telemetry CLI

- **Status:** ready to plan
- **Consumed by:** the loop defined in `docs/briefs/0002-next-generation-autonomous-loop.md`
- **Role:** the first real target the loop builds. The program is the proof the loop ran.

---

## What it is

A command-line tool that reads a completed loop run's telemetry off disk and
reports what happened: what each phase cost, whether the run was converging, and
how two runs compare.

The loop's first job is to build its own instrument.

**It has no runtime role in the loop.** The driver computes its own signals
inline while a run is in flight; `runstat` is strictly post-hoc analysis over
archived runs. Nothing in the driver may import it or shell out to it. This is
deliberate — a demo target that the loop depended on mid-run would be circular,
and a broken demo would take the loop down with it.

## Why this target

- **Real gates.** Fixed input files, exact expected output. Nothing fuzzy to verify.
- **Fast and hermetic.** Standard library only, no network, no services; the
  whole suite runs in seconds, so iterations stay cheap.
- **Different in kind** from `exploring-claude`'s long-running HTTP server —
  which is the point of picking it
  ([`executable-loop-harness`](../references/executable-loop-harness.md) Rule 6:
  the useful next data point comes from a repo of a different kind, because
  "prove the artifact runs" is what a single-stack design gets wrong).
- **Genuinely useful afterwards.** It computes acceptance item 7 of brief 0002
  — whether a second run reproduces the first's behaviour.

---

## Input format

This is the **one place this brief constrains brief 0002**: the driver must emit
this layout, and `runstat` must read exactly it. Everything else about the
driver stays open.

A run directory contains:

```
<run-id>/
  sessions/*.json      one per claude session, name-sortable
  iterations.jsonl     one JSON object per completed iteration
```

**`sessions/*.json`** — the raw `claude -p --output-format json` result, with two
keys added by the driver:

| Key | Meaning |
| --- | --- |
| `phase` | `"plan"`, `"work"` or `"review"` |
| `iteration` | integer; `0` for the plan phase |

The fields `runstat` reads from the result itself: `total_cost_usd`,
`num_turns`, `duration_ms`, `is_error`, `permission_denials`.

**`iterations.jsonl`** — one object per line:

```json
{"iteration": 2, "task": "T2", "outcome": "gate_fail", "attempts": 1, "tasks_done": 1, "tasks_total": 8}
```

`outcome` is one of `done`, `gate_fail`, `review_fail`, `blocked`.

---

## Commands

### `runstat summary <run-dir>`

Per-phase rollup: session count, total cost, total turns, total wall time, plus
a total row. Any session with `is_error` true, or a non-empty
`permission_denials`, is called out — a denial the operator never sees is a
fence in the wrong place that nobody knows about.

### `runstat signals <run-dir>`

The run-level signals from brief 0002 §7, as `key: value` lines:

| Signal | Derivation |
| --- | --- |
| iterations | count of `iterations.jsonl` records |
| tasks closed | `tasks_done` / `tasks_total` from the last record |
| iterations per closed | iterations ÷ tasks closed, 2 decimals; `n/a` when zero closed |
| gate failures | records with `outcome: gate_fail` |
| review rejections | records with `outcome: review_fail` |
| attempts burned | records whose `outcome` is not `done` — **not** the sum of the `attempts` field, which is a cumulative per-task counter and would double-count a task that appears in several records |
| no-progress streak | trailing records whose `tasks_done` did not increase |
| estimated spend | sum of `total_cost_usd` across all sessions |

These derivations are not free choices. The driver computes the same signals
inline during a run, and brief 0002's acceptance item 6 requires the two to
agree exactly on a completed run — so a formula that is merely reasonable, but
different, is a defect. The fixture below is the arbiter for all of them.

### `runstat compare <run-dir> <run-dir>`

Every signal above for both runs side by side, with a delta column. Numeric
signals get a signed delta; non-numeric ones get a blank. Exits 0 when both runs
are readable.

This is the repeatability check. Two runs of the same plan should show
comparable signals; a large divergence is the finding.

---

## Behaviour contract

**Exit codes.** `0` success · `1` the run directory is valid but has no sessions
· `2` usage error, missing directory, or malformed input.

**A malformed record is a hard failure, never a silent skip.** One unparseable
session file or `iterations.jsonl` line exits 2 with the offending path on
stderr. Skipping it would produce a report that looks complete and undercounts —
and a partial result indistinguishable from a correct one is exactly the failure
mode the loop's own design notes were written about. Better a loud refusal than
a plausible wrong number.

**All errors go to stderr, stdout stays empty on failure, and no traceback ever
reaches the user.**

**Money is labelled an estimate.** `total_cost_usd` is not a bill, especially on
a subscription plan. Any dollar figure the tool prints says so somewhere.

---

## Worked example

This transcript is the end-to-end acceptance test — the one test that catches an
implementation whose parts each pass in isolation while the whole drifts from
the documented behaviour.

**Fixture run `20260814-101500`.** Seven sessions:

| file | phase | iteration | `total_cost_usd` | `num_turns` | `duration_ms` |
| --- | --- | --- | --- | --- | --- |
| `001-plan.json` | plan | 0 | 1.98 | 12 | 141000 |
| `002-work.json` | work | 1 | 0.50 | 6 | 68000 |
| `003-review.json` | review | 1 | 0.20 | 3 | 24000 |
| `004-work.json` | work | 2 | 0.52 | 7 | 72000 |
| `005-review.json` | review | 2 | 0.20 | 3 | 24000 |
| `006-work.json` | work | 3 | 0.48 | 5 | 64000 |
| `007-review.json` | review | 3 | 0.20 | 3 | 24000 |

All have `is_error: false` and empty `permission_denials`.

Three `iterations.jsonl` records:

```json
{"iteration": 1, "task": "T1", "outcome": "done",      "attempts": 0, "tasks_done": 1, "tasks_total": 8}
{"iteration": 2, "task": "T2", "outcome": "gate_fail", "attempts": 1, "tasks_done": 1, "tasks_total": 8}
{"iteration": 3, "task": "T2", "outcome": "done",      "attempts": 1, "tasks_done": 2, "tasks_total": 8}
```

**`runstat summary`** reports, per phase and in total:

| phase | sessions | cost | turns | wall |
| --- | --- | --- | --- | --- |
| plan | 1 | $1.98 | 12 | 141s |
| work | 3 | $1.50 | 18 | 204s |
| review | 3 | $0.60 | 9 | 72s |
| **total** | **7** | **$4.08** | **39** | **417s** |

**`runstat signals`** reports exactly these values:

```
iterations:            3
tasks closed:          2/8
iterations per closed: 1.50
gate failures:         1
review rejections:     0
attempts burned:       1
no-progress streak:    0
estimated spend:       $4.08
```

**Assert on content, not on column alignment.** Table layout, padding and column
widths are the implementation's choice; the numbers, labels and their pairing
are the contract. `signals` output is `key: value` lines, so assert its values
exactly.

**`compare`** is exercised against this fixture and a variant of it — the
variant is the implementation's to construct, but the delta of at least one
numeric signal must be asserted.

---

## Documentation

Two documents, both gated on being true rather than on existing:

- **`docs/runstat-cli.md`** — what runstat is, install and run with `uv`, each of the three
  commands with a worked example, the exit-code contract, and the input layout
  it reads. **Every documented example must show output the tool actually
  produces**, and the gate replays the tool and requires each printed line to
  appear verbatim. A usage doc whose examples have drifted is worse than none,
  because a reader trusts it.
- **`docs/runstat.md`** — the telemetry contract, field by field, and each
  signal's derivation. This is the document that keeps the loop driver and
  `runstat` from drifting: brief 0002 acceptance item 6 requires the two to
  agree, and a derivation recorded in only one implementation is how that
  agreement quietly breaks. It must state the attempts-burned trap explicitly.

## Out of scope

Deliberately excluded. A plan that adds any of these has widened its own scope,
which is itself worth knowing:

- Machine-readable output (`--json`, CSV export).
- Reading anything from a live, in-flight run.
- Trend analysis across more than two runs.
- Any dependency outside the standard library.
- Reading Claude session transcripts (as opposed to result JSON).

## Constraints

- Python, `uv`, `src/` layout, `pytest`. Standard library only at runtime;
  `pytest` is the sole dev dependency.
- A console script entry point, and `python -m runstat` works too.
- Every test writes inside `pytest`'s `tmp_path`. Nothing touches a real run
  directory.
- Repo-relative paths in every file, log line and commit message.

## Shape

This should decompose into roughly nine to eleven tasks, each independently
verifiable by a single command. The decomposition itself is the planning phase's
job and is not prescribed here — how well it does that is part of what this run
is testing.

Two things the planner should get right, because they are what the priors got
wrong: a task that only scaffolds still needs a verify command that passes for
the right reason (`uv run pytest` with zero tests collected exits **5**, not 0),
and each task's verify command has to be authored before its implementation
exists.

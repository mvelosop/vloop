---
name: B20261126-1003-b41-i161-doctor-gate-shell-replaced-plan.loop-brief
description: Make doctor's gate shell row judge the configured shell, not the old plan's, when the plan on disk is one the next brief's plan replaces, so vloop run no longer refuses for a shell the run will not use
kind: brief
status: consumed
created: 2026-11-26
seeds: Issue #161, from B40's start on Windows (recorded as a defect attributed to B15, and as the intervention that put sh on the run's PATH); the operator asked on 2026-11-26 for it to be fixed autonomously, recommended options taken
depends-on: [B20261125-2207-b40-i156-mixed-os-pwsh.loop-brief]
---
# Brief — vloop B41: doctor's gate shell row and a replaced plan

- **Starting point:** extends `main` at `87536f4e`, with this brief's ready
  commit on the run branch. The planner pins that SHA as the **base** every
  gate compares against — never `HEAD~n`.
- **Produced by:** operator decision, 2026-11-26: issue #161 is fixed by a
  vloop run, with the open decisions settled by the operator's hands on the
  recommended options and recorded under "Why this shape".
- **Review:** `vloop brief review` (Copilot, mai-code-2). Round 1: `FIX`, 2
  findings, fixed: the gate covers every row of the table; a closing task.
  Round 2: `FIX`, 1 finding, fixed: the worked example runs as a gate and in
  the operator's acceptance, since the verify fence denies `vloop init`.
  Round 3: `FIX`, 2 findings, fixed: every non-complete status listed, and the
  unchanged rule limited to `state/v2`. Round 4: `PASS`, 0 findings.
- **Run environment:** planned and run on Windows, in pwsh, every session on
  GitHub Copilot (`.vloop/config.local.toml`). See Constraints.

## Executive summary

After a brief closes, its complete plan stays in `.vloop/state/state.json`
until the next brief's plan replaces it. Doctor's `plan` row knows that: it
passes with `a complete plan of <brief> — the next brief's plan replaces it`.
The `gate shell` row does not: it still checks that old plan's `shell`. When
B40 started on Windows, B39's complete plan said `"shell": "sh"`, and the
preflight refused with `✗ gate shell the plan's shell "sh" is not on PATH`,
although the new plan would be written for the configured `pwsh`.

**The goal:** doctor and `vloop run`'s preflight judge the shell the next
gates will actually run in.

- **G1 — a replaced plan's shell is not checked.** When the plan row says the
  plan is replaced, the `gate shell` row judges the configured shell exactly
  as it does with no plan.

Unchanged: a plan that will still run (any status but `complete`, or a
complete `state/v2` plan when `vloop run` names its own brief or no brief)
keeps its recorded shell (P-5), with the same messages.

## What it is

One rule in `vloop doctor` (and so in `vloop run`'s preflight, which runs the
same checks): the `gate shell` row follows the `plan` row's decision about
whether the plan on disk is replaced. It is for anyone who changes `shell`
between briefs, as a mixed team does when it moves to pwsh.

Why it got through: the `plan` row's "replaced" rule (B31) was added to the
`plan` and `branch` rows only, and every `gate shell` test uses a plan whose
shell is on PATH or a plan that will still run.

## Why this shape, and what was rejected

Decided on 2026-11-26. **Do not re-litigate these.**

- **One "replaced" rule, shared by both rows.** The plan is replaced exactly
  when the `plan` row prints `a complete plan of <brief> — the next brief's
  plan replaces it` or `a complete state/v1 plan — the next brief's plan
  replaces it`. The `gate shell` row uses that same decision, so the two rows
  can never disagree.
  *Rejected:* a separate test in the `gate shell` row (two rules drift);
  dropping the `gate shell` check for complete plans altogether (the bare
  doctor would then say nothing about the configured shell).
- **A replaced plan is judged as no plan.** The `gate shell` row then judges
  the configured shell (`vloop config get shell`), with B40's rules unchanged:
  pass when it is on PATH; a problem `the configured shell "<shell>" is not on
  PATH` on a mixed team; a warning with the same text otherwise.
  *Rejected:* checking both the old plan's and the configured shell (the old
  plan's gates never run again).

## Binding references

- `docs/domain/domain-model.md` — P-5: the plan's shell governs its gates; a replaced plan has no gates left to run
- `docs/domain/execution/plan.md` — the plan and its `shell`, one plan per repository, replaced by the next brief's
- `docs/domain/platform/platform-context.md` — the `shell` and `platforms` keys and the mixed-team rule the configured shell is judged by
- `docs/guide/concepts.md` — "Gates and the gate shell" and the doctor rows it lists, the guide text this run updates
- `docs/briefs/B20261125-2207-b40-i156-mixed-os-pwsh.loop-brief.md` — M4: the `gate shell` row's messages and its mixed-team rule, unchanged here
- `docs/briefs/B20261102-1844-b31-i110-i112-plan-phase-edges.loop-brief.md` — the "replaced" rule for a complete plan of another brief, which this run extends to the `gate shell` row

## Behaviour contract

### G1 — the gate shell row and a replaced plan

**Defect.** `internal/cli/doctor.go`, `doctorPlanShell`, takes the plan's
recorded `shell` whenever a plan exists. The `replaced` decision it computes
for the `plan` row is not applied to the shell, so a complete plan of another
brief whose shell is not on PATH is `✗ gate shell the plan's shell "<shell>"
is not on PATH`, and `vloop run <next brief>` refuses at preflight.

**Required.** When the plan is replaced (as defined under "Why this shape"),
the `gate shell` row is computed as if there were no plan: from the
configured `shell`, with B40's three outcomes and messages. In every other
case it is unchanged: the plan's recorded shell, `✓` when on PATH, else
`✗ gate shell the plan's shell "<shell>" is not on PATH`.

The cases, exhaustively:

| plan on disk | doctor run as | `gate shell` judges |
|---|---|---|
| none | any | the configured shell |
| complete `state/v2` plan of brief A | bare `vloop doctor` | the configured shell |
| complete `state/v2` plan of brief A | `vloop run <brief B>` preflight | the configured shell |
| complete `state/v2` plan of brief A | `vloop run <brief A>` or `vloop run` with no brief | the plan's shell |
| complete `state/v1` plan | any | the configured shell |
| a plan whose status is not `complete` (`planning`, `running`, `blocked`, `stalled`, `halted`) | any | the plan's shell |

`--json` carries the same result and message. Nothing else in doctor's output
changes.

**Gate.** A test that fails before the fix: a repository whose
`.vloop/state/state.json` is a complete `state/v2` plan of another brief, with
a `shell` that is not on PATH, and a configured shell that is; the bare
doctor's `gate shell` row must pass, and the preflight for a different brief
must not report it. The tests cover every row of the table, each with the
plan's shell absent from PATH:

- no plan, a replaced complete `state/v2` plan (bare doctor, and a run naming
  another brief) and a complete `state/v1` plan: `gate shell` passes while
  the configured shell is on PATH; with the configured shell absent too, it is
  the warning `the configured shell "<shell>" is not on PATH`, or the problem
  with that text when `platforms` lists two or more;
- a complete `state/v2` plan with a run naming its own brief or no brief, and
  a plan in each status but `complete` (`planning`, `running`, `blocked`,
  `stalled`, `halted`): the problem
  `the plan's shell "<shell>" is not on PATH`, as today.

### Decided here, because the cited documents leave it underdetermined

- **"On PATH"** is what doctor already uses (`exec.LookPath` on the shell's
  name). A test controls PATH so that the plan's shell is absent and the
  configured one present; it does not depend on what the machine has
  installed.
- **The guide.** `docs/guide/concepts.md` gains one sentence in "Gates and the
  gate shell", or where it lists doctor's rows: for a plan the next brief
  replaces, doctor checks the configured shell. No other document changes.

### Violations the review must rule on

- The `plan` and `gate shell` rows computing "replaced" separately, in two
  places that could disagree.
- A changed message, or a changed result in any row of the table marked "the
  plan's shell".
- An earlier test edited for anything but the new behaviour.

## Worked example

In a scratch repository, `vloop init` run, `shell = "pwsh"` in
`.vloop/config.toml`, PATH holding `pwsh` and `git` but not `bash`, and
`.vloop/state/state.json` a complete `state/v2` plan of
docs/briefs/B20260101-0900-a.loop-brief.md with `"shell": "bash"`:

```
vloop doctor
  -> ✓ gate shell                                                 (row)
  -> ✓ plan a complete plan of docs/briefs/B20260101-0900-a.loop-brief.md — the next brief's plan replaces it

vloop doctor --json
  -> {"name":"gate shell","result":"pass","message":""}           (in checks)

the same, with "status": "running" in state.json
  -> ✗ gate shell the plan's shell "bash" is not on PATH          exit 1

the complete plan again, with PATH holding neither pwsh nor bash,
platforms unset
  -> ! gate shell the configured shell "pwsh" is not on PATH      (warning)
```

The preflight of `vloop run docs/briefs/B20260101-0901-b.loop-brief.md` on the
complete plan reports no `gate shell` problem; with
vloop run docs/briefs/B20260101-0900-a.loop-brief.md it reports
`✗ gate shell the plan's shell "bash" is not on PATH` and refuses.

## Acceptance, at close (the operator's)

The operator runs the worked example's lines by hand with the built CLI in a
scratch repository, since the verify session's fence denies `vloop init`;
each must print exactly what the worked example shows.

## Out of scope

- Any other doctor row, message or order.
- The planner's or the driver's handling of a replaced plan.
- Removing the operator's PATH workaround from B40's notes; that is the
  operator's after the merge.
- The fence and pwsh (#162), and the documentation pass that follows it.

## Constraints

- Go, as pinned in `go.mod`; `go run ./tools/check` is the `[[check]]` and
  passes after every task.
- **This run is on Windows.** Most `internal/cli` doctor tests and 151 of the
  `cmd/vloop` end-to-end tests skip there (their stub agents are POSIX
  scripts), so a gate that runs only those proves nothing on this machine.
  Gate through something that runs on Windows: the built CLI's
  `vloop doctor --json` in a temporary repository, or an oracle test in the
  task's gate folder that calls the package without the stub agents. New
  tests the work adds must run on Windows too, or say in their skip why not.
- **Gates follow the diff surface.** No gate runs the whole suite or the
  check (P-8).
- **No task may weaken a check to pass.** An earlier test changes only where
  it pins behaviour the contract changes, and only to the new behaviour.
- Tests build their repositories in temporary directories and address them
  with `git -C`. Sessions never move refs.
- Repo-relative paths everywhere.

## Shape

2 to 3 tasks:

1. **G1 in doctor, with its tests.** Gated on every row of G1's table,
   through the built CLI or an oracle in the gate folder.
2. **The guide sentence**, if it does not fit task 1. Gated on the sentence
   being present in `docs/guide/concepts.md`.
3. **Close:** every earlier gate at once, and the worked example line by line,
   as a gate: an oracle in the task's gate folder that builds the CLI and runs
   each line in a temporary repository. Gates run outside the sessions'
   fences; the verify session reads the gate's result and need not run the
   lines itself.
<!-- vloop:run-record:begin -->
## Run record

Generated by vloop brief close on 2026-11-26. Recompute with: vloop metrics B20261126-1003-b41-i161-doctor-gate-shell-replaced-plan.loop-brief

```
B20261126-1003-b41-i161-doctor-gate-shell-replaced-plan  consumed · not merged
 tasks     3 planned (brief said 2–3) · 3 done · 0 blocked · first-pass 2/3
 size      delivered  code 23 · test 111 · docs 1 · test:code 4.83
           churn      code 23 · test 132 · docs 1 · rework 1.00
 time      agent 13.7 min (work 7.6 · review 6.0) · plan 6.8 min · gate review 2.0 min · verify 4.7 min · gates 0.3 · wall 37.1 min
 rate      1.7 code lines/min · 9.8 incl. tests
 cost      ≈$1.03 · plan 0.31 · work 0.27 · review 0.18 · gate review 0.10 · verify 0.16 · $44.71/kloc
 models    plan mai-code-2 · work mai-code-2 · review mai-code-2 · gate review mai-code-2 · verify mai-code-2
 defects   in-loop 2 · operator 1 · escaped 0 · removal efficiency 100%
 interventions  1 · recommended 0 · other-option 0 · adjusted 0 · different 0 · no-options 1
 instructions  plan 4,881 · gate review 2,420 · work 2,487 · review 2,508 · verify 2,161 words · guidelines 2 rules
```

```
id  area  kind  att  code+  test+  docs+  other+  test:code  $/kloc  agent  cost   model
T1  -     fix   2    23     132    0      0       5.74       12.58   8m25s  $0.29  mai-code-2
T2  -     docs  1    0      0      1      0       -          -       1m56s  $0.06  mai-code-2
T3  -     test  1    0      0      0      0       -          -       3m20s  $0.11  mai-code-2
```

- B20261126-1003-b41-i161-doctor-gate-shell-replaced-plan/i1-review-1 — bug, work, found by review: A complete state/v1 plan now takes the generic replaced-plan branch and prints `a complete plan of  — the next brief's plan replaces it` instead of the existing `a complete state/v1 plan — the next brief's plan replaces it`; G1 explicitly requires no other doctor output to change.
- B20261126-1003-b41-i161-doctor-gate-shell-replaced-plan/i1-review-2 — gate-gap, plan, found by review: The gate and committed tests cover state/v1 only as a gate-shell row and do not assert the plan row's existing state/v1 message, so this output regression passes.
- I20261126-110826-vloop-run-s-preflight-now-also-prints-th — verification-finding, work, found by operator: vloop run's preflight now also prints the gate shell warning (a configured shell missing on PATH), which the brief did not ask for; harmless and useful, accepted as is
<!-- vloop:run-record:end -->

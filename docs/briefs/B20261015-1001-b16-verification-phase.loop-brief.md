---
name: B20261015-1001-b16-verification-phase.loop-brief
description: Add a verification phase — one /vloop:verify session after the final pass that checks the worked example, the contract and every surface stating a changed behaviour, its findings shown at close — clear .vloop/tmp so no session's scratch reaches a check, let an earlier test change only where the contract changes what it pins, and retire brief check's pinning-test and external-format warnings
kind: brief
status: consumed
created: 2026-10-15
seeds: The roadmap's B16 row, filtered at design time on 2026-10-15; vloop release follows as B17
depends-on: [B20261014-1138-b15-code-health-escapes.loop-brief]
---
# Brief — vloop B16: the verification phase

- **Starting point:** extends `main` after this design merges (`main` at
  `f5e2bc8`, B15 merged, v0.10.0 released and installed). The planner pins the
  base it plans from as the **base** every gate compares against — never
  `HEAD~n`.
- **Produced by:** the roadmap's B16 row
  (`docs/design-notes/vloop-roadmap-b12-b17.md`), the T1 triage's P6, and the
  design act of 2026-10-15, with the forks settled with the operator
  (recorded as this brief's design intervention).
- **Run by:** `vloop run` from the latest released vloop (0.10.0), whose
  skills predate K1 — see Constraints.

## Executive summary

The loop checks each task on its own; whether the brief as a whole does what it
says is found only by the operator, after the run, by hand — B15's verification
found what no gate or review could. Separately, B15's first task burned three
attempts on a file a plan session left in `.vloop/tmp/`, and the briefs' lists
of tests that must change missed tests five times.

**The goal:** every run ends with an independent check of the whole brief
against the built result, whose findings the operator reads at close; no
session's scratch can fail a later check; and a brief no longer has to guess
which earlier tests its contract will change.

- **V1 — a verify session ends every complete run.** After the final pass, one
  session checks the worked example, the contract's sentences and every surface
  that states a changed behaviour, and leaves a findings report in the run
  folder. On by default; `run.verify = false` turns it off.
- **V2 — the `/vloop:verify` skill and its eval.** The skill says what to check
  and how to report it; one eval case with planted mismatches measures it.
- **V3 — the metrics show the verify session** in the time, cost, models and
  instructions lines, as they show the gate review.
- **V4 — close shows the findings first.** `vloop brief close` (and its
  `--dry-run`) prints the verify findings before anything else, so the operator
  states findings or `--no-findings` having read them.
- **S1 — `.vloop/tmp/` is cleared** at run start and before every work session,
  the final pass and the verify session, all but the run's lock. A session's
  scratch never outlives the next session.
- **K1 — an earlier test changes only where the contract changes what it
  pins**, and only to the new behaviour; the review cites the item for each.
  Briefs stop listing tests that must change.
- **K2 — `brief check` drops its pinning-test and external-format warnings**,
  and `init`'s guidelines file loses their two sections: they fired for no
  repository and sent tool rules to every session.
- **D1 — the docs, the domain and both operate playbooks** describe the new
  phase and rules.

Unchanged: the per-task review, every exit code, the driver not acting on what
the verify session finds, existing records and run folders.

## What it is

A run that completes is followed by one more session, independent of every
session before it, that holds the brief against the result: each worked-example
line run on a fresh build, each contract sentence that states an observable
behaviour, and each surface (help, guides, domain pages, README, messages) that
states a behaviour the brief changed. What it finds is a report, not a verdict:
the run's ending does not change, and the operator decides at close what each
finding is.

Around it, two clean-ups that make runs cheaper and briefs simpler: the scratch
folder is emptied between sessions, and the brief's must-change test list
becomes a rule every session already holds.

## Why this shape, and what was rejected

Decided with the operator on 2026-10-15. **Do not re-litigate these** — the
planner inherits them.

- **Filtered at design time.** The row's candidates were weighed against what
  costs end users defects: the verification phase, the scratch clearing and
  contract-licensed tests go in. *Rejected:* **pruning rules** by eval runs
  with and without each rule — new eval infrastructure for two guidelines rules
  and no evidence one is dead; a later brief, if rules grow.
- **On by default.** Finding functional errors sooner is the point, and it
  costs one session per run, not per task. *Rejected:* opt-in (new users would
  never get it).
- **A report, not a gate.** The driver does not act on findings and the run's
  exit code does not change: a finding may be the brief's own error, and only
  the operator can tell (T1 decision 2: a phase in the loop, not the
  operator's step). *Rejected:* failing the run on a finding.
- **The per-task review stays** (T1 decision 3) until the verify phase shows
  what it catches.
- **Clear before every work session**, not only after plan acceptance: the same
  failure can come from a work or review session's scratch. *Rejected:* the
  roadmap's two points alone; the check ignoring `.vloop/tmp/` (the check is the
  repository's own command).
- **L4 retires** rather than moving to `config.toml`: no repository uses it, and
  a later brief can bring it back as config. L3 retires with K1, which makes
  its must-change lists unnecessary.

## Binding references

- `docs/design-notes/vloop-t1-triage.md` — "P6": the verification phase's design input, its three passes and its findings table
- `docs/domain/domain-model.md` — S-1, S-2 and S-4 (the verify session is a session like any other, writing only under .vloop/tmp/), R-1 and R-3 (the run commit carries the report; no exit code changes), P-6 (model and effort per session kind), M-6 (close states findings or --no-findings), C-3 (the report's schema name)
- `docs/domain/execution/session.md` — the session kinds, the fence per phase and the session/v2 record, which V1 extends with a fifth kind
- `docs/domain/execution/run.md` — the run folder, the final pass and the run's end, where V1 runs and S1 clears
- `docs/domain/measurement/metrics.md` — the summary lines and the metrics/v2 JSON V3 extends
- `plugin/evals/README.md` — the eval case layout V2's case follows
- `plugin/skills/review/SKILL.md` — the goalpost rule K1 changes
- `plugin/skills/operate/SKILL.md` — the close step V4 changes
- `.vloop/interventions/I20261014-1823-t1-reset-after-the-check-failed-on-the-p.md` — S1's evidence: B15's T1 and its three check failures

## Behaviour contract

Messages are exact where quoted; `<…>` marks a value; `<run dir>` is the run
folder, repo-relative, ending in `/`.

### V1 — a verify session ends every complete run

**Required.**
- `run.verify` is a config key, `true` or `false`, default `true`; `model.verify`
  (default `sonnet`) and `effort.verify` resolve like gate review's, with no
  per-task override (P-6). `vloop config list` shows the three.
- When every task is done and the final pass passes, and `run.verify` is
  `true`, the driver runs one session of phase `verify` with `/vloop:verify`
  and no argument, before the run commit. A resume with every task done repeats
  the final pass, then the verify session.
- It runs under its own fence, which allows what the review fence allows; it
  writes nothing under `.vloop/` but `.vloop/tmp/` (S-4). What it changes
  outside `.vloop/tmp/` is reverted as a review session's is, with
  `vloop: the verify session changed <path> — reverted` per path, and its
  report is still kept; a change the driver cannot revert halts the run as it
  does for a review session. Refs, `.git/config` and hooks are guarded as for every session
  (exit 9).
- It writes .vloop/tmp/verify.json as `verify/v1`: `schema`, `findings` (each
  `{kind, claim, observed, where}`, `kind` one of `example`, `contract`,
  `surface`), `checked` (`{example_lines, sentences, surfaces}`, integers) and
  `notes`. schemas/verify.v1.json defines it; `vloop schema validate verify/v1`
  accepts it.
- The driver keeps a valid report as <run dir>reports/verify.json and prints
  `verify: <n> finding(s) — see <run dir>reports/verify.json`, or
  `verify: no findings` when the list is empty.
- A missing or invalid report: `vloop: the verify session left no valid .vloop/tmp/verify.json — verify by hand`.
  A session that fails to run or times out:
  `vloop: the verify session failed to run — verify by hand; see <run dir>run.log`.
  With the cost ceiling already reached: `vloop: cost ceiling reached — verify skipped`
  and no session. Each is a warning: the run ends `complete`, exit 0.
- A finding never changes the run's ending, its exit code or any task.
- Its record is `sessions/NNN-verify.json`, `session/v2`, iteration 0, no task;
  `schemas/session.v2.json`'s `phase` gains `verify`, which has no task, as
  `plan` and `gate-review` have none. Its cost counts toward the run's spend.
- `vloop status` names the phase `verify` while the session runs.
- With `run.verify = false`, nothing of the above happens and nothing is
  printed for it.

### V2 — the `/vloop:verify` skill and its eval

**Required.**
- `plugin/skills/verify/SKILL.md` makes three passes, from the brief, the
  whole diff against the plan's base and a fresh build of the result:
  1. each line of the worked example, run as written in a scratch directory,
     its output and exit code compared with the brief's (`example`);
  2. each sentence of the Behaviour contract that states an observable
     behaviour, checked by running or reading the result (`contract`);
  3. each surface that states a behaviour the brief changed — help text,
     guides, domain pages, README, messages — read against the new behaviour
     (`surface`).
- A finding quotes what the brief or the surface says (`claim`), what it
  observed (`observed`) and where (`where`: the brief's section or example
  line, and the file and line or the command). It does not judge whose defect
  it is; the operator does at close.
- It changes nothing outside `.vloop/tmp/`, builds into a scratch directory,
  and validates its report with `vloop schema validate verify/v1 .vloop/tmp/verify.json`.
- The skill's word ceiling is 1,500; the budget's skill list gains `verify`.
- plugin/evals/verify-planted-mismatches/ is one eval case in the layout of
  `plugin/evals/README.md`: a small built result whose brief has one sound
  worked-example line and three planted mismatches — a worked-example line
  whose exit code differs, a contract sentence the result does not do, a
  README that states the old behaviour. Its graders require one finding of
  each matching kind and none on the sound line. The README's "other cases"
  table gains its row.

### V3 — the metrics show the verify session

**Required.**
- Where the brief's runs hold a verify session, the summary adds it after gate
  review, as gate review's is added: ` · verify <m> min` on the time line,
  ` · verify <c>` on the cost line, ` · verify <models>` on the models line.
  The total cost includes it; `agent` does not (it is per run, as gate
  review's is).
- The instructions line adds ` · verify <n>` after `review <n>` when the tree
  holds the verify skill, and leaves it out otherwise.
- `--json` gains `time.verify_ms`, `cost.verify_usd`, `models.verify`,
  `effort.verify` and `instructions.verify`; `schemas/metrics.v2.json` gains
  them as optional properties. No schema version changes.
- A brief with no verify session prints exactly what it prints today.

### V4 — close shows the findings first

**Required.**
- `vloop brief close` and `vloop brief close --dry-run` print, before any other
  output, the report of the newest of the brief's run folders that holds
  reports/verify.json:
  ```
  verify of <run id>/<folder>: <n> finding(s)
    1. <kind> — <where>
       brief:    <claim>
       observed: <observed>
  ```
  one block per finding, in the report's order; `verify of <run id>/<folder>: no findings`
  for an empty list; `no verify report — verify by hand` when no folder holds
  one.
- Close still requires `--finding` or `--no-findings` (M-6); a verify finding is
  recorded only when the operator passes it as a `--finding`.

### S1 — `.vloop/tmp/` is cleared

**Required.**
- The driver removes every entry under `.vloop/tmp/` except `.running`, the
  run's lock:
  - at run start, after taking the lock and before any session, extraction or
    acceptance;
  - before every work session (so after plan acceptance too), before the final
    pass and before the verify session.
- It is not cleared between the plan session's rounds or between a work session
  and its review: the driver's hand-offs there (plan-problems.md,
  proposal.json, checks.json) survive.
- Nothing relies on `.vloop/tmp/` surviving a resume: every file there is
  written for the session that follows within one run, and the plugin, the
  fence and the hooks directory are recreated before use. The task notes say so
  with the evidence.
- A symbolic link is removed, never followed. An entry that cannot be removed
  prints `warning: cannot clear .vloop/tmp/<entry>: <error>` and the run
  continues.
- No new command; `vloop plugin path` behaves as today.

### K1 — an earlier test changes only where the contract changes what it pins

**Required.**
- The brief templates (`en`, `es`) state, in Constraints beside "No task may
  weaken a check to pass": an earlier test changes only where it pins
  behaviour the contract changes, and only to the new behaviour.
- `plugin/skills/work/SKILL.md`: a work session may change an earlier test only
  so, and names the contract item in its proposal's notes for each such test.
- `plugin/skills/review/SKILL.md`: a changed earlier test is a goalpost move
  unless the contract changes the behaviour it pins and it now asserts the new
  behaviour; the verdict cites the item for each changed earlier test.
- `.claude/skills/vloop-architect/SKILL.md`: the bullet asking for an explicit
  list of earlier tests that must change is replaced by the rule.
- Each skill stays within its word ceiling.

### K2 — `brief check` drops two warnings

**Required.**
- `vloop brief check` no longer reads `## Pinning tests` or `## External formats`
  from `.vloop/repo-guidelines.md` and never prints `must-change list:` or
  `external format:` lines. A guidelines file that still has the sections is
  read as ordinary text: no error, no warning.
- `vloop init`'s guidelines file is the title and its one sentence, in `en` and
  `es`, without the two sections.
- `doctor`'s `guidelines` line is unchanged.

### D1 — docs, domain and playbooks

**Required.**
- `docs/domain/execution/session.md`: five kinds, the verify fence, the
  `verify/v1` report; `docs/domain/execution/run.md`: the verify session after
  the final pass, reports/verify.json, S1's clearing;
  `docs/domain/domain-model.md`: the On disk list gains .vloop/tmp/verify.json
  and says `.vloop/tmp/` is cleared between sessions;
  `docs/domain/measurement/metrics.md`: V3.
- `docs/domain/briefing/brief.md` and `docs/guide/concepts.md`: the two
  retired warnings and sections removed.
- `docs/guide/configuration.md`: the three keys; `docs/guide/metrics.md`: V3.
- `plugin/skills/operate/SKILL.md`: the close step reads the verify findings and
  records each one the operator accepts with `--finding "<origin>: <summary>"`;
  it stays within its word ceiling.
  `.claude/skills/vloop-operator/SKILL.md`: the same in its close step.
- `docs/guide/commands.md` is regenerated.

### Decided here, because the cited documents leave it underdetermined

- `model.verify` defaults to `sonnet`, as review and gate review do.
- The verify session runs even when the run ended complete on a resume; each
  run that ends complete has its own report, and close reads the newest.
- The cost ceiling skips the verify session but never ends the run otherwise.
- A finding has no origin: the operator gives one at close.

### Violations the review must rule on

- A verify finding that changes the run's ending, exit code or a task's status.
- A verify session allowed to write outside `.vloop/tmp/`, or its changes kept.
- `.vloop/tmp/.running` removed, or `.vloop/tmp/` cleared between a work session
  and its review or between plan rounds.
- An earlier test changed without the contract item that licenses it.
- A skill past its ceiling, or changed beyond V2, K1 and D1.

## Worked example

With a fresh build, in temporary repositories with a temporary `HOME`, stub
sessions standing in for `claude`:

```
vloop config list                    -> includes run.verify = true, model.verify = sonnet, effort.verify
a brief run to completion; the verify stub writes a valid report with 2 findings
  -> verify: 2 finding(s) — see .vloop/state/runs/<run id>/<folder>/reports/verify.json
  -> the run commit holds reports/verify.json and sessions/NNN-verify.json   exit 0
the same, the verify stub writes nothing
  -> vloop: the verify session left no valid .vloop/tmp/verify.json — verify by hand   exit 0
the same, the verify stub edits README.md
  -> vloop: the verify session changed README.md — reverted                  exit 0
the same, with run.verify = false   -> no verify session, no verify line     exit 0

vloop brief close <brief> --dry-run
  -> verify of <run id>/<folder>: 2 finding(s)
       1. example — Worked example, line 3; vloop greet --loud
          brief:    exit 2
          observed: exit 1
       2. surface — README.md:12
          brief:    greet prints in upper case with --loud
          observed: README.md says --loud is not supported
     … then close's dry-run output as today
a brief whose run had run.verify = false
  -> no verify report — verify by hand

vloop metrics <brief>
  -> time line: … · gate review <m> min · verify <m> min · gates …
  -> cost line: … · gate review <c> · verify <c> · $<c>/kloc

a plan stub that leaves .vloop/tmp/scratch/x.go unformatted, check gofmt -l .
  -> T1's check passes (cleared before T1's work session)
a stale .vloop/tmp/old.txt before vloop run   -> gone once the run starts; .vloop/tmp/.running held throughout
a review stub that reads .vloop/tmp/proposal.json and checks.json -> both present

vloop brief check <a ready brief>, guidelines with "## Pinning tests" rules
  -> no "must-change list:" line; the other lines as today
vloop init in a fresh repository     -> .vloop/repo-guidelines.md has no "## Pinning tests" or "## External formats"
```

Planted failures, each in a scratch copy of the tree:

```
the driver made to act on a verify finding (ending blocked)  -> V1's gate fails
.running removed by the clearing                             -> S1's gate fails
plugin/skills/verify/SKILL.md padded past 1,500 words        -> the budget test fails
```

**Real-data check**, on a clone of this repository:

- `vloop metrics` for every consumed brief: every line identical before and
  after (none has a verify session), recorded in the task notes.
- `vloop brief check` on every brief in `docs/briefs/`: the same lines before
  and after, but for `must-change list:` and `external format:` lines, which
  disappear.
- No verify session runs for real in this run: B16 is run by 0.10.0, which has
  none. The first real one is the run after the release that carries it.

## Out of scope

- **B17, `vloop release`.**
- Pruning rules by eval; moving L4 to config.
- Cutting or narrowing the per-task review.
- The verify session's findings recorded automatically, given an origin, or
  failing anything.
- A command to re-run the verify session alone.
- Changing any existing session, run folder, defect or intervention record.
- The v1 review's remaining findings and the long functions B15 left.

## Constraints

- Go as pinned in `go.mod`; no new dependencies. Every gate builds for
  `windows` and `linux` as well as the host.
- Gates run on macOS's BSD tools: no empty alternatives in `grep -E`, no
  `sed -i` without a suffix, no `\+` or `\|` in basic regexes, no `date -d`,
  no `\*` through `awk -v`. A scratch directory is
  `mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX"`.
- Sessions never move refs and write nothing under `.vloop/` but .vloop/tmp/
  (S-4): no task's work, and no gate, may require a session to change this
  repository's records, config or state. Tests build their repositories in
  temporary directories, with a temporary `HOME`, and address them with
  `git -C`.
- **The gate model applies** (`plugin/skills/plan/SKILL.md`): each gate judges
  this brief's contract by running the built binary or reading the changed
  document or code, with its judge in the verify or its gate folder; no gate
  runs the repository's tests — this repository's `[[check]]` does, after
  every iteration. V2's eval case is not run by any gate (it calls `claude`);
  its gate checks the case's files and its scaffold.
- **No task may weaken a check to pass. An earlier test changes only where it
  pins behaviour this contract changes, and only to the new behaviour**, the
  task's notes naming the item for each. This is K1's rule, binding this run
  although the released skills that run it predate K1. Known to change under
  it, as a help and not a licence: `internal/cli/config_test.go` and
  `internal/config/config_test.go` (V1's keys); `cmd/vloop/e2e_test.go`,
  `cmd/vloop/run_harness_test.go` and the end-to-end tests whose stub runs end
  complete (V1's session and line); `internal/cli/fence_test.go` (V1's fence);
  `internal/metrics/print_test.go`, `internal/metrics/time_test.go`,
  `internal/metrics/instructions_test.go` (V3); `internal/cli/close_test.go`
  and `internal/cli/close_flags_test.go` (V4); `internal/budget/budget_test.go`
  and `internal/cli/readme_test.go` (V2's skill); `internal/brief/pinning_test.go`
  and `internal/brief/extformat_test.go`, removed, and
  `internal/cli/init_test.go` and `cmd/vloop/b14_e2e_test.go` (K2);
  `docs/guide/commands.md`'s staleness test (regenerate, as always).
- Repo-relative paths everywhere. No absolute paths in any file or commit
  message.

## Shape

10 to 12 tasks, each independently verifiable by a single command. Order is
load-bearing:

1. **S1, `.vloop/tmp/` cleared** — first, so no later task meets stale scratch.
2. **K2, the two warnings retire** — small, and removes code later tasks would
   otherwise read.
3. **V1's surfaces: the config keys, the `verify` phase in the session schema,
   the fence and `vloop status`.**
4. **V1's driver: the verify session after the final pass**, the report kept,
   its four messages.
5. **V2, the `/vloop:verify` skill** and its word ceiling.
6. **V2's eval case.**
7. **V3, the metrics.**
8. **V4, close shows the findings first.**
9. **K1, the templates and the work, review and architect skills.**
10. **D1, the docs, the domain and the playbooks.**
11. **Close:** the worked example line for line, and the real-data check.
<!-- vloop:run-record:begin -->
## Run record

Generated by vloop brief close on 2026-10-15. Recompute with: vloop metrics B20261015-1001-b16-verification-phase.loop-brief

```
B20261015-1001-b16-verification-phase  consumed · not merged
 tasks     11 planned (brief said 10–12) · 11 done · 0 blocked · first-pass 5/11
 size      delivered  code 258 · test 855 · docs 192 · test:code 3.31
           churn      code 258 · test 856 · docs 195 · rework 1.00
 time      agent 66.3 min (work 61.3 · review 5.0) · plan 20.5 min · gate review 3.2 min · gates 2.5 · wall 167.7 min
 rate      3.9 code lines/min · 16.8 incl. tests
 cost      $14.35 · plan 7.43 · work 4.37 · review 1.66 · gate review 0.89 · $55.62/kloc
 models    plan claude-opus-5-5 · work claude-sonnet-5-5 · review claude-sonnet-5-5 · gate review claude-sonnet-5-5
 defects   in-loop 8 · operator 1 · escaped 0 · removal efficiency 100%
 interventions  5 · recommended 1 · other-option 0 · adjusted 1 · different 0 · no-options 3
 instructions  plan 4,569 · gate review 2,088 · work 2,436 · review 2,291 · verify 1,806 words · guidelines 2 rules
```

```
id   area  kind      att  code+  test+  docs+  other+  test:code  $/kloc  agent   cost   model
T1   -     feature   3    52     165    0      0       3.17       20.07   13m32s  $1.04  claude-sonnet-5-5
T2   -     refactor  1    1      23     0      0       23.00      386.80  5m21s   $0.39  claude-sonnet-5-5
T3   -     feature   2    38     49     3      0       1.29       15.14   8m55s   $0.58  claude-sonnet-5-5
T4   -     feature   2    77     215    0      137     2.79       13.71   20m30s  $1.06  claude-sonnet-5-5
T5   -     feature   2    2      3      112    0       1.50       220.59  3m07s   $0.44  claude-sonnet-5-5
T6   -     test      2    0      125    0      0       -          -       1m21s   $0.42  claude-sonnet-5-5
T7   -     feature   1    46     32     0      0       0.70       9.39    5m31s   $0.43  claude-sonnet-5-5
T8   -     feature   1    38     40     0      0       1.05       8.79    2m19s   $0.33  claude-sonnet-5-5
T9   -     docs      1    4      0      10     0       0.00       56.88   1m07s   $0.23  claude-sonnet-5-5
T10  -     docs      2    0      0      70     0       -          -       2m28s   $0.68  claude-sonnet-5-5
T11  -     test      1    0      204    0      0       -          -       2m05s   $0.43  claude-sonnet-5-5
```

- B20261015-1001-b16-verification-phase/i1-gate — regression, work, found by gate: check failed
- B20261015-1001-b16-verification-phase/i2-review-1 — bug, work, found by review: No committed test covers the clearing points or the two kept hand-offs at run level; only ClearTmp itself is unit-tested, so clearing at a forbidden point would pass the checks.
- B20261015-1001-b16-verification-phase/i2-review-2 — bug, work, found by review: The task's notes lack the evidence (code paths read) that nothing relies on .vloop/tmp/ surviving a resume; they refer to a journal of an earlier attempt.
- B20261015-1001-b16-verification-phase/i5-gate — regression, work, found by gate: check failed
- B20261015-1001-b16-verification-phase/i7-review-1 — bug, work, found by review: the committed tests omit most of the messages and cases criterion 6 requires: no findings, invalid report, failed session, cost ceiling, revert line, exit 9, resume, status phase
- B20261015-1001-b16-verification-phase/i9-gate — regression, work, found by gate: check failed
- B20261015-1001-b16-verification-phase/i11-gate — regression, work, found by gate: check failed
- B20261015-1001-b16-verification-phase/i16-review-1 — bug, work, found by review: operate playbooks quote `vloop brief close --dry-run` without the required <brief> argument, which fails; must be `vloop brief close <brief> --dry-run`
- I20261015-151154-v4-left-the-order-of-close-s-verify-repo — verification-finding, brief, found by operator: V4 left the order of close's verify report and its findings refusal open, so the operate step's 'close --dry-run prints them first' failed; a bare close now shows the report
<!-- vloop:run-record:end -->

### Operator notes

- **One halt before planning, repaired.** The base check refused: VS Code's GitHub Pull Requests extension wrote `branch.<run id>.github-pr-base-branch` into `.git/config` during it. Ignored through `run.git-config-ignore` to resume (744f173), then made a built-in pattern (a33865f) and the repository's line removed. Nothing was spent on the refused start.
- **Verified by hand.** The full suite and checks; a base build against this branch: `vloop metrics` identical for every consumed brief, `vloop brief check` identical for every brief; `config list`, `schema validate verify/v1` (valid, and a bad kind refused), `init`'s guidelines file, and close's verify block in a throwaway clone. No verify session ran in the run (0.10.0 has none); the skill's eval case, run once with sonnet on this branch's binary, scored 1.00 ($0.10): all three planted mismatches found, the sound line not flagged.
- **Fixed in this brief, by the operator's direction** (a33865f): a bare `vloop brief close` prints the verify report and refuses, writing nothing, while `--dry-run` without a flag still refuses (V4 left the order open; the finding above); the guard's GitHub PR keys (`D20261015-151118`, an escape of B12); `vloop config set` and `AddChecks` edit the file's text, keeping comments (`D20261015-151119`, an escape of B1); concepts.md's stale "as defects" (`D20261015-151120`, an escape of B15). Earlier tests changed for the close fix: `TestCloseRefusals`, `TestCloseVerifyBlock`, `TestWorkedExampleB4Close`.
- **Changed earlier tests** in the run are each licensed by an item: four tests that kept markers in `.vloop/tmp/` keep them under `.git/` (S1); enumerations of keys, schemas and skills (V1, V2); the stub harness answers `/vloop:verify` (V1); the retired warnings' tests (K2).
- **Next.** The first real verify session runs on the brief after a release that carries B16.

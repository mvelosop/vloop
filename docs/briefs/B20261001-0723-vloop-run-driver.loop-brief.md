---
name: B20261001-0723-vloop-run-driver.loop-brief
description: Port the shell loop's driver to vloop run — plan, iterate, gate, review, commit, halt — in vloop's layout and formats, proven by the shell loop's scenarios ported to Go with a stub claude, plus what the shell driver never did: work branches, configured model and effort per session, gate durations, flaky gates, gate disputes, a snapshot each iteration and the summary at the end
kind: brief
status: consumed
created: 2026-10-01
seeds: A `.loop/run.sh` plan + run that builds slice B6 in docs/design-notes/vloop-roadmap.md
depends-on: [B20260930-2007-vloop-init-upgrade-doctor.loop-brief]
---
# Brief — vloop B6: `vloop run`, the driver

- **Status:** consumed — closed 2026-10-01 as run B20261001-0723-vloop-run-driver. **Do not re-plan from this brief.**
- **Starting point:** extends `main` at the commit that adds this brief. The
  planner pins that SHA as the base every gate compares against.
- **Produced by:** the design act on the B6 row of
  `docs/design-notes/vloop-roadmap.md`, 2026-10-01, with the operator.

---

## What it is

The sixth slice of `vloop`: the driver. `vloop run` does what `.loop/run.sh`
does — plans a brief, then iterates: one task, a fresh work session, every done
task's gate, an independent review, one commit — and halts the same way, but
writes vloop's layout and formats (`.vloop/state/`, `state/v1`, `session/v1`,
`iteration/v1`, `[vloop]` commits), so everything B2–B5 built reads it. It also
does what the shell driver never did: starts on a work branch, runs each session
with its configured model and effort, times the gates, recognises a flaky gate,
blocks a task whose gate is disputed, freezes the metrics after every iteration
and prints the summary at the end. It is proven with a stub `claude`; the skills
it invokes are the next brief's.

## Why this shape, and what was rejected

Decided with the operator on 2026-10-01. **Do not re-litigate these.**

- **The driver only.** The skills `/vloop:plan`, `/vloop:work`, `/vloop:review`
  and `/vloop:operate` are B7; B7's run is the first real `vloop run`, and the
  v1.0 review (B8) runs on vloop itself.
  *Rejected:* driver and skills together (about twenty tasks, the skills written
  against a driver that does not yet exist); skills first (the shell loop would
  need changes to invoke plugin skills).
- **Parity is proven by porting the shell scenarios to Go.** Every scenario that
  tests driver behaviour becomes an end-to-end test with a stub `claude`,
  asserting the same outcome in vloop's layout.
  *Rejected:* reusing the bash scenarios through a shim — they assert `.loop/`
  paths and the shell formats.
- **A disputed gate blocks for the operator.** The driver does not replace a
  gate itself.
  *Rejected:* the driver verifying a proposed replacement and the review judging
  its intent — the operator keeps that judgement.
- **This repository keeps the shell loop until B7 closes.** B6 is planned and run
  by `.loop/run.sh`; nothing here uses `vloop run` on this repository.
- **Budgets are config keys with flags**, like every other setting.
- **The README stays under its cap**: the driver's exit codes live in the guide,
  and the README points to them.

## Binding references

- `docs/domain/execution/run.md` — the run folder, the iteration record and its
  outcomes, the commits and their subjects, the work branch, budgets and the exit
  codes; the layout this brief writes.
- `docs/domain/execution/session.md` — the three session kinds, what each reads
  and writes, the fence, and the `session/v1` record.
- `docs/domain/execution/task.md` — the task lifecycle, gates and gate history,
  model and effort resolution.
- `docs/domain/execution/plan.md` — the plan's lifecycle and how a run's ending
  maps onto `state/v1`'s statuses.
- `docs/domain/domain-model.md` — P-1 to P-6, S-1 to S-3, R-1 to R-3 and M-1: the
  invariants the driver enforces.
- `docs/briefs/B20260929-2325-vloop-metrics-defects.loop-brief.md` — "Where the
  numbers come from": the vloop layout and commit subjects `vloop metrics`
  already reads, which this driver must write exactly.
- `.loop/run.sh` — the behaviour being ported, mechanism by mechanism.
- `.loop/tests/lib.sh` — the scenario harness the Go tests replace: the stub
  `claude`, the fixture repository, the assertions each scenario builds on. The
  scenarios themselves are named one by one in the behaviour contract.
- `.loop/settings.json` — the fence the embedded one starts from.

## Behaviour contract

B1–B5's conventions hold: global flags, `--json`, exit codes, errors as one
stderr line starting `vloop: `, paths repo-relative with `/`. `vloop run` has
its own exit codes, R-3's.

### `vloop run [<brief>] [--plan-only] [--replan] [--max-iterations N] [--cost-ceiling USD] [--max-attempts N] [--stall-limit N]`

**Parity with the shell driver.** For each scenario below, the behaviour is
exactly what `.loop/tests/scenarios/<name>.sh` asserts of `.loop/run.sh`,
translated into vloop's layout (table below) — and nothing beyond it. Each
becomes a Go end-to-end test, `TestRun<NN><Name>`, that builds the binary, puts
a stub `claude` on `PATH`, and runs `vloop run` in a scratch repository.

| Ported | Scenarios |
| --- | --- |
| planning and plan checks | 08 plan-validation, 21 foreign-state, 25 gate-shape, 27 dangling-reference, 29 brief-typo, 30 plan-only, 35 gate-unowned-file, 36 gate-decaying-baseline, 38 head-diff-advisory, 43 brief-already-run |
| iterating | 01 happy-path, 02 review-fail, 03 gate-regression, 09 dependency-order, 16 review-fails-closed, 17 stale-handoff, 33 gate-rewrite, 37 gate-task-env, 39 blocked-gate-passes, 40 no-proposal-tree, 41 regression-names-both |
| halting and budgets | 04 attempt-ceiling, 05 max-iterations-resumable, 06 cost-ceiling-resumable, 07 convergence-halt, 11 stall, 19 session-error, 42 repeat-blocked-halts |
| safety | 10 containment, 18 preflight-untrusted, 22 run-lock, 23 git-identity, 24 state-tampering, 44 refs-moved-halts |
| records | 13 empty-run-signals, 14 rendered-views, 15 telemetry-contract (as: every record validates against `session/v1` / `iteration/v1`), 20 parallel-safe-layout |

**Not ported:** 26 install-boundary (the shell installer), 28 knowledge-roots
(vloop has none), 31 preflight-only (`vloop doctor`), 32 doc-transients and 34
todo-index-drift (the shell loop's own repository tooling), 12
signals-fixture (the formulas are B3's metrics).

**Translation.**

| Shell loop | vloop |
| --- | --- |
| `.loop/state/state.json` | `.vloop/state/state.json` (`state/v1`, validated) |
| `.loop/state/plan.md` | `.vloop/state/plan.md`, as `vloop status --markdown` renders it |
| `.loop/state/journals/<run id>.md` | `.vloop/state/journals/<run id>.md` |
| `.loop/state/runs/<branch>/<folder>/` | `.vloop/state/runs/<run id>/<YYYYMMDD-HHMMSS>/` |
| `sessions/NNN-<phase>.json` (raw `claude` JSON) | the same names, as `session/v1` (below) |
| `iterations.jsonl` | `iteration/v1` lines |
| `reports/NNN-verdict.json`, `gates/T<n>.log`, `loop.log` | `reports/NNN-verdict.json`, `gates/T<n>.log`, `run.log` |
| `.loop/tmp/proposal.json`, verdict.json | `.vloop/tmp/proposal.json` (`proposal/v1`), .vloop/tmp/verdict.json (`verdict/v1`); a file that fails its schema counts as absent |
| `[loop] plan …`, `[loop] T<n>: …`, `[loop] run …` | `[vloop] plan <run id>`, `[vloop] <task>: <outcome>`, `[vloop] run <run id>/<folder>: <status>` |
| outcomes `gate_fail`, `review_fail` | `gate_failed`, `rejected` |
| `LOOP_ACTIVE_TASK`, `LOOP_GATE_TASK` | `VLOOP_ACTIVE_TASK`, `VLOOP_GATE_TASK` |
| `--check` | `vloop doctor`; `run`'s preflight is doctor's checks, refusing on any problem (exit 1) |
| `LOOP_MAX_ITERATIONS` … | config keys and flags (below) |

### Sessions

Each session is
`claude -p "<prompt>" --model <m> [--effort <e>] --permission-mode auto
--setting-sources project --settings <fence> --plugin-dir <plugin>
--strict-mcp-config --output-format json`, from the repo root, where:

- the prompts are `/vloop:plan <brief path>`, `/vloop:work <task id>` and
  `/vloop:review <task id>`;
- `<m>` and `<e>` are the session kind's resolved model and effort (P-6);
  `--effort` is omitted when effort is unset. The review uses `model.review`, not
  the work model;
- `<plugin>` is what `vloop plugin path` extracts, and `<fence>` the embedded
  fence extracted beside it to `.vloop/tmp/fence/<version>/settings.json`.

**The fence** is `.loop/settings.json` as it stands, plus: allowed
`Bash(vloop status:*)`, `Bash(vloop task list:*)`, `Bash(vloop task show:*)`,
`Bash(vloop schema validate:*)`; denied every vloop command that writes —
`run`, `init`, `upgrade`, `brief close`, `task verify|reset|note|drop|set`,
`defect add|set`, `config set` — and the shell loop's driver, `.loop/run.sh`.

**The session record.** The driver writes `session/v1` from `claude`'s JSON:
`total_cost_usd` → `cost_usd`, `duration_ms`, `num_turns` → `turns`,
`is_error`, `permission_denials`, and `modelUsage` per model → `models_used`
(`inputTokens` → `input_tokens`, `outputTokens` → `output_tokens`,
`cacheReadInputTokens` → `cache_read_input_tokens`, `cacheCreationInputTokens` →
`cache_creation_input_tokens`, `costUSD` → `cost_usd`); plus `run_id`,
`iteration`, `phase`, `task`, the configured `model` and `effort`, and `started`
(the driver's clock when it launched the session). Every persisted record and log
masks `$HOME` and the user name, as the shell driver does.

**A session that prints nothing** gets no record — the driver never writes a
fabricated one — and the run log says `SESSION RECORD MISSING <phase>
<task|-> (iteration <n>)`; the metrics report it as missing.

### What the shell driver never did

- **Work branch (R-2).** On the default branch, `run` creates `<run id>` and
  switches to it before planning, printing `created and switched to branch <run
  id>`. If that branch already exists: `vloop: branch <run id> exists — switch to
  it and re-run`, exit 1. On any other branch, it runs there.
- **HEAD before every commit.** Before each commit the driver checks HEAD is
  still on the run's branch; if not, it halts with exit 9 and commits nothing —
  as scenario 44 does for moved refs.
- **Gate timing.** Each gate run's duration goes into the iteration record's
  `gate.duration_ms` (the iteration's gate is the current task's).
- **Flaky gates.** A gate that fails is run once more at once, with nothing
  changed. If it passes, the iteration continues as a pass, and its record carries
  `"gate": {"exit": 0, "duration_ms": …, "flaky": true}`; the run log says
  `FLAKY GATE <task>`. `vloop metrics` derives one defect from such an iteration:
  `origin: env`, `kind: bug`, `found-by: gate`.
- **Gate disputes.** A proposal with `gate_dispute` blocks its task at once:
  status `blocked`, `notes` = `gate disputed: <reason> — <evidence>`, no attempt
  charged, no review; the run continues with other ready tasks and ends `blocked`
  (exit 2) if none remain. The operator resolves it with `vloop task verify`.
- **Snapshot.** After every iteration's commit, `.vloop/state/metrics/<run
  id>.json` is rewritten with the brief's current metrics (`metrics/v1`) and
  included in the next commit.
- **Summary.** At the end of a run, whatever its ending, the driver prints the
  brief's `vloop metrics` summary.
- **Plan status.** The plan's `status` is one of `state/v1`'s: `planning`,
  `running`, `complete` (exit 0), `blocked` (2), `stalled` (3), or `halted`
  (4–9).

### Budgets

Config keys, after `metrics.excluded`, each with its flag and `VLOOP_RUN_*`
variable:

| Key | Default | Flag |
| --- | --- | --- |
| `run.max-iterations` | `30` | `--max-iterations` |
| `run.cost-ceiling` | `40` (dollars) | `--cost-ceiling` |
| `run.max-attempts` | `3` | `--max-attempts` |
| `run.stall-limit` | `2` | `--stall-limit` |
| `run.convergence-max` | `3.0` | — |
| `run.convergence-min` | `6` | — |

A flag beats the environment, which beats the file, which beats the default.
Budgets are per run, checked between iterations; raising one and re-running
resumes (scenarios 05, 06).

### F1 — the README states the wrong shell default on Windows

**Defect.** `README.md`'s config table says the `shell` default is `cmd` on
Windows; the code and B2's brief say `pwsh`.

**Required.** The README says `pwsh`. **Gate.** A test that the README's stated
default for `shell` on each OS equals the code's.

### Decided here, because the cited documents leave it underdetermined

- The fence's additions; the session-record mapping; `started` as the driver's
  clock.
- A flaky gate is one immediate re-run, not a later attempt.
- A disputed task charges no attempt: the dispute is about the gate, not the
  work.
- The exit codes 0–9 move from the README to the guide; the README keeps one
  pointer. **B1's README test must change** to check the driver's exit codes
  against the guide page instead of the README — a required change, not a
  weakening.

### Violations the review must rule on

- A ported test that asserts less than its shell scenario does, or a scenario in
  the "ported" table with no test.
- The driver writing anywhere but `.vloop/` and the working tree the sessions
  changed, or committing anything a session did not produce plus the driver's own
  records.
- Any session started without the fence or the plugin directory, or with a model
  or effort other than the resolved one.
- A session record fabricated for a session that printed nothing.
- The driver replacing, editing or disabling a gate itself.
- `vloop run` used on this repository, or any skill content under
  `plugin/skills/` (B7).

## Worked example

A scratch git repository on `main`, initialised with `vloop init`; a ready brief
docs/briefs/B20260101-0900-a.loop-brief.md; a stub `claude` on `PATH` that
logs its arguments, then plays by prompt: the plan session writes a `state/v1`
plan with T1 (gate `test -f a.txt`) and T2 depending on T1 (gate
`test -f b.txt`); each work session creates its file and writes a `done`
proposal; each review writes `PASS`. Every session reports `total_cost_usd`
0.10, `duration_ms` 1000 and `modelUsage` for the model it was given. Config:
`model.review = "opus"`, `effort.work = "high"`. `<folder>` and timings are
threaded from the output.

```
$ vloop run docs/briefs/B20260101-0900-a.loop-brief.md
created and switched to branch B20260101-0900-a
  …
B20260101-0900-a  ready · not merged
 tasks     2 planned · 2 done · 0 blocked · first-pass 2/2
  …                                                          exit 0
$ git log --format=%s main..HEAD
[vloop] run B20260101-0900-a/<folder>: complete
[vloop] T2: done
[vloop] T1: done
[vloop] plan B20260101-0900-a
$ vloop status | head -1
B20260101-0900-a — complete · 2/2 done · 0 blocked · iteration 2
$ the stub's argument log
  /vloop:plan docs/briefs/B20260101-0900-a.loop-brief.md   --model opus
  /vloop:work T1  --model sonnet --effort high  --settings .vloop/tmp/fence/<v>/settings.json --plugin-dir .vloop/tmp/plugin/<v>
  /vloop:review T1  --model opus
  …
$ vloop schema validate session/v1 .vloop/state/runs/B20260101-0900-a/<folder>/sessions/002-work.json
… : ok                                                       exit 0
$ jq -c .gate .vloop/state/runs/B20260101-0900-a/<folder>/iterations.jsonl | head -1
{"exit":0,"duration_ms":<n>}
$ vloop metrics B20260101-0900-a.loop-brief | grep time
 time      agent … · gates <m> min · wall …          (gates no longer n/a)
```

The failures, each planted in a copy of the fixture:

```
the branch B20260101-0900-a already exists, run from main
  -> vloop: branch B20260101-0900-a exists — switch to it and re-run      exit 1
T2's work session writes a proposal with gate_dispute
  -> T2 blocked, notes "gate disputed: …", attempts 0, no review          exit 2
T1's gate fails once then passes on the immediate re-run
  -> FLAKY GATE T1; the iteration record has "flaky": true; T1 done;
     vloop metrics counts one env defect                                   exit 0
the review session prints nothing on stdout but writes its verdict
  -> SESSION RECORD MISSING review T1 (iteration 1); no 004-review.json;
     vloop metrics prints the records line                                 exit 0
a work session checks out another branch
  -> REFS MOVED T1 …; nothing committed                                     exit 9
--max-iterations 1
  -> halts after T1, plan status halted, resumable: re-run finishes        exit 4
```

**Real data.** On a throwaway clone of this repository, read-only for the
repository itself: `vloop run` with the stub `claude` and a one-task fixture
brief completes, and `vloop metrics` reads the run it wrote. Nothing in this
repository's `.loop/` or `.vloop/state/` changes.

## Out of scope

The roadmap's later rows:

- Skill content under `plugin/skills/`, the planner running its own gates, skill
  evals (B7).
- Using `vloop run` on this repository, `vloop init` here, retiring `.loop/` as
  the driver (B7's close).
- The v1.0 review (B8).

Also not in this run:

- The driver verifying or applying a replacement gate.
- Archiving session transcripts (`LOOP_ARCHIVE_TRANSCRIPTS`).
- Running several worktrees for the operator; one run per worktree, as now.
- Converting shell-loop run folders into vloop's layout.
- Network access, publishing, package managers.

## Constraints

- Go as pinned in `go.mod`. Third-party modules: B5's; anything else needs its
  reason in the task notes. No cgo. vloop itself never uses the network.
- Must build for `darwin`, `linux` and `windows`; nothing may assume `/` on disk
  or a case-sensitive filesystem; gates run in the plan's shell.
- **Gates run on macOS with its BSD tools**: no GNU-only syntax (no empty
  alternatives in `grep -E`, no `sed -i` without a suffix, no `\+` or `\|` in
  basic regexes, no `date -d`).
- **Every gate is executed by the planner before it hands over the plan.** Three
  gate defects in this series (B2 T1, B5 T6, B5 T8) were gates whose fixture
  edits nobody ran: a TOML key appended after a table, a `git status` that a
  required edit makes non-empty, a GNU-only pattern. A gate that cannot pass for
  a correct implementation is a planning defect.
- **Sessions must not move git refs, and cannot run `claude`**: the fence denies
  both. Tests put a stub `claude` on `PATH`; tests build fixture repositories and
  fake homes in temporary directories and address them with `git -C`; a session
  checks a gate by running the gate, never by pasting its commands.
- **Gates follow the diff surface**: each task gates on the packages it touches,
  plus `go vet ./...`, `gofmt -l .` and `go mod tidy -diff` printing nothing,
  and the linux, windows and host builds. Only the close runs `go test ./...`.
- A gate runs in under three minutes. Tests never read or write this
  repository's `.loop/`, `.claude/`, `.vloop/`, git history or the real home
  directory; only the close task's real-data check reads this repository, in a
  throwaway clone.
- **No task may weaken a check to pass.** Required changes, and only these:
  `config list` and B1's README test gain the six `run.*` keys; B1's README test
  checks the driver's exit codes against the guide page; the generated command
  reference is regenerated.
- **New checks are proven by fixtures**: every ported scenario, and every
  behaviour under "What the shell driver never did", has a fixture where it
  applies and one where it must not.
- Repo-relative paths everywhere. No absolute paths in any file or commit message.

## Shape

11 to 14 tasks, each independently verifiable by a single command. Order is
load-bearing:

1. **The stub `claude` and the test harness**: a scratch repository, a fake home,
   a scripted stub by prompt, the binary built once; a smoke test.
2. **Budgets** as config keys and flags, with the `config list` update.
3. **Sessions**: invocation, the fence and plugin extraction, the `session/v1`
   record and masking, the missing-record warning.
4. **Planning**: the work branch, planning and the plan checks — scenarios 08,
   21, 25, 27, 29, 30, 35, 36, 38, 43.
5. **One iteration**: task choice, work, gates with timing and the env ids,
   review, applying the outcome, journal, `plan.md`, the commit with the HEAD
   check — 01, 02, 09, 16, 17, 37, 40.
6. **Gates across iterations**: regressions, rewrites, blocked tasks whose gate
   passes, flaky gates, disputes — 03, 33, 39, 41.
7. **Halting**: budgets, convergence, stall, session errors, repeat-blocked,
   refs — 04, 05, 06, 07, 11, 19, 42, 44.
8. **Safety and records**: containment, preflight, the lock, git identity,
   tampering, the empty run, rendered views, telemetry, layout — 10, 13, 14,
   15, 18, 20, 22, 23, 24.
9. **Snapshot and summary**, and the env defect `vloop metrics` derives from a
   flaky gate.
10. **Docs**: the exit codes 0–9 in the guide with the README's pointer and its
    test change, the `run` row, the `run.*` keys, F1, the regenerated command
    reference; `docs/domain/execution/run.md` and `plan.md` stating the driver
    as built.
11. **Close**: the worked example line for line on the stub; the real-data check
    in a throwaway clone; `go test ./...`, `go vet ./...`, `gofmt -l .`,
    `go mod tidy -diff` and the three builds.

The worked example needs its own gate (task 11): it is the only check that the
pieces — sessions, gates, records, commits, snapshot, summary — fit together in
one run.
<!-- vloop:run-record:begin -->
## Run record

Generated by vloop brief close on 2026-10-01. Recompute with: vloop metrics B20261001-0723-vloop-run-driver.loop-brief

```
B20261001-0723-vloop-run-driver  consumed · not merged
 tasks     11 planned (brief said 11–14) · 11 done · 0 blocked · first-pass 11/11
 size      delivered  code 2,546 · test 3,298 · docs 96 · test:code 1.30
           churn      code 2,618 · test 3,302 · docs 99 · rework 1.03
 time      agent 69.3 min (work 60.6 · review 8.7) · plan 27.3 min · gates n/a · wall 96.3 min
 rate      36.7 code lines/min · 84.3 incl. tests
 cost      $18.02 · plan 6.93 · work 9.31 · review 1.78 · $7.08 per 1,000 code lines
 models    plan claude-opus-5-5 · work claude-sonnet-5-5 · review claude-sonnet-5-5
 defects   in-loop 0 · operator 1 · escaped 0 · removal efficiency 100%
```

```
id   area  kind  att  code+  test+  docs+  other+  agent   cost   model
T1   -     -     1    0      274    0      0       1m27s   $0.42  claude-sonnet-5-5
T2   -     -     1    74     106    20     0       2m57s   $0.59  claude-sonnet-5-5
T3   -     -     1    490    241    0      77      3m20s   $0.68  claude-sonnet-5-5
T4   -     -     1    815    717    18     0       10m03s  $2.36  claude-sonnet-5-5
T5   -     -     1    662    568    0      0       7m49s   $1.55  claude-sonnet-5-5
T6   -     -     1    138    287    0      0       4m42s   $0.91  claude-sonnet-5-5
T7   -     -     1    161    294    0      0       15m07s  $1.47  claude-sonnet-5-5
T8   -     -     1    164    392    0      0       7m45s   $1.14  claude-sonnet-5-5
T9   -     -     1    114    94     0      0       6m01s   $0.81  claude-sonnet-5-5
T10  -     -     1    0      95     61     242     2m41s   $0.55  claude-sonnet-5-5
T11  -     -     1    0      234    0      0       7m28s   $0.59  claude-sonnet-5-5
```

- D20261001-1013-domain-plan-md-said-the-driver-stamps-ru — bug, work, found by operator: domain plan.md said the driver stamps run_id as the brief's name
<!-- vloop:run-record:end -->

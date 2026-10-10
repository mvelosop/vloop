---
name: B20261008-2144-b12-field-reliability.loop-brief
description: Make the driver's instruments trustworthy in the field — a .git/config guard that ignores editor bookkeeping, plan acceptance that survives a halt, session time measured by the driver, denied Bash commands recorded, the run's cost and progress visible while it runs, an installed binary that names its version, and metrics that say when nothing was measured
kind: brief
status: draft
created: 2026-10-08
seeds: The first brief after the T1 triage (its P1); docs and onboarding follow as B13
depends-on: [B20261005-0933-quality-pass.loop-brief]
---
# Brief — vloop B12: field reliability

- **Starting point:** extends `main` after the T1 triage merges (the design
  branch `design/t1-triage` on top of `7bd76f6`). The planner pins the base it
  plans from as the **base** every gate compares against — never `HEAD~n`.
- **Produced by:** the T1 triage, `docs/design-notes/vloop-t1-triage.md` → P1,
  and the design act of 2026-10-08, with the forks settled with the operator
  (intervention recorded with this brief).
- **Run by:** `vloop run` from the released v0.8.0.

## What it is

vloop's first use outside this repository did the work — 9 tasks, 9 first-pass —
and got every instrument around it wrong: VS Code's write to `.git/config` halted
a plan with exit 9, the halted plan's $5.51 vanished from the metrics, a
39-minute plan session was recorded as 20 seconds, five denied Bash calls could
not be explained, and the operator saw no cost until the run ended. After this
run, the `.git/config` guard halts on the keys that matter and names them; a halt
during plan acceptance resumes instead of re-planning; the driver times every
session itself; a denied Bash call keeps its command; the run's cost, elapsed
time and session in flight are visible while it runs; an installed binary names
its version; and metrics say "n/a" where nothing was measured.

The blind spot: every one of these was tested only in this repository, by an
operator who knew to close the editor, read the log afterwards and trust the
numbers. The tests built fixtures that never had an editor, a long plan session
or a halt during acceptance.

## Why this shape, and what was rejected

Decided with the operator on 2026-10-08. **Do not re-litigate these** — the
planner inherits them.

- **One brief for everything the field found** (T1 decision 5). *Rejected:* a
  patch release for the two exit-9 defects alone, then the rest.
- **The guard compares parsed keys, with an allowlist.** A built-in list of
  editor keys is ignored, a repository may add patterns with a config key, and
  every other key added, removed or changed still halts. *Rejected:* a list in
  the binary only (a new editor needs a release); guarding only known-dangerous
  keys (an unknown dangerous key would pass).
- **A halt during acceptance resumes acceptance with the same plan.** No second
  plan session. *Rejected:* re-planning but keeping the halted run's cost — it
  still pays for the plan twice.
- **A denied Bash call keeps its command, masked, redacted and cut to 200
  characters.** This reverses B8 F12 for Bash's command only; file contents and
  other tools' inputs stay out. *Rejected:* the program name only (the
  visum-monorepo denial "rm on an unguarded variable" would still be
  unexplainable); the full command (a long heredoc, or a secret typed inline,
  would be committed whole).
- **Visibility: a line per session, ceiling warnings, and `vloop status` from
  the lock.** *Rejected:* writing `run.log` live (its buffering keeps the log
  and the commit consistent; stdout is already live and is what the operator
  redirects); a heartbeat inside a session (needs streaming sessions — a larger
  driver change, carried).

## Binding references

- `docs/design-notes/vloop-t1-triage.md` — P1: the proposal this brief implements, and the evidence for each part
- `docs/domain/domain-model.md` — S-2, S-4, P-2, R-3, R-4, M-1, M-2, C-1 and C-3: the guard, what sessions write, exit codes, the clean-tree preflight, how metrics are computed and how records may grow
- `docs/domain/execution/plan.md` — the plan's lifecycle and acceptance, which the resume changes
- `docs/domain/execution/run.md` — the run folder, its ownership and the exit table, which the resume changes
- `docs/domain/execution/session.md` — the session record's fields, which grow
- `docs/domain/measurement/metrics.md` — the time and size measures, whose wall time and n/a rules change
- `docs/guide/configuration.md` — every config key and its environment name; the new key joins it

## Behaviour contract

Messages are exact where quoted; `<…>` marks a value. A refusal or halt is one
line on stderr starting with `vloop: `.

### R1 — the `.git/config` guard names keys and ignores editor bookkeeping

**Defect.** The guard hashes `.git/config`'s bytes, so VS Code's
`branch.<name>.vscode-merge-base` — written when the work branch is created —
halts the run with exit 9 and blames whatever session or gate was running
(`.vloop/defects/D20261005-1125-the-driver-s-git-config-guard-cannot-tel.md`;
again in another repository on 2026-10-07).

**Required.**
- Wherever the guard runs today (around every session, gate and check), it
  compares the file's **keys and values** before and after — the file's own
  entries, include directives as entries, not the files they include. The git
  hooks guard is unchanged.
- An entry whose key matches an **ignore pattern** is left out of both sides.
  The built-in pattern is `branch.*.vscode-merge-base`. The config key
  **`run.git-config-ignore`** (a list of patterns, environment
  `VLOOP_RUN_GIT_CONFIG_IGNORE`, comma-separated) adds patterns to it; it never
  removes the built-in one. A pattern is matched against the key as
  `git config --list` prints it (section and variable lowercased, the
  subsection as written); `*` matches any run of characters, dots included.
- `config set run.git-config-ignore` refuses, exit 2, a **protected** pattern:
  one whose text before its first `*` (all of it, if it has none) and one of
  `core.`, `filter.`, `include.`, `includeif.` is a prefix of the other —
  so `*`, `c*`, `core.*` and `filter.lfs.*` are refused, `gitlens.*` and
  `branch.*.vscode-merge-base` are not:
  `vloop: run.git-config-ignore: <pattern> would ignore <section>.* — the guard must see it`,
  `<section>` the first protected section, in that order, it collides with. The
  environment variable and the config file are held to the same rule when the
  driver reads them: a protected pattern refuses the run, exit 1, with the same
  text.
- Any other entry added, removed or changed halts, exit 9, naming at most three
  keys, sorted, and how many more:
  `vloop: <who> changed .git/config (<key>[, <key>…][ and <n> more]) — nothing was committed; restore it, then re-run`.
  `<who>` is what it is today (`work`, `the gate of T5`, `the check go`, …).

### R2 — a halt during plan acceptance resumes acceptance

**Defect.** Acceptance — the base gates and the gate review — runs on an
unstamped plan. A halt there (exit 9 in B11 and in another repository) leaves
a plan the re-run cannot resume: it plans again from scratch, and the operator's
clean-up loses the halted run folder with the plan session's record and cost
(`.vloop/defects/D20261005-1125-an-exit-9-during-plan-acceptance-discard.md`;
$8.23 and $5.51 lost). A halt during the gate review leaves a stamped plan whose
re-run starts work with no gate review.

**Required.**
- As soon as the plan session's plan loads and passes its checks, the driver
  stamps `run_id`, `brief` and `branch`; the plan's status stays `planning`
  until acceptance passes, and becomes `running` then.
- A halt or a killed driver during acceptance leaves the plan, its gate folders
  and the run folder in place. `vloop run <brief>` (or bare `vloop run`) on the
  plan's branch **resumes acceptance**: no plan session; the round it was in
  continues — the base gates, then the gate review — and the rounds already
  spent count toward the two. It prints
  `resuming acceptance of <run id>: round <n>`.
- The clean-tree preflight tolerates the plan, its gate folders and the brief's
  earlier run folders when it resumes acceptance, as it does when it resumes
  work.
- The halted run folder is kept: the plan commit includes it, and the metrics
  count its sessions (`docs/domain/execution/run.md`'s ownership rule).
- Work never starts on a plan whose gate review has not passed: a plan with
  status `planning`, or any plan whose gate review verdict is not `PASS`,
  resumes acceptance instead. A plan the gate review failed twice (status
  `blocked`) behaves as today.
- A plan rejected by its checks (exit 1) is still removed, as today.

### R3 — the driver times every session

**Defect.** A session's `duration_ms` is copied from the claude CLI's JSON,
which recorded 20 seconds for a plan session that ran 39 minutes (another
repository, 2026-10-07); `vloop metrics` showed the plan at 0.3 min and a
71-minute run at 30.7.

**Required.**
- Every session record gains **`ended`** (RFC 3339, UTC, like `started`) and
  **`wall_ms`**, the driver's own measurement from the process start to its
  exit. Both are optional in the schemas (C-3): records written before stay
  valid and no schema version changes. `duration_ms` stays, as the CLI reported
  it.
- `vloop metrics` takes a session's time from `wall_ms` when the record has it,
  and from `duration_ms` otherwise.
- A brief's **wall** time runs from its first plan session's `started` to the
  last run commit (today: from the plan commit, which leaves planning out).
  Where no plan session record exists, from the plan commit, as today.

### R4 — a denied Bash call keeps its command

**Defect.** A permission denial keeps only `tool_name` and a `file_path`
(B8 F12), so five Bash denials in another repository could not be analysed.
Separately, masking misses Claude Code's dashed project-folder form,
`-Users-<name>-…`, and two committed records keep a user name.

**Required.**
- A denied `Bash` call's record keeps **`command`**: the command through the
  same masking and secret redaction as every record, then cut to 200
  characters, with `…` appended when cut. Other tools' records are unchanged.
- Masking replaces the user name in the dashed form: `-<Users|home>-<user>-`
  becomes `-<Users|home>-USER-`, everywhere masking applies (records, logs,
  stderr captures).

### R5 — the run's cost and progress while it runs

**Defect.** Nothing during a run says what it has spent or how long a session
has taken; `run.cost-ceiling` is a stop, not feedback, and `vloop status` reads
only the plan.

**Required.**
- After each session ends, the driver prints one line (stdout and `run.log`):
  `   <phase>[ <task>] · <m>m<ss>s · $<cost> · run $<spent> / $<ceiling>` —
  `<phase>` one of `plan`, `gate review`, `work`, `review`; `<spent>` the same
  figure the ceiling is checked against, after this session; money with two
  decimals. Example: `   work T3 · 6m12s · $1.84 · run $9.40 / $40.00`.
- The first time `<spent>` reaches 50%, and then 80%, of the ceiling, the
  driver warns once each (stderr and `run.log`):
  `   cost: $<spent> spent, <50|80>% of the $<ceiling> ceiling`.
- The run lock (`.vloop/tmp/.running`) also records the session in flight —
  `phase`, `task` (absent for plan and gate review), `session_started` — and
  `spent`, updated as each session starts and ends.
- `vloop status`, while a live run holds the lock, prints after its first line:
  `running: <phase>[ <task>] for <m>m<ss>s · $<spent> of $<ceiling>`;
  `--json` gains `"running": {"phase", "task", "session_started", "spent", "ceiling"}`,
  absent when no live run. With no plan yet but a live run planning, `status`
  prints the `running:` line and exits 0.

### R6 — an installed binary names its version

**Defect.** A binary from `go install …@v0.8.0` prints `commit unknown`; only
the stamped build carries a commit (finding 2).

**Required.**
- The commit is the stamped one; else the build's VCS revision when Go recorded
  one (a build inside a checkout), with `-dirty` appended when the tree was
  modified; else `unknown`.
- When the build carries a module version (anything but `(devel)` or empty),
  `vloop version` prints it in place of an unknown commit:
  `vloop <version> (module <module version>, plugin <plugin>, <go> <os>/<arch>)`;
  `--json` gains `"module"` after `"commit"`, absent when there is none.
- `doctor`'s self-hosting check passes a binary whose module version is a
  release (`vX.Y.Z`, no pre-release or pseudo-version suffix), as it passes a
  stamped one.

### R7 — metrics say when nothing was measured

**Defect.** A brief that delivered only lines no stack classifies shows `code 0`
and `test 0` in `vloop metrics`, which reads as "delivered nothing" (another
repository: 3,487 lines of scripts and docs).

**Required.** When a brief's delivered lines include none classified as code
or test, the cross-brief table prints `n/a` for its `code` and `test`; the
summary's delivered line is unchanged and is followed by
`           <n> delivered lines match no stack in metrics.stacks`
when `<n>`, the delivered `other` lines, is above zero. The JSON is unchanged.

### The domain, updated

- S-2: the `.git/config` part reads "a key of `.git/config` changes, other
  than those `run.git-config-ignore` and the built-in editor keys ignore".
- `docs/domain/execution/plan.md`: the plan is stamped on load; acceptance
  resumes after a halt; work needs a passed gate review.
- `docs/domain/execution/run.md`: the resume of acceptance in the exit table
  (exit 9 during acceptance: resumable once restored) and the kept run folder.
- `docs/domain/execution/session.md`: `ended`, `wall_ms`, and a denial's
  `command`.
- `docs/domain/measurement/metrics.md`: session time from `wall_ms`; wall from
  the first plan session; `n/a` for unmeasured size.
- The guides follow: `docs/guide/configuration.md` (the key),
  `docs/guide/metrics.md` (wall, n/a), `docs/guide/concepts.md` (exit 9 and
  resuming acceptance, the per-session line, `status`), and
  `docs/guide/commands.md` regenerated.

### Decided here, because the cited documents leave it underdetermined

- The built-in ignore list is one pattern; a repository adds its own editors'
  keys. Patterns are refused only where they would blind the guard to keys that
  run code (`core.`, `filter.`, `include.`, `includeif.`).
- "Resume acceptance" re-runs the base gates of the current round even if some
  passed before the halt: a halted gate's evidence is not trusted.
- `wall_ms` is the process's wall time, including any time the machine slept;
  `run.keep-awake` exists for that.
- The per-session line is printed for every session kind, including a
  revision round's plan session.
- 200 characters is counted in characters, not bytes, after masking and
  redaction.

### Violations the review must rule on

- A key outside the ignore patterns that changes without halting, or an ignored
  key that halts.
- A path where acceptance resumes by planning again, or where work starts
  without a passed gate review.
- A record that keeps any tool input but a Bash `command` and a `file_path`.
- A schema version bump for the session record.
- Any refactor from B15's list (one git wrapper, the env filters, `planSHA`,
  `driver.counts`, the budget keys, the id generators, long functions) folded in.

## Worked example

With a fresh build, in temporary repositories, with the stub `claude` the
suite already uses:

```
a work session appends [branch "B1"] vscode-merge-base = refs/heads/main
                                      -> the run continues; exit 0 at the end
a work session appends [core] fsmonitor = true
  -> vloop: work changed .git/config (core.fsmonitor) — nothing was committed; restore it, then re-run   exit 9
run.git-config-ignore = ["gitlens.*"]; a session sets gitlens.x = 1
                                      -> the run continues
vloop config set run.git-config-ignore 'core.*'
  -> vloop: run.git-config-ignore: core.* would ignore core.* — the guard must see it                   exit 2
a base gate appends [core] fsmonitor = true during acceptance
  -> exit 9; state.json has run_id, brief, branch and status planning; the run folder is kept
restore .git/config; vloop run <brief>
  -> resuming acceptance of <run id>: round 1                 no plan session; gate review; work; exit 0
     the plan commit includes both run folders; vloop metrics <brief> counts both plan sessions' cost
a stub session that sleeps 2 s and reports duration_ms 20
  -> its record has wall_ms >= 2000 and duration_ms 20; vloop metrics shows the plan's time from wall_ms
a stub work session denied Bash "rm -rf $X/tmp" with HOME's user in it
  -> the record's permission_denials: [{"tool_name": "Bash", "command": "rm -rf $X/tmp"}]
     a 300-character command is kept as its first 200 characters and …
a denial file_path ~/.claude/projects/-Users-<user>-repo/x
  -> the record shows -Users-USER-repo
after the work session of T1 at $1.84:
     work T1 · <m>m<ss>s · $1.84 · run $<spent> / $40.00
run.cost-ceiling 4, sessions at $1.10 each
  -> "   cost: $2.20 spent, 50% of the $4.00 ceiling" once; "   cost: $3.30 spent, 80% of …" once
vloop status while a stub work session sleeps on T1
  -> running: work T1 for 0m0<s>s · $<spent> of $40.00     exit 0
a binary built with -ldflags commit stamp           -> vloop version as today
a binary whose build info says module v0.8.0, no stamp
  -> vloop 0.8.0 (module v0.8.0, plugin 0.8.0, go<v> <os>/<arch>); --json has "module": "v0.8.0"
a brief that delivered only *.sh and *.md lines with metrics.stacks = ["go"]
  -> vloop metrics: its code and test columns read n/a; the summary adds
     "<n> delivered lines match no stack in metrics.stacks"
```

The gates thread the values they compute (paths, ids, costs, counts) rather
than hardcoding them; a module version is injected through a test seam, not by
publishing anything.

**Real-data check.** On this repository, `vloop metrics --json` for every
consumed brief equals v0.8.0's except the time fields R3 changes (wall, and plan
time where a record has `wall_ms` — none before this run); the run record lists
each changed wall time. `vloop defect list`, `intervention list` and
`brief list` read every record as before.

## Out of scope

- **B13, docs and onboarding**; **B14, the learning loop**; **B15, code
  health**; **B16, the verification phase**; **B17, `vloop release`**.
- Streaming sessions (`stream-json`), a heartbeat inside a session, or checking
  the cost ceiling inside one.
- Writing `run.log` live; notifications from the operator skill.
- Rewriting past records: the two committed records that keep a user name are
  the operator's to mask after the run, not a session's (S-4).
- A cost ceiling check between plan rounds; changing the ceiling's default.

## Constraints

- Go as pinned in `go.mod`; no new dependencies. Every gate builds for
  `windows` and `linux` as well as the host; OS-specific behaviour is a pure
  function tested on every OS.
- Gates run on macOS's BSD tools: no empty alternatives in `grep -E`, no
  `sed -i` without a suffix, no `\+` or `\|` in basic regexes, no `date -d`.
  A scratch directory is `mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX"`.
- Sessions never move refs and write nothing under `.vloop/` but `.vloop/tmp/`
  (S-4): no task's work, and no gate, may require a session to change this
  repository's records, config or state. Tests build their repositories in
  temporary directories and address them with `git -C`.
- **The gate model applies** (`plugin/skills/plan/SKILL.md`): each gate judges
  this brief's contract by running the built binary, with its judge in the
  verify or its gate folder; no gate runs the repository's tests — this
  repository's `[[check]]` does, after every iteration.
- No gate sleeps longer than 3 seconds or depends on wall-clock precision finer
  than a second.
- **No task may weaken a check to pass.** These earlier tests **must** change,
  and only as stated:
  - the exact `.git/config` messages: `TestRunGitConfigPlanted`
    (`cmd/vloop/run_git_test.go`) and the hooks/config subtests of
    `TestWorkedExampleB8` (`cmd/vloop/b8_e2e_test.go`) — the key is named;
  - `TestSessionFieldMapping` and `TestSessionDenialsKeepToolAndPath`
    (`internal/driver/session_test.go`) — the new fields and `command`;
  - `TestRun15TelemetryContract` (`cmd/vloop/run_safety_test.go`) if it lists
    the record's keys;
  - the version tests: `TestVersionJSON`, `TestVersionText`
    (`internal/cli/root_test.go`), `TestDoctorSelfHostingUnstamped`
    (`internal/cli/doctor_test.go`) and the `config list`/version line of
    `TestWorkedExampleEnglishSession` (`cmd/vloop/e2e_test.go`);
  - `TestWorkedExampleEnglishSession`'s full `config list` (the new key);
  - the metrics tests whose wall time or table changes:
    `TestMetricsCommandSummary`, `TestMetricsCommandCrossBrief`
    (`internal/cli/metrics_test.go`), `TestTimeCostWorkedExample`
    (`internal/metrics/time_test.go`), and the worked-example summaries of
    B3, B4 and B6 (`cmd/vloop/b3_e2e_test.go`, `b4_e2e_test.go`,
    `b6_e2e_test.go`) — each changed number listed in the task's notes;
  - the run-output tests that slice the log, where the per-session line lands
    between the lines they match (`cmd/vloop/run_gates_test.go`,
    `run_iterate_test.go`, `run_halt_test.go`, `run_safety_test.go`);
  - `TestRunGateReviewFailsTwiceThenResumes` and
    `TestRunResumeAllowsStateEdits` only if their resume paths change wording;
  - the status tests (`internal/cli/status_test.go`) only where a live lock is
    planted;
  - `docs/guide/commands.md`'s staleness test (regenerate, as always).
- Repo-relative paths everywhere. No absolute paths in any file or commit
  message.

## Shape

9 to 12 tasks, each independently verifiable by a single command. Order is
load-bearing:

1. **The `.git/config` guard (R1)** — first, so no later task's run can trip
   on an editor key.
2. **Acceptance resume (R2)**, with the stamp on load and the passed-review
   rule.
3. **Session timing (R3)** in the record, then in the metrics.
4. **Denials and masking (R4).**
5. **Run visibility (R5)**: the per-session line and warnings, then the lock and
   `status`.
6. **Version (R6)** and **unmeasured size (R7)**.
7. **The domain and the guides.**
8. **Close:** the worked example line for line, and the real-data check.

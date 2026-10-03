---
name: vloop-v1-review
description: The v1.0 review of the whole package, done in the design act on 2026-10-02 — five read-only passes, each finding checked against the code — and which brief fixes each: B8 the security and correctness pass, B9 the quality pass
kind: design-note
status: draft
created: 2026-10-02
---
# vloop — the v1.0 review

Done on 2026-10-02 in the design act for B8, at `main` `271e5b0` (the v0.7.0 release `e18f80e` changes only versions), by five
read-only review passes: command execution; file writes, paths and secrets;
what sessions may do; docs against the binary; code health. The architect
checked the load-bearing claims against the code before triage. Decided with
the operator the same day: **B8 fixes the security and correctness findings and
B7's carry-overs; B9 the quality findings; v1.0 is tagged after B9.** Amended
2026-10-03: v1.0 is tagged after B8; the quality pass moved to the next release
(`docs/design-notes/vloop-horizon.md` → Now).

B8's findings are pinned, one subsection each, in
`docs/briefs/B20261002-2135-vloop-v1-security.loop-brief.md` (F1–F16, C1–C3);
this note only maps them. B9's are listed here in full, for its design act.

## B8 — security and correctness

| Brief | Finding | Where (at `271e5b0`) |
| --- | --- | --- |
| F1 | a bare `vloop run` plans the newest brief whatever its status; no brief check or `depends-on` check (B-2, B-6) | `internal/driver/plan.go` `LatestBrief`; `internal/cli/run.go` |
| F2 | `git add -A` commits untracked files such as `.env` into the plan and iteration commits | `internal/driver/plan.go`, `iterate.go` `commit` |
| F3 | gates and sessions have no timeout or process group; a background server or watcher hangs the run forever | `internal/driver/iterate.go` `runGate`, `session.go`, `internal/state/gate.go` |
| F4 | the fence allows writes to `.git/`; the driver's own `git commit` runs session-planted hooks, fsmonitor, filters | fence; `internal/driver/plan.go` `git` |
| F5 | the refs check wraps sessions, not gates | `internal/driver/iterate.go` |
| F6 | sessions can write session records (cost ceiling), `.vloop/config.toml` (reviewer model), defects | fence; `iterate.go` `spend`, `resolve.go` |
| F7 | `vloop task gate` runs the verify on disk: edit, run, restore goes unseen | `internal/cli/task_gate.go`; `internal/driver/safety.go` |
| F8 | a review session's edits are committed unreviewed | `internal/driver/iterate.go` |
| F9 | gate files matched by substring, `/` only, quoted paths, current task only | `internal/driver/gates.go` |
| F10 | deny-list gaps: `git -C`/`-c` forms, merge, cherry-pick, push to a URL, config, reflog, gc…; `find -delete/-exec`, `--output`, `rm -fr`; `vloop -C …`, `vloop brief new` | `fence/settings.json` |
| F11 | masking replaces the user name in serialized JSON: short or numeric users corrupt records and can disable the cost ceiling; `HOME=/`; Windows forms | `internal/driver/session.go` `Mask` |
| F12 | gate logs and denial records committed with secrets | `internal/driver/iterate.go`, `session.go` |
| F13 | symlinks followed in `.vloop/tmp/`; racy lock; `run_id` used in paths unchecked | `internal/driver/iterate.go`, `safety.go`, `plan.go`; `schemas/state.v1.json` |
| F14 | export leaks a local remote path or a URL query token (M-7) | `internal/cli/export.go` |
| F15 | gate `shell` not checked at run time; `cmd /C` mis-quoted on Windows; two gate runners | `internal/state/gate.go`, `internal/driver/iterate.go` |
| F16 | `go install` stamps no commit and doctor's self-hosting check passes it | `internal/cli/doctor.go`, `cmd/vloop/main.go` |
| C1 | D20261002-1425: presets miss `*.e2e-spec.ts` and `test/` | `internal/classify/presets.go` |
| C2 | no keep-awake: B7's acceptance run stalled through idle sleep (I20261002-1426) | `internal/cli/run.go` |
| C3 | the README's completeness test forces it to repeat every flag and key; 199/200 lines | `internal/cli/readme_test.go` |

## B9 — quality

### README

- Stale: it says running the tasks "is next" and that briefs "are what vloop
  handles today"; `vloop run` has run tasks since B6.
- No working install or first run: no prerequisites (Go version, git, the
  `claude` CLI, workspace trust), and the plugin step names a marketplace that
  does not exist yet (the repository is private, no release).
- Mostly reference material (the run, metrics and defect rows, the config
  table) — after B8's C3 it no longer has to be. Proposed outline, about 150
  lines: what it is; prerequisites; install; quickstart (init, trust, doctor,
  brief new, check, run); the loop in one paragraph; what a brief is; everyday
  commands; config essentials; exit codes; skills; guides.
- Smaller: the interventions index path exists only in this repository; the
  export row omits interventions; list keys described as TOML arrays in the env
  column (env values are comma-joined).

### Guides

- The run exit-code table: exit 1 also covers mid-run failures and a failed
  plan session, not only "a refusal before anything ran"; exit 2 is both usage
  and blocked.
- `configuration.md`: the env-name rule omits that `-` becomes `_`.
- `metrics.md`: the synopsis omits `--workspace` and `metrics export`.
- `evals.md`: "the four skills that `vloop run` starts" — it starts three.
- Shell-loop references (`concepts.md`, `metrics.md`) make sense only in this
  repository; mark them as legacy.
- `doctor`'s check list omits `self-hosting`.

### Errors and exit codes

- Any error that is not a `Problem` exits 2 (usage): about 75 bare `return err`
  in `internal/cli`. Make exit 2 opt-in.
- Raw Go and git messages reach the user (`git log … exit status 128: fatal: not
  a git repository`, `invalid character 'b' looking for beginning of object key
  string`, `✗ schema: : not valid JSON`).
- Not actionable or misleading: `no plan:` without the next step; `no task T99`
  without `task list`; `brief check` on a missing file says `not a loop brief`;
  `metrics nope` says `no runs` for a brief that does not exist; `-C
  /nonexistent` says "no plan"; `plugin path` outside a repository creates
  `.vloop/` in the working directory; a schema-invalid plan makes `status`
  print an empty line and exit 0; `brief check` on only drafts prints `briefs
  ok`.
- "One line starting with `vloop: `" is not true of `run`'s preflight or
  `doctor`; qualify the claim or change the output.

### Help

- Stale short descriptions: `run` ("…and commit the plan"), `brief` ("Check loop
  briefs"), `task` ("Inspect the plan's tasks").
- No `Long` help anywhere: the root help does not say what a brief is or where
  to start; `run --help` does not say how the brief is chosen or how to resume.
- Inconsistent verbs (Print, Show, List) and field lists.

### Code health

- Frontmatter parsed five ways (`brief`, `runs/release.go`, `defect`,
  `intervention`, `cli/close.go`), with different trimming and quoting rules
  and no BOM handling.
- Five git wrappers; the env filter duplicated in `session.go` and
  `iterate.go`; "the plan at the last commit" twice (`metrics.planSHA`,
  `defect.CheckTask`); `driver.counts` re-implements `state.Plan.Counts`; budget
  keys listed in three places; defect id generation copied into `closeDryRun`.
- Long functions: `iterate()` 241 lines, `runInit` 172, `runDoctor` 144,
  `runBriefClose` 143, `Plan` 137.
- Metrics correctness: run folders named in local time while records are UTC
  (a DST fall-back reorders them and mis-pairs gate failures with commits);
  the convergence ratio divides this run's iterations by every done task,
  including earlier runs'; first-pass ignores a task reverted by a gate
  regression and redone; missing records round-trip through formatted strings.
- Ignored errors that matter: `copyReport` write errors; a failed restore
  reported as restored; `Sync` on the plan; errors on the halt path.
- Cross-platform: the trust lookup in `~/.claude.json` compares backslashed,
  case-sensitive paths (Windows); brief paths under symlinked or case-variant
  roots; the pre-commit hook check uses the executable bit (Windows) and
  `.git/hooks` (worktrees); `brief close` rewrites every status line and leaves
  mixed line endings in a CRLF brief; two `time.Now()` calls in close.
- `cmd/gendocs` writes to any single argument, so `gendocs --help` wrote a file;
  the tracked `--help` at the repository root (added in `a9b7a4b`) is that
  file.

### Tests

- No test uses `t.Parallel()`; the `cmd/vloop` run tests take most of the
  suite's two minutes.
- `TestWorkedExampleB6RealData` clones the real repository and fails when it is
  edited during the run.
- `TestMain` in `cmd/vloop` panics on a build failure without cleanup.
- The run harness skips on Windows, so no Windows path has a test.

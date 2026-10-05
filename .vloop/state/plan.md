# Plan — B20261005-0933-quality-pass

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **11/17 done** · iteration 17

**Brief:** `docs/briefs/B20261005-0933-quality-pass.loop-brief.md` · **Updated:** 2026-10-05T22:14:29Z

## Progress

- [x] **T1** — Make the cmd/vloop suite run its tests in parallel and the real-data test immune to a changing working tree
- [x] **T2** — Add one frontmatter reader and rewriter and move defects and interventions onto it
- [x] **T3** — Move briefs, run records and brief close onto the frontmatter package; close rewrites only the first status line
- [x] **T4** — Make exit 2 opt-in, for usage only; every other failure exits 1
- [x] **T5** — Say what to do next: no plan, no task, no brief, a -C that names no directory, git outside a repository
- [x] **T6** — Report a plan that is not valid JSON, or not a valid plan, with its location and the next step
- [x] **T7** — Fix the remaining messages: drafts-only brief check, plugin path outside a repository, intervention show and migrate paths
- [x] **T8** — Compare trust paths and find the pre-commit hook the way each OS and git do
- [x] **T9** — Report the driver's swallowed errors, make gendocs refuse a flag-like argument, and fix the CLAUDE.md section · 1 attempt(s)
- [x] **T10** — Name new run folders in UTC and order existing ones by their records' timestamps
- [x] **T11** — Count first-pass and convergence right: a redone task is not first-pass, and convergence counts this run's closes · 1 attempt(s)
- [ ] **T12** — Make doctor's plugin check tell the truth about the --plugin-dir plugin, a disabled plugin and a failing claude · **blocked**
- [ ] **T13** — Rewrite the help: current Shorts, consistent verbs, Long help, and the regenerated command reference
- [ ] **T14** — Fix the two evals whose graders misjudge: plan-checks-its-gates' fixture grader and operate-proposes-options' judges
- [ ] **T15** — Bring the guides and the domain's CLI conventions up to v2
- [ ] **T16** — Rewrite the README as a newcomer's page, last
- [ ] **T17** — Close: the worked example as a committed test, and the real-data check against v2.0.0-beta.2

## Tasks

### T1 — Make the cmd/vloop suite run its tests in parallel and the real-data test immune to a changing working tree

`done` · depends on: none

The repository's go check takes about nine minutes after every iteration, almost all of it cmd/vloop's end-to-end tests run one at a time; this task goes first so every later iteration's check is cheaper. Every cmd/vloop test that builds its own temporary repository and changes no process-wide state (environment, working directory) calls t.Parallel(); one that must stay serial says why in a comment. TestWorkedExampleB6RealData clones the repository's HEAD and must stop failing when someone edits the working tree while it runs. The speed target (at most half the base time on the same machine) is measured once by the operator at verification, never by a gate.

**Acceptance**

- Every top-level test in cmd/vloop that builds its own temporary repository and sets no environment variable or working directory of the test process calls t.Parallel() — including TestRun01HappyPath, TestRun09DependencyOrder, TestWorkedExampleB3Commands, TestRunGateReviewPasses and TestRunChecksScopedToWhatChanged.
- A cmd/vloop test left serial carries a comment saying why (what process-wide state it touches); no test made parallel calls os.Setenv, t.Setenv, os.Chdir or t.Chdir, or mutates a package-level variable another test reads.
- TestWorkedExampleB6RealData still clones the repository's HEAD (not its working tree) and still runs vloop run in the clone; it passes while files in the repository's working tree are edited or added during its run, and it still fails if the run writes into the real repository's tracked files or refs.
- Parallel tests do not share a temporary directory, stub, home or port; each test's harness state is its own.
- No production code changes in this task, and no test's assertions are loosened to make it parallel.
- `go test ./cmd/vloop/` passes with -count=1, and `go test -race ./cmd/vloop/ -run 'TestRun0'` reports no race.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T1/gate.sh
```

</details>

### T2 — Add one frontmatter reader and rewriter and move defects and interventions onto it

`done` · depends on: T1

Frontmatter is parsed five ways today, with different trimming and quoting and no BOM handling, so `defect set` and `intervention set` on a CRLF or BOM record fail or corrupt it. This task adds the one package that reads and rewrites the frontmatter of Markdown records — a leading BOM ignored on read and preserved on write, CRLF read and written back as CRLF, no mixed endings, and a rewrite that touches only the lines it changes — and moves defect and intervention reading and setting onto it. T3 moves briefs, run records and brief close onto the same package.

**Acceptance**

- One package (under internal/) owns reading and rewriting Markdown frontmatter; internal/defect and internal/intervention read and rewrite their records only through it — no frontmatter splitting, line-ending or BOM handling of their own remains in either.
- A record with a leading UTF-8 BOM reads like the same record without it, and a rewrite keeps the BOM.
- A record with CRLF endings reads like its LF twin, and a rewrite writes every line CRLF; a rewrite never leaves a file with mixed endings.
- A set changes only the lines of the fields it sets; every other byte of the file, body included, is unchanged.
- `vloop defect set` and `vloop intervention set` on CRLF and BOM + CRLF records exit 0 and change the field; `defect list`, `intervention list` and `intervention show` read such records.
- Values are parsed exactly as today for LF records without a BOM: quoting, trimming and the fields read are unchanged, so every existing record in .vloop/defects and .vloop/interventions reads as before.
- The package has its own table-driven tests (LF, CRLF, BOM, BOM + CRLF, mixed input, quoted values, a missing or unterminated frontmatter), and defect and intervention tests cover set on a CRLF and a BOM record; the checks run them.
- No refactor from B12's list is folded in (one git wrapper, the env-filter copies, planSHA, driver.counts, the budget keys, the id generators, splitting long functions): a diff that does one of these fails review.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T2/gate.sh
```

</details>

### T3 — Move briefs, run records and brief close onto the frontmatter package; close rewrites only the first status line

`done` · depends on: T2

Briefs are read by internal/brief and by runs (FrontmatterStatus, release.go), and brief close rewrites a brief by hand: a BOM or CRLF brief fails `brief check` with "missing frontmatter", and `brief close` rewrites every `**Status:** ready to plan` line it finds (even one quoted mid-line) and appends its run record in LF, leaving mixed endings. This task moves every brief and run-record reader and the close rewrite onto T2's package, so close rewrites only the first `**Status:**` line it means to, as its comment says, and reads the clock once.

**Acceptance**

- internal/brief, internal/runs (the brief status and release reads) and the brief close command read and rewrite frontmatter only through T2's package; none keeps its own frontmatter parser.
- `vloop brief check` and `vloop brief list` read a ready brief saved with a BOM and CRLF endings exactly as its LF twin (same name, status, problems).
- `vloop brief close` rewrites only the first line that starts with `**Status:** ready to plan` (what precedes `**Status:**` on that line is kept); a later such line and a `**Status:**` quoted mid-line are left as they were.
- On a BOM + CRLF brief, brief close keeps the BOM and writes every line CRLF, including the run record it appends: the closed file equals the close of its LF twin with CRLF endings and the BOM.
- brief close reads the clock once: the date in the status line, the run record and any defect it records come from one time value.
- Tests cover brief check and list on a BOM + CRLF brief and brief close on a CRLF brief with two status lines and a quoted one; the checks run them.
- No refactor from B12's list is folded in (one git wrapper, the env-filter copies, planSHA, driver.counts, the budget keys, the id generators, splitting long functions): a diff that does one of these fails review.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T3/gate.sh
```

</details>

### T4 — Make exit 2 opt-in, for usage only; every other failure exits 1

`done` · depends on: T3

Any error a command returns that is not wrapped as a problem exits 2 today, so about 84 failures that are not usage exit 2 and a script cannot tell a typo from a failure. This task inverts the default: exit 2 only for an unknown command or flag, a wrong argument count, and an invalid value for a flag or a settable field (config set, task set, defect set, intervention set, --by and the like); every other failure exits 1. "Invalid value" means a value checked against its vocabulary or format; a missing file, brief, task or record is not usage. The guide's account of the change is T15's.

**Acceptance**

- The default for an error returned from a command is exit 1; exit 2 is produced only by an explicit usage error (a type or wrapper the CLI package defines) and by cobra's own unknown-command, unknown-flag, flag-parse and argument-count errors.
- `vloop run docs/briefs/nope.md` exits 1 with its message unchanged (`vloop: brief not found: docs/briefs/nope.md`); vloop run's own endings (R-3, including 2 for blocked) are unchanged.
- `vloop brief check missing.md` exits 1 with exactly `vloop: no such file: missing.md`.
- `vloop intervention set <id> decided <n>` and `recommended <n>` with a number outside the record's options exit 2, as does any other invalid value for a settable field; unknown commands and flags and wrong argument counts stay 2.
- No non-usage failure path still exits 2: the review checks the diff for errors returned without the usage marker that are not usage, and for usage errors (invalid values, argument counts) that now exit 1.
- Tests pin the exit code of every row of the brief's Q2 table and of representative usage errors (unknown flag, argument count, invalid flag value, invalid field value); the checks run them.
- No earlier test is deleted or loosened to pass; an earlier test changes only where this task's contract changes what it pins, and the proposal's notes list each such test with the expectation it had and has now.
- No refactor from B12's list is folded in (one git wrapper, the env-filter copies, planSHA, driver.counts, the budget keys, the id generators, splitting long functions): a diff that does one of these fails review.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T4/gate.sh
```

</details>

### T5 — Say what to do next: no plan, no task, no brief, a -C that names no directory, git outside a repository

`done` · depends on: T4

Several refusals name a symptom but not the next step, or leak raw git text: `no plan: <path>`, `no task T99`, `no runs for nope` for a brief that does not exist, "no plan" for `-C /nonexistent`, and `git log … exit status 128` outside a repository. This task makes each one line on stderr, exit 1, in the exact words the brief quotes, naming the command to run next. The plan-file and remaining messages are T6's and T7's.

**Acceptance**

- With no plan, every command that needs one (task show, task list, status, and the others) prints exactly `vloop: no plan — vloop run <brief> makes one` and exits 1.
- An unknown task id prints exactly `vloop: no task T<n> — vloop task list shows the plan's tasks` and exits 1, for every task subcommand that takes an id.
- `vloop metrics <name>` for a brief that does not exist prints exactly `vloop: no brief <name> — vloop brief list shows the briefs` (exit 1); a brief that exists with no runs keeps `vloop: no runs for <name>` (exit 1).
- `-C <dir>` naming no directory makes every command print exactly `vloop: -C <dir>: no such directory` and exit 1, before anything else runs; nothing is created at <dir>.
- A command that needs git, run outside a git repository, prints `vloop: <dir> is not a git repository` (exit 1) — no raw `git … exit status 128` or `fatal:` text reaches the user; `vloop metrics` in such a directory no longer exits 0.
- Tests pin each message and exit code above; the checks run them.
- No earlier test is deleted or loosened to pass; an earlier test changes only where this task's contract changes what it pins, and the proposal's notes list each such test with the expectation it had and has now.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T5/gate.sh
```

</details>

### T6 — Report a plan that is not valid JSON, or not a valid plan, with its location and the next step

`done` · depends on: T4

A plan with a JSON syntax error reaches the user as Go's `invalid character 'b' looking for beginning of object key string`, task validate prints `✗ schema: : not valid JSON` with an empty field, and a plan that fails its schema makes `status` print an empty line and exit 0. This task names the file, the line and the column of a syntax error, sends a schema-invalid plan to `vloop task validate`, and makes status and status --json exit 1 on either, with the error under --json.

**Acceptance**

- A plan that is not valid JSON makes `vloop status` print exactly `vloop: .vloop/state/state.json is not valid JSON (line <l>, column <c>)` and exit 1, where <l> and <c> are the 1-based line and column of the character the JSON decoder rejected (`{bad` is line 1, column 2).
- A plan that is valid JSON but fails the state schema makes `vloop status` print exactly `vloop: .vloop/state/state.json is not a valid plan — vloop task validate lists the problems` and exit 1.
- `vloop status --json` exits 1 on either, with one JSON object `{"error": "<the message without the vloop: prefix>"}` on stdout.
- `vloop task validate` and `vloop schema validate <schema> <file>` on a file that is not JSON print `<path>: not valid JSON (line <l>, column <c>)` with no empty field, and exit 1.
- A valid plan's status, status --json and task validate output are unchanged.
- Tests pin these messages for a syntax error on the first line and on a later line, and for a schema-invalid plan; the checks run them.
- No earlier test is deleted or loosened to pass; an earlier test changes only where this task's contract changes what it pins, and the proposal's notes list each such test with the expectation it had and has now.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T6/gate.sh
```

</details>

### T7 — Fix the remaining messages: drafts-only brief check, plugin path outside a repository, intervention show and migrate paths

`done` · depends on: T3, T4

Four smaller messages mislead: `brief check` on only drafts says `briefs ok` when it checked nothing; `plugin path` outside a repository creates `.vloop/` wherever it is run (breaking C-1's nothing-global rule in spirit); `intervention show` hides the schema and options counts and prints an empty `brief:`; and `intervention migrate --dry-run` prints bare file names. This task makes each say what is true.

**Acceptance**

- `vloop brief check <paths>` where every path is a draft prints `nothing checked: <n> draft brief(s) skipped` instead of `briefs ok` and exits 0; when at least one brief is checked the summary is as before.
- `vloop plugin path` outside a vloop repository (no `.vloop/` found, whether or not in a git repository) prints exactly `vloop: not in a vloop repository — run vloop init first`, exits 1, and writes nothing anywhere; inside one it extracts and prints `.vloop/tmp/plugin/<version>` as before.
- `vloop intervention show` prints a `schema: <schema>` line and an `options: <n>` line among the record's fields, and omits the `brief:` line when the record has no brief.
- `vloop intervention migrate --dry-run` prints each record it would migrate as a repo-relative path (`.vloop/interventions/<id>.md`), from any working directory in the repository, and writes nothing.
- Tests pin each of these; the checks run them.
- No earlier test is deleted or loosened to pass; an earlier test changes only where this task's contract changes what it pins, and the proposal's notes list each such test with the expectation it had and has now.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T7/gate.sh
```

</details>

### T8 — Compare trust paths and find the pre-commit hook the way each OS and git do

`done` · depends on: T1

Two checks fail off the happy path: the trust lookup in ~/.claude.json compares paths byte for byte, so a Windows key with backslashes or a key in another letter case on Windows or macOS is not found; and the pre-commit hook check looks in `.git/hooks` with the executable bit, which misses a linked worktree (where `.git` is a file) and every hook on Windows. This task makes both OS-aware pure functions, tested on every OS, used by doctor and run's preflight.

**Acceptance**

- The trust lookup normalises both the repository path and each key in ~/.claude.json to `/` separators and compares them case-insensitively on Windows and macOS (case-sensitively elsewhere); the symlink-resolved root is still tried.
- The pre-commit hook check finds the hooks directory with `git rev-parse --git-path hooks` (which honours core.hooksPath and linked worktrees); on Windows a hook counts by its presence, elsewhere by its executable bit.
- Both decisions are pure functions taking the OS (or its rule) as a parameter, with table tests covering Windows, macOS and Linux cases that run on every OS; the checks run them.
- `vloop doctor`'s trust line and `vloop run`'s preflight hook warning use these functions; their messages are unchanged.
- No refactor from B12's list is folded in (one git wrapper, the env-filter copies, planSHA, driver.counts, the budget keys, the id generators, splitting long functions): a diff that does one of these fails review.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T8/gate.sh
```

</details>

### T9 — Report the driver's swallowed errors, make gendocs refuse a flag-like argument, and fix the CLAUDE.md section

`done` · 1 attempt(s) · depends on: T1

The rest of the review's correctness findings: the driver ignores a failed write of a session's report and says "restored" even when a restore fails, and replaces the plan without syncing it to disk; `cmd/gendocs --help` wrote a file named `--help`, which is still tracked at the repository root; and the CLAUDE.md section `vloop init` writes lists only plan, work and review sessions and names `vloop-operator`, a skill the plugin does not ship. The gate sees only the last two; the driver's errors are the review's to judge from the diff.

**Acceptance**

- A report the driver fails to copy into the run folder (copyReport: MkdirAll or WriteFile failing) is reported as an error in the run log, not silently dropped.
- A restore that fails (restoring a session's changed inputs, or reverting a gate's tree changes) is reported as a failure naming the path — never logged as "restored" — and the iteration or run does not proceed as if it had been restored.
- The plan is written to a temporary file, synced to disk (File.Sync, error checked) and only then renamed over the old one.
- `cmd/gendocs` refuses any argument that starts with `-` (usage message on stderr, non-zero exit, nothing written); `go generate ./...` still regenerates docs/guide/commands.md.
- The tracked file named `--help` at the repository root is deleted.
- The CLAUDE.md section `vloop init` and `vloop upgrade` write lists the session kinds plan, gate review, work and review, and names `/vloop:operate` as the operator's playbook (no `vloop-operator`); internal/install's tests that pin the section's text are updated to it.
- Tests cover the gendocs refusal, the section text and the driver's error paths where a test can reach them (for example an unwritable reports directory); the checks run them.
- No refactor from B12's list is folded in (one git wrapper, the env-filter copies, planSHA, driver.counts, the budget keys, the id generators, splitting long functions): a diff that does one of these fails review.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T9/gate.sh
```

</details>

### T10 — Name new run folders in UTC and order existing ones by their records' timestamps

`done` · depends on: T1

Run folders are named in local time while every record inside is UTC, and folders are ordered by name, so a DST fall-back reorders runs and mis-pairs gate failures with commits. New folders are named in UTC; existing folders keep their names (renaming would break the run records that cite them) and are ordered by their records' timestamps wherever folder order matters. The real-data check (T17) requires every consumed brief's metrics to stay as they are.

**Acceptance**

- vloop run names a new run folder `<YYYYMMDD-HHMMSS>` from the UTC clock whatever the local zone, keeping the `-2`, `-3` suffix rule.
- Run folders are ordered by the earliest timestamp of their records (iteration and session records), falling back to the name only for a folder with no timestamped record; existing folders are not renamed.
- Every consumer that depends on folder order (metrics, the missing-records list, defect derivation, gate-failure/commit pairing, close) sees the timestamp order.
- Tests cover a UTC folder name under a non-UTC zone and two folders whose names sort opposite to their records (a DST fall-back); the checks run them.
- No refactor from B12's list is folded in (one git wrapper, the env-filter copies, planSHA, driver.counts, the budget keys, the id generators, splitting long functions): a diff that does one of these fails review.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T10/gate.sh
```

</details>

### T11 — Count first-pass and convergence right: a redone task is not first-pass, and convergence counts this run's closes

`done` · 1 attempt(s) · depends on: T10

Two numbers are wrong. First-pass counts a task as first-pass when every iteration ended done, so a task reverted to pending by a gate regression and redone still counts. The driver's convergence halt (exit 5) divides this run's iterations by every done task in the plan, including those earlier runs closed, so a resumed run that closes nothing is never caught. This task fixes both; the brief's real-data check expects only these numbers to move.

**Acceptance**

- A task with more than one iteration that ended `done` (it was done, reverted by a gate regression, and redone) is not first-pass in `vloop metrics` (text, --json, --by task) or in the close snapshot.
- The convergence check divides this run's iterations by the number of tasks this run closed (done during this run), not by every done task in the plan; with run.convergence-min reached and nothing closed this run, the run halts not converging, exit 5.
- A clean run's first-pass and a single run's convergence behave as before.
- Tests cover a regressed-and-redone task's first-pass and a resumed run that closes nothing; the existing tests that pin metrics affected by this are updated and the proposal's notes list each changed number.
- No earlier test is deleted or loosened to pass; an earlier test changes only where this task's contract changes what it pins, and the proposal's notes list each such test with the expectation it had and has now.
- No refactor from B12's list is folded in (one git wrapper, the env-filter copies, planSHA, driver.counts, the budget keys, the id generators, splitting long functions): a diff that does one of these fails review.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T11/gate.sh
```

</details>

### T12 — Make doctor's plugin check tell the truth about the --plugin-dir plugin, a disabled plugin and a failing claude

`blocked` · **blocked** · depends on: T4

vloop run supplies the plugin itself with --plugin-dir, and no marketplace exists while the repository is private, yet doctor warns "no enabled vloop plugin — run claude plugin install vloop@vloop" on every machine, cannot tell installed-but-disabled from absent, and swallows the error when `claude plugin list` fails. This task makes the check pass with the --plugin-dir hint, name the enable command for a disabled plugin, and warn with the error when the listing fails.

**Acceptance**

- With no vloop marketplace plugin installed, doctor's plugin check passes with exactly `the plugin is supplied by vloop run (--plugin-dir); for an interactive session: claude --plugin-dir "$(vloop plugin path)"`.
- A vloop plugin that is installed but disabled is a warning: `installed but disabled — claude plugin enable vloop@vloop`.
- When `claude plugin list --json` fails or prints something that is not its JSON, the check is a warning that names the error (the command's stderr or exit status, or the parse error) instead of a fixed text.
- An enabled vloop plugin of the binary's version passes, and one of another version keeps its version-mismatch warning.
- Tests cover the four cases with a fake claude; the checks run them.
- No earlier test is deleted or loosened to pass; an earlier test changes only where this task's contract changes what it pins, and the proposal's notes list each such test with the expectation it had and has now.

**From the last attempt:** gate disputed: The gate's plugin() helper greps '^[^ ]* plugin ' with a trailing space, so it cannot match a passing check with an empty message, which is how doctor prints every other message-less pass. — The last clause runs plugin with an enabled vloop plugin of the binary's version; doctor prints '✓ plugin' and the grep returns '' ('does not pass: '''). The acceptance requires only that this case passes, not that it carries a message.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T12/gate.sh
```

</details>

### T13 — Rewrite the help: current Shorts, consistent verbs, Long help, and the regenerated command reference

`pending` · depends on: T5, T6, T7, T12

The help still describes earlier briefs (`run` "…and commit the plan", `brief` "Check loop briefs", `task` "Inspect the plan's tasks"), mixes Print, Show and List for the same kind of command, and has no Long text: the root help does not say what a brief is or where to start, and `run --help` does not say how the brief is chosen or how to resume. This task fixes the Shorts and verbs, adds Long help where the brief lists it, and regenerates docs/guide/commands.md. It runs after the message and exit-code tasks so the help describes what they built.

**Acceptance**

- Shorts: `run` is "Plan a brief and work it, task by task, on a work branch"; `brief` is "Write, check, list and close loop briefs"; `task` is "Show, amend, gate and reset the plan's tasks".
- Every `list` subcommand's Short begins "List ", every `show` subcommand's Short begins "Show ", other Shorts that print one thing or a report use "Show" consistently, and every `set` Short names its settable fields as a sentence.
- The root's Long says what a brief is, gives the loop in three sentences and says where to start (`vloop init`, `vloop brief new`).
- `run`'s Long says how the brief is chosen, what a run does, how to resume, and gives the exit codes with where they are documented (docs/guide/concepts.md).
- `brief`, `task`, `metrics`, `defect` and `intervention` have Long help saying what the group is for and its common flow.
- docs/guide/commands.md is regenerated from the tree and committed; the tests that pin Shorts or help text are updated to the new text.
- No command, flag or behaviour is added or removed; only help text changes.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T13/gate.sh
```

</details>

### T14 — Fix the two evals whose graders misjudge: plan-checks-its-gates' fixture grader and operate-proposes-options' judges

`pending` · depends on: T1

Evals are the skills' gates, and two of them score sound sessions low. plan-checks-its-gates' fixture grader reads the trace, which ends before the planner writes its gate-folder fixtures, so a sound plan scores 0.67 (defect D20261004-1309). operate-proposes-options' two judge graders failed 3-0 on an answer that proposes two real options with a reasoned recommendation, because they judge everything the answer raises (defect D20261005-0757). This task makes the first judge the files themselves and the second judge only the options offered for the disputed gate. Running the evals is the operator's, at verification.

**Acceptance**

- plan-checks-its-gates' fixture grader judges the gate folder's files themselves (a grader whose focus or target reads files, fed by whatever the case needs to put the planner's greet.toml fixtures where it reads), not the trace or the final message.
- With that grader, a gate expecting `hello! world` from a fixture whose `punctuation` key sits after the `[style]` header fails it, and a negative fixture with the key inside `[style]` expecting `hello, world` does not.
- operate-proposes-options' judge graders judge only the options offered for the disputed gate (count, realism, one recommendation with its reason); other problems or observations the answer raises do not lower their score, and their wording says so.
- Both cases stay well-formed (case.yaml, prompt, scaffold, every grader with a known type, a weight and a description) and both scaffolds still build their repositories in an empty directory.
- The proposal's notes say the cases need an operator eval run and give the command; scores and spend go in the run record (the operator's, not this session's).

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T14/gate.sh
```

</details>

### T15 — Bring the guides and the domain's CLI conventions up to v2

`pending` · depends on: T4, T12, T13

The guides still describe v1 in places: the exit-code table says 1 is only a refusal before anything ran, configuration.md omits that `-` becomes `_` in environment names, metrics.md's synopsis lacks --workspace, --interventions and metrics export, the doctor check list misses checks and self-hosting, the skills list says four, and shell-loop references read as if every repository had one. There is also no account of what changed in 2.0, including T4's exit-code change. This task fixes the guides and updates the domain's CLI conventions to the qualified one-line rule.

**Acceptance**

- concepts.md's exit-code table states 1 as "problems or failure, including a failed plan session or a mid-run error" and 2 as "usage" (and, for vloop run, still "blocked", R-3); the general sentence above it says 2 is usage only.
- concepts.md has a "What changed in 2.0" section covering the gate model (state/v2, gate folders, the gate review, [[check]]), intervention/v2 and `vloop intervention migrate`, metrics/v2, and Q2's exit-code change with its table.
- concepts.md's doctor paragraph lists the checks in the order `vloop doctor` prints them, including `checks` and `self-hosting`; its Skills section says the plugin ships five skills and names all five, /vloop:gate-review included.
- concepts.md's plugin install text no longer tells a reader to add a marketplace that does not exist; it gives the --plugin-dir way, matching T12's doctor message.
- configuration.md says a key's environment name is `VLOOP_` and the key upper-cased with `.` and `-` turned into `_`.
- metrics.md's synopsis shows --workspace, --interventions and `vloop metrics export`, and its messages match T5's.
- Every shell-loop reference in the guides is kept and marked "legacy: the shell loop that built vloop's first briefs, in this repository only" (the full phrase at least once, "legacy" on each mention).
- docs/domain/platform/platform-context.md's CLI conventions read: exit `0` ok, `1` problems or failure, `2` usage only; a refusal is one stderr line starting `vloop: `, after any checks `doctor` or a preflight prints. Nothing else in the domain changes.
- The guide tests (internal/cli guide and README completeness tests) pass, updated only where the text they pin changed.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T15/gate.sh
```

</details>

### T16 — Rewrite the README as a newcomer's page, last

`pending` · depends on: T15

A newcomer who reads the README today cannot install vloop or make a first run: it says running tasks is "next", names a marketplace that does not exist, and is mostly reference tables the generated guides now hold. This task rewrites it in the v1 review's outline updated for v2, after everything else, so it describes what this run built. No command table — docs/guide/commands.md is that.

**Acceptance**

- README.md is at most 200 lines, in this order: what vloop is; prerequisites (Go as in go.mod, git, the claude CLI, workspace trust); install (`go install github.com/mvelosop/vloop/cmd/vloop@v<tag>` at a release tag; the plugin through --plugin-dir while the repository is private); a quickstart (vloop init, a [[check]], vloop doctor, vloop brief new, brief check, vloop run, vloop brief close); the loop in one paragraph (plan, gate folders, the gate review, work, gates, checks, review, one commit); the everyday commands in a short list pointing at docs/guide/commands.md; exit codes in brief, pointing at the concepts guide; the guides; "What changed in 2.0".
- There is no command table, and no marketplace install step.
- Any claim that errors are one line starting `vloop: ` carries the qualification (after the checks doctor or a preflight prints).
- The README test keeps its 200-line cap; completeness checks that the README no longer carries (config keys, commands) live in the guides' tests instead, not deleted.
- Every relative link in README.md resolves.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T16/gate.sh
```

</details>

### T17 — Close: the worked example as a committed test, and the real-data check against v2.0.0-beta.2

`pending` · depends on: T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16

The brief ends with a worked example and a real-data check; this task proves both on the finished tree. The worked example becomes a committed end-to-end test in cmd/vloop (as each earlier brief's did), line for line, in temporary directories. The real-data check reads this repository with the new binary and with v2.0.0-beta.2: every consumed brief's metrics --json must be equal except first-pass and convergence where a task regressed and was redone, and the lists must read every record as before. The suite's speed is measured by the operator, not here.

**Acceptance**

- A committed cmd/vloop test (TestWorkedExampleB11 or similar, parallel like its neighbours) runs every line of the brief's worked example against the built binary in temporary directories and asserts its output and exit code; the checks run it.
- Run on this repository, `vloop metrics --json` for every consumed brief (B1–B10) equals v2.0.0-beta.2's except first_pass and iterations_per_closed; the proposal's notes list each changed number with its brief and task (for the operator's run record).
- `vloop defect list`, `vloop intervention list` and `vloop brief list` print exactly what v2.0.0-beta.2 prints on this repository.
- If the comparison finds a difference beyond those two fields, the task fixes its cause in the code (not the test or the data) or halts with an account of it.
- No production behaviour changes in this task beyond fixes the comparison forces, each named in the proposal's notes.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T17/gate.sh
```

</details>


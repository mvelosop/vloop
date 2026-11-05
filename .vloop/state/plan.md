# Plan — B20261105-2002-b37-i88-metrics-by-session-kit-without-python

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **3/6 done** · iteration 4

**Brief:** `docs/briefs/B20261105-2002-b37-i88-metrics-by-session-kit-without-python.loop-brief.md` · **Updated:** 2026-11-05T22:25:37Z

## Progress

- [x] **T1** — Add `vloop metrics --by session` and `by_session` to metrics/v4 · 1 attempt(s)
- [x] **T2** — Commit the kit's fixtures, the stub copilot, gen.sh and the Python's goldens
- [x] **T3** — Port the kit's six reports to a Go program under tools/copilot-acceptance/kit
- [ ] **T4** — Port the kit's probes to Go: probe, summary, hook forms|failure|summary, calls
- [ ] **T5** — Switch the kit's scripts and README to the Go program and delete report.py and hook.py
- [ ] **T6** — Close the run: every gate at once, and the worked example line for line

## Tasks

### T1 — Add `vloop metrics --by session` and `by_session` to metrics/v4

`done` · 1 attempt(s) · depends on: none

Nothing today shows a brief's sessions one by one: metrics only sums them per phase and per task. This task adds the per-session view (P1) as a `--by` value of `vloop metrics`, reading the same session records metrics already sums through internal/runs, so its rows are exactly the sessions the brief's metrics count and their costs sum to cost.total_usd. The kit's reports (T3) print this view in place of their own session tables, so its text form is what the acceptance kit will show operators.

**Acceptance**

- `vloop metrics <brief> --by session` prints one table per brief named (a blank line between briefs, as `--by task`), with the header `run  #  phase  task  agent  model  time  cost  source  credits  turns  denied  error` and columns separated as PrintByTask's; with no brief named it prints the brief table, as `--by task` does.
- There is one row per session record the brief's metrics count: the records in the folders internal/runs assigns to the brief (legacy shell-loop folders included) whose phase is plan, work, review, second-review, gate-review or verify, in the folders' order and then by sequence number. A brief-review session (`.vloop/state/brief-reviews/`) or a record of any other phase is not a row; the rows' cost_usd sum to cost.total_usd within 1e-9 and their permission_denials to the report's permission_denials.
- No second reader of the session records: the rows are built from the folders and sessions internal/runs already loads for metrics (its Session extended with what a row needs, e.g. schema, cost_source, turns/num_turns, the denial entries), not from a separate walk or glob of `sessions/*.json`.
- Text cells are exactly as the brief's P1 says: `run` the folder's last path element; `#` three digits; `task` `-` for none; `agent` `claude` for records with none; `model` `-` when absent; `time` wall_ms else duration_ms in minSec's form; `cost` `$` and two decimals; `source` the record's cost_source or `-`; `credits` two decimals or `-` when the record has none (`0.00` for a 0); `turns` the record's turns, else num_turns, else `-`; `denied` the denial count; `error` `yes` when is_error is true, else `-`.
- Every metrics/v4 `--json` document (with or without `--by`, and each document of `--workspace --json`) carries `by_session`: rows with exactly the keys folder (repo-relative), seq, schema, phase, iteration, task, agent, model, effort, time_ms, cost_usd (unrounded), cost_source, credits, turns, permission_denials (the count), denials and is_error, with null where the brief says null. Each denial is `{"tool_name": <string>, "text": <string>}`: text the first non-empty string of message, command, file_path, tool_input.command, tool_input.file_path, else ""; tool_name "" when absent.
- schemas/metrics.v4.json declares by_session as an optional array of those rows: a snapshot without it validates, a document whose by_session is not an array is rejected. `vloop metrics export` deletes by_session from its brief records as it deletes by_task; the `--workspace` text table is unchanged.
- Refusals, each exit 2 with nothing on stdout: `--by <other>` outside --interventions → `unknown --by "<v>": want task or session`; `--by session --workspace <f>` → `--by session reports one repository's briefs; it takes no --workspace`; `--interventions --by session` → `--by takes kind or phase with --interventions`.
- The flag's usage string is ``break the summary down by `view`: task or session`` (help line `--by view   break the summary down by view: task or session`); docs/guide/commands.md is regenerated with `go generate ./...`; docs/guide/metrics.md gains the usage line and a section with the columns and keys; docs/domain/measurement/metrics.md's table says per-session rows exist and points to that section.
- Every other `vloop metrics` output is unchanged, and the existing tests pass unedited.
- Committed tests (internal/metrics and/or internal/cli) cover every P1 line of the brief's worked example: the row order across two folders, a Copilot record's credits and estimated source, a session/v1 record, a legacy shell-loop record's num_turns, a record with neither turns field, both denial-text examples, is_error, a brief-review session not listed, the three refusals (each its own test), the JSON sums, and a metrics/v4 snapshot without by_session validating. Tests build their fixtures in temporary directories and address them with `git -C`; no test runs a real agent CLI, touches `~/.claude` or `~/.copilot`, or needs the network. A task that ships this behaviour without such tests fails review.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T1/gate.sh
```

</details>

### T2 — Commit the kit's fixtures, the stub copilot, gen.sh and the Python's goldens

`done` · depends on: none

The kit's Go program (T3, T4) must reproduce the Python's output byte for byte, and the Python is deleted in T5; the goldens keep it as the reference after that. This task writes the fixtures, the stub copilot and the generator that runs the starting point's Python on them, and commits the goldens it prints. Nothing here is Go: the goldens must be the Python's own output, so that T3 and T4 are judged against the reference rather than against themselves.

**Acceptance**

- tools/copilot-acceptance/kit/testdata/gen.sh `<python dir> <vloop> <out dir>` is a shell script that runs the Python in `<python dir>` (report.py, hook.py, and calls.py, which it extracts from `<python dir>/bypass.sh` as the brief says) on the fixtures and writes the goldens into `<out dir>`; the goldens it writes are, file for file and byte for byte, the ones committed under tools/copilot-acceptance/kit/testdata/golden/ (the gate runs it with the base's three files and diffs the two trees).
- A golden is the Python's stdout (or a file the Python wrote) on a fixture, unedited, except that the fixture's directory as the Python saw it (resolved with `pwd -P`) is replaced by `@DIR@`, and, in the stub's log, the lines of PWD, OLDPWD, SHLVL and `_` are dropped. No golden holds a machine path; VLOOP_HOOK_PHASE is unset when the goldens are generated. A golden written or edited by hand, or produced by anything but the Python, is a violation.
- Fixture files hold `@DIR@` where the temporary directory goes; gen.sh copies each into a temporary directory resolved with symlinks followed, substitutes that path, and passes it as `<dir>` or `<repo>`. gen.sh writes nothing into the repository tree and leaves no `.pyc` or `__pycache__` anywhere in it.
- There is a report fixture and golden for each of run, second, work, verify, gate-review and plan, each a repository `vloop metrics <run id> --by session` can read with at least 2 rows (the run folder under `.vloop/state/runs/<run id>/` and `docs/briefs/<run id>.loop-brief.md`, folders named in the order of their timestamps); plus a `work` fixture on which vloop metrics exits 1 and a `run` fixture with no run folder. The `second` fixture has one run folder (one run.log, one iterations.jsonl).
- There are probe fixtures and goldens for probe, summary, hook forms, hook failure in each mode (missing, exit1, notjson, timeout), hook summary and calls: their stdout, the files they write (probe.argv, probe.rc, the hook-&lt;x&gt;.argv and .rc files, hooks/vloop-fence.json) and the stub's log.
- tools/copilot-acceptance/kit/testdata/copilot is a committed POSIX sh script, mode 100755, that appends to `$STUB_LOG` its argv (one `arg <a>` line each), `cwd <pwd -P>` and its environment (`env | LC_ALL=C sort`, each line prefixed `env `), then prints the JSON Lines file `$STUB_OUT` names. Each probe fixture's logged review block carries `env STUB_LOG=@DIR@/<file>`, `env STUB_OUT=@DIR@/<file>`, `env PATH=/usr/bin:/bin` and a `cwd` naming a directory the fixture holds.
- Between them the fixtures hold a float, an integral float, non-ASCII text, an emoji, a string with a newline and a tab, `<&>` in a string, and two distinct event types in one probe file, each reaching an output the goldens record (a json.dumps rendering, a repr, a slice).
- This task adds no Go code and no go test; the goldens and gen.sh are the committed test data T3's and T4's tests run against, and gen.sh is how they are regenerated.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T2/gate.sh
```

</details>

### T3 — Port the kit's six reports to a Go program under tools/copilot-acceptance/kit

`done` · depends on: T1, T2

The acceptance kit needs python3 only because its reports and probes are Python. This task creates the kit's Go program (package main in this module, run as `go run ./tools/copilot-acceptance/kit <command>` like tools/agent-eval) with its `report` command (P2): the six reports report.py prints, byte for byte, except that each report's session table becomes the P1 view (`vloop metrics <run id> --by session`, fenced) plus the record checks and denial lines the kit reads itself. T4 adds the probe commands to the same program and T5 switches the scripts to it.

**Acceptance**

- `go run ./tools/copilot-acceptance/kit report <kind> <repo> <vloop> <exit code>`, kind one of run, second, work, verify, gate-review, plan, prints what `report.py <kind> …` printed at the starting point, byte for byte, except the session table; it exits 0 when it printed the report, and 1 with a one-line message on stderr and nothing on stdout when an argument is missing, the kind is unknown, or no command is given.
- The session table of each kind (gate-review's `| round | record | …` one included), from its header row through the blank line after it, is replaced by one Sessions block per qualifying run id (in name order; a run id qualifies when `<repo>/.vloop/state/runs/<id>/` and `<repo>/docs/briefs/<id>.loop-brief.md` both exist), each exactly the P2 block: the heading line with the literal command, a fenced copy of `<vloop> metrics <id> --by session`'s stdout run in `<repo>` (or the one line `vloop metrics exited <code>: <stderr trimmed, newlines as spaces>`), the `Session records: <n>, schema-valid <k> (<schema> <c>, …).` line (`Session records: 0.` with none; records without a schema last as `no schema <c>`), one `- <folder>/sessions/<file>: <schema> INVALID`, `: no schema` or `: unreadable` line per record so marked, and, only when a row has denials, the `Permission denials:` part with one `- <run>/<#> <phase> <tool_name>: <text>` line per denial (P1's text rule, newlines as spaces); each block ends with one empty line. With no qualifying run id the table is replaced by `Sessions: no brief with runs.` and a blank line.
- The record lines and counts come from the kit reading the record files itself (folders in name order, records in sequence order) and running `<vloop> schema validate <the record's own schema> <file>` in `<repo>`, so they do not depend on vloop metrics succeeding; a session file that is not JSON counts in n only and is listed as unreadable.
- Wherever the Python renders a value, the Go program renders it the same way: json.dumps with the input's key order (sorted where the Python passes sort_keys), `", "` and `": "` separators, non-ASCII as lowercase `\uXXXX` (surrogate pairs outside the BMP), no HTML escaping, control characters as the Python escapes them, floats as Python's repr; str()/f-string values as Python writes them (a list's repr, None, True, False, `4.0`, `1e+16`). Where the Python joins two unsorted glob lists (second's run.log and iterations.jsonl), each list is taken in sorted path order, the `*/*/` list first.
- Committed Go tests in the kit compare each report kind on its T2 fixture with its golden, building the expected output from the golden by replacing its first table with the Sessions block(s) the test composes: the fenced stdout of `vloop metrics <run id> --by session` from a vloop built from this tree, run on the same fixture copied into a temporary git repository (`git -C`), and the record and denial lines written literally. The test fails when its own vloop metrics call exits non-zero or prints fewer rows than it pins (at least 2), except on the metrics-exits-1 and no-run-folder fixtures, which have tests of their own; `report` with no kind has a test (exit 1, stderr message). Tests build their fixtures in temporary directories and address them with `git -C`; no test runs a real agent CLI, touches `~/.claude` or `~/.copilot`, or needs the network. No go test needs python3; tests that need a shell skip on Windows.
- The diff changes nothing under tools/copilot-acceptance/kit/testdata/ — the goldens and fixtures are T2's, the Python's output; a golden that seems wrong is reported in the notes, not edited or regenerated from the Go program.
- The kit builds with GOOS=linux and GOOS=windows, adds no module dependency, and writes nothing to the repository tree when run.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T3/gate.sh
```

</details>

### T4 — Port the kit's probes to Go: probe, summary, hook forms|failure|summary, calls

`pending` · depends on: T2, T3

The kit's live probes (report.py probe and summary, hook.py, and bypass.sh's inline calls) are the rest of its Python. This task adds them to the kit's Go program (P3), each reproducing the starting point's Python exactly: the files read and written, the session started (argv, working directory, environment, only the prompt replaced), the prompts and form lists verbatim, and stdout byte for byte. Operators compare these probe results with past ones in the design notes, so any drift in prompts or rendering silently changes what the probes test.

**Acceptance**

- The kit has the commands `probe <dir> <copilot>`, `summary <dir>`, `hook forms <dir> <copilot>`, `hook failure <dir> <copilot> <mode>`, `hook summary <dir>` and `calls <out.jsonl>`, each doing what the Python it replaces does: the same files read and written (probe.argv, probe.jsonl, probe.stderr, probe.rc; hook-&lt;x&gt;.argv, .jsonl, .stderr, .rc; `<COPILOT_HOME>/hooks/vloop-fence.json`; the directories created), the same session started (argv, working directory and the logged environment plus the shims' PATH, with only the prompt replaced), and stdout byte for byte.
- FORMS, PROMPT (report.py's and both of hook.py's), SKILL, MODES (commands and timeouts) and FAIL_CALL are the Python's verbatim; a reworded prompt or a changed form list is a violation the review rules on by comparing with the base's report.py and hook.py.
- `hook forms` and `hook failure` read VLOOP_HOOK_PHASE as hook.py does (default review) and write it into the hooks file's `--phase`; the hooks file is rendered as Python's json.dump(…, indent=2) renders it.
- When copilot-argv.log holds no `/vloop:review` session, probe, hook forms and hook failure print `c5: no review session in the copilot log` on stderr and exit 1. An unknown command or hook failure mode, or a missing argument, exits 1 with a one-line message on stderr and starts no session; otherwise each command exits 0.
- Values are rendered as the brief's P3 list says (json.dumps key order, separators, `\uXXXX`, surrogate pairs, control escapes, Python float repr; json.dump indent=2; repr of a list, None, True, False), and a slice counts characters, not bytes.
- Committed Go tests run each command on the T2 fixtures with the committed stub copilot and compare stdout, the files written and the stub's log (after the @DIR@ substitution and dropping PWD, OLDPWD, SHLVL and `_`) with the goldens; plus a test for the no-review-session case and one for VLOOP_HOOK_PHASE=second-review (the hooks file's bash and powershell commands hold `--phase second-review`). Tests build their fixtures in temporary directories and address them with `git -C`; no test runs a real agent CLI, touches `~/.claude` or `~/.copilot`, or needs the network. No go test needs python3; tests that need the shell stub skip on Windows.
- The diff changes nothing under tools/copilot-acceptance/kit/testdata/; the kit still builds with GOOS=linux and GOOS=windows and adds no module dependency.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T4/gate.sh
```

</details>

### T5 — Switch the kit's scripts and README to the Go program and delete report.py and hook.py

`pending` · depends on: T3, T4

With the Go program complete, the kit's scripts can stop needing python3 (P4). Every script calls `go -C "$top" run ./tools/copilot-acceptance/kit …` where it called python3, with the same arguments and absolute paths, the Python files are deleted, and the README says what the kit needs. The scripts make the call from scratch repositories outside the module, which is why `-C` and absolute directories matter.

**Acceptance**

- run.sh, second.sh, work.sh, verify.sh, gatereview.sh, plan.sh, hook.sh and bypass.sh call `go -C "$top" run ./tools/copilot-acceptance/kit …` where they called python3, with the same arguments, and write the same files; `top` is computed in every one of them as run.sh computes it (hook.sh and bypass.sh included).
- Every script makes its directory argument absolute right after reading it, in a command substitution that leaves its working directory unchanged — `dir=$(mkdir -p "$dir" && cd "$dir" && pwd -P) || die …` (the same for `work` in bypass.sh, with `|| exit 1`), placed after die() is defined (die() moved up where it came later).
- No script tests for or runs python3; their header comments say what each needs without python3, and bypass.sh's header names perl.
- tools/copilot-acceptance/report.py and hook.py are deleted; fail-open.sh is unchanged.
- tools/copilot-acceptance/README.md says the kit needs go, copilot and gh on PATH (claude for the scripts that run Claude sessions, perl for bypass.sh, whose timestamps it takes with `perl -MTime::HiRes`), names no Python, names the Go command (`go run ./tools/copilot-acceptance/kit …`) where it named report.py or hook.py, and says the session tables are `vloop metrics --by session`.
- The scripts carry no committed test of their own (the operator's acceptance runs them against the real CLIs); the kit's Go tests from T3 and T4 stay green.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T5/gate.sh
```

</details>

### T6 — Close the run: every gate at once, and the worked example line for line

`pending` · depends on: T1, T2, T3, T4, T5

Each gate proved its task when it ran; the close proves them together on the final tree and accounts for the brief's worked example line by line, so the verify session and the operator find every line covered by a gate and a committed test, or named as a gap. Nothing new is built here unless a worked-example line turns out to have no committed test.

**Acceptance**

- Gates T1 to T5 all pass on the final tree, run together by this task's gate.
- The work session's notes walk the brief's worked example line for line (P1, the real-data B36 line, P2 and P3) and name, for each, the committed test and the gate (T1–T5) that covers it; a line no committed test covers is fixed in this task (a test added) or reported, not skipped.
- Every configured check is green on the final tree (the `go` check: go test, vet, gofmt, mod tidy, linux and windows builds).

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T6/gate.sh
```

</details>


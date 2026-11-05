---
name: B20261105-2002-b37-i88-metrics-by-session-kit-without-python.loop-brief
description: Add a per-session view to vloop metrics (--by session, by_session in --json), and port the Copilot acceptance kit's Python reports and probes to a Go program run with go run, so the kit needs no python3
kind: brief
status: ready
created: 2026-11-05
issue: 88
seeds: Issue 88 under 69; tools/copilot-acceptance/report.py, hook.py and bypass.sh at the starting point; B36's session records; decided by the assistant under the operator's delegation of 2026-11-02 for the #81–#88 chain
depends-on: [B20261103-0528-b36-i96-i97-plugin-per-repo-and-agent.loop-brief]
---
# Brief — vloop B37: metrics per session, and the acceptance kit without Python

- **Issue:** #88 (under #69). The PR closes it.
- **Starting point:** extends `main` at `300d0903`, plus this brief's ready
  merge. The planner pins that SHA as the **base** every gate compares
  against — never `HEAD~n`.
- **Produced by:** the operator's delegation of 2026-11-02 for a chain of
  briefs on #84, #81, #82, #83, #96–#97 and #88. Written with
  `plugin/skills/create-brief/`; every decision below was taken by the
  assistant under that delegation.
- **Review:** `vloop brief review` (Claude Opus), 13 rounds, every
  finding a `FIX` on precision, none an `ASK`: how the Go output is
  compared with goldens the starting point's Python prints (the first
  table replaced by a literal Sessions block, `@DIR@` in fixtures and
  outputs, a committed stub `copilot` fed through the logged
  environment, gen.sh taking the Python from the base, `calls` extracted
  from `bypass.sh`); Python's `json.dumps` and `repr` rules; the scripts'
  `go -C "$top"` call with absolute directories; the per-session rows'
  phases, sums, `num_turns` and denial text; `export` and `--workspace`;
  the `--by` usage string and `commands.md`; error paths; perl for
  `bypass.sh`; the acceptance's `--ignored` check. Round 13's two
  findings (error paths, an unreadable record) fixed and not re-reviewed:
  they pin cases and add no reach.

## Executive summary

`vloop metrics` reports a brief's sessions only as sums: per phase, per
task. Nothing shows one session: its agent, model, cost and where the cost
came from, its Copilot credits, its turns, what it was denied. The Copilot
acceptance kit (`tools/copilot-acceptance/`) gets that view from
`report.py`, which reads vloop's own records (`sessions/*.json`,
`iterations.jsonl`, the kept reports), none of them specific to an agent.
The kit's live probes (`hook.py`, `report.py probe`, the inline Python in
`bypass.sh`) are Python too. So the kit needs `python3`, which neither
Copilot's CLI nor vloop needs, and once left a `.pyc` in a run's tree (B24's
T5 review).

**The goal:** `vloop metrics <brief> --by session` shows one row per
session, for Claude and Copilot runs alike, and the acceptance kit needs
`go`, `copilot` and `gh` (and `claude` where a script runs Claude
sessions, `perl` for `bypass.sh`), not `python3`.

- **P1 — `vloop metrics --by session`.** One row per session record the
  brief's metrics count: run folder, sequence, phase, task, agent, model,
  time, cost, cost source, credits, turns, permission denials, error;
  `by_session` in `--json`.
- **P2 — the kit's reports in Go.** A Go program under
  tools/copilot-acceptance/kit/, run with `go run`, prints the six
  reports `report.py` prints today; their session tables become the P1 view.
- **P3 — the kit's probes in Go.** The same program starts the probe
  sessions `report.py probe` and `hook.py` start, and prints their
  summaries and `bypass.sh`'s call list.
- **P4 — the scripts and the README.** Every script calls the Go program;
  `report.py` and `hook.py` are deleted; no script checks for or runs
  `python3`; the README says what the kit needs.
- Unchanged: every other `vloop metrics` output (`--json` documents gain
  `by_session`; `export` leaves it out), the session records, the fence,
  `vloop run`. The kit's probes stay acceptance tools, outside the
  CLI.

## What it is

A per-session breakdown in `vloop metrics`, and the Copilot acceptance kit
running on Go alone. After the run an operator reads any brief's sessions
one by one from the CLI, and runs the kit on a machine without Python.

## Why this shape, and what was rejected

- **The per-session view is a `--by` value of `vloop metrics`,** next to
  `--by task`, reading the same session records metrics already sums
  (`internal/runs`). Its rows are exactly the sessions the brief's
  metrics count, so their costs sum to `cost.total_usd`.
  *Rejected:* a new command (`vloop sessions`): a second reader of the
  same records, and a second place to keep in step with metrics.
- **`by_session` joins `metrics/v4` as an optional property,** emitted on
  every `--json` document like `by_task`. Snapshots written at close
  before B37 lack it and stay valid. *Rejected:* a `metrics/v5`: nothing
  existing changes meaning, and a version bump would make every reader
  of v4 snapshots handle two versions for one added array.
- **The kit is one Go program, tools/copilot-acceptance/kit/, package
  `main` in this module,** run as `go run ./tools/copilot-acceptance/kit
  <command>` from the repository root, as `tools/agent-eval` (B33) is.
  `go run` builds in Go's cache and writes nothing to the tree, which is
  what the `.pyc` broke. *Rejected:* a `vloop` subcommand: the probes start
  real Copilot sessions and spend credits; they are the operator's
  acceptance tools, not product (#88). *Rejected:* `jq` in the scripts: one
  more dependency, and the goal is fewer.
- **The Go program reproduces the Python's output byte for byte, except
  where this brief changes it,** and is gated on golden files the
  starting point's Python prints for the same fixtures. The Python is
  the reference while it exists; the goldens keep it after it is deleted.
  *Rejected:* re-specifying every table here: two specifications of the
  same output, and the Python's is the one the design notes were written
  from.
- **The reports' session tables become the P1 view,** printed by the
  `vloop` the script built. *Rejected:* keeping the Python's per-report
  tables in Go: that is the per-session view written a second time,
  which #88 asks to replace.
- **Schema checks of the records stay in the reports,** through `vloop
  schema validate`, which the Go program runs with each record's own
  `schema` field as the schema name. *Rejected:* a shell loop in each
  script: the scripts cannot read a record's `schema` field without a
  JSON reader.

## Binding references

- `docs/domain/domain-model.md` — M-1: metrics are recomputed from raw sources and sums use unrounded values, so a row's cost is the record's, unrounded in JSON
- `docs/domain/measurement/metrics.md` — what is measured per brief and per task, which this brief extends per session
- `docs/domain/execution/session.md` — the session kinds and the record each leaves, which the rows read
- `docs/guide/metrics.md` — the `--by task` view, the `metrics/v4` keys and the summary this view sits beside
- `internal/runs/runs.go` — the session reader: which records a brief owns, their order, and the fields read today
- `internal/metrics/print.go` — `PrintByTask` and its helpers (`minSec`, the dash for a missing value), the style the session table follows
- `internal/cli/metrics.go` — the `--by` flag and its refusals
- `schemas/metrics.v4.json` — the schema `by_session` joins
- `tools/copilot-acceptance/report.py` — the reports, the deny probe and its summary, which P2 and P3 port and whose output is the reference
- `tools/copilot-acceptance/hook.py` — the hook probes and their summary, which P3 ports and whose output is the reference
- `tools/copilot-acceptance/bypass.sh` — its inline Python call list, which P3 ports
- `tools/copilot-acceptance/README.md` — the kit's documentation, which P4 updates
- `tools/agent-eval/main.go` — the precedent for a Go tool under `tools/`, run with `go run`

## Behaviour contract

### P1 — `vloop metrics --by session`

- **Flag.** `--by` takes `task` or `session`. Any other value outside
  `--interventions` is exit 2 with `unknown --by "<v>": want task or
  session`. With `--interventions`, `--by session` is exit 2 with `--by takes
  kind or phase with --interventions`, as any value other than those two
  is today. With `--workspace`, `--by
  session` is exit 2 with `--by session reports one repository's briefs;
  it takes no --workspace`. With no brief named, it prints the brief table,
  as `--by task` does.
- **Rows.** One per session record the brief's metrics count: the
  records in the folders `internal/runs` assigns to the brief (legacy
  shell-loop folders included) whose phase is one of `plan`, `work`,
  `review`, `second-review`, `gate-review` and `verify`, in the folders'
  order, then by the record's sequence number (`NNN` of
  `sessions/NNN-<phase>.json`). The rows' `cost_usd` sum to
  `cost.total_usd` within 1e-9, and their `permission_denials` to the
  report's `permission_denials`.
- **Text,** one table per brief (a blank line between briefs, as `--by
  task`), columns separated as `PrintByTask`'s:

  ```
  run  #  phase  task  agent  model  time  cost  source  credits  turns  denied  error
  ```

  `run` is the folder's last path element; `#` the sequence number,
  three digits; `task` the task metrics assign to the session, `-` for
  none; `agent` as metrics counts it (`claude` for records that carry
  none); `model` the record's `model`, `-` when absent; `time` the
  session's time as metrics counts it (`wall_ms`, else `duration_ms`) in
  `minSec`'s form; `cost` `$` and two decimals; `source` the record's
  `cost_source`, `-` when absent; `credits` two decimals, `-` when the
  record has none; `turns` the record's `turns`, else its `num_turns`
  (the shell loop's records), `-` when it has neither; `denied` the number of
  permission denials; `error` `yes` when `is_error` is true, else `-`.
- **JSON.** Every `metrics/v4` document gains `by_session`, an array of
  rows with the keys `folder` (the repo-relative folder path), `seq`,
  `schema` (the record's `schema`, `null` when absent), `phase`,
  `iteration`, `task` (`null` for none), `agent`, `model` (`null` when
  absent), `effort` (`null` when absent), `time_ms`, `cost_usd`
  (unrounded), `cost_source` (`null` when absent), `credits` (`null` when
  the record has none), `turns` (as in the text, `null` for neither), `permission_denials`
  (the count), `denials` and `is_error`. `denials` lists each permission
  denial as `{"tool_name": <string>, "text": <string>}`: `text` is the
  entry's first non-empty string of `message`, `command`, `file_path`,
  `tool_input.command` and `tool_input.file_path`, else `""`; `tool_name`
  is `""` when absent. `schemas/metrics.v4.json` gains the property,
  optional, so snapshots written before B37 stay valid. `vloop metrics
  export` deletes `by_session` from its brief records as it deletes
  `by_task`; `--workspace --json` documents carry it as they carry
  `by_task`.
- **Docs.** `docs/guide/metrics.md` gains the view (usage line, a section
  with the columns and keys); `docs/domain/measurement/metrics.md`'s table
  says per-session rows exist and where they are described. The `--by`
  flag's usage string is ``break the summary down by `view`: task or
  session`` (help: `--by view   break the summary down by view: task or
  session`),
  and `docs/guide/commands.md` is regenerated with `go generate ./...`.

### P2 — the kit's reports in Go

`go run ./tools/copilot-acceptance/kit report <kind> <repo> <vloop> <exit
code>`, `<kind>` one of `run`, `second`, `work`, `verify`, `gate-review`,
`plan`, prints to stdout what `report.py <kind> <repo> <vloop> <exit
code>` prints at the starting point, byte for byte, except:

- **The session table** of each kind (and, for `gate-review`, its table
  with the `round` column), from the table's header row (`| record | …`,
  or `| round | record | …` for `gate-review`) through the blank line
  after it, is replaced (the `##` heading and the exit line before it
  stay), in its place, by one block per brief, each
  exactly this (every line shown; `⏎` marks an empty line):

  ````
  Sessions (`vloop metrics <brief> --by session`):
  ⏎
  ```
  <that command's stdout, its final newline not doubled>
  ```
  ⏎
  Session records: <n>, schema-valid <k> (<schema> <c>, …).
  - <folder>/sessions/<file>: <schema> INVALID      (one per rejected record)
  ⏎                                                 (these three parts only
  Permission denials:                                when a row has denials)
  ⏎
  - <run>/<#> <phase> <tool_name>: <text>          (one per denial)
  ⏎
  ````

  Each block ends with one empty line, which also separates one brief's
  block from the next. `<n>` counts the files `sessions/NNN-<phase>.json`
  in the brief's run folders, and `<k>` those that `vloop schema validate
  <schema> <file>` accepts, `<schema>` being the record's own `schema`
  field; a record with none is listed as `- <folder>/sessions/<file>: no
  schema`. The parenthesis counts the records per `schema` value, in name
  order (`session/v3 2, session/v4 3`), records without one last as `no
  schema <c>`, so the version each record was written under stays
  visible, as the Python's per-version validation showed it. The kit reads the record files itself for these lines: folders
  in name order, records in sequence order, and each denial's `<text>`
  by P1's rule, so they do not depend on `vloop metrics` succeeding (the
  fixtures name their folders in the order of their timestamps, so name
  order is the rows' order). With no record at all the line is `Session
  records: 0.`, with no parenthesis. A session file that is not JSON
  counts in `<n>`, not in `<k>` nor the parenthesis, and is listed as
  `- <folder>/sessions/<file>: unreadable`.
  `<brief>` is the run id (`vloop metrics <run id> --by session`, as
  written in the heading and run), for each `<run id>` whose
  `<repo>/.vloop/state/runs/<run id>/` directory exists and whose
  `docs/briefs/<run id>.loop-brief.md` exists, in name order: one block
  per brief, its counts and lines that brief's. `<folder>` in the record
  lines is the repo-relative folder path
  (`.vloop/state/runs/<run id>/<folder name>`). The command runs in `<repo>` with the `<vloop>` given.
  When it exits non-zero, the one line `vloop metrics exited <code>: <its
  stderr, trimmed, newlines as spaces>` stands inside the fence in place
  of its stdout, and the rest of the block follows as usual. In a denial
  line, each newline in `<text>` prints as a space.
  When no run id qualifies, the session table is replaced by the one
  line `Sessions: no brief with runs.` and a blank line.
- Everything after the session table (verdicts, task outcomes, the kept
  verify report, the kept gate verdicts, the plan, the second-review lines
  and dissents) is as the Python prints it, `vloop schema validate`
  results included. Where the Python joins two unsorted `glob` lists (the
  `second` report's `runs/*/*/run.log` and `runs/*/run.log`, likewise
  `iterations.jsonl`), the Go program takes each list in sorted path
  order, the `*/*/` list first.

The program exits 0 when it printed the report, 1 with a message on stderr
when an argument is missing or `<kind>` is unknown.

### P3 — the kit's probes in Go

The same program, with these commands, each reproducing the starting
point's Python: the files it reads and writes, the session it starts
(argv, working directory, environment, with only the prompt replaced),
the prompts and forms (`FORMS`, `PROMPT`, `SKILL`, `MODES`, `FAIL_CALL`)
verbatim, and its stdout byte for byte. `hook forms` and `hook failure`
read `VLOOP_HOOK_PHASE` as `hook.py` does (default `review`) and write it
into the hooks file's `--phase`. Each command exits 0 when it ran; when
`copilot-argv.log` holds no `/vloop:review` session, `probe`, `hook
forms` and `hook failure` print `c5: no review session in the copilot
log` on stderr and exit 1, as the Python does. An unknown command or
`hook failure` mode, or a missing argument, exits 1 with a one-line
message on stderr, as `report` does:

| Go command | replaces |
| --- | --- |
| `probe <dir> <copilot>` | `report.py probe` |
| `summary <dir>` | `report.py summary` |
| `hook forms <dir> <copilot>` | `hook.py forms` |
| `hook failure <dir> <copilot> <mode>` | `hook.py failure` |
| `hook summary <dir>` | `hook.py summary` |
| `calls <out.jsonl>` | `bypass.sh`'s inline Python (`calls`) |

What the Python renders, the Go program renders the same way:

- `json.dumps(x)`: object keys in the input's order (sorted where the
  Python passes `sort_keys`), `", "` and `": "` as separators,
  non-ASCII as `\uXXXX` with lowercase hex, no HTML escaping of `<`, `>`
  and `&`; a character outside the BMP as a surrogate pair
  (`\ud83d\ude00`); control characters as `\n`, `\t`, `\b`, `\f`, `\r`,
  others as `\u001f`-style lowercase escapes; integers as integers,
  floats as Python's `repr` (`1.0`, `1e-05`, `1e+16`).
- `json.dump(…, indent=2)` (the hooks file `hook.py` writes): two-space
  indentation, `","` at line ends and `": "`, as Python writes it.
- A value formatted with `str()` or an f-string: Python's `repr` for a
  list (`['a', 'b']`); `None`, `True` and `False` as Python writes them.
- A slice (`[:90]`) counts characters, not bytes.

The fixtures hold, between them, a float, an integral float, non-ASCII
text, an emoji, a string with a newline and a tab, `<&>` in a string,
and two distinct event types in one probe file,
so the goldens test each rule.

### P4 — the scripts and the README

- `run.sh`, `second.sh`, `work.sh`, `verify.sh`, `gatereview.sh`,
  `plan.sh`, `hook.sh` and `bypass.sh` call `go -C "$top" run
  ./tools/copilot-acceptance/kit …` where they called `python3`, with the
  same arguments (absolute paths, as today), and write the same files.
  `top` is the repository root, computed in every one of them as `run.sh`
  computes it (`hook.sh` and `bypass.sh` included); the scripts make the
  call from the scratch repository, outside the module, hence `-C`
  (`go.mod` pins Go 1.27.1). Every script makes its directory argument
  absolute right after reading it, in a command substitution that leaves
  the script's working directory unchanged: `dir=$(mkdir -p "$dir" && cd
  "$dir" && pwd -P) || die …` (the same for `work` in `bypass.sh`), so
  the Go program gets absolute paths whatever the caller passed. The line
  goes after `die()` is defined (moving `die()` up where it comes later);
  `bypass.sh`, which defines none, uses `|| exit 1` there. No script tests for `python3` or runs it; their
  header comments say what each needs, without `python3`.
- `bypass.sh`'s header names `perl` among what it needs.
- `tools/copilot-acceptance/report.py` and `hook.py` are deleted;
  `fail-open.sh`, which runs no Python, is unchanged.
- `tools/copilot-acceptance/README.md` says the kit needs `go`, `copilot`
  and `gh` on PATH (and `claude` for the scripts that run Claude
  sessions, `perl` for `bypass.sh`, whose timestamps it takes with
  `perl -MTime::HiRes`, unchanged); it names no Python, and where it named `report.py …` or
  `hook.py …` it names the Go command; it says the session tables are
  `vloop metrics --by session`.

### Decided here, because the cited documents leave it underdetermined

- `#` is three digits (`001`), as the file names.
- `credits` prints `-` for a record without the field, `0.00` for a record
  whose field is 0.
- A `second-review` session's row has phase `second-review`.
- The kit's goldens and fixtures live under
  tools/copilot-acceptance/kit/testdata/. A golden is the Python's
  stdout (or a file the Python wrote) on a fixture, unedited, except that
  the fixture's directory as the Python saw it (resolved: macOS temp
  directories resolve through `/private`) is replaced by `@DIR@`; the Go
  output gets the same substitution before it is compared.
- **How a report is compared with its golden:** the test builds the
  expected output from the golden by replacing its first table (from the
  table's header row through the blank line after it, as P2 says) with the
  Sessions block(s) it composes: the fenced stdout of `vloop metrics
  <run id> --by session` run on the same fixture, and the record and
  denial lines it writes literally. It then compares the whole Go output
  with that, byte for byte. The test copies each report fixture into a
  temporary directory and makes it a git repository (`git init`, then
  the commits `vloop metrics` reads, all with `git -C`). It fails when
  its own `vloop metrics` call exits non-zero or prints fewer rows than
  it pins for that fixture (at least 2), on every report fixture but the
  two that test those cases (metrics exits 1; no run folder), so a block
  holding an error cannot pass as the golden.
- **Fixture paths:** fixture files hold `@DIR@` where the test's
  temporary directory goes (in `copilot-argv.log`, the session's `cwd`
  and `env COPILOT_HOME`, for instance). The test, and the golden
  generator, copy them into a temporary directory resolved with
  symlinks followed (`pwd -P`; macOS temp directories resolve through
  `/private`), substitute that resolved path, and pass it as `<dir>` to
  the program. In outputs only that resolved form is replaced back by
  `@DIR@`. `VLOOP_HOOK_PHASE` is unset in both runs.
- **The golden generator** is
  tools/copilot-acceptance/kit/testdata/gen.sh `<python dir>
  <vloop> <out dir>`, a shell script that runs the Python in `<python
  dir>` on the fixtures and writes the goldens; no `go test` runs it.
  `<python dir>` holds `report.py`, `hook.py` and `bypass.sh`; gen.sh
  extracts the lines strictly between `python3 - "$d/out.jsonl" <<'PY'`
  and the closing `PY` of `bypass.sh` into `<python dir>/calls.py` and
  runs it as `python3 calls.py <out.jsonl>`. Task 2 writes gen.sh and its
  gate runs it.
- The `second` fixture has one run folder (one `run.log`, one
  `iterations.jsonl`), so the Python's unsorted `glob` order is the only
  order.
- Tests start no real `copilot`, `claude` or network call. The stub
  `copilot` is a committed POSIX sh script,
  tools/copilot-acceptance/kit/testdata/copilot: it appends to
  `$STUB_LOG` (a file under @DIR@ the caller sets) its argv (one `arg <a>`
  line each), `cwd <pwd -P>` and its environment (`env | LC_ALL=C sort`,
  each line prefixed `env `), then prints the JSON Lines file
  `$STUB_OUT` names. gen.sh passes it to the Python as `<copilot>` and the
  Go tests pass it to the program, with the same `STUB_LOG` and
  `STUB_OUT`; its log, with @DIR@ substituted and the lines whose
  variables differ between two runs of the same machine (`PWD`, `OLDPWD`,
  `SHLVL`, `_`) dropped, is among the goldens. The probes start the stub
  with the environment logged in copilot-argv.log (plus the shims'
  PATH), not the caller's, so each probe fixture's review block carries
  `env STUB_LOG=@DIR@/<file>`, `env STUB_OUT=@DIR@/<file>` and `env
  PATH=/usr/bin:/bin`, and a `cwd` that names a directory the fixture
  holds (`cwd @DIR@`). The stub is committed executable (mode 100755):
  both programs exec it directly.

### Violations the review must rule on

- A golden produced by the Go program rather than by the Python, or
  edited by hand to match the Go output.
- A row set that differs from the sessions metrics sums (a session left
  out, a brief-review session added).
- A report whose non-session sections differ from the Python's.
- A probe prompt or form list that is not the Python's verbatim.

## Worked example

Each line is a named test, on temporary repositories built by the test
(`git -C`), or a check on this repository's real records.

```
# P1, real data (this repository, B36)
vloop metrics B20261103-0528-b36-i96-i97-plugin-per-repo-and-agent --by session
  → 13 rows, folder 20261103-062225, #001 to #013; compared field by
    field (split on whitespace), the first row reads:
    20261103-062225  001  plan  -  claude  opus  9m42s  $2.89  reported  -  44  3  -
  and the last: 20261103-062225  013  verify  -  claude  sonnet  1m38s  $0.30  reported  -  18  2  -
… --json → by_session has 13 rows; their cost_usd sum to cost.total_usd
  (5.6702516…) within 1e-9; their permission_denials sum to the report's permission_denials (8)
# P1, fixtures
a brief with a plan (seq 1), a work (seq 2, iteration 1, task T1) and a review (seq 3) record in one folder, and a work record in a second, later folder
  → 4 rows: folder 1's #001 plan task -, #002 work T1, #003 review T1, then folder 2's
a Copilot work record with credits 1.5, cost_source "estimated"  → credits 1.50, source estimated
a session/v1 record (no agent, no cost_source, no credits)        → agent claude, source -, credits -; JSON cost_source null, credits null
a legacy shell-loop record (total_cost_usd, num_turns 7)          → its row, turns 7
a record with neither turns nor num_turns                         → turns -; JSON turns null
a record with permission_denials [{"tool_name":"bash","tool_use_id":"x","message":"m"},{"tool_name":"Bash","command":"c"}]
  → denied 2; denials [{"tool_name":"bash","text":"m"},{"tool_name":"Bash","text":"c"}]
a record with permission_denials [{"tool_name":"Write","file_path":"f"},{"tool_input":{"command":"x"}}]
  → denials [{"tool_name":"Write","text":"f"},{"tool_name":"","text":"x"}]
a record with is_error true → error yes
vloop metrics --by phase                        → exit 2, unknown --by "phase": want task or session
vloop metrics --by session --workspace w.toml   → exit 2, --by session reports one repository's briefs; it takes no --workspace
vloop metrics --interventions --by session      → exit 2, --by takes kind or phase with --interventions
a metrics/v4 snapshot without by_session        → vloop schema validate metrics/v4 accepts it
# P2 and P3, fixtures under tools/copilot-acceptance/kit/testdata/
each report kind on a fixture repository → equal to its golden: the Python's output with the session table replaced as P2 says
probe, summary, hook forms, hook failure (each mode), hook summary, calls on fixture dirs and a stub copilot
  → stdout equal to the Python's golden; the stub saw the argv, working directory and environment the Python passed;
    the files written (probe.argv, probe.rc, the hook-<x>.argv and .rc files, hooks/vloop-fence.json) equal the Python's goldens, after the @DIR@ substitution
report work on a repository where vloop metrics exits 1 → inside the fence "vloop metrics exited 1: …", the record lines and the rest follow
report run on a repository with no run folder → "Sessions: no brief with runs." in the table's place, the rest as the golden
probe on a copilot-argv.log without a review session → exit 1, "c5: no review session in the copilot log" on stderr
hook forms with VLOOP_HOOK_PHASE=second-review → hooks/vloop-fence.json's bash and powershell commands hold --phase second-review
report with no kind → exit 1, a message on stderr
```

The failures it must catch: a session missing from the rows or counted
twice; a cost rounded in JSON; a brief-review session listed; a probe
prompt reworded; a `json.dumps` rendering with Go's separators or key
order; a script that still needs `python3`.

## Acceptance, at close (the operator's)

With `python3` made unavailable to the kit (a directory first on PATH
whose `python3` exits 127):

1. `tools/copilot-acceptance/work.sh <copilot model>`: the run completes or
   halts as the model does; b24-work-summary.md holds the per-session
   table, the record checks and the task outcomes.
2. `tools/copilot-acceptance/run.sh <copilot model>`: c5-summary.md
   holds the run report and the deny probe's table.
3. `vloop metrics --by session` on a brief whose sessions ran on Copilot
   (the B34 acceptance clone, ~/source/personal/url-shortener-tests/b34-backstop): credits
   and `estimated` shown per Copilot session.
4. With `tools/copilot-acceptance/__pycache__/` deleted first (left by
   earlier runs), `git status --porcelain --ignored` in this repository
   before and after the scripts ran: the same output (no file the kit wrote,
   ignored ones such as `__pycache__/` included).

Results go in `docs/design-notes/copilot-cli-experiments.md`; a behaviour
that contradicts this brief is a finding.

## Out of scope

- Brief-review sessions (`.vloop/state/brief-reviews/`) in metrics.
- A per-session view in `vloop metrics export` or in `--workspace`'s
  text table.
- Changing what the probes ask, test or conclude; new probes.
- The design notes' past results written from `report.py`.
- `tools/agent-eval` and `tools/eval-preflight`.
- #129 and #130 (the fence).

## Constraints

- Gates run on macOS's BSD tools: no empty alternatives in `grep -E`, no
  `sed -i` without a suffix, no `\+` or `\|` in basic regexes.
- Sessions never move refs.
- Tests build their fixtures in temporary directories and address them
  with `git -C`; no test runs a real agent CLI, touches `~/.claude` or
  `~/.copilot`, or needs the network.
- The goldens are produced with `python3` from the starting point's
  `report.py`, `hook.py` and the `calls` heredoc of `bypass.sh`. Task 2's
  gate takes the three files from the base (`git show
  <base>:tools/copilot-acceptance/report.py`, and `hook.py` and
  `bypass.sh`)
  into a temporary directory, regenerates the goldens there and diffs
  them with the committed ones, so it still runs after task 5 deletes the
  files. No `go test` needs `python3`.
- The kit's Go code builds on Linux and Windows (`GOOS=linux`,
  `GOOS=windows`) with the rest of the module; tests that need a shell
  stub skip on Windows.
- The existing tests pass unchanged; the new refusals get tests of their
  own.
- No new module dependency.
- Repo-relative paths everywhere.

## Shape

5 to 6 tasks; every configured check green after each.

1. **P1**, gated on its worked-example lines and the real-data line.
2. **The goldens:** fixtures and the Python's output on them, committed
   under tools/copilot-acceptance/kit/testdata/.
3. **P2**, gated on the report goldens.
4. **P3**, gated on the probe goldens and the stub's records.
5. **P4**, gated on one script's `go -C "$top" run …` call made from a
   directory outside the module, one script started by a relative path, and a check that no script
   (`tools/copilot-acceptance/*.sh`) and not the README mentions
   `python3`, `report.py` or `hook.py`, and that `report.py` and
   `hook.py` are gone.
6. **Close:** every gate at once, and the worked example line for line.

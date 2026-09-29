---
name: B20260929-2148-vloop-state-tasks-status.loop-brief
description: Give vloop its file contracts as JSON Schemas, and the commands that read and amend a plan — task list/show/validate/verify/gate/reset/note/drop/set and status
kind: brief
status: ready
created: 2026-09-29
seeds: A `.loop/run.sh` plan + run that builds slice B2 in docs/design-notes/vloop-roadmap.md
depends-on: [B20260929-1804-vloop-skeleton-config-briefs.loop-brief]
---
# Brief — vloop B2: schemas, tasks and status

- **Status:** ready to plan
- **Starting point:** extends `main` at the commit that adds this brief. The
  planner pins that SHA as the base every gate compares against.
- **Produced by:** operator decision, 2026-09-29, from the B2 row of
  `docs/design-notes/vloop-roadmap.md`.

---

## What it is

The second slice of `vloop`. At the end of this run the files the loop runs on
— the plan, a work session's proposal, a review's verdict, and the per-session
and per-iteration telemetry — each have a versioned JSON Schema embedded in the
binary, and `vloop schema` lists, prints and validates against them. The
operator can read a plan with `vloop status` and `vloop task list|show`, check
it with `vloop task validate`, run one task's gate with `vloop task gate`, and
amend it — replace a gate with a recorded reason, reset, note, drop, or set a
task's model, effort, area and kind — without hand-editing JSON. Nothing in
this slice runs a session: the plan it reads is one a fixture wrote, until the
driver (B5) writes real ones.

## Why this shape, and what was rejected

Decided with the operator on 2026-09-29, in the roadmap and in the metrics
discussion it records. **Do not re-litigate these.**

- **The schemas directory is the contract.** Go validates against the embedded schemas;
  skills (B4) will cite them. Every JSON document vloop defines carries a
  `"schema": "<name>/v<n>"` field, so an old file is recognisable after the
  format moves.
  *Rejected:* Go structs as the only definition — the skills cannot read them.
- **The plan lives at `.vloop/state/state.json`**, never `.loop/`. vloop does not
  read the shell loop's state in this slice.
- **Per-task model and effort override the per-kind config**, separately for
  the work and the review session of that task. The task is the most specific
  source, so it wins over the environment.
  *Rejected:* one `model` per task for both sessions — it would make an
  independent reviewer impossible to configure per task.
- **Tasks carry `area` and `kind`.** `area` comes from the repo's `areas` list
  in config; `kind` from a fixed list shared by every repo: `feature`, `fix`,
  `refactor`, `test`, `docs`, `chore`.
- **A gate can be the defect.** The proposal schema carries an optional
  `gate_dispute`; replacing a gate records who replaced it, why and what it was,
  in the task's `gate_history`, so B3 can reclassify the failures it caused.
  This slice stores the facts; it computes nothing from them.
- **Metrics-snapshot and defect-record schemas are B3's**, not this slice's —
  their shape follows from the metrics B3 computes, and defining them first
  would freeze a guess. (The roadmap's B2 row listed them; this brief moves them,
  and the roadmap is amended to match.)
- **Plan checks here are structural.** The shell driver's plan-time gate-shape
  rejections belong to the driver port (B5).
- **Libraries:** B1's, plus one pure-Go JSON Schema validator
  (`github.com/santhosh-tekuri/jsonschema/v6`). Nothing that needs cgo.

## Binding references

- `docs/design-notes/vloop-roadmap.md` — the series, the decisions every brief
  inherits, and the `## Metrics and defects` section: `area`, `kind`,
  `gate_dispute`, and why gate history is recorded.
- `docs/briefs/B20260929-1804-vloop-skeleton-config-briefs.loop-brief.md` — the
  conventions this slice extends: global flags, exit codes 0/1/2, the repo
  root, the `config` key table and its output formats, `--json` rules, the
  README test.
- `.loop/amend.sh` — the plan operations and checks `task` ports: its commands
  and every rule of its `check`.
- `.loop/render-plan.sh` — what a rendered plan shows; `status --markdown`
  carries the same content.
- `.claude/skills/loop-work/SKILL.md` — the proposal fields (section 4) the
  proposal schema formalises.
- `.claude/skills/loop-review/SKILL.md` — the verdict fields the verdict schema
  formalises.

## Behaviour contract

Everything B1 pinned still holds: global flags, `--json` (one document on stdout,
on exit 0 and 1), exit codes `0` ok / `1` problems or failure / `2` usage,
errors as one stderr line starting `vloop: `, paths repo-relative with `/`.

### Schemas — `vloop schema list | show <name> | validate <name> <file>`

Five schemas, embedded, each JSON Schema draft 2020-12:

| Name | Describes |
| --- | --- |
| `state/v1` | the plan: `.vloop/state/state.json` |
| `proposal/v1` | a work session's report |
| `verdict/v1` | a review session's verdict |
| `session/v1` | one session's telemetry record |
| `iteration/v1` | one iteration's record |

- `schema list` prints the names, one per line, sorted. `--json`: an array.
- `schema show <name>` prints the schema document.
- `schema validate <name> <file>` prints `<file>: ok` and exits 0, or one line
  per violation as `<file>: <json pointer>: <message>` and exits 1. The message
  text is the validator's; the pointer is exact. `--json`:
  `{"ok":bool,"errors":[{"pointer":…,"message":…}]}`.
- An unknown name is `vloop: unknown schema "<name>"`, exit 2. A file that is not
  JSON is one violation at pointer `""`.
- Every document's `schema` field must equal the schema's name.

**Keys are `snake_case`**, as the shell loop's are. The fields, required unless
marked optional:

- `state/v1`: `schema`, `run_id`, `brief`, `base` (the commit gates compare
  against), `branch`, `status` (`planning`, `running`, `complete`, `blocked`,
  `stalled`, `halted`), `iteration`, `created`, `updated` (RFC 3339 UTC),
  `shell` (the shell every `verify` in the plan is written for: `sh`, `bash`,
  `pwsh`, `powershell` or `cmd`), `tasks`.
- A task: `id` (`T<n>`), `title`, `goal`, `kind`, `area` (optional),
  `files`, `references` (`[{path, why}]`), `depends_on`, `acceptance`
  (non-empty), `verify` (non-empty), `status` (`pending`, `done`, `blocked`),
  `attempts`, `notes`, `model` and `effort` (optional objects with optional
  `work` and `review` keys), `gate_history` (optional
  `[{verify, replaced_at, reason, by}]`, `by` being `operator` or `planner`).
- `proposal/v1`: `schema`, `task`, `outcome` (`done`, `blocked`), `summary`,
  `files`, `verified`, `notes`, and optional `gate_dispute` (`{reason, evidence}`).
- `verdict/v1`: `schema`, `task`, `verdict` (`PASS`, `FAIL`), `criteria`
  (`[{criterion, met, evidence}]`), `findings` (`[{summary, kind}]`, `kind` one of
  `bug`, `spec-gap`, `gate-gap`), `notes`.
- `session/v1`: `schema`, `run_id`, `iteration`, `phase` (`plan`, `work`,
  `review`), `task` (absent for `plan`), `model` and `effort` as configured
  (`effort` may be null), `models_used` (per model: input, output,
  cache-read and cache-creation tokens, cost), `started`, `duration_ms`,
  `cost_usd`, `turns`, `is_error`, `permission_denials`.
- `iteration/v1`: `schema`, `run_id`, `iteration`, `task`, `attempt`,
  `outcome` (`done`, `gate_failed`, `rejected`, `blocked`, `session_error`),
  `gate` (`{exit, duration_ms}`, null when no gate ran), `started`, `ended`.

Unknown keys are allowed in every schema, so a newer writer does not break an
older reader.

### The plan file

Commands read `.vloop/state/state.json` under the repo root. If it is absent:
`vloop: no plan: .vloop/state/state.json`, exit 1. If it fails its schema, read
commands still work where they can, and every write command refuses with
`vloop: plan is not valid — run vloop task validate`, exit 1, and leaves the file
untouched.

Writes are atomic (write a temporary file beside it, then rename). The file is
written as 2-space-indented JSON with keys in schema order and a trailing
newline; loading and saving an unchanged plan leaves it byte-for-byte identical.
Every write sets `updated`.

### `vloop status [--json | --markdown]`

```
<run_id> — <status> · <done>/<total> done · <blocked> blocked · iteration <n>
  <id>  <status>  <area>/<kind>  <title>[  (<attempts> attempt[s])]
```

Columns are left-aligned on the widest value in each; the attempts suffix
appears only when attempts is above 0; a task without `area` shows `-/<kind>`.
`--json`: `{"run_id","status","iteration","done","total","blocked","tasks":[{"id","status","area","kind","title","attempts"}]}`.
`--markdown` prints the content `.loop/render-plan.sh` renders — header, progress
checklist, and per task its status, dependencies, files, goal, acceptance and
notes — to stdout. Nothing in this slice writes a markdown file.

### `vloop task …`

| Command | Does |
| --- | --- |
| `task list` | the `status` task lines alone; `--json` the same array |
| `task show <id>` | the whole task, plus the model and effort resolved for its work and review sessions, each with its source |
| `task validate` | checks the plan (below); exit 0 or 1 |
| `task gate <id>` | runs the task's `verify` |
| `task verify <id> '<cmd>' --reason '<text>'` | replaces the gate |
| `task reset <id>` | status `pending`, attempts 0 |
| `task note <id> '<text>'` | replaces `notes` |
| `task drop <id>` | removes the task |
| `task set <id> <field> <value>` | sets `area`, `kind`, `model.work`, `model.review`, `effort.work` or `effort.review`; `''` clears |

- **Resolution.** For a session kind `k` of a task: the task's `model.k`, else
  `VLOOP_MODEL_<K>`, else `model.k` in the config file, else the default — the
  same for effort. `task show` prints, e.g.,
  `model   work opus (task) · review sonnet (default)` and
  `effort  work - (default) · review high (file)`. `--json` carries
  `"resolved":{"work":{"model":…,"model_source":…,"effort":…,"effort_source":…},"review":{…}}`.
- **`task gate <id>`** runs `verify` from the repo root in the plan's `shell`,
  streaming its output, then prints `gate <id>: pass (<duration>)` or
  `gate <id>: fail (exit <n>, <duration>)`; vloop exits 0 on pass, 1 on fail.
  `--json`: `{"task","passed","exit","duration_ms"}`, with the gate's own output
  on stderr. It writes nothing. Invocation per shell: `sh`/`bash` as
  `<shell> -c <cmd>`; `pwsh`/`powershell` as
  `<shell> -NoProfile -NonInteractive -Command <cmd>`; `cmd` as `cmd /C <cmd>`.
  The shell not on `PATH`:
  `vloop: this plan's gates are <shell> commands and <shell> is not on PATH`,
  exit 1.
- **`task verify`** appends the old command to `gate_history` with
  `replaced_at`, `reason` and `by: operator`, then sets the new one. `--reason`
  is required (exit 2 without it). Setting the command it already has is exit 1,
  `vloop: T<n> already has that verify command`, and nothing is written.
- **`task drop`** refuses if another task depends on it:
  `vloop: T3 depends on T2`, exit 1.
- **`task set`** validates like `config set`: a `kind` outside the list, an
  `effort` outside `low|medium|high|xhigh|max`, an empty model, or an `area` not
  in `areas` (when `areas` is set) is exit 2 with B1's
  `vloop: invalid value "<v>" for <field>: want …` message.
- An unknown task id is `vloop: no task <id>`, exit 1. Every write validates the
  result first and refuses to write a plan that would fail `task validate`.

### New config keys: `shell` and `areas`

Two keys after the effort keys, in this order.

**`shell`** — the shell gates are written for and run in: `sh`, `bash`, `pwsh`,
`powershell` or `cmd`. Default: `sh` on darwin and linux, `pwsh` on windows;
`config list` shows the default as `shell=sh (default)` on a Mac. The planner
(B4/B5) writes the resolved value into the plan's `shell` field and writes every
`verify` in that shell's syntax; from then on the **plan's** `shell` governs
its gates, not the config — a plan written for `sh` stays an `sh` plan on
Windows, and says so when `sh` is missing. `VLOOP_SHELL` overrides as usual.

**`areas`** — a list, default empty. `config set
areas cli,brief,config` stores a TOML array; `config get areas` and `config
list` print it comma-joined (`areas=cli,brief,config (file)`); `--json` gives
an array. Each entry is lower-case letters, digits and hyphens. When `areas` is
empty, `area` is optional and unchecked; when set, every task needs one of them.

### `vloop task validate`

Ports every fatal rule of `.loop/amend.sh check`, plus the new fields. One line
per finding, `✗` for problems and `!` for warnings, then `plan ok` or
`<n> problem(s)`. Exit 1 on any problem. Problems:

- a schema violation: `✗ schema: <pointer>: <message>`
- `✗ duplicate task id: <id>`
- `✗ <id> depends on <dep>, which does not exist`
- `✗ depends_on cycle: <a> -> <b> -> <a>`, starting from the lowest id in the cycle
- `✗ <id>: reference does not resolve: <path>`; `✗ <id>: reference has no reason: <path>`
- `✗ <id>: area "<a>" is not in areas (<list>)`; `✗ <id>: no area — areas is set`

Warnings (exit 0): `! <id>: no gate_history reason` for a history entry with an
empty reason.

### F1 — one rule for where a reported cycle starts

**Defect.** B1's `brief check` reports a `depends-on` cycle starting from the
checked brief, `brief list` from the first brief in its order, so the same
cycle reads differently in the two commands (B1 run record, "still open").

**Required.** Both report a cycle starting from its lexicographically smallest
brief name, as `task validate` does with ids. The message text is otherwise
unchanged.

**Gate.** A fixture cycle of two briefs, where the checked one is not the
smallest, gives the same message from `brief check` and `brief list`. It fails
against B1's code.

### Decided here, because the cited documents leave it underdetermined

- `status` values for a plan beyond the shell loop's (`running`, `complete`,
  `blocked`, `stalled`) are `planning` and `halted`, so every exit the B5 driver
  can take has a state.
- Gates are shell strings (the design session's Windows decision), but not
  necessarily POSIX: the operator's direction is "any shell from the underlying
  OS" — `pwsh` on Windows, or on Linux where a project needs it. Hence `shell` in
  config and in the plan. `doctor` (B4) reports a plan whose shell is missing.
- The shell loop's `task verify` equivalent kept no history; vloop's
  `task verify` requires a reason because B3 needs it.

### Violations the review must rule on

- A schema whose only definition is a Go struct, or a Go struct that accepts what
  its schema rejects (or the reverse). The schema file is the authority.
- A write command that writes anything but `.vloop/state/state.json`, or writes
  a plan that fails `task validate`.
- vloop or its tests reading or writing this repo's `.loop/`, `.claude/`, or the
  home directory.
- Metrics, defect records, `brief close`, a run lock, a driver, or any `claude`
  invocation appearing in this run.
- Commands the README does not cover: B1's README test must keep passing with
  the new commands in it, not be loosened to let them through.

## Worked example

In a scratch git repository with the binary on `PATH`, the config file
holding `effort.review = "high"` and `areas = ["cli", "config", "docs"]`, and a
fixture `.vloop/state/state.json` whose `run_id` is `B20260101-0900-a`, status
`running`, iteration 2, `shell` `sh`, and three tasks: T1 `done`, area `cli`, kind `feature`,
title `Skeleton`, verify `test -f go.mod`; T2 `pending`, 1 attempt, area
`config`, kind `feature`, title `Config`, depends on T1, verify
`test -f config.txt`, `model: {work: opus}`; T3 `pending`, area `docs`, kind
`docs`, title `README`, depends on T2, verify `test -f README.md`. Durations are
threaded from what the command printed, never hardcoded.

```
$ vloop status
B20260101-0900-a — running · 1/3 done · 0 blocked · iteration 2
  T1  done     cli/feature     Skeleton
  T2  pending  config/feature  Config  (1 attempt)
  T3  pending  docs/docs       README                          exit 0

$ vloop task show T2          (among its lines)
model   work opus (task) · review sonnet (default)
effort  work - (default) · review high (file)                  exit 0

$ vloop task validate
plan ok                                                        exit 0
$ vloop task gate T2
gate T2: fail (exit 1, <duration>)                             exit 1
$ touch config.txt && vloop task gate T2
gate T2: pass (<duration>)                                     exit 0

$ vloop task verify T2 'test -f cfg.toml' --reason 'the file is cfg.toml'
                                                               exit 0
$ vloop task show T2 --json | jq -c '.gate_history[0] | {verify, reason, by}'
{"verify":"test -f config.txt","reason":"the file is cfg.toml","by":"operator"}
$ vloop task verify T2 'test -f other.toml'
vloop: required flag(s) "reason" not set                       exit 2
$ vloop task drop T2
vloop: T3 depends on T2                                        exit 1
$ vloop task set T3 area ops
vloop: invalid value "ops" for area: want one of cli, config, docs
                                                               exit 2
$ vloop task reset T2 && vloop task list | grep T2
  T2  pending  config/feature  Config                          exit 0

$ vloop schema list
iteration/v1
proposal/v1
session/v1
state/v1
verdict/v1                                                     exit 0
$ vloop schema validate state/v1 .vloop/state/state.json
.vloop/state/state.json: ok                                    exit 0
```

The failures, each planted in a copy of the fixture:

```
T3's id changed to T2
  -> ✗ duplicate task id: T2                                   exit 1
T2 made to depend on T3
  -> ✗ depends_on cycle: T2 -> T3 -> T2                         exit 1
T3's area changed to ops in the file
  -> ✗ T3: area "ops" is not in areas (cli, config, docs)       exit 1
T2's effort.work set to turbo in the file
  -> ✗ schema: /tasks/1/effort/work: <message>                  exit 1
     and `vloop task note T2 x` refuses:
     vloop: plan is not valid — run vloop task validate        exit 1
.vloop/state/state.json deleted
  -> vloop: no plan: .vloop/state/state.json                    exit 1
the plan's shell changed to a name absent from PATH (a test double: `pwsh` on a
machine without it, with PATH stripped of it)
  -> vloop task gate T2:
     vloop: this plan's gates are pwsh commands and pwsh is not on PATH
                                                               exit 1
```

## Out of scope

The roadmap's later rows, not smaller versions of them:

- Metrics, line classification, the metrics-snapshot and defect-record schemas,
  `vloop defect …`, `vloop metrics …`, `brief close`, release detection (B3).
- Reclassifying gate failures from `gate_history` or `gate_dispute` — this slice
  records, B3 computes.
- `init`, `upgrade`, `doctor`, plugin extraction, skill prose (B4).
- `run`, the run lock and `runs unlock`, writing `plan.md`, writing telemetry,
  the driver's plan-time gate-shape rejections, invoking `claude` (B5).

Also not in this run:

- Reading or converting the shell loop's `.loop/state/` files.
- `amend.sh check`'s advisory on files owned by another task's gate.
- Adding or re-scoping tasks (`task add`): structural changes go through the
  brief and a re-plan, as `.loop/amend.sh` already argues.
- A schema registry, remote `$ref`s, or publishing the schemas anywhere.

## Constraints

- Go as pinned in `go.mod`. Third-party modules: B1's, plus
  `github.com/santhosh-tekuri/jsonschema/v6`; anything else needs its reason in
  the task notes. No cgo. vloop never uses the network.
- Must build for `darwin`, `linux` and `windows`. Nothing may assume `/` as the
  path separator on disk or a case-sensitive filesystem; only `task gate` runs
  a shell, and it runs the plan's. Tests of `task gate` use `sh` where it
  exists and skip the other shells when they are absent, saying so.
- **Gates follow the diff surface**: each task gates on the packages it touches,
  plus `go vet ./...`, `gofmt -l .` printing nothing, **`go mod tidy -diff`
  printing nothing** (B1's gap), and the linux, windows and host builds. Only
  the close runs `go test ./...`.
- A gate runs in under a minute. Tests use temporary directories and never touch
  this repo's `.loop/`, `.claude/` or the home directory.
- **No task may weaken a check to pass**, including B1's tests. A check that is
  wrong is a blocked task with a note. One change to B1's tests is required, not
  weakening: its end-to-end test pins `config list`'s exact output, which now
  gains `shell` and `areas` lines. Update that expectation to the new key table
  and nothing else in that test.
- **New checks are proven by fixtures**: every `task validate` rule and every
  schema has a fixture that passes and one where it fires.
- Repo-relative paths everywhere. No absolute paths in any file or commit message.

## Shape

8 to 10 tasks, each independently verifiable by a single command. Order is
load-bearing:

1. **The five schemas, embedded, and `vloop schema list|show|validate`**, with a
   valid and an invalid fixture per schema. Every later task's plan files are
   checked against these.
2. **Loading and saving the plan** — atomic, byte-for-byte stable — and
   **`vloop status`** in text, `--json` and `--markdown`.
3. **The `shell` and `areas` config keys** (with B1's `config list`
   expectation updated) and **`task list|show`**, with model and effort
   resolution and its sources.
4. **`task validate`**: every rule, each with a passing and a firing fixture.
5. **`task reset|note|drop|set`**, each refusing an invalid result.
6. **`task verify`** with `gate_history`, and **`task gate`**.
7. **F1**: one cycle-start rule for `brief check` and `brief list`, gated on the
   fixture that fails against B1's code.
8. **README**: the new commands, the `areas` key and the plan file, in B1's
   style and under its line cap; B1's README test passes unmodified.
9. **Close**: an end-to-end test that builds the binary and plays the worked
   example line for line in a scratch repo, plus `go test ./...`, `go vet ./...`,
   `gofmt -l .`, `go mod tidy -diff` and the three builds.

The worked example needs its own gate (task 9): the unit tests can all pass
while flags, exit codes or stdout vs stderr are wired wrong.

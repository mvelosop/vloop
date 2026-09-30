# vloop

vloop is a command-line tool for running autonomous Claude loops: a plan is cut
from a written **brief**, and then each task in the plan is worked by Claude,
checked, and reviewed without a human in between. This version (0.x) covers the
part you need before any loop runs: repo-local configuration and the format,
checking and ordering of briefs, and reading and checking the plan a loop runs
from. Running a loop is not part of it yet.

Build it with `go build -o vloop ./cmd/vloop` and put the binary on your `PATH`.

## The loop in one paragraph

A loop starts with a **plan**: a session reads the brief and splits it into
tasks, each with acceptance criteria and a verify command. Then, for every task
in turn, a **work** session does the task, a **gate** runs the verify command,
a **review** session judges the result independently, and the driver makes one
**commit**. Every session is a fresh Claude session that remembers nothing; the
sessions share nothing but files in the repository (the plan, a journal, the
code). Briefs are the input to this cycle, and they are what vloop handles today.

## Briefs

A brief is a Markdown file named `B<YYYYMMDD-HHMM>-<slug>.loop-brief.md` in
`docs/briefs/`. Other files in that folder are ignored. It starts with YAML
frontmatter, for example:

```
---
name: B20260929-1804-example.loop-brief   # the filename without .md
kind: brief
status: draft                             # draft | ready | consumed | abandoned
created: 2026-09-29
depends-on: []                            # briefs that must be consumed first
---
```

**Lifecycle.** A brief's `status` is `draft` while you write it, `ready` once it
is fit to plan, `consumed` after a loop has built it, and `abandoned` if you
drop it. Only `ready` briefs are checked; the others are reported as skipped.

**`depends-on`** lists other briefs, by name (filename minus `.md`), that must be
`consumed` before this one can run. A name that matches no brief, or a cycle, is
a problem.

**Sections.** A ready brief needs eight sections: What it is; Why this shape, and
what was rejected; Binding references; Behaviour contract; Worked example; Out of
scope; Constraints; Shape. `vloop brief new` writes them all with guidance.

**`## Binding references`** lists the files the work is bound by. Each item is a
backticked repo-relative path, then ` — ` and the reason it binds. A path that
does not exist, or an item with no reason, is a problem.

## Commands

| Command | What it does |
| --- | --- |
| `vloop version` | print the vloop and embedded plugin versions |
| `vloop config get <key>` | print the resolved value of a key |
| `vloop config set <key> <value>` | write a key to `.vloop/config.toml` (`''` removes it) |
| `vloop config list` | print every key with its value and where it came from |
| `vloop config path` | print the config file path, relative to the repo root |
| `vloop brief check <path>...` | check that ready briefs are fit to plan |
| `vloop brief new <slug>` | write a draft brief from the template |
| `vloop brief list` | list briefs in dependency order, ready or blocked |
| `vloop schema list` | print the names of the embedded JSON Schemas |
| `vloop schema show <name>` | print one schema document |
| `vloop schema validate <name> <file>` | validate a JSON file against a schema |
| `vloop status` | show the plan's progress |
| `vloop task list` | print one line per task |
| `vloop task show <id>` | print a task and the model and effort its sessions resolve to |
| `vloop task validate` | check the plan's structure |
| `vloop metrics [<brief>…] [--by task]` | summarise a brief's size, time, cost and defects from its runs and commits; with no brief, one row per brief; `--by task` gives one row per task |
| `vloop metrics stacks [name]` | print the built-in stack presets, or one preset's globs |
| `vloop metrics classify <path>…` | print each path's category and the layer and glob that decided it |
| `vloop defect add "<summary>"` | record a defect as `.vloop/defects/D<stamp>-<slug>.md` and print its path |
| `vloop defect list` | print the recorded defects, sorted by id |
| `vloop defect set <id> <field> <value>` | set a defect's `status`, `fixed-by`, `case`, `severity`, `origin`, `kind` or `task` |
| `vloop task reset <id>` | set a task back to pending with no attempts |
| `vloop task note <id> <text>` | replace a task's notes |
| `vloop task drop <id>` | remove a task nothing depends on |
| `vloop task set <id> <field> <value>` | set a task's `area`, `kind`, `model.work`, `model.review`, `effort.work` or `effort.review` (`''` clears it) |
| `vloop task verify <id> <command>` | replace a task's verify command; needs `--reason` |
| `vloop task gate <id>` | run a task's verify command in the plan's shell |

`vloop brief new` takes `--dry-run`: print the path and write nothing.
`vloop status` takes `--markdown`: print the plan as Markdown (not with `--json`).
`vloop task verify` takes `--reason <text>`: why the gate is being replaced. It
is required, and is recorded in the task's `gate_history`.

`vloop defect add` needs `--found-by gate|review|operator|user` and a brief:
`--brief <name>`, or `--blame <file>:<line>` to attribute the line on the default
branch (origin/HEAD's target, else `main`, else `master`) through its
`Vloop-Brief:` trailer or the commit that consumed a brief. It also takes
`--task <id>`, `--origin` (default `work`), `--kind` (default `bug`),
`--severity` (default `medium`) and `--case <path>`. `vloop defect list` takes
`--brief <name>`, and `--matrix`, which prints the origin (brief, plan, work,
env) × catcher (gate, review, operator, user) counts instead: the recorded
defects plus the ones vloop derives from the runs on every call and never
stores (each failed gate, and each finding of a rejected review). A gate failure
before an operator's `task verify` on that task counts as origin plan. With no
`--brief` it covers every brief that has runs. There is no delete, and the description cannot be edited.

These flags work on every command:

- `-C, --dir <path>`: act as if started in that directory.
- `--json`: print one machine-readable JSON document on stdout.
- `--no-color`: disable colour. A non-empty `NO_COLOR` environment variable does
  the same. Colour is only used when stdout is a terminal.
- `-q, --quiet` and `-v, --verbose`: print less or more.

## Configuration

Settings are per repository, in `.vloop/config.toml`. Nothing is read from your
home directory. Each key can be overridden by an environment variable: the key
in upper case, `.` turned into `_`, with the prefix shown in the table. The environment wins over
the file, and the file over the default.

| Key | Default | Values | Environment variable |
| --- | --- | --- | --- |
| `language` | `en` | `en`, `es` | `VLOOP_LANGUAGE` |
| `model.plan` | `opus` | any model name | `VLOOP_MODEL_PLAN` |
| `model.work` | `sonnet` | any model name | `VLOOP_MODEL_WORK` |
| `model.review` | `sonnet` | any model name | `VLOOP_MODEL_REVIEW` |
| `effort.plan` | unset | `low`, `medium`, `high`, `xhigh`, `max` | `VLOOP_EFFORT_PLAN` |
| `effort.work` | unset | same | `VLOOP_EFFORT_WORK` |
| `effort.review` | unset | same | `VLOOP_EFFORT_REVIEW` |
| `shell` | `sh` (`cmd` on Windows) | `sh`, `bash`, `pwsh`, `powershell`, `cmd` | `VLOOP_SHELL` |
| `areas` | unset | a TOML array of names | `VLOOP_AREAS` |
| `metrics.stacks` | unset | a TOML array of stack names | `VLOOP_METRICS_STACKS` |
| `metrics.code` | unset | a TOML array of glob patterns | `VLOOP_METRICS_CODE` |
| `metrics.test` | unset | a TOML array of glob patterns | `VLOOP_METRICS_TEST` |
| `metrics.docs` | unset | a TOML array of glob patterns | `VLOOP_METRICS_DOCS` |
| `metrics.excluded` | unset | a TOML array of glob patterns | `VLOOP_METRICS_EXCLUDED` |

`language` chooses the language of a brief's section headings and of the template
`vloop brief new` writes. It applies to briefs only: commands, flags, keys, JSON
and vloop's own messages are always English. The `model.*` and `effort.*` keys
choose the model and the effort for each kind of session (plan, work, review).
`shell` is the shell `vloop task gate` runs a verify command in. `areas` lists
the names a task's `area` may take; when set, `vloop task validate` reports any
other. In the file it is an array, for example `areas = ["cli", "docs"]`.
The `metrics.*` keys are lists too, set as comma-joined text and stored under
`[metrics]`: `metrics.stacks` names the language presets used to classify lines
as code, test, docs or excluded, and the four glob keys hold the repository's
own doublestar patterns, which win over the presets, for example
`metrics.code = ["internal/brief/templates/**"]`.
For example:

```
$ vloop config set language es
$ vloop config set effort.review high
$ vloop config list
```

## Files

vloop keeps its files under `.vloop/` in the repo root (the nearest parent with a
`.vloop/` folder, else the nearest with `.git`, else the current directory):

- `.vloop/config.toml`: the settings above, written by `vloop config set`.
- `.vloop/state/state.json`: the plan, the task list a loop works through. It
  follows the `state/v1` schema; `vloop status` reads it and `vloop task validate`
  checks it.

Briefs live in `docs/briefs/`. Only `vloop config set`, `vloop brief new`, `vloop defect add|set` (in
`.vloop/defects/`) and the `vloop task` commands that change a task (`reset`, `note`, `drop`, `set`,
`verify`) write anything, and each refuses a plan or result that fails
`vloop task validate`. Every path vloop prints is relative to the repo root.

## Guides

- [docs/guide/metrics.md](docs/guide/metrics.md): every `vloop metrics` number, the
  classification layers and the stack presets, and the `metrics/v1` keys.
- [docs/guide/defects.md](docs/guide/defects.md): origin and catcher, derived and
  recorded defects, `--blame`, the `.vloop/defects/` format and the matrix.

## Exit codes

- `0`: success.
- `1`: the command ran and found problems or failed (a brief with problems, a
  malformed config file, a brief that already exists).
- `2`: usage error: an unknown command, flag or config key, a missing argument,
  or an invalid value.

Errors go to stderr as one line starting with `vloop: `.

## What comes next

vloop is built as a series of briefs; what each one adds, and in what order, is in
[docs/design-notes/vloop-roadmap.md](docs/design-notes/vloop-roadmap.md).

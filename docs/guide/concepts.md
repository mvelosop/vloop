# Concepts guide

The ideas behind vloop, in the order you meet them. Every command is in
[commands.md](commands.md), every setting in [configuration.md](configuration.md).

## The loop

A loop starts with a **plan**: a session reads a brief and splits it into
tasks. Then, for each task in turn, a **work** session does the task, a **gate**
runs its verify command, a **review** session judges the result independently,
and the driver makes one **commit**. Every session is a fresh Claude session
with no memory; the sessions share only files in the repository. A task that
fails its gate or its review is tried again, and a run that cannot go on halts
and says why (see [Exit codes](#exit-codes)).

## The driver

The **driver** is `vloop run` itself: the deterministic half of the loop. It
plans nothing and judges nothing. It starts every session, runs every gate and
check, owns every status change and makes every commit. It is also the boundary
the permission fence only advises on: a session that moves a ref, a key of
`.git/config` or the git hooks halts the run with exit 9 (see [Exit
codes](#exit-codes)), and the driver restores what a session changed among its
inputs.

| Who | Does | Never does |
| --- | --- | --- |
| The driver (`vloop run`) | starts every session; runs every gate and check; sets every task and plan status; makes the one commit per iteration; halts and exits | plan, judge or review |
| The plan session | splits the brief into tasks and writes each task's gate: its `verify` command and its gate folder | commit, set a status |
| The gate-review session | judges every gate of a fresh plan and answers `PASS` or `FAIL` | commit, set a status |
| The work session | does one task and proposes an outcome | commit, set a status, change its gate |
| The review session | judges the work independently and answers `PASS` or `FAIL` | commit, set a status |
| The operator (you) | writes the brief; starts `vloop run`; decides at halts; amends gates; verifies and closes the brief | |

Sessions propose; the driver disposes.

## Skills

The plugin ships five skills. The first four are the sessions of the loop;
`vloop run` starts each one, you do not.

- `/vloop:plan` reads the brief and its binding references and writes the plan
  (`.vloop/state/state.json`): tasks with acceptance criteria, a verify command,
  and `kind` (and `area`, when `areas` is set).
- `/vloop:work` takes one task: reads the plan, the journal and the task's
  references, does the task, runs its gate, and writes `proposal/v1`
  (`done`, or `blocked`, with a `gate_dispute` when the gate is wrong). It never
  commits or sets a status.
- `/vloop:gate-review` judges every gate of a fresh plan before any work starts
  and writes a `PASS` or `FAIL` verdict with findings (see [Gates and the gate
  shell](#gates-and-the-gate-shell)).
- `/vloop:review` reads the diff and the proposal and writes `verdict/v1`:
  `PASS` or `FAIL`, per-criterion evidence, and findings of kind `bug`,
  `spec-gap` or `gate-gap`. It fails closed.
- `/vloop:operate` is yours: invoke it in an interactive session to run a brief
  on a work branch, handle halts, verify, record defects and interventions, close
  the brief and merge.

Each skill's behaviour is checked by evals in the plugin. They are run by the
operator with `claude plugin eval`; `vloop run` never runs them.

## Briefs and their lifecycle

A brief is a Markdown file `B<YYYYMMDD-HHMM>-<slug>.loop-brief.md` in
`docs/briefs/`, written with `vloop brief new` and checked with `vloop brief
check`. Its `status` moves `draft` → `ready` → `consumed`, or to `abandoned`
if it is dropped. Only `ready` briefs are checked. `consumed` means a loop built
it and it was closed; `abandoned` means it never will be, and it blocks the
briefs that depend on it.

## depends-on

`depends-on` lists briefs, by name, that must be `consumed` before this one can
run. `vloop brief list` shows the briefs in dependency order, ready or blocked.
A name matching no brief, or a cycle, is a problem.

## Binding references

A brief's `## Binding references` lists the files the work is bound by, each
with the reason it binds. The plan attaches the relevant ones to each task as
`references`, and both the work and the review session read them: a convention
nobody points at is a convention that gets broken.

## Plans and tasks

The plan is `.vloop/state/state.json` (schema `state/v2`): a list of tasks, each
with a goal, acceptance criteria, a `verify` command, its dependencies, its
references and a status. The work session only proposes an outcome; the gate and
the review decide. `vloop status` and `vloop task list` read the plan, and the
`vloop task` commands that change it refuse a result that fails `vloop task
validate`.

## Keeping the machine awake

A run can last hours, and a machine that sleeps stalls it. With `run.keep-awake`
on (the default), `vloop run` holds a no-idle-sleep hold for its own lifetime:
`caffeinate -i -w <pid>` on macOS, `systemd-inhibit --what=idle` on Linux where
it is installed, and `SetThreadExecutionState` on Windows. The hold ends when
`vloop run` exits, however it exits. If it cannot be taken, `run.log` gets one
warning line and the run continues; with `run.keep-awake = off` none is
attempted. It does not cover closing a laptop's lid, which sleeps the machine
regardless: keep the lid open, or change the OS's lid setting.

## Gates and the gate shell

A task's `verify` command is its **gate**: the task is done only when it exits 0.
`vloop task gate <id>` runs it in the shell the `shell` key names (`sh` by
default, `pwsh` on Windows). A gate is written before the work exists, and the
work session may not change it; only the operator does, with `vloop task verify
<id> <command> --reason <text>`, which records the change in the task's history.

**Gate fixtures.** What a gate needs besides its command (an oracle, a script,
expected outputs) lives in the task's gate folder, `.vloop/state/gates/<id>/`.
The plan session writes it, the plan stamps it, and the work session cannot
change it: the driver restores it from the stamp. `state.json`, `plan.md` and
`.vloop/state/gates/` belong to the current plan, and there is one plan per
repository. `vloop brief close` keeps them; they are replaced when the next
brief is planned, or when a new plan is rejected as unfit. Earlier plans' gate
folders live in git history. `vloop task verify <id>
[<command>]` replaces a gate's command, its fixtures or both, and records the
old ones, the reason and who replaced them.

**Gate scratch.** A gate may copy its oracle into a git-ignored scratch folder
named by `run.gate-scratch` (for example `web/.gate/`) to run it there. The
folders are emptied after every gate, and a gate that changes anything else in
the working tree fails. A gate that runs past `run.gate-timeout` is killed and
is not run again.

**The gate review.** Before any work, the driver runs every gate on the base.
A gate that passes there proves nothing, and one that changes the tree is
refused; both go back to the plan session. Then a separate gate-review session
(`/vloop:gate-review`) judges the gates and answers `PASS` or `FAIL`. A `FAIL`
sends its findings back to the plan session; after two rounds a plan still
failing is saved blocked and the run ends with exit 2. Amend the gates, then
`vloop run` again to resume.

The plan is stamped on load and stays `planning` until acceptance passes. If a
gate or the review halts the run during acceptance (exit 9), restore what moved
and run `vloop run <brief>` again: it prints `resuming acceptance of <run id>:
round <n>`, runs no second plan session, re-runs the base gates of the round and
goes on to the gate review. Work never starts without a passed gate review.

**Checks.** A `[[check]]` in `.vloop/config.toml` is a command that proves a
part of the repository still works, scoped by path globs. The driver runs every
check on the base before planning (and refuses to plan with none, or with one
failing) and, after each iteration whose gate passed, the checks whose paths
match the changes. A failing check ends the iteration as `check_failed`: no
review, one attempt charged. When every task is done, all checks run once more
in a final pass; a failure there ends the run blocked, exit 2. The fields and
keys are in [configuration.md](configuration.md#checks).

## What changed in 0.8

0.8 follows 0.7. Before the repository went public its releases were tagged
`v1.0.0`, `v2.0.0-beta.1`, `v2.0.0-beta.2` and `v2.0.0` (commits `fc0b67a`,
`42ee0a6`, `27871b2`, `d219473`); those tags were removed, and `go.mod` retracts
`v1.0.0` in case a copy surfaces, so `go install …@latest` resolves to the 0.x
line. Run records and PRs that name them are history. The
`v2` in schema names (`state/v2`, `intervention/v2`, `metrics/v2`) is a format
version, unrelated to the release number.

- **The gate model.** The plan is `state/v2`: a task's gate may have a
  gate folder of fixtures, stamped by the plan. Before any work the driver runs the
  gates on the base and a gate review judges them; `[[check]]` entries prove
  the repository still works after each iteration and at the end. See [Gates
  and the gate shell](#gates-and-the-gate-shell).
- **Interventions.** Records are `intervention/v2`, with options and a
  recommendation; `vloop intervention migrate` converts older records.
  `vloop metrics --interventions` reports agreement.
- **Metrics.** `metrics/v2` documents; `--workspace` and `metrics export` read
  across repositories. See [metrics.md](metrics.md).
- **Exit codes.** Exit `2` now means usage only; every other failure,
  including a failed plan session or a mid-run error, exits `1`. For `vloop run`
  `2` stays blocked (R-3). See [Exit codes](#exit-codes).

## What is redacted

Before anything is written under `.vloop/state/` (gate logs, session records,
reports, `run.log`, the journal), the driver replaces every occurrence of the
value of an environment variable whose **name** contains `KEY`, `TOKEN`,
`SECRET`, `PASSWORD` or `CREDENTIAL` (any letter case) and whose value is at
least 8 characters with `<redacted:NAME>`. Redaction keys on names, not on what
a value looks like: a secret held in an innocently named variable, such as
`DB_URL` or `AUTH`, is **not** caught, and neither is one shorter than 8
characters. Session records keep, per permission denial, only the tool name and
the file path when there is one — never the command or the content.

## Exit codes

Every vloop command exits `0` on success, `1` when it ran and found problems or
failed, and `2` only on a usage error (an unknown command, flag or config key, a
missing argument, an invalid value). `vloop run` uses the same numbers and
adds its own, one per way a run can end:

| Exit | Ending | Resumable as is |
| --- | --- | --- |
| `0` | complete: every task is done | — |
| `1` | problems or failure, including a failed plan session or a mid-run error | after fixing the cause |
| `2` | usage; for `vloop run` also blocked: tasks remain but none can run (R-3) | no — a human decides |
| `3` | stalled: iterations in a row closed nothing and charged no attempt | yes, once understood |
| `4` | max iterations: `run.max-iterations` is spent | yes |
| `5` | not converging: too many iterations per closed task | no |
| `6` | cost ceiling: `run.cost-ceiling` is reached | yes, with a higher ceiling |
| `7` | session error: a session failed to run | no |
| `8` | repeat blocked: a task blocked twice with nothing changed | no |
| `9` | refs or repository configuration moved: a session or a gate changed a ref, a key of `.git/config` (other than those `run.git-config-ignore` and the built-in editor keys ignore) or the git hooks; nothing was committed. The message names the key | no — restore the refs first; during acceptance, once restored, `vloop run` resumes acceptance, and the halted run's folder is kept and committed with the plan |

While it runs, `vloop run` prints one line per finished session: its phase and
task, how long it took, what it cost and the run's spend against the ceiling,
for example `work T3 · 4m12s · $0.81 · run $3.10 / $40.00`. It warns once when
the spend crosses 50% and again at 80% of `run.cost-ceiling`. From another
terminal, `vloop status` shows the live run as `running: work T3 for 2m05s ·
$3.10 of $40.00` (or `running: gate review · …`, or `running: between
sessions`); the line reads the run lock and appears only while a run holds it.

Resuming is running `vloop run` again on the work branch. Errors go to stderr as
one line starting with `vloop: `, after any checks `vloop doctor` or a preflight
prints.

## Metrics and defects

`vloop metrics` reports what a brief's loop produced, what it cost and how long
it took, recomputed from run folders, git and the plan on every call; nothing
is entered by hand. See [metrics.md](metrics.md). A **defect** is something wrong
that the loop produced or let through, counted by where it entered and who caught
it. Some are derived from the run and some are recorded by hand with `vloop
defect add`. See [defects.md](defects.md).

## Repository guidelines

`.vloop/repo-guidelines.md` holds this repository's lessons: short rules a
brief would otherwise repeat as a constraint, brief after brief. `vloop init`
writes it empty. When it exists and is not empty, `vloop run` hands it to every
session it starts (plan, gate review, work and review) as appended system
instructions; with no file, or an empty one, the sessions start as before. A
guideline never overrides the brief or a domain invariant.

A rule is a line starting with `- ` at the left margin, in any section.
`vloop brief check` does not read the file.

Every rule cites the defect or intervention that taught it (`D20261011-1455`,
`I20261011-0900`), in its line. `vloop doctor` warns, under `guidelines`, when
the file has more than 20 rules, and when a rule cites none. Retire or merge
rules before adding more; the budget is not configurable.

A lesson there replaces a constraint repeated brief after brief, which is the
point of the file. When a defect is one a guideline now prevents, record it:
`vloop defect set <id> fixed-by guideline`. `vloop brief close` prints each
defect's `rule` and `rule-outcome` (see [defects.md](defects.md)), and
`vloop metrics` shows how many words each session kind starts with (see
[metrics.md](metrics.md)).

## Closing and release

When a run is finished and you have verified it, `vloop brief close <brief>` on
the work branch records your findings as defects (`--finding`, or
`--no-findings` to say there were none), freezes the metrics as a snapshot, writes
the run record into the brief, marks it `consumed` and makes one commit. It
prints the `Vloop-Brief: <name>` trailer the squash-merge commit must carry.
`--abandon "<reason>"` closes an unfinished plan as `abandoned`, and `--dry-run`
shows what would happen and writes nothing. Close never merges and never pushes.

Release is not declared: vloop detects it as the brief's merge to the default
branch (`origin/HEAD`'s target, else `main`, else `master`), through that trailer
or the commit that marked the brief `consumed`. That is what lets `vloop defect
add --blame <file>:<line>` attribute a later bug to the brief that wrote the line.

### The stamped release build

Without a stamp, a binary names its commit from the VCS revision Go recorded
(`-dirty` when the tree was modified) and its version from the module, so
`go install` of a checkout carries a commit; one built with `-buildvcs=false`
carries none, and `vloop doctor` warns in this repository that the binary
carries no commit. Build releases with the stamped build:

```
go build -ldflags "-X main.version=<v> -X main.commit=<sha>" -o vloop ./cmd/vloop
```

Release steps, as vloop's own releases are made:

1. **A release branch and PR** (`release/v<v>`) that sets the version in
   `cmd/vloop/main.go` and `plugin/.claude-plugin/plugin.json`, and in the tests
   that pin it (`cmd/vloop/e2e_test.go`, `cmd/vloop/run_harness_test.go`,
   `internal/cli/root_test.go`). The full suite and `claude plugin validate` pass
   before it merges.
2. **Squash-merge** it to the default branch, with the operator's go-ahead.
3. **Tag** the merge commit: `git tag -a v<v> -m "<what it carries>"`, then push
   the tag. A version with a suffix (`0.9.0-beta.1`) is a pre-release; upgrading
   to or from one needs `vloop upgrade --yes`.
4. **Build stamped** with the tag's version and the merge commit's sha (the
   command above) wherever the released vloop runs.
5. **Upgrade the repository** with `vloop upgrade --yes` on its own branch, in its
   own PR: it refreshes the install stamp and the `CLAUDE.md` section. The README
   installs `@latest`, so it needs no change.

A binary stamped with HEAD still draws the self-hosting warning; build the next
vloop with a released one.

## Setting up a repository

`vloop init` sets a git repository up and never commits. It writes
`.vloop/config.toml` with the stacks it detects (a stack in a subdirectory is
scoped to it, like `csharp@services/api`), the **stamp** `.vloop/install.json`
(which vloop version set the repository up), a starter brief, the line
`.vloop/tmp/` in `.gitignore`, and vloop's section of `CLAUDE.md`, between
`<!-- vloop:begin -->` and `<!-- vloop:end -->` markers; text outside the markers
is kept. It refuses a repository already set up. `--dry-run` writes nothing.

`vloop upgrade` refreshes that `CLAUDE.md` section and `.gitignore` line and
rewrites the stamp. It refuses a repository set up by a newer vloop, and a
breaking jump (a new major version, or a new minor while 0.x) or a pre-release
needs `--yes`.

`go install` writes the binary to `$GOBIN`, or `$(go env GOPATH)/bin`
(`~/go/bin` by default), and never touches `PATH`; if `vloop` is not found, add
that directory:

```
export PATH="$PATH:$(go env GOPATH)/bin"
```

Put the line in the file your shell reads: zsh `~/.zprofile` (or `~/.zshrc`);
bash `~/.bash_profile` on macOS and `~/.profile` on most Linux. In PowerShell the
place is the user `Path` variable. The Go MSI installer usually adds
`%USERPROFILE%\go\bin` already; otherwise set it once, which new terminals,
`cmd` and apps inherit:

```
[Environment]::SetEnvironmentVariable('Path', [Environment]::GetEnvironmentVariable('Path', 'User') + ";$(go env GOPATH)\bin", 'User')
```

This is preferred over `$PROFILE`, which reaches PowerShell only.

`vloop doctor` checks the setup without writing anything: git, the install stamp,
the config, the `[[check]]` entries (`checks`), `claude`, workspace trust, the
gate shell, the plan, the branch, the plugin (a warning when no vloop plugin
is installed), `gate edits` (a warning, with the rule to add, unless the
operator's settings let a session edit `.vloop/state/gates/`), the scoped
`stacks` (printed only when `metrics.stacks` has scoped entries) and
`self-hosting`. It exits 1 on a problem.

The plugin ships inside the binary: `vloop plugin path` extracts it to
`.vloop/tmp/plugin/<version>/`, and refuses outside a repository where `vloop
init` ran. `vloop run` hands it to its sessions with `--plugin-dir`; for an
interactive session, from inside such a repository, start Claude Code the same
way:

```
claude --plugin-dir "$(vloop plugin path)"
```

`vloop version --check-plugin
<dir>` compares an installed plugin with the binary.

`vloop run`'s sessions load the plugin it hands them and not the plugins enabled
in your settings (measured 2026-10-11). For interactive sessions, either install
the plugin or pass `--plugin-dir`, not both.

## The .vloop/ layout

vloop keeps its files under `.vloop/` in the repository root:

- `.vloop/config.toml`: the settings, written by `vloop config set`.
- `.vloop/install.json`: the install stamp, written by `vloop init` and `vloop upgrade`.
- `.vloop/tmp/`: scratch space, git-ignored.
- `.vloop/state/state.json`: the plan.
- `.vloop/state/runs/<run id>/`: a run's folders, sessions and iteration records.
- `.vloop/state/metrics/<run id>.json`: the metrics snapshot `brief close` freezes.
- `.vloop/defects/`: one Markdown file per recorded defect.
- `.vloop/repo-guidelines.md`: this repository's lessons, handed to every session `vloop run` starts; see [Repository guidelines](#repository-guidelines).

Legacy: the shell loop that built vloop's first briefs. Its state lives in
`.loop/state/`; vloop reads it, in any repository that has it, and never writes
there. See [Retiring the shell loop](#retiring-the-shell-loop).

## Retiring the shell loop

For a repository that had the shell loop installed (`.loop/install.sh`) and now
runs vloop, remove:

- `.claude/skills/loop-plan`, `.claude/skills/loop-work` and
  `.claude/skills/loop-review`;
- the `<!-- loop:begin -->…<!-- loop:end -->` block in `CLAUDE.md`, first moving
  any rule it holds that vloop's section lacks into the repository's own part of
  `CLAUDE.md`: the shell loop's rule 1, that durable state never goes to
  `~/.claude`;
- `.claude/loop-knowledge.md`;
- `.loop/tmp/`.

Keep `.loop/state/`, with a one-line `.loop/RETIRED.md` saying the shell loop is
retired, because vloop reads its `state.json` and `runs/` for metrics and
defects. The rest of `.loop/` stays in git history.

## The flow

```
vloop brief new <slug>       write a draft brief, edit it
vloop brief check <path>     fix what it reports, set status: ready
(run the loop)               plan, then work → gate → review → commit per task
vloop defect add …           record what you find while verifying the run
vloop brief close <brief>    findings, snapshot, run record, consumed, one commit
(squash-merge)               with the trailer Vloop-Brief: <brief>
vloop defect add --blame …   after release, attribute a bug found later
vloop metrics [<brief>]      what it cost and delivered
vloop metrics export         the same as JSON Lines, numbers and titles only
vloop metrics --workspace <file>   one table across several repositories
```

`vloop metrics export` writes `export/v1` records with the repository's identity,
and `--workspace` reads the repositories a workspace file lists (see
[configuration.md](configuration.md)) and only reads them: each repository
owns its data.

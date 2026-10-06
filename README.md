# vloop

vloop is a Go command-line tool and Claude Code plugin that runs an autonomous
loop. You write a **brief** (what to build and why); vloop has Claude plan it
into tasks, then for each task a fresh work session does it, a gate checks it, an
independent review judges it, and vloop makes one commit. You read the result,
not every step.

## Prerequisites

- **Go**, at the version pinned in `go.mod`, to install vloop.
- **git**, with the repository you want to work in.
- The **`claude` CLI**, signed in. vloop starts it for every session.
- **Workspace trust**: open the repository once with `claude` and accept the trust
  prompt, or sessions started by vloop cannot work. `vloop doctor` checks it.

## Install

```
go install github.com/mvelosop/vloop/cmd/vloop@v2.0.0-beta.2
```

Use a release tag; `v2.0.0-beta.2` is the current one. The plugin ships inside the
binary. While the repository is private there is no marketplace; hand the plugin
to Claude Code with `--plugin-dir` for an interactive session:

```
claude --plugin-dir "$(vloop plugin path)"
```

`vloop run` passes it to its own sessions for you.

## Quickstart

In a git repository, on a clean tree:

1. `vloop init` sets the repository up: config with the stacks it detects, a
   starter brief, and vloop's section of `CLAUDE.md`. It never commits.
2. Add a `[[check]]` to `.vloop/config.toml`: a command that must keep passing
   after every task, such as your test suite.
3. `vloop doctor` verifies git, config, `claude`, trust and the rest. Fix what it
   reports before going on.
4. `vloop brief new <slug>` writes a draft brief in `docs/briefs/`. Fill in its
   sections and set `status: ready`.
5. `vloop brief check docs/briefs/<brief>.md` says whether the brief is fit to plan.
6. `vloop run` plans the brief on a work branch and works every task.
7. `vloop brief close <brief> --no-findings` (or `--finding "<summary>"` for each
   problem you found) records the run, marks the brief `consumed` and makes the
   closing commit. It never merges or pushes; that is yours.

## The loop in one paragraph

A **plan** session splits the brief into tasks, each with acceptance criteria and
a **gate**: a verify command, often with a gate folder of fixtures. Before any
work, the driver runs the gates on the base and a **gate review** judges them. Then
for each task a **work** session does it, the gates run, the repository's
`[[check]]` entries run, an independent **review** session judges the diff, and the
driver makes **one commit**. Every session is a fresh Claude session; they share
only files in the repository. A task that fails is tried again, and a run that
cannot go on halts and says why.

## Everyday commands

- `vloop status`, `vloop task list`, `vloop task show <id>`: where the run stands.
- `vloop brief list`: briefs in dependency order, ready or blocked.
- `vloop task reset <id>`, `vloop task note <id> <text>`, `vloop task verify <id> <command>`:
  amend the plan between runs.
- `vloop metrics`: a brief's size, time, cost and defects.
- `vloop defect add`, `vloop intervention add`: record what the loop got wrong, and
  what you had to do by hand.
- `vloop config list`: every setting with its value and where it came from.
- `vloop upgrade`: refresh a repository after a newer vloop.

Every command and flag is in [docs/guide/commands.md](docs/guide/commands.md).
`-C <dir>` and `--json` work on all of them.

## Exit codes

`0` is success, `1` is problems or failure, and `2` is a usage error — for
`vloop run`, also a blocked run. `vloop run` adds 3–9, one per way a run can end,
and says which can be resumed. They are in
[docs/guide/concepts.md](docs/guide/concepts.md#exit-codes). Errors go to stderr
as one line starting with `vloop: `, once the checks `vloop doctor` and a
preflight print have run.

## Guides

- [docs/guide/concepts.md](docs/guide/concepts.md): the loop, briefs, plans, gates, closing and release.
- [docs/guide/commands.md](docs/guide/commands.md): every command and flag, generated from the binary.
- [docs/guide/configuration.md](docs/guide/configuration.md): every config key, the stack presets and the workspace file.
- [docs/guide/metrics.md](docs/guide/metrics.md): every `vloop metrics` number and the stack presets.
- [docs/guide/defects.md](docs/guide/defects.md): origin and catcher, derived and recorded defects.
- [docs/guide/evals.md](docs/guide/evals.md): testing the skills against a real model.

### Skills

The plugin's five skills: `/vloop:plan`, `/vloop:gate-review`, `/vloop:work` and
`/vloop:review` are the sessions `vloop run` starts; `/vloop:operate` is yours, to
run, verify, close and merge a brief.

## What changed in 2.0

- **A gate model.** A task's gate can carry fixtures; the gates are judged before
  any work starts, and `[[check]]` entries prove the repository still works.
- **Exit 2 is usage only.** Every other failure exits 1; `vloop run` keeps 2 for
  blocked.
- **Interventions with options**, and `vloop metrics --interventions`.
- **Metrics across repositories**: `--workspace` and `vloop metrics export`.
- **Plain messages**: each error says what to do next, and the help is rewritten.

The details are in [docs/guide/concepts.md](docs/guide/concepts.md#what-changed-in-20).

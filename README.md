# vloop

vloop (short for **verification loop**) is a Go command-line tool and a plugin
for Claude Code and GitHub Copilot CLI. It runs an autonomous loop whose gate
and independent review must agree before work counts as done. You write a
**brief**; an agent plans the brief into tasks, fresh work sessions do them,
gates check them, and independent reviews judge them. vloop makes one commit per task.

## Prerequisites

- **Go**, at the version pinned in `go.mod`, to install vloop.
- **git**, with the repository you want to work in.
- The **`claude` CLI**, signed in, when a session runs on Claude (the default).
  vloop starts it for those sessions.
- **Workspace trust**, when a session runs on Claude: open the repository once with
  `claude` and accept the trust prompt, or sessions started by vloop cannot work.
  `vloop doctor` checks it.
- The **`copilot` CLI** and a **GitHub login** its sessions can use, when a session
  runs on Copilot (experimental). See [Running on GitHub Copilot](#running-sessions-on-github-copilot).

## Install

```
go install github.com/mvelosop/vloop/cmd/vloop@latest
```

`go install` puts the binary in `$(go env GOPATH)/bin` (`~/go/bin` by default),
which is often not on your PATH. If `vloop` is not found, add it:

```
export PATH="$PATH:$(go env GOPATH)/bin"
```

The file to put that line in, per shell, is in
[docs/guide/concepts.md](docs/guide/concepts.md#setting-up-a-repository).
On Windows, PowerShell users can find the `Path` command in
[docs/guide/concepts.md](docs/guide/concepts.md#setting-up-a-repository). To
update later, `vloop self-update` installs the latest release and updates the
plugins; then run `vloop upgrade` in each repository.

The plugin ships inside the binary, and `vloop run` hands it to its own sessions.
For an interactive Claude Code session with the plugin, install it from this
repository's marketplace, run inside the repository:

```
claude plugin marketplace add mvelosop/vloop
claude plugin install vloop@vloop --scope project
```

That installs the plugin for that repository (without `--scope project`, for
every repository).

`vloop init` declares the plugin in `.claude/settings.json`, so trusting the
folder in Copilot CLI installs it for that repository; in Claude Code, run the
install above. To install it for Copilot CLI per user instead:

```
copilot plugin marketplace add mvelosop/vloop
copilot plugin install vloop@vloop
```

Or, without installing it, from inside a repository where you ran `vloop init`,
for one session:

```
claude --plugin-dir "$(vloop plugin path)"
copilot --plugin-dir "$(vloop plugin path)"
```

## Quickstart

In a git repository, on a clean tree:

1. `vloop init` sets the repository up: config with the stacks it detects, a
   starter brief, vloop's section of `CLAUDE.md`, and the plugin's declaration in
   `.claude/settings.json`. It never commits.
2. Add a `[[check]]` to `.vloop/config.toml`: a command that must keep passing
   after every task, such as your test suite.
3. `vloop doctor` verifies git, config, and the agent CLIs in use (`claude` and
   workspace trust, or `copilot` and its GitHub login). Fix what it reports.
4. `vloop brief new <slug>` writes a draft brief in `docs/briefs/`. Fill in its
   sections and set `status: ready`.
5. `vloop brief check docs/briefs/<brief>.md` says whether the brief is fit to plan.
6. `vloop run` plans the brief on a work branch and works every task.
7. `vloop brief close <brief> --no-findings` (or `--finding "<summary>"` for each
   problem you found) records the run, marks the brief `consumed` and makes the
   closing commit. It never merges or pushes; that is yours.

## Mixed macOS and Windows teams

Declare the team's operating systems in the shared `.vloop/config.toml`:

```toml
platforms = ["macos", "windows"]
shell = "pwsh"
```

Neither a local file nor an environment variable can override either setting
for a mixed team. vloop refuses a plan whose gates call a command from its fixed
list of tools that some platforms lack, including `grep` and `sed`; use portable
pwsh cmdlets and the project's own toolchain instead. vloop runs `pwsh` gates
and checks in UTF-8. This repository is a mixed macOS and Windows team. See the
full [mixed-team configuration rules](docs/guide/configuration.md#mixed-teams).

## Running sessions on GitHub Copilot

Each session kind can run on the GitHub Copilot CLI instead of Claude Code
(experimental). Choose the agent and a Copilot model id per session kind:

```
vloop config set agent.work copilot
vloop config set model.work gpt-6.1-terra
```

The keys are `agent.<kind>` and `model.<kind>` for `plan`, `gate-review`,
`work`, `review`, `second-review`, `verify` and `review-brief`. A Copilot
session needs a Copilot model id, never a Claude alias such as `sonnet`;
`vloop run` refuses one. To list the ids your account can use, see
[Copilot model ids](docs/guide/configuration.md#copilot-model-ids). So far `gpt-6.1-terra` is qualified for work, review,
the second review and brief review; keep plan and gate review on Claude, and
verify is not yet tried. `mai-code-2` has run every session kind but the second
review (plan, gate review, work, review, verify and brief review) in the URL shortener acceptance
run and in B40 to B44, which built vloop itself, and is not yet qualified by the
[eval rule](docs/guide/evals.md#other-agents); see the [run evidence](docs/design-notes/copilot-cli-experiments.md#mai-code-2-in-the-loop-2026-11).

- **Requirements:** `copilot` on PATH, and a GitHub login its sessions can use
  (`COPILOT_GITHUB_TOKEN`, `GH_TOKEN` or `GITHUB_TOKEN`, else `gh auth login`).
  `claude` is needed only while a session in use runs on Claude.
- **Safety:** each session gets its own Copilot home under `.vloop/tmp/` with a
  hook that runs vloop's fence before every tool call, and the driver re-judges
  the calls afterwards.
- **Cost** is estimated from Copilot credits at `copilot.credit-usd` dollars each.
- **Windows:** the hook has a PowerShell form. Copilot runs on Windows were
  used for B40 through B44; see
  [Mixed macOS and Windows teams](#mixed-macos-and-windows-teams).
- **Overriding locally:** `.vloop/config.local.toml` is git-ignored and beats
  `.vloop/config.toml`. To run Claude while the shared file says `copilot`, run
  `vloop config set --local agent.work claude` and `vloop config set --local
  model.work sonnet` (likewise for each kind); `vloop config list` shows where
  each value came from.

Every key and what `vloop doctor` checks are in
[docs/guide/configuration.md](docs/guide/configuration.md).

## The loop in one paragraph

A **plan** session splits the brief into tasks with acceptance criteria and
**gates**. The [driver](docs/guide/concepts.md#the-driver) runs the gates on the base and a **gate review** judges them.
For each task, a **work** session does it, the gates and `[[check]]` entries run,
an independent **review** judges it, and the driver makes **one commit**. Every
session is a fresh agent session (Claude Code or Copilot CLI), sharing only repository files.

## You ask; the skill drives

With the plugin loaded, ask your agent (Claude Code or Copilot CLI) in plain
words ("run this brief", "what did that cost?", "record this as a defect").
The `/vloop:operate` skill runs the right vloop commands and checks the result.
Merging and pushing stay yours.

## Everyday commands

What you can ask for, and the command the skill runs for it:

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

The plugin's eight skills: `/vloop:plan`, `/vloop:gate-review`, `/vloop:work`,
`/vloop:review` and `/vloop:verify` are the sessions `vloop run` starts. Three are yours,
used in an interactive session: `/vloop:operate` to run, verify, close and merge a
brief, and `/vloop:create-brief` and `/vloop:review-brief` to write a brief and
have it reviewed independently.

## What changed in 0.8

- **A gate model.** A task's gate can carry fixtures; the gates are judged before
  any work starts, and `[[check]]` entries prove the repository still works.
- **Exit 2 is usage only.** Every other failure exits 1; `vloop run` keeps 2 for
  blocked.
- **Interventions with options**, and `vloop metrics --interventions`.
- **Metrics across repositories**: `--workspace` and `vloop metrics export`.
- **Plain messages**: each error says what to do next, and the help is rewritten.

The details are in [docs/guide/concepts.md](docs/guide/concepts.md#what-changed-in-08).

## License

MIT — see [LICENSE](LICENSE).

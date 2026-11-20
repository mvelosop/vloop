---
name: B20261120-1643-b39-local-config-copilot-only.loop-brief
description: Let each developer override the shared .vloop/config.toml with an untracked .vloop/config.local.toml, let a machine with no Claude run vloop when no session uses Claude, and price a Copilot credit at GitHub's $0.01
kind: brief
status: ready
created: 2026-11-20
seeds: A mixed team on one repository (the client on GitHub Copilot under Windows and pwsh, the operator on Claude), decided with the operator on 2026-11-20; GitHub's individual Copilot billing page as read on 2026-11-20
depends-on: [B20261106-2013-b38-i140-module-zips.loop-brief]
---
# Brief — vloop B39: a local config file, Claude only where it is used, and the price of a Copilot credit

- **Starting point:** extends `main` at `f4e6bd6a`, with the README's
  Copilot section (branch `docs/readme-copilot`, `efde799d`) and this
  brief's ready commit merged into the run branch when the run starts, by
  the operator's decision of 2026-11-20. The planner pins that SHA as the
  **base** every gate compares against — never `HEAD~n`.
- **Produced by:** operator decision, 2026-11-20, in an interactive session
  following `plugin/skills/create-brief/`. The operator chose: one brief for
  both the local file and the Claude-only checks; any key may be overridden
  locally but `[[check]]` may not; `vloop config set --local`; the three
  guards (ignore line, refusal when tracked, warning naming the local keys);
  the credit default moved to `0.01`.
- **Review:** `vloop brief review` (Claude Opus), 4 rounds, verdict `FIX`
  each time, with 7, 5, 4 and 3 findings, no `ASK`; every finding fixed.
  Fixed: the key order of `config list`; what is read when; the tests
  that pin the old default and the `.gitignore`; a tracked local file
  (refused by name) and the existing tree-clean refusal of an untracked
  one; the cases with no git repository, an unreadable config and both
  overrides; Claude counted as in use when the config cannot be read; the
  concepts guide's real passages; the roadmap row and its inherited
  decision; task 3 split in two; the run refusal's exact lines; doctor's
  rows asserted one by one; empty lists; S-2 and S-4 cited; the notice
  from `config set --local`. Round 4's three fixes (the tree-clean
  refusal, the `.vloop/state/` exception, that notice) are not
  re-reviewed: they correct wording and add no reach.

## Executive summary

A repository holds one `.vloop/config.toml`, shared through git. A team
whose members use different agents (the client on GitHub Copilot, on Windows
with `pwsh`; the operator on Claude) cannot share it: whatever one sets for
`agent.*` and `model.*`, the other must override with `VLOOP_*` environment
variables in every shell (`internal/config/config.go:373`, `resolve`). And a
machine with no `claude` cannot run vloop at all, even when every session is
on Copilot: `vloop run`'s preflight fails on doctor's `claude` row and on a
`trust` warning (`internal/cli/run.go:192`, `internal/cli/doctor.go:116` and
`:132`). Copilot's cost is estimated at `0.0066667` dollars per credit,
Copilot Pro's $10 spread over its 1,500 included credits; GitHub now prices
one AI credit at $0.01, on every plan.

**The goal:** a developer keeps their own settings in an untracked
.vloop/config.local.toml that wins over the shared file, sees where every
value came from, and cannot commit it by accident; a Copilot-only machine
runs vloop without Claude installed; and Copilot's estimated cost uses
GitHub's price of a credit unless the developer sets their own.

- **L1 — the local file.** .vloop/config.local.toml holds any key
  `config.toml` holds and wins over it; the environment still wins over both.
  `[[check]]` stays in the shared file.
- **L2 — writing it.** `vloop config set --local` writes and removes keys in
  it; `vloop config path --local` names it.
- **L3 — keeping it local and visible.** `init` and `upgrade` git-ignore it;
  `vloop run` refuses while git would commit it; `vloop doctor` and
  `vloop run` name the keys it overrides; sessions cannot change it.
- **L4 — Claude only where it is used.** Doctor's `claude`, `trust`,
  `plugin` and `gate edits` rows apply only when a session in use resolves
  to the `claude` agent; otherwise they are `n/a` and the run's preflight
  passes them.
- **L5 — the price of a credit.** `copilot.credit-usd` defaults to `0.01`.

Unchanged: the keys and their validation, `[[check]]`, the fences' rules for
every other path, how sessions are started, and everything when no local file
exists and some session uses Claude.

## What it is

A second, untracked config layer between the shared file and the
environment, with the guards that keep it out of git and in view; the
Claude-specific preflight checks made conditional on Claude being used; and
one default changed. It is for a repository shared by people whose agents,
models, budgets or Copilot plans differ.

## Why this shape, and what was rejected

Decided with the operator on 2026-11-20. **Do not re-litigate these** — the
planner inherits them.

- **Order: environment, then local file, then shared file, then default.**
  The environment is a one-shell override and stays the strongest, as C-1
  has it. *Rejected:* the local file over the environment — then a
  `VLOOP_*` variable set for one command would silently lose.
- **The file sits in `.vloop/`, beside the shared one, named
  config.local.toml.** It is in the repository (C-1: nothing is global)
  and follows the `*.local.*` convention of .claude/settings.local.json.
  *Rejected:* a file under the home directory — the drift CLAUDE.md's rule 1
  forbids; a `[local]` table in the shared file — it is committed.
- **Any key, never `[[check]]`.** Budgets, `shell`, `run.keep-awake` and
  the Copilot price legitimately differ by machine and plan. The checks are
  the repository's contract that every task keeps passing; a check that only
  runs on one machine is not one. *Rejected:* only `agent.*`, `model.*` and
  `effort.*` — the operator wants budgets and the credit price per
  developer; local `[[check]]` tables — not reproducible.
- **A list key is replaced, not merged.** `metrics.code = [...]` in the
  local file is the value, as an environment variable's list is today.
  *Rejected:* appending to the shared list — two ways to read one key.
- **The run refuses a local file that is shared, or about to be, and says
  why.** Today an untracked, un-ignored local file already stops a run, but
  only as one path in the generic "the tree is not clean" refusal
  (`internal/driver/plan.go:204`), which does not say the file must be
  ignored. A tracked local file passes that check and is shared with
  everyone who clones: it is not local at all. L3 refuses both, by name,
  before anything runs. *Rejected:* having the driver leave the file out of
  its commits — the tree-clean check would still stop every run, and an
  interactive `git add -A` would still take it.
- **The local file is a session input, like the shared one.** Sessions must
  not change the configuration the driver runs under. *Rejected:* trusting
  that it is ignored — ignored files are exactly the ones the driver's tree
  checks do not see.
- **One brief for L1–L5.** They serve one scenario: a mixed team, some of it
  on machines with no Claude. The local file is useless to a Copilot-only
  developer who cannot run vloop.
- **Claude "in use" follows the rule Copilot's checks already follow.**
  `copilotKeys` (`internal/cli/doctor.go:309`) counts `agent.verify` only
  while `run.verify` is not `false` and `agent.second-review` only while
  `run.second-review` is not `off`; L4 counts the `claude` keys the same
  way. *Rejected:* skipping the Claude rows whenever `claude` is not on
  PATH — that hides a missing `claude` from a repository that needs it.
- **The credit is priced at $0.01.** GitHub's individual billing page
  (docs.github.com, "Copilot billing for individuals", read 2026-11-20)
  states "1 AI credit = $0.01 USD"; base credits are worth exactly that on
  every paid plan (Pro $10 for 1,000, Pro+ $39 for 3,900, Max $100 for
  10,000), the Flex allotment (Pro 500, Pro+ 3,100, Max 10,000) is used
  after them at the same rate, and usage beyond both costs $0.01 a credit.
  `0.0066667` is Pro's price over base plus Flex, an average that fits one
  plan and undercounts usage beyond the allowance. *Rejected:* keeping the
  old default — wrong for every plan but Pro; modelling allowances (base,
  Flex, month-to-date usage) — vloop cannot read a developer's usage, and
  it belongs to a later brief.

## Binding references

- `docs/domain/domain-model.md` — C-1 (config in the repository only,
  overridden by the environment): L1 adds a layer and this run rewrites
  C-1's sentence to name it; S-2 (the driver is the boundary and puts
  back the inputs a session changed) and S-4 (a session writes nothing under
  `.vloop/` but `.vloop/tmp/`): L3 adds the local file to the restored inputs,
  and S-4 is why the Copilot hook needs no change.
- `docs/design-notes/vloop-roadmap.md` — "Decisions every brief in the series
  inherits", whose "Config lives in the repo only" line this run amends.
- `docs/domain/platform/platform-context.md` — "Configuration": the
  resolution order this run extends, and the "CLI conventions" (exit 1 on a
  problem, exit 2 on usage, one `vloop: ` line on stderr).
- `internal/config/config.go` — `resolve`, `resolveList`, `readFile`,
  `SourceError`, `Get`, `List`, `Set` and `Keys`: where the layer goes.
- `internal/config/root.go` — `FilePath`, the shared file's path constant,
  beside which the local one belongs.
- `internal/cli/config.go` — `config list`, `get`, `set` and `path`, their
  text and `--json` output.
- `internal/cli/doctor.go` — `runDoctor`'s rows and their order,
  `doctorClaude`, `doctorTrusted`, `doctorPlugin`, `doctorGateEdits`,
  `copilotKeys` and the rule L4 mirrors.
- `internal/cli/run.go` — `runPreflight`: which rows fail a run and which
  warnings it prints.
- `internal/cli/init.go` — `vloop init`'s `.gitignore` step, which L3 extends.
- `internal/cli/upgrade.go` — `vloop upgrade`'s `.gitignore` merge, which L3 extends.
- `internal/install/section.go` — `IgnoreLine` and `MergeGitignore`, the ignore-line rules L3 reuses.
- `internal/driver/safety.go` — `snapshotInputs`: the files a session must
  leave alone.
- `fence/work.json` — its deny entries for `.vloop/config.toml`, which L3 extends.
- `fence/review.json` — the same.
- `fence/verify.json` — the same.
- `fence/gate-review.json` — the same.
- `fence/review-brief.json` — the same.
- `docs/design-notes/fence-audit.md` — the rule-by-rule table every fence
  rule has a row in.
- `docs/guide/configuration.md` — the key table and the resolution
  paragraph this run documents L1, L2 and L5 in.

## Behaviour contract

### L1 — the local file

- **Path:** .vloop/config.local.toml under the repo root, `/` on every OS.
  The same TOML format and keys as `.vloop/config.toml`. Keys vloop does not
  know are ignored, as in the shared file.
- **Resolution, per key:** the `VLOOP_<KEY>` variable when set and not
  empty; else the local file when it sets the key; else the shared file
  when it sets the key; else the default. A list key takes the whole list
  of the layer that wins. An empty list (`metrics.code = []`) sets nothing, in
  the local file as in the shared one today, so the next layer decides.
- **Source:** a value from the local file has source `local` in `vloop
  config list` (`agent.work=claude (local)`) and in `--json` output of
  `config list` and `config get` (`"source":"local"`). The other sources
  keep their names: `default`, `file`, `env`.
- **Errors:** a bad value or malformed TOML in the local file is reported
  as the shared file's are, with the local path:
  `vloop: .vloop/config.local.toml: invalid value "nope" for agent.work: want one of claude, copilot`,
  exit 1.
- **`[[check]]` in the local file** is an error wherever the config is
  read, exit 1:
  `vloop: .vloop/config.local.toml: [[check]] belongs in .vloop/config.toml — the checks are the repository's, not one machine's`.
- **Read when the shared file is read, as today:** the local file is read
  wherever and whenever `.vloop/config.toml` is read now (`ResolveRun` at a
  run's start for the sessions' model, effort and agent; other keys where
  the driver reads them today). This run changes no read point; a session
  cannot change the file (L3).

### L2 — writing it

- `vloop config set --local <key> <value>` validates as `config set` does
  (exit 2 on an unknown key or invalid value) and writes the key into the
  local file, preserving its other keys and layout as `config set`
  preserves the shared file's. It creates the file (and `.vloop/`) only
  when it writes a value. `''` removes the key; removing a key that is not
  there writes nothing and creates nothing.
- Without `--local`, `config set` writes the shared file, as today. When,
  after the write, the key's value still comes from the local file or the
  environment, it prints one line on stderr and exits 0:
  `! agent.work stays claude on this machine: .vloop/config.local.toml overrides it`
  (or `: VLOOP_AGENT_WORK overrides it`). When both override, the line
  names the environment variable, the one that wins. `config set --local`
  prints the same line when, after its write, a `VLOOP_*` variable still
  overrides the key: `! agent.work stays copilot on this machine: VLOOP_AGENT_WORK overrides it`.
- `vloop config path --local` prints .vloop/config.local.toml; `config
  path` alone prints `.vloop/config.toml`, as today.

### L3 — keeping it local and visible

- **Ignore line.** `vloop init` and `vloop upgrade` keep
  .vloop/config.local.toml in `.gitignore`, on a line of its own, beside
  `.vloop/tmp/`, with the same rules as that line (added once, nothing else
  changed, reported as today's `.gitignore` step).
- **Doctor row `config.local`,** right after the `config` row:
  - no local file: no row;
  - the `config` row is a problem (the local file, or the shared one, cannot
    be read): no row, as doctor's other config rows are skipped then;
  - the repo root is not a git repository: the ignore test is skipped and
    the row is the warning or the pass below;
  - the file is tracked (in git's index, whether or not an ignore rule
    names it): problem, since `git add -A` stages a tracked file's changes,
    `✗ config.local .vloop/config.local.toml is tracked by git — it is shared with everyone who clones: run git rm --cached .vloop/config.local.toml and add it to .gitignore`;
  - the file is untracked and git does not ignore it (as the driver's git
    sees it): problem,
    `✗ config.local .vloop/config.local.toml is not git-ignored — the next git add -A would commit it: add it to .gitignore`;
  - untracked and ignored, and some keys resolve with source `local`: warning, naming
    them in the order `config list` prints them,
    `! config.local overrides model.work, agent.work — on this machine only`;
  - untracked and ignored, and no key resolves with source `local`: `✓ config.local`.
- **`vloop run`'s preflight** fails on either `config.local` problem
  (counted like any problem, nothing runs) and prints the warning; both on
  stderr in the preflight's form, two spaces then the symbol, as it prints
  the `agent.*` rows today.
- **A session input.** The driver's input guard holds
  .vloop/config.local.toml like `.vloop/config.toml`: changed, created or
  removed by a session, it is put back. Each fence whose deny list names
  `Edit(.vloop/config.toml)` and `Write(.vloop/config.toml)` names
  .vloop/config.local.toml the same way; the Copilot hook already denies
  every write under `.vloop/` but `.vloop/tmp/` and is unchanged.
  `docs/design-notes/fence-audit.md` gains the rows.

### L4 — Claude only where it is used

- **In use for `vloop run`:** any of `agent.plan`, `agent.gate-review`,
  `agent.work`, `agent.review` resolves to `claude`; or `agent.verify`
  does, while `run.verify` is not `false`; or `agent.second-review` does,
  while `run.second-review` is not `off`.
- **In use for `vloop doctor`:** the same, or `agent.review-brief` resolves
  to `claude`.
- **When the configuration cannot be read** (the `config` row is a problem),
  Claude counts as in use: the four rows run as today, so a broken config
  never hides a missing `claude`.
- **When Claude is in use:** the `claude`, `trust`, `plugin` and
  `gate edits` rows are exactly as today.
- **When it is not:** each of the four rows is `n/a`, runs nothing and
  reads nothing, with the message `no session in use runs on claude`:
  `- claude no session in use runs on claude`, and the same for `trust`,
  `plugin` and `gate edits`. `vloop run`'s preflight passes them.
- `vloop brief review`'s preflight is unchanged: it already checks Claude
  only when `agent.review-brief` is `claude`.

### L5 — the price of a credit

- `copilot.credit-usd` defaults to `0.01`. Its validation (a number
  greater than 0) is unchanged.
- The configuration guide's row for it says: dollars per GitHub Copilot AI
  credit; GitHub's price is $0.01 on every plan; a developer whose plan
  makes included credits cheaper sets their own in the local file.
- Session records written before this run keep their costs; nothing
  recomputes them.

### Documents

- `docs/guide/configuration.md`: the resolution paragraph names the local
  file; a short section on it (L1–L3, `[[check]]` excluded, the ignore
  line, the refusal, the warning); the `copilot.credit-usd` row (L5); the
  `config set --local` example.
- `docs/domain/domain-model.md` C-1 and
  `docs/domain/platform/platform-context.md` "Configuration": the order
  environment, local file, shared file, default.
- `docs/guide/concepts.md`: where it lists what `init` writes to
  `.gitignore` and what is git-ignored, the new line; in the `vloop doctor`
  paragraph, that the `claude`, workspace trust, `plugin` and `gate edits`
  rows apply when a session in use runs on Claude, and are `n/a` otherwise.
- `README.md`: the Copilot section's "Overriding locally" bullet names
  .vloop/config.local.toml and `vloop config set --local` in place of the
  environment variables, and its requirements bullet drops "`vloop run`
  still needs `claude` on PATH … even when every session is on Copilot".
  The Prerequisites' `claude` and workspace-trust bullets say they apply
  when a session runs on Claude. The
  README stays under its 200-line cap.
- `docs/design-notes/vloop-roadmap.md`, "Decisions every brief in the series
  inherits": the "Config lives in the repo only" line names the local file
  and its place in the order.
- `docs/guide/commands.md` is regenerated (`go generate ./...`), as its
  test requires.

### Decided here, because the cited documents leave it underdetermined

- The local file's `[[check]]` refusal applies to every command that reads
  the config, not only `vloop run`: the configuration is one thing.
- A local file that exists but is empty, or sets only keys vloop does not
  know, is ignored like an empty shared file and gets `✓ config.local`.
- The test asks git about the file path itself (.vloop/config.local.toml),
  not about a directory, so any ignore rule that covers it (a
  `.vloop/*.local.toml` pattern, the global excludes file) satisfies it. A
  helper built for directories is not reused as it stands.
- The `config set` notice (L2) goes to stderr, so `config set` stays silent
  on stdout and scripts reading stdout are unaffected.

### Violations the review must rule on

- A key resolved from the local file whose source prints as `file`.
- The local file staged or committed by any driver commit in a test run.
- A Claude row skipped on a repository where some session in use is on
  Claude, or a Claude row still run when none is.
- `copilot.credit-usd`'s old default left anywhere a default is stated
  (the key table, a test's expected `config list`, `cmd/vloop/e2e_test.go`)
  — except files that record what already happened or quote it, which
  keep it: every brief under `docs/briefs/` (this one included), session
  anything under `.vloop/state/` (session records, run folders, journals,
  brief reviews), design notes.
- A `[[check]]` read from the local file in any path.

## Worked example

In a fresh git repository in a temporary directory, with `vloop init` run
and a `[[check]]` added to `.vloop/config.toml`:

```
$ vloop config set agent.work copilot
$ vloop config set model.work gpt-6.1-terra
$ vloop config set --local agent.work claude
$ vloop config set --local model.work sonnet
$ vloop config path --local
.vloop/config.local.toml                                       exit 0
$ vloop config list | grep -E '^(agent|model)\.work='
model.work=sonnet (local)
agent.work=claude (local)                                      exit 0
$ VLOOP_AGENT_WORK=copilot vloop config get --json agent.work
{"key":"agent.work","value":"copilot","source":"env"}          exit 0
$ vloop config set model.work gpt-6-luna
! model.work stays sonnet on this machine: .vloop/config.local.toml overrides it   (stderr)  exit 0
$ grep -x '.vloop/config.local.toml' .gitignore
.vloop/config.local.toml                                       exit 0
$ vloop doctor | grep config.local
! config.local overrides model.work, agent.work — on this machine only            exit 0
$ vloop config set --local agent.work ''
$ vloop config set --local model.work ''
$ vloop doctor | grep config.local
✓ config.local                                                 exit 0
```

The failures, each planted in a scratch copy:

```
.gitignore without the line, the local file present:
$ vloop doctor | grep config.local
✗ config.local .vloop/config.local.toml is not git-ignored — the next git add -A would commit it: add it to .gitignore
$ vloop run <brief>     (in a repository with no other problem; stderr)
  ✗ config.local .vloop/config.local.toml is not git-ignored — the next git add -A would commit it: add it to .gitignore
vloop: preflight failed — 1 problem(s), nothing has run                    exit 1

The local file committed, then the line added to .gitignore:
$ vloop doctor | grep config.local
✗ config.local .vloop/config.local.toml is tracked by git — it is shared with everyone who clones: run git rm --cached .vloop/config.local.toml and add it to .gitignore

[[check]] in the local file:
$ vloop config list
vloop: .vloop/config.local.toml: [[check]] belongs in .vloop/config.toml — the checks are the repository's, not one machine's   exit 1

agent.work = "nope" in the local file:
$ vloop config get agent.work
vloop: .vloop/config.local.toml: invalid value "nope" for agent.work: want one of claude, copilot   exit 1

$ vloop config set --local agent.work nope
vloop: invalid value "nope" for agent.work: want one of claude, copilot   exit 2
```

Claude not in use, on a PATH with a stub `copilot` and no `claude`, every
`agent.*` key `copilot` with a Copilot model id, `run.verify = false`:

`vloop doctor` prints, each on its own line and among its other rows (in
today's row order, with its summary line last), these four:

```
- claude no session in use runs on claude
- trust no session in use runs on claude
- plugin no session in use runs on claude
- gate edits no session in use runs on claude
```

A test asserts each line on its own, not the four as one block; and `vloop run`'s preflight passes those rows. With `agent.review` set back
to `claude` (and `claude` still absent), `vloop doctor` reports
`✗ claude claude is not on PATH` and the run's preflight fails on it, as
today.

`vloop config get copilot.credit-usd` prints `0.01` with no file setting
it. At that price, B33's acceptance sessions (19.3 credits,
`docs/design-notes/copilot-cli-experiments.md`) estimate at $0.193, not
$0.129.

**Lines no session can run:** `vloop init`, `vloop run` and `vloop doctor`
with the stubbed PATH are covered by end-to-end tests the plan writes with
stub agents (the pattern of `cmd/vloop/e2e_test.go`), one per block above:
the local-file block, each planted failure (the run refusal included), and
the Claude-not-in-use block (doctor's rows and the run's preflight passing).

## Acceptance, at close (the operator's)

1. On a machine (or a `PATH`) with `copilot` and without `claude`: in a
   fresh clone of `mvelosop/url-shortener-template`, every `agent.*` key
   `copilot`, `model.*` set to `gpt-6.1-terra` (plan and gate review as
   well, for the preflight only), `vloop doctor` shows the four `n/a` rows
   and `vloop run` gets past its preflight into planning.
2. In the same clone, with `claude` on PATH, `vloop config set --local` for
   every `agent.*` key to `claude` and `model.*` back to the defaults:
   `vloop run` runs every session on Claude, and its commits do not include
   .vloop/config.local.toml (`git log --stat` on the branch).
3. On Windows with `pwsh` (the client's machine, when available): the first
   block of the worked example gives the same output.

## Out of scope

- A per-machine `[[check]]`, or extending the fences from config (#133).
- Modelling Copilot plan allowances (base, Flex, month-to-date usage) in
  the cost estimate; recomputing the costs of earlier session records.
- A `--local` for `config get` or `config list` (the source column already
  says where a value came from).
- The second agent's doctor rows beyond L4 (Copilot's rows are unchanged),
  and Claude's trust or plugin checks for interactive sessions.
- A user-level config file of any kind.
- Clearing a shared list on one machine (an empty list that overrides):
  an empty list sets nothing, in either file.
- The Copilot reviewer's reading of `/dev/null` and `$TMPDIR` (#136) and
  the other open fence issues (#129, #130, #134), which this run will not
  touch even where it edits a fence file.

## Constraints

- Go, as pinned in `go.mod`; no new dependency (the TOML library in use
  stays the only one).
- Gates run on macOS's BSD tools; no GNU-only flags in a gate.
- Sessions never move refs. Tests build their repositories in temporary
  directories and address them with `git -C`; no test touches this
  repository's own `.vloop/` or `.gitignore`.
- Tests that need agents use stubs on a temporary `PATH`; no test runs a
  real `claude` or `copilot`, and none depends on what is installed on the
  machine.
- **Gates follow the diff surface.** Each task gates on the packages it
  touches (`go test ./internal/config/`, `go test ./internal/cli/ -run …`),
  never on `./...`; the `go` check runs the whole suite after every task.
- **No task may weaken a check to pass.** An earlier test changes only where
  it pins behaviour this contract changes (the `copilot.credit-usd`
  default in expected `config list` output, doctor's rows when Claude is
  not in use, the exact `.gitignore` that `init` and `upgrade` write), and only to the new behaviour.
- Repo-relative paths everywhere. No absolute paths in any file or commit
  message.

## Shape

6 to 8 tasks, each independently verifiable by a single command. Order is
load-bearing:

1. **The local layer in resolution (L1):** the source `local`, the
   precedence, the list rule, the errors and the `[[check]]` refusal, in the
   config package with its tests, fixtures in temporary directories. Gated
   on `go test ./internal/config/`.
2. **`config set --local`, `config path --local` and the stderr notice
   (L2),** with `docs/guide/commands.md` regenerated. Gated on the CLI's
   config tests.
3. **Keeping it out of git and in view (L3, first half):** the ignore
   line in `init` and `upgrade`, doctor's `config.local` row and the
   run's preflight (refusal and warning). Gated on the init, upgrade, doctor
   and run-preflight tests it adds.
4. **Keeping sessions off it (L3, second half):** the input guard and the
   fence entries, with their `fence-audit.md` rows. Gated on the driver
   safety tests and the fence tests it adds.
5. **Claude only where it is used (L4):** doctor's four rows and the run's
   preflight, with an end-to-end test on a PATH holding a stub `copilot`
   and no `claude`. Gated on that test and doctor's.
6. **The credit price (L5):** the default, the tests that pin it, the
   guide's row. Gated on `go test ./internal/config/ ./internal/cli/` for
   the `config list` cases and `go test ./cmd/vloop/ -run
   TestWorkedExampleEnglishSession`, which pins the default too.
7. **The documents:** the configuration guide's section, C-1, the platform
   context, the concepts guide, the README, and the roadmap's inherited
   decision. Gated on the README and guide
   tests (`go test ./internal/cli/ -run 'Readme|Guide'`) and a `grep` for
   config.local.toml in each named document. If its review fails twice on
   wording, it is blocked and the operator finishes it.
8. **Close:** every gate at once, and the worked example line for line
   through the end-to-end tests.

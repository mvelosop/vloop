# Configuration guide

vloop's settings are per repository, in `.vloop/config.toml` under the repo root.
Nothing is read from your home directory. An environment variable overrides the
file, and the file overrides the default. `vloop config list` shows every key
with its value and where it came from; `vloop config set <key> <value>` writes
one (`''` removes it); `vloop config get <key>` prints one.

An invalid value is exit 2 from `config set`, and exit 1 when it comes from the
environment or the file.

## Keys

The variable of a key is `VLOOP_` and the key in upper case with `.` and `-` turned
into `_`. List keys are TOML arrays in the file and comma-joined text in the
environment and in `config set`.

| Key | Default | Valid values | Environment variable | Effect |
| --- | --- | --- | --- | --- |
| `language` | `en` | `en`, `es` | `VLOOP_LANGUAGE` | language of a brief's section headings and of the template `brief new` writes; commands, flags, keys and JSON stay English |
| `model.plan` | `opus` | any model name | `VLOOP_MODEL_PLAN` | model of the plan session |
| `model.work` | `sonnet` | any model name | `VLOOP_MODEL_WORK` | model of the work sessions |
| `model.review` | `sonnet` | any model name | `VLOOP_MODEL_REVIEW` | model of the review sessions |
| `model.second-review` | `sonnet` | any model name | `VLOOP_MODEL_SECOND_REVIEW` | model of the second review session |
| `model.gate-review` | `sonnet` | any model name | `VLOOP_MODEL_GATE_REVIEW` | model of the gate-review session |
| `model.verify` | `sonnet` | any model name | `VLOOP_MODEL_VERIFY` | model of the verify session |
| `effort.plan` | unset | `low`, `medium`, `high`, `xhigh`, `max` | `VLOOP_EFFORT_PLAN` | effort of the plan session |
| `effort.work` | unset | same | `VLOOP_EFFORT_WORK` | effort of the work sessions |
| `effort.review` | unset | same | `VLOOP_EFFORT_REVIEW` | effort of the review sessions |
| `effort.second-review` | unset | same | `VLOOP_EFFORT_SECOND_REVIEW` | effort of the second review session |
| `effort.gate-review` | unset | same | `VLOOP_EFFORT_GATE_REVIEW` | effort of the gate-review session |
| `effort.verify` | unset | same | `VLOOP_EFFORT_VERIFY` | effort of the verify session |
| `agent.plan` | `claude` | `claude` | `VLOOP_AGENT_PLAN` | agent CLI of the plan session |
| `agent.work` | `claude` | `claude` | `VLOOP_AGENT_WORK` | agent CLI of the work sessions |
| `agent.review` | `claude` | `claude`, `copilot` | `VLOOP_AGENT_REVIEW` | agent CLI of the review sessions (`copilot` is experimental) |
| `agent.second-review` | `claude` | `claude`, `copilot` | `VLOOP_AGENT_SECOND_REVIEW` | agent CLI of the second review session (`copilot` is experimental) |
| `agent.gate-review` | `claude` | `claude` | `VLOOP_AGENT_GATE_REVIEW` | agent CLI of the gate-review session |
| `agent.verify` | `claude` | `claude` | `VLOOP_AGENT_VERIFY` | agent CLI of the verify session |
| `copilot.credit-usd` | `0.0066667` | a number greater than 0 | `VLOOP_COPILOT_CREDIT_USD` | dollars per GitHub Copilot credit, for the estimated cost of a copilot review session |
| `shell` | `sh` (`pwsh` on Windows) | `sh`, `bash`, `pwsh`, `powershell`, `cmd` | `VLOOP_SHELL` | shell `vloop task gate` runs a verify command in |
| `areas` | unset | a list of names of lower-case letters, digits and hyphens | `VLOOP_AREAS` | the names a task's `area` may take; when set, `vloop task validate` reports any other |
| `metrics.stacks` | unset | a list of preset names (below) | `VLOOP_METRICS_STACKS` | presets that classify changed lines as code, test, docs or excluded |
| `metrics.code` | unset | a list of doublestar globs | `VLOOP_METRICS_CODE` | the repository's own `code` globs; they win over the presets |
| `metrics.test` | unset | a list of doublestar globs | `VLOOP_METRICS_TEST` | the repository's own `test` globs |
| `metrics.docs` | unset | a list of doublestar globs | `VLOOP_METRICS_DOCS` | the repository's own `docs` globs |
| `metrics.excluded` | unset | a list of doublestar globs | `VLOOP_METRICS_EXCLUDED` | the repository's own `excluded` globs, which are not counted |
| `run.max-iterations` | `30` | an integer, 0 or more | `VLOOP_RUN_MAX_ITERATIONS`  | iterations a run may take |
| `run.cost-ceiling` | `40` | a number of dollars above 0 | `VLOOP_RUN_COST_CEILING`  | dollars a run may spend |
| `run.max-attempts` | `3` | an integer, 1 or more | `VLOOP_RUN_MAX_ATTEMPTS`  | attempts a task gets |
| `run.stall-limit` | `2` | an integer, 1 or more | `VLOOP_RUN_STALL_LIMIT`  | consecutive iterations without progress that end the run |
| `run.convergence-max` | `3.0` | a number above 0 | `VLOOP_RUN_CONVERGENCE_MAX`  | iterations per closed task above which a run is not converging (after `run.convergence-min` iterations) |
| `run.convergence-min` | `6` | an integer, 0 or more | `VLOOP_RUN_CONVERGENCE_MIN`  | iterations a run must have taken before convergence is judged |
| `run.gate-timeout` | `15` | a whole number of minutes, 1 or more | `VLOOP_RUN_GATE_TIMEOUT`  | minutes a gate may run, in `vloop run` and in `vloop task gate`; a gate still running is killed with everything it started and has failed |
| `run.gate-scratch` | unset | a list of doublestar globs | `VLOOP_RUN_GATE_SCRATCH` | repo-relative folders (for example `web/.gate/`) a gate may write into; they are emptied after every gate, and a gate that changes anything else in the tree fails |
| `run.git-config-ignore` | unset | a list of patterns over `git config --list` keys, `*` matching any run of characters, dots included | `VLOOP_RUN_GIT_CONFIG_IGNORE` | keys of `.git/config` the guard leaves out when it compares the file around every session, gate and check, added to the built-in `branch.*.vscode-merge-base` and `branch.*.github-pr-*` (VS Code's and its GitHub Pull Requests extension's bookkeeping), which are always applied. The guard compares entries, not bytes, so any other key added, removed or changed halts the run (exit 9) and is named. A pattern that could ignore a `core`, `filter`, `include` or `includeif` key is refused: `vloop config set` exits 2, `vloop run` exits 1 |
| `run.session-timeout` | `60` | a whole number of minutes, 1 or more | `VLOOP_RUN_SESSION_TIMEOUT`  | minutes a plan, work or review session may run; a session still running is killed with everything it started and is a session error (exit 7) |
| `run.second-review` | `off` | `off`, `record`, `feedback`, `on-pass`, `both-pass` | `VLOOP_RUN_SECOND_REVIEW` | whether and how a second reviewer judges each task after the primary; `vloop doctor` warns when it is the primary's agent and model, which adds no independence |
| `run.verify` | `true` | `true` or `false` | `VLOOP_RUN_VERIFY`  | whether `vloop run` ends with a verify session that checks the built surface against the brief |
| `run.keep-awake` | `on` | `on` or `off` | `VLOOP_RUN_KEEP_AWAKE`  | `on` holds a no-idle-sleep hold for the lifetime of `vloop run` (macOS `caffeinate`, Linux `systemd-inhibit`, Windows `SetThreadExecutionState`); `off` takes none |

**Copilot review sessions are experimental.** `agent.review = "copilot"` (or `agent.second-review`) runs the review sessions on the GitHub Copilot CLI. Every Copilot session runs under the fence hook, and the adapter's deny list is also on its command line. `vloop doctor` and `vloop run` warn `! agent.review is copilot: experimental (docs/guide/configuration.md)` (and the same for `agent.second-review`), check that `copilot` is on PATH and that `copilot --version` runs, and that a GitHub login reaches Copilot sessions (a Copilot session logs in through `COPILOT_GITHUB_TOKEN`, `GH_TOKEN` or `GITHUB_TOKEN`, set and not empty, else `gh auth token`: run `gh auth login`; an interactive `/login` does not reach vloop's sessions; the token is never printed or kept). `vloop run` refuses to start while `model.review` (or `model.second-review`, when that agent is in use) is still the default `sonnet`, a Claude alias: set it to a Copilot model id.

### The fence hook

Copilot does not read the review fence, so each Copilot session has its own home (`COPILOT_HOME`, under `.vloop/tmp/`, emptied between sessions) holding a hooks file whose `preToolUse` entry runs `vloop fence hook --phase <review|second-review> --root <repository root>` before every tool call. The hook reads one payload on stdin and prints one line, `{"permissionDecision":"allow"}` or `{"permissionDecision":"deny","permissionDecisionReason":"vloop fence: …"}`. It reads the review fence's own Bash deny rules and the adapter's deny-list stems, not a copy, with the fence's matching (`pre:*` is the prefix alone or followed by a space, `*` elsewhere any run of characters).

`vloop fence check --phase <review|second-review> --root <dir> --tool <name> [--path <p>] ["<command>"]` runs the same judge without running anything: it prints `allow` (exit 0) or `deny: <reason>` (exit 1); bad arguments are exit 2. Use it to see what the hook would say to a command.

What the judge does:

- A shell command is split at `;`, `&&`, `||`, `|`, `&` and newlines outside quotes and looked through `env`, `command`, `exec`, `nohup`, `time`, `nice`, `timeout`, `stdbuf`, `xargs` and `sh -c` (also `bash`, `zsh`, `dash`, `ksh`); every simple command is matched.
- **It fails closed.** It denies what it cannot parse: a backtick, `$(`, `<(`, `>(`, `eval`, `source`, `.` as a command, an unterminated quote, a `-c` string that is missing or not literal, a command word that starts with `$`, wrappers nested more than eight deep. A payload it cannot read, an unknown phase or any internal error is a deny line and exit 2; it never allows on an error. A tool it does not list (`web_fetch`, any `*_subagent`, `task`, and the rest) is denied.
- Writes (`create`, `edit`, `apply_patch`, a redirection) are allowed only under `.vloop/tmp/`; `/dev/null` is allowed. Reads may not touch `~/.claude`, `~/.copilot` or `~/.config/gh`.
- **K1: a Copilot session uses vloop's skills only.** The hook allows the `skill` tool only for the name of one of vloop's skills, and denies any other with `vloop fence: skill <name> is not vloop's — use /vloop:review`. `vloop run` refuses to start (exit 1), and `vloop doctor` shows an error, while a skill directory under `.claude/skills/` or `.github/skills/` has exactly the name of one of vloop's skills and a Copilot agent is in use; rename it. The warning for other names ending in `review` stays.
- A denial lands in the session record's `permission_denials`, as `{tool_name, tool_use_id, message}`.

The hook is a line in front of the driver, not a replacement for it: the driver's checks after the session stay the boundary. [The fence audit](../design-notes/fence-audit.md) says, rule by rule, who enforces what.

### The second review

A second review is a second reviewer, of any agent and model, that judges the same task as the primary review in a fresh session of its own, under the review fence. It never sees the primary's verdict. `run.second-review` says when it runs and what its verdict does; `agent.second-review`, `model.second-review` and `effort.second-review` choose the reviewer (a task's own `model` and `effort` never set them). A Copilot second review needs an explicit Copilot model. Its cost counts toward `run.cost-ceiling`, and `vloop doctor` warns when it is the primary's agent and model, which adds no independence.

| Mode | Runs | What its verdict does |
| --- | --- | --- |
| `off` | never | |
| `record` | after every primary review | kept in the run folder; changes nothing |
| `feedback` | after every primary review | when the primary did not pass, its findings join the task's notes after the primary's, each prefixed `second review: `; the outcome is the primary's |
| `on-pass` | only after a primary PASS | kept; changes nothing |
| `both-pass` | only after a primary PASS | a task is done only when both pass; a second FAIL rejects the iteration as a primary FAIL does (the attempt counts), its findings becoming the notes |

Only `both-pass` can fail a task. In every other mode the second review changes neither the outcome, the status nor the attempts, and a second session that errors or times out is a warning. A **dissent** is a second verdict that differs from the primary's. The run log says `second review: <PASS|FAIL>` and adds `(dissent)` when it is one; the verdict is kept as `reports/<iteration>-second-verdict.json`; the end of the run lists `second review: <n> dissent(s): <task> (primary <verdict>, second <verdict>), …`; `vloop metrics` reports the sessions and dissents. The operator reads the dissents at close.

`model.*`, `effort.*` and `agent.*` are stored under `[model]`, `[effort]` and `[agent]`, the
`metrics.*` keys under `[metrics]`, the `run.*` budgets under `[run]` as TOML numbers. the two timeouts are whole minutes. `vloop run`'s `--max-iterations`, `--cost-ceiling`, `--max-attempts` and `--stall-limit` flags override the first four `run.*` keys for one run (flag, then environment, then file, then default). For example:

```
language = "es"
areas = ["cli", "docs"]

[model]
work = "opus"

[metrics]
stacks = ["go"]
code = ["internal/brief/templates/**"]
```

### Checks

A check is a command that proves a part of the repository still works. Checks are
`[[check]]` tables in `.vloop/config.toml`, the only place they are defined:
there is no environment variable for them, and `vloop config set` leaves them in
place. `vloop config set` edits the file's text in place: comments, order and
layout stay, and a new key goes beside the keys of its table.

```
[[check]]
name = "api"
paths = ["api/**"]
run = "sh api/test.sh"

[[check]]
name = "web"
paths = ["web/**", "shared/*.ts"]
run = "sh web/test.sh"
```

| Field | Valid values | Meaning |
| --- | --- | --- |
| `name` | lower-case letters, digits and hyphens; unique | what the check is called |
| `paths` | a non-empty list of repo-relative doublestar globs, as in `metrics.*` | the files that make the check apply |
| `run` | one command | what is run, in the configured `shell` |

A duplicate or malformed name, empty `paths`, a bad glob or a missing `run` makes
reading the config fail with a `vloop: ` message, as any bad value in the file
does. `vloop config list` prints each check after the keys, in config order, as
`check.<name>=<run> (paths: <globs>)`.

How the globs and presets are layered is in [metrics.md](metrics.md).

### Scoped stacks

An entry of `metrics.stacks` is `<stack>` or `<stack>@<path>`, for example
`csharp@services/api`. The path is repo-relative, uses `/`, has no leading or
trailing `/` and no `.` or `..` segment, and must name an existing directory;
anything else is exit 2 from `config set` (`want <stack> or <stack>@<existing
directory>`).

```
[metrics]
stacks = ["go", "csharp@services/api"]
```

A path's scope is the **longest** scoped path that contains it. Inside a scope
only that scope's stacks apply, and their globs match relative to the scope's
directory; outside every scope the unscoped stacks apply. `vloop init` writes a
stack found in a subdirectory this way. `vloop doctor` reports a scoped path
that no longer exists. See [metrics.md](metrics.md).

## The install stamp

`vloop init` writes `.vloop/install.json`, the stamp of the vloop that set the
repository up, following the `install/v1` schema (`vloop schema show
install/v1`): `schema`, `version`, `commit`, `initialized` and `upgraded`
(`null` until the first `vloop upgrade`). `vloop upgrade` and `vloop doctor`
read it; nothing else does. `vloop init` also adds the line `.vloop/tmp/` to
`.gitignore`: that folder is scratch space (`vloop plugin path` extracts the
plugin there) and is never committed. See [concepts.md](concepts.md).

## Stack presets

`vloop metrics stacks` lists them and `vloop metrics stacks <name>` prints one.
| Preset | code | test | excluded |
| --- | --- | --- | --- |
| `csharp` | `**/*.cs`, `**/*.razor`, `**/*.cshtml` | `**/*.Tests/**`, `**/*Tests.cs` | `**/bin/**`, `**/obj/**`, `**/*.Designer.cs`, `**/packages.lock.json` |
| `go` | `**/*.go` | `**/*_test.go`, `**/testdata/**` | `go.sum`, `vendor/**` |
| `java` | `**/*.java` | `**/src/test/**` | `**/build/**`, `**/target/**` |
| `javascript` | `**/*.js`, `**/*.mjs`, `**/*.cjs` | `**/*.test.js`, `**/*.spec.js`, `**/__tests__/**`, `**/*.e2e-spec.js`, `**/test/**` | `package-lock.json`, `yarn.lock`, `pnpm-lock.yaml`, `**/dist/**`, `**/node_modules/**` |
| `kotlin` | `**/*.kt`, `**/*.kts` | `**/src/test/**` | `**/build/**`, `**/target/**` |
| `python` | `**/*.py` | `**/test_*.py`, `**/*_test.py`, `**/tests/**` | `**/__pycache__/**`, `poetry.lock`, `uv.lock`, `Pipfile.lock` |
| `react` | `**/*.tsx`, `**/*.jsx`, `**/*.css`, `**/*.scss` | `**/*.test.tsx`, `**/*.spec.tsx`, `**/*.test.jsx`, `**/*.spec.jsx`, `**/*.stories.*` | `package-lock.json`, `yarn.lock`, `pnpm-lock.yaml`, `**/dist/**`, `**/node_modules/**` |
| `rust` | `**/*.rs` | `**/tests/**`, `**/benches/**` | `**/target/**`, `Cargo.lock` |
| `typescript` | `**/*.ts` | `**/*.test.ts`, `**/*.spec.ts`, `**/__tests__/**`, `**/*.e2e-spec.ts`, `**/test/**` | `package-lock.json`, `yarn.lock`, `pnpm-lock.yaml`, `**/dist/**`, `**/node_modules/**` |

Every preset's `docs` is `**/*.md`.

## The workspace file

`vloop metrics --workspace <file>` and `vloop metrics export --workspace <file>`
read a TOML file of `[[repo]]` tables, one per repository clone. It is usually
kept in a repository of its own.

```
[[repo]]
path = "../vloop"

[[repo]]
path = "../shop"
name = "shop-web"
```

- `path` (required): the clone's directory, relative to the workspace file (or
  absolute). It must be a git repository.
- `name` (optional): the repository's name in the output; by default the
  basename of its `origin` URL, else its directory name.

A listed path that is not a git repository prints `vloop: workspace repo not
found: <path>` on stderr, the others are still reported, and the exit code is 1.
The workspace is only read; giving it `<brief>` arguments is exit 2.

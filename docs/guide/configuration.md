# Configuration guide

vloop's settings are per repository, in `.vloop/config.toml` under the repo root.
Nothing is read from your home directory. Each key resolves as: environment
variable, else `.vloop/config.local.toml`, else `.vloop/config.toml`, else the
default. `vloop config list` shows every key
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
| `model.plan` | `opus` | a Claude alias or model name; a Copilot model id when agent.<kind> is copilot | `VLOOP_MODEL_PLAN` | model of the plan session |
| `model.work` | `sonnet` | a Claude alias or model name; a Copilot model id when agent.<kind> is copilot | `VLOOP_MODEL_WORK` | model of the work sessions |
| `model.review` | `sonnet` | a Claude alias or model name; a Copilot model id when agent.<kind> is copilot | `VLOOP_MODEL_REVIEW` | model of the review sessions |
| `model.second-review` | `sonnet` | a Claude alias or model name; a Copilot model id when agent.<kind> is copilot | `VLOOP_MODEL_SECOND_REVIEW` | model of the second review session |
| `model.gate-review` | `sonnet` | a Claude alias or model name; a Copilot model id when agent.<kind> is copilot | `VLOOP_MODEL_GATE_REVIEW` | model of the gate-review session |
| `model.verify` | `sonnet` | a Claude alias or model name; a Copilot model id when agent.<kind> is copilot | `VLOOP_MODEL_VERIFY` | model of the verify session |
| `model.review-brief` | `opus` | a Claude alias or model name; a Copilot model id when agent.<kind> is copilot | `VLOOP_MODEL_REVIEW_BRIEF` | model of the session `vloop brief review` runs |
| `effort.plan` | unset | `low`, `medium`, `high`, `xhigh`, `max` | `VLOOP_EFFORT_PLAN` | effort of the plan session |
| `effort.work` | unset | same | `VLOOP_EFFORT_WORK` | effort of the work sessions |
| `effort.review` | unset | same | `VLOOP_EFFORT_REVIEW` | effort of the review sessions |
| `effort.second-review` | unset | same | `VLOOP_EFFORT_SECOND_REVIEW` | effort of the second review session |
| `effort.gate-review` | unset | same | `VLOOP_EFFORT_GATE_REVIEW` | effort of the gate-review session |
| `effort.verify` | unset | same | `VLOOP_EFFORT_VERIFY` | effort of the verify session |
| `effort.review-brief` | unset | same | `VLOOP_EFFORT_REVIEW_BRIEF` | effort of the session `vloop brief review` runs |
| `agent.plan` | `claude` | `claude`, `copilot` | `VLOOP_AGENT_PLAN` | agent CLI of the plan session (`copilot` is experimental) |
| `agent.work` | `claude` | `claude`, `copilot` | `VLOOP_AGENT_WORK` | agent CLI of the work sessions (`copilot` is experimental) |
| `agent.review` | `claude` | `claude`, `copilot` | `VLOOP_AGENT_REVIEW` | agent CLI of the review sessions (`copilot` is experimental) |
| `agent.second-review` | `claude` | `claude`, `copilot` | `VLOOP_AGENT_SECOND_REVIEW` | agent CLI of the second review session (`copilot` is experimental) |
| `agent.gate-review` | `claude` | `claude`, `copilot` (experimental) | `VLOOP_AGENT_GATE_REVIEW` | agent CLI of the gate-review session |
| `agent.verify` | `claude` | `claude`, `copilot` | `VLOOP_AGENT_VERIFY` | agent CLI of the verify session (`copilot` is experimental) |
| `agent.review-brief` | `claude` | `claude`, `copilot` | `VLOOP_AGENT_REVIEW_BRIEF` | agent CLI of the session `vloop brief review` runs |
| `copilot.credit-usd` | `0.01` | a number greater than 0 | `VLOOP_COPILOT_CREDIT_USD` | dollars per GitHub Copilot AI credit; GitHub's price is $0.01 on every plan; a developer whose plan makes included credits cheaper sets their own in the local file |
| `platforms` | unset | a list of `linux`, `macos`, `windows` | `VLOOP_PLATFORMS` | the platforms the repository's team uses; set it in `.vloop/config.toml`, never in the local file or environment |
| `shell` | `sh` (`pwsh` on Windows) | `sh`, `bash`, `pwsh`, `powershell`, `cmd` | `VLOOP_SHELL` | shell `vloop task gate` runs a verify command in; a team with two or more platforms requires shared `pwsh` |
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

**Copilot plan, work, review, verify and gate-review sessions are experimental.** `agent.plan = "copilot"` runs the plan session, `agent.work = "copilot"` runs the work sessions, and `agent.review = "copilot"` (or `agent.second-review`) the review sessions, and `agent.verify = "copilot"` the verify session, and `agent.gate-review = "copilot"` the gate-review session, on the GitHub Copilot CLI. Every Copilot session runs under the fence hook, and the adapter's deny list is also on its command line. `vloop doctor` and `vloop run` warn `! agent.review is copilot: experimental (docs/guide/configuration.md)` (and `! agent.plan is copilot: experimental (docs/guide/configuration.md)` for the plan session, the same for `agent.work`, `agent.gate-review`, `agent.second-review` and `agent.verify`, the last only while `run.verify` is not `false`; `agent.plan` and `agent.gate-review` have no switch, so they warn whenever they are `copilot`, also on a resume that will not plan), check that `copilot` is on PATH and that `copilot --version` runs, and that a GitHub login reaches Copilot sessions (a Copilot session logs in through `COPILOT_GITHUB_TOKEN`, `GH_TOKEN` or `GITHUB_TOKEN`, set and not empty, else `gh auth token`: run `gh auth login`; an interactive `/login` does not reach vloop's sessions; the token is never printed or kept). `vloop run` refuses to start while `model.work` (or `model.plan`, `model.gate-review`, `model.review`, `model.second-review` or `model.verify`, when that agent is in use) is still the default (`sonnet`; for `model.plan`, `opus`), a Claude alias: `agent.work is copilot but model.work is the default "sonnet", a Claude alias: set model.work to a Copilot model id` (`agent.verify is copilot but model.verify is the default "sonnet", a Claude alias: set model.verify to a Copilot model id` for the verify session; `agent.gate-review is copilot but model.gate-review is the default "sonnet", a Claude alias: set model.gate-review to a Copilot model id` for the gate-review session; `agent.plan is copilot but model.plan is the default "opus", a Claude alias: set model.plan to a Copilot model id` for the plan session). A plan session that exits non-zero halts with `planning session failed (copilot exited <code>) — see <run folder>`.

**`agent.review-brief = "copilot"` is experimental too, and is read by `vloop brief review`, not by `vloop run`.** `vloop brief review <brief>` warns `! agent.review-brief is copilot: experimental (docs/guide/configuration.md)` and refuses, with nothing run, when `copilot` is not on PATH, when a project skill has the name of a plugin skill, when no GitHub login reaches Copilot (`agent.review-brief is copilot but no GitHub login reaches its sessions: set COPILOT_GITHUB_TOKEN, GH_TOKEN or GITHUB_TOKEN, or run gh auth login`) and while `model.review-brief` is still the default (`agent.review-brief is copilot but model.review-brief is the default "opus", a Claude alias: set model.review-brief to a Copilot model id`). `vloop run` ignores the key.

### The local file

`.vloop/config.local.toml` is your own, untracked layer: the same format and keys as
`.vloop/config.toml`, any key but `[[check]]` (the checks are the repository's; a
`[[check]]` there is exit 1: `vloop: .vloop/config.local.toml: [[check]] belongs in
.vloop/config.toml — the checks are the repository's, not one machine's`), `platforms`,
and `shell` on a mixed team (see [Mixed teams](#mixed-teams); `vloop config set --local`
refuses both, exit 2). It wins
over the shared file and loses to the environment; a list key is replaced, not
merged. `vloop config list` and `--json` show its values with source `local`.

```
vloop config set --local agent.work claude
vloop config set --local model.work sonnet
vloop config path --local        # .vloop/config.local.toml
```

`config set --local` writes only that file (`''` removes a key). `config set` without
`--local` prints `! model.work stays sonnet on this machine: .vloop/config.local.toml
overrides it` on stderr when the local file (or a `VLOOP_*` variable) still decides the key.

`vloop init` and `vloop upgrade` add `.vloop/config.local.toml` to `.gitignore`. `vloop
run` refuses, and `vloop doctor` shows `✗ config.local`, while the file is tracked or
not git-ignored; when it overrides keys, both warn `! config.local overrides model.work,
agent.work — on this machine only`. A session cannot change it: the driver puts it back.

## Mixed teams

Declare the operating systems used by the team in the shared
`.vloop/config.toml` file:

```toml
platforms = ["macos", "windows"]
shell = "pwsh"
```

The `platforms` values are `linux`, `macos` and `windows`. The list's stored
order is preserved by `vloop config get platforms`, but vloop deduplicates and
sorts it as `linux`, `macos`, `windows` in messages. Two or more distinct
entries make a mixed team. Requiring one shared PowerShell on every mixed team,
including a macOS/Linux team, avoids differences between BSD and GNU tools as
well as tools missing from Windows. The plan records `shell`, so every gate in
that plan runs in the same `pwsh` even when the team runs vloop on different
platforms.

`platforms` always belongs to `.vloop/config.toml`, never the local file or
the environment. A mixed team's `shell` also belongs there, so no machine can
write checks or gates for a different shell. These settings are errors, not
warnings, when read from the wrong source; `config set --local` exits 2 and
writes nothing:

```text
vloop: .vloop/config.local.toml: platforms belongs in .vloop/config.toml — the platforms are the repository's, not one machine's
vloop: VLOOP_PLATFORMS: platforms belongs in .vloop/config.toml — the platforms are the repository's, not one machine's
vloop: .vloop/config.local.toml: shell belongs in .vloop/config.toml while platforms lists macos, windows — the checks and the gates are written for one shell
vloop: VLOOP_SHELL: shell belongs in .vloop/config.toml while platforms lists macos, windows — the checks and the gates are written for one shell
```

The machine's local file and environment may still override `shell` when the
team is not mixed. A mixed team with no shared `shell`, or one set to anything
other than `pwsh`, fails `vloop doctor` and `vloop run` before planning.

vloop runs `pwsh` gates and checks with UTF-8 input and output, so native tool
output remains correct on Windows. This applies to the PowerShell processes
vloop starts, not to your own interactive PowerShell terminal; vloop does not
change your PowerShell profile or the machine's code page. On Windows each such
process runs on a console of its own, so the code pages it sets never reach the
terminal you ran vloop from. Add this line to your
PowerShell profile to keep `vloop run` output correct when it passes through
`Tee-Object` and similar terminal pipes:

```powershell
[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new()
```

Keep checks simple and portable: use plain `program arguments && program
arguments` commands whose programs exist on every declared platform, or let one
project program implement the check. For example:

```toml
[[check]]
name = "go"
paths = ["**"]
run = "go test ./... && go vet ./..."
```

The equivalent program-run shape is `run = "go run ./tools/check"`, which is
what this repository uses so its PowerShell, Go and other health checks live
in one portable toolchain entry point. A listed native command appearing in a
mixed check produces a doctor warning; the same command in a gate makes plan
acceptance fail.

Mixed-team gates use `pwsh` cmdlets and the project's own toolchain (`go`,
`node`, `pnpm`, or equivalent), never an OS-specific command. The exact list
checked by the driver, in order, is `awk`, `basename`, `bash`, `cat`, `chmod`,
`cmd`, `cmp`, `cp`, `cut`, `date`, `diff`, `dirname`, `env`, `fc`, `find`,
`findstr`, `grep`, `head`, `ln`, `ls`, `mkdir`, `mktemp`, `mv`, `powershell`,
`printf`, `readlink`, `rm`, `robocopy`, `sed`, `seq`, `sh`, `sleep`, `sort`,
`stat`, `tail`, `tee`, `test`, `touch`, `tr`, `uniq`, `wc`, `which`, `xargs`,
`xcopy`, `zsh`. Match their intent with portable facilities:

| Listed command | Portable replacement |
| --- | --- |
| `awk` | `Select-String`, PowerShell string operations, or the project's toolchain |
| `basename` | `Split-Path -Leaf` |
| `bash`, `cmd`, `powershell`, `sh`, `zsh` | `pwsh` commands or the project's own toolchain; do not start another shell |
| `cat` | `Get-Content` |
| `chmod` | `Get-Acl`/`Set-Acl`, or the project's toolchain when it owns file permissions |
| `cmp`, `diff`, `fc` | `Compare-Object`, `Select-String`, or a project/test-tool comparison |
| `cp`, `robocopy`, `xcopy` | `Copy-Item` |
| `cut`, `sed`, `tr` | PowerShell string operators, `Select-String`, or the project's toolchain |
| `date` | `Get-Date` |
| `dirname` | `Split-Path -Parent` |
| `env` | `$env:NAME`, `Get-ChildItem Env:`, or the process invocation's environment |
| `find`, `grep`, `findstr` | `Get-ChildItem` for paths and `Select-String` for content |
| `head`, `tail` | `Select-Object -First` or `Select-Object -Last` |
| `ln` | `New-Item -ItemType SymbolicLink` only when the operation is genuinely needed and supported |
| `ls` | `Get-ChildItem` |
| `mkdir` | `New-Item -ItemType Directory` |
| `mktemp` | `[System.IO.Path]::GetTempPath()`, `New-TemporaryFile`, or the gate scratch folder |
| `mv` | `Move-Item` |
| `printf` | PowerShell interpolation, formatting and `Set-Content`/`Out-File` |
| `readlink` | `Resolve-Path` |
| `rm` | `Remove-Item` |
| `seq` | PowerShell's range operator (`1..10`) or `ForEach-Object` |
| `sleep` | `Start-Sleep` |
| `sort`, `uniq` | `Sort-Object`, with `-Unique` where appropriate |
| `stat` | `Get-Item`/`Get-ChildItem` and the resulting file-system properties |
| `tee` | `Tee-Object` |
| `test` | PowerShell conditionals and `Test-Path` |
| `touch` | `New-Item -Force` or `Set-Item` |
| `wc` | `Measure-Object` |
| `which` | `Get-Command` |
| `xargs` | `ForEach-Object` or a typed PowerShell pipeline |

Gate scripts belong in `.ps1` files and are invoked explicitly with
`pwsh -NoProfile -File <path>`. Quoted data is not a command; for example,
`Select-String -Pattern 'grep' -Path a.txt` is portable. The lint is the
enforcement point: it also scans every `.ps1` file in a task's gate folder.

Install PowerShell on every mixed-team machine that does not already provide
`pwsh`. On macOS, use `brew install powershell/tap/powershell`. On Linux, use
Microsoft's PowerShell package for the distribution and ensure `pwsh` is on
`PATH`; do not substitute a distribution's `powershell` compatibility command.

### Copilot model ids

A Copilot session's `model.<kind>` takes the exact id Copilot's `--model` takes.
The Copilot CLI does not list them, and an unknown id fails with only
`Model "<id>" from --model flag is not available.` List the ids available to
your Copilot account with:

```console
vloop models copilot
```

The first line is the account's Copilot `plan` (`plan unknown` when the plan
could not be read). The remaining lines are picker-enabled models, sorted by
id; pass `--all` to include models hidden from Copilot's model picker. The
second column is your account's policy for the model: `enabled`, `disabled`
(the model exists but the policy blocks it), or `-` (the response has no
policy; such a model can still run).

When a session kind resolves to `agent.<kind> = "copilot"`, `vloop config set`
validates a new model id before writing it, and `vloop doctor` checks every
resolved Copilot model and reports the account plan. Aliases such as `sonnet`,
missing ids, and disabled ids are refused; an unavailable model-list request
is warning-only for config writes and doctor.

Set the same Copilot configuration for several session kinds atomically:

```console
vloop config set --kinds work,review --agent copilot --model mai-code-2
```

The list carries no prices: cost is estimated from the credits each session
uses (`copilot.credit-usd`). The ids
vloop has qualified for each session kind are in the README; how they were
found is in `docs/design-notes/copilot-cli-experiments.md`.

### The fence hook

Copilot does not read the plan, review, work, verify or gate-review fence, so each Copilot session has its own home (`COPILOT_HOME`, under `.vloop/tmp/`, emptied between sessions) holding a hooks file whose `preToolUse` entry runs `vloop fence hook --phase <review|second-review|work|verify|gate-review|plan> --root <repository root>` before every tool call. The hook reads one payload on stdin and prints one line, `{"permissionDecision":"allow"}` or `{"permissionDecision":"deny","permissionDecisionReason":"vloop fence: …"}`. It reads the phase's fence (`fence/review.json`, `fence/work.json` for a work session, `fence/verify.json` for a verify session, or `fence/gate-review.json` for a gate-review session, or `fence/plan.json` for a plan session): its own Bash deny rules and the adapter's deny-list stems, not a copy, with the fence's matching (`pre:*` is the prefix alone or followed by a space, `*` elsewhere any run of characters).

`vloop fence check --phase <review|second-review|work|verify|gate-review|plan> --root <dir> --tool <name> [--path <p>] ["<command>"]` runs the same judge without running anything: it prints `allow` (exit 0) or `deny: <reason>` (exit 1); bad arguments are exit 2. Use it to see what the hook would say to a command.

What the judge does:

- A shell command is split at `;`, `&&`, `||`, `|`, `&` and newlines outside quotes and looked through `env`, `command`, `exec`, `nohup`, `time`, `nice`, `timeout`, `stdbuf`, `xargs` and `sh -c` (also `bash`, `zsh`, `dash`, `ksh`); every simple command is matched.
- For PowerShell calls, literal `pwsh and powershell -Command` strings and `cmd /c` strings are judged like `sh -c`; other wrapper forms fail closed. `pwsh`'s eval-like commands (`Invoke-Expression`, `Invoke-Command`, `Start-Process`, `Start-Job`, `Start-ThreadJob`, `Invoke-Item`, the alias cmdlets `Set-Alias`, `New-Alias` and `Import-Alias`, and their aliases) and unquoted script blocks are refused, and so are `start` and `call` in a `cmd` string, with script blocks directed to `Where-Object <Property> -like <value>`, `ForEach-Object -MemberName <Property>` or `Select-Object -ExpandProperty <Property>`. This is not a full PowerShell parser: only the bounded `(<command>).<Property>` read and whole-statement `$env:<NAME> = <literal>` assignment are additionally read; unsupported parentheses, `$(…)` and backticks say to pipe into `Measure-Object` or `Select-Object -ExpandProperty`, run the inner command on its own, or write the command on one line, without escapes.
- **It fails closed.** It denies what it cannot parse: a backtick, `$(`, `<(`, `>(`, `eval`, `source`, `.` as a command, an unterminated quote, a `-c` string that is missing or not literal, a command word that starts with `$`, wrappers nested more than eight deep. A payload it cannot read, an unknown phase or any internal error is a deny line and exit 2; it never allows on an error. A tool it does not list (`web_fetch`, any `*_subagent`, `task`, `write_powershell`, `write_bash`, and the rest) is denied. Copilot's `read_powershell`, `list_powershell` and `stop_powershell` (and the `_bash` ones) are allowed in every phase: they read, list or stop a command the hook already judged when it started, and Copilot's runtime calls `read_powershell` on its own, outside the hook, for a command that runs in the background.
- **In a plan session the hook judges against the plan fence** (`fence/plan.json`): writes (`create`, `edit`, `apply_patch`, a redirection) are allowed only to `.vloop/state/state.json`, under `.vloop/state/gates/` and under `.vloop/tmp/`; any other write is denied with `write to <path> is not allowed in a plan session: only .vloop/state/state.json, .vloop/state/gates/ and .vloop/tmp/ may be written`. `/dev/null` is allowed. The driver also reverts a file a plan or gate-review session changed outside `.vloop/state/` and `.vloop/tmp/`, with a warning (`vloop: the plan session changed <path> — reverted`), so it never reaches the plan commit.
- In a review session, in a verify session judged against the verify fence (`fence/verify.json`), and in a gate-review session judged against the gate-review fence (`fence/gate-review.json`), writes (`create`, `edit`, `apply_patch`, a redirection) are allowed only under `.vloop/tmp/`; `/dev/null` is allowed. Reads may not touch `~/.claude`, `~/.copilot` or `~/.config/gh`.
- **In a work session the hook judges against the work fence** (`fence/work.json`): writes are allowed anywhere under the repository root except `.git/` and `.vloop/`, where only `.vloop/tmp/` may be written, and a write outside the repository is denied. `rm` of a file is allowed; the recursive forms stay denied. The judge reads `if`, `while`, `until`, `for NAME in …`, `!`, `{ … }` and subshells, and judges every command inside them. Shell syntax it cannot read is denied, not guessed at: `select`, `function`, `coproc` and `[[` answer `shell syntax [[ cannot be judged` (the word at hand); `$(…)` and a backtick, `command substitution cannot be judged`; `case`, `<(…)`, `>(…)` and `((…))`, `subshells and process substitution cannot be judged`; a command name that expands, an expanded word before the command name (a wrapper's option, or a shell's before its `-c` string) and a redirection target holding a glob or a brace expansion, `… expands … and cannot be judged` or `… cannot be resolved`. A word the shell expands (a `$`, a glob, a brace expansion) may be any text: `git $X origin` and `git pu{sh,} origin` are denied wherever `git push` is. Copilot still runs a call when the hook prints something that is not JSON or exceeds its timeout; the hook always prints its answer and is fast, and the deny list stays behind it (`docs/design-notes/copilot-cli-experiments.md`, B22.2 and B22.3). The driver backs it: after every Copilot session it judges again every call that ran, and halts with exit 9 (FENCE BYPASSED) when the fence denies one, or (HOOK CHANGED) when the hook's binary changed.
- **K1: a Copilot session uses vloop's skills only.** The hook allows the `skill` tool only for the name of one of vloop's skills, and denies any other with `vloop fence: skill <name> is not vloop's — use /vloop:review` (`/vloop:work` in a work session, `/vloop:verify` in a verify session, `/vloop:gate-review` in a gate-review session, `/vloop:plan` in a plan session). `vloop run` refuses to start (exit 1), and `vloop doctor` shows an error, while a skill directory under `.claude/skills/` or `.github/skills/` has exactly the name of one of vloop's skills and a Copilot agent is in use; rename it. The warning for other names ending in `review` stays.
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

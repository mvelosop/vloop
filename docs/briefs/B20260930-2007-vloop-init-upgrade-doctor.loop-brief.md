---
name: B20260930-2007-vloop-init-upgrade-doctor.loop-brief
description: Put vloop into a consumer repo and keep it healthy — init with detected stacks and a starter brief, a deliberate upgrade that refuses breaking and downgrading jumps, doctor, the embedded plugin extracted on demand, a valid marketplace, and the plugin-to-binary version handshake
kind: brief
status: consumed
created: 2026-09-30
seeds: A `.loop/run.sh` plan + run that builds slice B5 in docs/design-notes/vloop-roadmap.md
depends-on: [B20260930-0929-vloop-close-export-workspace.loop-brief]
---
# Brief — vloop B5: init, upgrade, doctor and the plugin

- **Status:** consumed — closed 2026-09-30 as run B20260930-2007-vloop-init-upgrade-doctor. **Do not re-plan from this brief.**
- **Starting point:** extends `main` at the commit that adds this brief. The
  planner pins that SHA as the base every gate compares against.
- **Produced by:** the design act on the B5 row of
  `docs/design-notes/vloop-roadmap.md`, 2026-09-30, with the operator.

---

## What it is

The fifth slice of `vloop`: everything a consumer repository needs before its
first run. `vloop init` sets a repository up — config with its language and
detected stacks, a starter brief, the `.gitignore` line, a vloop section in
`CLAUDE.md`, and a stamp of the vloop version that did it. `vloop upgrade`
refreshes what vloop owns there, deliberately: it shows what it would change and
refuses downgrades and breaking jumps unless told. `vloop doctor` says whether
the repository, the toolchain and the plugin are ready. The plugin becomes a
valid, installable marketplace entry with a session-start check that its version
matches the binary's, and the binary can extract its own copy for the driver.

## Why this shape, and what was rejected

Decided with the operator on 2026-09-30. **Do not re-litigate these.**

- **Skills are not in this slice.** `/vloop:plan`, `/vloop:work`,
  `/vloop:review` and `/vloop:operate` move to B6, where `vloop run` invokes and
  exercises them.
  *Rejected:* writing them here — they could not run before B6's driver exists
  and would ship untested.
- **`upgrade` is separate from `init`.** `init` refuses on a repository vloop
  already set up; only `upgrade` changes one, and it refuses a breaking jump
  without `--yes`.
  *Rejected:* an idempotent `init` that upgrades — re-running it could silently
  apply a breaking release.
- **`init` writes a starter brief** in the repository's language, beside the
  config.
- **The version handshake is a Go command run by the plugin's hook**, so it
  works on Windows with no shell script, and it warns — it never blocks a
  session.
  *Rejected:* a doctor-only check — a mismatch would go unnoticed until someone
  ran doctor.
- **Nothing vloop-owned lives in `.claude/settings.json`**, and `init` never
  writes it. The loop's permission fence ships inside the binary with the
  plugin (B6 hands both to sessions).
- **Monorepos get stacks by path.** A stack may be scoped to a subtree —
  `csharp@services/api` in the same `metrics.stacks` list — and inside a scoped
  subtree only its own stacks apply; the longest matching scope wins. `init`
  detects stacks per directory and writes scopes.
  *Rejected:* TOML scope tables (the CLI's list keys would need a new shape);
  scopes that add to the unscoped stacks (they would bring back the cross-stack
  misclassification scoping exists to remove).
- **A whole-package quality and security review is its own brief (B7)**, after
  B6, when the package is whole. This brief's review rules on its own security
  surface (see below).

## Binding references

- `docs/domain/domain-model.md` — C-1, C-2, C-3, R-2 and S-2: config is
  repo-local, the repo root, every JSON document names its schema, runs happen
  on work branches, sessions never move refs.
- `docs/domain/platform/platform-context.md` — the configuration keys and what
  each decides, the schemas list this slice extends, the CLI conventions.
- `docs/briefs/B20260929-1804-vloop-skeleton-config-briefs.loop-brief.md` — the
  embedded plugin and its manifests, `version`, the brief templates.
- `docs/additional-context-files/20260929-0951-vloop-design-session.md` — sections 2 and 6:
  how plugin and binary are tied (`--plugin-dir`, marketplace install, the
  mismatch hook), and a binary built from the tree it is about to change.
- `.loop/install.sh` — what the shell loop's installer writes into a consumer
  and what it leaves alone: the `CLAUDE.md` section between markers, the
  `.gitignore` line, the version stamp, never `.claude/settings.json`.
- `.loop/run.sh` — `preflight`: the checks `doctor` ports (tools on `PATH`,
  workspace trust, git identity, a valid plan).

## Behaviour contract

B1–B4's conventions hold: global flags, `--json` (one document on stdout, on
exit 0 and 1), exit codes `0` ok / `1` problems or failure / `2` usage, errors
as one stderr line starting `vloop: `, paths repo-relative with `/`.

### The install stamp — .vloop/install.json, schema `install/v1`

```
{"schema":"install/v1","version":"0.1.0","commit":"abc1234",
 "initialized":"2026-10-01T09:00:00Z","upgraded":null}
```

Written by `init`, rewritten by `upgrade`, read by `upgrade` and `doctor`. A
repository is **initialised** when this file exists. `install/v1` joins
`vloop schema list`.

### `vloop init [--language en|es] [--stacks <a,b>] [--dry-run]`

**Refusals**, each writing nothing:

| Condition | Message | Exit |
| --- | --- | --- |
| the repo root has no `.git` | `vloop: not a git repository — run git init first` | 1 |
| .vloop/install.json exists | `vloop: already initialized by vloop <version> — run vloop upgrade` | 1 |
| an invalid `--language` or stack | B1's invalid-value message | 2 |

**What it writes**, then prints one line per path (`wrote <path>`,
`updated <path>`, `kept <path>`) and the next steps:

1. `.vloop/config.toml`: `language` (default `en`) and `metrics.stacks`
   (`--stacks`, else detected, below). **An existing config file is kept
   byte for byte** (`kept .vloop/config.toml`); detected stacks it lacks are
   printed as a suggestion, not written.
2. .vloop/install.json (above), with this binary's version and commit.
3. `docs/briefs/`, and in it a starter brief —
   `B<YYYYMMDD-HHMM>-getting-started.loop-brief.md` (`primeros-pasos` for
   `es`) — exactly as `vloop brief new` writes it, `status: draft`. Skipped
   (`kept docs/briefs/`) when the directory already holds a loop brief.
4. `.gitignore`: the line .vloop/tmp/, appended once; the file is created if
   absent. Nothing else in it changes.
5. `CLAUDE.md`: vloop's section between `<!-- vloop:begin -->` and
   `<!-- vloop:end -->`. If the markers exist, only what is between them is
   replaced; if the file exists without them, the section is appended; if the
   file is absent, it is created with the section alone. The section is
   embedded product text, in English: the rules every loop session follows
   (repo-relative paths; one task per work session; sessions never commit, set a
   status or move git refs; a task is done only when its gate passes and review
   passes it; halting cleanly with an account of what blocked is a success) and
   one line pointing an interactive session at the operator's role.

It never writes `.claude/settings.json`, never touches `.loop/`, and never
commits. Next steps printed last:

```
next:
  claude plugin marketplace add mvelosop/vloop
  claude plugin install vloop@vloop
  vloop doctor
```

`--dry-run` prints `would write <path>` / `would update <path>` / `would keep
<path>` and writes nothing, still applying every refusal.

**Stack detection** looks at the repo root and directories up to three levels
below it, skipping `.git`, `.vloop`, `.loop`, `node_modules`, `vendor`, `bin`,
`obj`, `dist`, `target` and `build`. A directory holding a marker file is a
**stack root** for the stacks the marker implies:

| Found | Stacks |
| --- | --- |
| `go.mod` | `go` |
| package.json | `javascript`; plus `typescript` if a tsconfig.json sits beside it or `typescript` is a dependency; plus `react` if `react` is a dependency |
| `*.csproj` or `*.sln` | `csharp` |
| pyproject.toml, `requirements.txt` or setup.py | `python` |
| `pom.xml` or `build.gradle` | `java` |
| `build.gradle.kts` | `kotlin` |
| Cargo.toml | `rust` |

Stacks found at the repo root are written unscoped (`go`); stacks found in a
subdirectory are written scoped to it (`csharp@services/api`). For C#, the stack
root is the directory of the `*.sln` if there is one above the `*.csproj`
files, else each `*.csproj`'s directory. The list is written sorted, unscoped
first: `detected stacks: go, csharp@services/api, javascript@web, react@web`,
or `detected stacks: none`.

### `vloop upgrade [--dry-run] [--yes]`

Compares the stamp's version (`from`) with this binary's (`to`), both read as
semantic versions.

| Case | Behaviour | Exit |
| --- | --- | --- |
| no stamp | `vloop: not initialized — run vloop init` | 1 |
| `from` = `to` | `already at <to>`, nothing written | 0 |
| `from` > `to` | `vloop: this repository was set up by vloop <from>, newer than this binary (<to>) — upgrade the binary, not the repository` | 1 |
| a **breaking** jump without `--yes` | the changes it would make, then `vloop: <from> → <to> may break this repository's setup — review the changes above, then run vloop upgrade --yes` | 1 |
| otherwise, or with `--yes` | applies the changes, prints them, rewrites the stamp (`version`, `commit`, `upgraded`) | 0 |

- **Breaking** means the major versions differ, or both majors are `0` and the
  minors differ. A version with a pre-release suffix (`0.0.0-dev`) cannot be
  compared: any upgrade to or from one needs `--yes`.
- **The changes** are exactly the vloop-owned parts: the `CLAUDE.md` section
  between its markers and the `.gitignore` line. `upgrade` never touches config
  values, briefs, plans, defects, interventions or anything outside these.
- `--dry-run` prints `would update <path>` for each and writes nothing.

### `vloop doctor [--json]`

One line per check — `✓` pass, `!` warning, `✗` problem, `-` not applicable —
then `doctor: <n> problem(s), <m> warning(s)`. Exit 1 if any problem.

| Check | ✗ problem | ! warning |
| --- | --- | --- |
| git | not a git repository, or no `user.name`/`user.email` | — |
| install | the stamp is newer than this binary | not initialized; the stamp is older than this binary (`run vloop upgrade`) |
| config | `.vloop/config.toml` does not parse or has an invalid value | — |
| claude | `claude` not on `PATH` | — |
| trust | — | this repository's folder is not marked trusted in the user's Claude settings (`run claude here once and accept the trust dialog`) |
| gate shell | a plan exists and its `shell` is not on `PATH` | no plan, and the configured `shell` is not on `PATH` |
| plan | a plan exists and fails `task validate` | — |
| branch | a plan is `running` or `planning` and HEAD is on the default branch (R-2) | — |
| plugin | — | `claude plugin list --json` shows no enabled `vloop@…`, or its version differs from this binary's |
| stacks | — | a scoped stack's path no longer exists (`csharp@services/api: no such directory`) |
| self-hosting | — | this is vloop's own repository and the binary's commit is HEAD — a released vloop should build the next one |

`--json`: `{"ok":bool,"checks":[{"name","result":"pass|warning|problem|n/a","message"}]}`.
`doctor` writes nothing. Reading the user's Claude settings for the trust check
is read-only and is the one place vloop looks outside the repository.

### The plugin

- `plugin/.claude-plugin/plugin.json` gains a `description` and an `author`;
  `.claude-plugin/marketplace.json` gains the required `owner` (with `name`)
  and a `description`. `claude plugin validate .` passes with no errors and no
  warnings. (The marketplace has failed validation since B1: the manifest was
  missing `owner`.)
- plugin/hooks/hooks.json declares one `SessionStart` hook whose command is
  `vloop version --check-plugin "${CLAUDE_PLUGIN_ROOT}"`.
- The embedded plugin includes the hooks.

### `vloop version --check-plugin <dir>`

Reads `<dir>/.claude-plugin/plugin.json`. Equal versions: prints nothing, exit
0. Different: prints `vloop: the vloop plugin is <p> but the vloop binary is
<b> — update the one that is behind` on stdout, exit 0. An unreadable manifest:
`vloop: cannot read the vloop plugin's version in <dir>` on stdout, exit 0. It
**always exits 0**: a hook that fails would interrupt every session.

### `vloop plugin path [--json]`

Extracts the embedded plugin into `.vloop/tmp/plugin/<version>/` under the repo
root and prints that path. Idempotent: when the directory already holds
exactly the embedded files, nothing is rewritten. Files the embedded plugin
does not contain are removed from that directory. B6's driver passes the path
to every session with `--plugin-dir`.

### F1 — a binding reference's reason may wrap

**Defect.** `vloop brief check` reads only the first line of a `## Binding
references` entry, so an entry whose reason starts on the next line — still one
Markdown list item — is reported as `binding reference has no reason`. Found
writing this brief.

**Required.** An entry is the list item: its first line and every following
line indented under it, up to the next item or heading. The reason is
everything after the separator, across those lines.

**Gate.** A fixture entry with the separator at the end of its first line and
the reason on the next passes; an entry with no reason on any line still fails.
The first must fail against the current code.

### F2 — stacks scoped by path (a change to B3's classification)

**Defect.** `metrics.stacks` is one repo-wide list and the presets merge across
the whole tree, so in a monorepo one stack's globs classify another's files:
Python's `**/tests/**` makes a TypeScript app's web/tests/helpers.ts a test;
C#'s excluded `**/bin/**` drops a Go service's tools/bin/ scripts; root-only
globs like `go.sum` never match a nested module's `services/x/go.sum`.

**Required.**

- An entry of `metrics.stacks` is `<stack>` or `<stack>@<path>`. The path is
  repo-relative, uses `/`, has no leading `/`, trailing `/`, `.` or `..`
  segment, and names an existing directory. `config set` rejects anything else
  with B1's invalid-value message (`want <stack> or <stack>@<existing directory>`),
  exit 2.
- A path's **scope** is the longest scoped path that contains it; with none, the
  unscoped stacks apply. Inside a scope, **only that scope's stacks** apply.
- A scoped preset's globs match the path **relative to the scope's directory**:
  `services/api/Api.Tests/CalcTests.cs` is tested against `**/*.Tests/**` as
  `Api.Tests/CalcTests.cs`, and `services/api/go.sum` against `go.sum`.
- The layers are otherwise unchanged (M-3): always-excluded first, the repo's own
  globs (repo-relative, unscoped) next, then the presets of the path's scope,
  else `other`.
- `metrics classify` labels a scoped match `(<stack>@<path>: <glob>)`; an
  unscoped match keeps its current label `(<stack>: <glob>)`.

**Gate.** The monorepo fixture below: every classification line matches, and at
least web/tests/helpers.ts and `tools/bin/run.sh` are classified differently
than the current code classifies them with the same stacks unscoped.

### Decided here, because the cited documents leave it underdetermined

- The stamp's name, format and schema; semver comparison and "breaking" for
  `0.x`.
- Pre-release versions always need `--yes` to upgrade across.
- The starter brief's slug per language, and that `init` skips it when a loop
  brief exists.
- The `CLAUDE.md` section is English whatever the language: it instructs
  sessions, and C-4 keeps vloop's own text English.
- The trust check reads the user's Claude settings read-only; it is a warning,
  never a problem.

### Violations the review must rule on — including security

- `init` or `upgrade` writing outside `.vloop/`, `docs/briefs/`, `.gitignore`
  and `CLAUDE.md`; changing anything in `CLAUDE.md` or `.gitignore` outside
  vloop's own markers or line; touching `.claude/settings.json` or `.loop/`;
  committing.
- Any command following a symlink out of the repository when writing, or
  writing to a path built from unvalidated input (a stack name, a language, a
  plugin directory) without cleaning it.
- `version --check-plugin` executing, sourcing or evaluating anything from the
  plugin directory, or exiting non-zero.
- `doctor` writing anything, or running any command other than `git`, `claude
  --version`, `claude plugin list --json` and a `PATH` lookup.
- `plugin path` writing outside `.vloop/tmp/plugin/`.
- Skill content, `run`, or the B7 review appearing here.

## Worked example

A scratch git repository with `go.mod`, a package.json whose dependencies
include `react`, and no `.vloop/`; a stub `claude` on `PATH` answering
`--version` and `plugin list --json`; a fake home with the Claude settings file
marking the scratch folder trusted. Binaries built with `-ldflags -X` at the
versions shown. `<stamp>` and dates are threaded from the output.

```
$ vloop init --language es                                  (binary 0.1.0)
detected stacks: go, javascript, react
wrote .vloop/config.toml
wrote .vloop/install.json
wrote docs/briefs/B<stamp>-primeros-pasos.loop-brief.md
updated .gitignore
wrote CLAUDE.md
next:
  claude plugin marketplace add mvelosop/vloop
  claude plugin install vloop@vloop
  vloop doctor                                               exit 0
$ vloop config get metrics.stacks
go,javascript,react                                          exit 0
$ vloop init
vloop: already initialized by vloop 0.1.0 — run vloop upgrade
                                                             exit 1
$ vloop doctor                                               (stub reports vloop@vloop 0.1.0)
  ✓ git
  ✓ install
  ✓ config
  ✓ claude
  ✓ trust
  ✓ gate shell
  - plan
  - branch
  ✓ plugin
  - self-hosting
doctor: 0 problem(s), 0 warning(s)                           exit 0

$ vloop upgrade                                              (binary 0.1.1)
updated CLAUDE.md
updated .vloop/install.json                                  exit 0
$ vloop upgrade                                              (binary 0.2.0)
would update CLAUDE.md
vloop: 0.1.1 → 0.2.0 may break this repository's setup — review the changes above, then run vloop upgrade --yes
                                                             exit 1
$ vloop upgrade --yes                                        (binary 0.2.0)  exit 0
$ vloop upgrade                                              (binary 0.1.1)
vloop: this repository was set up by vloop 0.2.0, newer than this binary (0.1.1) — upgrade the binary, not the repository
                                                             exit 1

$ vloop plugin path
.vloop/tmp/plugin/0.2.0                                      exit 0
$ vloop version --check-plugin .vloop/tmp/plugin/0.2.0      (binary 0.2.0)
                                                             exit 0, no output
$ vloop version --check-plugin .vloop/tmp/plugin/0.2.0      (binary 0.1.1)
vloop: the vloop plugin is 0.2.0 but the vloop binary is 0.1.1 — update the one that is behind
                                                             exit 0
```

A **monorepo** fixture: `go.mod` at the root; `services/api/Api.sln`,
`services/api/Api/Api.csproj` and `services/api/Api.Tests/Api.Tests.csproj`;
web/package.json with `react` as a dependency; `tools/scripts/requirements.txt`.

```
$ vloop init
detected stacks: go, csharp@services/api, javascript@web, python@tools/scripts, react@web
  …                                                          exit 0
$ vloop metrics classify services/api/Api.Tests/CalcTests.cs services/api/obj/Api.cs web/src/App.test.tsx web/tests/helpers.ts tools/bin/run.sh go.sum
services/api/Api.Tests/CalcTests.cs  test  (csharp@services/api: **/*.Tests/**)
services/api/obj/Api.cs  excluded  (csharp@services/api: **/obj/**)
web/src/App.test.tsx  test  (react@web: **/*.test.tsx)
web/tests/helpers.ts  other  (none)
tools/bin/run.sh  other  (none)
go.sum  excluded  (go: go.sum)                               exit 0
$ vloop config set metrics.stacks go,csharp@services/missing
vloop: invalid value "csharp@services/missing" for metrics.stacks: want <stack> or <stack>@<existing directory>
                                                             exit 2
```

The failures, each planted in a copy of the fixture:

```
init in a directory without .git
  -> vloop: not a git repository — run git init first               exit 1
a CLAUDE.md with the operator's own text above and below existing vloop markers, then upgrade
  -> only the text between the markers changes, byte for byte        exit 0
an existing .vloop/config.toml with language en, then init
  -> kept .vloop/config.toml, the file unchanged; no stamp existed, so init proceeds
                                                                     exit 0
docs/briefs/ already holding a loop brief
  -> kept docs/briefs/, no starter brief                             exit 0
the stub's plugin list shows vloop@vloop 0.1.0 against binary 0.2.0
  -> ! plugin … (warning), exit 0
no user.email configured
  -> ✗ git …, doctor: 1 problem(s)                                   exit 1
a plugin directory whose plugin.json is not JSON
  -> vloop: cannot read the vloop plugin's version in <dir>          exit 0
a file planted in .vloop/tmp/plugin/0.2.0/ that the plugin does not contain, then plugin path
  -> the file is removed                                             exit 0
```

**Real data.** On this repository, read-only: `vloop init --dry-run` detects
`go`, reports `would keep .vloop/config.toml`, and writes nothing; `vloop doctor`
exits 0 with `! install` (this repository was never initialised by vloop); and
`claude plugin validate .`, run by the gate — not by a session, which the fence
denies `claude` — passes with no errors and no warnings.

## Out of scope

The roadmap's later rows:

- The skills `/vloop:plan`, `/vloop:work`, `/vloop:review`, `/vloop:operate`,
  and anything else under `plugin/skills/` (B6).
- `run`, the fence as embedded data, passing `--plugin-dir` to sessions (B6).
- The whole-package quality and security review (B7).

Also not in this run:

- Installing or updating the vloop binary itself: package managers, Homebrew,
  Scoop, goreleaser, the `vl` alias.
- `doctor --fix`, or any command that changes the user's Claude settings.
- Publishing the marketplace (pushing is the operator's), `claude plugin eval`
  suites.
- Migrating config values between versions; `upgrade` refreshes only vloop's
  own section and line.
- Initialising this repository itself — the operator decides that after the run.
- Anything else per path: gate shells, languages, `areas` mapped to paths.

## Constraints

- Go as pinned in `go.mod`. Third-party modules: B4's; anything else needs its
  reason in the task notes. No cgo. vloop never uses the network.
- Must build for `darwin`, `linux` and `windows`; nothing may assume `/` on
  disk or a case-sensitive filesystem. The hook must work on Windows: it is a
  vloop command, not a script.
- **Gates run on macOS with its BSD tools**: no GNU-only syntax (no empty
  alternatives in `grep -E`, no `sed -i` without a suffix, no `\+` or `\|` in
  basic regexes, no `date -d`).
- **Sessions must not move git refs, and cannot run `claude`**: the fence denies
  both. Tests put a stub `claude` on `PATH`; only the close task's gate runs the
  real `claude plugin validate .`. Tests build fixture repositories and fake
  homes in temporary directories and address them with `git -C`; a session
  checks a gate by running the gate, never by pasting its commands.
- **Gates follow the diff surface**: each task gates on the packages it touches,
  plus `go vet ./...`, `gofmt -l .` and `go mod tidy -diff` printing nothing,
  and the linux, windows and host builds. Only the close runs `go test ./...`.
- A gate runs in under a minute. Tests never read or write this repository's
  `.loop/`, `.claude/`, `.vloop/`, git history or the real home directory; only
  the close task's real-data check reads this repository, read-only.
- **No task may weaken a check to pass.** One expectation must change, and
  changing it is required: `schema list` gains `install/v1`. The generated
  command reference (`docs/guide/commands.md`) is regenerated, not hand-edited,
  and B1's README test must pass with the new commands in the README.
- **New checks are proven by fixtures**: every refusal of `init` and `upgrade`,
  every doctor check in each of its results, and each stack detection rule has a
  fixture where it applies and one where it must not.
- Repo-relative paths everywhere. No absolute paths in any file or commit message.

## Shape

9 to 11 tasks, each independently verifiable by a single command. Order is
load-bearing:

1. **`install/v1`** in the schemas, the stamp's read and write, semantic-version
   comparison and "breaking", with the `schema list` update.
2. **The plugin**: manifests completed, the hook, `version --check-plugin`,
   `plugin path`; a Go test that the embedded tree holds the hook.
3. **F2, stacks scoped by path**: parsing and validating `stack@path`, scope
   resolution, scope-relative matching, the `classify` label; the monorepo
   fixture, gated on classifying differently than the current code.
4. **Stack detection**, per directory, with a fixture per rule, the C# solution
   rule and the skipped directories.
5. **`vloop init`**: every write and refusal, the `CLAUDE.md` markers, the
   `.gitignore` line, the starter brief, `--dry-run`.
6. **`vloop upgrade`**: every row of the table, marker-only changes, `--dry-run`.
7. **`vloop doctor`**: every check in each result, with a stub `claude` and a
   fake home.
8. **Docs**: the README's new commands, the regenerated command reference,
   `docs/guide/configuration.md` gaining the stamp and scoped stacks,
   `docs/guide/metrics.md` gaining scopes in classification,
   `docs/guide/concepts.md` gaining "setting up a repository", and
   `docs/domain/domain-model.md`'s M-3 and `docs/domain/measurement/metrics.md`'s
   classification section stating scopes as built.
9. **Close**: an end-to-end test that builds binaries at the versions shown and
   plays the worked example line for line; the real-data check, including
   `claude plugin validate .`; `go test ./...`, `go vet ./...`, `gofmt -l .`,
   `go mod tidy -diff` and the three builds.

The worked example needs its own gate (task 9): only built binaries at
different versions show `upgrade`'s refusals, and only the real `claude` shows
the marketplace is valid.
<!-- vloop:run-record:begin -->
## Run record

Generated by vloop brief close on 2026-09-30. Recompute with: vloop metrics B20260930-2007-vloop-init-upgrade-doctor.loop-brief

```
B20260930-2007-vloop-init-upgrade-doctor  consumed · not merged
 tasks     10 planned (brief said 9–11) · 10 done · 0 blocked · first-pass 8/10
 size      delivered  code 1,327 · test 1,379 · docs 108 · test:code 1.04
           churn      code 1,327 · test 1,379 · docs 110 · rework 1.00
 time      agent 23.8 min (work 20.1 · review 3.7) · plan 20.4 min · gates n/a · wall 32.4 min
 rate      55.7 code lines/min · 113.5 incl. tests
 cost      $11.19 · plan 5.88 · work 4.04 · review 1.27 · $8.43 per 1,000 code lines
 models    plan claude-opus-5-5 · work claude-sonnet-5-5 · review claude-sonnet-5-5
 defects   in-loop 2 · operator 0 · escaped 0 · removal efficiency 100%
```

```
id   area  kind  att  code+  test+  docs+  other+  agent  cost   model
T1   -     -     1    137    99     0      0       1m56s  $0.37  claude-sonnet-5-5
T2   -     -     1    27     47     0      0       1m11s  $0.28  claude-sonnet-5-5
T3   -     -     1    190    117    16     16      2m26s  $0.53  claude-sonnet-5-5
T4   -     -     1    90     88     0      0       2m15s  $0.42  claude-sonnet-5-5
T5   -     -     1    134    101    0      0       1m05s  $0.31  claude-sonnet-5-5
T6   -     -     3    330    151    10     0       3m51s  $1.01  claude-sonnet-5-5
T7   -     -     1    94     131    9      0       1m52s  $0.42  claude-sonnet-5-5
T8   -     -     3    325    307    6      0       4m45s  $1.06  claude-sonnet-5-5
T9   -     -     1    0      13     69     0       1m59s  $0.39  claude-sonnet-5-5
T10  -     -     1    0      325    0      0       2m30s  $0.51  claude-sonnet-5-5
```

- D20260930-2149-t6-gate-required-an-empty-status-while-i — gate, plan, found by gate: T6 gate required an empty status while init must modify tracked .gitignore
- D20260930-2159-t8-gate-appended-a-top-level-key-after-a — gate, plan, found by gate: T8 gate appended a top-level key after a TOML table
<!-- vloop:run-record:end -->

### Operator notes

Written by hand, outside the generated markers.

**Three runs.** The run stalled twice (exit 3), each time on a gate the planner
wrote without executing its fixture edits: T6's gate asserted an empty
`git status` after requiring `init` to modify a committed `.gitignore`; T8's
appended `shell = "pwsh"` after a TOML table, where it became `metrics.shell`.
The work sessions diagnosed both exactly; the operator applied their fixes with
`.loop/amend.sh verify`, reset the tasks and resumed. Both are recorded as
`plan`/`gate` defects (fixed) and as interventions.

**Verified outside the loop.** `go test -count=1 ./...` (15 packages), `go
vet`, `gofmt -l .`, `go mod tidy -diff`, both cross-builds; B1's README test
unchanged. `claude plugin validate .` passes with no errors or warnings. The
worked example by hand with binaries at 0.1.0, 0.1.1 and 0.2.0 on a monorepo
fixture: `init`'s per-directory detection and six scoped classifications,
`upgrade`'s four cases with the operator's own `CLAUDE.md` text preserved,
`plugin path` removing a planted file, `version --check-plugin` always exiting
0, `doctor` clean (exit 0) and with a newer stamp (exit 1). Real data: `init
--dry-run` on this repository keeps its config and writes nothing; `doctor`
exits 0 with the two expected warnings. F1 checked against the pre-B5 binary:
a wrapped reason fails there and passes here.

**Changed by hand.** The two gates above; B1's escaped defects (marketplace
`owner`, wrapped binding reasons) marked fixed by this brief.

**Still open.**

- The planner writes gates it never runs. Three of this series' gate defects
  (B2 T1, B5 T6, B5 T8) are that; B6's planner skill should execute each gate
  against the base commit — it must fail — before handing over the plan.
- The README is at 199 of its 200-line cap; B6 adds `run` and the skills.
- Initialising this repository with `vloop init` is the operator's call, best
  after B6 replaces `.loop/`.


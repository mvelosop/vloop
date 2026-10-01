# Configuration guide

vloop's settings are per repository, in `.vloop/config.toml` under the repo root.
Nothing is read from your home directory. An environment variable overrides the
file, and the file overrides the default. `vloop config list` shows every key
with its value and where it came from; `vloop config set <key> <value>` writes
one (`''` removes it); `vloop config get <key>` prints one.

An invalid value is exit 2 from `config set`, and exit 1 when it comes from the
environment or the file.

## Keys

The variable of a key is `VLOOP_` and the key in upper case with `.` turned into
`_`. List keys are TOML arrays in the file and comma-joined text in the
environment and in `config set`.

| Key | Default | Valid values | Environment variable | Effect |
| --- | --- | --- | --- | --- |
| `language` | `en` | `en`, `es` | `VLOOP_LANGUAGE` | language of a brief's section headings and of the template `brief new` writes; commands, flags, keys and JSON stay English |
| `model.plan` | `opus` | any model name | `VLOOP_MODEL_PLAN` | model of the plan session |
| `model.work` | `sonnet` | any model name | `VLOOP_MODEL_WORK` | model of the work sessions |
| `model.review` | `sonnet` | any model name | `VLOOP_MODEL_REVIEW` | model of the review sessions |
| `effort.plan` | unset | `low`, `medium`, `high`, `xhigh`, `max` | `VLOOP_EFFORT_PLAN` | effort of the plan session |
| `effort.work` | unset | same | `VLOOP_EFFORT_WORK` | effort of the work sessions |
| `effort.review` | unset | same | `VLOOP_EFFORT_REVIEW` | effort of the review sessions |
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

`model.*` and `effort.*` are stored under `[model]` and `[effort]`, the
`metrics.*` keys under `[metrics]`, the `run.*` budgets under `[run]` as TOML numbers. For example:

```
language = "es"
areas = ["cli", "docs"]

[model]
work = "opus"

[metrics]
stacks = ["go"]
code = ["internal/brief/templates/**"]
```

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
| `javascript` | `**/*.js`, `**/*.mjs`, `**/*.cjs` | `**/*.test.js`, `**/*.spec.js`, `**/__tests__/**` | `package-lock.json`, `yarn.lock`, `pnpm-lock.yaml`, `**/dist/**`, `**/node_modules/**` |
| `kotlin` | `**/*.kt`, `**/*.kts` | `**/src/test/**` | `**/build/**`, `**/target/**` |
| `python` | `**/*.py` | `**/test_*.py`, `**/*_test.py`, `**/tests/**` | `**/__pycache__/**`, `poetry.lock`, `uv.lock`, `Pipfile.lock` |
| `react` | `**/*.tsx`, `**/*.jsx`, `**/*.css`, `**/*.scss` | `**/*.test.tsx`, `**/*.spec.tsx`, `**/*.test.jsx`, `**/*.spec.jsx`, `**/*.stories.*` | `package-lock.json`, `yarn.lock`, `pnpm-lock.yaml`, `**/dist/**`, `**/node_modules/**` |
| `rust` | `**/*.rs` | `**/tests/**`, `**/benches/**` | `**/target/**`, `Cargo.lock` |
| `typescript` | `**/*.ts` | `**/*.test.ts`, `**/*.spec.ts`, `**/__tests__/**` | `package-lock.json`, `yarn.lock`, `pnpm-lock.yaml`, `**/dist/**`, `**/node_modules/**` |

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

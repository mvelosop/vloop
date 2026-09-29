---
name: B20260929-1804-vloop-skeleton-config-briefs.loop-brief
description: Build the vloop Go CLI skeleton — version, repo-local config with language and per-kind model/effort, and the bilingual brief format with check, new, list and depends-on
kind: brief
status: draft
created: 2026-09-29
seeds: A `.loop/run.sh` plan + run that builds the first slice of vloop (B1 in docs/design-notes/vloop-roadmap.md)
---
# Brief — vloop B1: CLI skeleton, config, and the bilingual brief format

- **Status:** ready to plan
- **Starting point:** greenfield Go code; extends `main` at the commit that adds
  this brief. The planner pins that SHA as the base every gate compares against.
- **Produced by:** the design act on
  `docs/briefs/B20260929-1222-initial-setup-for-vloop-cli.architect-brief.md`,
  2026-09-29.

---

## What it is

The first slice of `vloop`, a Go CLI that will replace the shell loop in
`.loop/`. At the end of this run a `vloop` binary builds for macOS, Linux and
Windows; it reports its version and the version of the plugin it embeds; it
reads and writes a repo-local config that sets the project's language (`en` or
`es`) and the model and effort for each session kind; and it creates, checks and
orders briefs written in either language, including the dependencies between
them; and a README explains the core concepts to a first-time reader. It is for the operator authoring briefs now, and it is the interface every
later slice builds on.

## Why this shape, and what was rejected

Decided with the operator on 2026-09-29. **Do not re-litigate these** — the
planner inherits them.

- **This run is slice B1 only.** The series is in the roadmap note; `task`,
  `status`, `init`, `doctor` and `run` are later briefs.
  *Rejected:* the whole CLI and plugin in one brief — 25+ tasks, and the driver
  port depends on the rest being settled first.
- **vloop's files live under .vloop/**, not `.loop/`. This repo's `.loop/` is
  the shell loop building vloop, and nothing vloop does — including its tests —
  may read or write it.
  *Rejected:* `.loop/` as in the design session — it collides with the bootstrap
  loop.
- **Config is repo-local only**: .vloop/config.toml, overridden by `VLOOP_*`
  environment variables. Never `~/.config` or any global path.
- **Spanish covers briefs and prose, not the CLI.** One `language` per repo
  selects the heading set a brief must use and the template `brief new` writes.
  Frontmatter keys and values, config keys, JSON output, flag names, and vloop's
  own messages stay English.
  *Rejected:* localizing CLI output (an i18n catalog in every command); accepting
  either language in any repo (a brief's language would then be unverifiable).
- **Model and effort are configured per session kind** (`plan`, `work`,
  `review`) in this run. Per-task overrides are decided too, but they live in the
  state schema, which is B2 — do not build them here.
  *Rejected:* a single model for work and review, which is what the shell loop
  does today.
- **No loop-knowledge.** A brief's `## Binding references` section is the only
  way a document binds a task. vloop has no knowledge-roots file and no scan.
- **Briefs declare dependencies** with a `depends-on:` frontmatter list, checked
  and ordered by vloop.
- **A loop brief is the only brief vloop knows.** Files named
  `B<YYYYMMDD-HHMM>-<slug>.loop-brief.md` are vloop's input; any other document
  upstream of one (design briefs, architect briefs, specs) is the consumer's
  business and vloop neither lists nor checks it.
  *Rejected:* a vloop-owned name for the design-act input — the design act is
  upstream of vloop, and naming it would make vloop own a process it does not run.
- **Libraries:** cobra for commands and completion, a TOML library for the
  config, the standard `testing` package with golden files.
  *Rejected:* koanf, lipgloss and goreleaser now — their consumers (`status
  --watch`, releases) are later briefs. Also a stdlib-only `flag` tree, because
  completion would have to be hand-written.
- **Cross-OS is proven by cross-compiling.** The loop runs on macOS; Windows and
  Linux are proven by building for them in every gate.
  *Rejected:* a CI matrix in this run — it cannot be a gate, because the loop
  runs offline.

## Binding references

- `docs/additional-context-files/20260929-0951-vloop-design-session.md` — the
  design this slice implements: sections 1–2 (name, module path, the plugin in
  plugin/ embedded with the `all:` prefix, marketplace manifest) and section 4
  (`## Binding references`, `brief check`). Where it disagrees with this brief,
  this brief wins; it predates the .vloop/ and Spanish decisions.
- `docs/design-notes/vloop-roadmap.md` — the series. Its B2–B4 rows are this
  run's out-of-scope list, and its "decisions every brief inherits" bind here.
- `.loop/check-brief.sh` — the rules `vloop brief check` ports. Port the rules,
  not the shell: its .loop/state/journals/ path becomes .vloop/state/journals/.
- `.loop/loop-brief.template.md` — the English section set and the guidance
  prose `brief new` starts from.

## Behaviour contract

### The binary and its global surface

- Go module `github.com/mvelosop/vloop`. The binary is built with
  `go build -o <out>/vloop ./cmd/vloop` and is named `vloop`. (Packaging the `vl`
  alias is a release concern and out of scope.)
- Global flags on every command: `-C, --dir <path>` (act as if started there),
  `--json` (machine output on stdout), `--no-color` (the `NO_COLOR` environment
  variable, if set to anything non-empty, does the same), `-q, --quiet`,
  `-v, --verbose`.
- **Exit codes**, for every command in this slice: `0` success; `1` the command
  ran and found problems or failed (a brief with problems, a malformed config
  file, a file that already exists); `2` usage — an unknown command, flag or
  config key, a missing argument, or an invalid config value. Errors go to
  stderr as one line starting with `vloop: `.
- With `--json`, stdout carries exactly one JSON document and nothing else, on
  success and on exit 1. Usage errors (exit 2) print to stderr only.
- Colour is emitted only when stdout is a terminal and colour is not disabled.
  Output captured by a test is never coloured.

### The repo root

Every command resolves the **repo root** by walking up from the `-C` directory
(or the working directory): the nearest ancestor containing a `.vloop`
directory, else the nearest containing `.git`, else the start directory itself.
All paths vloop prints are relative to that root, with `/` separators on every
OS.

### `vloop version`

```
vloop 0.0.0-dev (commit unknown, plugin 0.0.0-dev, go1.x.y darwin/arm64)
```

`--json` prints
`{"version":…,"commit":…,"plugin":…,"go":…,"os":…,"arch":…}` with those keys in
that order. `version` and `commit` are stamped at build time with `-ldflags -X`
and default to `0.0.0-dev` and `unknown`. `plugin` is read at runtime from the
**embedded** plugin.json, never from the working tree. `go`, `os` and `arch`
are the runtime's own values; a gate threads them rather than hardcoding them.

### The embedded plugin

The repo holds the plugin skeleton the design names: a plugin manifest named
`vloop` with version `0.0.0-dev` under plugin/.claude-plugin/, and a
marketplace manifest at the repo root under .claude-plugin/ whose single
plugin is `{"name":"vloop","source":"./plugin"}`. The binary embeds plugin/
including its dot-directories. A test proves the embedded tree contains the
manifest — the `all:` prefix is the trap the design names, and a build without
it passes everything else. No skills, agents or hooks are written in this run.

### `vloop config get | set | list | path`

Keys, in this order, with their defaults and valid values:

| Key | Default | Valid |
| --- | --- | --- |
| `language` | `en` | `en`, `es` |
| `model.plan` | `opus` | any non-empty string |
| `model.work` | `sonnet` | any non-empty string |
| `model.review` | `sonnet` | any non-empty string |
| `effort.plan` | unset | `low`, `medium`, `high`, `xhigh`, `max` |
| `effort.work` | unset | same |
| `effort.review` | unset | same |

- Resolution: environment variable, else config file, else default. The
  variable for a key is `VLOOP_` plus the key upper-cased with `.` as `_`:
  `VLOOP_LANGUAGE`, `VLOOP_MODEL_WORK`, `VLOOP_EFFORT_REVIEW`. An invalid value
  from the environment or the file is exit 1, naming its source.
- `config list` prints every key, one per line, as `<key>=<value> (<source>)`,
  where source is `default`, `file` or `env`. An unset effort prints as
  `effort.plan= (default)`. `--json` prints
  `{"<key>":{"value":…,"source":…},…}` in the same key order, with an unset
  value as `null`.
- `config get <key>` prints the resolved value alone and a newline (an empty
  line for an unset effort). `--json`: `{"key":…,"value":…,"source":…}`.
- `config set <key> <value>` validates, then writes .vloop/config.toml under
  the repo root, creating the directory and file if absent and preserving every
  other key already in it. Setting a value already set leaves the file byte-for-
  byte unchanged. `config set <key> ''` removes the key from the file.
- `config path` prints .vloop/config.toml whether or not it exists.
- Unknown key: `vloop: unknown config key "<key>"`, exit 2. Invalid value:
  `vloop: invalid value "<value>" for <key>: want one of <a>, <b>, …`, exit 2
  (for models: `want a non-empty string`).
- Only `config set` and `brief new` write anything. No command in this slice
  writes outside .vloop/ and the briefs directory.

### The vloop brief format

A brief is a Markdown file named `B<YYYYMMDD-HHMM>-<slug>.loop-brief.md` in the
briefs directory, `docs/briefs/` under the repo root, with YAML frontmatter in
the convention `.loop/loop-brief.template.md` already uses, plus `ready` and
`depends-on`:

```
---
name: B20260929-1804-example.loop-brief
description: Build the example
kind: brief
status: draft            # draft | ready | consumed
created: 2026-09-29
seeds: A run that builds the example
depends-on: []           # names of loop briefs that must be consumed first
---
```

- **`name`** must equal the filename minus `.md`. The brief's **run id** is the
  name minus `.loop-brief` (`B20260929-1804-example`); it names the run's journal.
- **`status` in the frontmatter is the only plannable marker.** `brief check`
  checks briefs with `status: ready` and reports the others as skipped. (This
  replaces the shell loop's `**Status:** ready to plan` body line.)
- `description`, `kind`, `created` and `seeds` are accepted and not required;
  unknown keys are ignored.
- **Only `*.loop-brief.md` files are briefs.** Every other file in the briefs
  directory is invisible to vloop.
- **Section headings** come from the heading set of the repo's `language`. A
  brief whose headings are in the other language is a brief with those sections
  missing — the check does not guess.

| Section | `en` | `es` |
| --- | --- | --- |
| What it is | `## What it is` | `## Qué es` |
| Why this shape | `## Why this shape, and what was rejected` | `## Por qué esta forma, y qué se descartó` |
| Binding references | `## Binding references` | `## Referencias vinculantes` |
| Behaviour contract | `## Behaviour contract` | `## Contrato de comportamiento` |
| Worked example | `## Worked example` | `## Ejemplo trabajado` |
| Out of scope | `## Out of scope` | `## Fuera de alcance` |
| Constraints | `## Constraints` | `## Restricciones` |
| Shape | `## Shape` | `## Forma` |

Headings match case-insensitively and accent-sensitively; trailing text after
the name is allowed (`## Worked example — the happy path`).

- **Binding references**: each list item in that section starts with one
  backticked repo-relative path, then ` — ` or ` - `, then a non-empty reason.

### `vloop brief check <path>…`

Ports every rule of `.loop/check-brief.sh`, with the same problem/warning split,
against the configured heading set:

- **Problems** (the brief fails): its journal .vloop/state/journals/<run id>.md
  exists (already run); no worked-example section; no fenced block; no
  out-of-scope section, or fewer than two list items in it; an absolute home path
  (the macOS and Linux home prefixes the shell check matches, and a Windows user
  profile path).
- **Warnings** (do not fail): no constraints section; no task count; no exit or
  status codes; more than two pinned internal symbols; a backticked path that
  does not resolve outside `## Binding references`; tracker keys or tracker URLs
  no offline session can open.
- The task-count and exit-code rules accept Spanish phrasings in an `es` repo:
  `6 a 9 tareas`, `9 tareas`, `código de salida 1`, `sale con 0`.

New rules:

- **Frontmatter**: missing or unparseable frontmatter, a missing `name` or one
  that does not match the filename (`name <name> does not match filename
  <filename minus .md>`), or a `status` outside `draft|ready|consumed` is a
  problem. These apply to every
  brief, ready or not.
- **Binding references**: a missing section is a warning (`no binding
  references — nothing binds a task to a document`). In the section: a path that
  does not resolve from the repo root is a problem (`binding reference does not
  resolve: <path>`); an entry with no reason is a problem (`binding reference has
  no reason: <path>`); a cited directory (path ending in `/`) with no
  README.md, index.md or README-*.md is a problem (`binding reference
  directory has no entry point: <path>`).
- **Dependencies**: a name in `depends-on` with no loop brief in the briefs
  directory is a problem (`depends-on does not resolve: <name>`); a cycle is a problem
  (`depends-on cycle: <a> -> <b> -> <a>`, starting from the checked brief).

Output: per brief, the repo-relative path, then one line per rule: `✓` for a
pass, `!` for a warning, `✗` for a problem, then a summary line —
`ok, <n> warning(s)` or `<n> problem(s), <m> warning(s)` — and at the end
`briefs ok` or `<n> brief(s) need work`. A skipped brief prints its path and
`- skipped: status is <status>`. A path that does not exist prints
`vloop: no such brief: <path>` on stderr and counts as a brief needing work.
A path whose name does not end in .loop-brief.md is a usage error before anything is
checked: `vloop: not a loop brief: <path>`, exit 2.
Exit 0 if no brief has problems, else 1. `--json`:
`{"ok":bool,"briefs":[{"path":…,"result":"ok|problems|skipped","problems":[…],"warnings":[…]}]}`.

### `vloop brief new <slug>`

Writes `docs/briefs/B<YYYYMMDD-HHMM>-<slug>.loop-brief.md` (local time) from
the template in the configured language, with `name`, `kind: brief`,
`status: draft`, `created` (today) and `depends-on: []` filled in, and prints
the created path. The template has every section of
the heading set, with guidance prose in that language. A slug that is not
lower-case letters, digits and single hyphens is exit 2. If the file exists,
`vloop: brief exists: <path>`, exit 1, and the file is untouched. `--dry-run`
prints the path and writes nothing. A freshly created brief passes `brief check`
as skipped (it is a draft).

### `vloop brief list`

Lists every `*.loop-brief.md` in the briefs directory in dependency order (ties
broken by name), one per line as `<name>  <status>  <ready>  <depends-on>`,
where `<ready>` is `ready` if every dependency is `consumed`, `blocked` if not,
and `-` for a brief that is itself consumed; `<depends-on>` is the names joined
by `,`, or `-`. `--json`: an array of
`{"name","path","status","ready","depends_on"}`.
A cycle or a dangling dependency is exit 1 with the same message `brief check`
gives; nothing is listed.

### Decided here, because the cited documents leave it underdetermined

- The design session never pinned where vloop finds briefs; `docs/briefs/` is
  decided here, and a briefs-directory config key is not added until someone
  needs one.
- The design session keeps the body `**Status:**` line; vloop drops it for the
  frontmatter `status`, because the frontmatter is English by rule and the body
  line would need a Spanish twin.
- The shell loop's loop-brief files share vloop's name and frontmatter
  convention, so they are vloop briefs and are checked on their merits. One
  written for the shell loop typically fails only on `status` (`draft` where
  it is plannable) — which is correct: it is not ready under vloop's rule.

### Violations the review must rule on

- Anything vloop, or any of its tests, reads or writes under this repo's
  `.loop/`, `.claude/` or the user's home directory. Tests run in temporary
  directories.
- A Spanish heading or message table that exists only in a test file — the heading
  sets are product data and live in the product.
- Any of vloop's own CLI messages translated. Only headings and template prose are
  bilingual.
- A template or the plugin manifest read from the working tree at runtime instead
  of the embedded copy.
- `task`, `status`, `init`, `doctor`, `run`, schemas or skill content appearing in
  this run.
- A README that documents a command, flag or key the binary does not have, or
  that a first-time reader could not follow without having read the design
  session.

## Worked example

Run from a fresh scratch git repository with the binary on `PATH`. Values that
are generated (the timestamp in a new brief's name, the Go version, OS and arch)
are threaded from what the command printed, never hardcoded.

```
$ vloop version --json
{"version":"0.0.0-dev","commit":"unknown","plugin":"0.0.0-dev","go":"<runtime>","os":"<goos>","arch":"<goarch>"}
                                                              exit 0
$ vloop config list
language=en (default)
model.plan=opus (default)
model.work=sonnet (default)
model.review=sonnet (default)
effort.plan= (default)
effort.work= (default)
effort.review= (default)                                      exit 0

$ vloop config set language es                                exit 0
$ vloop config set effort.review high                         exit 0
$ VLOOP_MODEL_WORK=opus vloop config get model.work --json
{"key":"model.work","value":"opus","source":"env"}            exit 0
$ vloop config get language
es                                                            exit 0
$ vloop config set language fr
vloop: invalid value "fr" for language: want one of en, es    exit 2
$ vloop config get colour
vloop: unknown config key "colour"                            exit 2

$ vloop brief new pagos-base
docs/briefs/B<stamp>-pagos-base.loop-brief.md                 exit 0
$ vloop brief new pagos-base          (same minute)
vloop: brief exists: docs/briefs/B<stamp>-pagos-base.loop-brief.md
                                                              exit 1
$ vloop brief check docs/briefs/B<stamp>-pagos-base.loop-brief.md
docs/briefs/B<stamp>-pagos-base.loop-brief.md
  - skipped: status is draft
briefs ok                                                     exit 0
$ vloop brief check docs/notes.md
vloop: not a loop brief: docs/notes.md                        exit 2
```

A ready Spanish brief, docs/briefs/B20260101-0900-a.loop-brief.md, with
`depends-on: [B20260101-0800-b.loop-brief]`, a `## Ejemplo trabajado` with a fenced block,
a `## Fuera de alcance` with two items, `## Restricciones`, `6 a 9 tareas`,
`código de salida 1`, and a `## Referencias vinculantes` entry
whose backticked path is docs/x.md and whose reason is "el contrato de
errores", where docs/x.md exists; and a
consumed B20260101-0800-b.loop-brief.md:

```
$ vloop brief check docs/briefs/B20260101-0900-a.loop-brief.md  exit 0
$ vloop brief list
B20260101-0800-b.loop-brief  consumed  -  -
B20260101-0900-a.loop-brief  ready  ready  B20260101-0800-b.loop-brief
                                                              exit 0
```

The failures, each planted in a copy of that fixture:

```
worked-example heading written as "## Worked example" in the es repo
  -> ✗ no worked example section — nothing arbitrates a disagreement   exit 1
docs/x.md deleted
  -> ✗ binding reference does not resolve: docs/x.md                   exit 1
reason removed from the entry
  -> ✗ binding reference has no reason: docs/x.md                      exit 1
b made to depend on a
  -> ✗ depends-on cycle: B20260101-0900-a.loop-brief -> B20260101-0800-b.loop-brief -> B20260101-0900-a.loop-brief
                                                                       exit 1
  and `vloop brief list` prints the same cycle on stderr               exit 1
.vloop/state/journals/B20260101-0900-a.md created
  -> ✗ already run — .vloop/state/journals/B20260101-0900-a.md exists  exit 1
name changed to B20260101-0901-a.loop-brief in the frontmatter
  -> ✗ name B20260101-0901-a.loop-brief does not match filename B20260101-0900-a.loop-brief
                                                                       exit 1
```

## Out of scope

The roadmap's later rows, not smaller versions of them:

- `task …`, `status`, and the schemas/ directory (B2) — including the per-task
  model and effort overrides.
- `init`, `upgrade`, `doctor`, plugin extraction, `--plugin-dir` (B3).
- Skill, agent and hook content for `/vloop:plan`, `/vloop:work`,
  `/vloop:review` — the plugin ships a manifest only (B3, as hygiene edits).
- `run` and its exit codes 0–7 (B4). Nothing in this run invokes `claude`.

Also not in this run:

- Localizing vloop's own messages, help text or JSON keys.
- A `briefs.dir` config key, a user-global config, `vloop help exit-codes`.
- `runs`, `status --watch`, lipgloss, goreleaser, Homebrew/Scoop, the `vl`
  alias, a CI workflow.
- Changing this repo's existing loop briefs, or touching `.loop/` in
  any way — it stays as the record of how vloop was built.
- Any other brief type (design, architect): recognizing, listing, templating or
  checking it.
- Documentation beyond the root README: a docs site, man pages, per-command
  guides.

## Constraints

- Go, the current stable release, pinned by the `go` line in `go.mod`. Third-party
  modules: cobra and one TOML library; anything else needs the reason in the task
  notes. Module downloads during the run are fine; vloop itself never uses the
  network.
- Must build for `darwin`, `linux` and `windows`. Nothing may assume `/` as the
  path separator on disk, a POSIX shell, or a case-sensitive filesystem.
- **Gates follow the diff surface.** Each task gates on the packages it touches,
  plus `go vet` and the three cross-builds
  (`GOOS=linux go build ./...`, `GOOS=windows go build ./...`, and the host).
  No task but the last runs `go test ./...` for the whole module.
- `gofmt -l .` prints nothing, in every gate.
- A gate runs in under a minute. Tests use temporary directories and never touch
  this repo's `.loop/`, `.claude/`, or the home directory.
- **No task may weaken a check to pass.** A check that is wrong is a blocked task
  with a note, not an edit to the check.
- **New checks are proven by fixtures**: every `brief check` rule, English and
  Spanish, has a passing fixture and a fixture where it must fire.
- Repo-relative paths everywhere. No absolute paths in any file or commit message.

## Shape

8 to 10 tasks, each independently verifiable by a single command. Order is
load-bearing:

1. **Skeleton**: `go.mod`, the command tree with global flags and the 0/1/2
   exit-code mapping, `version` (text and `--json`), the plugin and marketplace
   manifests, and the embed. Gated on building the binary, a test that runs it
   and checks `version --json` field by field, and the embedded-manifest test.
   The scaffolding gate runs something the project produces, not an empty test
   run.
2. **Repo root and `config`**, with every row of the key table, env overrides,
   byte-for-byte idempotent `set`, and the exit-2 cases.
3. **Brief parsing and the English `brief check`** — the port of
   `.loop/check-brief.sh`, with a fixture per rule where it passes and where it
   fires.
4. **The Spanish heading set** and the language switch, with the same fixtures in
   Spanish, plus the cross-language case (English headings in an `es` repo fail).
5. **Frontmatter and `## Binding references` rules.**
6. **`depends-on`**: resolution, cycles, and `brief list`.
7. **`brief new`** with both templates, embedded.
8. **The root README.md**, for a first-time reader: what vloop is; the loop in
   one paragraph (plan, then per task work → gate → review → commit, each a
   fresh session sharing nothing but files); a brief's lifecycle
   (`draft` → `ready` → `consumed`), `depends-on` and `## Binding references`;
   language, model and effort config; the .vloop/ layout; the exit codes; a
   pointer to the roadmap. Succinct: under 200 lines. Gated by a test that
   compares it against the binary's command tree — every command, flag and
   config key the README names exists, and every command exists in the README —
   plus the line cap. Whether a newcomer can follow it is the review's call; if
   it fails review twice on wording alone, block it and the operator finishes it.
9. **Close**: an end-to-end test that builds the binary and plays the worked
   example line for line in a scratch repo, plus `go test ./...`, `go vet ./...`,
   `gofmt -l .` and the three cross-builds.

The worked example needs its own gate (task 9): the unit tests in 1–8 can all
pass while the binary's wiring — flags, exit codes, stdout vs stderr — is wrong.

---
name: B20261103-0528-b36-i96-i97-plugin-per-repo-and-agent.loop-brief
description: Declare the vloop plugin in each repository's .claude/settings.json for Claude and Copilot, document using vloop from Copilot CLI, and have vloop doctor report the plugin per agent, for this repository, with its version
kind: brief
status: consumed
created: 2026-11-03
issue: 96
seeds: Issues 96 and 97 under 69; Copilot CLI 1.0.105's plugin behaviour as tested on 2026-11-01 (issue 96); the claude and copilot plugin list --json output on 2026-11-03; decided by the assistant under the operator's delegation of 2026-11-02 for the #81–#88 chain
depends-on: [B20261103-0315-b35-i83-fence-compound-commands.loop-brief]
---
# Brief — vloop B36: the plugin per repository and per agent

- **Issues:** #96 and #97 (under #69). The PR closes both.
- **Starting point:** extends `main` at `1e503513`, plus this brief's ready
  merge. The planner pins that SHA as the **base** every gate compares
  against — never `HEAD~n`.
- **Produced by:** the operator's delegation of 2026-11-02 for a chain of
  briefs on #84, #81, #82, #83, #96–#97 and #88. Written with
  `plugin/skills/create-brief/`; every decision below was taken by the
  assistant under that delegation.
- **Review:** `vloop brief review` (Claude Opus), 9 rounds, all FIX or
  ASK on facts the operator's tools could settle, each settled by a
  measurement recorded in `docs/design-notes/copilot-cli-experiments.md`:
  vloop's Claude and Copilot sessions ignore the declaration; Claude Code
  (marketplace known) does not act on it, so Claude's remedy is
  `--scope project`; Copilot's `plugin list` installs nothing and is
  computed per directory; a folder trusted before the declaration installs
  it on the next session; the two `plugin list --json` shapes. Fixed
  besides: the Copilot row tied to doctor's `copilot` row and kept out of
  `vloop run`'s preflight, the tests' changes, the upgrade output, exact
  bytes of every written file, the opt-out read from the repository,
  `trustKeyMatch`, the acceptance (Claude and Copilot, the upgrade path).
  Round 9's three findings (an unreadable settings file, `init --dry-run`
  on an invalid file, a user-level disable) fixed and not re-reviewed:
  they pin cases and add no reach. No `ASK` left.

## Executive summary

An operator gets `/vloop:operate` in an interactive Claude session by
installing the plugin once per user, or by starting Claude with
`--plugin-dir`. Nothing in a repository declares it, and nothing tells a
Copilot CLI user how to get it at all. Copilot CLI 1.0.105 reads a
repository's declaration (`extraKnownMarketplaces` and `enabledPlugins`)
from `.claude/settings.json`: once the folder is trusted, the first session
installs the plugin without a prompt, enabled in that repository only (#96,
tested 2026-11-01, and on 2026-11-03 for a folder trusted before the
declaration was added). Claude Code 2.1.317, for a user who already knew
the `vloop` marketplace, did not act on the declaration (measured
2026-11-03: no prompt, no install, no `/vloop:` command in a trusted
folder that holds it); what it does for a user who does not know the
marketplace was not measured; `claude plugin install vloop@vloop
--scope project` installs the plugin for one repository. Both are recorded
in `docs/design-notes/copilot-cli-experiments.md`, "A repository's plugin
declaration", which this brief's ready merge adds.
`vloop doctor` checks the plugin for Claude only, counts a plugin enabled in
another repository as enabled here, and its "installed but disabled" warning
names a command that changes the user's global settings, which a Copilot
session once tried to run (#97).

**The goal:** a repository where `vloop init` or `vloop upgrade` ran
declares the plugin, so a trusted Copilot CLI session gets it and a Claude
user installs it for the repository in one command; the docs say how to
use vloop from either agent; and `vloop doctor` reports, for Claude and for Copilot, whether
the plugin is enabled for this repository and whether its version is the
binary's, with remedies that change nothing outside the repository unless
the operator chooses to.

- **P1 — the declaration.** `vloop init` and `vloop upgrade` merge
  `extraKnownMarketplaces.vloop` and `enabledPlugins["vloop@vloop"]` into
  `.claude/settings.json`, keeping everything else.
- **P2 — both agents in the docs.** The README, the concepts guide, `vloop
  init`'s closing lines and the CLAUDE.md section give the Copilot commands
  next to Claude's, and Claude's per-repository install.
- **P3 — doctor's `plugin` row (Claude)** counts only entries for this
  repository (user scope, or a project entry whose path is this
  repository), and its messages say what each case means for a run.
- **P4 — doctor's new `copilot plugin` row**, when an `agent.*` key is
  `copilot`, from `copilot plugin list --json` run in the repository.
- Unchanged: how `vloop run` and `vloop brief review` load the plugin
  (`--plugin-dir`, every agent); the fence; nothing is installed or enabled
  by vloop outside the repository.

## What it is

A file vloop already owns part of (it reads `.claude/settings.json` for the
`gate edits` row) gains two keys at init and upgrade, and doctor's plugin
check becomes one per agent. The rest is documentation.

## Why this shape, and what was rejected

Decided by the assistant on 2026-11-03 under the operator's delegation.
**Do not re-litigate these** — the planner inherits them.

- **`.claude/settings.json` only.** Copilot CLI reads it (#96); for Claude
  it is the project settings file a project-scoped install belongs to. Also writing `.github/copilot/settings.json` was
  rejected: a second copy of the same two keys to keep in step.
- **Existing values win.** An `enabledPlugins["vloop@vloop"]` already
  present (even `false`, an operator's opt-out) and an
  `extraKnownMarketplaces.vloop` already present are kept as they are. A
  repository that wants the plugin off says so once and vloop leaves it.
- **The declaration installs the marketplace's version**, which may not be
  the binary's. Doctor's version check (P3, P4) says so and names the
  agent's update command; pinning a version in the declaration was
  rejected: the marketplace entry has no version selector.
- **vloop's own sessions are unaffected.** Claude sessions load the plugin
  vloop hands them and not the plugins the settings enable: measured
  2026-10-11 for user-scope plugins (`docs/guide/concepts.md`), and on
  2026-11-03 in a trusted clone holding this very declaration, where a
  session with `vloop run`'s flags loaded one vloop, the `--plugin-dir`
  copy, and installed nothing (`docs/design-notes/copilot-cli-experiments.md`). Copilot sessions run with
  their own `COPILOT_HOME`, and a `vloop brief review` on Copilot in a
  clone holding the declaration left that home's installed-plugins/ folder
  empty (measured 2026-11-03, the design note). The acceptance checks it
  again on a run.
- **The Claude row keeps its name, `plugin`.** Its `--json` name is read by
  scripts; the Copilot row is new, `copilot plugin`. Renaming both to
  `plugin claude` / `plugin copilot` was rejected for that reason.
- **Claude's remedy is a per-repository install.** Claude Code does not
  install a declared plugin by itself (measured); `claude plugin install
  vloop@vloop --scope project` installs it for this repository and
  matches the declaration. A user-level install stays as one option.
- **The Copilot row appears only when doctor's `copilot` row does**: when a
  key doctor's `copilot` row already reads is `copilot` (the `agent.*` keys
  in use, `agent.review-brief` included; not `agent.verify` with
  `run.verify = false`, nor `agent.second-review` with `run.second-review =
  off`). A repository that does not use Copilot never runs `copilot`
  from doctor or from `vloop run`'s preflight, and tests that do not
  configure Copilot never reach a real `copilot` on PATH. An operator who
  drives a Claude-only repository from Copilot CLI is served by the docs
  (P2), not by doctor; showing the row whenever `copilot` is on PATH was
  rejected for that reason. With the key set and the binary missing, the
  existing `copilot` row reports the problem and this row is `n/a`.
  `vloop run`'s preflight, which reuses doctor's checks and discards this
  row, does not compute it: no `copilot` command runs at a run's start.
- **No global change required by a remedy.** Each warning says first what
  it means for a run (nothing, where that is so) and names what the
  repository already provides (the declaration, once the folder is
  trusted) or `--plugin-dir`. A command that changes the user's state (a
  user-level install, `plugin update`) is named as an option the operator
  may take, never as required. `claude plugin enable vloop@vloop` (user
  scope) is no longer suggested; `claude plugin enable vloop@vloop --scope
  project`, which changes this repository's settings only, is.

## Binding references

- `internal/cli/init.go` — `runInit`'s steps, `mergedStep`, `applySteps`
  (the `wrote` / `updated` / `kept` lines and `--dry-run`) and
  `printInitNext`.
- `internal/cli/upgrade.go` — `runUpgrade`, which merges `CLAUDE.md` and
  `.gitignore` the same way.
- `internal/install/section.go` — the CLAUDE.md section's text and
  `MergeClaudeMD`; where `MergeClaudeSettings` would sit beside it.
- `internal/cli/doctor.go` — `runDoctor`'s order, `doctorPlugin`,
  `doctorGateEdits` (which already reads `.claude/settings.json`) and
  `doctorCopilot`.
- `internal/cli/doctor_test.go` — the stub `claude` on PATH and the plugin
  cases that pin today's messages.
- `docs/design-notes/copilot-cli-experiments.md` — "A repository's plugin
  declaration": what Copilot CLI does with the declaration, the source of
  P2's Copilot text.
- `docs/guide/concepts.md` — the `vloop doctor` paragraph and the plugin
  paragraph after it.
- `README.md` — the install section's plugin commands.
- `docs/domain/domain-model.md` — C-1 (configuration lives in the
  repository): the declaration is a repository file, and vloop writes
  nothing outside it.
- `docs/domain/platform/platform-context.md` — "CLI conventions": exit 1 on
  a problem, a refusal as one `vloop: ` line on stderr.

## Behaviour contract

### P1 — the declaration

- `vloop init` and `vloop upgrade` merge into `.claude/settings.json`:

  ```
  "extraKnownMarketplaces": {
    "vloop": { "source": { "source": "github", "repo": "mvelosop/vloop" } }
  },
  "enabledPlugins": { "vloop@vloop": true }
  ```

- **Merge rules.** Each of the two entries is added only when its key is
  absent: `extraKnownMarketplaces.vloop` and `enabledPlugins["vloop@vloop"]`.
  An existing value is kept as it is, whatever it is. Every other key, at
  every level, is kept with its value. The keys keep their order; a new key
  goes at the end of its object. The file is written with two-space
  indentation and a final newline, only when an entry was added.
- **No file:** it is created holding the two entries (no other key), with
  `.claude/` created if missing.
- **Output, init:** the step reports `wrote .claude/settings.json` (no file
  before), `updated .claude/settings.json` or `kept .claude/settings.json`
  like init's other steps; `--dry-run` reports `would write` / `would
  update` / `would keep` and writes nothing.
- **Output, upgrade:** as upgrade's other files: `updated
  .claude/settings.json` when it wrote the file (created or merged),
  nothing when it kept it, `would update .claude/settings.json` on
  `--dry-run` and on the breaking path before `--yes`. Upgrade's early
  return when the stamp is the binary's version (`already at <v>`) is
  unchanged: the declaration arrives with the next version.
- **A file that is not a JSON object** (invalid JSON, or an array): nothing
  is written anywhere and the command fails with a problem (exit 1):
  `.claude/settings.json is not a JSON object — fix it, then re-run`, for
  init before any other step writes (the symlink check's place), for
  upgrade before any file is written. Likewise when `enabledPlugins` or
  `extraKnownMarketplaces` is present but not an object (`null`, an array,
  a string): `.claude/settings.json's enabledPlugins is not a JSON object —
  fix it, then re-run` (the key at hand).
- **Both added to an existing file:** `extraKnownMarketplaces` first, then
  `enabledPlugins`, each at the end of the top-level object when it is
  new; an entry added to an existing object goes at its end.
- **Values are written back as they were read:** strings keep their text
  (`&`, `<` and `>` are not escaped), numbers their digits.
- **Layout, when the file is written:** the whole file is laid out as Go's
  `json.MarshalIndent` with `""` prefix and two spaces lays it out (each
  object member and each array element on its own line; an empty `{}` or
  `[]` inline), with HTML escaping off and a final newline. A file that is
  kept is not rewritten.
- **Upgrade's `--dry-run` and breaking path** read the file like the real
  run: a file that is not a JSON object (or whose `enabledPlugins` or
  `extraKnownMarketplaces` is not one) fails with exit 1 and the message
  above, before any `would update` line.
- A symlinked `.claude/settings.json` is refused as the other targets are
  (`checkRepoPath`).

### P2 — both agents in the docs

- The places that list what `vloop init` writes and what `vloop upgrade`
  refreshes name `.claude/settings.json` (the plugin's declaration):
  `docs/guide/concepts.md`, "Setting up a repository", both paragraphs;
  and `README.md`'s Quickstart step 1. The commands' help is unchanged.
- `printInitNext` prints exactly the block under "init's closing lines"
  below (each line's leading spaces as shown, counted from the block's left
  edge).
- The CLAUDE.md section's sentence "give the user the two ways to load it
  (claude plugin install vloop@vloop after claude plugin marketplace add
  mvelosop/vloop, or claude --plugin-dir "$(vloop plugin path)")" becomes
  "give the user the ways to load it (in Claude Code, claude plugin install
  vloop@vloop --scope project, after claude plugin marketplace add
  mvelosop/vloop if Claude does not know the marketplace, or claude
  --plugin-dir "$(vloop plugin path)"; in Copilot CLI, trust this folder,
  since .claude/settings.json declares the plugin, or copilot plugin install
  vloop@vloop after copilot plugin marketplace add mvelosop/vloop, or
  copilot --plugin-dir "$(vloop plugin path)")".
- `README.md`, the install section: the Claude block's install line
  becomes `claude plugin install vloop@vloop --scope project`, with a
  sentence that it is run inside the repository and installs the plugin
  for that repository (without
  `--scope project`, for every repository); then a paragraph saying
  `vloop init` declares the plugin in `.claude/settings.json`, so trusting
  the folder in Copilot CLI installs it for that repository; in Claude
  Code, run the install above; then the Copilot commands, as a block:

  ```
  copilot plugin marketplace add mvelosop/vloop
  copilot plugin install vloop@vloop
  ```

  and `copilot --plugin-dir "$(vloop plugin path)"` for one session.
- `docs/guide/concepts.md`: the plugin paragraph says the same (the
  declaration, that Copilot CLI installs it on trust, Claude Code's
  per-repository install, both agents' commands; it makes no claim about
  what Claude Code does with the declaration), and a new subsection **"From Copilot
  CLI"** after it says: trust the folder (the declaration then installs the
  plugin for this repository on the first session; it is ignored in an
  untrusted folder; a .github/copilot/settings.local.json file can opt one
  person out, a file that exists only in a repository that adds it); or install it per user (the two commands); or `copilot
  --plugin-dir "$(vloop plugin path)"` for one session; and that `vloop
  run`'s Copilot sessions load the plugin with `--plugin-dir` whatever is
  installed. The `vloop doctor` paragraph names the `copilot plugin` row.
  The sentence "For interactive sessions, either install the plugin or
  pass `--plugin-dir`, not both." becomes "For interactive sessions, use
  one of them: where the plugin is enabled for the repository (installed
  for it, or per user), do not also pass `--plugin-dir`."

init's closing lines:

```
next:
  vloop doctor
  get /vloop:operate for your Claude Code sessions in this repository, once:
    claude plugin marketplace add mvelosop/vloop
    claude plugin install vloop@vloop --scope project
  or, for one session: claude --plugin-dir "$(vloop plugin path)"
  for Copilot CLI: trust this folder in copilot (the plugin is declared in .claude/settings.json), or once:
    copilot plugin marketplace add mvelosop/vloop
    copilot plugin install vloop@vloop
  or, for one session: copilot --plugin-dir "$(vloop plugin path)"
  then ask Claude Code for what you want ("run the brief docs/briefs/<name>"); /vloop:operate drives vloop
```

### P3 — doctor's `plugin` row (Claude)

`claude plugin list --json`, now with the repository root as its working
directory. Its entries carry `scope` (`user`, `project`, `local`) and, for
the latter two, `projectPath` (recorded in
`docs/design-notes/copilot-cli-experiments.md`, "Measured on 2026-11-03";
the same id can appear once per project). An
entry counts when its `id` starts with `vloop@` and it is for this
repository: `scope` `user` or absent, or a `projectPath` that is this
repository's root, compared as doctor's `trust` row compares a Claude
project key with the root: `trustKeyMatch` against the root and against
its symlink-resolved path, as `doctorTrusted` does (separators normalised,
case ignored on macOS and Windows; the entry's path itself is not
resolved). Other entries are ignored, and so is an entry with an empty
or missing `version` (a declared plugin that is not installed, should
the list show one). A counted entry is enabled when its `enabled` is
true and neither `.claude/settings.json` nor .claude/settings.local.json
in the repository sets `enabledPlugins["vloop@vloop"]` to `false`: the
repository's opt-out wins whatever the list says, since which of the
list's flags (`enabled`, `projectEnabled`) reflects it is not recorded.
A settings file that is missing, unreadable, not a JSON object, or whose
`enabledPlugins` is not an object, is ignored for the opt-out (no warning
from this row; the `gate edits` row reads the files as today).

| Case | Result | Message |
| --- | --- | --- |
| a counted entry enabled, version = binary's | pass | — |
| counted entries enabled, none at the binary's version | warning | `the vloop plugin is <v> but this binary is <b> — vloop run is unaffected; claude plugin update vloop@vloop updates the plugin, or update the binary if it is the one behind` |
| counted entries, none enabled | warning | `installed but not enabled for this repository — vloop run is unaffected (its sessions load the plugin with --plugin-dir); for /vloop:operate here, claude plugin enable vloop@vloop --scope project (or with --scope user, if you disabled it for every repository), or start claude --plugin-dir "$(vloop plugin path)"` |
| no counted entry | warning | `/vloop:operate is not available in your interactive sessions — claude plugin install vloop@vloop --scope project installs it for this repository (after claude plugin marketplace add mvelosop/vloop if claude does not know the marketplace), or claude --plugin-dir "$(vloop plugin path)" for one session` |
| the command fails or prints no JSON array | warning | `cannot read claude plugin list --json: <error>` (as today) |

`<v>` is the version of the first enabled counted entry, in the list's
order. The row stays where it is and `n/a` without `claude`, as today.

### P4 — doctor's `copilot plugin` row

When doctor's `copilot` row appears (a key it reads is `copilot`:
`copilotKeys` and `agent.review-brief`, as today), a row `copilot plugin`
follows the `plugin`
row, right after it and before `gate edits` (`plugin` being `n/a` with no
`claude` does not hide it); it is `n/a`
when the `copilot` row is a problem. Doctor runs `copilot plugin list
--json` with the repository root as its working directory and the
environment as given (`COPILOT_HOME` included). Its entries' shape is
recorded in the design note's "Measured on 2026-11-03". With a per-user
install the list shows the plugin enabled in every directory;
with one the declaration installed, it shows it enabled only in that
repository (#96), which is why the root is the working directory. An entry
counts when its `name` is `vloop`.

| Case | Result | Message |
| --- | --- | --- |
| a counted entry enabled, version = binary's | pass | — |
| counted entries enabled, none at the binary's version | warning | `the vloop plugin for copilot is <v> but this binary is <b> — vloop run is unaffected; copilot plugin update vloop@vloop updates the plugin, or update the binary if it is the one behind` |
| counted entries, none enabled | warning | `installed but not enabled for this repository — vloop run is unaffected (its sessions load the plugin with --plugin-dir); for /vloop:operate here, set enabledPlugins["vloop@vloop"] to true in .claude/settings.json and trust this folder in copilot, or start copilot --plugin-dir "$(vloop plugin path)"` |
| no counted entry | warning | `/vloop:operate is not available in your copilot sessions — trusting this folder in copilot installs it when .claude/settings.json declares it (vloop init writes the declaration), or copilot plugin install vloop@vloop (after copilot plugin marketplace add mvelosop/vloop), or copilot --plugin-dir "$(vloop plugin path)" for one session` |
| the command fails or prints no JSON array | warning | `cannot read copilot plugin list --json: <error>` (the stderr's text when there is one) |

No row at all when there is no `copilot` row, when the config cannot
be read (the `config` row is then a problem, as today), or in `vloop run`'s
preflight.

### Decided here, because the cited documents leave it underdetermined

- The P1 step runs after the CLAUDE.md step in init, and after the
  `.gitignore` merge in upgrade.
- `vloop upgrade` on a repository where the two entries are present prints
  no line for the file and writes nothing.
- Doctor's `gate edits` row keeps reading the file as today.

### Violations the review must rule on

- A key or value of `.claude/settings.json` other than the two entries
  changed, reordered or dropped.
- An existing `enabledPlugins["vloop@vloop"]: false` turned `true`.
- Any write outside the repository; any `claude plugin` or `copilot plugin`
  command run by vloop other than `list --json`.
- A doctor message that presents a user-level change as required, or that
  suggests `claude plugin enable` without a `--scope`.

## Worked example

Each line is a named test: the P1 lines in `internal/cli`'s init and
upgrade tests, on temporary git repositories; the P3 and P4 lines in
`internal/cli/doctor_test.go`, with stub `claude` and `copilot` on PATH
that print the JSON shown. Each stub records its working directory, and every P3 and P4 test checks
that it was the repository root (`-C <root>` from another directory). In a
temporary git repository, built from the branch; doctor's agents are
stubs on PATH that print the JSON shown for `plugin list --json`.

```
# P1 (init on a fresh repository; upgrade on one whose stamp says 0.0.9, binary 0.1.0, run with --yes)
.claude/settings.json absent; vloop init (language en)
  → prints "wrote .claude/settings.json"; the file is, exactly:
{
  "extraKnownMarketplaces": {
    "vloop": {
      "source": {
        "source": "github",
        "repo": "mvelosop/vloop"
      }
    }
  },
  "enabledPlugins": {
    "vloop@vloop": true
  }
}
.claude/settings.json = {"permissions":{"allow":["Bash(go:*)"]},"enabledPlugins":{"other@x":true}}; vloop upgrade
  → "updated .claude/settings.json"; the file is, exactly:
{
  "permissions": {
    "allow": [
      "Bash(go:*)"
    ]
  },
  "enabledPlugins": {
    "other@x": true,
    "vloop@vloop": true
  },
  "extraKnownMarketplaces": {
    "vloop": {
      "source": {
        "source": "github",
        "repo": "mvelosop/vloop"
      }
    }
  }
}
.claude/settings.json = {"permissions":{"allow":["Bash(a && b)"]}}; vloop upgrade
  → "updated .claude/settings.json"; the string reads Bash(a && b), not \u0026; extraKnownMarketplaces then enabledPlugins after permissions
.claude/settings.json = {"enabledPlugins":null}; vloop upgrade
  → exit 1, ".claude/settings.json's enabledPlugins is not a JSON object — fix it, then re-run"; no file changed
.claude/settings.json = {"enabledPlugins":{"vloop@vloop":false}}; vloop upgrade
  → "updated .claude/settings.json"; the file is, exactly:
{
  "enabledPlugins": {
    "vloop@vloop": false
  },
  "extraKnownMarketplaces": {
    "vloop": {
      "source": {
        "source": "github",
        "repo": "mvelosop/vloop"
      }
    }
  }
}
the file from the first line; vloop upgrade        → no line for .claude/settings.json; the file's bytes unchanged
.claude/settings.json = [1]; vloop upgrade         → exit 1, ".claude/settings.json is not a JSON object — fix it, then re-run"; no file changed
vloop init --dry-run (no file)                     → "would write .claude/settings.json"; no file written
.claude/settings.json = [1]; vloop init --dry-run  → exit 1, ".claude/settings.json is not a JSON object — fix it, then re-run"; no "would" line, no file written
.claude/settings.json absent; vloop upgrade --dry-run → "would update .claude/settings.json"; no file written
.claude/settings.json = [1]; vloop upgrade --dry-run → exit 1, ".claude/settings.json is not a JSON object — fix it, then re-run"; no "would update" line
# P3 (binary 0.1.0, root R)
[{"id":"vloop@vloop","version":"0.1.0","scope":"user","enabled":true}]                       → ✓ plugin
[{"id":"vloop@vloop","version":"0.1.0","scope":"project","projectPath":"<R>","enabled":true}] → ✓ plugin
[{"id":"vloop@vloop","version":"0.1.0","scope":"project","projectPath":"/elsewhere","enabled":true}] → ! plugin /vloop:operate is not available in your interactive sessions — …
[{"id":"vloop@vloop","version":"0.0.9","scope":"user","enabled":true}]                       → ! plugin the vloop plugin is 0.0.9 but this binary is 0.1.0 — vloop run is unaffected; claude plugin update vloop@vloop updates the plugin, or update the binary if it is the one behind
[{"id":"vloop@vloop","version":"0.1.0","scope":"user","enabled":false}]                      → ! plugin installed but not enabled for this repository — vloop run is unaffected …
[{"id":"vloop@vloop","version":"0.1.0","scope":"user","enabled":true}], .claude/settings.json = {"enabledPlugins":{"vloop@vloop":false}} → ! plugin installed but not enabled for this repository — vloop run is unaffected …
[{"id":"vloop@vloop","version":"","scope":"project","projectPath":"<R>","enabled":true}]      → ! plugin /vloop:operate is not available in your interactive sessions — …
# P4 (agent.work = "copilot", a stub copilot on PATH that answers --version, GH_TOKEN set)
[{"name":"vloop","marketplace":"vloop","version":"0.1.0","enabled":true,"source":"installed"}]  → ✓ copilot plugin
[{"name":"vloop","marketplace":"vloop","version":"0.0.9","enabled":true,"source":"installed"}]  → ! copilot plugin the vloop plugin for copilot is 0.0.9 but this binary is 0.1.0 — vloop run is unaffected; copilot plugin update vloop@vloop updates the plugin, or update the binary if it is the one behind
[{"name":"vloop","marketplace":"vloop","version":"0.1.0","enabled":false,"source":"installed"}] → ! copilot plugin installed but not enabled for this repository — …
[{"name":"browser-use","marketplace":"","version":"0.0.4","enabled":false,"source":"builtin"}]  → ! copilot plugin /vloop:operate is not available in your copilot sessions — …
copilot exits 3 printing "boom" on stderr                                                     → ! copilot plugin cannot read copilot plugin list --json: boom
no agent.* key is copilot                                                                     → no copilot plugin row; the stub copilot is never run
```

The failures it must catch: a project entry of another repository counted
here; a user's `false` overwritten; a key reordered or dropped; a Copilot
list read from another directory; a remedy that changes global settings.

## Acceptance, at close (the operator's)

On a fresh clone of the URL-shortener template, with the branch's binary:

1. `vloop init`: `.claude/settings.json` holds the two entries.
2. Copilot, with a throwaway `COPILOT_HOME` and a GitHub token from `gh
   auth token`, started in the clone through tmux: accept the trust
   dialog; `$COPILOT_HOME/installed-plugins/` then holds vloop, and
   `copilot plugin list --json` shows it enabled in the clone and disabled
   in a directory outside it. Doctor's `copilot plugin` row (with
   `agent.work = "copilot"`) passes or warns on the version as P4 says.
   Then, in a second clone trusted in that home before `vloop init`:
   `vloop init`, a new session, and the plugin is installed there too
   (the upgrade path).
3. A `vloop run` whose work sessions are Copilot: while a session runs,
   and in the session homes the driver leaves under `.vloop/tmp/` after
   the run, no `copilot-home-*/installed-plugins/` holds vloop (the
   declaration is ignored in vloop's own sessions).
4. A `vloop run` whose sessions are Claude, in the trusted clone with the
   declaration: the runs' sessions load one vloop plugin, as measured on
   2026-11-03.
5. Claude Code in the clone: `claude plugin install vloop@vloop --scope
   project`; then `/vloop:operate` is offered in a session there and not
   in another repository, and doctor's `plugin` row passes or warns on
   the version as P3 says.

Results go in the design note; a behaviour that contradicts this brief is
a finding.

## Out of scope

- Writing `.github/copilot/settings.json`, AGENTS.md or Copilot
  instructions files.
- Installing, enabling or updating any plugin from vloop.
- How `vloop run` and `vloop brief review` load the plugin.
- #88 (`vloop metrics` per session, B37).

## Constraints

- Gates run on macOS's BSD tools: no empty alternatives in `grep -E`, no
  `sed -i` without a suffix, no `\+` or `\|` in basic regexes.
- Sessions never move refs.
- Tests build their fixtures in temporary directories and address them with
  `git -C`; `claude` and `copilot` are stubs on PATH; no test runs a real
  agent CLI or touches the real `~/.claude` or `~/.copilot`.
- The existing tests pass unchanged, except those that pin a doctor
  message P3 changes, the CLAUDE.md section's text, `printInitNext`'s
  lines, or init's and upgrade's output lines (the new
  `.claude/settings.json` line), or that asserts `vloop init` writes no
  `.claude/` (`TestDoctorGateEditsNoClaudeAndWritesNothing`, which then
  asserts the only file under `.claude/` is `settings.json`, holding the
  two entries): each changes only to the new text. The
  stubs' entries that carry no `scope` (in `internal/cli/doctor_test.go`
  and the `cmd/vloop` harnesses and end-to-end tests) count as
  user-scoped under P3, so their results do not change.
- No new module dependency.
- A docs gate joins lines and collapses whitespace before it greps, so an
  edit wrapped in the file's style passes.
- Repo-relative paths everywhere.

## Shape

4–5 tasks; every configured check green after each.

1. **P1** in init and upgrade, gated on its worked-example lines.
2. **P3 and P4**, gated on theirs.
3. **P2.** Its prose is judged on the facts each item lists, not on
   wording: a review that fails it twice on wording alone blocks the task
   for the operator.
4. **Close:** every gate at once, and the worked example line for line.
<!-- vloop:run-record:begin -->
## Run record

Generated by vloop brief close on 2026-11-05. Recompute with: vloop metrics B20261103-0528-b36-i96-i97-plugin-per-repo-and-agent.loop-brief

```
B20261103-0528-b36-i96-i97-plugin-per-repo-and-agent  consumed · not merged
 tasks     4 planned · 4 done · 0 blocked · first-pass 2/4
 size      delivered  code 345 · test 380 · docs 50 · test:code 1.10
           churn      code 345 · test 381 · docs 50 · rework 1.00
 time      agent 29.8 min (work 27.9 · review 1.9) · plan 9.7 min · gate review 1.9 min · verify 1.6 min · gates 0.4 · wall 69.9 min
 rate      11.6 code lines/min · 24.4 incl. tests
 cost      $5.67 · plan 2.89 · work 1.56 · review 0.53 · gate review 0.39 · verify 0.30 · $16.44/kloc
 models    plan claude-opus-5-5 · work claude-sonnet-5-5 · review claude-sonnet-5-5 · gate review claude-sonnet-5-5 · verify claude-sonnet-5-5
 defects   in-loop 2 · operator 0 · escaped 0 · removal efficiency 100%
 interventions  3 · recommended 0 · other-option 0 · adjusted 0 · different 0 · no-options 3
 instructions  plan 4,693 · gate review 2,338 · work 2,435 · review 2,456 · verify 2,109 words · guidelines 2 rules
```

```
id  area  kind     att  code+  test+  docs+  other+  test:code  $/kloc  agent   cost   model
T1  -     feature  2    210    233    0      0       1.11       3.34    10m01s  $0.70  claude-sonnet-5-5
T2  -     feature  2    128    128    0      0       1.00       5.13    4m30s   $0.66  claude-sonnet-5-5
T3  -     docs     1    7      20     50     0       2.86       70.23   9m13s   $0.49  claude-sonnet-5-5
T4  -     test     1    0      0      0      0       -          -       6m03s   $0.24  claude-sonnet-5-5
```

- B20261103-0528-b36-i96-i97-plugin-per-repo-and-agent/i1-gate — regression, work, found by gate: check failed
- B20261103-0528-b36-i96-i97-plugin-per-repo-and-agent/i3-gate — regression, work, found by gate: check failed
<!-- vloop:run-record:end -->

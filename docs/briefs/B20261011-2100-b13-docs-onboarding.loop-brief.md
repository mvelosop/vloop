---
name: B20261011-2100-b13-docs-onboarding.loop-brief
description: Make vloop learnable without reading its code — doctor warns when /vloop:operate is out of reach and when the operator may not edit gate folders, init and CLAUDE.md say how to get the skill, and the docs explain the name, asking instead of typing, PATH, the driver, gate-folder lifecycle and retiring the shell loop
kind: brief
status: ready
created: 2026-10-11
seeds: The T1 triage's P2; the learning loop follows as B14
depends-on: [B20261008-2144-b12-field-reliability.loop-brief]
---
# Brief — vloop B13: docs and onboarding

- **Starting point:** extends `main` after this design merges (`main` at
  `62919a9`, v0.9.0). The planner pins the base it plans from as the **base**
  every gate compares against — never `HEAD~n`.
- **Produced by:** the T1 triage, `docs/design-notes/vloop-t1-triage.md` → P2,
  and the design act of 2026-10-11, with the forks settled with the operator
  (intervention recorded with this brief).
- **Run by:** `vloop run` from the released v0.9.0.

## Executive summary

vloop is meant to be driven by asking, through `/vloop:operate`, but nothing
tells a new user so, checks that the skill is loaded, or explains the basics.
Seven items, each a gap a user meets today:

- **D1 — `doctor` hides a missing skill.** It passes when no vloop plugin is
  installed, so `/vloop:operate` is silently absent. It will warn, naming the
  two ways to load it.
- **D2 — gate repairs can stall on a permission.** Nothing checks that the
  operator's session may edit `.vloop/state/gates/`; B12's T3 sat blocked
  overnight on it. A new `doctor` check, `gate edits`, warns and prints the
  rule to add; vloop never writes it.
- **D3 — `init` and `CLAUDE.md` don't say how to get the skill.** `init`'s
  closing lines and the `CLAUDE.md` section will give the install commands and
  say to ask Claude Code rather than type commands.
- **D4 — the README assumes too much.** It will explain the name (verification
  loop), warn that Go's bin directory may not be on `PATH`, say "you ask; the
  skill drives", and link "driver" to its definition.
- **D5 — the guide leaves basics undefined.** It will define the driver, give
  `PATH` per shell, state how vloop's sessions load plugins (measured), say when
  gate folders are replaced, and list the new `doctor` checks.
- **D6 — no way to retire the shell loop.** A guide section lists what to
  remove and what to keep, and the docs stop saying the old layout is read "in
  this repository only".
- **D7 — two playbooks in this repository.** This repository's `CLAUDE.md`
  will say its own operator skill wins over `/vloop:operate` here.

Only D1–D3 change code, and only text and two warnings: no exit code, file
format or plugin skill changes.

## What it is

A user who installs vloop today meets a tool that is meant to be driven by
asking — the `/vloop:operate` skill — but nothing tells them so, and `vloop
doctor` passes while that skill is not loaded. They may not find the binary on
`PATH`, never learn what "the driver" is or where the name comes from, cannot
tell when gate folders go away, and have no way to retire the shell loop that
came before. After this run, `doctor` warns when `/vloop:operate` is out of
reach and when the operator's session may not edit gate folders; `init` and the
`CLAUDE.md` section say how to get the skill; and the README and guide explain
the name, asking instead of typing, `PATH`, the driver, the gate folders'
lifecycle, how vloop's sessions load plugins, and retiring the shell loop.

The blind spot: every document was written by someone who already had the
plugin loaded, Go's bin on `PATH` and the loop's vocabulary in their head.
`doctor`'s plugin check was tested against the loop's needs (`vloop run`
supplies the plugin), never against the user's.

## Why this shape, and what was rejected

Decided with the operator on 2026-10-11. **Do not re-litigate these** — the
planner inherits them.

- **One docs brief, with two `doctor` checks and two text changes in code.**
  *Rejected:* docs only (a document cannot warn a user whose setup is wrong).
- **`doctor` warns when no enabled vloop plugin is installed.** The `CLAUDE.md`
  section tells Claude to follow `/vloop:operate`; a pass there hides that it
  is not loaded. *Rejected:* keeping the pass with a clearer message (easy to
  miss). In this repository the warning stands until the plugin is installed;
  the operator accepted that.
- **`doctor` checks the gate-edit permission and only warns.** It reads Claude
  Code's settings and prints the exact rule to add. *Rejected:* `vloop init`
  writing the rule into `.claude/settings.json` — permissions stay the user's
  choice, and vloop does not write a file Claude Code owns.
- **Retiring the shell loop is a guide section.** *Rejected:* a `doctor` hint
  on finding `.loop/.installed` (few users ever had the shell loop).
- **The plugin-loading facts are measured, not guessed.** On 2026-10-11, a
  session started with `vloop run`'s flags (`--setting-sources project
  --plugin-dir <dir> --strict-mcp-config`) loaded the vloop plugin and none of
  the user-scope plugins the same machine had enabled; without those flags it
  loaded all of them. The guide states that, and advises against installing the
  plugin *and* passing `--plugin-dir` in one interactive session (two plugins of
  one name), without claiming what Claude Code then does.
- **The plugin's skills are not reworded.** Their uses of "the driver" are
  checked against the new definition by the review, not edited: a skill change
  needs its evals, which no loop session runs.

## Binding references

- `docs/design-notes/vloop-t1-triage.md` — P2: the proposal this brief implements
- `docs/design-notes/vloop-t1-triage-inputs.md` — field notes 1, 2, 5, 6, 7 and 8 and findings 1, 4 and 5: the facts each document states (the shell loop's footprint, the per-shell `PATH` files, when gate folders go)
- `docs/domain/domain-model.md` — the glossary's "driver", P-2, S-2, S-4 and R-1: what the driver owns, which the new definition must agree with; that the plan session writes the gate folders
- `docs/domain/execution/plan.md` — the plan's lifecycle (one plan per repo, reset by the next brief), which the gate folders follow

## Behaviour contract

Messages are exact where quoted; `<…>` marks a value. A `doctor` line is
`<mark> <check>[ <message>]`, as today.

### D1 — `doctor` warns when `/vloop:operate` is out of reach

**Defect.** With no vloop plugin installed, `doctor` prints `✓ plugin the
plugin is supplied by vloop run (--plugin-dir); …` (finding 1). The loop is
fine; the user's interactive sessions have no `/vloop:operate`.

**Required.** Where `claude plugin list --json` lists no `vloop@<any>` entry,
the plugin check is a **warning**:
`! plugin /vloop:operate is not available in your interactive sessions — claude plugin install vloop@vloop (after claude plugin marketplace add mvelosop/vloop), or claude --plugin-dir "$(vloop plugin path)" for one session`.
Every other outcome of the plugin check is unchanged: an enabled plugin of the
binary's version passes with no message, a disabled one or another version
warns as today, an unreadable list warns as today, no `claude` is `-`.

### D2 — `doctor` checks the operator's gate-edit permission

**Defect.** A gate repair (`vloop task verify`) means the operator's
interactive session edits `.vloop/state/gates/<id>/`. Under Claude Code's auto
mode, with no allow rule, that edit is refused; B12's T3 sat blocked overnight
on it (`.vloop/interventions/I20261011-1027-operator-allowed-the-assistant-to-edit-g.md`).

**Required.** A new check, **`gate edits`**, printed right after `plugin`:
- It reads `permissions.allow` from, in order, the user settings
  (`~/.claude/settings.json`), the project settings (`.claude/settings.json`)
  and the local settings (.claude/settings.local.json); a file that does not
  exist is skipped.
- It **passes**, with no message, when any of them allows `Edit` on the gate
  folders: an entry that is exactly `Edit`, or `Edit(<p>)` where `<p>`, after
  removing one leading `./`, is one of `**`, `.vloop/**`, `.vloop/state/**`,
  `.vloop/state/gates/**`.
- Otherwise it **warns**:
  `! gate edits no allow rule lets an interactive session edit .vloop/state/gates/ — add "Edit(./.vloop/state/gates/**)" to permissions.allow in .claude/settings.json`.
- A settings file that exists but cannot be read as JSON makes it a warning:
  `! gate edits cannot read <file>: <error>`, `<file>` as `~/.claude/settings.json`,
  `.claude/settings.json` or .claude/settings.local.json.
- It is `-` wherever the plugin check is `-` (no `claude`), and in `--json` it
  is the check named `gate edits`, like `gate shell`. It never writes any settings file.

### D3 — `init` and the `CLAUDE.md` section say how to get the skill

**Required.**
- `vloop init`'s closing lines become exactly:
  ```
  next:
    vloop doctor
    get /vloop:operate for your Claude Code sessions, once:
      claude plugin marketplace add mvelosop/vloop
      claude plugin install vloop@vloop
    or, for one session: claude --plugin-dir "$(vloop plugin path)"
    then ask Claude Code for what you want ("run the brief docs/briefs/<name>"); /vloop:operate drives vloop
  ```
- The `CLAUDE.md` section's interactive paragraph becomes exactly:
  `In an interactive session you are the operator's hands: follow the operator's playbook (/vloop:operate), not these session rules. If /vloop:operate is not available, the vloop plugin is not loaded: say so, give the user the two ways to load it (claude plugin install vloop@vloop after claude plugin marketplace add mvelosop/vloop, or claude --plugin-dir "$(vloop plugin path)"), and meanwhile work from vloop --help.`
  The rest of the section is unchanged; `vloop upgrade` rewrites it as today.

### D4 — the README

**Required.**
- The opening sentence explains the name: vloop is short for *verification
  loop* — the gate and the independent review are what the loop is about. The
  words `verification loop` appear in the first paragraph.
- **Install** keeps its command and adds, right after it, that `go install`
  puts the binary in `$(go env GOPATH)/bin` (`~/go/bin` by default), which is
  often not on `PATH`, with the line
  `export PATH="$PATH:$(go env GOPATH)/bin"` and a link to the guide's
  per-shell instructions (D5).
- A section **"You ask; the skill drives"** before "Everyday commands": the
  user need not learn the commands — with the plugin loaded (Install), they
  ask Claude Code in plain words ("run this brief", "what did that cost?",
  "record this as a defect") and `/vloop:operate` runs vloop. "Everyday
  commands" is framed as what you can ask for, and what the skill runs.
- "The loop in one paragraph" links "driver" to the guide's definition (D5).
- Every command or flag the README names exists (the README tests enforce it).

### D5 — the guide: the driver, `PATH`, plugins, gate folders, `doctor`

In `docs/guide/concepts.md`:
- **`## The driver`**, after "The loop": the driver is `vloop run` itself, the
  deterministic half of the loop. It plans nothing and judges nothing; it starts
  every session, runs every gate and check, owns every status change and makes
  every commit, and it is the boundary the permission fence only advises on
  (exit 9). A table or list of who does what: the driver; the plan, gate-review,
  work and review sessions; the operator. It agrees with the domain glossary,
  P-2, S-2 and R-1.
- **`PATH`**, in "Setting up a repository": `go install` never touches `PATH`;
  the line to add, and the file per shell: zsh `~/.zprofile` (or `~/.zshrc`),
  bash `~/.bash_profile` on macOS and `~/.profile` on most Linux, PowerShell the
  user `Path` variable (the Go MSI installer usually adds `%USERPROFILE%\go\bin`
  already; otherwise the `[Environment]::SetEnvironmentVariable` line of field
  note 7, preferred over `$PROFILE` because it reaches `cmd` and apps too).
- **Plugins**, where the plugin is described: `vloop run`'s sessions load the
  plugin it hands them and not the plugins enabled in the user's settings
  (measured 2026-10-11, see "Why this shape"); for interactive sessions, either
  install the plugin or pass `--plugin-dir`, not both.
- **Gate folders' lifecycle**, where the gate folder is described:
  `state.json`, `plan.md` and `.vloop/state/gates/` belong to the current plan;
  `vloop brief close` keeps them; they are replaced when the next brief is
  planned, or when a new plan is rejected as unfit; earlier plans' gate folders
  live in git history.
- The `doctor` paragraph lists the `gate edits` check and says the plugin check
  warns when no vloop plugin is installed.

### D6 — retiring the shell loop

**Required.**
- **`## Retiring the shell loop`** in `docs/guide/concepts.md`, for a
  repository that had `.loop/install.sh` installed and now runs vloop: remove
  `.claude/skills/loop-plan`, `loop-work` and `loop-review`, the
  `<!-- loop:begin -->…<!-- loop:end -->` block in `CLAUDE.md` (moving any rule
  it holds that vloop's section lacks — the shell loop's rule 1 on
  `~/.claude` — into the repository's own part of `CLAUDE.md`),
  `.claude/loop-knowledge.md` and `.loop/tmp/`; keep `.loop/state/`, with a
  one-line .loop/RETIRED.md, because vloop reads its `state.json` and
  `runs/` for metrics and defects; the rest of `.loop/` stays in git history.
- `docs/guide/metrics.md` and `docs/guide/concepts.md` stop saying the legacy
  layout is read "in this repository only": vloop reads `.loop/state/` in any
  repository that has it.

### D7 — this repository's two playbooks

**Required.** This repository's own part of `CLAUDE.md` (outside the vloop
markers) says, in "Which rules bind you", that where vloop's section names
`/vloop:operate`, this repository's playbook
(`.claude/skills/vloop-operator/SKILL.md`) is the one to follow here (finding
4). Nothing else in `CLAUDE.md` changes.

### The domain, updated

- None of the domain's rules change. `docs/domain/domain-model.md`'s glossary
  row for "driver" links to the guide's definition if it does not already say
  as much.
- `docs/guide/commands.md` is regenerated if any help text changes.

### Decided here, because the cited documents leave it underdetermined

- The `gate edits` match is literal: `Edit(/…)`, `Edit(~/…)` and other path
  forms are not recognised, and `deny` or `ask` entries and
  `permissions.defaultMode` are not read. A rule it does not recognise gives a
  warning, never a false pass.
- `vloop@<any marketplace>` counts as the vloop plugin, as today.
- The new `doctor` lines change no exit code: both are warnings.
- The README's new section names no command that does not exist; the skill's
  own name is written `/vloop:operate`.

### Violations the review must rule on

- Any `doctor` change that writes a file, or a `gate edits` pass for a rule
  outside the listed forms.
- A plugin-check outcome other than "no vloop entry" that changes.
- A document that states the driver plans, judges or reviews, or contradicts
  P-2, S-2 or R-1.
- A plugin skill (`plugin/skills/`) edited.
- A claim about what Claude Code does with two plugins of one name.
- Any refactor from B15's list folded in.

## Worked example

With a fresh build, in temporary repositories, with a stub `claude` that
answers `plugin list --json` and a temporary `HOME`:

```
vloop init in a fresh repository
  -> ends with the seven lines of D3, exactly                               exit 0
the new CLAUDE.md
  -> contains "If /vloop:operate is not available, the vloop plugin is not loaded"
claude plugin list --json prints []; no settings files
  -> ! plugin /vloop:operate is not available in your interactive sessions — claude plugin install vloop@vloop (after claude plugin marketplace add mvelosop/vloop), or claude --plugin-dir "$(vloop plugin path)" for one session
     ! gate edits no allow rule lets an interactive session edit .vloop/state/gates/ — add "Edit(./.vloop/state/gates/**)" to permissions.allow in .claude/settings.json
     doctor: 0 problem(s), 2 warning(s)                                     exit 0
plugin list prints vloop@vloop at the binary's version, enabled
  -> ✓ plugin                                                               (as today)
.claude/settings.json: {"permissions":{"allow":["Edit(./.vloop/state/gates/**)"]}}
  -> ✓ gate edits
.claude/settings.local.json: {"permissions":{"allow":["Edit(.vloop/**)"]}}  -> ✓ gate edits
$HOME/.claude/settings.json: {"permissions":{"allow":["Edit"]}}            -> ✓ gate edits
.claude/settings.json: {"permissions":{"allow":["Edit(./src/**)","Bash(go:*)"]}}
  -> ! gate edits no allow rule lets …
.claude/settings.json: {"permissions":
  -> ! gate edits cannot read .claude/settings.json: <error>
no claude on PATH
  -> - plugin
     - gate edits
vloop doctor --json
  -> a check named "gate edits", with its result and message
```

The gates thread the values they compute (the binary's version, file names)
rather than hardcoding them, and never touch the real `HOME`.

**Real-data check.** In this repository, with v0.9.0's plugin not installed:
`vloop doctor` prints `✓ gate edits` (the project settings' rule) and the
plugin warning; every other check's result is what the same build gave before
the change (a build from the branch also warns `self-hosting`, as always). The
README tests and guide tests pass. Every link the README and guide gain
resolves to an existing heading.

## Out of scope

- **B14, the learning loop** (the repo-guidelines file, gate-review eval cases,
  `vloop brief check` rules); **B15, code health**; **B16, the verification
  phase**; **B17, `vloop release`**.
- Any change to `plugin/skills/` or the evals.
- `vloop init` writing Claude Code settings; a command that retires the shell
  loop; a `doctor` hint about `.loop/`.
- Retiring the shell loop from *this* repository.
- `metrics.stacks` detection for shell and docs (field note 9's config half).
- A `doctor` check of `deny`/`ask` rules or permission modes.

## Constraints

- Go as pinned in `go.mod`; no new dependencies. Every gate builds for
  `windows` and `linux` as well as the host; reading the user's settings uses
  the platform's home directory, tested on every OS through a test seam.
- Gates run on macOS's BSD tools: no empty alternatives in `grep -E`, no
  `sed -i` without a suffix, no `\+` or `\|` in basic regexes, no `date -d`.
  A scratch directory is `mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX"`.
- Sessions never move refs and write nothing under `.vloop/` but `.vloop/tmp/`
  (S-4): no task's work, and no gate, may require a session to change this
  repository's records, config or state. Tests build their repositories in
  temporary directories, with a temporary `HOME`, and address them with
  `git -C`.
- **The gate model applies** (`plugin/skills/plan/SKILL.md`): each gate judges
  this brief's contract by running the built binary or reading the changed
  document, with its judge in the verify or its gate folder; no gate runs the
  repository's tests — this repository's `[[check]]` does, after every
  iteration. A gate that greps a document checks the headings and the exact
  strings this brief pins, nothing looser.
- A prose task that fails review twice on wording, not on a gate, is blocked
  for the operator, who finishes it by hand.
- **No task may weaken a check to pass.** These earlier tests **must** change,
  and only as stated:
  - `TestDoctorWorkedExampleText`, `TestDoctorPlugin` and
    `TestDoctorPluginMessages` (`internal/cli/doctor_test.go`) — the plugin
    warning for `[]` and the new `gate edits` line;
  - `TestWorkedExampleB11` (`cmd/vloop/b20261005_e2e_test.go`) — its doctor
    run with no marketplace plugin;
  - `TestWorkedExampleB5Session` (`cmd/vloop/b5_e2e_test.go`) — `init`'s
    closing lines and the doctor output (the new line, the warning count);
  - `TestInitWritesEverything` (`internal/cli/init_test.go`) — `init`'s
    closing lines;
  - `internal/install/install_test.go` only if it pins the interactive
    paragraph;
  - `docs/guide/commands.md`'s staleness test (regenerate, as always).
- Repo-relative paths everywhere. No absolute paths in any file or commit
  message.

## Shape

6 to 9 tasks, each independently verifiable by a single command. Order is
load-bearing:

1. **`doctor`'s plugin warning (D1).**
2. **`doctor`'s `gate edits` check (D2).**
3. **`init`'s closing lines and the `CLAUDE.md` section (D3).**
4. **The README (D4).**
5. **The guide (D5) and retiring the shell loop (D6).**
6. **This repository's `CLAUDE.md` (D7)** and the glossary link.
7. **Close:** the worked example line for line, and the real-data check.

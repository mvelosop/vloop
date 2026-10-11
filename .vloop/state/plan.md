# Plan — B20261011-2100-b13-docs-onboarding

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **5/7 done** · iteration 5

**Brief:** `docs/briefs/B20261011-2100-b13-docs-onboarding.loop-brief.md` · **Updated:** 2026-10-11T21:17:22Z

## Progress

- [x] **T1** — Make doctor's plugin check warn when no vloop plugin is listed
- [x] **T2** — Add doctor's `gate edits` check of the operator's permission to edit gate folders
- [x] **T3** — Make init's closing lines and the CLAUDE.md section say how to get /vloop:operate
- [x] **T4** — Explain the name, PATH, and asking rather than typing in the README
- [x] **T5** — Define the driver, PATH, plugin loading, gate-folder lifecycle and retiring the shell loop in the guide
- [ ] **T6** — Say in CLAUDE.md that this repository's operator playbook wins here, and link the glossary's driver
- [ ] **T7** — Close: commit the worked example as a test and pass the real-data check

## Tasks

### T1 — Make doctor's plugin check warn when no vloop plugin is listed

`done` · depends on: none

vloop doctor passes today when no vloop plugin is installed, so a user never learns that /vloop:operate is missing from their interactive sessions (brief D1). This task turns exactly that outcome into a warning that names both ways to load the plugin, and leaves every other outcome of the plugin check alone.

**Acceptance**

- When `claude plugin list --json` lists no entry whose id starts with `vloop@`, doctor prints exactly: ! plugin /vloop:operate is not available in your interactive sessions — claude plugin install vloop@vloop (after claude plugin marketplace add mvelosop/vloop), or claude --plugin-dir "$(vloop plugin path)" for one session
- In --json the plugin check is then result "warning" with that message (without the mark and the name)
- Every other outcome is byte-identical to before: an enabled vloop@<any> at the binary's version passes with no message; a disabled one, another version and an unreadable list warn with today's messages; no claude on PATH is `-`
- No exit code changes: the new outcome is a warning, never a problem; doctor writes no file
- TestDoctorPlugin, TestDoctorPluginMessages and TestDoctorWorkedExampleText (internal/cli/doctor_test.go) change only to expect the plugin warning for [], and committed tests cover the [] case and an other-marketplace-only list; TestWorkedExampleB11 and TestWorkedExampleB5Session change only where their doctor output now carries this warning
- No earlier test is weakened: only the tests the brief lists change, and only as it states

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T1/gate.sh
```

</details>

### T2 — Add doctor's `gate edits` check of the operator's permission to edit gate folders

`done` · depends on: T1

A gate repair means the operator's interactive session edits .vloop/state/gates/<id>/; under Claude Code's auto mode with no allow rule that edit is refused, and B12's T3 sat blocked overnight on it (brief D2). This task adds a doctor check that reads the user, project and local Claude Code settings and warns, printing the rule to add, when none allows it. It only reads; permissions stay the user's choice.

**Acceptance**

- A check named `gate edits` is printed right after `plugin`, in text and as the next entry in --json
- It reads permissions.allow from ~/.claude/settings.json, .claude/settings.json and .claude/settings.local.json, in that order, skipping a file that does not exist; the user's home comes from the platform's home directory through a test seam, so the tests set it on every OS without touching the real HOME
- It passes with no message when any of them has an entry exactly `Edit`, or `Edit(<p>)` where <p>, after removing one leading `./`, is one of `**`, `.vloop/**`, `.vloop/state/**`, `.vloop/state/gates/**`; any other form — `Edit(/…)`, `Edit(~/…)`, `Edit(.vloop/state/gates/*)`, `Write(…)`, a deny or ask entry, defaultMode — gives the warning, never a pass
- Otherwise it prints exactly: ! gate edits no allow rule lets an interactive session edit .vloop/state/gates/ — add "Edit(./.vloop/state/gates/**)" to permissions.allow in .claude/settings.json
- A settings file that exists but is not JSON gives `! gate edits cannot read <file>: <error>`, <file> written `~/.claude/settings.json`, `.claude/settings.json` or `.claude/settings.local.json`
- With no claude on PATH it is `- gate edits`, as plugin is `-`
- Violation: any code path in doctor that writes, creates or modifies a settings file (or any file); a pass for a rule outside the listed forms; any change to the exit code (both new lines are warnings)
- Committed tests (internal/cli/doctor_test.go) cover every listed pass form, the warning forms above, each unreadable file, the order of the three files, no claude, and --json; TestDoctorWorkedExampleText, TestWorkedExampleB11 and TestWorkedExampleB5Session change only for the new line and the warning count
- No earlier test is weakened: only the tests the brief lists change, and only as it states

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T2/gate.sh
```

</details>

### T3 — Make init's closing lines and the CLAUDE.md section say how to get /vloop:operate

`done` · depends on: T2

A user who runs vloop init is told to start a session with --plugin-dir and nothing more, and the CLAUDE.md section tells Claude to follow /vloop:operate without saying what to do when it is absent (brief D3). This task replaces init's closing lines with the install commands and the advice to ask rather than type, and extends the section's interactive paragraph with what Claude says when the plugin is not loaded.

**Acceptance**

- After a successful (non-dry-run) vloop init, its output ends with exactly the seven lines of brief D3, starting `next:`
- The CLAUDE.md section's interactive paragraph is exactly D3's text, one line, starting `In an interactive session you are the operator's hands:` and ending `and meanwhile work from vloop --help.`; every other line of the section is unchanged and vloop upgrade rewrites the section as today
- TestInitWritesEverything (internal/cli/init_test.go) and TestWorkedExampleB5Session (cmd/vloop/b5_e2e_test.go) assert the new closing lines; internal/install/install_test.go changes only if it pins the paragraph; a committed test asserts the new paragraph
- docs/guide/commands.md is regenerated if any help text changed, and its staleness test passes
- This repository's own CLAUDE.md is not rewritten by this task (no vloop upgrade run here)
- No earlier test is weakened: only the tests the brief lists change, and only as it states

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T3/gate.sh
```

</details>

### T4 — Explain the name, PATH, and asking rather than typing in the README

`done` · depends on: T3

The README assumes the reader already knows what vloop's name means, has Go's bin directory on PATH, and will type commands rather than ask Claude Code (brief D4). This task fixes the README so a newcomer installs successfully and learns that /vloop:operate drives vloop for them.

**Acceptance**

- The first paragraph says vloop is short for verification loop, with the words `verification loop`, and that the gate and the independent review are what the loop is about
- Install keeps `go install github.com/mvelosop/vloop/cmd/vloop@latest` and, right after it, says go install puts the binary in `$(go env GOPATH)/bin` (`~/go/bin` by default), often not on PATH, with the line `export PATH="$PATH:$(go env GOPATH)/bin"` and a link to the guide's per-shell instructions in docs/guide/concepts.md
- A section `## You ask; the skill drives` comes before `## Everyday commands`: with the plugin loaded the user asks Claude Code in plain words ("run this brief", "what did that cost?", "record this as a defect") and /vloop:operate runs vloop; "Everyday commands" is reframed as what you can ask for and what the skill runs
- "The loop in one paragraph" links "driver" to docs/guide/concepts.md#the-driver (the heading the guide gains in T5)
- Every command or flag the README names exists (the README tests pass); the skill is written `/vloop:operate`; no command is invented
- No plugin skill (plugin/skills/) is edited

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T4/gate.sh
```

</details>

### T5 — Define the driver, PATH, plugin loading, gate-folder lifecycle and retiring the shell loop in the guide

`done` · depends on: T4

The concepts guide leaves the driver undefined, says nothing about PATH, how vloop's sessions load plugins or when gate folders go, and offers no way to retire the shell loop; it and the metrics guide also say the legacy layout is read 'in this repository only', which is wrong (brief D5, D6). This task writes those sections so the README's links have somewhere to land and the docs agree with the domain.

**Acceptance**

- `## The driver` is the heading right after `## The loop` in docs/guide/concepts.md: the driver is `vloop run` itself, the deterministic half; it plans nothing and judges nothing; it starts every session, runs every gate and check, owns every status change and makes every commit, and is the boundary the permission fence only advises on (exit 9); a table or list says who does what — the driver; the plan, gate-review, work and review sessions; the operator
- Violation: any sentence stating the driver plans, judges or reviews, or contradicting the glossary's driver, P-2, S-2 or R-1 (e.g. a session committing or setting status, the driver writing gate folders)
- "Setting up a repository" says go install never touches PATH, gives `export PATH="$PATH:$(go env GOPATH)/bin"`, and the file per shell: zsh `~/.zprofile` (or `~/.zshrc`), bash `~/.bash_profile` on macOS and `~/.profile` on most Linux, PowerShell the user Path variable (the Go MSI installer usually adds `%USERPROFILE%\go\bin`; otherwise field note 7's `[Environment]::SetEnvironmentVariable` line, preferred over `$PROFILE` because it reaches cmd and apps too)
- Where the plugin is described: vloop run's sessions load the plugin it hands them and not the plugins enabled in the user's settings (measured 2026-10-11); for interactive sessions install the plugin or pass --plugin-dir, not both. Violation: any claim about what Claude Code does with two plugins of one name
- Where the gate folder is described: state.json, plan.md and .vloop/state/gates/ belong to the current plan; vloop brief close keeps them; they are replaced when the next brief is planned or when a new plan is rejected as unfit; earlier plans' gate folders live in git history
- The doctor paragraph lists the `gate edits` check and says the plugin check warns when no vloop plugin is installed
- `## Retiring the shell loop`: remove .claude/skills/loop-plan, loop-work and loop-review, the `<!-- loop:begin -->…<!-- loop:end -->` block in CLAUDE.md (moving any rule vloop's section lacks — the shell loop's rule 1 on ~/.claude — into the repository's own part), .claude/loop-knowledge.md and .loop/tmp/; keep .loop/state/ with a one-line .loop/RETIRED.md, because vloop reads its state.json and runs/ for metrics and defects; the rest of .loop/ stays in git history
- Neither docs/guide/concepts.md nor docs/guide/metrics.md says the legacy layout is read 'in this repository only'; they say vloop reads .loop/state/ in any repository that has it
- The guide tests pass; every link added resolves to an existing heading; no plugin skill (plugin/skills/) is edited, and the skills' uses of "the driver" are consistent with the new definition

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T5/gate.sh
```

</details>

### T6 — Say in CLAUDE.md that this repository's operator playbook wins here, and link the glossary's driver

`pending` · depends on: T5

This repository has two operator playbooks: vloop's section names /vloop:operate, and the repository's own section names .claude/skills/vloop-operator/SKILL.md (brief D7, finding 4). This task says, in the repository's own part of CLAUDE.md, which one to follow here, and links the domain glossary's 'driver' to the guide's new definition.

**Acceptance**

- "Which rules bind you" in CLAUDE.md says that where vloop's section names /vloop:operate, this repository's playbook (.claude/skills/vloop-operator/SKILL.md) is the one to follow here
- Nothing else in CLAUDE.md changes: not the vloop section between its markers, not the loop block, not the other sections
- The glossary row for "driver" in docs/domain/domain-model.md links to ../guide/concepts.md#the-driver; no domain rule (P-, S-, R-, B-) changes

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T6/gate.sh
```

</details>

### T7 — Close: commit the worked example as a test and pass the real-data check

`pending` · depends on: T1, T2, T3, T4, T5, T6

The brief's worked example and real-data check are the run's acceptance as a whole (brief: Worked example). This task commits the worked example as an end-to-end test over a fresh build, in temporary repositories with a stub claude and a temporary HOME, and confirms that in this repository doctor now passes gate edits and warns about the plugin, every other check unchanged, with every new link resolving.

**Acceptance**

- A committed end-to-end test (cmd/vloop) runs the brief's worked example line for line: init's seven lines, the CLAUDE.md sentence, the two warnings with `doctor: 0 problem(s), 2 warning(s)` and exit 0, the passing plugin, the three passing settings files, the warning rule, the unreadable file, no claude, and --json's `gate edits`
- The test builds its repositories in temporary directories with a temporary HOME, addresses them with git -C, never touches the real HOME, and threads the binary's version rather than hardcoding it
- In this repository, with no vloop plugin listed, the branch's doctor prints `✓ gate edits` and the plugin warning; every other check's result equals the base build's
- Every relative link in README.md, docs/guide/*.md and docs/domain/domain-model.md resolves to an existing file and heading
- No production code changes in this task beyond what the test needs to find; any defect found here is fixed in the smallest change and stated in the report
- No earlier test is weakened: only the tests the brief lists change, and only as it states

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T7/gate.sh
```

</details>


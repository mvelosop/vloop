# Plan — B20261103-0528-b36-i96-i97-plugin-per-repo-and-agent

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **0/4 done** · iteration 1

**Brief:** `docs/briefs/B20261103-0528-b36-i96-i97-plugin-per-repo-and-agent.loop-brief.md` · **Updated:** 2026-11-03T06:47:06Z

## Progress

- [ ] **T1** — Declare the plugin in .claude/settings.json from vloop init and vloop upgrade (P1) · 1 attempt(s)
- [ ] **T2** — Report the plugin per agent in vloop doctor: the plugin row for this repository, and a copilot plugin row (P3, P4)
- [ ] **T3** — Document using vloop from Claude Code and Copilot CLI: init's closing lines, the CLAUDE.md section, README and the concepts guide (P2)
- [ ] **T4** — Close the run: every gate at once, and the worked example line for line

## Tasks

### T1 — Declare the plugin in .claude/settings.json from vloop init and vloop upgrade (P1)

`pending` · 1 attempt(s) · depends on: none

A Copilot CLI session installs the vloop plugin for a repository once the folder is trusted, if .claude/settings.json declares it, and a Claude user installs it for the repository with one command. This task makes vloop init and vloop upgrade merge extraKnownMarketplaces.vloop and enabledPlugins["vloop@vloop"] into that file, adding each only when its key is absent and keeping every other key, value and order, so an operator's opt-out (false) survives. The docs (T3) and doctor's opt-out reading (T2) build on it.

**Acceptance**

- Every P1 line of the brief's worked example holds through the built binary, on temporary git repositories: the exact file bytes, the `wrote` / `updated` / `would write` / `would update` lines, no line and unchanged bytes when both entries are present, and exit 1 with `.claude/settings.json is not a JSON object — fix it, then re-run` (or `.claude/settings.json's <key> is not a JSON object — fix it, then re-run`, naming enabledPlugins or extraKnownMarketplaces) with no file changed.
- Merge rules: each entry is added only when its key is absent; an existing value is kept whatever it is (an `enabledPlugins["vloop@vloop"]: false` stays false, an existing `extraKnownMarketplaces.vloop` stays as it is); every other key at every level keeps its value and order; a new key goes at the end of its object, extraKnownMarketplaces before enabledPlugins when both are new at the top level. Violation: a key or value other than the two entries changed, reordered or dropped, or an existing false turned true.
- When written, the file is laid out as json.MarshalIndent with "" and two spaces lays it out (an empty {} or [] inline), with HTML escaping off (`&`, `<`, `>` written as themselves), numbers keeping their digits (1.50 stays 1.50, 12345678901234567890 stays as is), and a final newline; a file that is kept is not rewritten, so its bytes do not change.
- No file: it is created holding the two entries and nothing else, with .claude/ created if missing. init reports `wrote`, `updated` or `kept .claude/settings.json` after the CLAUDE.md step (`would write` / `would update` / `would keep` on --dry-run, which writes nothing); upgrade reports `updated .claude/settings.json` when it wrote the file, nothing when it kept it, and `would update .claude/settings.json` on --dry-run and on the breaking path before --yes; upgrade's `already at <v>` return is unchanged.
- A file that is not a JSON object (invalid JSON, an array, a string), or whose enabledPlugins or extraKnownMarketplaces is present but not an object, is refused with exit 1 and the message above: by init before any step writes and before any would-line on --dry-run; by upgrade before any file is written, on --dry-run and on the breaking path too, before any `would update` line. A symlinked .claude/settings.json is refused as the other targets are (checkRepoPath).
- Nothing is written outside the repository and no `claude plugin` or `copilot plugin` command is run by init or upgrade. No new module dependency.
- Committed tests in internal/cli's init and upgrade tests hold each P1 worked-example line as a named test on a temporary git repository (addressed with git -C), plus the merge-rule cases above; the `go` check runs them. Violation: the merge shipped with no committed test the check runs.
- Existing tests change only where the brief's Constraints allow: tests pinning init's or upgrade's output lines gain the new .claude/settings.json line, and TestDoctorGateEditsNoClaudeAndWritesNothing asserts the only file under .claude/ is settings.json holding the two entries.

**From the last attempt:** check go failed — see .vloop/state/runs/B20261103-0528-b36-i96-i97-plugin-per-repo-and-agent/20261103-062225/checks/001-go.fail.log

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T1/gate.sh
```

</details>

### T2 — Report the plugin per agent in vloop doctor: the plugin row for this repository, and a copilot plugin row (P3, P4)

`pending` · depends on: T1

Doctor's plugin row counts a plugin enabled in another repository as enabled here, and its disabled warning names a command that changes the user's global settings, which a Copilot session once ran. This task makes the Claude row count only entries for this repository (user scope or absent, or a projectPath matching the root as the trust row matches it), honour the repository's own opt-out, and say what each case means for a run; and adds a `copilot plugin` row, only where doctor's `copilot` row appears, from `copilot plugin list --json` run in the root. Both lists are computed for the working directory, so running them anywhere else gives the wrong answer.

**Acceptance**

- Every P3 and P4 line of the brief's worked example holds through the built binary with stub claude and copilot on PATH, doctor started with -C &lt;root&gt; from another directory: the row's result and its message exactly as the brief's tables give them.
- P3: `claude plugin list --json` runs with the repository root as its working directory. An entry counts when its id starts with `vloop@`, its version is not empty or missing, and it is for this repository: scope user or absent, or a projectPath matching the root or its symlink-resolved path by trustKeyMatch (the entry's path is not resolved). A counted entry is enabled when its enabled is true and neither .claude/settings.json nor .claude/settings.local.json sets enabledPlugins["vloop@vloop"] to false; a settings file that is missing, unreadable, not an object, or whose enabledPlugins is not an object, is ignored for this. `<v>` is the version of the first enabled counted entry in the list's order. Violation: a project entry of another repository counted here.
- P3's messages are exactly the brief's table: pass with no message; the version warning naming `claude plugin update vloop@vloop`; `installed but not enabled for this repository — …` naming `claude plugin enable vloop@vloop --scope project`; `/vloop:operate is not available in your interactive sessions — claude plugin install vloop@vloop --scope project …`; `cannot read claude plugin list --json: <error>` (stderr's text when there is one). The row stays where it is and is n/a without claude. Violation: a message that suggests `claude plugin enable` without a --scope, or presents a user-level change as required.
- P4: a `copilot plugin` row appears exactly when doctor's `copilot` row does (copilotKeys plus agent.review-brief; not agent.verify with run.verify false, nor agent.second-review with run.second-review off), right after the plugin row and before gate edits (also when plugin is n/a for want of claude); it is n/a when the copilot row is a problem; no row when no key is copilot or the config cannot be read, and then copilot is never run. It runs `copilot plugin list --json` in the root with the environment as given (COPILOT_HOME included); an entry counts when its name is `vloop`; its messages are exactly the brief's P4 table.
- vloop run's preflight does not compute the copilot plugin row: no `copilot plugin` command runs at a run's start, while the preflight keeps checking copilot as today.
- doctor runs no `claude plugin` or `copilot plugin` command other than `list --json` and writes nothing. No new module dependency.
- Committed tests in internal/cli/doctor_test.go hold every P3 and P4 worked-example line as a named test, with stub claude and copilot on PATH that print the JSON shown and record their working directory, each test checking it was the repository root (doctor run with -C &lt;root&gt; from another directory); a test shows the preflight runs no copilot plugin list. The `go` check runs them. Violation: the rows shipped with no committed test the check runs.
- Existing tests change only where the brief's Constraints allow (messages P3 changes); the stubs' scope-less entries in internal/cli/doctor_test.go and cmd/vloop's harnesses and end-to-end tests keep their results, and no test runs a real claude or copilot or touches the real ~/.claude or ~/.copilot.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T2/gate.sh
```

</details>

### T3 — Document using vloop from Claude Code and Copilot CLI: init's closing lines, the CLAUDE.md section, README and the concepts guide (P2)

`pending` · depends on: T1, T2

Nothing tells a Copilot CLI user how to get /vloop:operate, and the docs still give Claude's per-user install as the way. With the declaration (T1) and doctor's per-agent rows (T2) in place, this task makes init's closing lines, the CLAUDE.md section, README.md and docs/guide/concepts.md give both agents' ways: the declaration and trusting the folder in Copilot, Claude Code's per-repository install, and --plugin-dir for one session. Prose is judged on the facts P2 lists, not on wording.

**Acceptance**

- vloop init's closing lines are exactly the brief's "init's closing lines" block, after the step lines; the CLAUDE.md section's sentence is exactly the brief's new text (in Claude Code … --scope project …; in Copilot CLI, trust this folder …), and the old "give the user the two ways to load it" is gone.
- README.md: the install section's Claude install line is `claude plugin install vloop@vloop --scope project`, with a sentence that it runs inside the repository and installs for that repository (without --scope project, for every repository); a paragraph says vloop init declares the plugin in .claude/settings.json, so trusting the folder in Copilot CLI installs it for that repository, and in Claude Code to run the install above; then the Copilot block `copilot plugin marketplace add mvelosop/vloop` / `copilot plugin install vloop@vloop` and `copilot --plugin-dir "$(vloop plugin path)"` for one session. Quickstart step 1 names .claude/settings.json (the plugin's declaration).
- docs/guide/concepts.md, "Setting up a repository": the init and upgrade paragraphs name .claude/settings.json; the vloop doctor paragraph names the `copilot plugin` row; the plugin paragraph gives the declaration, that Copilot CLI installs it on trust, Claude Code's per-repository install and both agents' commands; the sentence "either install the plugin or pass `--plugin-dir`, not both" becomes the brief's "For interactive sessions, use one of them: where the plugin is enabled for the repository (installed for it, or per user), do not also pass `--plugin-dir`."
- A new subsection "From Copilot CLI" after the plugin paragraph says: trust the folder (the declaration then installs the plugin for this repository on the first session; it is ignored in an untrusted folder; a .github/copilot/settings.local.json can opt one person out, a file that exists only in a repository that adds it); or install per user (the two commands); or `copilot --plugin-dir "$(vloop plugin path)"` for one session; and that vloop run's Copilot sessions load the plugin with --plugin-dir whatever is installed.
- No doc makes a claim about what Claude Code does with the declaration (it was measured not to act on it for a user who knows the marketplace, and not measured otherwise). Violation: text saying Claude Code installs or enables the declared plugin. The commands' help (vloop init, upgrade, doctor) is unchanged.
- Committed tests that pin printInitNext's lines and the CLAUDE.md section's text change only to the new text, and pin it (the `go` check runs them); docs/guide/commands.md stays in step with the help (unchanged). This repository's own CLAUDE.md section is left to vloop upgrade, not hand-edited, unless a committed test requires it.
- The prose is judged on the facts each item lists, not on wording.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T3/gate.sh
```

</details>

### T4 — Close the run: every gate at once, and the worked example line for line

`pending` · depends on: T3

Each gate proved its task when it ran; the close proves them together on the final tree and accounts for the brief's worked example line by line, so the verify session and the operator find every line covered by a gate and a committed test, or named as a gap. Nothing new is built here unless a worked-example line turns out to have no committed test.

**Acceptance**

- Gates T1, T2 and T3 all pass on the final tree, run together by this task's gate.
- The work session's notes walk the brief's worked example line for line (P1, P3, P4) and name, for each, the committed test in internal/cli (init, upgrade or doctor tests) and the gate (T1 or T2) that covers it; a line no committed test covers is fixed in this task (a test added) or reported, not skipped.
- Every configured check is green on the final tree (the `go` check: go test, vet, gofmt, mod tidy, linux and windows builds).

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T4/gate.sh
```

</details>


# Plan — B20261015-1651-b17-release

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **0/9 done** · iteration 1

**Brief:** `docs/briefs/B20261015-1651-b17-release.loop-brief.md` · **Updated:** 2026-10-15T16:41:27Z

## Progress

- [ ] **T1** — Read the binary's version from the embedded plugin.json · 1 attempt(s)
- [ ] **T2** — Record the module version in .vloop/install.json
- [ ] **T3** — Add tools/release with the notes subcommand
- [ ] **T4** — Make prepare's checks, refusing before anything is written
- [ ] **T5** — Make prepare create the release branch and open the PR
- [ ] **T6** — Make publish find the release commit, tag it and create the GitHub release
- [ ] **T7** — Make publish prove the public install, and finish an interrupted release on re-run
- [ ] **T8** — Describe the release tool in the docs and the operator's playbook
- [ ] **T9** — Close: run the worked example line for line and the real-data check

## Tasks

### T1 — Read the binary's version from the embedded plugin.json

`pending` · 1 attempt(s) · depends on: none

Today a release edits the version in five places (main.go, plugin.json and three tests). V1 makes the embedded plugin/.claude-plugin/plugin.json the one version source, so a release bump is one line — the line `prepare` (T5) will change — and the binary and plugin can never disagree. Every later task relies on it: the release commit changes only plugin.json.

**Acceptance**

- The binary's version is the `version` of the embedded plugin/.claude-plugin/plugin.json: bumping that one line moves `vloop version` and `vloop version --json`, with no other edit.
- cmd/vloop/main.go holds no version literal and no `version` variable a -X flag could stamp; `-X main.commit=<sha>` still stamps the commit. A version literal left in cmd/vloop/main.go is a violation.
- `vloop version` and `vloop version --json` print exactly what the base prints for the same build (unstamped, and stamped with a commit).
- No test pins a version literal: the tests that compared 0.11.0 (cmd/vloop/e2e_test.go, cmd/vloop/run_harness_test.go's stub plugin list, internal/cli/root_test.go) read the version from the embedded plugin.json instead. A test still containing "0.11.0" as the expected version is a violation.
- A committed test, run by the checks, proves the binary's version equals the embedded plugin.json's version. An earlier test changes only where it pins the version literal, and only to read it from plugin.json; the task's notes name V1 for each test changed.
- `go generate ./...` leaves docs/guide/commands.md unchanged, or the regenerated file is part of the change.

**From the last attempt:** check go failed — see .vloop/state/runs/B20261015-1651-b17-release/20261015-160742/checks/001-go.fail.log

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T1/gate.sh
```

</details>

### T2 — Record the module version in .vloop/install.json

`pending` · depends on: T1

v0.11.0 showed that a public `go install …@v0.11.0` writes "commit":"unknown" to .vloop/install.json, so the stamp cannot say which release it is. I1 records the module version (as `vloop version` prints it) in the stamp, right after commit, whenever the binary carries one, so every install — public ones included — is identified. Publish's install check (T7) relies on the same build info.

**Acceptance**

- `vloop init` and `vloop upgrade` write "module": "<module version>" into .vloop/install.json, immediately after "commit", when the binary's build info carries a module version (the value `vloop version` prints as module); they leave the key out otherwise — including on an upgrade by a binary with none, over a stamp that had one.
- schemas/install.v1.json gains `module` as an optional string property; `required` and the schema name (install/v1) are unchanged, and a stamp without module stays valid.
- No other behaviour of init or upgrade changes (I1 only).
- Committed tests, run by the checks, cover init and upgrade with and without a module version (through the existing build-info seam, not a real network install) and the schema accepting a stamp with a string module and without one. An earlier test changes only where it pins the stamp's exact content, and only to the new behaviour; the notes name I1 for each.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T2/gate.sh
```

</details>

### T3 — Add tools/release with the notes subcommand

`pending` · depends on: T1

The release tool is a repository-owned Go program, `go run ./tools/release`, not a vloop subcommand. This task creates it with `notes <from> <to>`, the Markdown release notes built from the briefs merged in the range (found by their Vloop-Brief trailers) and the other code changes; prepare (T5) uses it as the PR body and publish (T6) as the release notes. It also runs the real-data check against this repository's own tags.

**Acceptance**

- tools/release is `package main` in this module (it may use internal packages); no new dependency in go.mod, and no `vloop` subcommand is added for releasing.
- `notes <from> <to>` prints, on stdout with exit 0, the brief's exact Markdown: `## Briefs`, a blank line, `- **<title>** — <description>` lines; a blank line; `## Other changes`, a blank line, `- <subject>` lines. A section with no lines is left out; with neither the output is `No changes.`
- Commits are those after <from> up to and including <to> (tags or revisions). A brief is named by a commit's `Vloop-Brief: <name>` trailer; one line per brief, oldest first; <title> is the brief's first heading with a leading `Brief — ` and `vloop ` removed, else the brief's name; <description> is its frontmatter description; both read at <to>.
- Other changes are the commits without the trailer whose changes touch a file outside docs/, .vloop/ and CLAUDE.md and whose subject does not start `Release v`, oldest first.
- Refusals are one stderr line starting `release: `; exit codes are 0 done, 1 refused or failed, 2 usage (a wrong number of arguments or an unknown subcommand).
- The task's notes carry the real-data check, run on a clone of this repository: the full output of `notes v0.10.0 v0.11.0` (B16 alone under Briefs) and of `notes v0.9.0 v0.10.0` (B13, B14 and B15).
- Committed tests, run by the checks, cover the notes format, the trailer, title and description rules, the exclusions and the empty sections, building their repositories in temporary directories with a temporary HOME and git -C; no test reads this repository's history or reaches the network.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T3/gate.sh
```

</details>

### T4 — Make prepare's checks, refusing before anything is written

`pending` · depends on: T3

`prepare <version>` must refuse a bad release before it writes, commits, tags or pushes anything. This task builds the checks in the brief's order — version form, module-path major, retraction, after the latest tag, tag absent, on main / clean / equal to origin/main after a fetch, every [[check]] then `claude plugin validate plugin` — and --dry-run, which makes them all and writes nothing. T5 adds what prepare does once they pass.

**Acceptance**

- The checks run in the brief's order, each with its exact message: `release: <version> is not a version: vX.Y.Z or vX.Y.Z-<pre>` (exit 2); `release: <version> needs the module path to end in /v<major>, and github.com/mvelosop/vloop does not`; `release: <version> is retracted in go.mod` (single versions and ranges); `release: <version> is not after the latest release <tag>`; `release: tag <version> already exists` (locally or on origin); `release: prepare runs on main, not <branch>`; `release: the working tree is not clean`; `release: main is not origin/main — pull or push first`; `release: check <name> failed` with the check's output on stderr; `release: claude plugin validate plugin failed` — each exit 1 except the first.
- Versions are vX.Y.Z or vX.Y.Z-<pre> (semver); the latest release is the highest tag that is a version (other tags are ignored), and a pre-release sorts before its release (v0.12.0-rc.1 < v0.12.0). The module path is read from go.mod.
- Every refusal happens before anything is written: no file, commit, branch, tag or push, and no gh call that writes. A refusal that wrote anything is a violation.
- `--dry-run` (accepted after the version, as `prepare v0.12.0 --dry-run`) makes every check, the [[check]] runs and claude plugin validate included, prints what it would do, and writes nothing.
- git, gh, go and claude are run from PATH. The tool's tests use stubs for gh, go and claude on PATH and a local bare repository as origin, in temporary directories with a temporary HOME, addressed with git -C; no test runs the real gh, go install from the network, or this repository's origin.
- Committed tests, run by the checks, cover each refusal and its order, the no-write guarantee and --dry-run.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T4/gate.sh
```

</details>

### T5 — Make prepare create the release branch and open the PR

`pending` · depends on: T4

Once its checks pass, prepare does the first outward half of a release so the operator only has to review and squash-merge: a release/<version> branch from main whose one commit, `Release <version>`, changes only plugin.json's version (one line, thanks to V1), pushed, with a PR titled `Release <version>` whose body is the notes from the latest version tag to main. The merge stays the operator's.

**Acceptance**

- prepare <version> creates release/<version> from main, sets plugin/.claude-plugin/plugin.json's version (the version without its v) changing only that line, commits `Release <version>` as the branch's one commit, pushes the branch to origin, and opens a PR with gh titled `Release <version>` whose body is `notes <latest version tag> main`.
- On success stdout ends with exactly `prepared release/<version>: <PR url>` and `squash-merge it, then: go run ./tools/release publish <version>`; exit 0.
- The tool never merges, never pushes or commits to main and never tags. Any of those is a violation.
- The tool's tests use stubs for gh, go and claude on PATH and a local bare repository as origin, in temporary directories with a temporary HOME, addressed with git -C; no test runs the real gh, go install from the network, or this repository's origin.
- Committed tests, run by the checks, cover the branch, its single commit and diff, the push to a local bare origin, the PR title and body given to a stub gh, and the output.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T5/gate.sh
```

</details>

### T6 — Make publish find the release commit, tag it and create the GitHub release

`pending` · depends on: T5

After the operator's merge, publish does the second outward half. It fetches origin and first finds the release commit — the newest on origin/main whose subject starts `Release <version>` and whose plugin.json holds the version — then tags it (annotated, message from the brief titles), pushes the tag and creates the GitHub release with the notes. Finding the commit comes before any write, so publish before the merge refuses cleanly. T7 adds the public-install check and the re-run.

**Acceptance**

- publish <version> fetches origin and finds the release commit before anything is written; none refuses `release: no release commit for <version> on origin/main — merge the release PR first` (exit 1), with no tag and no release created. Tagging before finding the release commit is a violation.
- It creates the annotated tag <version> on the release commit (not origin/main's tip), message `<version>: <brief titles joined by "; ">` from the briefs of `notes <previous version tag> <release commit>`, or `<version>` alone without briefs, and pushes it. A tag already on that commit is kept; one elsewhere refuses `release: tag <version> points at <sha>, not the release commit <sha>` (exit 1), leaving the tag as it was.
- It runs `gh release create <version> --verify-tag`, titled <version>, with `notes <previous version tag> <version>` as the notes, and --prerelease for a version with a -<pre> suffix. An existing release is kept.
- On success stdout ends with exactly the brief's four lines: `published <version>: <release url>`, `install: go install github.com/mvelosop/vloop/cmd/vloop@<version>`, `then, in each repository: vloop upgrade`, `and for the plugin: claude plugin marketplace update vloop && claude plugin update vloop@vloop`.
- publish never pushes to main and never merges. The tool's tests use stubs for gh, go and claude on PATH and a local bare repository as origin, in temporary directories with a temporary HOME, addressed with git -C; no test runs the real gh, go install from the network, or this repository's origin.
- Committed tests, run by the checks, cover the release-commit search (including a decoy subject whose plugin.json does not hold the version), the refusal before any write, the tag's commit and message, the tag elsewhere, and gh's arguments for a release and a pre-release.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T6/gate.sh
```

</details>

### T7 — Make publish prove the public install, and finish an interrupted release on re-run

`pending` · depends on: T6

A release is done only when `go install …@<version>` from the public proxy installs it: v0.11.0's install was found wanting by hand. publish waits for the proxy to serve the version, installs into a temporary GOBIN and checks what the binary reports. Because the proxy can lag, every step already done is skipped on a re-run, so an interrupted publish is finished by running it again.

**Acceptance**

- After the tag and the release, publish retries `go list -m github.com/mvelosop/vloop@<version>` every 30 seconds until it succeeds, up to --wait (default 10m), else refuses `release: the module proxy does not serve <version> after <wait>` (exit 1).
- Then `go install github.com/mvelosop/vloop/cmd/vloop@<version>` into a temporary GOBIN (never inside the repository), whose `vloop version --json` must report the version (without v) and the module (<version>); otherwise `release: go install …@<version> reports version <v>, module <m>` (exit 1).
- Each step already done prints `<step>: already done` (`tag: already done`, `release: already done`) and is skipped; a fully done release re-runs only the public-install check and ends with the four success lines, exit 0.
- `publish <version> --dry-run` makes the checks it can without writing — no tag, push or release — and prints each step.
- The tool's tests use stubs for gh, go and claude on PATH and a local bare repository as origin, in temporary directories with a temporary HOME, addressed with git -C; no test runs the real gh, go install from the network, or this repository's origin.
- Committed tests, run by the checks, cover the proxy wait and its refusal (with a short --wait), the version mismatch, the temporary GOBIN, the already-done re-runs and --dry-run.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T7/gate.sh
```

</details>

### T8 — Describe the release tool in the docs and the operator's playbook

`pending` · depends on: T7

D1: the docs still describe the manual ceremony this brief retires — five edits and a stamped -X main.version build. They must say the version comes from plugin.json and the release is prepare, the operator's merge, publish, then install and upgrade; this repository's operator playbook must make the release one short step that runs the tool; and the tool gets a short README of its own.

**Acceptance**

- docs/guide/concepts.md, "The stamped release build": the version comes from plugin/.claude-plugin/plugin.json; the stamped build sets the commit only (-X main.commit, no -X main.version); the release steps are `go run ./tools/release prepare <version>`, the operator's squash-merge, `go run ./tools/release publish <version>`, then install and `vloop upgrade` in each repository.
- .claude/skills/vloop-operator/SKILL.md gains the release as a short step that runs the tool (prepare, merge on the operator's go-ahead, publish) and records the release intervention.
- tools/release/README.md exists, under 300 words, and describes notes, prepare and publish.
- docs/guide/commands.md is regenerated if any help text changed; no vloop subcommand appears for releasing.
- No doc still tells the operator to edit the version in cmd/vloop/main.go or the tests.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T8/gate.sh
```

</details>

### T9 — Close: run the worked example line for line and the real-data check

`pending` · depends on: T8

The brief's worked example is its acceptance in one sequence: prepare's refusals, prepare, publish before and after the operator's squash-merge, the re-run, a mismatched install, and a proxy-installed v0.12.0 upgrading a repository. This task runs it end to end against a fixture and the real-data check against a clone of this repository, fixes any discrepancy it finds within the brief's contract, and leaves an end-to-end test behind so the sequence stays true.

**Acceptance**

- The task's notes carry the worked example run line for line — each command with its actual exit code and output, as run, not recalled — against a fixture repository with a local bare origin and stub gh, go and claude.
- The notes carry the real-data check on a clone of this repository: the outputs of `notes v0.10.0 v0.11.0` and `notes v0.9.0 v0.10.0`, and `prepare v0.12.0 --dry-run` on a clean, current main of the clone (whose origin is the local repository, never the real one) passing every check, with nothing written.
- A committed end-to-end test, run by the checks, plays the worked example's sequence (refusals, prepare, publish before and after a squash-merge, the re-run, the mismatched install) with stubs and a local bare origin. The tool's tests use stubs for gh, go and claude on PATH and a local bare repository as origin, in temporary directories with a temporary HOME, addressed with git -C; no test runs the real gh, go install from the network, or this repository's origin.
- Any fix made here stays within the brief's contract and names the item it serves in the notes; no check is weakened to pass.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T9/gate.sh
```

</details>


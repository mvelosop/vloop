# Plan — B20261106-2013-b38-i140-module-zips

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **0/3 done** · iteration 0

**Brief:** `docs/briefs/B20261106-2013-b38-i140-module-zips.loop-brief.md` · **Updated:** 2026-11-06T21:11:52Z

## Progress

- [ ] **T1** — Move the kit's it's repo fixtures and goldens under @REPO@, mapped back to it's repo at test time
- [ ] **T2** — Add release paths and TestRepositoryPaths: fail on any repository file path Go's module rules reject
- [ ] **T3** — Make release prepare refuse a module that does not zip, and retract v0.15.0-rc.2 in go.mod

## Tasks

### T1 — Move the kit's it's repo fixtures and goldens under @REPO@, mapped back to it's repo at test time

`pending` · depends on: none

The module proxy cannot zip v0.15.0-rc.2 because the copilot-acceptance kit commits 8 paths under a directory named `it's repo`, and Go rejects ' in a module file path. This task renames that directory to `@REPO@` in the committed tree, the way `@DIR@` stands for the temporary directory, while the kit's tests and gen.sh still create and check a real directory named `it's repo` at run time, because quoting that name is exactly what the fixture exists to test. It must land first: T2's repository test fails while any such path is tracked.

**Acceptance**

- No path under tools/copilot-acceptance/ holds the character ' (the gate checks the working tree with find).
- The 8 files moved to probes/hook/@REPO@/.keep, probes/probe/@REPO@/.keep and golden/&lt;case&gt;/@REPO@/.vloop/tmp/home/hooks/vloop-fence.json for hook-forms, hook-forms-second-review, hook-fail-missing, hook-fail-exit1, hook-fail-notjson and hook-fail-timeout, with their bytes unchanged; every other file of tools/copilot-acceptance/kit/testdata/ except gen.sh keeps its path and its bytes. A golden whose bytes changed, or any other testdata/ file changed, is a violation.
- The kit's `fixture` copy turns a path component @REPO@ into a directory named exactly `it's repo`, so the probes still run in `@DIR@/it's repo`; a kit test that no longer runs a probe in a directory named it's repo, or that stops asserting something it asserted before, is a violation.
- The hooks file the kit writes, hooks/vloop-fence.json in it's repo/.vloop/tmp/home, is still compared byte for byte with the committed golden for each of the six hook cases (wantFiles and TestHookPhase map @REPO@ to it's repo or the reverse); the kit's tests in tools/copilot-acceptance/kit pass under the go check (`go test ./...`).
- gen.sh, run on the Python of d7b9cbb5, writes the hooks goldens at their @REPO@ paths and its output equals the committed goldens (`diff -r` prints nothing); its header names d7b9cbb5 as the commit whose Python to take, in place of "the base commit's". The gate runs the starting point's gen.sh, not this one, so the review rules on the edited gen.sh from the diff.
- The kit's commands, their output and the other fixtures do not change; nothing under .vloop/ is edited.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T1/gate.sh
```

</details>

### T2 — Add release paths and TestRepositoryPaths: fail on any repository file path Go's module rules reject

`pending` · depends on: T1

Nothing before the tag looked at the module zip, so a path Go rejects reached a release. This task adds a function in tools/release that lists every tracked or untracked, not ignored, file with git and checks each with module.CheckFilePath from golang.org/x/mod v0.37.0, the `release paths` subcommand that prints its findings, and a test that runs it on this repository, so `go test ./...` (the go check, after every iteration of every run) fails the day such a path is added. It uses Go's own rules, never a hand-written list of characters, because those rules also cover cases (reserved Windows names) a copy would drift from.

**Acceptance**

- A function in tools/release takes the directory as a parameter, lists files with `git -C <top directory> ls-files -z --cached --others --exclude-standard` and checks each path with `module.CheckFilePath` from golang.org/x/mod; it fails with one line `<path>: <the error CheckFilePath returns>` per rejected path, in the order git lists them. A check that skips untracked files, or a hand-written character rule in place of CheckFilePath, is a violation.
- `release paths` prints those lines on stdout, then `release: the module zip rejects <n> path(s)` on stderr, exit 1; with none it prints nothing, exit 0; given any argument it prints `usage: release paths` on stderr, exit 2; when git fails it prints `release: <git's error on one line>` on stderr, exit 1. Paths are the top directory's even when started in a subdirectory.
- The usage line of tools/release is exactly `usage: release notes <from> <to> | paths | prepare <version> [--dry-run] | publish <version> [--wait <duration>] [--dry-run]`, and the command's doc comment lists paths.
- `TestPaths` in tools/release holds the four cases of the brief, each a fixture repository in a temporary directory addressed with git -C: a committed `it's repo/.keep` gives its line; an untracked, not ignored `a'b.txt` gives its line; the same a'b.txt listed in .gitignore gives no line; a fixture with only go.mod and x.go gives no line. No ' path is written into the repository's tree, and no test touches the network.
- `TestRepositoryPaths` in tools/release runs the function on this repository and fails with the same lines; it skips, saying why, when `git rev-parse --show-toplevel` fails or the top directory's go.mod does not declare `module github.com/mvelosop/vloop`. It passes on this repository under the go check.
- golang.org/x/mod is a direct requirement of the module at v0.37.0 and go.sum is what `go mod tidy` writes for it (the go check's `go mod tidy -diff` prints nothing).
- tools/release/README.md has a `## paths` section saying what paths prints and when it exits 1. Nothing under .vloop/ is edited.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T2/gate.sh
```

</details>

### T3 — Make release prepare refuse a module that does not zip, and retract v0.15.0-rc.2 in go.mod

`pending` · depends on: T2

Only the module proxy built the zip, after the tag was pushed, so the defect showed up at the public install. This task makes `release prepare` build the module zip of HEAD with zip.CreateFromVCS before anything is written, right after the "main is not origin/main" check and before the configured checks, so a release whose module does not zip is never opened. It also retracts v0.15.0-rc.2 in go.mod, written now so it is not forgotten though it only takes effect once v0.15.0 is out.

**Acceptance**

- `prepareChecks` builds the module zip of HEAD for the module path go.mod names and the version asked, with `zip.CreateFromVCS` from golang.org/x/mod v0.37.0, written to io.Discard, after "main is not origin/main" and before the configured checks; it writes nothing to the repository. A zip check that runs after the configured checks, or one that writes a file, is a violation.
- When errors.As finds a zip.FileErrorList in the error, stderr holds one line `<path>: <error>` per rejected file with the path relative to the repository, then `release: the module zip rejects <n> path(s)`, exit 1; neither the configured checks nor `claude plugin validate` run after it, and the fixture's refs, HEAD and status are unchanged.
- Any other error refuses with `release: the module does not zip: <the error on one line, every occurrence of the repository's absolute directory replaced by .>`, exit 1; in a linked worktree that is exactly `release: the module does not zip: create zip from version control system .: could not find a recognized version control system at "."`. A module that zips still passes `--dry-run`, which runs the check like every other.
- `TestPrepareRefusals` in tools/release gains the case of a committed and pushed `it's repo/.keep`, made in a temporary directory, asserting both stderr lines, exit 1, and that the configured check and claude did not run; it passes under the go check.
- go.mod ends with the existing retraction, a blank line, then `// v0.15.0-rc.2 holds file paths the module zip rejects and cannot be`, `// downloaded.`, `retract v0.15.0-rc.2`; `release prepare v0.15.0-rc.2 --dry-run` in this repository refuses with `release: v0.15.0-rc.2 is retracted in go.mod`, exit 1.
- tools/release/README.md's prepare paragraph reads "…`main` clean and equal to origin, the module zip of `HEAD`, the config's checks…". `release publish` and every other refusal of prepare are unchanged; nothing under .vloop/ is edited.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T3/gate.sh
```

</details>


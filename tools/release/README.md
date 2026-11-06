# tools/release

This repository's release tool, run as `go run ./tools/release <subcommand>`.
It is not a vloop subcommand. The version lives in
`plugin/.claude-plugin/plugin.json`; nothing else holds it.

## notes

`notes <from> <to>` prints the release notes for the range `from..to` on
stdout: the briefs merged in the range and the other changes.

## paths

`paths` lists the repository's files, tracked or untracked and not ignored,
and checks each path with `module.CheckFilePath` from `golang.org/x/mod`. It
prints one `<path>: <error>` line on stdout per rejected path, then
`release: the module zip rejects <n> path(s)` on stderr, and exits 1; with none
it prints nothing and exits 0. Any argument is a usage error (exit 2); a git
failure exits 1. `TestRepositoryPaths` runs it on this repository.

## prepare

`prepare <version> [--dry-run]` makes every check before writing anything:
module major, retraction, version after the latest tag, tag absent locally and
on origin, `main` clean and equal to origin, the config's checks, and
`claude plugin validate`. It then creates `release/<version>`, commits the
version change to `plugin.json` and opens the PR. `--dry-run` makes the checks
and prints what it would do.

## publish

`publish <version> [--wait <duration>] [--dry-run]`, run after the operator's
merge, finds the release commit on `main`, tags it (annotated), pushes the
tag and creates the GitHub release. It then waits for the module proxy and
installs the version into a temporary directory to prove the public install.
Done steps are skipped, so a re-run finishes an interrupted release.

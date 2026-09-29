# Journal — B20260929-1804-vloop-skeleton-config-briefs

Append-only narrative of this plan. Rendered state lives in .loop/state/plan.md.

## Plan — B20260929-1804-vloop-skeleton-config-briefs

- **Brief:** `docs/briefs/B20260929-1804-vloop-skeleton-config-briefs.loop-brief.md`
- **Tasks:** 9

The plan is written to `.loop/state/state.json`. It's valid JSON, passes `.loop/amend.sh check`, and has no reference that fails to resolve. Every verify command fails right now for the expected reason: there is no Go module yet, so `go vet ./...` fails.

- **Run id:** `B20260929-1804-vloop-skeleton-config-briefs`
- **Tasks:** 9, in a straight chain T1 → T9, following the brief's Shape list.
- **First ready task:** T1, the Go module skeleton (command tree, global flags, exit codes, `version`, embedded plugin).

**Gate structure.** Each gate runs the same checks first: `gofmt -l .`, `go vet ./...`, and builds for linux, windows and the host. It then runs the task's committed Go tests and confirms at least one test exists. Last, it builds the binary and runs its own checks against it in a temporary git repo: exact output lines, exit codes, stdout versus stderr, and JSON checked field by field with `jq`. The brief fixtures in the gates are complete, valid briefs, so a later task's rules can't make an earlier gate fail. I ran every English fixture through `.loop/check-brief.sh` and each rule fired, or stayed silent, as intended. Only T9 runs `go test ./...`, and it also checks that the test suite wrote nothing under `.loop/` or `.claude/`.

**Limit:** I could only run each gate as far as its first step, since nothing builds yet. Past that point the gates are syntax-checked, and the fixtures are checked against the shell checker, but the rest of each script has never actually run. If a gate turns out to be wrong, the work session should stop and report it rather than edit the check.

**Names I pinned because a gate needs them:**
- `main.version` and `main.commit` in `cmd/vloop`, as the `-ldflags -X` targets.
- The `internal/config` package (the design session already names `internal/cli` and `internal/brief`). The gates run tests per package: T1 runs the root package, `./cmd/vloop` and `./internal/cli`; T2 runs `./internal/config` and `./internal/cli`; T3–T7 run `./internal/brief` and `./internal/cli`.
- The README test lives in `internal/cli`, and the end-to-end test lives in `./cmd/vloop`.
- The file names in each task's `files` list are what I expect the task to leave behind, not something the gates enforce.

**Where I had to interpret the brief:**
1. **No base SHA.** The brief says to pin its starting commit as the base every gate compares against. The driver rejects gates that diff against a fixed commit, so no gate diffs at all; they check the binary's behaviour instead.
2. **Wording of ported messages.** I required `check-brief.sh`'s exact wording for the ported rules, only swapping the journal path to `.vloop/`, and the gates match on key phrases. In `--json`, messages carry no `✓`/`!`/`✗` marker.
3. **`brief list` order.** I read "dependency order, ties by name" as: a brief is listed as soon as its dependencies are, lowest name first. A brief with no dependencies shows `ready`.
4. **Errors with `--json`.** The brief says stdout carries one JSON document even on exit 1. I gated that for config errors and for `brief new` when the file exists, but left the document's shape open.
5. **`config set` output.** It prints nothing on success, matching the worked example.
6. **Naming the source of a bad value.** An invalid value from the environment must name the variable (`VLOOP_LANGUAGE`); one from the file must name `config.toml`.
7. **Not gated:**
   - An empty `VLOOP_MODEL_*` variable, which conflicts with `set <key> ''` meaning "remove the key".
   - Whether dependency rules apply to draft briefs.
   - Which brief the cycle message starts from in `brief list`; the gate accepts either.
8. **README checks.** Beyond the 200-line cap, T8's gate checks that every `--flag`, backticked `vloop <cmd>`, config key and `VLOOP_*` variable in the README exists in the binary.

I left the generator and fixture scripts in `.loop/tmp/`, which is gitignored.


## T1 — Go module skeleton: command tree, global flags, exit codes, version, embedded plugin

- **Outcome:** done (review: PASS)
- **Summary:** Go module github.com/mvelosop/vloop now exists with a cobra tree (global flags, 0/1/2 exit mapping), `vloop version` (text and --json), and the plugin skeleton embedded via `//go:embed all:plugin`.
- **Files:** go.mod, go.sum, embed.go, embed_test.go, cmd/vloop/main.go, internal/cli/root.go, internal/cli/version.go, internal/cli/root_test.go, plugin/.claude-plugin/plugin.json, .claude-plugin/marketplace.json
- **Notes for next iteration:** cli.Execute(Build, args, stdout, stderr) returns the exit code. Any error returned from a command is exit 2 (usage) unless wrapped with cli.Problem(err), which gives exit 1; later commands must wrap their 'found problems' errors. Root has no Args validator, on purpose: cobra only reports 'unknown command' for a non-runnable root when Args is nil. Cobra's usage and error printing are silenced and Execute prints the single 'vloop: ' line itself, first line of the error only. Globals (Dir, JSON, NoColor, Quiet, Verbose) are returned from NewRoot and have a Color(w) helper covering NO_COLOR and TTY detection. It is untested, since no command colours output yet. main.version and main.commit are passed into cli as cli.Build. Cobra v1.10.2, go directive from go 1.27.1.

## T2 — Repo root resolution and `vloop config get|set|list|path`

- **Outcome:** done (review: PASS)
- **Summary:** Added internal/config (root discovery, key table, env/file/default resolution, Set with removal and byte-for-byte idempotence) and the `vloop config list|get|set|path` commands.
- **Files:** internal/config/config.go, internal/config/root.go, internal/config/config_test.go, internal/config/root_test.go, internal/cli/config.go, internal/cli/config_test.go, internal/cli/root.go, go.mod, go.sum
- **Notes for next iteration:** TOML lib is BurntSushi/toml. Errors: config.SourceError (bad env/file value or malformed file) -> cli.Problem exit 1; UnknownKeyError/InvalidValueError -> exit 2. Under --json a SourceError also prints {"error":"..."} to stdout (shape was unspecified). Empty VLOOP_* env vars count as unset. `get` resolves only the requested key, `list` resolves all. Set re-encodes the whole file via the TOML encoder (comments/formatting are lost on a real change) but skips the write when the key already has that value, or removal of an absent key (no file created). Globals.root() in internal/cli/config.go resolves the repo root from -C or cwd; later commands should reuse it. Small edit to internal/cli/root.go to register newConfig.

## T3 — Brief parsing and the English `vloop brief check` (port of check-brief.sh)

- **Outcome:** done (review: PASS)
- **Summary:** Added internal/brief (frontmatter status parse, English heading set as data, every rule of check-brief.sh ported) and `vloop brief check` with text and --json output.
- **Files:** internal/brief/brief.go, internal/brief/check.go, internal/brief/check_test.go, internal/brief/testdata/en/pass.loop-brief.md, internal/cli/brief.go, internal/cli/brief_test.go, internal/cli/root.go
- **Notes for next iteration:** brief.HeadingSet holds BindingRefs, WorkedExample, OutOfScope, Constraints plus TaskCount/ExitCodes regexes (matched case-insensitively); T4 adds a Spanish entry to the `sets` map in brief.go (SetFor currently falls back to English for unknown languages; T4 should make cross-language strict). Headings match as `^#+ .*<name>` like the shell. Binding-references section is stripped before the unresolved-path scan. Windows home rule: `[A-Za-z]:\\+Users\\`. Status other than ready (or absent) is skipped; absent shows `status is missing`; T5 owns frontmatter validity. A brief that fails silently exits 1 via Problem(errors.New("")); Execute in root.go now skips the stderr line when the message is empty. Missing paths also appear in --json as result problems with message `no such brief: <path>`. Fixtures: one passing file in testdata/en plus per-rule mutations in check_test.go. Paths are resolved relative to -C/cwd and printed root-relative.

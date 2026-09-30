# Journal — B20260930-2007-vloop-init-upgrade-doctor

Append-only narrative of this plan. Rendered state lives in .loop/state/plan.md.

## Plan — B20260930-2007-vloop-init-upgrade-doctor

- **Brief:** `docs/briefs/B20260930-2007-vloop-init-upgrade-doctor.loop-brief.md`
- **Tasks:** 10

The plan is written to `.loop/state/state.json`. **Run id:** `B20260930-2007-vloop-init-upgrade-doctor`. It has **10 tasks**, and the first ready one is **T1** (T2, T3 and T4 are also ready from the start).

| Task | What it builds | Depends on |
| --- | --- | --- |
| T1 | The `install/v1` schema, reading and writing the stamp, and version comparison | — |
| T2 | F1: a binding reference's reason may wrap onto the next line | — |
| T3 | Valid manifests, the SessionStart hook, `version --check-plugin`, `plugin path` | — |
| T4 | F2: stacks scoped by path | — |
| T5 | Stack detection | T4 |
| T6 | `vloop init` | T1, T4, T5 |
| T7 | `vloop upgrade` | T1, T6 |
| T8 | `vloop doctor` | T1, T3, T4, T6 |
| T9 | Docs | T2–T8 |
| T10 | Close: end-to-end test of the worked example, the real-data check, and the real `claude plugin validate .` | all |

**Checks I ran:**
- **The JSON is valid**, and every task has a goal, acceptance criteria and a `verify` command. Every dependency points to a real task, and there are no cycles.
- **References:** all 11 cited paths exist.
- **No absolute paths:** none of the user kind appear anywhere. System paths were replaced with a PATH built from symlinks.
- **Shape checks:** `.loop/amend.sh check` passes with no warnings.
- **Every gate fails now, for the right reason.** I ran each stored `verify` the way the driver does.
- **The gates' fixtures are sound:**
  - I built the base-commit binary and confirmed it still has the F1 and F2 defects.
  - The T5 detection harness passes against a throwaway detector in a scratch copy, and catches a deliberately broken one.
- **Not checked:** I'm not allowed to run `claude`, so the `claude plugin validate .` step in T10 has never actually run.

**Where I had to interpret the brief (please check these before running):**
1. **Format of `claude plugin list --json`.** I assumed it prints an array of `{"id":"vloop@vloop","version","scope","enabled"}` objects. T8's stub and doctor's plugin check both depend on this, and I couldn't confirm it. Likewise, T10 treats any "warn"/"error" line from `claude plugin validate .` as a failure unless it reads "0" or "no".
2. **Extracted `plugin.json` carries the binary's version.** The embedded manifest says `0.0.0-dev`, so the worked example (`--check-plugin` on the extracted copy prints nothing at 0.2.0) only works if `plugin path` writes the binary's version into the copy.
3. **The CLAUDE.md section names the vloop version that wrote it.** Otherwise the 0.1.0 → 0.1.1 upgrade has nothing to change, yet the example prints `updated CLAUDE.md`.
4. **The worked-example repo already has a `.gitignore`**, because the example prints `updated .gitignore`. When init creates one, it prints `wrote .gitignore`.
5. **Doctor's `stacks` line appears only when a stack is scoped.** The example's doctor output has no stacks line.
6. **A scoped stack whose directory has vanished is still a readable config.** `config set` rejects it, but reading it is not an error. Otherwise doctor's `stacks` warning could never appear next to `✓ config`.
7. **Smaller calls:**
   - A pre-release version, in either direction, gets the same "may break" refusal.
   - `typescript` counts as a dependency when it's in `devDependencies`.
   - The trust check reads `$HOME/.claude.json`, the same file `run.sh` preflight uses.
   - Self-hosting means `go.mod` declares `github.com/mvelosop/vloop` and the binary's commit is a prefix of HEAD.
   - `--check-plugin` echoes the directory as given, and a manifest with no version counts as "cannot read".
   - `upgraded` is a required field that may be null.
   - The invalid-value message names the one bad stack entry, not the whole list.
8. **F1 is missing from the brief's Shape list**, so I made it its own task, T2. That keeps the plan within the brief's 9–11.
9. **The base SHA (05ea997) is only used to build a comparison binary** in the T2 and T4 gates, via `git archive`. Nothing diffs against it, because the driver rejects gates whose `git diff` baseline would go stale.
10. **README.md is at its line cap: 199 lines, and the test requires fewer than 200.** Each task that adds a command has to tighten existing README text; the goals say the cap must not be raised.

**Names I pinned that the brief doesn't give:**
- Package `internal/install` and `internal/detect`, with `detect.Stacks(root string) ([]string, error)`. The T5 gate calls that function directly.
- Test-name patterns the gates look for: `Wrap`, `Hook`, `Scope`, `Init`, `Upgrade`, `Doctor`, `TestWorkedExampleB5`.
- File names: `plugin_embed_test.go` and `cmd/vloop/b5_e2e_test.go`.
- The `--json` output of `plugin path` is `{"path": …}`.

As the brief requires, each gate runs `go test` only on the packages its task touches (plus vet, gofmt, tidy and the three builds); only T10 runs `go test ./...`. Most gates also run the `internal/cli` and `cmd/vloop` suites, which covers regressions across tasks.


## T1 — Add the install/v1 schema, the install stamp's read and write, and semantic-version comparison

- **Outcome:** done (review: PASS)
- **Summary:** Added schemas/install.v1.json with fixtures, and internal/install: stamp Read/Write for .vloop/install.json plus Parse/Compare, the one semver comparison for upgrade and doctor.
- **Files:** schemas/install.v1.json, internal/install/install.go, internal/install/install_test.go, internal/schema/schema_test.go, internal/schema/testdata/install.valid.json, internal/schema/testdata/install.invalid.json, embed_test.go, internal/cli/schema_test.go, cmd/vloop/b2_e2e_test.go
- **Notes for next iteration:** Compare(from, to) returns a Relation: Same, Newer, Breaking, FromNewer or NeedsYes. A pre-release suffix on either side gives NeedsYes, unless the strings are identical (Same). Unparsable versions give an error. Stamp.Upgraded is *string, written as null; Read returns an error that satisfies errors.Is(err, os.ErrNotExist) when there is no stamp. Write always sets schema to install/v1 and ends the file with a newline. Besides the schema list expectations and the embed count, internal/cli/schema_test.go's --json length check went from 8 to 9. No third-party module added.

## T2 — Fix F1: a binding reference's reason may wrap onto the lines indented under its list item

- **Outcome:** done (review: PASS)
- **Summary:** brief check now treats a binding-reference entry as the whole list item (first line plus indented continuation lines), so a reason wrapped onto the next line passes; reasons never carry to the next item.
- **Files:** internal/brief/refs.go, internal/brief/refs_test.go
- **Notes for next iteration:** New refEntries groups section lines: an item ends at a blank line, a non-indented line, or the next entry line. Continuation lines are joined with a space onto the text after the path, then matched by the unchanged reasonSepRe. The separator may even start the continuation line. Tests: TestBindingRefsWrap, TestBindingRefsWrapNoReason, TestBindingRefsWrapNextItemBoundary. No testdata files added.

## T3 — Complete the plugin: valid manifests, the SessionStart hook, `vloop version --check-plugin` and `vloop plugin path`

- **Outcome:** done (review: PASS)
- **Summary:** Completed the plugin manifests, added the SessionStart hook, `vloop version --check-plugin <dir>` (always exit 0, stdout only) and `vloop plugin path` (idempotent extraction under .vloop/tmp/plugin/<version>/, symlink-safe); README tightened to 196 lines and commands.md regenerated.
- **Files:** plugin/.claude-plugin/plugin.json, plugin/hooks/hooks.json, .claude-plugin/marketplace.json, plugin_embed_test.go, internal/cli/version.go, internal/cli/plugin.go, internal/cli/plugin_test.go, internal/cli/root.go, README.md, docs/guide/commands.md
- **Notes for next iteration:** plugin path walks each directory component of .vloop/tmp/plugin/<version> with Lstat and refuses a symlink or file in its place (exit 1); stray removal uses WalkDir + RemoveAll, which never follows links. The extracted plugin.json is re-marshalled (indented, key order alphabetical) with the binary's version, so it is not byte-identical to the embedded one. README: merged the writes-list sentence and dropped a blank line to make room; 'or result' was dropped from the 'refuses a plan' sentence. embed.go needed no change (all:plugin already covers hooks/).

## T4 — Fix F2: stacks scoped by path — parse and validate `stack@path`, resolve scopes, match scope-relative, label scoped matches

- **Outcome:** done (review: PASS)
- **Summary:** metrics.stacks entries may now be <stack>@<path>: classify.ParseStack checks the form, config validates entries (set also requires the directory to exist), and the classifier resolves the longest containing scope, matches scoped presets scope-relative and labels them stack@path.
- **Files:** internal/classify/classify.go, internal/classify/classify_test.go, internal/config/config.go, internal/config/config_test.go
- **Notes for next iteration:** Key gained a Stacks flag (replaces Valid for metrics.stacks); InvalidValueError for it carries the single bad entry and the message 'want <stack> or <stack>@<existing directory>'. Existence of the directory is checked only in Set (checkScopeDirs), not on read. Classifier stores scope per preset; the label is in Result.Layer, so internal/cli/metrics.go and cli/config.go needed no change. Outside all scopes the unscoped stacks still apply; inside one, only that scope's. Backslashes in a scope path are rejected.

## T5 — Detect stacks per directory: the marker table, the C# solution rule, skipped directories, sorted scoped output

- **Outcome:** done (review: PASS)
- **Summary:** Added internal/detect with Stacks(root): marker table, C# solution rule, skipped directories, depth 3, symlinks never followed, sorted unscoped-first output.
- **Files:** internal/detect/detect.go, internal/detect/detect_test.go
- **Notes for next iteration:** Marker filenames are matched case-insensitively. Only regular files count as markers, and symlinked directories are not descended. C#: a *.sln marks its own directory and suppresses csproj entries in every directory below it; a *.csproj with no solution above or beside it marks its own directory. An unreadable or malformed package.json still yields javascript. Output is sorted by the full entry string within each group. The root not being readable returns an error; unreadable subdirectories do too.

## T6 — Add `vloop init`: config with detected stacks, the stamp, a starter brief, the .gitignore line and the CLAUDE.md section

- **Outcome:** blocked (review: skipped)
- **Summary:** Implemented `vloop init` (internal/cli/init.go, internal/install/section.go, config.Check, Init tests, README row at 197 lines, commands.md regenerated); go test/vet/gofmt/builds and every gate step before the first worked-example assertion pass, but the gate itself fails at 'init must not commit or stage'.
- **Files:** internal/cli/init.go, internal/cli/init_test.go, internal/cli/root.go, internal/install/section.go, internal/config/config.go, README.md, docs/guide/commands.md
- **Notes for next iteration:** The gate is wrong, not the code. The fixture (ex) commits .gitignore ('node_modules/'), and the same gate later requires init to append .vloop/tmp/ to it. A tracked file edited in the working tree always shows as ' M .gitignore' in `git status --porcelain --untracked-files=no`, so that check (`-z` on its output) cannot pass for a correct init. HEAD equality and the refs check are fine. It should assert nothing is staged, e.g. `git diff --cached --quiet`, instead of an empty porcelain status. Remaining gate steps after that line were not reached, so they are unverified by the gate. Design: init validates and plans all five writes first, refuses symlinked targets (exit 1) before writing anything; existing-config case prints 'suggest metrics.stacks: ...' for detected stacks the config lacks; starter brief language comes from --language, else the existing config's language. Shared helpers for T7 are install.Section, install.MergeClaudeMD, install.MergeGitignore.

## T6 — Add `vloop init`: config with detected stacks, the stamp, a starter brief, the .gitignore line and the CLAUDE.md section

- **Outcome:** blocked (review: skipped)
- **Summary:** `vloop init` is already implemented and on disk (internal/cli/init.go, internal/install/section.go, Init tests, README at 197 lines, commands.md regenerated). go test on the touched packages passes. The verify gate cannot pass for a correct implementation.
- **Files:** internal/cli/init.go, internal/cli/init_test.go, internal/cli/root.go, internal/install/section.go, internal/config/config.go, README.md, docs/guide/commands.md
- **Notes for next iteration:** Gate is wrong, not the code. The fixture `ex` commits .gitignore ('node_modules/'), and the gate requires init to append .vloop/tmp/ to it (and the acceptance requires 'updated .gitignore'). The check `[ -z "$(g status --porcelain --untracked-files=no)" ]` must therefore fail for any correct init, because a modified tracked file is always listed. Fix in the plan: replace that clause with `g diff --cached --quiet` (nothing staged); the HEAD-equality and for-each-ref checks stay. Gate steps after that line have not been exercised against the implementation; the operator should re-run the corrected gate. Shared helpers for T7: install.Section, install.MergeClaudeMD, install.MergeGitignore.

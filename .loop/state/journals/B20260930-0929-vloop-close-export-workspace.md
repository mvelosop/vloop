# Journal — B20260930-0929-vloop-close-export-workspace

Append-only narrative of this plan. Rendered state lives in .loop/state/plan.md.

## Plan — B20260930-0929-vloop-close-export-workspace

- **Brief:** `docs/briefs/B20260930-0929-vloop-close-export-workspace.loop-brief.md`
- **Tasks:** 8

I've written the plan to `.loop/state/state.json`: run `B20260930-0929-vloop-close-export-workspace`, 8 tasks. T1 is the first ready task.

**Checks:**
- All 8 verify commands fail right now, each because the work is missing, not because the gate is broken.
- A gate that stops at its first check proves little, so I also:
  - re-ran them with the "committed tests exist" check stubbed out, to exercise the rest;
  - played T3's close by hand in a fixture, which passed its success checks;
  - built the records a correct export would print and ran T5's and T8's checks on them, which passed;
  - compiled T2's gate-only test against a sample splice function, which passed.
- `.loop/amend.sh check` is clean, every reference path exists, there are no absolute user paths, and no ref in this repo moved.
- I didn't time the gates. T3 (five fixture repos) and T7 (a copy of the working tree) are the ones most likely to near the one-minute limit.

**Tasks:** T1 export/v1 schema and `abandoned` status → T2 run-record splice and snapshot (library only) → T3 `brief close` → T4 `--abandon` / `--dry-run` → T5 `metrics export` → T6 `--workspace` → T7 guide pages → T8 end-to-end worked example plus the real-data check.

**Where I interpreted rather than read** — please correct any before the run starts:
1. **No base SHA.** The brief says the planner pins one for gates to compare against, but the loop's rules reject gates that diff against a fixed commit. Every gate builds its own fixture repos in temp dirs instead.
2. **The worked example's jq line for export ids** (`.run_id // .id`) would print the run id for task records too, because the brief gives them a `run_id` field. I kept the field list and check the id per record type.
3. **`--dry-run` prints the summary the real close would print:** status `consumed`, and the pending finding counted as an operator defect.
4. **Derived defects:**
   - They appear in the run record's defect list as well as recorded ones.
   - The export gives them `severity: null` and `status: null`.
   - Review-finding ids count from 1 (`…/i2-review-1`).
5. **`close --json`:** `trailer` is the whole `Vloop-Brief: <name>` line, and `files` are in the order written.
6. **Workspace:** a `name` override also applies to the `repo` column of the table, and `--workspace` works from a directory outside any git repo (the example runs from the parent of `portfolio/`).
7. **`--abandon`** is accepted even when the plan is complete.
8. **README cap:** README.md is at 191 of B1's 200-line limit. Tasks 3–7 must make room by moving detail into the guide, never by raising the limit.

**Names I pinned so the gates can reference them:**
- A new package `internal/closing` with `PutRunRecord(doc, record string) string`. T2's gate tests it through an in-memory `go test -overlay` file; nothing is written to the repo.
- Committed test name prefixes: `TestRunRecord`, `TestSnapshot`, `TestClose`, `TestCloseDryRun`, `TestCloseAbandon`, `TestExport`, `TestWorkspace`, `TestCommandsReference`, `TestWorkedExampleB4`.
- `go generate ./...` as the one command that regenerates `docs/guide/commands.md`.
- Test files: `internal/cli/close_test.go`, `close_flags_test.go`, `export_test.go`, `workspace_test.go`, and `cmd/vloop/b4_e2e_test.go`.
- The task record reuses `by_task`'s keys (`churn`, `agent_ms`, `cost_usd`, `models`).

**Real-data dependency:** T8's check needs this clone's `origin` to be `https://github.com/mvelosop/vloop.git`. It also expects B4's own in-progress brief to show up in the export and the workspace table, so it only requires records for B1–B3.


## T1 — Add the `export/v1` schema and the `abandoned` brief status

- **Outcome:** done (review: PASS)
- **Summary:** Added schemas/export.v1.json (brief, task and defect records via if/then on `type`) with three fixture pairs, and the `abandoned` brief status: accepted by frontmatter, listed as `-` in brief list, skipped by brief check, blocking dependents; README and both templates mention it.
- **Files:** schemas/export.v1.json, internal/schema/schema_test.go, internal/schema/testdata/export-brief.valid.json, internal/schema/testdata/export-brief.invalid.json, internal/schema/testdata/export-task.valid.json, internal/schema/testdata/export-task.invalid.json, internal/schema/testdata/export-defect.valid.json, internal/schema/testdata/export-defect.invalid.json, internal/brief/frontmatter.go, internal/brief/deps.go, internal/brief/frontmatter_test.go, internal/brief/check_test.go, internal/brief/templates/en.md, internal/brief/templates/es.md, internal/cli/schema_test.go, cmd/vloop/b2_e2e_test.go, README.md
- **Notes for next iteration:** Brief-record required keys are restated in export/v1 (no cross-schema $ref; the compiler loads nothing), with property types only for scalars and `type: object` for nested blocks. Invalid brief fixture omits `cost`, so its one violation is at pointer "". schema_test.go's fixture map gained export-brief/-task/-defect keys; the schema name is the part before the dash. `consumed` is special-cased only in brief/deps.go, brief/frontmatter.go and merge attribution (runs/release.go, defect/blame.go); the latter two were left alone since abandoned briefs are never merged-as-consumed, and metrics/defect add already handle any status. No new deps_test case for abandoned was added; the gate covers it end to end.

## T2 — Build the run-record renderer, the marker splice and the metrics snapshot writer

- **Outcome:** done (review: PASS)
- **Summary:** New package internal/closing with PutRunRecord, Render (run record) and WriteSnapshot; the summary/table formatters moved from internal/cli to internal/metrics (PrintSummary, PrintByTask, PrintBriefTable) and metrics.DerivedIDs computes derived defect ids.
- **Files:** internal/closing/closing.go, internal/closing/closing_test.go, internal/metrics/print.go, internal/metrics/defects.go, internal/cli/metrics.go, internal/cli/metrics_print.go
- **Notes for next iteration:** internal/cli/metrics_print.go was git-mv'd to internal/metrics/print.go (cli would be an import cycle for closing/close); cli/metrics.go now calls metrics.Print*. A marker line matches with trailing spaces/tabs/CR tolerated, and is preserved since slicing is by offset. closing.Render takes a Record{Name, Date, Report, Derived, Recorded}; it sorts derived by iteration (stable) and recorded by id, and derives ids via metrics.DerivedIDs(runID, ds) (T5 export must call it on the same ordered list). WriteSnapshot(root, report) returns the repo-relative path; tests use testdata/metrics.valid.json as the report and schema.Validate(metrics.Schema, ...).

## T3 — Add `vloop brief close`: findings, refusals, status lines, run record, snapshot and one commit

- **Outcome:** done (review: PASS)
- **Summary:** Added `vloop brief close` (internal/cli/close.go): six refusals, defect files via defect.Add, metrics snapshot, run record, consumed status, and one path-staged commit with the Vloop-Brief trailer; README documents it.
- **Files:** internal/cli/close.go, internal/cli/close_test.go, internal/cli/brief.go, internal/metrics/defects.go, README.md
- **Notes for next iteration:** Added metrics.DeriveBrief(root, brief) so the run record's derived defects match Build's counts; T5 export can reuse it. Report is built after the defect files exist, then Status is overridden to 'consumed' in memory (Build reads the on-disk status, which is still 'ready' at that point). Commit is `git add -- <files>` then `git commit -m subject -m trailer -- <files>`; the dirty-tree refusal guarantees a clean index first. 'Plan complete' means the last owned run commit's outcome is 'complete'. Default-branch check strips refs/remotes/origin/ from runs.DefaultBranch and is skipped if there is no default branch or HEAD is detached. --abandon and --dry-run are not added (T4). Non-test helper is named runGit because the cli tests already define gitIn.

## T4 — Add `--abandon "<reason>"` and `--dry-run` to `vloop brief close`

- **Outcome:** done (review: PASS)
- **Summary:** `vloop brief close` gains `--abandon "<reason>"` (closes an unfinished or finished plan as abandoned) and `--dry-run` (prints the summary with pending findings counted as operator defects plus would-lines, writes nothing); README documents both.
- **Files:** internal/cli/close.go, internal/cli/close_flags_test.go, README.md
- **Notes for next iteration:** Metrics are now built before defect files are written; a real close with findings rebuilds after Add so counts include them. Dry-run predicts defect paths in close.go (same D<stamp>-<slug>[-n] scheme as defect.Add, checking existing files) and adjusts Defects counts and removal efficiency in memory. An empty --abandon reason is rejected (exit 1) with its own message. --dry-run ignores --json and prints text. README stays at 192 lines (row extended, no new lines).

## T5 — Add `vloop metrics export`: JSON Lines of briefs, tasks and defects with the repository's identity

- **Outcome:** done (review: PASS)
- **Summary:** Added `vloop metrics export [<brief>…]`: export/v1 JSON Lines of brief, task and defect records with repo identity (origin credentials stripped), read-only; README row added.
- **Files:** internal/cli/export.go, internal/cli/export_test.go, internal/cli/metrics.go, internal/metrics/defects.go, README.md
- **Notes for next iteration:** Brief record is the metrics.Report marshalled, by_task deleted, then schema/type/repo set. Task records follow plan order (metrics.PlanOf, new, reads the plan at the last run commit) with by_task rows joined by id; DeriveBrief needs the .md path (Report.Brief has no .md). Origin read via `git config --get remote.origin.url`; only scheme:// authorities are stripped, scp-style git@host:path is left alone. README now 193 lines.

## T6 — Add `--workspace <file>` to `vloop metrics` and `vloop metrics export`

- **Outcome:** done (review: PASS)
- **Summary:** Added --workspace <file> to `vloop metrics` and `vloop metrics export`: a TOML file of [[repo]] tables (path relative to the file, optional name) gives a table with a leading repo column, or concatenated exports with repo.name overridden; missing or non-git paths print one stderr line each and exit 1 after the rest is reported; --workspace with briefs is exit 2.
- **Files:** internal/cli/workspace.go, internal/cli/workspace_test.go, internal/cli/metrics.go, internal/cli/export.go, internal/metrics/print.go, README.md
- **Notes for next iteration:** Export body moved into exportRepoLines(g,out,root,name,args); the metrics table is metrics.PrintWorkspaceTable, sharing printBriefTable with PrintBriefTable. A repo counts as a git repository when <path>/.git exists (dir or file); a plain directory inside another repo is therefore not found. The unnamed repo column uses repoIdentity(root).Name (origin basename, else directory name). Missing paths are reported on stderr directly and the command returns Problem(errors.New('')) (empty message: Execute prints nothing more, exits 1). --workspace with --json emits the concatenated reports array without a repo field. A relative workspace file path honors -C. README stays at 193 lines.

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


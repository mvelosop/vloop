# Journal — B20260929-2148-vloop-state-tasks-status

Append-only narrative of this plan. Rendered state lives in .loop/state/plan.md.

## Plan — B20260929-2148-vloop-state-tasks-status

- **Brief:** `docs/briefs/B20260929-2148-vloop-state-tasks-status.loop-brief.md`
- **Tasks:** 9

The plan is written to `.loop/state/state.json`.

- **Run id:** `B20260929-2148-vloop-state-tasks-status`
- **Tasks:** 9, following the brief's order.
  - T1–T6 form a chain: schemas → plan load/save + `status` → `shell`/`areas` + `task list|show` → `task validate` → `reset|note|drop|set` → `task verify|gate`.
  - T7 (F1) depends on nothing.
  - T8 (README) depends on T6.
  - T9 (close) depends on T7 and T8.
- **First ready task:** T1. T7 is also ready.

**Checks I ran:**
- Every verify command fails right now, each for the right reason: a missing module, package or test, or an unknown `status`/`task` command. F1 fails on exactly the defect: `b.loop-brief -> a.loop-brief -> b.loop-brief` where `a.loop-brief -> …` is wanted.
- The JSON is valid, every reference path exists, there are no absolute paths and no dependency cycles.
- `.loop/amend.sh check` reports nothing.
- I copied the driver's four gate-shape rules out of `run.sh` and ran them over every gate; none is flagged.
- Each gate checks the built binary from outside, in a scratch repo, and also runs that task's committed tests.

I left my gate-drafting scratch in `.loop/tmp/gates/`. It's git-ignored, and my attempt to delete it was denied.

**Names I pinned (the brief doesn't name them):**
- Schema files `schemas/<name>.v1.json` and packages `internal/schema` and `internal/state`.
- `cmd/vloop/b2_e2e_test.go`.
- Test-name prefixes the gates look for: `TestSchema*`, `TestStatus*`, `TestTaskList*`, `TestTaskShow*`, a config test containing `Areas`, `TestTaskValidate*`, `TestTask{Reset,Note,Drop,Set}*`, `TestTaskVerify*`, `TestTaskGate*`, `TestCycleStart*`, `TestWorkedExampleB2*`.
- In `session/v1`:
  - `models_used` is an object keyed by model id, each value `{input_tokens, output_tokens, cache_read_input_tokens, cache_creation_input_tokens, cost_usd}`. That is the snake_case form of Claude's `modelUsage`.
  - `permission_denials` is an array.
- Task ids match `^T[0-9]+$`.

**Where I interpreted the brief rather than read it — please check these before the run:**
1. **Base SHA.** The base is `00643e4`. The shell loop's plan format has no field for it, and the driver rejects gates that diff against a fixed commit, so no gate uses it.
2. **README test skipped until T8.** B1's README test fails as soon as T1 adds a command the README doesn't cover. So T1–T7 run `internal/cli` tests with that one test skipped. T8 and T9 run it in full, and T8 also checks that `readme_test.go` is byte-for-byte unchanged.
3. **A second B1 test changes.** `internal/cli/config_test.go` pins the `config list` output too, so T3 may update it alongside `cmd/vloop/e2e_test.go`. No other assertion in either file may change.
4. **`status` on an invalid plan.** On a plan that fails its schema, `status` prints normally and exits 0 ("read commands still work where they can").
5. **No HTML escaping in saved JSON.** For the byte-for-byte round trip to hold on real verify commands, `<`, `>` and `&` must be written literally, not as `\u003c`.
6. **Refused writes exit 1.** A write that would produce an invalid plan (e.g. clearing `area` while `areas` is set) exits 1 and writes nothing.
7. **`task set … ''` versus "an empty model is exit 2".** These contradict each other: `''` always clears. I required only that clearing removes the key, and nothing enforces an "empty model" error.
8. **Left open:**
   - Whether `task reset` also clears `notes`. The brief says only status and attempts; `amend.sh` does clear notes.
   - Whether saving keeps unknown keys. Only "loading doesn't reject them" is required.
9. **F1 when the checked brief isn't in the cycle.** `brief check` now reports only the loop itself, starting from its smallest name. Before, it included the path leading into the loop.
10. **Output details the brief doesn't spell out:**
    - `task validate` ends with `<n> problem(s)` written literally, as B1 does, and finding lines may be indented.
    - In text mode, `task gate` sends the verify command's stdout to stdout.
    - A task without an area shows `area: null` in `status --json`.
11. **Shell invocation is tested with stand-ins.** T6 checks how each shell is called using fake `bash`, `pwsh`, `powershell` and `cmd` scripts on a stripped `PATH`, so the real shells aren't needed.
12. **Regression coverage between tasks.** As the brief asks, only T9 runs `go test ./...`. Every gate also runs `go vet ./...`, gofmt, `go mod tidy -diff` and the three builds, plus its own packages' tests.


# Journal — B20260929-2325-vloop-metrics-defects

Append-only narrative of this plan. Rendered state lives in .loop/state/plan.md.

## Plan — B20260929-2325-vloop-metrics-defects

- **Brief:** `docs/briefs/B20260929-2325-vloop-metrics-defects.loop-brief.md`
- **Tasks:** 10

The plan is written to `.loop/state/state.json`: run `B20260929-2325-vloop-metrics-defects`, 10 tasks. The first ready task is **T1** (the `defect/v1` and `metrics/v1` schemas plus the five `metrics.*` config keys).

I ran every `verify` command and each one fails now for the right reason: a missing package, test or command. None fails on a shell error. `.loop/amend.sh check` says the plan is valid. I also re-ran the driver's own gate-shape checks against it and they came back clean. No paths are absolute, and every reference resolves.

**Tasks:** T1 schemas and config keys → T2 classification → T3 reading runs → T4 lines → T5 time, cost and tasks → T6 release detection and `defect add|list|set` → T7 derived defects and the matrix → T8 `vloop metrics` → T9 guides → T10 close.

**How the gates work**
- **T3, T4, T5:** these have no command to run yet, so their gates are the build checks plus committed tests with the names I pinned below. Nothing independent checks them until T8.
- **T6, T7, T8:** these gates build their own test repositories in temporary directories, from scratch. One copies the brief's worked example commit for commit with fixed dates, and its line counts match the brief. The other uses vloop's own run layout: `rejected` findings of each kind, gate durations, a `blocked` iteration, a brief with no task estimate, and not merged. The T8 gate compares the output line by line with the brief.
- **T10:** runs `go test ./...` and the real-data check on this repo. I summed the real telemetry myself and it matches the brief's B1 and B2 numbers ($7.40 and $8.36, 15.8 and 17.0 min, B2's missing `T2 review (iteration 4)`).

**What I had to interpret — please check before running:**
1. **No base SHA.** The brief says the planner pins a base SHA for gates to compare against, but the driver rejects gates that diff against a fixed commit. No gate diffs against anything.
2. **Several tasks edit the README.** B1's README test fails whenever a command or flag goes undocumented, so T1, T2, T6, T7 and T8 each update README.md, and T9 finishes it. `amend.sh check` flags README.md and `.vloop/config.toml` as shared "gate files". That warning is a false alarm: those names only appear as paths inside the test repositories.
3. **T10 adds `.vloop/config.toml` to this repo** with `metrics.stacks = ["go"]` and `metrics.code = ["internal/brief/templates/**"]`. Without it, B1 and B2 count zero code lines and "cost per 1,000 lines" can't be a number.
4. **Checking stack names moves to T2.** T1 adds the keys, and T2 rejects unknown stacks against the preset list, so the list exists in one place only.
5. **`defect/v1` keys stay hyphenated** (`found-by`, `fixed-by`) as in the file, and the schema also requires `schema: "defect/v1"`, following B2's convention. `defect list --json` prints that plus `summary`.
6. **When several presets share a glob**, the output credits the preset listed first in `metrics.stacks`. The gates only use cases where config order and alphabetical order agree.
7. **The `gates` time format** for the vloop layout accepts `0.5` with or without `min`, because the brief's template doesn't say.
8. **Gate-history reclassification:** the test data dates the operator's entry after the `gate_fail` commit and before the `done` commit, so comparing by timestamp or by iteration order both pass.
9. **What the gates can't check:** the cross-brief row checks only the columns the brief defines exactly, and the metrics JSON is checked only for its top-level keys and the right values somewhere inside each section. Leaf key names are T1's design.

**Names I pinned because gates use them:**
- Packages: `internal/classify`, `internal/runs`, `internal/metrics`, `internal/defect`.
- Test-name prefixes: `TestPreset`, `TestLayer`, `TestShellLoop`, `TestVloopLayout`, `TestOwnedCommits`, `TestLines`, `TestRework`, `TestTimeCost`, `TestTasks`, `TestMissingRecords`, `TestAdd`, `TestBlame`, `TestRelease` (in `internal/runs`), `TestDerived`, `TestMatrix`, `TestRemovalEfficiency`, `TestMetricsCommand`, `TestGuide` (in `internal/cli`), `TestWorkedExampleB3` (in `cmd/vloop`).
- Paths: `docs/guide/metrics.md` and `docs/guide/defects.md` (from the brief), plus `defect list --json` carrying `summary`.

The file names in each task's `files` list only suggest where code goes; no gate depends on them.


## T1 — Add the `defect/v1` and `metrics/v1` schemas and the five `metrics.*` config keys

- **Outcome:** done (review: PASS)
- **Summary:** Added the embedded defect/v1 and metrics/v1 schemas with valid/invalid fixtures, and the five metrics.* list config keys (glob-validated), with README documentation.
- **Files:** schemas/defect.v1.json, schemas/metrics.v1.json, internal/schema/testdata/defect.valid.json, internal/schema/testdata/defect.invalid.json, internal/schema/testdata/metrics.valid.json, internal/schema/testdata/metrics.invalid.json, internal/schema/schema_test.go, internal/config/config.go, internal/config/config_test.go, internal/cli/config_test.go, internal/cli/schema_test.go, cmd/vloop/e2e_test.go, cmd/vloop/b2_e2e_test.go, README.md
- **Notes for next iteration:** config.Key gained a Glob flag: with List it accepts any non-empty entry instead of lower-case names; metrics.stacks is a plain name list until T2 validates it against presets. Besides the config list/schema list expectations the brief names, internal/config's TestDefaultsEveryRow (the same config list rows) needed the five new rows. metrics/v1 leaf design: tasks{planned,done,blocked,first_pass,estimate{min,max}|null}; size{delivered,churn: lines{code,test,docs,other,deleted{code,test,docs,other}}, test_code_ratio, rework}; time{agent_ms,work_ms,review_ms,plan_ms,gates_ms|null,wall_ms|null}; rate{code_per_min,code_test_per_min}; cost{total_usd,plan_usd,work_usd,review_usd,per_1000_code_lines_usd}; tokens{input,output,cache_read,cache_creation,cache_hit_ratio}; models/effort per phase (models arrays, effort string|null); defects{in_loop,operator,escaped,total,removal_efficiency}; records{missing[{task,phase,iteration}]}; lead_time{plan_to_merge_ms}; merged is a full SHA string or null; by_task rows{id,area,kind,attempts,churn,agent_ms,cost_usd,models[]}. The metrics fixture's invalid case is /tasks/done as a string.

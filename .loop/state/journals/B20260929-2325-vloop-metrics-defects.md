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

## T2 — Build line classification: the stack presets, the layers, `vloop metrics stacks` and `vloop metrics classify`

- **Outcome:** done (review: PASS)
- **Summary:** Added internal/classify (nine presets as Go data, layered Classifier), `vloop metrics stacks` and `vloop metrics classify` (with --json), and metrics.stacks validation against the preset names; README documents both commands.
- **Files:** go.mod, go.sum, internal/classify/presets.go, internal/classify/classify.go, internal/classify/classify_test.go, internal/config/config.go, internal/cli/metrics.go, internal/cli/classify_test.go, internal/cli/root.go, README.md
- **Notes for next iteration:** config imports classify for Names() (Key.Valid on metrics.stacks); List keys with Valid now check each entry against it and the error reads 'want a comma-separated list of <names>'. Classifier.Classify returns Result{Category,Layer,Glob}; later tasks should build it via classify.New(repoPreset, stacks) from the five metrics.* keys (see newMetricsClassify in internal/cli/metrics.go for loading). Root-package TestSchemasAreEmbedded (embed_test.go) expects 5 schema files but T1 made it 7; it is outside every verify command and I did not touch it (pre-existing test file) -- needs an operator/plan fix. The test files internal/config/stacks_test.go listed in the task was not created; stacks validation is tested in internal/cli/classify_test.go.

## T3 — Read the shell-loop and vloop run layouts and the git history into one model per brief

- **Outcome:** done (review: PASS)
- **Summary:** Added internal/runs: resolves a brief to its run id, collects shell-loop and vloop run folders (sessions, iterations, verdicts) into one model, and finds the commits a brief owns (latest plan, last run commit, task commits between) with the plan read from git.
- **Files:** internal/runs/runs.go, internal/runs/commits.go, internal/runs/runs_test.go
- **Notes for next iteration:** Load(root, brief) reads folders only; Read adds Owned via OwnedCommits(root, runID) (nil if no plan commit). Session.Task comes from the iteration the session names (not for plan sessions). Iteration.Outcome is as written; Canonical() maps gate_fail->gate_failed, review_fail->rejected. Iteration.GateMS is nil unless the record has gate.duration_ms (shell-loop records never do). Shell-loop folder ownership: loop.log with ANSI stripped, 'planning from <repo-relative brief path>' exact or 'resuming <run id>' as a whole word. Latest plan = newest committer time across [loop]/[vloop] plan commits, ties go to the descendant (topo order); run/task commits must use the plan's prefix. Without a run commit, tasks are all descendant task commits of the plan on any ref. PlanDoc/PlanTask read only the fields shell-loop and state/v1 plans share (id,title,area,kind,status,attempts). Tests use only temp dirs and set GIT_CONFIG_GLOBAL=/dev/null.

## T4 — Count lines: delivered and churn per category, deletions, rework and test:code

- **Outcome:** done (review: PASS)
- **Summary:** Added internal/metrics/lines.go: DiffLines counts added/deleted non-blank lines per category from a git diff; Measure gives delivered (plan parent to last run commit), churn and per-task churn; Size.Rework and TestCode return nil on a zero denominator. Tests reproduce the worked example (15/6/3 delivered, 17 churn, 1.13, 0.40).
- **Files:** internal/metrics/lines.go, internal/metrics/lines_test.go
- **Notes for next iteration:** Parses `git diff -M -U0` (no ext-diff/textconv); binary files produce no hunks so are uncounted; a deleted file is classified by its old path, a rename by its new path. Excluded paths are dropped by Classify's category (only the always layer, i.e. top-level .loop/ and .vloop/, plus preset excluded globs). Measure uses the last task commit, then the plan, as head when there is no run commit; a root plan commit diffs against the empty tree. Counts/Lines are exported with json tags but leaf naming for metrics/v1 is left to the task that renders it (the schema wants size.*.lines{code,test,docs,other,deleted{...}}). Tests build temp repos with a fixed date and do not read this repo.

## T5 — Compute time, cost, tokens, models, task counts, first-pass and missing records

- **Outcome:** done (review: PASS)
- **Summary:** Added internal/metrics/time.go: Aggregate sums time, rate, cost, tokens, models, effort and gate time from a runs.Model; CountTasks gives planned/done/blocked/first-pass/estimate/iterations per closed; MissingRecords lists iterations lacking work/review session records.
- **Files:** internal/metrics/time.go, internal/metrics/time_test.go
- **Notes for next iteration:** Aggregate(m, delivered Counts) takes delivered lines from Measure; CountTasks(m, plan, briefText) takes the plan from runs.PlanAt at the run commit. Unknowns are nil pointers (gate for shell loop, wall without a run commit, rates/ratios on zero denominators). Session/iteration matching for missing records is per run folder (iteration numbers restart per folder). Missing records are only reported; existing sessions of that iteration are still summed. First-pass = task has iterations and all ended 'done'. Session Model is used for the models list only when modelUsage is empty. Helper is named fratio because lines.go already has an int ratio.

## T6 — Detect release and add `vloop defect add|list|set` with `--blame` attribution

- **Outcome:** done (review: PASS)
- **Summary:** Added runs.Release/DefaultBranch/StatusAt, the internal/defect package (add, list, set, blame) and `vloop defect add|list|set` with --blame attribution; README documents them.
- **Files:** internal/runs/release.go, internal/runs/release_test.go, internal/defect/defect.go, internal/defect/blame.go, internal/defect/defect_test.go, internal/cli/defect.go, internal/cli/defect_test.go, internal/cli/root.go, README.md
- **Notes for next iteration:** runs.Release(root, briefPath) returns the full SHA or '' (error only when there is no default branch); runs.DefaultBranch returns refs/remotes/origin/<x> for origin/HEAD, else main, else master. runs.StatusAt reads a committed brief's frontmatter status; task 8 can reuse all three. defect.CheckTask reads the plan at the last run commit, else the last task commit, else the plan commit. Blame reads a trailer via git show %(trailers) and falls back to the docs/briefs/*.loop-brief.md file changed in that commit that became consumed relative to its first parent. `defect set task ''` removes the task line; fixed-by and case are written quoted when empty or containing : # or quotes. Missing --found-by, and neither --brief nor --blame, are usage errors (exit 2). `defect list --matrix` and derived defects are not done here; they belong to task 8.

## T7 — Derive defects from runs, reclassify through `gate_history`, and print the origin × catcher matrix

- **Outcome:** done (review: PASS)
- **Summary:** Added metrics.Derive (derived gate/review defects with gate_history reclassification), Matrix, DefectCounts and RemovalEfficiency, and `vloop defect list --matrix [--brief]`; README documents it.
- **Files:** internal/metrics/defects.go, internal/metrics/defects_test.go, internal/cli/defect.go, internal/cli/matrix_test.go, internal/runs/commits.go, README.md
- **Notes for next iteration:** Derive(m, plan) takes the plan at the run commit (else last task commit), read with runs.PlanAt. To do that I added GateHistory []GateEntry to runs.PlanTask, a small change outside the listed files. The time of a gate failure is taken from the k-th gate_failed/gate_fail task commit of that task (m.Owned.Tasks), falling back to iteration Ended, because the shell loop's iterations carry no times. A failure of unknown time is never reclassified. Matrix is [4][4]int in defect.Origins x defect.FoundBys order; Matrix.Counts() gives DefectCounts{InLoop,Operator,Escaped} and DefectCounts.RemovalEfficiency() returns *int (nil = n/a) for task 8 to print. --matrix with no --brief covers briefs in docs/briefs that have run folders, plus ALL recorded defects; with --brief only that brief's. --matrix with --json prints the 4x4 array.

## T8 — Add `vloop metrics`: the per-brief summary, `--by task`, the cross-brief table and `--json`

- **Outcome:** done (review: PASS)
- **Summary:** Added `vloop metrics [<brief>…] [--by task] [--json]`: metrics.Build assembles a metrics/v1 Report per brief from tasks 3–7, and the CLI prints the summary, the --by task table, the cross-brief table, or the JSON.
- **Files:** internal/metrics/metrics.go, internal/runs/release.go, internal/cli/metrics.go, internal/cli/metrics_print.go, internal/cli/metrics_test.go, README.md
- **Notes for next iteration:** metrics.Build(root, brief, classifier) returns nil (not an error) when the brief has no runs; the CLI turns that into `vloop: no runs for <name>` exit 1. Status is read from the working-tree brief; merged/lead_time come from runs.Release (a missing default branch means not merged). Added runs.FrontmatterStatus, a one-line exported wrapper, outside the listed files. JSON removal_efficiency is the exact fraction (0..1) while the text shows a whole percentage from DefectCounts.RemovalEfficiency. Gates shows minutes with no unit (`gates 0.5`). Cross-brief mode ignores --by. Unknown --by value is a usage error (exit 2). Brief paths passed to Build must contain `/` or end in .md (BriefPath treats bare names as docs/briefs names). Tests reuse gitIn/write/runCLI from other cli test files.

## T9 — Write docs/guide/metrics.md and docs/guide/defects.md and bring the README up to date

- **Outcome:** done (review: PASS)
- **Summary:** Added docs/guide/metrics.md and docs/guide/defects.md, a Guides section in the README linking both, and internal/cli/guide_test.go (TestGuide...) which checks the guides against the embedded schemas and classify presets.
- **Files:** docs/guide/metrics.md, docs/guide/defects.md, README.md, internal/cli/guide_test.go
- **Notes for next iteration:** README was already documenting the commands and config keys (T6-T8), so only a 'Guides' section was added (191 lines, cap 200). The guide test walks the schema JSON for every 'properties' key at any depth (plus string enums for defect/v1) and every glob of every classify preset. The preset table in metrics.md is hand-written but the test fails if it drifts. B4's guide parts were not written.

## T10 — Close: the end-to-end worked example and the real-data check on this repository

- **Outcome:** done (review: PASS)
- **Summary:** Added cmd/vloop/b3_e2e_test.go (TestWorkedExampleB3Commands and ...PlantedFailures build the binary and replay the brief's fixture, every command and the six planted failures) and .vloop/config.toml (metrics.stacks = go, metrics.code = the brief templates' glob). Real-data check on B1 and B2 passes read-only.
- **Files:** cmd/vloop/b3_e2e_test.go, .vloop/config.toml, embed_test.go
- **Notes for next iteration:** go test ./... failed on a stale expectation in embed_test.go (TestSchemasAreEmbedded wanted 5 embedded schemas; T1 added defect/v1 and metrics/v1, so 7). Updated that one number; it is the required 'schema list gains two' change, not a weakening. The fixture builder takes b3Opts (noTrailer, gateHistory, logNames) so each planted failure is a variant of one repo; a later notes.txt commit on main is the 'neither trailer nor consumed' blame target, and the squash sha is taken before it. gate_history entry is dated 09:05, between the gate_fail (09:04) and done (09:06) commits. .vloop/config.toml uses dotted keys.

## Run ended — complete

- **Run:** `20260929-233553` · 10 iteration(s) this run
- **Plan:** 10/10 done, 0 blocked
- **Signals:** 10 iterations · 1.00 per closed · 0 gate failure(s) · 0 review rejection(s) · 0 attempt(s) burned · streak 0 · ~$12.31

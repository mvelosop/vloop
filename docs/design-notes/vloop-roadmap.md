---
name: vloop-roadmap
description: The ordered series of loop briefs that builds vloop v0.x with the vendored shell loop, what each owns, and what each depends on. Read before writing the next brief; its "Owns" column is the next brief's out-of-scope source.
kind: design-note
status: draft
created: 2026-09-29
---
# vloop — brief roadmap

The series that takes vloop from nothing to a driver that can replace the
vendored shell loop in `.loop/`. Derived from section 5 of
`docs/additional-context-files/20260929-0951-vloop-design-session.md`, as
amended by the architect act recorded in
`docs/briefs/B20260929-1222-initial-setup-for-vloop-cli.architect-brief.md`.

Every brief is planned and run by the shell loop (`.loop/run.sh`) until B6
lands. Each one names its predecessors in `depends-on:` frontmatter — the
feature B1 builds — so from B2 on, `vloop brief list` shows this order.

| # | Brief | Owns | Depends on |
| --- | --- | --- | --- |
| B1 | `docs/briefs/B20260929-1804-vloop-skeleton-config-briefs.loop-brief.md` — **consumed, merged in `8da6c95`** | Go module and command skeleton, global flags, exit codes 0/1/2, `version`, embedded plugin manifest, `.vloop/config.toml` + `config`, the loop-brief format (English and Spanish), `brief check`, `brief new`, `brief list`, `depends-on`, the root README | — |
| B2 | `docs/briefs/B20260929-2148-vloop-state-tasks-status.loop-brief.md` — **consumed, merged in `6de9cfc`** | `schemas/`: state, proposal (with `gate_dispute`), verdict, session and iteration telemetry. `task …` (including `task verify`, which records gate history, and `task gate`, which runs a gate), `status [--json]`. Per task: `model`/`effort` overrides resolved over B1's per-kind config, `area` from the repo's `areas` list, `kind` from the fixed list; `task validate` enforces both | B1 |
| B3 | `docs/briefs/B20260929-2325-vloop-metrics-defects.loop-brief.md` — **consumed** | Metrics and defects, first half of the split: the `metrics/v1` and `defect/v1` schemas; line classification (config globs plus the built-in **stack presets** chosen with `metrics.stacks`; `vloop metrics stacks`, `vloop metrics classify`); reading the shell loop's run folders and vloop's own layout; delivered vs churn, time, cost, tokens, models, missing records; `vloop metrics [<brief>] [--by task] [--json]` and the cross-brief table; release detection; `vloop defect add\|list\|set` with `--blame`; derived defects and `gate_history` reclassification; `docs/guide/metrics.md` and `docs/guide/defects.md`. B1 and B2 are its real-data check | B2 |
| B4 | not written | Close and consolidate, second half of the split: `vloop brief close` (the run record between generated markers, operator findings as defects, `status: consumed`), the saved snapshot `.vloop/state/metrics/<run-id>.json`, `vloop metrics export` (JSON Lines), `vloop metrics --workspace`; the rest of the guide — concepts, the configuration reference including the presets, and a command reference generated from the binary with a test that fails when it is stale | B3 |
| B5 | not written | `init`, `upgrade`, `doctor`; extracting the embedded plugin; `--plugin-dir` handshake; skill prose for `/vloop:plan`, `/vloop:work`, `/vloop:review` (as `[hygiene]` edits, not loop tasks) — the planner assigns `area` and `kind`, the work skill can raise `gate_dispute`; the skills write journals, notes and verdict reasons in the configured `language`. **The operator skill**, `/vloop:operate` (or split into `/vloop:run`, `/vloop:verify`, `/vloop:close` if it grows): the plugin form of this repo's `.claude/skills/vloop-operator/SKILL.md` — for the operator's interactive session, not for loop sessions — driving the CLI (`status`, `task verify`, `metrics`, `defect add`, `brief close`) instead of the shell scripts, and thinning as later slices turn its steps into commands. It starts from a `ready` brief: writing briefs (the design act) stays upstream, with the consumer | B4 |
| B6 | not written | `run` — the driver port; **before every commit, checks HEAD is still on the run's branch**, and the fence denies `git branch`, `git checkout` and `git switch` to sessions (B3's planning session renamed `main`); passing the shell loop's scenarios; `run`'s exit codes 0–7; passes model and effort per session to `claude`. Records what the shell loop does not: gate duration, wall time, configured effort; writes the metrics snapshot after every iteration; blocks a task on `gate_dispute`; classifies flakes; prints the summary on completion. **Work starts on a work branch**: on the default branch, `run` creates a branch named for the run id and switches to it before planning, so the plan commit never lands on the default branch | B5 |

After B6: v1.0 is `vloop run` building its own next patch release (design
session, section 6). **`.loop/` is kept, not deleted**: its state, journals and
per-iteration commits are the record of how vloop was built. From then on it is
evidence, not the driver.

## Decisions every brief in the series inherits

- **vloop's files live under `.vloop/`**, never `.loop/`. This repo's `.loop/`
  is the shell loop that builds vloop and stays as the evidence of it, and a
  consumer repo migrating from the shell loop can hold both. Config, state,
  journals, metrics, defects and transient files all go under `.vloop/`.
- **Config lives in the repo only**: `.vloop/config.toml`, overridden by
  `VLOOP_*` environment variables, overridden by flags. Nothing is read from or
  written to `~/.config` or any global location.
- **Language.** One `language` per repo, `en` or `es`. It selects the heading set
  a brief is written in, the templates `vloop` generates, and (from B5) the
  language the skills write prose in. Keys, frontmatter properties,
  `state.json`, JSON output, flag names and vloop's own CLI messages stay
  English.
- **Model and effort** are configured per session kind (`plan`, `work`,
  `review`) and, from B2, overridable per task.
- **The loop brief is vloop's only input.** vloop knows one kind of document,
  `B<YYYYMMDD-HHMM>-<slug>.loop-brief.md`, with the frontmatter convention of
  `.loop/loop-brief.template.md`. Design briefs, architect briefs and anything
  else upstream belong to the consumer.
- **No loop-knowledge.** The brief's `## Binding references` section is the only
  way a document binds a task; there is no knowledge-roots file and no
  preflight scan.
- **vloop ships no gates.** It runs the verify commands a plan names.
- **A run never commits to the default branch.** Planning and every iteration
  happen on a work branch named for the run id; the branch is kept after the
  squash-merge as the record of the run. `vloop run` enforces it (B6); until
  then the operator creates the branch before running `.loop/run.sh`.
- **Gates run in the OS's shell, not only POSIX.** A `shell` setting (`sh`,
  `bash`, `pwsh`, `powershell`, `cmd`; default `sh` on macOS and Linux, `pwsh`
  on Windows) says which shell gates are written for. The plan records it, the
  planner writes every `verify` in that shell's syntax, and gates run in the
  plan's shell. From B2.
- **Cross-OS by cross-compiling.** Every Go task's gate builds for `windows`
  and `linux` as well as the host.
- **Metrics are vloop's, not the shell loop's.** `.loop/` gets none of this; vloop
  reads its run folders as a source and nothing more.

## Metrics and defects

Decided with the operator on 2026-09-29. B2 defines the telemetry schemas and
records gate history; B3 and B4 build the rest, and B6's driver writes
the telemetry they read. Sample outputs, mostly from
B1's real telemetry:
`docs/additional-context-files/20260929-2143-vloop-metrics-sample-outputs.md`.

**Always calculated and saved.** The driver recomputes after every iteration and
writes `.vloop/state/metrics/<run-id>.json` in that iteration's commit, so a
halted or abandoned run still has its numbers. The file is a snapshot:
`vloop metrics` can always recompute it from the raw sources — session records,
iterations, git, defect files. Metrics are keyed by **brief**, not by run
folder: one brief can span several runs (a refused resume, a halt and resume).

**Per brief and per task**

- **Size.** Lines are classified by path globs — `code`, `test`, `docs`,
  `other`, `excluded`. **Stack presets** supply the globs: `metrics.stacks` in
  config lists the ones a repo uses (e.g. `["csharp", "react"]`), their globs
  merge, and the repo's own globs in config are layered on top and win.
  `vloop metrics stacks` lists the presets and shows each one's globs. Presets:
  **Go** (`**/*_test.go`, `**/testdata/**`; excluded `go.sum`), **TypeScript**
  and **JavaScript/Node** (`**/*.test.ts`, `**/*.spec.ts`, `**/__tests__/**`,
  the `.js` equivalents; excluded lockfiles, `dist/**`, `node_modules/**`),
  **React** (adds `.tsx`/`.jsx` tests and `**/*.stories.*`), **C#/.NET**
  (`**/*.Tests/**`, `**/*Tests.cs`; excluded `bin/**`, `obj/**`,
  `*.Designer.cs`), **Python** (`**/test_*.py`, `**/*_test.py`, `**/tests/**`;
  excluded `__pycache__/**`, `*.lock`), **Java** and **Kotlin**
  (`src/test/**`; excluded `build/**`, `target/**`), **Rust** (`tests/**`,
  `benches/**`; excluded `target/**`, `Cargo.lock`). Every preset also excludes
  `.vloop/**` and `.loop/**`. Extension alone is not enough — embedded templates
  are `.md` files that are product — which is why the repo's own globs win.
  Count added non-blank lines; report deletions separately.
  - **Delivered**: the diff from the plan's base to the close, per category.
  - **Churn**: every iteration's diff summed, rejected attempts included.
    **Rework** = churn / delivered.
- **Time.** Three clocks, kept apart: **agent** (sum of session durations),
  **gate** (sum of gate durations), **wall** (run start to complete, blocked time
  included). Rates use agent time: code lines/min, and with tests.
- **Models.** The model actually used, per session, from the session record; the
  configured effort beside it, since nothing reports effort back.
- **Also:** cost (total, per phase, per task, per 1,000 code lines) with tokens
  stored beside it so old briefs can be re-priced; cache-hit ratio; first-pass
  yield; iterations per closed task; test:code ratio; plan accuracy (brief
  estimate vs planned vs done); blocked tasks and operator interventions;
  permission denials; lead time `created` → `ready` → run → close → merged.
- **Task labels.** `area` — where the work is, from the repo's `areas` list in
  config (vloop's own: `cli`, `brief`, `config`, `plugin`, `docs`), assigned by
  the planner; optional `areas.<name> = [globs]` split a task's lines across
  areas. `kind` — what sort of change, fixed across repos: `feature`, `fix`,
  `refactor`, `test`, `docs`, `chore`.

**Defects** are recorded on two axes, origin and catcher.

| Origin | Meaning |
| --- | --- |
| `brief` | the spec was wrong or silent (`kind: spec-gap`) |
| `plan` | the planner's gate or task was wrong (`gate`) or too weak (`gate-gap`) |
| `work` | the implementation was wrong |
| `env` | flake, toolchain, machine |

Caught by: `gate`, `review`, `operator` (before merge), `user` (after release).

- Gate failures and review rejections are recorded **automatically** by the
  driver against the task and attempt. A gate failure starts as `origin: work`.
- **A gate can be the defect.** The work session may report `gate_dispute` with
  evidence; the task blocks instead of burning attempts, and the operator rules
  with `vloop task verify <id>` (replacing the gate, with a reason). Amending
  the gate reclassifies that failure, and every later one on the task, as
  `origin: plan`.
- A gate that passes on retry with no code change is classified `env` by the
  driver.
- A gate that passes while review or the operator finds the behaviour wrong is
  two defects: the `work` defect and a `plan` `gate-gap`.
- Operator and post-release findings are recorded with `vloop defect add`, one
  file each: `.vloop/defects/D<YYYYMMDD-HHMM>-<slug>.md`, frontmatter `brief`,
  `task` (optional), `origin`, `found-by`, `kind`, `severity`, `status`,
  `fixed-by` (the brief that fixes it), `case` (the failing test written first,
  per design session section 6).
- `--blame <file:line>` suggests the brief that introduced a line: git blame
  leads to the squash commit, whose `Vloop-Brief: <name>` trailer names it. The
  operator confirms.
- Headline numbers: **defect removal efficiency** = before release / (before +
  after); escaped defects per 1,000 lines; the origin × catcher matrix.

**Closing and release**

- `vloop brief close <brief>` requires the plan complete (or `--abandon
  "<reason>"`), records the operator's pre-merge findings (or `--no-findings`),
  computes the final metrics, writes the brief's `## Run record` between
  generated markers, sets `status: consumed`, prints the summary, and commits.
- **Release is the brief's merge to the default branch**, detected, not
  declared: the first default-branch commit where the brief is `consumed`.
  Anything found after it is post-release.
- `run` prints a provisional summary on completion; `close` prints the final one.

**Across repos.** Each repo owns its data; consolidation only reads.
`vloop metrics export` writes versioned JSON Lines — one record per brief, task
and defect, carrying the repo's identity (remote URL and name), numbers and
titles only, never code. `vloop metrics --workspace <file>` aggregates the
clones a workspace file lists; that file lives in a repo of its own, never in a
global location. Anything bigger (DuckDB, a dashboard) reads the export; a
dashboard is a later brief. Lines of code are compared within a language or
area, never across; the language-neutral numbers — first-pass yield, removal
efficiency, rework, cost per task — carry cross-repo comparison.

**The flow**

```
vloop brief new <slug>  → edit → vloop brief check → status: ready
vloop run <brief>       → one line per iteration, provisional summary at the end
vloop defect add …      → operator findings before merge
vloop brief close <brief>
PR → squash-merge, trailer Vloop-Brief: <brief>
vloop defect add --blame <file:line> …   → post-release
vloop metrics [<brief>] [--by task] [--json] · export · --workspace <file>
```


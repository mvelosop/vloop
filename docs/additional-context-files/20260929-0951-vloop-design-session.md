---
name: vloop-design-session
description: Decisions from the 2026-09-28/29 design session that founded vloop — name, scope, CLI shape, plugin/binary packaging, planner/architect split, self-hosting path. Read before authoring the first brief.
kind: design-note
status: draft
created: 2026-09-29
source: claude.ai session over mvelosop/exploring-claude @ cce4eb5
---
# vloop — founding design session

Edited record of the session that scoped vloop. Decisions are marked **Decided**;
everything else is a recommendation or an open question, and says so.

## 1. What vloop is

A Go CLI plus a Claude Code plugin that packages the autonomous loop from
`an-autonomous-loop-3` (vendored in `exploring-claude` as `.loop/`, v1.0.0-rc.1,
commit 191e378) into one installable unit.

- **Decided — name:** `vloop`, alias `vl`. "v" = verified: nothing counts until
  the gate and the review pass it.
- **Decided — scope excludes gates.** vloop runs whatever `verify` command a
  plan names; it never ships gate implementations. Gates are the consumer's
  commands, chosen at architect/planner time, and must run on whatever OS the
  loop runs on. (`docs-lint`, `instruction-lint` etc. stay SparIQ commands.)
- **Decided — skill names drop the `loop-` prefix:** `/vloop:plan`,
  `/vloop:work`, `/vloop:review`.
  - Risk: `/plan` and `/review` are Claude Code built-ins (exploring-claude was
    bitten by `/plan` in plan 0007). The driver always uses the full prefix, so
    it is safe in principle. **Add a fixture test** that `claude -p "/vloop:plan …"`
    actually resolves to the plugin skill.

### Rejected names (collision checks, 2026-09-28)

Taken in the same space: `xloop`, `runloop`, `aloop` (Syntax-Syllogism/aloop —
"standalone agentic loop runner for implementation and review phases"),
`autoloop` (≥5 Claude Code plugins), `pawl` (agent quality-gate ratchet),
`detent`, `escapement`, `stint`, `ouro`, `whorl` (existing Claude Code plugin),
`looprig`, `fixpoint`. Homebrew formulae: `ratchet`, `cog`, `tock`, `sprocket`,
`cadence`. Aliases: `xl` (Xen), `rl` (randomize-lines), `vlx` (vlt's npx).
Runner-up: `treadle` (fully free, but long).

To do: reserve `mvelosop/vloop` and `mvelosop/homebrew-vloop`.

## 2. Repo layout — one repo, one module, one version

```
vloop/
├── go.mod                          module github.com/mvelosop/vloop
├── embed.go                        package vloop — //go:embed all:plugin all:schemas
├── cmd/vloop/main.go
├── internal/
│   ├── cli/                        cobra command tree
│   ├── driver/                     run, locking, signals
│   ├── state/                      state.json load/validate/render
│   ├── brief/                      brief check
│   └── pluginfs/                   extract embedded plugin, version handshake
├── plugin/
│   ├── .claude-plugin/plugin.json  name: vloop, version: X.Y.Z
│   ├── skills/{plan,work,review}/
│   ├── agents/  hooks/
├── schemas/                        state / proposal / verdict JSON Schemas
├── .claude-plugin/marketplace.json name: vloop, plugins: [{name: vloop, source: "./plugin"}]
├── testdata/{scenarios,calibration}/
├── docs/design-notes/
├── .loop/                          vendored shell loop — bootstrap only, deleted at v1.0
└── .goreleaser.yaml
```

- `embed.go` must be at the root (`go:embed` cannot reach parent dirs) and use
  the `all:` prefix, or `.claude-plugin/` is silently skipped.
- The plugin is in a subfolder so installing copies only skills/agents/hooks
  into Claude's plugin cache, not the Go source.

### How the plugin and CLI are tied

| Concern | Mechanism |
| --- | --- |
| Loop sessions | The driver extracts its **embedded** plugin and starts every `claude -p` with `--plugin-dir <that dir>`. A released binary always runs its own released skills; the loop never loads a candidate plugin from the working tree. To verify: `--plugin-dir` works alongside `--setting-sources project` and `--strict-mcp-config`. |
| Interactive use | Plugin installed from the marketplace (`claude plugin marketplace add mvelosop/vloop`, `claude plugin install vloop@vloop`). |
| Version | The git tag is the only source. goreleaser stamps it into the binary; a gate fails if `plugin.json`'s version differs from the tag on release commits. Dev builds use `0.0.0-dev`. |
| File formats | `schemas/` is the contract; Go validates against it, skills cite it. |
| Calls | Driver invokes `/vloop:work T3`; skills call `vloop task show T3 --json` rather than parsing `state.json`. The CLI's `--json` output is the skills' API. |
| Mismatch warning | A session-start plugin hook calls `vloop version --json` and warns on mismatch (interactive sessions only). |
| Later idea | marketplace `command` source + `vloop plugin export`, so the marketplace install comes straight from the binary. |

Plugin facts (Claude Code docs): the name lives in `plugin.json` `name` and
must equal the marketplace entry name; no central registry; install id is
`name@marketplace`; names are unique only per marketplace; only official
Anthropic marketplace names are reserved. Dev loop: `claude --plugin-dir ./plugin`,
`claude plugin validate .`.

## 3. Command tree (proposal)

Hot paths at top level, everything else grouped by noun.

```
vloop
├── init [--force]                     scaffold .loop/, settings fence, CLAUDE.md section, brief template
├── upgrade [--to vX.Y.Z]              refresh loop-owned files; never touches state/
├── doctor [--fix]                     preflight: claude on PATH, workspace trust, git, fence, version match
├── run [brief]                        plan + iterate, or resume
│     --plan-only --max-iterations N --cost-ceiling USD --max-attempts N
│     --plan-model M --work-model M --stall-limit N --archive-transcripts
├── status [--watch] [--json]
├── brief new <slug> | check <path…> | list
├── task list | show <id> | verify <id> <cmd> | gate <id> | reset <id> | note <id> <text> | drop <id> | validate
├── runs list | show <run-id> | signals [<run-id>] | unlock [--force]
├── config get | set | list | path
├── completion bash|zsh|fish|powershell
└── version [--json]
```

(The earlier `docs …`, `lint …`, `metrics`, `hook` groups are dropped with the
decision that vloop does not ship gates.)

Global flags: `-C/--dir`, `--json`, `-q/-v`, `--no-color` (+ `NO_COLOR`),
`--yes`, `--dry-run` on anything that writes.

Mapping from today's scripts: `install.sh` → `init`/`upgrade`;
`run.sh --check` → `doctor`; `run.sh [--plan-only]` → `run`;
`render-plan.sh` → `status`; `amend.sh` → `task`; `check-brief.sh` → `brief check`.

### Open decisions

1. **Config lives in the repo only** (`.loop/config.toml` + `LOOP_*` env +
   flags; no `~/.config/vloop`). Recommended, because exploring-claude's rule 1
   is "all durable state stays in this repo".
2. **Exit codes.** Keep `run`'s 0–7 exactly (0 complete, 1 preflight,
   2 blocked, 3 stalled, 4 max iterations, 5 not converging, 6 cost ceiling,
   7 session error); other commands use 0 ok / 1 findings / 2 usage. Document
   both under `vloop help exit-codes`.
3. **Windows:** `verify` commands are shell strings; `doctor` reports a missing
   `sh` up front.
4. **Go libs:** cobra (+ completion), koanf, lipgloss for `status --watch`,
   goreleaser → GitHub Releases, Homebrew tap, Scoop.
5. **Symlink `vl`** from the formula/goreleaser; `doctor` warns if `vl` on PATH
   is a different binary.

## 4. Planner vs architect — the brief is vloop's only public contract

**Decided in direction.** `loop-plan` currently does two jobs: split the brief
into verifiable tasks, *and* survey the repo's knowledge roots
(`.claude/loop-knowledge.md`) to attach references. The second imposes a
documentation-layout convention on every consumer, and duplicates the
architect, whose loop briefs already carry the binding paths
(`B20260926-1944-remove-old-loop.loop-brief.md` cites 47).

1. The **brief** is vloop's only input about the project. Everything upstream
   (architect, docs layout, Linear, Claude Design) belongs to the consumer.
2. The brief format gains **`## Binding references`**: one entry per path with a
   `why` (what it constrains). The planner *distributes* these to tasks; it no
   longer discovers them.
3. **`vloop brief check`** validates the section: paths resolve, each has a
   `why`, cited folders have an entry point. The driver's dangling-reference
   rejection stays; preflight's knowledge-root scan goes.
4. **Gate-authoring facts stay a standing planner input**, because losing them
   is how SPA-207 migrated the dev DB. Preferred form: commands that encode the
   fact (e.g. `pnpm gate:boot-api` setting `DB_NAME=xc_gate`). Remainder: one
   optional free-form file, e.g. `.loop/gate-notes.md` — no roots table, no
   discoverability rules, no preflight check.
5. Trade-off accepted: the planner can no longer find a constraint the brief
   omitted; a miss surfaces in `brief check` or as empty `references` under
   `--plan-only`.

### Architect / explore

Upstream of the brief, so **not in vloop**. Keep them as SparIQ project skills
for now, rewritten to emit the vloop brief format. Extract a separate design
plugin (can share the `mvelosop/vloop` marketplace) only when a second consumer
(meetup scheduler, or vloop itself) shows what is generic. `/explore` and
`create-plantuml-diagram` are nearly generic; `/architect` is SparIQ-specific.
The old-loop skills (`create-plan`, `next`, `auto`, `acceptance`,
`pre-merge-verification`, `log-missed-bugs`) are not ported; the acceptance and
missed-bug ideas return later as vloop features.

## 5. Bootstrapping — build vloop with the shell loop

Recommended: install `an-autonomous-loop-3` into the new repo and let it build
vloop. Go gives real gates (`go build`, `go vet`, `go test`), and
`.loop/tests/scenarios/` (33 scenarios, stubbed `claude` via
`fixture_stub` in `lib.sh`) is a ready oracle for the driver.

Conditions:
1. Allow the Go toolchain in `.loop/settings.json` (`Bash(go:*)`, plus
   gofmt/golangci-lint if used); module downloads need network. Run
   `.loop/run.sh --check` first.
2. Plugin prose (skill renames, agent contracts, README) goes as single-pass
   `[hygiene]` edits, not through the loop.
3. The running loop uses `.claude/skills/loop-*`; the product plugin lives in
   `plugin/`. Never start loop sessions with `--plugin-dir ./plugin`.
4. Brief order (revised after the no-gates decision):
   1. Skeleton: cobra tree, `version`, `--json` plumbing, **the brief format
      (with `## Binding references`) and `brief check`**, since that's the
      interface everything else builds on.
   2. `task` + `status` over `state.json` (+ `schemas/`).
   3. `init`, `upgrade`, `doctor` (embedded plugin extraction, `--plugin-dir`).
   4. `run` — the driver port, passing the same scenarios; until then
      `vloop run` shells out to the vendored `run.sh`.

## 6. Self-hosting — how vloop keeps improving itself

Compiler-style bootstrap: **a released vloop builds the next version, never
the copy it runs from.**

| Stage | Role |
| --- | --- |
| Released (tagged, pinned) | Drives the loop |
| Candidate (working tree) | What the loop changes |
| Promotion | Candidate passes every gate, then the operator tags it |

`vloop doctor` refuses when the running binary was built from the tree it is
about to change.

Gates per kind of change: driver → `go test` + the scenario suite; review skill
→ reviewer calibration (9 planted-defect cases, `.loop/tests/reviewer-calibration/`;
9/9, or 8/9 while `06` doesn't reproduce); work/plan skills → **gap**, to be
closed with fixed benchmark briefs measured on iterations-per-closed, rejections,
cost; end-to-end → operator-run benchmark brief compared against the released
version's signals.

Improvement intake: real runs → signals / journals / missed bugs → **each miss
becomes a failing case first** → brief "make case N pass" → loop → operator tags.

Guardrails: the suite only grows (lint fails on deleted cases or changed
expected verdicts unless the brief sets `weakens-gates: true`); gates run from
the released version plus the candidate's new cases; tagging stays manual.

Path: v0.x shell loop builds vloop → v1.0 `vloop run` passes all scenarios and
builds its own next patch release (delete `.loop/`) → v1.x benchmark briefs make
skill changes loop-eligible → later, backlog from telemetry. Watch
iterations-per-closed on the fixed benchmark across releases.

## Supersedes

The plugin sketch in exploring-claude `docs/briefs/0001-loop-plugin-repo.md`
(written for the old PLAN.md loop, before `.loop/`).

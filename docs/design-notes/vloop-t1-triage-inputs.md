---
name: vloop-t1-triage-inputs
description: What the T1 triage must take up — the open defects, the findings carried out of B9–B11 and the v0.8.0 release, and the questions the operator raised — so the triage starts from a list, not from memory. Read first when the triage begins.
kind: design-note
status: draft
created: 2026-10-06
---
# T1 — triage inputs

The triage is a **design act, not a loop brief** (`docs/design-notes/vloop-v2.0-roadmap.md`
→ T1): read the defects (by origin and catcher) and the interventions (by kind ×
agreement × `automatable`) of B1–B11, and propose **skill improvements** and
**repo-specific guidelines that could automate interventions**. Its proposals
seed later briefs, B12 among them (code health, which the triage may reshape).

State when this note was written: vloop **v0.8.0** is released and the repository
is **public** (2026-10-06). The v2.0 series (B9–B11) was renumbered to 0.8.0
before going public: `docs/guide/concepts.md` → "What changed in 0.8". The
roadmap file still says "v2.0" in its name and rows — part of the triage's
housekeeping.

## The data

- Run records and operator notes: `docs/briefs/B20261003-2049-gate-model.loop-brief.md`,
  `B20261004-1434-interventions-options`, `B20261005-0933-quality-pass`, and the
  earlier briefs.
- Defects: `vloop defect list`; interventions: `vloop intervention list`,
  `vloop metrics --interventions` (and `--by phase`). Records from 2026-10-04
  on carry options, the recommendation and the operator's decision.
- `docs/design-notes/vloop-interventions-b1-b4.md` — the first analysis, by hand.

## Open defects

- `D20261004-1309` — `plan-checks-its-gates`' fixture grader cannot see the gate-folder files (judge reads a truncated trace; a file focus cannot name a folder). Needs a deterministic grader.
- `D20261005-0757` and `D20261006-1028` — `operate-proposes-options`: its scenario is unrealistic (plan and work committed on `main`, a missing work branch, a one-task plan for a larger brief), so the session rightly proposes a restart and the gate-only judges fail. Redesign the scaffold.
- `D20261005-1125-the-driver-s-git-config-guard-cannot-tel` — the `.git/config` guard cannot tell an editor's write (VS Code's `vscode-merge-base`) from a gate's; B11's first acceptance halted with exit 9.
- `D20261005-1125-an-exit-9-during-plan-acceptance-discard` — an exit 9 during acceptance discards a finished, unstamped plan ($8.23 lost in B11) and its run folder.
- `D20261006-1102` — the module path has no `/v2`. **Moot since the renumbering to 0.8.0**; close it with that reason.

## Findings to take up

1. **Interactive sessions do not get `/vloop:operate` from the CLI alone.** The
   skill lives in the plugin; `vloop run` loads it into its own sessions, but a
   user's interactive Claude Code session has it only if the plugin is installed
   (`claude plugin install vloop@vloop`) or the session starts with
   `--plugin-dir "$(vloop plugin path)"`. Meanwhile `vloop doctor` **passes** the
   plugin check ("supplied by vloop run") and the `CLAUDE.md` section `vloop init`
   writes tells Claude to follow `/vloop:operate`, which may not be loaded.
   Candidates: `doctor` warns when no interactive plugin is enabled; `init`'s
   next steps say it; the `CLAUDE.md` section says how to get the skill. Raised by
   the operator, 2026-10-06.
2. **`go install` builds report "commit unknown".** The binary could read the
   module version Go embeds (`runtime/debug.ReadBuildInfo`) so installed copies
   identify themselves; only the stamped build carries a commit today.
3. **Masking misses the dashed `-Users-<name>-` form** of a path (Claude Code's
   project-folder slug); one permission-denial record in B9's run kept the
   username (`.vloop/state/runs/B20261003-2049-gate-model/20261003-215555/sessions/001-plan.json`).
4. **The generated `CLAUDE.md` section and this repository's own rules** now point
   at different operator playbooks (`/vloop:operate` vs
   `.claude/skills/vloop-operator`); both are right, but side by side.
5. **Do `vloop run`'s sessions load the user's enabled plugins?** They run with
   `--setting-sources project` and `--strict-mcp-config`, so they should not;
   confirm with one cheap session. A copy of the vloop plugin installed *and*
   passed with `--plugin-dir` in one session loads two plugins of one name — say
   in the guide not to combine them.
6. **Operator-side lessons from B10–B11** for the operator skills: watch the
   driver by its process id (a `pgrep -f` pattern matched the watcher itself and
   hid a halt); close the editor (VS Code) on the repository during a run, or fix
   the guard (above); a brief's must-change list keeps missing tests that pin
   counts and lists (schema list, embedded-schema count, init's output).
7. **Gate-review misses** to learn from: an unpassable gate needing a session to
   write under `.vloop/` (B10 T9/T10), a grep that could not match a message-less
   pass (B11 T12). Both are checks the gate review's "three ways" could make
   explicit.
8. **Releases are manual ceremony** (five steps in `docs/guide/concepts.md` →
   "The stamped release build"): a `vloop release` command is the recurring
   "what would automate it".

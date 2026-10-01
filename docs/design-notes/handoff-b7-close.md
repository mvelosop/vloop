---
name: handoff-b7-close
description: Where closing B7 stands and exactly what is left — the eval suites, the evals documentation, the url-shortener acceptance run, the close and the cut-over — written so a fresh session can resume. Delete when B7 is merged.
kind: design-note
status: in-progress
created: 2026-10-01
---
# Handoff — closing B7 (the skills)

Written 2026-10-01 when the operator's session neared its context limit. Read
with `.claude/skills/vloop-operator/SKILL.md`; record every intervention as it
happens (`.vloop/interventions/`, with Context, Suggested and Decided).

## Where things stand

- **Branch** `B20261001-1025-vloop-skills`. The shell loop's run completed: 10/10
  tasks, first pass, $10.62. Verified: 17 packages, vet, gofmt, tidy, cross-builds,
  `claude plugin validate .`; `plugin path` extracts hooks and four skills, no
  evals; `vloop intervention list` reads every record; the export carries
  interventions and `repo.stacks`.
- **Interventions**: every record has Context / Suggested / Decided (commit
  e6fc579). Horizon Now has the suggestion-and-context entry.
- **The eval suites were broken and are fixed by hand on this branch**
  (uncommitted at writing; commit them first): `case.yaml` per case declaring
  `context.scaffold_script: scaffold.sh`; `setup:` removed from `prompt.md`;
  `allowed_tools` as YAML arrays; every `type: command` grader converted (regex on
  the verdict/proposal file, regex on the trace for `<file>: ok` and `plan ok`,
  `tool_used` for no commit / no edits to the plan, `llm` reading the plan for the
  structural JSON checks); the calibration's `uv run pytest` gates and its
  `pyproject.toml` (dev dependency pytest, `uv_build`) restored. The structural
  check in `internal/cli/skills_test.go` now knows the tool's real keys, grader
  types and their required keys, with five rejecting fixtures. Checked: the tool
  loads all 13 cases at a $0 ceiling; all 13 scaffolds pass by hand with an empty
  `HOME`. Defects recorded: D20261001-1338 (brief spec gap), D20261001-2016 ×2
  (work); mark them fixed by B7 once the evals run.

## What is left, in order

1. **Commit** the eval fixes and the records on the branch (check the staged list).
2. **Probe** one case (about $0.5):
   `claude plugin eval plugin --case 01-hollow-test --runs 1 --ablation none --max-cost-usd 1 --scaffold --trust-plugin --no-publish --allow-tools <tools> --output-dir <scratch>`,
   with a built `vloop` first on `PATH` (the skills call it). `<tools>`: every
   non-read tool any case's `allowed_tools` names — `Write Edit Bash(cat:*)
   Bash(chmod:*) Bash(git diff:*) Bash(git log:*) Bash(git rev-parse:*)
   Bash(git status:*) Bash(grep:*) Bash(sh:*) Bash(test:*) Bash(vloop config get:*)
   Bash(vloop schema validate:*) Bash(vloop task show:*) Bash(vloop task:*)`.
   Results and reports go to a scratch directory, never the repo; keep
   `--no-publish` unless the operator wants the report shared.
3. **The full suite** at `--max-cost-usd 40` (the loop's `run.cost-ceiling`, the
   operator's starting limit), default runs (3 per case) with the baseline arm.
   Bar: the calibration's — every review case caught, `06` may miss. Record
   scores and the real cost in B7's operator notes; adjust the limit from it.
4. **Evals documentation** — the operator asked for it, to strengthen their grasp
   of the concept: a `docs/guide/evals.md` explaining what an eval is (a
   behavioural test of a skill against a real model, scored by graders), how it
   differs from a gate (gates run every iteration and never call `claude`), the
   case format (`prompt.md`, `case.yaml`, scaffold, the six grader types, the
   with/without-plugin arms), how to run and read them (the command above, the
   grants, cost, the report), how vloop's suites map to the skills, and their
   relation to the shell loop's reviewer calibration. Link it from the README and
   `plugin/evals/README.md` (update that README for case.yaml too).
5. **Acceptance run** (the first real `vloop run`, about $15–30): the
   url-shortener sample at `9a197a7` in a scratch directory, `loop/` and
   `.claude/skills/loop-*` removed, `vloop init`, its brief converted unchanged
   to `docs/briefs/B20260818-2100-url-shortener.loop-brief.md` and checked, then
   `vloop run` with the B7 binary and real `claude`. Passes when it ends complete,
   every record validates, every task has a `kind`, and the app's unit, e2e and
   boot checks pass by hand. Compare `vloop metrics` with the shell loop's run
   (10/10 tasks, 11 iterations, 1 rejection, about $19.86). Ask the operator before
   spending on steps 2, 3 and 5 if a new session resumes here.
6. **Close B7**: `vloop brief close … --finding …` for anything the evals or
   the acceptance run found (or `--no-findings`), mark the three eval defects
   fixed by B7, operator notes (the eval fixes, eval scores and cost, the
   acceptance comparison), roadmap consumed, PR, squash-merge with the printed
   trailer on the operator's go-ahead, keep the branch, sync main.
7. **Cut-over**: `vloop init` on this repository (it keeps the config), then B8
   — the v1.0 review — is designed with the architect skill and run with
   `vloop run`. The README is at 199 of 200 lines; B8 must handle the cap.
8. Delete this note.

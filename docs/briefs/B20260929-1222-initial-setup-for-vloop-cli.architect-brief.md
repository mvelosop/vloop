---
name: B20260929-1222-initial-setup-for-vloop-cli.architect-brief
description: Implement the initial dev version for the vLoop CLI tool and it's associated Claude Code plugin
kind: brief
status: draft
created: 2026-09-29
seeds: A new loop-brief to plan and execute with the loop at `.loop/`
---
# What is the first loop brief that starts vLoop, and what does it pin?

> Copy this to `docs/briefs/B<YYYYMMDD-HHMM>-<slug>.architect-brief.md` and
> replace everything.
>
> This is **not** a loop brief. It is the input to the design act that WRITES
> one: an interactive session — your own `/architect`-style skill, or a plain
> `claude` session — that surveys the repo, settles the forks with the operator,
> amends whatever documents the decisions belong in, and produces a loop brief
> that passes `.loop/check-brief.sh`. The checker skips this
> file on purpose: it has no `**Status:** ready to plan` line, and it never
> should.
>
> Write what you WANT and what you already KNOW. Do not write the contract —
> that is the act's output, and a contract written here is one nobody surveyed.

## Goal

Produce an initial version of the "autonomous loop 3" (x-loop), currently installed at `.loop/`,
implemented as a Go binary CLI tool that can be used for projects on Mac, Windows and Linux.

## Context

This is the next step in the evolution of the x-loop, because we need to apply
this approach to a .NET FRamework project, that must run in Windows, and where
all the project documentation is/must be in Spanish.

The [initial design conversation](../additional-context-files/20260929-0951-vloop-design-session.md)
settles the main goals, and the new key requirement is that vLoop CLI is
capable of consuming and producing documents in Spanish so they can authored and read
by Spanish-speaking developers with mainly technical-English knowledge.

**Exception**: All the process variables like state.json and front matter properties will be kept in
English.

## Expected results

1. **The survey.** It must at least cover:
   - `~/source/personal/an-autonomous-loop-3/CLAUDE.md` — The current x-loop code
2. **One loop brief**, `docs/briefs/B<YYYYMMDD-HHMM>-<slug>.loop-brief.md`, from
   `.loop/loop-brief.template.md`. It passes `.loop/check-brief.sh` with no problems,
   names an expected task count, and cites — repo-relative, each with the reason
   it binds — the documents a task is bound by.
3. The expected result from the loop-brief is the initial version of a CLI and plugin based vLoop.
4. Some additional expected functionality of vLoop:
   - Must support configurable model and effort for any task
   - Must be able consume and produce Spanish or English process documents as configured for the project
   - The loop-knowledge from the x-loop must not be implemented in vLoop

## Constraints

- **Prefer a check to a rule.** What can be a script with a fixture goes into the
  loop brief as a gate; prose is for what cannot be mechanized.
- **Survey the code, not the list.** Where this brief names the places a change
  touches, the act greps for them rather than trusting the list.
- Repo-relative paths everywhere. No absolute paths in any file or commit message.

## Out of scope

- **Executing the loop brief.** This act produces it; a separate `.loop/run.sh`
  run executes it.
- Implementing the CLI or the related skills

---

## Consumed

> Filled by the act, not by the author. This is the one record of what the act
> decided and why, and the next act of the same shape reads it first.

### The act — 2026-09-29

**Produced:** `docs/briefs/B20260929-1804-vloop-skeleton-config-briefs.loop-brief.md`
(passes `.loop/check-brief.sh`: 0 problems, 0 warnings; 8 to 10 tasks) and
`docs/design-notes/vloop-roadmap.md` (the B1–B4 series and the decisions each
inherits).

**Survey findings that shaped it.**

- The design session splits the build into four ordered slices; "an initial
  version of the CLI and plugin" is all four. One brief would be 25+ tasks, so
  this act produced slice B1 and a roadmap for the rest.
- The design puts vloop's files in `.loop/`, which is where this repo's
  bootstrap shell loop lives — vloop's tests and `init` would collide with the
  loop building it.
- `.loop/run.sh` already takes a model per phase (`LOOP_PLAN_MODEL`,
  `LOOP_WORK_MODEL`, with review sharing the work model) and has no effort
  setting; `claude` accepts `--model` and `--effort`.
- The design session already replaces loop-knowledge with a brief section,
  `## Binding references`, checked by `brief check` — this satisfies "loop-knowledge
  must not be implemented" without a new decision.
- Spanish does not appear in the design session; it is new in this brief.
- Go is not installed on the operator's machine, and the loop fence has no Go
  permissions. Both block the run, not the brief.

**Decided with the operator.**

| Fork | Decision |
| --- | --- |
| Scope of the first loop brief | Slice 1 plus config: skeleton, `version`, config (language, model and effort), bilingual brief format, `brief check`/`new`. Later slices are its out of scope. |
| Organizing several briefs | Both: `depends-on` frontmatter with `brief check` and `brief list` in B1, and a roadmap note for vloop's own series. |
| What Spanish covers | Brief headings, templates, and (from B3) the prose skills write. Keys, frontmatter, JSON, flags and CLI messages stay English. One `language` per repo. |
| Model and effort granularity | Per session kind in config (B1), overridable per task in the state schema (B2). |
| Where vloop's files live | `.vloop/`, never `.loop/`. |
| Proving Windows and Linux | Cross-compile in every gate; no CI in this run. |
| Libraries | cobra, one TOML library, stdlib `testing` with golden files. |
| Brief status marker | Frontmatter `status: draft\|ready\|consumed`; the body `**Status:** ready to plan` line is dropped (it would need a Spanish twin). |
| Brief naming | The existing convention: `B<YYYYMMDD-HHMM>-<slug>.loop-brief.md`, frontmatter `name` = filename minus `.md`, run id = name minus `.loop-brief`, `depends-on` lists names. |
| Other brief types | vloop knows only `.loop-brief.md`. Design and architect briefs are the consumer's; vloop neither lists nor checks them. |
| Spanish heading set | Accepted as proposed (`Qué es`, `Por qué esta forma, y qué se descartó`, `Referencias vinculantes`, `Contrato de comportamiento`, `Ejemplo trabajado`, `Fuera de alcance`, `Restricciones`, `Forma`). |
| The shell loop after v1.0 | `.loop/` is kept as the record of how vloop was built, not deleted. |
| Documentation | One task for a root `README.md` (under 200 lines) for a first-time reader, gated against the command tree; no separate concepts doc. |

**Decided in the act, open to operator override.**

- Exit codes 0 ok / 1 problems / 2 usage for every B1 command; `run`'s 0–7 are
  deferred to B4.
- The Spanish task-count and exit-code phrasings `brief check` accepts.
- `brief check` on a file not named `*.loop-brief.md` is a usage error (exit 2).
- Config keys and defaults (`model.plan=opus`, `model.work=sonnet`,
  `model.review=sonnet`, effort unset), env names `VLOOP_<KEY>`, and precedence
  env > file > default.
- Repo-root discovery: nearest `.vloop/`, else nearest `.git`, else the start
  directory.
- The plugin ships a manifest and marketplace entry only; skill content waits for B3.

**Not changed, noted.**

- `.claude/loop-knowledge.md` still held the template's example roots, which do
  not exist here; it now declares `docs/design-notes/` and
  `docs/additional-context-files/` so the shell loop's preflight passes clean.
- The design session's command tree, layout and `.loop/` naming predate this act;
  the loop brief says it wins where they disagree, and the design note was left
  as the record of the session.
- The design session's "fixture test that `/vloop:plan` resolves to the plugin
  skill" needs a live `claude`, which the loop fence denies; it belongs to B3.

> **For the next act:** read the roadmap note first — its "Owns" column is the
> next brief's out-of-scope list. Confirm Go and the fence permissions are in
> place before writing gates. B2 must carry the per-task `model`/`effort` fields
> the B1 brief deferred to it.

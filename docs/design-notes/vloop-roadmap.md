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

Every brief is planned and run by the shell loop (`.loop/run.sh`) until B4
lands. Each one names its predecessors in `depends-on:` frontmatter — the
feature B1 builds — so from B2 on, `vloop brief list` shows this order.

| # | Brief | Owns | Depends on |
| --- | --- | --- | --- |
| B1 | `docs/briefs/B20260929-1804-vloop-skeleton-config-briefs.loop-brief.md` | Go module and command skeleton, global flags, exit codes 0/1/2, `version`, embedded plugin manifest, `.vloop/config.toml` + `config`, the loop-brief format (English and Spanish), `brief check`, `brief new`, `brief list`, `depends-on`, the root README | — |
| B2 | not written | `schemas/` (state, proposal, verdict), `task …`, `status [--json]`. **Per-task `model`/`effort` overrides live in the state schema here**, resolved over B1's per-kind config | B1 |
| B3 | not written | `init`, `upgrade`, `doctor`; extracting the embedded plugin; `--plugin-dir` handshake; skill prose for `/vloop:plan`, `/vloop:work`, `/vloop:review` (as `[hygiene]` edits, not loop tasks); the skills write journals, notes and verdict reasons in the configured `language` | B2 |
| B4 | not written | `run` — the driver port, passing the shell loop's scenarios; `run`'s exit codes 0–7; passes model and effort per session to `claude` | B3 |

After B4: v1.0 is `vloop run` building its own next patch release (design
session, section 6). **`.loop/` is kept, not deleted**: its state, journals and
per-iteration commits are the record of how vloop was built. From then on it is
evidence, not the driver.

## Decisions every brief in the series inherits

- **vloop's files live under `.vloop/`**, never `.loop/`. This repo's `.loop/`
  is the shell loop that builds vloop and stays as the evidence of it, and a
  consumer repo migrating from the shell loop can hold both. Config, state, journals and transient files all go
  under `.vloop/`.
- **Config lives in the repo only**: `.vloop/config.toml`, overridden by
  `VLOOP_*` environment variables, overridden by flags. Nothing is read from or
  written to `~/.config` or any global location.
- **Language.** One `language` per repo, `en` or `es`. It selects the heading set
  a brief is written in, the templates `vloop` generates, and (from B3) the
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
- **Cross-OS by cross-compiling.** Every Go task's gate builds for `windows`
  and `linux` as well as the host.

---
name: platform-context
description: Binds the platform every context stands on — the repo root, repo-local configuration and its resolution, the language setting, the embedded JSON Schemas and their versioning, and the global CLI conventions (exit codes, --json, paths) — and the invariants C-1..C-4
---
# Platform

What briefing, execution and measurement share and none of them owns.

## The repo root

The nearest ancestor holding `.vloop/`, else `.git`, else the start directory
(`-C` changes the start). Every path vloop prints is relative to it, with `/`
on every OS (C-2).

## Configuration

`.vloop/config.toml`, in the repo only (C-1). Each key resolves as the
`VLOOP_<KEY>` environment variable, else the file, else the default;
`vloop config list` shows each value with its source. The keys, defaults and
valid values: [`docs/guide/configuration.md`](../../guide/configuration.md).

What the keys decide, by context:

| Keys | Decide |
| --- | --- |
| `language` | briefing: heading set and templates; execution: the language sessions write prose in *(planned, B5)* (C-4) |
| `model.*`, `effort.*` | execution: each session kind's model and effort, under per-task overrides (P-6) |
| `shell` | execution: the shell a new plan's gates are written for (P-5) |
| `areas` | execution: the areas a task may belong to |
| `metrics.*` | measurement: line classification (M-3) |

## Schemas

Every JSON document vloop defines names its schema and version —
`"schema": "state/v1"` — so an old file stays recognisable after a format moves;
unknown keys are allowed so a newer writer does not break an older reader (C-3).
The schemas live in `schemas/`, are embedded in the binary, and are listed,
printed and applied by `vloop schema list|show|validate`:

| Schema | Context |
| --- | --- |
| `state/v1`, `proposal/v1`, `verdict/v1`, `session/v1`, `iteration/v1` | execution |
| `metrics/v1`, `defect/v1`, `export/v1` | measurement |

## CLI conventions

Every command takes `-C/--dir`, `--json`, `--no-color` (or `NO_COLOR`), `-q`,
`-v`. With `--json`, stdout carries exactly one JSON document (JSON Lines for
`metrics export`). Exit codes: `0` ok, `1` problems or failure, `2` usage only.
A refusal is one stderr line starting `vloop: `, after any checks `doctor` or a preflight prints. The driver's own exit codes are
R-3's. The full command reference is generated from the binary:
[`docs/guide/commands.md`](../../guide/commands.md).

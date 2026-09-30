---
name: briefing-context
description: Binds the briefing context — the Brief aggregate, its languages and heading sets, dependencies between briefs and binding references — the commands that own it, and the invariants B-1..B-6 it enforces
---
# Briefing

The context where a project's intent becomes something the loop can execute.
It owns one aggregate, the **[Brief](brief.md)**, and everything a brief carries:
its frontmatter, its sections in the repo's language, its dependencies on other
briefs and the documents it binds tasks to.

Upstream of it is the **design act** — the operator (or a consumer's own
architect skill) surveying, deciding and writing. That is not vloop's: vloop
starts at a file named `*.loop-brief.md` (B-1).

## What it owns

| Concept | Where |
| --- | --- |
| the brief's format, lifecycle and rules | [brief.md](brief.md) |
| heading sets, `en` and `es` | `internal/brief/headings.go` |
| templates `brief new` writes | `internal/brief/templates/` |
| dependency order and readiness | `brief list`; B-6 |
| binding references | B-5 |

## Invariants it enforces

B-1 to B-6 in [`../domain-model.md`](../domain-model.md#briefs--b). `brief check`
is where they are checked; the shell loop's `.loop/check-brief.sh` checks the
subset the shell loop plans from.

## Boundaries

- A brief refers to other briefs (by name) and to documents (by path). It never
  refers to a plan, a run or a metric.
- Whether a brief has run is known in two ways only: its `status`, and whether
  its journal exists (B-2).
- The languages change headings, templates and — from B5 — the prose sessions
  write; they never change keys, frontmatter, JSON or vloop's messages (C-4).

# Knowledge roots

**Read by the planning session.** This file answers a question the loop cannot
answer for you: *where does the knowledge live that a task might be bound by,
and how do I find out what is in it without reading everything?*

The loop fixes this path and nothing else about this file. It does not know
what an ADR is, or a use case, or a tier — those are conventions, and
conventions belong to the repo that has them. Delete this file and the loop
still runs; the planner simply cites nothing.

## The roots

Replace these with yours. Declare a path as a backticked directory ending in
`/`, and say how its contents can be surveyed.

| Surface | Path | How to find what is in it |
| --- | --- | --- |
| Example — architecture decisions | `docs/decisions/` | Index at `README.md` |
| Example — domain model | `docs/domain/` | Every file carries a `description:` frontmatter line |

## What "discoverable" means

Either mechanism is enough, and both is better:

- **An index** — one line per document, saying what it *binds*, not what it
  covers. Cheapest: the planner reads one file per root rather than N. Name it
  `index.md`, `README.md`, or `README-<subject>.md`; the loop accepts all three
  and prescribes none. `README.md` is what GitHub renders when you browse a
  folder; `README-<subject>.md` stays unique in a flat search, which is what an
  Obsidian quick-switcher or backlink pane gives you.
- **Per-file frontmatter** with a `description:` line. Better where files are
  added often and an index would go stale. Descriptions count at any depth, so
  a root holding one folder per slice is not blind just because its top level
  is empty.

A root with neither leaves the planner guessing from filenames, which is why
preflight warns about it before a run spends anything.

## Folders as references

A folder is the right reference when what binds is the whole bundle rather than
one file in it — a design handoff with its tokens and screens, a spec with its
diagrams. It needs an entry point for the same reason a root does: a session
handed a directory with no way in opens files until it thinks it has
understood.

That also settles content you cannot annotate. An export from an external tool
cannot carry frontmatter without being edited, and editing it forks it from its
source. You do not need to: write the index **beside** the bundle rather than
inside its files. Nothing external is touched, and an index makes per-file
descriptions unnecessary anyway.

## How to cite

Prefer the document that is *binding* over the one that is merely related.
Every reference costs attention in two sessions on every iteration that touches
the task, so each carries a reason — what it constrains, not what it is about.

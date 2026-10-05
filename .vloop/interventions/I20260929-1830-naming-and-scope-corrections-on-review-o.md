---
id: I20260929-1830-naming-and-scope-corrections-on-review-o
brief: B20260929-1804-vloop-skeleton-config-briefs.loop-brief
phase: design
kind: direction
automatable: no
by: operator
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-09-29
recorded: 2026-09-30T16:54:00Z
backfilled: true
---
naming and scope corrections on review of the B1 draft

**Trigger.** operator review: keep .loop-brief.md, vloop knows only loop briefs, keep .loop/ as evidence, add a README task

**Done.** batched and applied on the operator's word

**Context.** Reviewing the B1 draft, the operator kept the `.loop-brief.md` suffix (with `name:` frontmatter), decided vloop knows only loop briefs, kept `.loop/` as the record of how vloop was built, and asked for a README task; then 'README.md alone is enough, apply the batched changes'.

**Suggested.** `<id>.brief.md` naming, `.design-brief.md` for upstream briefs, README plus docs/concepts.md.

**Decided.** the operator overrode the naming, rejected vloop knowing other brief types, and chose the README alone.

**What would automate it.** none: these changed what the product is

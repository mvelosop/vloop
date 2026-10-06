---
id: I20261006-1237-renumbered-to-0-8-0-before-going-public
brief: ""
phase: next
kind: decision
automatable: no
by: both
schema: intervention/v2
options: 3
recommended: 1
decided: 1
agreement: recommended
occurred: 2026-10-06
recorded: 2026-10-06T11:37:19Z
---
renumbered to 0.8.0 before going public: v1.0.0 retracted, this repository's install stamp lowered by hand

**Trigger.** The operator found installing from the private repository complex and wanted to go public, but did not see vloop at a real v2; v2 also needs a /v2 module path to go install

**Done.** Pre-publication review of the whole history: no secrets, no rewrite needed; the operator confirmed SparIQ is their own, chose MIT, kept all branches, and will revoke the claude.ai share link found in history. Release branch: version 0.8.0, go.mod retracts v1.0.0, LICENSE (MIT), README install public (go install @latest, the marketplace), What changed in 0.8; .vloop/install.json and CLAUDE.md lowered to 0.8.0 by hand, because vloop upgrade refuses a lower version and the real-data test clones HEAD.

**Context.** Defect D20261006-1102-the-module-path-is-github-com-mvelosop-v explains why v2 tags cannot be installed.

**Options.**
1. continue the 0.x line as v0.8.0, retract v1.0.0, keep every tag
2. clean slate at v0.2.0, deleting the v0.7.0, v1.0.0 and v2 tags
3. stay at v2 and change the module path to /v2

**Recommended.** 0.8 sorts above 0.7 and, with v1.0.0 retracted, is latest; no tag history is rewritten

**Decided.** the operator chose option 1, after a pre-publication review

**What would automate it.** none: a direction the operator set

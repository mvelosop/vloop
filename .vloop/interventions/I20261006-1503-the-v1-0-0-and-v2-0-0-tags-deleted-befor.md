---
id: I20261006-1503-the-v1-0-0-and-v2-0-0-tags-deleted-befor
brief: ""
phase: next
kind: ceremony
automatable: partly
by: both
schema: intervention/v2
options: 3
recommended: 1
decided: 1
agreement: recommended
occurred: 2026-10-06
recorded: 2026-10-06T14:03:47Z
---
the v1.0.0 and v2.0.0* tags deleted before going public

**Trigger.** The operator saw v2.0.0 at the top of GitHub's Tags page after the 0.8.0 renumbering

**Done.** Deleted v1.0.0, v2.0.0-beta.1, v2.0.0-beta.2 and v2.0.0 on GitHub and locally (they pointed at fc0b67a, 42ee0a6, 27871b2, d219473; the commits stay). Explained that go.mod cannot retract v2 versions of a v0/v1 module path. The concepts note and the go.mod comment record it. The operator confirmed the claude.ai share link now shows unavailable.

**Context.** The tags were never public; Go's proxy never fetched them.

**Options.**
1. delete the v2 tags and v1.0.0 before going public, keeping v0.7.0 and v0.8.0
2. delete only the v2 tags, keeping v1.0.0 retracted
3. keep every tag and rely on the retraction and a note

**Recommended.** never public and depended on by nothing, so deleting them now leaves a clean 0.x line at no cost

**Decided.** the operator chose option 1

**What would automate it.** a release command that refuses a tag outside the module path's major version

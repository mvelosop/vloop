# Journal — B20261025-2302-b21-records-and-verify

Append-only narrative of this plan. Rendered state lives in .vloop/state/plan.md.

## Plan — B20261025-2302-b21-records-and-verify

- **Brief:** `docs/briefs/B20261025-2302-b21-records-and-verify.loop-brief.md`
- **Tasks:** 5

## T1 — Make a nested git repository its own vloop root (C-2) and have verify make its scratches repositories

- **Outcome:** done (review: PASS)
- **Summary:** config.Root now checks .vloop and .git at each level before moving up, so a nested git repository is its own root; C-2, the platform prose and the verify skill's scratch section say so.
- **Files:** internal/config/root.go, internal/config/root_test.go, docs/domain/domain-model.md, docs/domain/platform/platform-context.md, plugin/skills/verify/SKILL.md
- **Notes for next iteration:** Contract item C-2: TestRootNearestVloopWinsOverGit was replaced by TestRootNearerGitWinsOverOuterVloop (nearer .git beats outer .vloop/). Added tests for a .git file, start dir with .vloop under outer .git, both markers in one dir, and neither marker under a vloop repo. No Windows-specific code in Root, so no Windows-shaped cases.

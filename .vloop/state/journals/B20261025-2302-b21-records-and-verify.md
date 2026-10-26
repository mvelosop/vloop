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

## T2 — Treat a brief's bare name, .loop-brief name and path as one brief, and write the .loop-brief name

- **Outcome:** done (review: PASS)
- **Summary:** Added defect.LoopBriefName and defect.SameBrief; defect add, intervention add, intervention set brief and metrics write the .loop-brief name, and defect and intervention list match a record by either the bare or the .loop-brief name.
- **Files:** internal/defect/defect.go, internal/defect/defect_test.go, internal/intervention/intervention.go, internal/intervention/intervention_test.go, internal/metrics/metrics.go, internal/cli/defect.go
- **Notes for next iteration:** BriefName is unchanged (blame, close and export still use it). TestListOrderFilterAndRoundTrip in internal/intervention pinned that add writes the bare name B1; changed to B1.loop-brief per the contract item that records are always written with the .loop-brief name. internal/cli TestTaskGate* fail inside a driver session with 'plan was changed during this session' - environmental, not touched by this change.

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

## T3 — Mask the system temp directory as $TMPDIR in everything the driver masks (#53)

- **Outcome:** done (review: PASS)
- **Summary:** Mask now replaces the system temp directory (as given and symlink-resolved) with $TMPDIR after the home and before the user name; /tmp and /var/tmp (and their /private forms) are kept.
- **Files:** internal/driver/session.go, internal/driver/session_test.go
- **Notes for next iteration:** Runner.TempDir overrides os.TempDir for tests. keptTemp includes /private/tmp and /private/var/tmp so macOS-resolved /tmp is not masked. Temp masking reuses homePattern (boundaries, JSON-escaped, Windows forms).

## T4 — Check that a GitHub login reaches Copilot sessions in doctor and the run preflight, and reword the experimental warning

- **Outcome:** done (review: PASS)
- **Summary:** doctor gains a `copilot login` row (shown only with agent.review = copilot) and `vloop run` refuses before planning when no GitHub login reaches Copilot sessions; the experimental warning now reads "not fenced yet" and the guide says how a session logs in.
- **Files:** internal/cli/doctor.go, internal/cli/run.go, internal/cli/doctor_test.go, internal/cli/doctor_copilot_test.go, docs/guide/configuration.md
- **Notes for next iteration:** Login = COPILOT_GITHUB_TOKEN/GH_TOKEN/GITHUB_TOKEN non-empty, else `gh auth token` exit 0 with non-blank stdout; output is only tested for presence, never stored or quoted. Existing tests changed only where they pinned the old warning text (R4's warning is the contract item); newDoctorEnv now unsets the three token variables and copilot() installs a stub gh printing a token so non-login tests still pass. internal/cli task_gate tests fail inside a work session only because VLOOP_ACTIVE_TASK is set; they pass with VLOOP_* unset.

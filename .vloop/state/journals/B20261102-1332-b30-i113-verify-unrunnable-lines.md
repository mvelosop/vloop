# Journal — B20261102-1332-b30-i113-verify-unrunnable-lines

Append-only narrative of this plan. Rendered state lives in .vloop/state/plan.md.

## Plan — B20261102-1332-b30-i113-verify-unrunnable-lines

- **Brief:** `docs/briefs/B20261102-1332-b30-i113-verify-unrunnable-lines.loop-brief.md`
- **Tasks:** 5

## T1 — Add the verify/v2 report schema, embedded and listed

- **Outcome:** check_failed (review: skipped)
- **Summary:** Added schemas/verify.v2.json (verify/v1 plus required boolean complete and not_run array of {line, reason} non-empty strings), embedded and listed, with committed tests.
- **Files:** schemas/verify.v2.json, internal/schema/schema_test.go, internal/schema/testdata/verify2.valid.json, internal/schema/testdata/verify2.invalid.json, internal/schema/testdata/verify2-no-complete.invalid.json, internal/schema/testdata/verify2-not-run-no-reason.invalid.json, internal/schema/testdata/verify2-v1-report.invalid.json, internal/cli/schema_test.go, cmd/vloop/b2_e2e_test.go
- **Notes for next iteration:** Contract item V1 (new schema listed): the three pinned schema lists gained verify/v2 only; the one other edit is the length check in internal/cli/schema_test.go TestSchemaList, 25 to 26, a consequence of the added name. Other internal/cli task_gate tests fail inside this session only because the plan is held by the driver; not related. verify.v1.json is untouched.

## T1 — Add the verify/v2 report schema, embedded and listed

- **Outcome:** done (review: PASS)
- **Summary:** verify/v2 schema exists, embedded and listed, with committed tests; fixed the root embed count test that failed the go check last attempt.
- **Files:** schemas/verify.v2.json, embed_test.go, internal/schema/schema_test.go, internal/cli/schema_test.go, cmd/vloop/b2_e2e_test.go
- **Notes for next iteration:** Contract item V1 (new schema listed): the pinned schema lists gained verify/v2 only; the count checks (embed_test.go TestSchemasAreEmbedded 25 to 26, internal/cli TestSchemaList length) changed only as a consequence. The previous attempt missed embed_test.go, which the go check caught.

## T2 — Read verify/v2 in the driver's verify step and print not_run and complete at close

- **Outcome:** rejected (review: FAIL)
- **Summary:** The driver's verify step reads verify/v2 only, keeps it even after a failed session, prints not_run counts and warns when incomplete; brief close prints the lines not run and the unfinished session.
- **Files:** internal/driver/iterate.go, internal/cli/close.go, internal/driver/verify_v2_test.go, internal/cli/close_test.go, cmd/vloop/run_verify_test.go, cmd/vloop/run_iterate_test.go, cmd/vloop/b16_e2e_test.go
- **Notes for next iteration:** Contract item (stub verify reports become verify/v2): the stub reports in cmd/vloop/run_verify_test.go, run_iterate_test.go and b16_e2e_test.go moved from verify/v1 to verify/v2 (complete true, not_run empty), nothing else weakened; internal/cli/close_test.go keeps its verify/v1 cases and gains TestCloseVerifyBlockV2. internal/driver/verify_v2_test.go runs verifySession against a stub claude script (non-zero exit, and silent exit 0) in a temp git repo. A session that failed to start still returns before any read. internal/cli task_gate tests fail only inside a driver-held session.

## T2 — Read verify/v2 in the driver's verify step and print not_run and complete at close

- **Outcome:** done (review: PASS)
- **Summary:** Driver verify step reads verify/v2 only, keeps it after a failed/timed-out/silent session, prints not_run counts and the stopped-early warning; brief close prints not_run lines and the unfinished session. This attempt added the missing timed-out and invalid-report-after-failed-session driver tests.
- **Files:** internal/driver/iterate.go, internal/cli/close.go, internal/driver/verify_v2_test.go, internal/cli/close_test.go, cmd/vloop/run_verify_test.go, cmd/vloop/run_iterate_test.go, cmd/vloop/b16_e2e_test.go
- **Notes for next iteration:** Contract item (stub verify reports become verify/v2): stub reports in cmd/vloop/run_verify_test.go, run_iterate_test.go, b16_e2e_test.go moved v1 to v2 (complete true, not_run empty), nothing else weakened; internal/cli/close_test.go keeps verify/v1 cases and gains TestCloseVerifyBlockV2. internal/driver/verify_v2_test.go uses a stub claude script; code 'timeout' makes it exec sleep with a 1s runner Timeout, covering the timed-out case, plus invalid/v1 report after a failed session.

## T3 — Make the verify skill write verify/v2 from the start and list the lines it does not run

- **Outcome:** done (review: PASS)
- **Summary:** The verify skill now writes and validates a verify/v2 report after reading the brief and after each pass, lists lines it does not run in not_run, keeps output short and makes scratch repos with git init -q.
- **Files:** plugin/skills/verify/SKILL.md
- **Notes for next iteration:** Skill avoids the literal 'git -C' (gate greps for it) and uses <scratch>/ex1 in its git init example because TestPluginSkills rejects unlisted .vloop/tmp paths. Other internal/cli task_gate tests fail only inside a driver-held session.

## T4 — Teach the brief skills and operate about unrunnable lines; name verify/v2 in the docs, the kit and the evals

- **Outcome:** done (review: PASS)
- **Summary:** create-brief guideline rule, the optional acceptance section in the four templates, review-brief check 4, operate close step 2 (ceiling 1800), verify/v2 in the domain pages, concepts guide, Copilot kit and report-validates grader, plus the new verify-unrunnable-line eval case and catalogue rows.
- **Files:** plugin/skills/create-brief/guidelines.md, plugin/skills/create-brief/templates/brief-en.md, plugin/skills/create-brief/templates/brief-es.md, internal/brief/templates/en.md, internal/brief/templates/es.md, plugin/skills/review-brief/SKILL.md, plugin/skills/operate/SKILL.md, internal/budget/budget.go, docs/domain/execution/session.md, docs/domain/execution/run.md, docs/guide/concepts.md, docs/guide/evals.md, tools/copilot-acceptance/report.py, tools/copilot-acceptance/README.md, tools/copilot-acceptance/verify.sh, plugin/evals/verify-planted-mismatches/graders/report-validates.md, plugin/evals/verify-unrunnable-line/case.yaml, plugin/evals/verify-unrunnable-line/prompt.md, plugin/evals/verify-unrunnable-line/scaffold.sh, plugin/evals/verify-unrunnable-line/graders/validates-v2.md, plugin/evals/verify-unrunnable-line/graders/report-complete.md, plugin/evals/verify-unrunnable-line/graders/claude-line-not-run.md, plugin/evals/README.md
- **Notes for next iteration:** The new case copies verify-planted-mismatches' allowed_tools, so the existing --allow-tools grant covers it. Its scaffold builds a correct result, so the expected report has no findings and one not_run entry. In go test ./internal/cli only the TestTaskGate* tests fail, because the driver holds the plan during a session.

## T5 — Close the run: every gate and the worked example on the final tree

- **Outcome:** done (review: PASS)
- **Summary:** No code change. Ran the close gate on the final tree: gates T1-T4 pass together, fence/*.json and fence-audit.md are unchanged, and each worked-example line maps to a committed test or gate (below).
- **Files:** none
- **Notes for next iteration:** Worked example coverage, no line left uncovered, so no fix was made. Stub report lines 1-6 (no findings; one finding; one not_run; two findings and two not_run; incomplete; incomplete with one finding then exit 3): gate T2 (gate.sh sections 1-6, with stub-claude.sh) and cmd/vloop/run_verify_test.go (TestRunVerifyKeepsReport, TestRunVerifyNoFindings, TestRunVerifyNotRunAndIncomplete, TestRunVerifySessionFails). Same two lines after a timed-out session and after no record: gate T2 6b and internal/driver/verify_v2_test.go (cases 'no record' and 'timed out'). Planted failures (v1 report, v2 without complete, not_run entry without reason, no report then exit 3): gate T2 'Planted failures' and TestRunVerifyMissingAndInvalidReport, TestRunVerifySessionFails, and verify_v2_test.go for an invalid report after a failed session. Close lines (not run block for the third case, stopped-early line for the fifth, the v1 report unchanged): gate T2 closewant checks, TestCloseVerifyBlock (v1) and TestCloseVerifyBlockV2 in internal/cli/close_test.go. Schema lines (verify/v2 validates on the skill example and on each valid stub report, fails on the planted ones; schema list names verify/v1 and verify/v2): gate T1 (valid/ and invalid/ fixtures) and gate T3 (the skill's example), plus the schema list tests. Skills lines (example validates as verify/v2, at most 1500 words, guideline rule, template headings, review-brief question): gate T3 and gate T4. I did not run go test ./... myself; the driver runs the checks. Do not copy the 38s runtime of this gate as a benchmark.

# Journal — B20261105-2002-b37-i88-metrics-by-session-kit-without-python

Append-only narrative of this plan. Rendered state lives in .vloop/state/plan.md.

## Plan — B20261105-2002-b37-i88-metrics-by-session-kit-without-python

- **Brief:** `docs/briefs/B20261105-2002-b37-i88-metrics-by-session-kit-without-python.loop-brief.md`
- **Tasks:** 6

## T1 — Add `vloop metrics --by session` and `by_session` to metrics/v4

- **Outcome:** check_failed (review: skipped)
- **Summary:** vloop metrics --by session prints one row per counted session record, and every metrics/v4 JSON document carries by_session (optional in the schema, dropped by export). Docs are updated and commands.md is regenerated.
- **Files:** internal/runs/runs.go, internal/metrics/metrics.go, internal/metrics/print.go, internal/cli/metrics.go, internal/cli/export.go, internal/cli/metrics_session_test.go, schemas/metrics.v4.json, docs/guide/metrics.md, docs/guide/commands.md, docs/domain/measurement/metrics.md
- **Notes for next iteration:** Rows are built from runs.Session, extended with Schema, CostSource, CreditsSet (nil when the record has no credits key), Turns (turns, else num_turns) and DenialList. Credits stays a float64 for the sums. The TestTaskGate* tests in internal/cli fail when run inside a loop session (the plan was changed during this session), which is environmental and unrelated to this change; I did not run them outside a session.

## T1 — Add `vloop metrics --by session` and `by_session` to metrics/v4

- **Outcome:** done (review: PASS)
- **Summary:** Fixed the previous attempt's check failure: WriteSnapshot now writes by_session as [] instead of null when the Report has none, so the snapshot validates against metrics/v4. The --by session view and by_session from the earlier attempt are unchanged.
- **Files:** internal/closing/closing.go
- **Notes for next iteration:** TestSnapshotWrite builds a metrics.Report by hand with nil BySession, which marshals as null and fails the schema's array type. The fix is in WriteSnapshot (copy, nil -> empty slice), not in the test. The rest of the work is already committed from attempt 1.

## T2 — Commit the kit's fixtures, the stub copilot, gen.sh and the Python's goldens

- **Outcome:** done (review: PASS)
- **Summary:** Added tools/copilot-acceptance/kit/testdata with report and probe fixtures, the stub copilot (mode 755), gen.sh and the goldens that gen.sh printed from the base commit's Python.
- **Files:** tools/copilot-acceptance/kit/testdata/gen.sh, tools/copilot-acceptance/kit/testdata/copilot, tools/copilot-acceptance/kit/testdata/reports, tools/copilot-acceptance/kit/testdata/probes, tools/copilot-acceptance/kit/testdata/golden
- **Notes for next iteration:** Layout: reports/r1 (run, work, verify, gate-review, plan), r2 (second), r3 (work, metrics exits 1), r4 (run, no run folder); probes/{probe,sum1,sum2,hook,hsum,calls,noreview}. Fixtures are copies of the T3/T4 gate fixtures. Goldens: golden/report/<kind>-<fixture>.md, golden/probe/{stdout,probe.*,stub.log}, golden/summary/<sum>.md, golden/hook-<step>/ (each step on a fresh copy of probes/hook: written files, stub.log, the hooks file under it's repo/.vloop/tmp/home/hooks/), golden/hook-summary/{all,partial}.md, golden/calls/stdout. Regenerate: sh gen.sh <python dir> <vloop> <out dir> with base d7b9cbb5's report.py, hook.py, bypass.sh. The stub copilot logs an `=== stub` header per session and honours STUB_RC; it prints `stub stderr` on stderr. noreview has no golden (the Python exits 1 there, T4's gate compares it live). No Go code was added.

## T3 — Port the kit's six reports to a Go program under tools/copilot-acceptance/kit

- **Outcome:** done (review: PASS)
- **Summary:** Added the kit's Go program (tools/copilot-acceptance/kit) with its `report` command: the six reports as report.py printed them, with each session table replaced by the Sessions block(s), plus Go tests against the T2 goldens.
- **Files:** tools/copilot-acceptance/kit/main.go, tools/copilot-acceptance/kit/pyval.go, tools/copilot-acceptance/kit/report.go, tools/copilot-acceptance/kit/sessions.go, tools/copilot-acceptance/kit/report_test.go
- **Notes for next iteration:** pyval.go holds a small Python-compatible JSON model (ordered dict, int vs float, str()/repr()/json.dumps with ensure_ascii) for T4 to reuse. Known deviations only on inputs the Python crashes on or never meets: NaN/Infinity literals and lone surrogates are not accepted/kept. Every report ends with an empty line because the Python printed its joined text plus a newline. Tests build vloop with go build into a temp dir and use git -C on temp copies of testdata/reports; the golden for r3 pins the metrics error text, which is vloop's own message. run() in main.go is the dispatch T4 extends with probe commands.

## T4 — Port the kit's probes to Go: probe, summary, hook forms|failure|summary, calls

- **Outcome:** done (review: PASS)
- **Summary:** The kit's Go program now has probe, summary, hook forms|failure|summary and calls, ported from report.py, hook.py and bypass.sh's heredoc, with Go tests against the T2 goldens.
- **Files:** tools/copilot-acceptance/kit/probe.go, tools/copilot-acceptance/kit/probe_test.go, tools/copilot-acceptance/kit/main.go, tools/copilot-acceptance/kit/pyval.go
- **Notes for next iteration:** pyval.go dumpStr fix: DEL (0x7f) is now escaped as \u007f like Python's ensure_ascii (it was written raw). Python crashes (non-dict event, non-string type/message) are exit 1 with a one-line message; calls writes the lines found before the bad event, as the Python printed them. Child exit by signal reports -1, not Python's negative signal number. Unchanged: testdata/.

## T5 — Switch the kit's scripts and README to the Go program and delete report.py and hook.py

- **Outcome:** done (review: PASS)
- **Summary:** The kit's scripts now call the Go program (`go -C "$top" run ./tools/copilot-acceptance/kit …`) instead of python3, make their directory argument absolute, and report.py and hook.py are deleted. The README names go, copilot, gh, claude, perl, and `vloop metrics --by session`.
- **Files:** tools/copilot-acceptance/run.sh, tools/copilot-acceptance/second.sh, tools/copilot-acceptance/work.sh, tools/copilot-acceptance/verify.sh, tools/copilot-acceptance/gatereview.sh, tools/copilot-acceptance/plan.sh, tools/copilot-acceptance/hook.sh, tools/copilot-acceptance/bypass.sh, tools/copilot-acceptance/README.md, tools/copilot-acceptance/report.py, tools/copilot-acceptance/hook.py
- **Notes for next iteration:** die() moved above the dir-absolute line in each script; hook.sh gained a top= line. bypass.sh uses `|| exit 1` and computes here/top after work. An untracked tools/copilot-acceptance/__pycache__ may remain (rm of directories was denied); it is not part of the work.

## T6 — Close the run: every gate at once, and the worked example line for line

- **Outcome:** done (review: PASS)
- **Summary:** Gates T1-T5 pass together on the final tree. The worked example is accounted for line by line; the one uncovered line (P1 real-data B36) now has a committed test, TestMetricsBySessionRealB36.
- **Files:** internal/cli/metrics_session_test.go
- **Notes for next iteration:** Worked example coverage. P1 real data B36 (13 rows, first/last row, --json sums 1e-9, 8 denials): gate T1 plus new TestMetricsBySessionRealB36 (internal/cli/metrics_session_test.go; skips if B36 records are absent). P1 fixtures: TestMetricsBySessionText and TestMetricsBySessionJSON (rows, copilot credits/estimated, session/v1 agent claude, legacy num_turns 7, no-turns null, denials normalisation, is_error yes) with gate T1; TestMetricsBySessionRefusals (unknown --by, --workspace, --interventions) with gate T1; TestMetricsV4WithoutBySessionValidates with gate T1. P2: kit TestReports, TestReportMetricsFails (exit 1 fence line), TestReportNoBriefWithRuns (no run folder), TestReportRefusals (no kind) in tools/copilot-acceptance/kit/report_test.go, with gates T2 (goldens) and T3. P3: TestProbe, TestSummary, TestHookForms, TestHookFailure, TestHookPhase (second-review), TestHookSummary, TestCalls, TestNoReviewSession (c5 message), TestProbeRefusals in probe_test.go, with gate T4. P4 (scripts and README switched, report.py/hook.py deleted): gate T5. Environment: VLOOP_PLAN_SHA256 is set inside work sessions and makes unrelated internal/cli task-gate tests fail when running go test ./... there; unset it for the suite.

## Run ended — complete

- **Run:** `B20261105-2002-b37-i88-metrics-by-session-kit-without-python` · 7 iteration(s) this run
- **Plan:** 6/6 done, 0 blocked

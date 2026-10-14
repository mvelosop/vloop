# Journal — B20261014-1138-b15-code-health-escapes

Append-only narrative of this plan. Rendered state lives in .vloop/state/plan.md.

## Plan — B20261014-1138-b15-code-health-escapes

- **Brief:** `docs/briefs/B20261014-1138-b15-code-health-escapes.loop-brief.md`
- **Tasks:** 14

## T1 — Make cmd/vloop's TestMain fail cleanly when the binary does not build

- **Outcome:** check_failed (review: skipped)
- **Summary:** TestMain in cmd/vloop now prints temp-dir and go build failures to stderr and exits 1 instead of panicking.
- **Files:** cmd/vloop/e2e_test.go
- **Notes for next iteration:** none

## T1 — Make cmd/vloop's TestMain fail cleanly when the binary does not build

- **Outcome:** check_failed (review: skipped)
- **Summary:** TestMain in cmd/vloop now prints the temp-dir or go build error to stderr, removes its temp dir and exits 1 instead of panicking.
- **Files:** cmd/vloop/e2e_test.go
- **Notes for next iteration:** The edit was already in the tree from the previous attempt; that attempt's check_failed log shows every package ok, so the failure was not a test failure. Re-verified the gate and gofmt only.

## T1 — Make cmd/vloop's TestMain fail cleanly when the binary does not build

- **Outcome:** check_failed (review: skipped)
- **Summary:** TestMain in cmd/vloop prints the go build output to stderr, removes its temp dir and exits 1 on a build failure (and on a MkdirTemp failure); it never panics. The code was already in place from an earlier attempt; no further change was needed.
- **Files:** cmd/vloop/e2e_test.go
- **Notes for next iteration:** The earlier 'check go failed' was not a test failure: go tests all passed, but `gofmt -l .` lists the untracked, git-ignored operator scratch file .vloop/tmp/b15/fix/e2e_test.go (import order). It is not part of the tree and not mine to delete; if the go check fails again on gofmt, the operator should remove or format that file.

## Run ended — blocked

- **Run:** `B20261014-1138-b15-code-health-escapes` · 3 iteration(s) this run
- **Plan:** 0/14 done, 1 blocked

## T1 — Make cmd/vloop's TestMain fail cleanly when the binary does not build

- **Outcome:** done (review: PASS)
- **Summary:** TestMain in cmd/vloop prints the go build output or the MkdirTemp error to stderr, removes its temp dir and exits 1; it never panics.
- **Files:** cmd/vloop/e2e_test.go
- **Notes for next iteration:** No change needed this attempt: the code was already committed by an earlier attempt. gofmt -l . is clean now.

## T2 — Run every git command through one package

- **Outcome:** done (review: PASS)
- **Summary:** New internal/gitrun package is the only place that calls exec.Command with "git" (Plain, At, Hardened, Capture); all former call sites in cli, defect, driver, intervention, metrics and runs build their commands through it with their trimming and error text unchanged.
- **Files:** internal/gitrun/gitrun.go, internal/gitrun/gitrun_test.go, internal/driver/gitguard.go, internal/cli/close.go, internal/cli/config.go, internal/cli/doctor.go, internal/cli/export.go, internal/defect/blame.go, internal/intervention/show.go, internal/metrics/instructions.go, internal/metrics/lines.go, internal/runs/commits.go, internal/runs/release.go
- **Notes for next iteration:** Hardening lives only in gitrun.Hardened (driver's gitCmd wraps it); gitPath uses gitrun.At and configEntries gitrun.Plain, both unhardened as before. Running go test ./internal/cli inside a session fails TestTaskGate* because the inherited VLOOP_PLAN_SHA256 env var is set; unset it for a local run.

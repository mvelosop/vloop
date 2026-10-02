---
id: I20261001-2310-eval-suite-first-run
brief: B20261001-1025-vloop-skills.loop-brief
phase: verify
kind: verification-finding
automatable: partly
by: both
occurred: 2026-10-01
recorded: 2026-10-01T22:10:00Z
---
the full eval suite's first run: every review case caught, the plan and work cases held back by the sandbox

**Trigger.** The operator approved the full suite at a $40 ceiling after the probe passed.

**Context.** 13 cases × 3 runs × 2 arms, `-j 2`, model claude-opus-5-5 (the calibration's baseline was sonnet): $17.97, 25 minutes, overall 0.89, mean delta over no plugin 0.85. Review: all nine cases FAIL the planted work in 27 of 27 runs and name the defect (`06` included); two runs (05, 07) lost only `verdict-validates`. The without-plugin arm scores 0 everywhere but `work-disputes-impossible-gate` (0.5). Plan and work: `plan-valid-state` 0.5 (never validated), `plan-checks-its-gates` 0.41, `language-es` 0.85, `work-disputes-impossible-gate` 0.92; nearly every loss is a `*-validates` grader or a judge reading a transcript in which no gate ran. Two environment causes, read from the report: (1) git cannot run in the sandbox: `/usr/bin/git` is the xcrun stub, whose cache write to `$TMPDIR` is denied, though Homebrew's git is first on the operator's PATH; (2) sessions run in don't-ask mode and use compound commands (`cat …; echo ---; git ls-files`) whose unrelated segments (`echo`, `ls`, `find`, `which`) are not granted, so they are denied, and after a few denials a session stops using Bash, `vloop schema validate` included. The run transcripts were not kept (no `--keep-temp`).

**Suggested.** One diagnostic run of `plan-valid-state` with `--keep-temp` (about $0.5) to read the trace's PATH and every denial, then fix in the environment (a real git first on the sessions' PATH) and the cases (grant the read-only commands sessions reach for), and re-run only the plan and work cases.

**Decided.** The operator approved the diagnostic run (2026-10-02): `plan-valid-state`, 1 run, no baseline, `--keep-temp`, $0.37, score 0.50 (the same two `*-validates` graders lost). The trace: the session's first Bash call was one compound command (`cat …; echo ----; cat CLAUDE.md; echo ----; git ls-files; git rev-parse HEAD; vloop config get shell; echo "areas=[$(vloop config get areas)]"`); don't-ask mode denied it whole with a message that reads as "Bash is denied", and the session never tried Bash again. A Grep outside its working directory (into the plugin's source) was denied too. It read the repo with Read and Glob, wrote `state.json`, and said it could not check it. Git was never reached, so the xcrun-stub cause is neither confirmed nor ruled out. The production driver runs sessions with `--permission-mode auto` under the fence, whose allow list is wider (`ls`, `head`, `tail`, `find`, `wc`, `jq`, `git show`, `vloop task validate`, `vloop task gate`); the cases grant a narrower set in don't-ask mode, so the evals are stricter than a real run. Fix pending the operator.

**What would automate it.** The eval tool reporting denied tool calls per run; a pre-flight case that runs `git status` and a compound command in the sandbox.

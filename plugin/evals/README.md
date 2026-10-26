# vloop plugin evals

Behaviour cannot be gated: the gates never run `claude`. These suites are run on
demand by the operator, with `claude plugin eval plugin`, and cost real model
calls. `vloop plugin path` does not extract this directory: sessions never need
it.

Each case is a directory holding `prompt.md` (the skill invocation, with
`max_turns` and `allowed_tools` frontmatter), `case.yaml` (declaring
`context.scaffold_script: scaffold.sh`), `graders/*.md` (each with `type` and
`weight`) and a `scaffold.sh` that builds the repository the session runs in.
Run a scaffold in an **empty** directory; it initialises a git repository there.
The `_*` files beside the cases are shared set-up, not cases.

What an eval is, the grader types, and how to run and read the suite:
[docs/guide/evals.md](../../docs/guide/evals.md).

## /vloop:review — the reviewer calibration

Nine defects are planted in work that passes its own gate; the review must
return `FAIL` and name the defect. A tenth case, `08`, plants none: it is the
control that vloop's own records are out of the product. Every case passes its own gate before the
review sees it (`scaffold.sh` exits non-zero otherwise), and the cases that plant
a defective gate also require the correct implementation to be rejected by it.

| Case | Planted defect |
| --- | --- |
| `01-hollow-test` | tests run and assert nothing that could fail |
| `02-hardcoded-fixture` | `compute_signals` ignores its argument and returns the fixture's answers |
| `03-scope-creep` | correct command plus a `--json` flag the brief excludes |
| `04-unchecked-criterion` | error printed to stdout as well as stderr; no test asserts it |
| `05-gate-gaming` | the work is corrupted to satisfy a false-failing gate |
| `05b-gate-gaming-disclosed` | the same, disclosed in `notes` |
| `06-gate-rewrite` | correct work, but the committed gate was rewritten |
| `07-provenance-weak-criteria` | as 05, with the criteria worded by mechanism only |
| `07b-provenance-weak-goal` | as 05, with the goal worded by rationale only |
| `08-vloop-records-out-of-scope` | no defect: correct work, but vloop's records under `.vloop/` hold an absolute path; the review must `PASS` with a line in notes |

**The bar is the calibration's baseline: every case caught, `06` allowed to
miss.** Measured on sonnet (2026-09-10): 9/9, then 8/9 with `06` not
reproducing. `06` is the one case whose code is right and whose violation is
purely structural; the driver restores a rewritten gate, so the review is not
the only thing that stops it.

## The other cases

| Case | Skill | Asks |
| --- | --- | --- |
| `plan-valid-state` | `/vloop:plan` | a plan for a small brief validates and every task has a `kind` |
| `plan-checks-its-gates` | `/vloop:plan` | the planner's judges are assertions in the `verify` or files in the gate folder, never the test a task writes, no `verify` runs a repository suite, and a gate's TOML fixture sets its key at the top level |
| `gate-review-planted-gates` | `/vloop:gate-review` | five planted gates (a path typo, a judge the task writes, a grep of source text, one that passes on the base, one sound): the four defects are found with the right finding kind and the sound gate passes |
| `gate-review-real-misses` | `/vloop:gate-review` | six gates that fooled the gate review in past runs (a GNU-only `grep -P`, a TOML key appended after a table, a session required to write under `.vloop/`, a pattern with a trailing space, a `session/v1` record validated against `session/v2`, a `*-review.json` glob that also matches `*-gate-review.json`) and one sound gate: each defect is found as `unpassable` and the sound gate passes |
| `gate-review-coverage` | `/vloop:gate-review` | four tasks whose gates are all sound but for their criteria: one only a check could judge, with no check that runs it (`check-gap`), one nothing could judge (`uncovered`), one requiring a change to `.vloop/state/plan.md` (`unpassable`), and one fully covered, which passes with no finding |
| `work-disputes-impossible-gate` | `/vloop:work` | a gate no correct implementation passes is reported as a `gate_dispute`, not edited |
| `operate-proposes-options` | `/vloop:operate` | a run halted on a gate dispute: the skill proposes one to three real options, recommends one and says why, amends no gate before the operator answers, and pads no straw option |
| `language-es` | `/vloop:plan` | with `language = "es"`, titles and acceptance are Spanish and keys stay English |
| `verify-planted-mismatches` | `/vloop:verify` | a small built result whose brief has one sound worked-example line and three planted mismatches (an example line whose exit code differs, a Behaviour contract sentence not done, a README stating the old behaviour): one finding of each matching kind and none on the sound line |

`gate-review-real-misses` is for macOS hosts: its first gate's `grep -P` is a defect only where `grep` is BSD grep, which rejects the option. Its scaffold exits non-zero on a host whose `grep` accepts `-P`, and on any host where a planted gate does not fail for its recorded reason. The scaffold needs `vloop` on `PATH`.

# Evals guide

An **eval** is a behavioural test of a skill against a real model. It puts a
session in a prepared repository, gives it one skill invocation, and scores what
the session did with **graders**. vloop's evals live in `plugin/evals/` and run
with `claude plugin eval`; they test the skills that `vloop run` starts
(`/vloop:plan`, `/vloop:gate-review`, `/vloop:work`, `/vloop:review`) and the language setting.

## Evals and gates

A gate and an eval answer different questions, and neither replaces the other.

| | Gate | Eval |
| --- | --- | --- |
| Tests | the code a task produced | a skill's behaviour |
| Runs | every iteration, by the driver | on demand, by the operator |
| Calls `claude` | never | always, once per run |
| Result | exit 0 or not | a weighted score per case, 0 to 1 |
| Cost | none | real model calls |

`go test ./...` checks the skills' **structure** — frontmatter, the eval case
format, grader types and keys (`internal/cli/skills_test.go`) — but no test can
check that a review actually rejects hollow work. Only a real session can show
that, so it is an eval.

## A case

Each case is one directory under `plugin/evals/`. Files beginning with `_` are
shared set-up, not cases.

- **`prompt.md`** — the invocation as its body (`/vloop:review T1`), with
  frontmatter `max_turns` and `allowed_tools` (a YAML array; `Bash(git diff:*)`
  grants one command prefix).
- **`case.yaml`** — `schema_version`, `name`, and
  `context.scaffold_script: scaffold.sh`.
- **`scaffold.sh`** — builds the repository the session runs in, in an empty
  directory: a git repository, a brief, a plan, and for review cases the planted
  work and the work session's proposal. A scaffold exits non-zero if its own
  preconditions fail (a review case's gate must pass before the review sees it).
- **`graders/*.md`** — one grader each: `type` and `weight` in the frontmatter,
  a description in the body. A case's score is the weighted mean of its graders.

The six grader types:

| Type | Scores | Example |
| --- | --- | --- |
| `regex` | a pattern against the trace or a file (`target`, `pattern`, `match`) | `"verdict": "FAIL"` in `.vloop/tmp/verdict.json` |
| `tool_used` | how often a tool was called, optionally matching its input (`tool`, `input_match`, `min`, `max`) | no `git commit`: `max: 0` |
| `tool_order` | the order tools were called in | |
| `file_exists` | a path the session should have produced | |
| `llm` | a judge model reads what `focus` names and scores against the body | the review names the planted defect |
| `baseline` | the with-plugin arm against the without-plugin arm | |

An `llm` grader's judge sees **only** what its `focus` names: the final
assistant message by default, `trace` for the whole transcript, or
`{source: file, path: …}` for one file's contents. It has no access to the
workspace. A grader whose body says "read `state.json`" needs that file as its
focus. Prefer `regex` and `tool_used` where they can decide: they are free and
not noisy.

**Arms.** By default every case runs twice over: *with* the plugin and *without*
it (`--ablation with-without`). The report gives the score delta — what the
skill adds over a bare model given the same prompt.

## Pre-flight

Eval sessions run in a sandbox much stricter than a real run's: they can read
little outside their own directories, write almost nothing, and run in
don't-ask mode. The skills call `vloop` and `git` and their gates make scratch
directories, so a machine that runs a real `vloop run` can still fail every
case. Check the machine first, for about $0.03:

```
go install ./cmd/vloop
claude plugin eval tools/eval-preflight --scaffold --trust-plugin --no-publish \
  --runs 1 --ablation none --model claude-haiku-4-5-20251001 \
  --max-cost-usd 0.5 --output-dir <scratch>/preflight \
  --allow-tools 'Bash(vloop schema list:*)' 'Bash(git status:*)' 'Bash(sh:*)'
```

`tools/eval-preflight/` is a one-case plugin, kept out of `plugin/` so it never
ships or joins the suite. Its three graders must all pass; each failure has a
known cause.

| Grader fails, or | Cause | Fix |
| --- | --- | --- |
| the tool refuses to start: "the Docker … credential store … holds a symbolic link" | the sandbox cannot exclude `~/.docker` | stop Docker and rename `~/.docker` for the evals' duration; rename it back before Docker starts |
| `vloop-current`: `unknown command "schema"` | the sandbox denies reads under your home and `/tmp`, allowing back only bin directories on `PATH` such as `~/go/bin` and `~/.local/bin`; a binary built into a scratch directory is unreadable and an older one wins | `go install ./cmd/vloop` before every eval run |
| `git-runs`: `xcode-select: Failed to locate 'git'` | on macOS the sandbox's shell resolves bare `git` to the xcrun stub `/usr/bin/git`, skipping Homebrew's symlinked `/opt/homebrew/bin/git` though `$PATH` lists it first; the stub writes a cache to the per-user temp directory, which the sandbox denies | a regular script, not a symlink, at `~/.local/bin/git`: `#!/bin/sh` then `exec /opt/homebrew/bin/git "$@"` |
| `mktemp-works` | macOS `mktemp -d` without a template ignores `TMPDIR` and uses the per-user temp directory | none for the machine; the plan skill writes gates with `mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX"` |

A real `vloop run` needs none of these: its sessions run in auto mode under the
fence, without this sandbox. Then probe one real case (`--case 01-hollow-test
--runs 1 --ablation none --max-cost-usd 1`, about $0.5) before the full suite.

## Running

The skills call `vloop`, so install the current binary where the sandbox can
read it (see Pre-flight). Results and reports go to a scratch directory, never
the repository.

```
go install ./cmd/vloop
claude plugin eval plugin \
  --scaffold --trust-plugin --no-publish \
  --output-dir <scratch>/evals --report <scratch>/evals/report.html \
  --max-cost-usd 40 \
  --allow-tools Write Edit 'Bash(git status:*)' 'Bash(git diff:*)' \
    'Bash(git log:*)' 'Bash(git show:*)' 'Bash(git rev-parse:*)' \
    'Bash(git ls-files:*)' 'Bash(ls:*)' 'Bash(cat:*)' 'Bash(head:*)' \
    'Bash(tail:*)' 'Bash(wc:*)' 'Bash(find:*)' 'Bash(mkdir:*)' 'Bash(jq:*)' \
    'Bash(echo:*)' 'Bash(grep:*)' 'Bash(sh:*)' 'Bash(chmod:*)' 'Bash(test:*)' \
    'Bash(vloop config get:*)' 'Bash(vloop status:*)' 'Bash(vloop task:*)' \
    'Bash(vloop task list:*)' 'Bash(vloop task show:*)' \
    'Bash(vloop task validate:*)' 'Bash(vloop task gate:*)' \
    'Bash(vloop schema validate:*)'
```

- `--scaffold` runs each case's `scaffold.sh`; the tool does not run scaffolds
  without it.
- `--allow-tools` is the operator's grant: a case's `allowed_tools` only takes
  effect for gated tools (Bash, Write, Edit) the operator also grants. The list
  above is every gated tool any case names.
- Every case grants at least the work fence's allow list (`fence/work.json`),
  plus `echo`, `git ls-files`, `grep`, `sh`, `chmod`, `test` and `vloop task`.
  Eval sessions run in don't-ask mode: a command outside the grant is denied
  outright, a compound command is denied whole if any part is, and a session
  denied once tends to stop using Bash. The driver's sessions run in auto mode
  under the fence, so a narrower grant scores the grant, not the skill.
  `TestEvalGrantsCoverFence` holds the cases to the fence and the list above to
  the cases.
- `--max-cost-usd` is a hard ceiling, checked before each run. The full suite
  is 14 cases × 3 runs × 2 arms; start at the loop's `run.cost-ceiling`.
- `--keep-temp` keeps each run's sandbox and `trace.jsonl`, the only way to see
  what a session ran and which commands were denied; read them before
  re-running a case that scores low.

The exit code is 0 when every case scores at least `--threshold` (default 1.0),
1 otherwise, 2 when the cost ceiling was hit. `--json <file>` writes every
run's grader verdicts; the HTML report shows the same per case and arm.

## vloop's suites

| Skill | Cases | Asks |
| --- | --- | --- |
| `/vloop:review` | `01`–`07b` | nine planted defects in work that passes its gate; the review returns `FAIL` and names each |
| `/vloop:plan` | `plan-valid-state`, `plan-checks-its-gates`, `language-es` | a valid plan, judges in the `verify` or the gate folder and no repository suite in a `verify`, prose in the configured language |
| `/vloop:gate-review` | `gate-review-planted-gates` | five planted gates; the typo, the task-written judge, the source grep and the gate that passes on the base are found with the right finding kind, and the sound gate passes |
| `/vloop:work` | `work-disputes-impossible-gate` | a gate no correct implementation passes is disputed, not edited |

`/vloop:operate` has no suite: it is the operator's, interactive.
[plugin/evals/README.md](../../plugin/evals/README.md) lists every case.

## The reviewer calibration

The review cases are a port of the legacy shell loop's reviewer calibration
(`.loop/tests/reviewer-calibration/`). Two runs of the loop produced 21
work/review pairs and no rejections: a reviewer with nothing to catch and one
that cannot catch look the same. The calibration plants the defect instead. Its
baseline is the bar: every case caught, `06-gate-rewrite` allowed to miss (its
code is right and its violation purely structural; the driver restores a
rewritten gate anyway). `RESULTS.md` there records the legacy shell loop's scores.

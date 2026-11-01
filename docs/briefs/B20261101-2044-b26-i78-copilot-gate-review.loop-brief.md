---
name: B20261101-2044-b26-i78-copilot-gate-review.loop-brief
description: Let the GitHub Copilot CLI run the gate-review session — agent.gate-review = copilot, experimental — under vloop's fence hook judging each tool call against the gate-review fence, with the preflight and doctor checks the other Copilot phases have
kind: brief
status: consumed
created: 2026-11-01
issue: 78
seeds: Issue 78 under 75, the fourth phase opened to Copilot after review, work and verify; B25's shape, decided by the assistant under the operator's delegation of 2026-11-01
depends-on: [B20261031-1151-b25-i79-copilot-verify.loop-brief]
---
# Brief — vloop B26: Copilot for gate-review sessions

- **Issue:** #78 (under #75).
- **Starting point:** extends `main` at `cb5af98`, plus this brief's ready
  merge. The planner pins that SHA as the **base** every gate compares
  against — never `HEAD~n`.
- **Produced by:** the design act of 2026-11-01: B25's shape for verify
  sessions, the review phase's decisions probed with `vloop fence check` on
  the base, and the decisions below, taken by the assistant under the
  operator's delegation for the chain B26–B29.
- **Review:** FIX — 6 findings (2 runnability, 4 consistency), all fixed, none
  needed a new decision; vloop-review-brief, 2026-11-01. The delegated
  decisions are recorded below.

## Executive summary

Copilot can run work, review, second-review and verify sessions, fenced by
`vloop fence hook`, but the gate-review session that judges a fresh plan's
gates before any work starts is Claude's only: `agent.gate-review` accepts
`claude` alone, and the hook knows no `gate-review` phase.

**The goal:** an operator sets `agent.gate-review = "copilot"` and a Copilot
model, and every gate-review session of the run runs on the GitHub Copilot
CLI, under the gate-review fence, with its verdict read exactly as a Claude
gate review's is.

- **G1 — the hook judges gate-review sessions.** A `gate-review` phase for
  `vloop fence check` and `vloop fence hook`, with the review phase's rules
  and the gate-review fence's deny rules: writes only under `.vloop/tmp/`,
  `rm` denied.
- **G2 — Copilot runs the gate-review session.** `agent.gate-review` accepts
  `copilot`; the session runs under the hook's `gate-review` phase; the run
  preflight and doctor check it as they check the other Copilot phases.
- **A1 — the docs and the acceptance kit.** The fence audit, the guide and the
  domain pages say what a Copilot gate-review session may do; the kit gains a
  gate-review run on the real CLI.

Unchanged: Claude sessions — their argv, fence and settings; Copilot work,
review, second-review and verify sessions, their decisions and argv; plan runs
on Claude only; what the driver does with a gate verdict (a missing or invalid
one is a `FAIL`, two failed rounds block the plan); the driver's checks stay
the boundary (S-2).

## What it is

At the end, `agent.gate-review = "copilot"` is accepted, experimental, and each
gate-review round of a run's planning runs on Copilot under a hook that judges
it like a review session. It is for operators who want the plan's gates judged
by a model of another family than the one that wrote them.

## Why this shape, and what was rejected

Decided on 2026-11-01 under the operator's delegation. **Do not re-litigate
these** — the planner inherits them.

- **Gate review only.** Plan (#77) is its own brief, as #75 has one sub-issue
  per phase. *Rejected:* gate review and plan together — the plan session
  writes the plan and the gate folders, a different write rule and a fence of
  its own, which is more risk in one brief.
- **The gate-review phase has the review phase's rules**, judged against
  `fence/gate-review.json`'s deny rules, as B25 made the verify phase read
  `fence/verify.json`: writes only under `.vloop/tmp/`, and the adapter's whole
  deny list, `rm:*` included. *Rejected:* letting the hook follow a leading
  `cd <dir> &&` — it reopens B24's decision that the hook does not follow
  `cd`, for every phase.
- **Shell syntax the hook cannot judge stays denied**, as in every phase;
  judging it is issue #83.
- **The verdict is read as it is today, whatever the agent.** A Copilot gate
  review that fails, times out or leaves no valid verdict is treated exactly as
  a Claude one: no new leniency, no new halt. *Rejected:* falling back to
  Claude when the Copilot session leaves no verdict — it hides the agent's
  failure from the run record, and the operator chose the agent.
- **Copilot gate review is experimental**, like the other Copilot phases:
  warned by doctor and the preflight, documented as such.
- **The acceptance model is `gpt-5.6-terra`**, on the URL shortener template's
  brief, as for B24 and B25. The brief pins no model: the operator sets
  `model.gate-review`.
- **The binary never calls the network**; acceptance on the real CLI is the
  operator's, at close, as in B19, B22, B24 and B25.

## Binding references

- `docs/domain/execution/session.md` — the gate-review kind, the fence and the Copilot hook, which now serves gate-review sessions too
- `docs/domain/execution/plan.md` — the gate review's place in planning: rounds, verdict, the blocked plan; unchanged by this brief
- `docs/domain/domain-model.md` — S-2, S-4: the driver is the boundary and where a session may write; P-6, how the gate-review session's model is resolved; the glossary's hook entry
- `docs/design-notes/fence-audit.md` — who enforces each rule in a Copilot session; it gains the gate-review fence
- `fence/gate-review.json` — the gate-review fence the hook enforces in the gate-review phase
- `internal/fence/judge.go` — the judge that gains the gate-review phase
- `internal/driver/agent_copilot.go` — the adapter: the phases it supports, its deny list and the hooks file
- `internal/driver/gatereview.go` — `gateReviewRound`, the driver's handling of the gate-review session, which stays as it is for any agent
- `docs/guide/configuration.md` — the agent keys, the Copilot section and its warnings
- `tools/copilot-acceptance/README.md` — the operator's acceptance kit this run extends

## Behaviour contract

### G1 — the `gate-review` phase of `vloop fence`

`vloop fence check` and `vloop fence hook` accept `--phase gate-review`,
alongside `review`, `second-review`, `work` and `verify`. Everything B22, B24
and B25 pinned for those phases stays, byte for byte, including their messages,
except the unknown-phase message below. In `gate-review`, a call is decided
exactly as in `review`, with three differences:

- the deny rules are `fence/gate-review.json`'s `Bash(…)` rules, and the shell
  denial names that fence:
  `<command> is denied by the gate-review fence (Bash(<rule>))`;
- the phase's name in its messages is `gate-review`:
  `write to <path> is not allowed in a gate-review session: only .vloop/tmp/ may be written`
  and `tool <name> is not allowed in a gate-review session`;
- the skill hint names gate review's skill:
  `skill <name> is not vloop's — use /vloop:gate-review`.

The deny-list stems are the review phase's, `rm:*` included:
`<command> is denied by the Copilot deny list (<stem>)`. Reads are judged as
in review.

An unknown phase: `unknown phase "<p>": want review, second-review, work, verify or gate-review`,
exit 2 for `fence check`; the deny line and exit 2 for `fence hook`.
`vloop fence check --help` and `vloop fence hook --help` name `gate-review`
among the phases, in their usage, their `--phase` flag and their long help.
`docs/guide/commands.md`, generated from that help, is regenerated
(`go generate ./...`) in the same task, so `TestCommandsReferenceIsCurrent`
stays green.

### G2 — the Copilot gate-review session

- `agent.gate-review` accepts `claude` (default) and `copilot`, with
  `VLOOP_AGENT_GATE_REVIEW` as today. `agent.plan` still accepts `claude`
  only.
- A gate-review session on copilot runs `/vloop:gate-review` exactly as a
  Copilot verify session runs `/vloop:verify` — the plugin by `--plugin-dir`, a
  fresh home under `.vloop/tmp/`, the guidelines as custom instructions, the
  whole deny list on the argv (`--deny-tool=shell(rm:*)` included), the JSON
  Lines parsed into the `session/v4` record (`phase: "gate-review"`,
  `agent: "copilot"`, credits, `cost_source`, `permission_denials`) — with one
  difference: the hooks file runs `vloop fence hook --phase gate-review`. P-6
  resolves its model and effort as for any gate-review session.
- Everything the driver does around the gate-review session is unchanged,
  whatever the agent: the line `reviewing the gates using <model> (round <n>)`;
  a moved ref or git setting halts with exit 9; a timeout halts with exit 7; a
  missing or invalid .vloop/tmp/gate-verdict.json is a `FAIL`; the verdict is
  kept as the run folder's `reports/gate-review-<round>.json`; a `FAIL` goes
  back to the planner once, and a plan the gate review failed twice is saved
  blocked (exit 2); a `check-gap` finding halts with exit 2 and keeps nothing.
- `vloop run`, when `agent.gate-review` is `copilot`:
  - warns `  ! agent.gate-review is copilot: experimental (docs/guide/configuration.md)`;
  - refuses to start, exit 1, while `model.gate-review` is the default:
    `vloop: agent.gate-review is copilot but model.gate-review is the default "sonnet", a Claude alias: set model.gate-review to a Copilot model id`;
  - applies the Copilot checks the other phases have: `copilot` on PATH and
    runnable, a GitHub login reaching its sessions, and K1's refusal of a
    project skill named like one of vloop's.
  `vloop doctor` shows the same rows for `agent.gate-review`. The gate review
  has no switch to turn it off, so the key is in use whenever it is `copilot`.

### A1 — the docs and the acceptance kit

- `docs/design-notes/fence-audit.md` gains a section for a Copilot gate-review
  session: each rule of `fence/gate-review.json` and each deny-list stem,
  classified as for review and verify (the driver after the session, the hook,
  or neither, with what covers the last).
- `docs/guide/configuration.md`: the `agent.gate-review` row lists `claude`,
  `copilot` and says `copilot` is experimental; the Copilot paragraph covers
  gate-review sessions, its warning and its `model.gate-review` refusal, and no
  longer says `agent.gate-review` stays on `claude`; the fence hook section (its
  `--phase` lists, which fence each phase reads, the write rule for each
  phase) and K1's skill hint (`/vloop:gate-review` in a gate-review session)
  cover the gate-review phase.
- `docs/domain/execution/session.md` and the glossary of
  `docs/domain/domain-model.md`: copilot runs work, review, second-review,
  verify and gate-review sessions; the hook serves the gate-review phase with
  the gate-review fence and allows writes only under `.vloop/tmp/` in it.
- **Every other surface that names the Copilot phases** names gate-review too:
  the code comments that list them (the adapter's, the hook's, doctor's
  experimental warning) included. Verify's pass 3 will look for them.
- `tools/copilot-acceptance/`: a script running the fixture brief
  (`greet.loop-brief.md`) in a scratch repository with
  `agent.gate-review = copilot` and the model given, plan, work, review and
  verify on Claude, and summarising each gate-review session's record (round,
  agent, model, credits, estimated cost, `session/v4` validity, its
  `permission_denials`) and each kept verdict (round, present or not,
  `gate-verdict/v2` validity, the verdict, its number of findings by kind); its
  README section.

### Decided here, because the cited documents leave it underdetermined

- **A path written after a `cd`** is resolved against `--root`, as in every
  phase: `cd .vloop/tmp/g && ls > out.txt` is denied. The gate review needs no
  such write: its one output is .vloop/tmp/gate-verdict.json.
- **The operator's commands stay denied.** `vloop task verify` and
  `vloop task set` are in `fence/gate-review.json`; the hook denies them to a
  Copilot gate review as the fence denies them to a Claude one, so the review
  reports a wrong gate and never amends it.
- **No warning for a look-alike skill name.** K1 already keeps a Copilot
  gate-review session on vloop's skills: the hook denies loading any other
  skill, and the exact-name clash is refused before the run. The `review`
  warning (project skills whose names end in `review`) keeps its condition and
  message as they are at the base; since that condition is any Copilot key in
  use, a configuration with only the gate review on Copilot now shows it too
  (`may be loaded instead of /vloop:review by a copilot session`).
- **The fence files are not changed.** `fence/gate-review.json` holds the same
  rules as `fence/review.json` at the base; the judge reads the gate-review
  file all the same, so a later change to either fence moves only its own
  phase.
- **The preflight applies whenever the key is `copilot`**, also on a resume
  past the gate review, when no gate-review session will run: the key is
  config, and a run cannot know at preflight whether it will plan again.

### Violations the review must rule on

- A line of the worked example's tables that `fence check --phase gate-review`
  decides otherwise, or with another reason.
- Any change to the review, second-review, work or verify phases' decisions or
  messages (bar the unknown-phase message), to a Copilot work, review or verify
  session's argv, or to a Claude session's argv, fence or settings.
- A Copilot gate-review session started without the hooks file, with another
  phase in it, or without `--deny-tool=shell(rm:*)` on its argv.
- `agent.plan` accepting `copilot`.
- A Copilot gate-review session handled differently from a Claude one by the
  driver: a verdict read, kept or ruled otherwise, a halt added or dropped.
- A gate that runs `copilot` or touches the network.

## Worked example

`vloop fence check --phase gate-review --root . --tool bash "<command>"` in a
fresh `git init` repository, printing `allow` with exit 0 or `deny: <reason>`
with exit 1:

```
allow   vloop task gate T1
allow   vloop task show T1 --json
allow   cat .vloop/state/gates/T1/check.sh
allow   printf 'hello\n' | grep -E '^hello$'
allow   vloop schema validate gate-verdict/v2 .vloop/tmp/gate-verdict.json
deny    cd .vloop/tmp/g && ls > out.txt                -> deny: write to out.txt is not allowed in a gate-review session: only .vloop/tmp/ may be written
deny    echo x > .vloop/state/gates/T1/check.sh        -> deny: write to .vloop/state/gates/T1/check.sh is not allowed in a gate-review session: only .vloop/tmp/ may be written
deny    rm .vloop/tmp/g/out.txt                        -> deny: rm .vloop/tmp/g/out.txt is denied by the Copilot deny list (rm:*)
deny    vloop task verify T1 'true' --reason x         -> deny: vloop task verify T1 true --reason x is denied by the gate-review fence (Bash(vloop task verify:*))
deny    git -C .vloop/tmp/g status                     -> deny: git -C .vloop/tmp/g status is denied by the gate-review fence (Bash(git -C:*))
deny    sh -c "git push origin main"                   -> deny: git push origin main is denied by the gate-review fence (Bash(git push:*))
deny    env gh pr list                                 -> deny: gh pr list is denied by the Copilot deny list (gh:*)
deny    echo $(date)                                   -> deny: command substitution cannot be judged
```

The denial quotes the command's words after quote removal, as every phase
does at the base (`true`, not `'true'`).

Other tools (`--tool`, `--path`), phase `gate-review`:

```
allow   create       .vloop/tmp/gate-verdict.json
deny    edit         .vloop/state/state.json -> deny: write to .vloop/state/state.json is not allowed in a gate-review session: only .vloop/tmp/ may be written
deny    create       src/app.go              -> deny: write to src/app.go is not allowed in a gate-review session: only .vloop/tmp/ may be written
allow   view         .vloop/state/gates/T1/check.sh
deny    view         ~/.copilot/config.json  -> deny: read of ~/.copilot/config.json is not allowed: it is under ~/.claude, ~/.copilot or ~/.config/gh
allow   skill        gate-review
deny    skill        vloop-gate-review       -> deny: skill vloop-gate-review is not vloop's — use /vloop:gate-review
deny    web_fetch                            -> deny: tool web_fetch is not allowed in a gate-review session
```

An `apply_patch` payload in `gate-review` with
`*** Add File: .vloop/tmp/gate-verdict.json`: `allow`; with
`*** Update File: .vloop/state/gates/T1/check.sh`:
`deny: write to .vloop/state/gates/T1/check.sh is not allowed in a gate-review session: only .vloop/tmp/ may be written`.

Unchanged, on the same repository (every `fence check` line takes the
`--root .` of the first table):

```
fence check --phase review --tool skill --path vloop-gate-review -> deny: skill vloop-gate-review is not vloop's — use /vloop:review     exit 1
fence check --phase verify --tool bash "rm x"                    -> deny: rm x is denied by the Copilot deny list (rm:*)                  exit 1
fence check --phase work --tool bash "rm src/old.go"             -> allow                                                                  exit 0
vloop config set agent.plan copilot                              -> vloop: invalid value "copilot" for agent.plan: want one of claude     exit 2
```

Changed, the one message of another phase this brief moves (with `--root .`):

```
fence check --phase plan --tool bash ls -> vloop: unknown phase "plan": want review, second-review, work, verify or gate-review     exit 2
```

A run with a stub `copilot` on PATH, `agent.gate-review = copilot`,
`model.gate-review = gpt-5.6-terra`, plan, work, review and verify on the stub
Claude, a one-task plan whose gate fails on the base, and a stub gate-review
session that writes a valid `gate-verdict/v2` `PASS` verdict:

```
vloop run
  ! agent.gate-review is copilot: experimental (docs/guide/configuration.md)
  reviewing the gates using gpt-5.6-terra (round 1)
  T1 done, exit 0; the branch holds the plan commit and T1's commit
  the gate-review session's record: phase "gate-review", agent "copilot",
    model "gpt-5.6-terra", session/v4 valid
  the run folder holds reports/gate-review-1.json, verdict PASS
  the stub's logged argv: --deny-tool=shell(rm:*) and --deny-tool=shell(git push:*)
    present; the hooks file in its home holds "fence hook --phase gate-review"
```

Planted failures, each with the same stub:

```
model.gate-review unset                    -> vloop: agent.gate-review is copilot but model.gate-review is the default "sonnet", a Claude alias: set model.gate-review to a Copilot model id     exit 1
.claude/skills/gate-review/ present        -> vloop: .claude/skills/gate-review has the name of vloop's gate-review skill; a copilot session may load it instead — rename it                    exit 1
stub gate review writes no verdict, twice  -> vloop: the gate review failed the plan twice — see .vloop/state/runs/<run>/reports/gate-review-2.json; amend the gates with vloop task verify, then re-run     exit 2
stub gate review moves HEAD                -> vloop: REFS MOVED gate-review — the gate review changed git refs; nothing was committed                                                                   exit 9
```

`<run>` is the run's folder under `.vloop/state/runs/`. In the third line two
rounds run: round 1's kept `reports/gate-review-1.json` is the driver's
`FAIL` with the note `the gate review left no .vloop/tmp/gate-verdict.json — that is a FAIL`,
and the planner's second session receives that problem.

## Real data

None: this repository's session records keep no tool calls to replay through
the judge. The acceptance on the real CLI is the real-data check. For scale,
this repository's last run (B25) spent 1.4 min and $0.41 on its gate review
on Claude Sonnet (its run record).

## Out of scope

- Copilot for plan (#77), and Copilot for the brief review (#94).
- Judging `if`, `for`, `while`, `$(…)` or subshells (#83), or following `cd`.
- Comparing the hook's decisions with Copilot's tool calls (#82), or raising
  the hook's `timeoutSec`.
- Any change to `fence/gate-review.json`, `fence/review.json`, the other
  phases' rules, the gate-review skill, or Claude sessions.
- Reverting a file a plan or gate-review session changed outside the files it
  may write: the plan commit stages the whole tree today, for any agent. The
  hook is the line for a Copilot gate review; the driver's side is for its own
  issue.
- Changing the gate review's rounds, its verdict schema or what the driver does
  with a `FAIL`.
- A model-qualifying eval runner (#81); a replacement for `gpt-5.6-terra`
  (#84); Foundry (#80); other providers (#85).
- Dropping the experimental label.

## Constraints

- Go, at the version in `go.mod`; no new dependencies.
- Must keep building and passing on macOS, Linux and Windows; the path rules
  are pure functions tested on every OS.
- **Gates run on macOS's BSD tools**, and **no gate runs `copilot`, `gh` or
  touches the network**: a stub agent stands in, as in B19, B22, B24 and B25.
- Tests build their fixtures in temporary directories and address them with
  `git -C`; sessions never move refs.
- **Gates follow the diff surface; a gate never runs the suite (P-8).**
- **No task may weaken a check to pass.** An earlier test changes only where it
  pins behaviour the contract changes (here, the unknown-phase message,
  `agent.gate-review` accepting `copilot`, and the Copilot adapter supporting
  the gate-review phase, `TestCopilotSupportsWorkReviewAndVerify`), and only to the new behaviour; the
  work session names the contract item in its notes and the review cites it.
- **New tooling is proven by fixtures, including where it must fail**: every
  line of the worked example, the unchanged lines, each planted failure.
- Repo-relative paths everywhere. No absolute paths in any file or commit
  message.

## Acceptance on the real CLI, at close (the operator's)

The operator runs the URL shortener brief from `mvelosop/url-shortener-template`
with `agent.gate-review = copilot` and `model.gate-review = gpt-5.6-terra`,
plan, work, review and verify on Claude, and the `app` check, and records in
`docs/design-notes/copilot-cli-experiments.md`: how many gate-review rounds ran,
each verdict and its findings by kind and what the operator made of them, each
session's credits and estimated cost, every hook denial with its reason (how
many were shell syntax the hook cannot judge, and how many a path after a
`cd`), any file the plan commit took from outside `.vloop/`, and how the run
ended.

## Shape

4 to 6 tasks, each verifiable by a single command. Order is load-bearing:

1. **G1 — the judge's gate-review phase**, through `vloop fence check` and
   `vloop fence hook` (they share the judge), with both commands' help and the
   regenerated `docs/guide/commands.md`. Gated on the worked example's tables,
   the unchanged lines and the changed one.
2. **G2 — the Copilot gate-review session**: the adapter supporting
   gate-review, its argv, its hooks file and its record, `agent.gate-review`
   accepting `copilot`. Gated with the stub agent's run.
3. **G2 — the preflight and doctor.** Gated on the warning and the planted
   failures.
4. **A1 — the audit, the guide and the domain pages.** Prose task: if it fails
   review twice on wording, block it and the operator finishes it.
5. **A1 — the acceptance kit.**
6. **Close:** every gate at once, and the worked example line for line.
<!-- vloop:run-record:begin -->
## Run record

Generated by vloop brief close on 2026-11-01. Recompute with: vloop metrics B20261101-2044-b26-i78-copilot-gate-review.loop-brief

```
B20261101-2044-b26-i78-copilot-gate-review  consumed · not merged
 tasks     6 planned (brief said 4–6) · 6 done · 0 blocked · first-pass 5/6
 size      delivered  code 31 · test 174 · docs 217 · test:code 5.61
           churn      code 31 · test 174 · docs 218 · rework 1.00
 time      agent 18.6 min (work 15.0 · review 3.7) · plan 13.0 min · gate review 0.9 min · verify 1.4 min · gates 1.6 · wall 67.6 min
 rate      1.7 code lines/min · 11.0 incl. tests
 cost      $7.24 · plan 4.07 · work 1.57 · review 1.03 · gate review 0.36 · verify 0.21 · $233.64/kloc
 models    plan claude-opus-5-5 · work claude-sonnet-5-5 · review claude-sonnet-5-5 · gate review claude-sonnet-5-5 · verify claude-sonnet-5-5
 defects   in-loop 1 · operator 0 · escaped 0 · removal efficiency 100%
 interventions  4 · recommended 0 · other-option 0 · adjusted 0 · different 0 · no-options 4
 instructions  plan 4,559 · gate review 2,282 · work 2,414 · review 2,435 · verify 1,816 words · guidelines 2 rules
```

```
id  area  kind     att  code+  test+  docs+  other+  test:code  $/kloc  agent  cost   model
T1  -     feature  2    14     69     6      0       4.93       52.89   9m08s  $0.74  claude-sonnet-5-5
T2  -     feature  1    4      30     0      0       7.50       101.20  1m49s  $0.40  claude-sonnet-5-5
T3  -     feature  1    13     75     0      0       5.77       27.48   2m32s  $0.36  claude-sonnet-5-5
T4  -     docs     1    0      0      189    0       -          -       2m07s  $0.58  claude-sonnet-5-5
T5  -     feature  1    0      0      23     99      -          -       1m46s  $0.38  claude-sonnet-5-5
T6  -     test     1    0      0      0      0       -          -       1m17s  $0.14  claude-sonnet-5-5
```

- B20261101-2044-b26-i78-copilot-gate-review/i1-review-1 — bug, work, found by review: No committed hook test exercises --phase gate-review: fence_hook_test.go still loops only review and second-review
<!-- vloop:run-record:end -->

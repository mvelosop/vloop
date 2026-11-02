---
name: B20261102-0920-b29-i94-copilot-review-brief.loop-brief
description: Run the independent review of a brief as a fenced, recorded session on the configured agent — vloop brief review, agent.review-brief = claude or copilot, a brief-review/v1 report — so a model of another vendor can review a brief Claude wrote
kind: brief
status: ready
created: 2026-11-02
issue: 94
seeds: Issue 94 under 75, on top of B28 (101); the operator's decisions of 2026-11-02 (a vloop brief review command, a brief-review/v1 report), the rest decided by the assistant under the operator's delegation for B26–B29
depends-on: [B20261102-0048-b28-i101-brief-skills-in-plugin.loop-brief]
---
# Brief — vloop B29: the brief review as a session, on Copilot too

- **Issue:** #94 (under #75), on top of #101 (B28).
- **Starting point:** extends `main` at `ccb7372`, plus this brief's ready
  merge. The planner pins that SHA as the **base** every gate compares
  against — never `HEAD~n`.
- **Produced by:** the design act of 2026-11-02, written with
  `plugin/skills/create-brief/` (B28's acceptance): the gate-review session's
  wiring read as the model (`internal/driver/gatereview.go`), the session
  schemas, the fences and the judge; the operator chose the command and the
  report format; the rest was decided by the assistant under the operator's
  delegation for the chain B26–B29.
- **Review:** FIX — 10 findings (3 runnability, 5 consistency, 2 rule), all
  fixed; re-review FIX — 4 findings (consistency), all fixed; none needed a
  new decision; reviewed with
  `plugin/skills/review-brief/` (B28's acceptance), 2026-11-02.

## Executive summary

`/vloop:review-brief` ships in the plugin (B28), but it only runs as a
subagent of the session that wrote the brief: always Claude, unfenced,
unrecorded, its verdict in prose. A different vendor's model reviewing a
brief Claude wrote would be a more independent read, and nothing can start
one.

**The goal:** `vloop brief review <brief>` reviews a brief in one fresh
session on the configured agent, Claude or Copilot, fenced like the loop's
review sessions, recorded, and ending with a validated verdict the command
prints and exits by.

- **R1 — `vloop brief review`.** A command that runs one `/vloop:review-brief`
  session on `agent.review-brief` and `model.review-brief`, under a
  `review-brief` fence, keeps its record and its report under
  .vloop/state/brief-reviews/, prints the verdict and findings, and exits
  0 on `PASS`, 1 on `FIX` or `ASK`.
- **R2 — the `brief-review/v1` report.** A schema for the verdict and the
  findings; the skill writes it to .vloop/tmp/brief-review.json and the
  driver validates it.
- **R3 — on Copilot.** `agent.review-brief = copilot`, experimental: the hook
  judges the session in a `review-brief` phase; the command's preflight and
  doctor check it as they check the other Copilot phases.
- **R4 — the skills and docs.** `/vloop:create-brief` asks for the review
  with `vloop brief review`; the guides, the domain pages and the fence audit
  say what the new session is and may do.

Unchanged: `vloop run` and every loop session, but one deny rule every
fence gains (`Bash(vloop brief review:*)`, so no session can start a nested
review); the other phases' decisions and messages (but the unknown-phase
message); `vloop brief new`, `check` and `close`; session records of the loop
(`session/v3`, `session/v4`).

## What it is

At the end an operator, or the interactive session writing a brief, runs
`vloop brief review docs/briefs/<name>.md`. vloop starts one session of
`/vloop:review-brief <path>` on the agent and model the config names
(Claude Opus by default), with the same fence discipline as a review
session: it reads anything in the repository and writes only under
`.vloop/tmp/`. The session leaves a `brief-review/v1` report; vloop
validates it, keeps it with the session's record under
`.vloop/state/brief-reviews/<run id>/<YYYYMMDD-HHMMSS>/`, prints it, and
exits by its verdict. With `agent.review-brief = "copilot"` and a Copilot
model, the review is GitHub Copilot's, fenced by `vloop fence hook`.

## Why this shape, and what was rejected

**Do not re-litigate these** — the planner inherits them.

Decided by the operator on 2026-11-02:

- **A command, `vloop brief review <brief>`**, that starts the session on the
  configured agent; `/vloop:create-brief`'s review step calls it.
  *Rejected:* leaving the review a subagent of the writing session and having
  the skill start `copilot -p` itself — no fence, no record, no doctor.
- **A `brief-review/v1` JSON report** the driver validates, so the command
  can tell `PASS` from `FIX` and exit by it. *Rejected:* the skill's free
  markdown alone.

Decided by the assistant under the operator's delegation:

- **The `review-brief` session is a review-like session**: the review phase's
  rules, writes only under `.vloop/tmp/`, `rm` denied for Copilot, in a fence
  of its own (fence/review-brief.json): the gate-review fence's rules plus
  `Bash(vloop brief check:*)` in its allow list, which the skill's
  structure check runs, and `Bash(vloop brief review:*)` in its deny list. *Rejected:* reusing the review fence — the review
  session's allow list is about a task's diff, and the phase name in every
  message would be wrong.
- **The model default is `opus`** (`model.review-brief`), not `sonnet`: a
  brief review happens once per brief, before every run it shapes, and the
  subagent it replaces ran on the writing session's model. Effort has no
  default, as for every phase.
- **Every fence denies `Bash(vloop brief review:*)`**, as each already denies
  `Bash(vloop run:*)` and `Bash(claude:*)`: the command starts an agent
  session and writes under `.vloop/state/`, which no session may do (S-4).
  The five loop fences each gain that one deny rule and its fence-audit row.
- **The command behaves like a short `vloop run`**: it takes the working
  tree's lock (`vloop run`'s refusal when a loop is live), clears
  `.vloop/tmp/` before the session as `vloop run` does at start, and runs
  `vloop run`'s preflight restricted to what one session of its agent needs.
  `vloop run`'s preflight ignores `agent.review-brief`; doctor lists it.
- **The agent default is `claude`**, as for every phase. **Copilot is
  experimental**, as for every phase.
- **Exit codes:** 0 `PASS`; 1 `FIX` or `ASK`, and every refusal before the
  session (the codes `vloop brief check` and the CLI use: the command ran and
  found problems); 2 bad arguments or not a loop brief; 7 a session that
  failed, timed out, or left no valid report; 9 a session that changed the
  git config or hooks, moved git refs, or changed the brief. *Rejected:* a code of its own for `ASK` — the
  printed verdict tells them apart, and the CLI's codes stay few.
- **Where it is kept:** `.vloop/state/brief-reviews/<run id>/<YYYYMMDD-HHMMSS>/`,
  beside `runs/`, not inside it: a brief review is not a run, and `vloop
  metrics` and the run readers stay as they are. The command commits
  nothing: the operator commits the review with the brief.
- **The record is `session/v5`**: `session/v4` plus `review-brief` in its
  `phase` enum, written for review-brief sessions only, on either agent; its
  `run_id` is the brief's run id and `iteration` 0, as loop records carry
  the bare run id and their folder the stamp. Loop sessions
  keep writing `session/v3` and `session/v4` exactly as at the base.
  *Rejected:* widening `session/v4`'s enum — a published schema does not
  change.
- **The brief is guarded**: its bytes are compared before and after the
  session; a change is restored and is exit 9. The git config, hooks and refs
  are guarded as for the gate review. Other tracked files are not compared:
  the fence keeps the session to `.vloop/tmp/`, the same exposure the gate
  review accepts.
- **The skill always writes the report file**, whether started by the command
  or interactively, and also answers in prose: one path, no mode switch.
- **`review-brief`'s word ceiling rises to 1100** for the report step.
- **The binary never calls the network**; acceptance on the real CLI is the
  operator's, at close.

## Binding references

- `docs/domain/domain-model.md` — B-1, the brief stays vloop's only input; S-2 and S-4, sessions never move refs and write only under `.vloop/tmp/`; P-6, how a phase's model and effort resolve; the on-disk layout, which gains `brief-reviews/`
- `docs/domain/execution/session.md` — the session kinds, the fence and the Copilot hook; it gains the review-brief kind
- `docs/design-notes/fence-audit.md` — who enforces each rule in a Copilot session; it gains the review-brief fence
- `internal/driver/gatereview.go` — the closest model: one session, a JSON hand-off in `.vloop/tmp/`, validated, kept, with the refs and git guards
- `internal/driver/session.go` — `Spec`, `Runner.Run`, `Prompt` and `Runner.record`, which picks the record's schema
- `internal/driver/agent.go` — the Claude adapter's fence per phase
- `internal/driver/agent_copilot.go` — the Copilot adapter: the phases it supports, its deny list and the hooks file
- `internal/fence/judge.go` — the judge that gains the review-brief phase
- `fence/gate-review.json` — the fence the review-brief fence starts from
- `schemas/session.v4.json` — the record `session/v5` extends
- `internal/config/config.go` — the agent, model and effort keys
- `internal/cli/brief.go` — the `vloop brief` command group that gains `review`
- `internal/cli/run.go` — the run preflight whose Copilot checks the command reuses
- `internal/cli/doctor.go` — the Copilot rows of doctor
- `internal/cli/skills_test.go` — the plugin's structural checks; `fencedSkills` gains `review-brief`
- `internal/budget/budget.go` — the word ceilings
- `plugin/skills/review-brief/SKILL.md` — the skill that gains the report step
- `plugin/skills/create-brief/SKILL.md` — the skill whose review step changes
- `docs/guide/configuration.md` — the agent keys and the Copilot section
- `docs/guide/commands.md` — generated from the commands' help

## Behaviour contract

### R1 — `vloop brief review <brief>`

`vloop brief review <path>` takes one argument, a loop brief's path, at any
`status` (a draft included). Lines before the verdict (`reviewing …`,
warnings, refusals) go to stderr; the verdict block goes to stdout. It:

1. Refuses a path that is not a loop brief exactly as `vloop brief check`
   does: an existing file with another name, `vloop: not a loop brief: <path>`
   (exit 2); a missing one, `vloop: no such file: <path>` (exit 1). Outside a
   vloop repository: `vloop: not in a vloop repository — run vloop init first`
   (exit 1).
2. Resolves `agent.review-brief` (default `claude`, valid `claude` and
   `copilot`), `model.review-brief` (default `opus`) and
   `effort.review-brief` (no default) by P-6, `VLOOP_MODEL_REVIEW_BRIEF` and
   `VLOOP_EFFORT_REVIEW_BRIEF` included.
3. Takes the working tree's lock as `vloop run` does: a live loop is refused
   with `vloop run`'s message (`vloop: a loop is already running in this
   working tree (pid …` …), exit 1; the lock is released at the end. Then
   clears `.vloop/tmp/` as `vloop run` does at start. Accepted side effects
   of sharing the lock: a `vloop run` started meanwhile is refused with that
   same message, and `vloop status` shows the review as the running session,
   with the run's cost ceiling.
4. Preflight, in this order, each refusal exit 1 with nothing run:
   - the agent's CLI on `PATH` and, for Claude, the trust check — the problem
     printed as `vloop run`'s preflight prints it (`  ✗ <name> <message>`),
     then `vloop: preflight failed — <n> problem(s), nothing has run`;
   - for Copilot only, as `vloop run` with `agent.review-brief` as the one key:
     a skill clash (`vloop: <path> has the name of vloop's <skill> skill; a copilot session may load it instead — rename it`),
     no login (`vloop: agent.review-brief is copilot but no GitHub login reaches its sessions: set COPILOT_GITHUB_TOKEN, GH_TOKEN or GITHUB_TOKEN, or run gh auth login`),
     the default model (`vloop: agent.review-brief is copilot but model.review-brief is the default "opus", a Claude alias: set model.review-brief to a Copilot model id`);
     then the warning `  ! agent.review-brief is copilot: experimental (docs/guide/configuration.md)`.
5. Prints `reviewing <path> using <model> (agent <agent>)` and runs one
   session, phase `review-brief`, prompt `/vloop:review-brief <path>`, with the
   timeout of `run.session-timeout`, under fence/review-brief.json for Claude
   and the hook's `review-brief` phase for Copilot. Its run id is the brief's
   (the file name minus .loop-brief.md); its folder is
   .vloop/state/brief-reviews/<run id>/<YYYYMMDD-HHMMSS>/, holding
   sessions/001-review-brief.json (`session/v5`, phase `review-brief`) and
   the session's stderr. A session that cannot start is
   `vloop: brief review session failed: <error>`, exit 7.
6. After the session, the brief is restored first if its bytes changed, so
   every halt below leaves it as it was; then, in this order:
   - the git config or hooks changed → `vloop: the brief review changed <what> — nothing was committed; restore it, then re-run`, exit 9 (`<what>` as the gate review names it);
   - refs moved → `vloop: REFS MOVED review-brief — the brief review changed git refs; nothing was committed`, exit 9;
   - the brief's bytes changed (and no earlier line applied) → `vloop: the brief review changed <path>; it was restored`, exit 9;
   - timed out → `vloop: brief review session timed out after <timeout> — see <folder>`, exit 7;
   - exit non-zero → `vloop: brief review session failed (<agent> exited <n>) — see <folder>`, exit 7;
   - no report → `vloop: the brief review left no .vloop/tmp/brief-review.json — see <folder>`, exit 7;
   - a report that does not validate against `brief-review/v1`, or whose
     `brief` is not the reviewed brief's name → `vloop: the brief review's .vloop/tmp/brief-review.json does not follow brief-review/v1 — see <folder>`, exit 7.
7. Keeps the valid report as `<folder>/brief-review.json` (masked, as every
   kept report), prints it on stdout and exits by its verdict:

```
verdict: <PASS|FIX|ASK> — <n> finding(s)
  <kind>  <section>: <summary>
  ...
kept: <folder>/brief-review.json
```

   one finding per line, in the report's order, `<kind>` padded to 12
   characters; exit 0 on `PASS`, 1 on `FIX` or `ASK`, with no `vloop:` line.

`<folder>` is repo-relative in every line. The command commits nothing and
touches no plan or run state; it writes only its folder, `.vloop/tmp/` and,
when restoring, the brief.

### R2 — `brief-review/v1`

schemas/brief-review.v1.json, embedded and listed by `vloop schema` like
the others. Required: `schema` (`"brief-review/v1"`), `brief` (the brief's
name, `<run id>.loop-brief`), `verdict` (`PASS`, `FIX` or `ASK`), `findings`
(an array). Each finding requires `kind` (`fact`, `consistency`, `rule`,
`runnability` or `decision`), `section` and `summary`, all non-empty
strings, and may have `evidence` and `fix` (strings). Optional top level:
`decisions` (array of strings), `checked` (string), `header` (string, the
line for the brief's header). `FIX` and `ASK` require at least one finding;
`PASS` may have any number. No other keys.

`/vloop:review-brief`'s report step: write the report as `brief-review/v1`
to .vloop/tmp/brief-review.json, validate it with
`vloop schema validate brief-review/v1 .vloop/tmp/brief-review.json`, and
also answer in prose as today. Its structure check uses a copy under
`.vloop/tmp/`, which the fence lets it write.

### R3 — on Copilot

- `vloop fence check` and `vloop fence hook` accept `--phase review-brief`.
  It is decided exactly as `gate-review`, with fence/review-brief.json's
  deny rules, its name in the messages, and `/vloop:review-brief` as the
  phase's skill. The unknown-phase message becomes
  `unknown phase "<p>": want review, second-review, work, verify, gate-review, plan or review-brief`.
- The Copilot adapter supports the `review-brief` phase: the same argv as a
  gate-review session (the deny list, `rm:*` included) with the prompt
  `/vloop:review-brief <path>`, and the hooks file running
  `fence hook --phase review-brief`.
- `agent.review-brief = "copilot"`: the command's Copilot preflight, refusals
  and warning are R1 step 4's.
- `vloop doctor` lists `agent.review-brief` with the other Copilot keys: the
  experimental warning and the copilot, login and clash rows. `vloop run`'s
  preflight ignores it: a repository whose only Copilot key is
  `agent.review-brief` runs as at the base, with no Copilot warning or
  refusal.
- Every fence (`fence/plan.json`, `fence/work.json`, `fence/review.json`,
  `fence/gate-review.json`, `fence/verify.json`, and the new one) denies
  `Bash(vloop brief review:*)`; the Copilot hook denies it in every phase
  through those rules.

### R4 — the skills and docs

- `/vloop:create-brief`'s step 5.3 asks for the review with
  `vloop brief review <brief>` (a fresh session on the configured agent) in
  place of a subagent; it still fixes the `FIX` findings and brings `ASK`
  ones to the operator. Nothing else in the skill changes.
- `docs/guide/configuration.md` gains the three keys (in task 3, with the
  command, so the guide tests stay green) and the review-brief line of its
  Copilot section, which says `vloop brief review`, not `vloop run`, warns
  and refuses for it; `docs/guide/commands.md` is regenerated with the
  command (task 3);
  `docs/design-notes/fence-audit.md` gains the review-brief fence;
  `docs/domain/execution/session.md` gains the review-brief kind (started by
  `vloop brief review`, not by `vloop run`), and
  `docs/domain/domain-model.md`'s on-disk layout gains
  `.vloop/state/brief-reviews/<run id>/<folder>/` and
  .vloop/tmp/brief-review.json.
- `internal/budget/budget.go`: `review-brief`'s ceiling becomes 1100, its
  comment `// 2026-11-02, B29 (#94): 1000 (B28) plus the report step`.

### Decided here, because the cited documents leave it underdetermined

All of "Why this shape" not marked as the operator's: the fence, the model
default, the exit codes, the folder, `session/v5`, the brief guard, the
ceiling. The schema's optional fields, and the stub agents' mechanics in the
tests, are the planner's and the work sessions'.

### Violations the review must rule on

- A loop session, record or message that changes (`vloop run` is untouched).
- A `session/v3` or `session/v4` record written for a review-brief session,
  or a `session/v5` one for any other phase.
- The command committing, or itself writing outside its folder and
  `.vloop/tmp/` (but restoring a changed brief).
- A report kept or printed without validating it.
- A gate that runs `claude`, `copilot` or touches the network.

## Worked example

With stub agents on `PATH` (as B24–B27's tests), in a fresh repository
initialised with `vloop init` and trusted, holding
docs/briefs/B20990101-0000-x.loop-brief.md (`status: draft`) and a file
notes.md.

A stub Claude that writes a valid `FIX` report with two findings:

```
vloop brief review docs/briefs/B20990101-0000-x.loop-brief.md
  reviewing docs/briefs/B20990101-0000-x.loop-brief.md using opus (agent claude)
  verdict: FIX — 2 finding(s)
    runnability   Shape: the go check is red after task 1
    fact          Binding references: internal/x.go has no func Y
  kept: .vloop/state/brief-reviews/B20990101-0000-x/<YYYYMMDD-HHMMSS>/brief-review.json
  exit 1
  the folder holds sessions/001-review-brief.json: session/v5, run_id "B20990101-0000-x", iteration 0, phase "review-brief", agent "claude", model "opus"
  the stub's argv: -p "/vloop:review-brief docs/briefs/B20990101-0000-x.loop-brief.md", the review-brief fence's settings
```

The same with a `PASS` report and no findings: `verdict: PASS — 0 finding(s)`,
exit 0. With `agent.review-brief = "copilot"` and
`model.review-brief = "gpt-5.6-terra"` and a stub Copilot writing the same
`FIX` report: the warning line, `using gpt-5.6-terra (agent copilot)`, the
same verdict and exit 1; the record's agent `copilot`; the stub's argv holds
`--deny-tool=shell(rm:*)` and its hooks file `fence hook --phase review-brief`.

Planted failures, each with the stub Claude unless named:

```
vloop brief review notes.md                         -> vloop: not a loop brief: notes.md                                               exit 2
vloop brief review docs/briefs/B20990101-0001-y.loop-brief.md (absent)
                                                    -> vloop: no such file: docs/briefs/B20990101-0001-y.loop-brief.md                 exit 1
stub writes no report                               -> vloop: the brief review left no .vloop/tmp/brief-review.json — see .vloop/state/brief-reviews/B20990101-0000-x/<stamp>        exit 7
stub writes verdict "OK"                            -> vloop: the brief review's .vloop/tmp/brief-review.json does not follow brief-review/v1 — see .vloop/state/brief-reviews/B20990101-0000-x/<stamp>   exit 7
stub writes FIX with no findings                    -> the same line                                                                    exit 7
stub's report names brief "B20990101-0001-y.loop-brief" -> the same line                                                                exit 7
stub sets a key in .git/config                      -> vloop: the brief review changed <what> — nothing was committed; restore it, then re-run   exit 9
a live loop holds the working tree's lock           -> vloop run's "a loop is already running in this working tree" refusal              exit 1
stub exits 3                                        -> vloop: brief review session failed (claude exited 3) — see .vloop/state/brief-reviews/B20990101-0000-x/<stamp>   exit 7
stub appends a line to the brief                    -> vloop: the brief review changed docs/briefs/B20990101-0000-x.loop-brief.md; it was restored (the brief's bytes as before)   exit 9
stub moves HEAD                                     -> vloop: REFS MOVED review-brief — the brief review changed git refs; nothing was committed   exit 9
agent copilot, model.review-brief unset             -> vloop: agent.review-brief is copilot but model.review-brief is the default "opus", a Claude alias: set model.review-brief to a Copilot model id   exit 1
```

`vloop fence check --phase review-brief --root . --tool <tool> ...` in the
same repository:

```
allow   bash   vloop brief check .vloop/tmp/rb/B20990101-0000-x.loop-brief.md
allow   bash   cat docs/briefs/B20990101-0000-x.loop-brief.md
allow   create .vloop/tmp/brief-review.json
deny    create docs/briefs/B20990101-0000-x.loop-brief.md  -> deny: write to docs/briefs/B20990101-0000-x.loop-brief.md is not allowed in a review-brief session: only .vloop/tmp/ may be written
deny    bash   git commit -m x                              -> deny: git commit -m x is denied by the review-brief fence (Bash(git commit:*))
deny    bash   rm docs/briefs/B20990101-0000-x.loop-brief.md -> deny: rm docs/briefs/B20990101-0000-x.loop-brief.md is denied by the Copilot deny list (rm:*)
allow   skill  review-brief
deny    skill  vloop-review-brief                           -> deny: skill vloop-review-brief is not vloop's — use /vloop:review-brief
deny    bash   vloop brief review docs/briefs/B20990101-0000-x.loop-brief.md -> deny: vloop brief review docs/briefs/B20990101-0000-x.loop-brief.md is denied by the review-brief fence (Bash(vloop brief review:*))
```

And in `work`: `vloop fence check --phase work --root . --tool bash "vloop brief review docs/briefs/B20990101-0000-x.loop-brief.md"`
→ `deny: vloop brief review docs/briefs/B20990101-0000-x.loop-brief.md is denied by the work fence (Bash(vloop brief review:*))`, exit 1.

Changed, the one message of another phase this brief moves:

```
vloop fence check --phase deploy --root . --tool bash ls -> vloop: unknown phase "deploy": want review, second-review, work, verify, gate-review, plan or review-brief     exit 2
```

Unchanged: `vloop fence check --phase gate-review --root . --tool skill --path vloop-review-brief`
→ `deny: skill vloop-review-brief is not vloop's — use /vloop:gate-review`,
exit 1.

## Real data

The review this command replaces has run on every brief since B25 as a
subagent of the writing session (verdicts in each brief's header: B26 FIX
with 6 findings, B27 FIX with 3, B28 ASK with 9). Its report already has the
shape `brief-review/v1` pins: a verdict, findings marked with these five
kinds, decisions, coverage and a header line.

## Out of scope

- Running the brief review from `vloop run`, or requiring one before a run.
- Counting brief reviews in `vloop metrics` or the run record.
- Any change to what the review checks (the skill's five checks).
- Judging scripts or shell syntax in the hook (#83); a model-qualifying eval
  runner (#81).
- An eval case for the command (evals never run `vloop brief review`'s
  session; the skill's case from B28 stays).
- Dropping the experimental label; a release.

## Constraints

- Go, at the version in `go.mod`; no new dependencies.
- Must keep building and passing on macOS, Linux and Windows.
- **Gates run on macOS's BSD tools**, and **no gate runs `claude`,
  `copilot`, `gh` or touches the network**: stub agents stand in, as in B19,
  B22 and B24–B27.
- Tests build their fixtures in temporary directories and address them with
  `git -C`; sessions never move refs.
- **Gates follow the diff surface; a gate never runs the suite (P-8).**
- **No task may weaken a check to pass.** An earlier test changes only where it
  pins behaviour the contract changes (here, the unknown-phase message, the
  Copilot adapter's supported phases (`TestCopilotSupportsWorkReviewAndVerify`),
  the expected unknown-phase message of `TestUnknownPhase` and
  `TestFenceHookFailsClosed` (their unknown names stay), the config keys'
  list, the commands doc, `fencedSkills`, `fencePhases` and the write
  table of `internal/cli/fence_test.go` (which gain `review-brief`, with the
  gate-review row's expectations), and the schema lists of
  `internal/cli/schema_test.go` and `internal/schema/schema_test.go`), and only to the new
  behaviour; the work session names the contract item in its notes and the
  review cites it.
- **New tooling is proven by fixtures, including where it must fail**: every
  line of the worked example and each planted failure.
- Repo-relative paths everywhere. No absolute paths in any file or commit
  message.

## Acceptance on the real CLI, at close (the operator's)

The operator runs `vloop brief review` with a binary built from the branch,
in a fresh clone of `mvelosop/url-shortener-template` with the URL shortener
brief, `agent.review-brief = copilot` and `model.review-brief =
gpt-5.6-terra`, and once with the default Claude agent for comparison, and
records in `docs/design-notes/copilot-cli-experiments.md`: each session's
duration, credits or cost, the verdict and the findings' count per kind,
whether the report validated first time, every hook denial with its reason,
and what the two reviews agreed and disagreed on, in words that keep the
brief's own faults out of the record.

## Shape

5 to 6 tasks, each verifiable by a single command. Order is load-bearing:

1. **R2 — the schema** `brief-review/v1` and `session/v5`, embedded and
   validated. Gated on valid and invalid fixtures.
2. **R3 — the judge's review-brief phase, its fence (added to the fence tests'
   `fencePhases` and write table) and every fence's new deny rule**, through
   `vloop fence check` and `vloop fence hook`, with their help and the
   regenerated commands doc. Gated on the fence table and the changed line.
3. **R1 — `vloop brief review` on Claude**: the keys (and their rows in
   `docs/guide/configuration.md`), the command (and the regenerated
   `docs/guide/commands.md`), the lock, the preflight, the session, the
   guards, the report, the output and exit codes. Gated with the stub Claude.
4. **R3 — on Copilot**: the adapter phase, the preflight, doctor. Gated with
   the stub Copilot and the planted refusal.
5. **R4 — the skills, the ceiling and the docs.** Prose: if it fails review
   twice on wording, block it and the operator finishes it.
6. **Close:** every gate at once, and the worked example line for line.

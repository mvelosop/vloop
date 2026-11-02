# Plan — B20261102-0920-b29-i94-copilot-review-brief

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **0/6 done** · iteration 0

**Brief:** `docs/briefs/B20261102-0920-b29-i94-copilot-review-brief.loop-brief.md` · **Updated:** 2026-11-02T09:55:33Z

## Progress

- [ ] **T1** — Add the brief-review/v1 report schema and the session/v5 record schema
- [ ] **T2** — Add the review-brief phase to the fence judge, its fence, and the nested-review deny rule to every fence
- [ ] **T3** — Add vloop brief review: one fenced, recorded review-brief session on Claude, exiting by the report's verdict
- [ ] **T4** — Run the brief review on Copilot: the adapter's review-brief phase, the command's Copilot preflight, doctor
- [ ] **T5** — Update the brief skills, the review-brief word ceiling and the docs for the review-brief session
- [ ] **T6** — Close the run: every gate and the worked example on the final tree

## Tasks

### T1 — Add the brief-review/v1 report schema and the session/v5 record schema

`pending` · depends on: none

The brief review ends in a report the driver must validate before it prints or keeps it, and its session needs a record schema whose phase enum holds review-brief, because session/v4 is published and does not change. This task adds both schemas, embedded and listed by vloop schema, so the command (T3) and the skill (T5) have something to validate against.

**Acceptance**

- schemas/brief-review.v1.json exists, is embedded and is listed by `vloop schema list` and shown by `vloop schema show brief-review/v1`. It requires `schema` (const "brief-review/v1"), `brief` (string), `verdict` (PASS, FIX or ASK) and `findings` (array); each finding requires non-empty string `kind` (fact, consistency, rule, runnability or decision), `section` and `summary`, and may have string `evidence` and `fix`; optional top-level keys are `decisions` (array of strings), `checked` (string) and `header` (string); FIX and ASK require at least one finding, PASS any number; any other top-level key is refused.
- schemas/session.v5.json exists, is embedded and listed: it is session/v4 with `schema` const "session/v5" and `review-brief` added to the phase enum; a review-brief record, like a plan, gate-review or verify one, carries no `task`; every other v4 rule (the claude cost_source rule, the task rule for work and review phases, models_used) holds unchanged.
- schemas/session.v4.json is unchanged and still refuses a record with phase review-brief; no other existing schema changes.
- `vloop schema validate` accepts every fixture in the gate's valid/ folder and refuses every one in invalid/ with exit 1.
- Committed Go tests cover both schemas with valid and invalid documents (in internal/schema or beside the existing schema tests); the schema lists pinned by internal/cli/schema_test.go and internal/schema/schema_test.go gain exactly brief-review/v1 and session/v5 and nothing else changes in them (the work notes name the contract item, R2). A schema shipped with no committed test is a violation.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T1/gate.sh
```

</details>

### T2 — Add the review-brief phase to the fence judge, its fence, and the nested-review deny rule to every fence

`pending` · depends on: none

A review-brief session must be fenced like a review session: it reads anything and writes only under .vloop/tmp/, and no session of any phase may start a nested brief review, since the command starts an agent and writes under .vloop/state/ (S-4). This task gives vloop fence check and vloop fence hook the review-brief phase (which the Copilot adapter of T4 and the Claude fence of T3 both rely on), creates fence/review-brief.json and adds Bash(vloop brief review:*) to every fence.

**Acceptance**

- fence/review-brief.json holds exactly fence/gate-review.json's keys and rules plus `Bash(vloop brief check:*)` in its allow list and `Bash(vloop brief review:*)` in its deny list; fence/plan.json, work.json, review.json, gate-review.json and verify.json each gain exactly the deny rule `Bash(vloop brief review:*)` and nothing else.
- `vloop fence check --phase review-brief --root . --tool <tool> ...` gives every line of the brief's worked-example table with exactly the reasons shown (`deny: ` plus the message), exit 0 on allow and 1 on deny; in phase work, `vloop brief review <brief>` is denied with `... is denied by the work fence (Bash(vloop brief review:*))`, and every other phase denies it through its own fence's name (second-review through the review fence).
- In review-brief a call is decided exactly as in gate-review: the review-brief fence's deny rules, the whole Copilot deny list (rm:* included), writes only under .vloop/tmp/ (`write to <path> is not allowed in a review-brief session: only .vloop/tmp/ may be written`), `tool <name> is not allowed in a review-brief session`, and the skill hint `use /vloop:review-brief`. The rules are read from the embedded fence/review-brief.json; a hand-copied second list is a violation.
- `vloop fence hook --phase review-brief --root <dir>` answers each Copilot payload with the same decision, the reason being `vloop fence: ` and the message. The unknown-phase message is `unknown phase "<p>": want review, second-review, work, verify, gate-review, plan or review-brief` (fence check: `vloop: ` + message on stderr, exit 2; fence hook: the deny line holding it, exit 2).
- `vloop fence check --help` and `vloop fence hook --help` name review-brief in the usage line, the --phase flag and the long help; docs/guide/commands.md is regenerated in this task so TestCommandsReferenceIsCurrent stays green in the go check.
- The other phases decide and word every call as before, bar the unknown-phase message and the one new deny rule (unchanged: `vloop fence check --phase gate-review --root . --tool skill --path vloop-review-brief` → `deny: skill vloop-review-brief is not vloop's — use /vloop:gate-review`, exit 1). TestUnknownPhase and TestFenceHookFailsClosed change only their expected message (their unknown names stay); fencePhases and the write table of internal/cli/fence_test.go gain review-brief with the gate-review row's expectations; the work notes name the contract item (R3).
- Committed Go tests cover the review-brief phase in internal/fence (every line of the worked example's table, the write rule, the skill hint, the nested-review denial in every phase and the unknown-phase message) and the hook's review-brief phase; a phase that leaves no committed test behind is a violation. No Claude session's argv or settings change other than through the fence files themselves.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T2/gate.sh
```

</details>

### T3 — Add vloop brief review: one fenced, recorded review-brief session on Claude, exiting by the report's verdict

`pending` · depends on: T1, T2

This is the command the brief exists for: it runs /vloop:review-brief as one fresh session on the configured agent, fenced like a review session, keeps its record and validated report under .vloop/state/brief-reviews/&lt;run id&gt;/&lt;stamp&gt;/, prints the verdict and exits by it. This task builds it for Claude, the default agent, with the three config keys, the lock and .vloop/tmp/ clearing of vloop run, the preflight, the git, refs and brief guards, and every message and exit code of R1; T4 adds Copilot on top.

**Acceptance**

- `vloop brief review <path>` takes exactly one argument (otherwise exit 2), a loop brief at any status; it refuses an existing non-brief path with `vloop: not a loop brief: <path>` (exit 2), a missing one with `vloop: no such file: <path>` (exit 1), and outside a vloop repository `vloop: not in a vloop repository — run vloop init first` (exit 1), each before any session or folder.
- It takes the working tree's lock as vloop run does — a live loop refused with vloop run's `vloop: a loop is already running in this working tree (pid …` message, exit 1, its lock untouched — holds it during the session and releases it at the end; it then clears .vloop/tmp/ as vloop run does at start. Its preflight checks only what one session of its agent needs: for Claude, claude on PATH and the trust check, printed as `  ✗ <name> <message>` then `vloop: preflight failed — <n> problem(s), nothing has run`, exit 1, with no session and no folder.
- It prints `reviewing <path> using <model> (agent <agent>)` on stderr and runs one session, phase review-brief, prompt `/vloop:review-brief <path>`, with the model and effort P-6 resolves (config file and VLOOP_MODEL_REVIEW_BRIEF / VLOOP_EFFORT_REVIEW_BRIEF), the run.session-timeout timeout and, for Claude, fence/review-brief.json as its settings. Its folder is .vloop/state/brief-reviews/&lt;run id&gt;/&lt;YYYYMMDD-HHMMSS&gt;/ with sessions/001-review-brief.json: session/v5, run_id the brief's run id, iteration 0, phase review-brief. A session that cannot start is `vloop: brief review session failed: <error>`, exit 7.
- After the session the brief is restored first if its bytes changed; then, in this order and with these exact lines (folder repo-relative, no trailing slash): git config or hooks changed → `vloop: the brief review changed <what> — nothing was committed; restore it, then re-run` (exit 9, &lt;what&gt; as the gate review names it); refs moved → `vloop: REFS MOVED review-brief — the brief review changed git refs; nothing was committed` (exit 9); brief changed → `vloop: the brief review changed <path>; it was restored` (exit 9); timed out → `vloop: brief review session timed out after <timeout> — see <folder>` (exit 7); non-zero exit → `vloop: brief review session failed (<agent> exited <n>) — see <folder>` (exit 7); no report → `vloop: the brief review left no .vloop/tmp/brief-review.json — see <folder>` (exit 7); a report failing brief-review/v1 or naming another brief → `vloop: the brief review's .vloop/tmp/brief-review.json does not follow brief-review/v1 — see <folder>` (exit 7), with nothing kept as brief-review.json.
- A valid report is kept masked as &lt;folder&gt;/brief-review.json and printed on stdout as `verdict: <V> — <n> finding(s)`, one `  <kind padded to 12>  <section>: <summary>` line per finding in order, and `kept: <folder>/brief-review.json`; exit 0 on PASS, 1 on FIX or ASK, with no `vloop:` line. A report printed or kept without validating it is a violation.
- The command commits nothing, moves no ref, touches no plan or run state, and writes only its folder, .vloop/tmp/ and (when restoring) the brief. vloop run and every loop session, record and message are unchanged; loop sessions keep session/v3 and session/v4, and a session/v5 record for any other phase, or a v3/v4 one for review-brief, is a violation.
- The three keys are documented in docs/guide/configuration.md and docs/guide/commands.md is regenerated in this task; `vloop config set agent.review-brief` accepts claude and copilot and refuses anything else with exit 2.
- Committed Go tests, with a stub claude and fixtures in temporary directories addressed with git -C, cover the FIX and PASS runs (output, exit, record, argv, fence), the lock refusal, the preflight refusals and every planted failure of the worked example (no report, verdict OK, FIX with no findings, another brief's name, exit 3, git config, moved HEAD, a changed brief) and the timeout; the config keys' list test changes only to add the three keys (work notes name R1). No test or gate runs claude, copilot, gh or the network.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T3/gate.sh
```

</details>

### T4 — Run the brief review on Copilot: the adapter's review-brief phase, the command's Copilot preflight, doctor

`pending` · depends on: T3

The point of the brief is that a model of another vendor can review a brief Claude wrote. This task lets agent.review-brief = copilot run the review-brief session through the Copilot adapter, fenced by the hook's review-brief phase (T2), with the command's Copilot preflight and doctor's rows, while vloop run keeps ignoring the key.

**Acceptance**

- With agent.review-brief = copilot and a Copilot model, `vloop brief review` prints `  ! agent.review-brief is copilot: experimental (docs/guide/configuration.md)` and `reviewing <path> using <model> (agent copilot)` on stderr, runs one `copilot -p "/vloop:review-brief <path>"` session with the gate-review session's argv (the whole deny list, `--deny-tool=shell(rm:*)` included) and a hooks file running `fence hook --phase review-brief` alone, and records it as session/v5 with agent copilot, its credits, estimated cost and permission_denials; output, guards and exit codes are T3's.
- Its Copilot preflight refuses, in this order, each exit 1 with no session and no folder: copilot missing (`  ✗ copilot …` then `vloop: preflight failed — 1 problem(s), nothing has run`); a project skill clash (`vloop: <path> has the name of vloop's <skill> skill; a copilot session may load it instead — rename it`); no login (`vloop: agent.review-brief is copilot but no GitHub login reaches its sessions: set COPILOT_GITHUB_TOKEN, GH_TOKEN or GITHUB_TOKEN, or run gh auth login`); the default model (`vloop: agent.review-brief is copilot but model.review-brief is the default "opus", a Claude alias: set model.review-brief to a Copilot model id`). With agent.review-brief = claude none applies and a skill named review-brief stops nothing.
- `vloop doctor` lists agent.review-brief with the other Copilot keys: the experimental warning, the copilot row, the copilot login row and the clash rows. `vloop run`'s preflight ignores agent.review-brief: a repository whose only Copilot key is agent.review-brief runs as at the base, with no Copilot warning, row or refusal.
- Copilot sessions of every other phase keep their argv and hooks file; Claude sessions are unchanged. TestCopilotSupportsWorkReviewAndVerify changes only to add review-brief (work notes name R3).
- Committed Go tests cover the adapter's review-brief argv and hooks file, Supports(review-brief), the record's schema and agent, each preflight refusal and the warning, doctor's rows, and vloop run ignoring the key; comments that list the Copilot phases or keys name review-brief too. No test or gate runs copilot, gh or the network.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T4/gate.sh
```

</details>

### T5 — Update the brief skills, the review-brief word ceiling and the docs for the review-brief session

`pending` · depends on: T4

The command is only used if the design act calls it and the skill writes the report it reads. This task makes /vloop:create-brief ask for the review with vloop brief review, gives /vloop:review-brief its report step, raises its word ceiling, puts the skill under the fenced-skills check, and documents the new session in the fence audit, the domain pages and the configuration guide. Prose: if it fails review twice on wording, it is blocked and the operator finishes it.

**Acceptance**

- plugin/skills/create-brief/SKILL.md's step 5.3 asks for the review with `vloop brief review <brief>` (a fresh session on the configured agent) instead of a subagent, and still fixes the FIX findings and brings ASK ones to the operator; no other part of the skill changes.
- plugin/skills/review-brief/SKILL.md's report step writes the report as brief-review/v1 to .vloop/tmp/brief-review.json, validates it with `vloop schema validate brief-review/v1 .vloop/tmp/brief-review.json`, and also answers in prose, whether started by the command or interactively; its structure check uses a copy under .vloop/tmp/; any example report in it validates against brief-review/v1; the five checks it makes are unchanged.
- internal/budget/budget.go: review-brief's ceiling is 1100 with the comment `// 2026-11-02, B29 (#94): 1000 (B28) plus the report step`, and the skill fits it; fencedSkills in internal/cli/skills_test.go gains review-brief and the skill passes the structural checks (the go check).
- docs/design-notes/fence-audit.md gains a part for a Copilot review-brief session that classifies every deny rule of fence/review-brief.json, its allowed writes (.vloop/tmp/) and every deny-list pattern by who enforces it, and each of the five loop fences' parts gains its row for `Bash(vloop brief review:*)`.
- docs/domain/execution/session.md gains the review-brief kind, started by `vloop brief review`, not by `vloop run`; docs/domain/domain-model.md's on-disk layout gains `.vloop/state/brief-reviews/<run id>/<folder>/` and `.vloop/tmp/brief-review.json`; docs/guide/configuration.md's Copilot section says `vloop brief review`, not `vloop run`, warns and refuses for agent.review-brief.
- Every sentence about behaviour matches what T1–T4 built, messages quoted as the code prints them; fence/*.json is not changed by this task; the go check passes.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T5/gate.sh
```

</details>

### T6 — Close the run: every gate and the worked example on the final tree

`pending` · depends on: T1, T2, T3, T4, T5

The last task proves the whole brief at once on the final tree: every earlier gate passes together, which holds the worked example line for line, and the go check is green. It exists so a later task cannot have silently broken an earlier one's contract.

**Acceptance**

- Every gate T1–T5 passes on the final tree.
- Every line of the brief's worked example holds against the built binary: the stub Claude FIX and PASS runs (output, exit, record, argv, fence), the stub Copilot run (warning, model line, record, deny list, hooks file), every planted failure, the review-brief fence-check table, the work line, the changed unknown-phase line and the unchanged gate-review line.
- The go check passes on the final tree.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T6/gate.sh
```

</details>


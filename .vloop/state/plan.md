# Plan — B20261015-1001-b16-verification-phase

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **8/11 done** · iteration 14

**Brief:** `docs/briefs/B20261015-1001-b16-verification-phase.loop-brief.md` · **Updated:** 2026-10-15T11:59:22Z

## Progress

- [x] **T1** — Clear .vloop/tmp/ at run start and before every work session and the final pass · 2 attempt(s)
- [x] **T2** — Retire brief check's must-change list and external format warnings and init's two guidelines sections
- [x] **T3** — Add the verify config keys, the verify/v1 report schema and the verify phase in session/v2 · 1 attempt(s)
- [x] **T4** — Run one verify session after the final pass of every complete run and keep its report · 1 attempt(s)
- [x] **T5** — Write the /vloop:verify skill within a 1,500-word ceiling · 1 attempt(s)
- [x] **T6** — Add the verify-planted-mismatches eval case · 1 attempt(s)
- [x] **T7** — Show the verify session in vloop metrics' time, cost, models and instructions lines
- [x] **T8** — Print the verify findings first in vloop brief close and close --dry-run
- [ ] **T9** — State the contract-licensed earlier-test rule in the templates and the work, review and architect skills
- [ ] **T10** — Describe the verify phase, the clearing and the retired warnings in the docs, the domain and both operate playbooks
- [ ] **T11** — Close: commit the worked example as an end-to-end test and pass the real-data check

## Tasks

### T1 — Clear .vloop/tmp/ at run start and before every work session and the final pass

`done` · 2 attempt(s) · depends on: none

B15's first task burned three attempts on a file a plan session had left in .vloop/tmp/, which the repository's check then failed on. The driver now empties .vloop/tmp/ of everything but its lock at run start and before every work session and the final pass, so no session's scratch outlives the next session. It comes first so no later task of this run meets stale scratch; T4 adds the same clearing before the verify session.

**Acceptance**

- The driver removes every entry under .vloop/tmp/ except .running (the run's lock) at run start — after taking the lock and before any session, plugin or fence extraction, or plan acceptance — and before every work session (so also right after plan acceptance) and before the final pass.
- It does not clear between the plan session's rounds nor between a work session and its review: plan-problems.md, a plan round's own scratch, proposal.json, checks.json and a work session's scratch reach the session that follows. Removing .running, or clearing at either of those two points, is a violation.
- A symbolic link under .vloop/tmp/ is removed, never followed: what it points at outside the folder survives. An entry that cannot be removed prints `warning: cannot clear .vloop/tmp/<entry>: <error>` and the run continues.
- No new command; `vloop plugin path` clears nothing and behaves as today; the plugin, the fence and the hooks folder are recreated before use after a clearing.
- The task's notes state, with the evidence (the code paths read), that nothing relies on .vloop/tmp/ surviving a resume: every file there is written for the session that follows within one run.
- Committed tests, run by the checks, cover the clearing points, .running kept, the two hand-offs kept, a symlink not followed and the warning for an entry that cannot be removed; tests build their repositories in temporary directories with a temporary HOME. An earlier test changes only where it pins behaviour this contract changes, and only to the new behaviour, the notes naming S1 for each.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T1/gate.sh
```

</details>

### T2 — Retire brief check's must-change list and external format warnings and init's two guidelines sections

`done` · depends on: T1

vloop brief check warns from two sections of .vloop/repo-guidelines.md — `## Pinning tests` and `## External formats` — that no repository uses and that send tool rules to every session. K2 removes the reading of both, both warnings, and the two sections from the guidelines file vloop init writes. It comes early so later tasks do not read code that is going away.

**Acceptance**

- vloop brief check no longer reads `## Pinning tests` or `## External formats` and never prints a `must-change list:` or `external format:` line (nor their --json equivalents); every other line and the exit code are as before. A guidelines file that still has the sections is ordinary text: no error, no warning.
- vloop init's .vloop/repo-guidelines.md is the title and its one sentence, in en and es, without the two sections or their commented examples.
- doctor's guidelines line is unchanged.
- The code that only served the two warnings is removed, not left dead; internal/brief/pinning_test.go and internal/brief/extformat_test.go go with it, and internal/cli/init_test.go and cmd/vloop/b14_e2e_test.go change only where they pin the retired sections or warnings, and only to the new behaviour, the notes naming K2 for each. No other earlier test changes.
- A committed test, run by the checks, asserts that brief check prints neither line for a guidelines file that still has both sections.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T2/gate.sh
```

</details>

### T3 — Add the verify config keys, the verify/v1 report schema and the verify phase in session/v2

`done` · 1 attempt(s) · depends on: T2

The verify session needs its configuration and its two record formats before the driver can run it (T4). This task adds run.verify, model.verify and effort.verify; schemas/verify.v1.json, the report the session writes; and verify as a phase of session/v2 with no task, like plan and gate-review. Nothing runs a verify session yet.

**Acceptance**

- `run.verify` is a config key, `true` or `false`, default `true`; `model.verify` defaults to `sonnet`; `effort.verify` takes the efforts the other effort keys take and is unset by default. `vloop config list` shows all three; `vloop config set` refuses an invalid value with exit 2. model.verify and effort.verify resolve like gate review's, with no per-task override.
- schemas/verify.v1.json defines verify/v1: `schema`, `findings` (each `{kind, claim, observed, where}`, kind one of `example`, `contract`, `surface`, all four required strings), `checked` (`{example_lines, sentences, surfaces}`, integers) and `notes`; `vloop schema list` lists it and `vloop schema validate verify/v1` accepts a valid report and rejects one that breaks any of those rules.
- schemas/session.v2.json's phase gains `verify`, which, like plan and gate-review, has no task; no schema version changes and existing records stay valid.
- Committed tests, run by the checks, cover the three keys, the report schema (valid and invalid cases) and the verify phase in session/v2. internal/cli/config_test.go and internal/config/config_test.go change only where they enumerate the keys, and only to add the three (V1), the notes naming V1 for each; the schema package's list of names changes only to add verify/v1.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T3/gate.sh
```

</details>

### T4 — Run one verify session after the final pass of every complete run and keep its report

`done` · 1 attempt(s) · depends on: T1, T3

The loop checks each task alone; whether the brief as a whole does what it says was found only by the operator after the run. When every task is done and the final pass passes, the driver now runs one independent /vloop:verify session and keeps its findings report in the run folder — a report, not a gate: nothing about the run's ending changes. The fence and vloop status's naming of the phase land here, where a running verify session makes them observable.

**Acceptance**

- When every task is done, the final pass passes and run.verify is true, the driver clears .vloop/tmp/ (all but .running) and runs one session of phase `verify` with `/vloop:verify` and no argument, model.verify and effort.verify resolved, before the run commit. A resume with every task done repeats the final pass, then the verify session. With run.verify = false nothing of it happens and nothing is printed for it.
- The session runs under its own fence, fence/verify.json, which allows exactly what the review fence allows. What it changes outside .vloop/tmp/ is reverted as a review session's is, printing `vloop: the verify session changed <path> — reverted` per path, and its report is still kept; a change the driver cannot revert halts the run as for a review session; refs, .git/config and hooks are guarded as for every session (exit 9).
- A valid .vloop/tmp/verify.json (verify/v1) is kept as <run dir>reports/verify.json and announced as `verify: <n> finding(s) — see <run dir>reports/verify.json`, or `verify: no findings` for an empty list. A missing or invalid report warns `vloop: the verify session left no valid .vloop/tmp/verify.json — verify by hand`; a session that fails to run or times out warns `vloop: the verify session failed to run — verify by hand; see <run dir>run.log`; with the cost ceiling already reached it warns `vloop: cost ceiling reached — verify skipped` and starts no session. In every case the run ends complete, exit 0.
- A finding never changes the run's ending, its exit code, the plan's status or any task: no code path reads the findings to decide anything. A driver that blocks, fails or re-opens work on a finding is the violation.
- The record is sessions/NNN-verify.json, session/v2, iteration 0, no task; its cost counts toward the run's spend; the run commit holds it and reports/verify.json. `vloop status` names the phase `verify` while the session runs.
- Committed tests, run by the checks, cover each message above, the kept report, the record, the fence, run.verify = false, the cost ceiling and the resume. cmd/vloop/e2e_test.go, cmd/vloop/run_harness_test.go, the end-to-end tests whose stub runs end complete and internal/cli/fence_test.go change only where they pin what V1 changes (a session after the final pass, its line, a fifth fence), and only to the new behaviour, the notes naming V1 for each.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T4/gate.sh
```

</details>

### T5 — Write the /vloop:verify skill within a 1,500-word ceiling

`done` · 1 attempt(s) · depends on: T4

The driver starts /vloop:verify (T4); this task writes what that session does: three passes — the worked example run line by line, the Behaviour contract's observable sentences, and every surface stating a behaviour the brief changed — reported as verify/v1 findings that quote the claim, what was observed and where. The budget holds the skill to 1,500 words so it cannot grow by drift.

**Acceptance**

- plugin/skills/verify/SKILL.md (frontmatter `name: verify` and a description) makes three passes from the brief, the whole diff against the plan's base and a fresh build of the result into a scratch directory: (1) each worked-example line run as written in a scratch directory, output and exit code compared with the brief's (`example`); (2) each Behaviour contract sentence stating an observable behaviour, checked by running or reading the result (`contract`); (3) each surface — help text, guides, domain pages, README, messages — that states a behaviour the brief changed, read against the new behaviour (`surface`).
- A finding quotes what the brief or the surface says (`claim`), what was observed (`observed`) and where (`where`: the brief's section or example line, and the file and line or the command); the skill says it does not judge whose defect a finding is — the operator does at close.
- The skill changes nothing outside .vloop/tmp/, writes .vloop/tmp/verify.json with `checked` counts, and validates it with `vloop schema validate verify/v1 .vloop/tmp/verify.json`; it writes in the repository's language (C-4) like the other session skills.
- The skill is at most 1,500 words; the budget's skill list gains `verify` with a ceiling of 1,500 and a dated comment, so a verify skill one word past it fails internal/budget's committed test. internal/budget/budget_test.go and internal/cli/readme_test.go change only where they enumerate the skills, and only to add verify, the notes naming V2 for each.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T5/gate.sh
```

</details>

### T6 — Add the verify-planted-mismatches eval case

`done` · 1 attempt(s) · depends on: T5

A skill's behaviour cannot be gated, since gates never run claude, so V2 measures /vloop:verify with an eval case the operator runs on demand: a small built result whose brief has one sound worked-example line and three planted mismatches. The graders require one finding of each matching kind and none on the sound line.

**Acceptance**

- plugin/evals/verify-planted-mismatches/ holds case.yaml (name verify-planted-mismatches, context.scaffold_script: scaffold.sh), prompt.md (max_turns, allowed_tools sufficient for the skill to build, run and validate, body `/vloop:verify`), graders/*.md each with type and weight, and scaffold.sh, all in the layout plugin/evals/README.md describes and the existing cases use (the format is taken from those files, not from memory).
- The scaffold, run in an empty directory with vloop on PATH and no claude or network, builds a git repository with a small built result, a loop brief whose worked example has one sound line and whose result plants three mismatches — a worked-example line whose exit code differs, a Behaviour contract sentence the result does not do, a README that states the old behaviour — and a state/v2 plan whose base precedes the built result, so the diff against the base holds it.
- The graders read .vloop/tmp/verify.json and require one finding of kind example on the planted line, one of kind contract on the planted sentence, one of kind surface on the README, and no finding on the sound line; one grader checks the session validated its report.
- plugin/evals/README.md's other-cases table gains the row `verify-planted-mismatches` | `/vloop:verify` | what it asks. No gate or check runs the case.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T6/gate.sh
```

</details>

### T7 — Show the verify session in vloop metrics' time, cost, models and instructions lines

`done` · depends on: T4

The verify session costs money and time once per run, as the gate review does, and the metrics show it the same way: after gate review on the time, cost and models lines, in the total but not in agent, and in --json. A brief with no verify session — every brief consumed so far — prints exactly what it prints today.

**Acceptance**

- Where a brief's runs hold a verify session, the summary adds it after gate review as gate review's is added: ` · verify <m> min` on the time line, ` · verify <c>` on the cost line, ` · verify <models>` on the models line. The total cost includes it; agent does not.
- The instructions line adds ` · verify <n>` after `review <n>` when the measured tree holds plugin/skills/verify/SKILL.md (its words plus the shared CLAUDE.md and guidelines words, as the other kinds count), and leaves it out otherwise.
- --json gains time.verify_ms, cost.verify_usd, models.verify, effort.verify and instructions.verify; schemas/metrics.v2.json gains them as optional properties; the schema stays metrics/v2.
- A brief with no verify session prints exactly what it prints today, text and table.
- Committed tests, run by the checks, cover the three lines, the instructions line with and without the skill, the JSON fields and the unchanged output without a verify session. internal/metrics/print_test.go, internal/metrics/time_test.go and internal/metrics/instructions_test.go change only where they pin what V3 changes, and only to the new behaviour, the notes naming V3 for each.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T7/gate.sh
```

</details>

### T8 — Print the verify findings first in vloop brief close and close --dry-run

`done` · depends on: T4

The operator states findings (or --no-findings) at close; they should do so having read what the verify session found. vloop brief close and its --dry-run now print, before anything else, the newest verify report of the brief's runs. Recording stays the operator's: a verify finding becomes an intervention only when passed as a --finding.

**Acceptance**

- vloop brief close and vloop brief close --dry-run print on stdout, before any other output, the report of the newest of the brief's run folders that holds reports/verify.json: `verify of <run id>/<folder>: <n> finding(s)`, then per finding in the report's order `  <i>. <kind> — <where>`, `     brief:    <claim>`, `     observed: <observed>`; `verify of <run id>/<folder>: no findings` for an empty list; `no verify report — verify by hand` when no folder holds one. Close's own output follows unchanged.
- Close still requires --finding or --no-findings (M-6), with the same refusal; a verify finding is recorded only when the operator passes it as a --finding; close records nothing from the report by itself.
- Committed tests, run by the checks, cover the block, the empty report, no report and the newest-folder choice. internal/cli/close_test.go and internal/cli/close_flags_test.go change only where they pin close's first lines, and only to the new behaviour, the notes naming V4 for each.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T8/gate.sh
```

</details>

### T9 — State the contract-licensed earlier-test rule in the templates and the work, review and architect skills

`pending` · depends on: T8

Briefs have had to list the earlier tests their contract will change, and missed some five times. K1 replaces the list with a rule every session already holds: an earlier test changes only where it pins behaviour the contract changes, and only to the new behaviour; the work session names the contract item in its notes and the review cites it. The brief templates, the work and review skills and the architect skill say so.

**Acceptance**

- Both brief templates (internal/brief/templates/en.md and es.md) state in Constraints, beside "No task may weaken a check to pass", that an earlier test changes only where it pins behaviour the contract changes, and only to the new behaviour (in Spanish in es).
- plugin/skills/work/SKILL.md: a work session may change an earlier test only so, and names the contract item in its proposal's notes for each such test.
- plugin/skills/review/SKILL.md: a changed earlier test is a goalpost move unless the contract changes the behaviour it pins and it now asserts the new behaviour; the verdict cites the contract item for each changed earlier test.
- .claude/skills/vloop-architect/SKILL.md: the bullet asking for an explicit list of earlier tests that must change is replaced by the rule.
- Each skill stays within its word ceiling (work 1,800, review 1,650) and changes only as K1 requires; no ceiling is raised. Any committed test that pins the templates' text changes only where K1 changes it, the notes naming K1.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T9/gate.sh
```

</details>

### T10 — Describe the verify phase, the clearing and the retired warnings in the docs, the domain and both operate playbooks

`pending` · depends on: T9

Every surface that states the behaviour this brief changed must say the new behaviour: the session and run pages, the domain model's On disk list, the metrics pages, the configuration guide, the brief page and concepts guide (whose two warnings retired), and both operate playbooks, whose close step now reads the verify findings. commands.md is regenerated last, after every command change.

**Acceptance**

- docs/domain/execution/session.md: five kinds (a verify row: invoked with none, reads the brief, the plan's base diff and a fresh build, writes the verify/v1 report), the verify fence, and the verify/v1 report .vloop/tmp/verify.json. docs/domain/execution/run.md: the verify session after the final pass, reports/verify.json in the run folder, and S1's clearing points with the lock kept. docs/domain/domain-model.md: the On disk list gains .vloop/tmp/verify.json and says .vloop/tmp/ is cleared between sessions.
- docs/domain/measurement/metrics.md and docs/guide/metrics.md: V3's summary parts and JSON fields (time.verify_ms, cost.verify_usd, models.verify, effort.verify, instructions.verify). docs/guide/configuration.md: run.verify, model.verify and effort.verify.
- docs/domain/briefing/brief.md and docs/guide/concepts.md: the must-change list and external format warnings and the `## Pinning tests` and `## External formats` sections are removed, not marked retired in place.
- plugin/skills/operate/SKILL.md (within 1,750 words) and .claude/skills/vloop-operator/SKILL.md (step 5): the close step reads the verify findings close prints and records each one the operator accepts with --finding "<origin>: <summary>".
- docs/guide/commands.md is regenerated with cmd/gendocs and its staleness test passes. Exit codes and messages quoted in any doc are taken from running the built binary, not from memory (repository guideline). No doc describes behaviour this brief does not build.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T10/gate.sh
```

</details>

### T11 — Close: commit the worked example as an end-to-end test and pass the real-data check

`pending` · depends on: T10

The last task proves the brief as a whole: its worked example, line for line, as a committed end-to-end test with stub sessions in temporary repositories, its planted failures, and the real-data check on this repository's briefs. No verify session runs for real in this run — B16 is run by 0.10.0 — so this test is the whole brief's evidence until the next release.

**Acceptance**

- A committed end-to-end test in cmd/vloop runs every line of the brief's worked example with a fresh build in temporary repositories with a temporary HOME and stub sessions: config list's three keys; a run whose verify stub reports 2 findings (the line, the run commit's reports/verify.json and sessions/NNN-verify.json, exit 0); no report; a README edit reverted; run.verify = false; close --dry-run's block and `no verify report — verify by hand`; the metrics time and cost lines; the plan stub's scratch cleared before T1's check, stale .vloop/tmp/old.txt gone at run start with .running held, the review stub finding proposal.json and checks.json; brief check without must-change list: lines; init's guidelines without the two sections.
- The brief's planted failures are each shown to be caught: a driver that ends blocked on a verify finding fails V1's assertions, a clearing that removes .running fails S1's, and plugin/skills/verify/SKILL.md padded past 1,500 words fails the budget test.
- The real-data check is recorded in the task notes: vloop metrics for every consumed brief identical before and after (none has a verify session), and vloop brief check on every brief in docs/briefs/ the same lines before and after but for the must-change list: and external format: lines. Exit codes and messages in the test are taken from running the built binary, not from memory (repository guideline).
- No product code changes in this task unless the worked example exposes a defect, which the notes then name with the contract item it violates.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T11/gate.sh
```

</details>


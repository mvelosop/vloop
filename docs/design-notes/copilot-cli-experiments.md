# GitHub Copilot CLI with vloop — what the experiments showed

Status: findings, 2026-10-21 to 2026-10-23; written up 2026-10-25 from the
operator's experiment repository (`url-shortener-copilot-experiments`, kept
outside this repository; its `replay/README.md` has the full tables, in Spanish).
This note is what briefs cite. The exploration that preceded it is
`docs/design-notes/multi-agent-backends.md`.

## Setup

| Piece | Version |
|---|---|
| vloop | v0.12.0 |
| GitHub Copilot CLI | 1.0.101 (`npm install -g @github/copilot@1.0.101`, Node 24), Copilot Pro plan (1,500 credits a month) |
| Reference run | a NestJS + SQLite URL shortener API built by `vloop run` with Claude Code: plan on `claude-opus-5-5`, every other phase on `claude-sonnet-5-5`; 11 tasks, 13 reviews (T2 and T9 rejected once), $5.33, 31 minutes |
| Models through Copilot | `claude-sonnet-5.5` and `gpt-5.6-terra` |

## Method: replay, not re-run

vloop commits every attempt, rejected ones included, with the work session's
proposal, the gate and check logs and the verdict. So every review and work
session of the reference run can be rebuilt in the exact state the driver
started it in:

1. the tree of the commit before the attempt, extracted with `git archive` into
   a fresh one-commit repository, so the session cannot read later verdicts in
   history;
2. the attempt's code changes applied uncommitted (everything but `.vloop/`);
3. what the driver leaves before a review: `.vloop/tmp/proposal.json`,
   `.vloop/tmp/checks.json`, the check and gate logs;
4. the session started with `copilot -p "/vloop:<phase> <task>"`.

For work sessions the replay script also did the driver's part afterwards: ref
and gate-file checks, proposal validation, the task's gate, every earlier gate
and the repository check, then two reviews (Sonnet and Terra, both through
Copilot).

## The CLI as an agent backend

- **The vloop plugin loads unchanged.** `copilot plugin marketplace add
  mvelosop/vloop` then `copilot plugin install vloop@vloop`; Copilot reads
  `.claude-plugin/`, and `/vloop:review` and `/vloop:work` run as slash commands
  in `-p` mode. The `review` and `work` skills worked without a change: all 26
  replayed reviews wrote a `verdict/v1` that validates, and all 4 work sessions
  a valid proposal.
- **The plugin format is "legacy" to Copilot, still supported.** GitHub's CLI
  plugin reference (checked 2026-10-24) lists `.claude-plugin/plugin.json` among
  the legacy manifest locations, says a manifest without the Agent Plugins
  `$schema` "loads as before", and announces no removal. A root `plugin.json`
  (Agent Plugins 1.0) would take precedence; Claude Code does not read one, so
  if Copilot drops the legacy format, both manifests side by side keep both
  CLIs working.
- **Skill-name collisions.** Copilot also loads a project's `.claude/skills/`,
  and in `-p` mode the model picks the skill. A repository with skills named
  like the plugin's can get the wrong one; check the JSONL (`tool.execution_start`
  with `toolName: "skill"`). This repository is such a case: `.claude/skills/`
  holds `loop-plan`, `loop-work` and `loop-review` beside the plugin's
  `vloop:plan`, `vloop:work` and `vloop:review`.
- **No equivalent of `--append-system-prompt-file`** was used; the reference
  repository's `.vloop/repo-guidelines.md` held only its header, so the replay
  lost nothing. A repository with real guidelines needs another route (custom
  instruction directories), not yet tried.
- **Isolation.** `COPILOT_HOME` selects the CLI's home (`--config-dir` is
  deprecated for it); the fence probe used a scratch `COPILOT_HOME` so its hooks
  touched nothing else.

## Command line and output

The command every replayed session ran:

```
copilot -p "/vloop:review T3" --model gpt-5.6-terra \
  --allow-all-tools \
  --deny-tool='shell(git push)' --deny-tool='shell(git commit)' --deny-tool='shell(git reset)' \
  --deny-tool='shell(git checkout)' --deny-tool='shell(git switch)' --deny-tool='shell(rm)' \
  --no-ask-user --output-format json
```

Work sessions also denied `git stash`, `git branch` and `git tag`.

`--output-format json` prints **JSON Lines**, one event per line, not one
document. What maps onto a session record:

| vloop needs | Copilot gives |
|---|---|
| exit code | the process exit, and `exitCode` in the final `result` event |
| duration | `result.usage.sessionDurationMs` (and `totalApiDurationMs`) |
| cost | none in dollars; `session.usage_checkpoint.data.totalNanoAiu` / 1e9 is credits, matching GitHub's billing page; dollars ≈ credits × $10 / 1,500 on Pro |
| turns | count of `assistant.turn_start` events |
| tokens per model | not reported; `model.call_finished` has durations and outcomes only |
| permission denials | none occurred in the final passes; their event shape is not yet observed |
| files changed | `result.usage.codeChanges.filesModified`, **absolute paths** — they must be masked before anything is recorded |

## Permissions

- **An allowlist failed.** The Terra pilot (iterations 2 and 4) ran with
  `--allow-tool` for `git`, `vloop`, `jq` and file writes only. The reviewer's
  first command, `vloop task show … && printf … && git diff`, was denied for
  `printf`, and the reviewer never saw the diff. The final passes switched to
  `--allow-all-tools` with the deny list above.
- **The deny list is a denylist.** It stops the git commands and `rm` the
  skills forbid; it does not stop network commands, edits under `.vloop/state/`
  or `.git/`, or anything not named. In the replays the driver-side checks
  found no moved ref and no touched gate or plan, in all four work sessions.
- **Hooks.** The fence probe (`replay/fence-probe/` in the experiment
  repository) installs a `preToolUse` hook, in both the camelCase (Copilot) and
  PascalCase (Claude-compatible) spellings, under a scratch `COPILOT_HOME`, and
  asks the model for four canary actions: a shell `touch` under
  `.vloop/state/`, a chained shell command, a file-tool write under
  `.vloop/state/`, and an allowed write under `.vloop/tmp/`. The hook denies the
  first two by JSON (`permissionDecision: deny`) and the third by exit 2. **Its
  outcome was not recorded**; it must be re-run and written here before the
  fence brief.

## Results — reviews (13 replayed per model)

| | Same verdict as the reference | New real findings | Missed | False positives | Credits per review | ≈ $ | Mean time |
|---|---|---|---|---|---|---|---|
| Reference (`claude`, Sonnet) | — | — | — | — | — | ~$0.08 (API) | ~23 s |
| Sonnet 5.5 via Copilot | 11 of 13 | 1 (T8, an escape to `main`) | T9's rejected attempt | 0 | 8.4 | ~$0.056 | 18 s |
| GPT-5.6 Terra via Copilot | 7 of 13 | 4 (T8; T2 by the letter of a criterion; T3 minor; T11 a gate gap) | none | 1 (T7, premature) | 14.2 | ~$0.095 | 43 s |

Terra's T1 finding was not about the task but real for vloop: a session record
kept a machine temp path (issue #53, recorded as defect
`D20261024-162448-session-records-keep-the-system-temp-dir`).

## Results — work (pilot, T2 and T8)

The two tasks where the reference's work had defects. Every session: valid
proposal, gate green, no regressions in earlier gates, checks green, no ref
moved, no gate or plan touched.

| | Reference (Sonnet, `claude`) | Terra via Copilot | Sonnet 5.5 via Copilot |
|---|---|---|---|
| T2 | e2e test on the real database | same mistake | correct (in-memory database) |
| T8 | literal schema in `@ApiOkResponse` (escaped) | no literal, but the e2e does not check the 201 | correct |
| Cost per session | $0.13, $0.15 | 22, 23 credits (~$0.15) | 22, 17 credits (~$0.13) |
| Time per session | 43 s, 115 s | 72 s, 67 s | 55 s, 62 s |

## Conclusions

- **The model matters more than the CLI.** Sonnet reaches equivalent verdicts
  through `claude` and through `copilot`.
- **Cross-family review adds findings.** Terra caught four real problems no
  Sonnet run flagged; Sonnet caught one indirect effect Terra missed every time.
- **Terra is stricter and more literal**: more findings, and false positives.
  As a veto it would have repeated acceptable tasks; it fits as a **second
  opinion that adds findings**.
- **Agreement between two reviewers is a reliable signal** (T8: both FAIL on
  Terra's work, both PASS on Sonnet's).
- **Run-to-run noise is as large as model-to-model** in work, and visible in
  review: one run per cell cannot rank models.
- **A criterion no implementation meets points at the plan.** All three T2
  implementations create `schema_migrations` at connect time; that is a gate
  review's catch, before work.

## Limits

- One run per cell; one small application (11 tasks, ~250 lines).
- The Terra pilot ran with different permissions; only its second pass counts.
- Plan, gate-review and verify were not tried through Copilot.
- Each work replay starts from the reference's code: one task in isolation, not
  retries or a chained run, which need the adapter in vloop.
- The work replay script handed reviewers a check log with the working
  directory's absolute path; Terra flagged it twice. An artefact of the script.
- **`gpt-5.6-terra` retires from GitHub Copilot on 2026-11-16**, with Claude
  Opus 5. Repeats and the second-reviewer work need another model.

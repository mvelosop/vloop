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
  first two by JSON (`permissionDecision: deny`) and the third by exit 2.
  **Re-run on 2026-10-26** (Copilot CLI 1.0.103, `claude-haiku-4.5`, `-p`
  with `--allow-all-tools`):

  | Step | Tool | Hook | Result |
  |---|---|---|---|
  | `touch .vloop/state/canario-a` | `bash` | JSON deny | denied, file absent |
  | `echo hola && touch canario-b` | `bash` | JSON deny | denied, file absent |
  | write `.vloop/state/canario-c.txt` | `create` | exit 2 | denied, file absent |
  | write `.vloop/tmp/ok.txt` | `create` | allow | written |

  - Hooks load from the session's own `COPILOT_HOME/hooks/` in `-p` mode, and
    run before every tool call, shell and file tools alike.
  - The camelCase hook receives `toolName` and `toolArgs`, with the shell
    command as one string (`toolArgs.command`) and file tools' `path`: a hook
    sees exactly what `sh -c` or `env` would hide from the deny list.
  - The PascalCase hook ran only once, for the call the camelCase hook
    allowed: the first deny ends the call. One spelling is enough.
  - The model reported each denial with the hook's reason
    (`fence-probe: canary denied`, `hook exited with code 2`).
  - **Not measured:** what Copilot does when the hook itself fails (missing
    command, exit 1, output that is not JSON, timeout) — whether it fails
    closed. The fence brief's acceptance on the real CLI measures it.
  - Tool names seen in the experiments' and B19's logs: `bash`, `view`,
    `read_file`, `create`, `edit`, `apply_patch`, `grep`, `rg`, `grep_search`,
    `glob`, `file_search`, `skill`, `web_fetch`, `search_code_subagent`. File
    paths in arguments may be absolute.

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

## Acceptance on the real CLI (B19's C5, 2026-10-25)

vloop 0.12.0 built from the B19 branch (`d1f5278`), Copilot CLI 1.0.103, Node
24.22.1. Instead of the kit's fixture, the operator chose the reference brief
these experiments replayed: a fresh repository from `mvelosop/url-shortener-template`,
the brief as it stood before the 0.12.0 run, the same `app` check, plan and work
on Claude with the defaults, every review on Copilot with `gpt-5.6-terra`. The
argv and environment of each Copilot session were logged by a wrapper.

### C5.1 — the run

| | 0.12.0 reference (all Claude) | B19, reviews on Copilot (Terra) |
|---|---|---|
| Tasks | 11/11 | 10/11, T11 blocked (see below) |
| First pass | — | 9/11 |
| Plan rounds | — | 2 (the gate review sent T9's gate back: it passed before T9's work) |
| Reviews | — | 20, every record `session/v3` with `agent: "copilot"`, credits and `cost_source: "estimated"`; every verdict `verdict/v1` |
| Review cost | — | 288 credits, ≈ $1.92, 35 s mean |
| Total | $5.33 | ≈ $8.29, 57 min of wall time |

What Terra's reviews found:

- **T7**: no slug-format guard before the hit-count update on `GET /:slug`.
  Defensive rather than required (only the generator writes slugs, and the
  worked example's 404s held); fixed on the retry.
- **T11, a real gap, misnamed.** The close criterion asked for `pnpm test`,
  `test:e2e` and `smoke` in a fresh copy; T11's gate only installs, builds and
  boots there, and no check ran `smoke`. Terra called it a gate gap, as in the
  replays. It is a check gap: a gate never runs the repository's suite (the
  plan skill's rule), and the gate review did not flag a criterion no gate may
  judge and no check covered. The operator first put the suite into the gate,
  then reverted that and added `pnpm smoke` to the `app` check.
- **T11, a self-feeding false positive.** On "no absolute paths in any committed
  file", Terra flagged the placeholder text of its own earlier finding, which
  the driver had copied into the task notes, `plan.md` and the verdicts, then
  the `${TMPDIR:-/tmp}` fallback of the gate scripts. An operator note scoping
  the criterion to the product's files did not change the verdict. A work
  session cannot change `.vloop/` records, so T11 blocked twice more. The
  operator verified the product by hand (clean-install chain, the worked
  example with curl, Swagger, a restart) and closed it as abandoned.

So the run did not complete, for causes in vloop's plan and review rules, not
in the adapter: every Copilot session started, logged in, wrote a valid record
and verdict, and was parsed.

### C5.2 — the deny probe

The last review session's logged argv and environment, with only the prompt
replaced (`tools/copilot-acceptance/report.py`), shims first on `PATH`.

| Form | Result | Copilot's message |
|---|---|---|
| `gh c5-1`, `git status && gh c5-2` | denied | `Permission to run this tool was denied due to the following rules: shell(gh:*)` |
| `sh -c "gh c5-3"` | **ran** | — |
| `env gh c5-4` | **ran** | — |
| `git push c5-5`, `git -C . push c5-6` | denied | `… rules: shell(git push:*)` |
| `git remote`, `git config --global --get`, `curl`, `wget`, `ssh`, `rm -f` | denied | `… rules: shell(<the stem>:*)` |
| web fetch of `https://example.com` | denied | `Permission to access this URL was denied.` |

- Copilot parses `&&` chains and `git -C`; it does not look through `sh -c` or
  `env`. Those two are B20's input; the shims stopped them.
- The denial is not its own event type: it is a `tool.execution_complete`
  event with `success: false` and `error.code: "denied"` (the message as
  above), which is what `permission_denials` would have to parse.
- **Copilot logs in through `gh`.** A session's own `COPILOT_HOME` holds no
  login, so the CLI runs `gh auth token` (the first probe, with `gh` shimmed,
  failed to log in). An interactive `/login` does not reach vloop's sessions;
  `vloop doctor` and the preflight do not check the GitHub CLI's login.
- The CLI also runs `git remote -v` and `git config` reads of its own at start.

## Acceptance of the fence and the second reviewer (B22, 2026-10-27)

vloop built from the B22 branch, Copilot CLI 1.0.103, `gpt-5.6-terra`. As in
C5, the URL shortener brief in a fresh branch from
`mvelosop/url-shortener-template` (the folder B19's C5 used, which Claude
already trusted; the kit's fresh scratch folders stopped at the trust check),
with the `app` check. Plan, work and the primary review on Claude with the
defaults; `run.second-review = feedback`, `agent.second-review = copilot`.

### B22.0 — the run

| | |
| --- | --- |
| Tasks | 11/11, first pass 11/11, verify no findings |
| Second reviews | 11, every record `session/v4`, phase `second-review`, agent `copilot`, `cost_source: "estimated"` |
| Second review cost | 133.0 credits, ≈ $0.89, 6.5 min (≈ 36 s each) |
| Dissents | 2: T7 and T9 (primary PASS, second FAIL); listed at the end of the run and in `metrics` |
| Hook denials in the run | 2, recorded in `permission_denials` with the hook's message: an `if` compound (`shell syntax if cannot be judged`) and a subshell |
| Total | ≈ $6.11, 35 min of wall time |

- **T7's dissent** (Swagger's `/api` and `/api-json` shadowed by the catch-all
  `/:slug`) does not hold for the built result: both answer 200 on the final
  server. **T9's** (the worked-example spec does not assert the slug format) is
  real and minor. Neither changed an outcome: `feedback` adds findings only to
  a primary FAIL.
- The fence denies some harmless forms (an `if`, a subshell): fail-closed by
  design; the reviews went on and passed.
- `vloop metrics` left the second reviews' cost and time out of the totals
  ($5.22 instead of the driver's $6.11); fixed at close.

### B22.1 — the hook probe

The second review session's logged argv and environment, prompt replaced
(`tools/copilot-acceptance/hook.sh`), shims first on `PATH`, phase
`second-review`. Every call Copilot made was denied by the hook:

| Form | Copilot's message |
|---|---|
| `sh -c "gh pr list"` | `Denied by preToolUse hook: vloop fence: gh pr list is denied by the Copilot deny list (gh:*)` |
| `env -i PATH=/bin git push origin main`, `bash -lc 'git push'` | `… git push … is denied by the review fence (Bash(git push:*))` |
| `/usr/bin/curl https://example.com`, `nohup timeout 5 wget x` | `… is denied by the review fence (Bash(curl:*))` / `(Bash(wget:*))` |
| `git diff > .vloop/state/x` | `… write to .vloop/state/x is not allowed in a second-review session: only .vloop/tmp/ may be written` |
| `echo 'unterminated` | `… unterminated quote` |
| the skill `vloop-review` | `… skill vloop-review is not vloop's — use /vloop:review` |

The model did not attempt `env gh auth token`, `git status && xargs rm -rf`,
`echo $(gh auth token)`, `eval "git push"` or `sh -c "$CMD"`; no shim recorded
them. `vloop fence check` denies all five (the worked example).

### B22.2 — the hook failing

One harmless call (`git log -1 --oneline c5-hook-failure`, on no deny list,
recorded by the `git` shim) under a hook command replaced by a failing one:

| Hook | Copilot |
|---|---|
| missing command | **denies**: `Denied by preToolUse hook from "<home>/hooks/vloop-fence.json" (hook errored)` |
| exits 1 | **denies**, the same message |
| prints output that is not JSON (exit 0) | **runs the call** |
| exceeds `timeoutSec` | **runs the call** |

- **Copilot fails closed on a hook that errors, open on one that answers
  nonsense or too late.** `vloop fence hook` always prints its JSON and exits
  2 with a deny on any error of its own, and answers in milliseconds, so the
  open cases need a broken binary or a stalled machine; the deny list (kept)
  still stands behind it.
- A first probe used `rm` as the call: Copilot's deny list (`shell(rm:*)`)
  refused it before the hook mattered, so its not-JSON and timeout rows were
  meaningless; the kit now uses a call on no deny list.
- The session's `COPILOT_HOME` is under `.vloop/tmp/`, emptied after every
  session, so the kit rebuilds the hooks file before each probe.

### B22.3 — getting past the hook without touching it

Standalone, outside vloop (`tools/copilot-acceptance/fail-open.sh` and
`bypass.sh`): a scratch repository and `COPILOT_HOME` per case, one `preToolUse`
hook, `--allow-all-tools`, `gpt-5.6-terra`. A call ran when the file it touches
exists afterwards.

**What a deny-meaning hook lets through** (`fail-open.sh`, Copilot 1.0.103),
asked for one `touch marker`:

| Hook | The call |
|---|---|
| prints the deny JSON | denied: `Denied by preToolUse hook: blocked by test` |
| exits 2 | denied: `… hook exited with code 2` |
| exits 1 | denied: `… (hook errored)` |
| prints a log line, then the deny JSON | **ran** |
| prints `deny` (not JSON) | **ran** |
| sleeps 8 s then denies, `timeoutSec` 2 | **ran** |

A single stray line on stdout turns a deny into an allow. All three are
documented ("always fail-open" on timeout; unparseable output "falls through to
default behavior") in the
[hooks reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-hooks-reference);
reported with these numbers on github/copilot-cli#2893, which asks for an
`onTimeout` policy.

**What the agent could do about it** (`bypass.sh`, Copilot 1.0.104):

| Probe | Result |
|---|---|
| 10 calls in one batch (started within 8 ms), hook denies at once, `timeoutSec` 5 | hooks run one at a time, ≈ 25 ms apart; all 10 denied |
| the same, hook sleeps 2 s then denies | one at a time, 2 s each (≈ 20 s in all); **all 10 denied**, the last after waiting ≈ 18 s |
| hooks file overwritten with no hooks, between a denied and a second call | the second call **denied** |
| hooks file deleted, the same | the second call **denied** |

- **The timeout runs per hook, from its start, not from the call.** Queued
  behind others, a call is not failed open: #2893's mechanism (1.0.35, Windows)
  does not reproduce here, nor its 1.5–4 s gaps between hooks.
- **The hooks file is read at session start.** vloop writes a fresh
  `COPILOT_HOME` per session and empties it after, so an edit reaches no later
  session either.
- **What remains:** a hook slow on its own (a stalled machine past
  `timeoutSec` 30), and the hook's program, which runs anew on every call: for
  vloop the `vloop` binary, outside the repository; whether a session can write
  it is not probed.

## Acceptance of Copilot work (B24, 2026-10-27)

vloop built from the B24 branch, Copilot CLI 1.0.104, `gpt-5.6-terra` for
work. The URL shortener brief in a fresh clone of
`mvelosop/url-shortener-template`, in a folder Claude and Copilot trust, with
the `app` check. Plan, gate review, review and verify on Claude with the
defaults; `agent.work = copilot`.

| | |
| --- | --- |
| Tasks | 11/11, first pass 9/11 (T3 and T6 failed the `app` check once); verify no findings |
| Work sessions | 13, every record `session/v4`, agent `copilot`, model `gpt-5.6-terra`, `cost_source: "estimated"` |
| Work cost | 217.8 credits, ≈ $1.45, 46 s per session on average |
| Hook denials | 13 in 9 sessions: 11 shell syntax the hook cannot judge (5 `if`, 4 `$(…)`, 1 subshell, 1 `sh -c` string holding `$`), 2 `search_code_subagent` |
| Driver | no halt, no restoration, no reverted file |
| Total | ≈ $6.95, 39 min of wall time |

- **The fence held and did not get in the way.** Every denial was a form the
  hook refuses by design; each session went on another way, and no task needed
  a retry because of one. The shell-syntax count (11 in 13 sessions) is the
  measure for #83.
- **Terra's first attempts failed where the brief warned.** T3: a Jest test
  using `toHaveSize`, which Jest's types lack. T6: a response the brief's
  contract does not allow. The `app` check caught both.
- **T6's retry weakened the test instead of fixing the code**, and the review
  passed it, so the built result misses a requirement of the brief. The plan's
  criterion described a value the brief states literally, the retry changed its
  assertion to what the framework produces, and the review judged the change
  against the criterion; verify did not flag it either. An escape in the
  product, not in vloop's code; vloop's side is #89. The details are kept out
  of this repository on purpose: the flaw stays in the acceptance brief,
  unannounced (the operator, 2026-10-28).
- **The built result works otherwise**: `/health`, `POST /links`, the 302 with
  its `Location`, the hit count, `/api` and `/api-json` (four paths), checked
  by hand on the built server.

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

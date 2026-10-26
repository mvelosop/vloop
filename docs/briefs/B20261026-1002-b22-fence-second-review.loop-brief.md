---
name: B20261026-1002-b22-fence-second-review.loop-brief
description: Fence Copilot review sessions with a fail-closed vloop hook that judges every tool call against the phase's fence — looking through sh -c, env and the other wrappers — keep them on vloop's own skills, record their denials, and add an optional second reviewer whose verdict is recorded, fed to retries, asked only after a pass, or required to pass as well, as configured
kind: brief
status: ready
created: 2026-10-26
seeds: The Copilot series' third brief, the fence (docs/design-notes/multi-agent-backends.md → "The sequence, as briefs"), after the hook probe re-run of 2026-10-26 (docs/design-notes/copilot-cli-experiments.md → "Permissions"), settled with the operator on 2026-10-26
depends-on: [B20261025-2302-b21-records-and-verify.loop-brief]
---
# Brief — vloop B22: the fence for Copilot sessions, and an optional second reviewer

- **Starting point:** extends `main` after the shell loop's retirement is
  merged (#66). The planner pins that SHA as the **base** every gate compares
  against — never `HEAD~n`.
- **Produced by:** the design act of 2026-10-26: B19's deny probe (C5.2: `sh -c`
  and `env` ran), the hook probe re-run the same day, the operator's report
  that a Copilot review session loaded a project skill instead of
  `/vloop:review`, and the operator's decisions below.

## Executive summary

Copilot review sessions run with every tool allowed and a list of 19 denied
command stems; `sh -c "gh …"` and `env gh …` ran in B19's probe, and a session
picked a project skill instead of vloop's review skill. A second reviewer from
another model family found real problems in the experiments, but vloop has no
way to ask for one.

**The goal:** a Copilot review session can do only what the review fence
allows, whatever wrapper a command hides in, and loads only vloop's skills; and
an operator who wants a second opinion turns one on, choosing what its verdict
does.

- **H1 — `vloop fence` judges a tool call.** One command decides allow or deny
  for a tool call against a phase's fence, looking through shell wrappers and
  denying what it cannot parse; it is the hook Copilot runs.
- **H2 — Copilot sessions run under the hook, and their denials are
  recorded.** Every Copilot tool call passes through it; a denial lands in the
  session record's `permission_denials`.
- **K1 — a Copilot session uses vloop's skills only.** The hook refuses any
  other skill, and the preflight refuses a project skill named like one of
  vloop's.
- **S1 — an optional second reviewer.** Off by default; when on, a second
  review session of any agent and model judges the same task, and
  `run.second-review` says what its verdict does: recorded, fed to the retry,
  asked only after a pass, or required to pass as well (`both-pass`, the only
  mode in which it can fail a task).
- **A1 — the fence audit and the docs.** Every fence rule classified by who
  enforces it; Copilot review stays experimental but is no longer "not fenced";
  the operator reads dissents at close.

Unchanged: Claude sessions — their argv, fence and settings; the primary
review decides every task unless `both-pass` is chosen; Copilot runs review sessions only; the driver's
checks stay the boundary (S-2).

## Why this shape, and what was rejected

Decided with the operator on 2026-10-26. **Do not re-litigate these** — the
planner inherits them.

- **The fence is a vloop hook, fail-closed.** Copilot runs a `preToolUse` hook
  before every tool call, shell and file tools alike, and honours a JSON deny
  and exit 2 (the probe of 2026-10-26). The hook is the vloop binary itself,
  judging against the same per-phase fence Claude sessions run under, so there
  is one rule set. What it cannot parse it denies.
  *Rejected:* a longer deny list (the next wrapper gets through); fencing Claude
  sessions with the same hook in this brief (Claude sessions stay unchanged
  since B18).
- **The deny list stays, as a second line.** `--deny-tool` still names the
  stems; the hook is the fence.
- **The second reviewer is optional, and what its verdict does is
  configuration.** The experiments found a stricter cross-family reviewer
  useful as a second opinion and costly as a veto (it would have repeated
  acceptable tasks), so three modes only add its opinion; `both-pass` makes it
  a veto for an operator who wants that cost. The choice is per repository, in
  config.
  *Rejected:* a fixed behaviour.
- **The second reviewer may be any agent and model.** Matching the primary is
  allowed, and warned about.
- **Copilot review stays experimental.** The "not fenced" wording goes; the
  label stays until more runs on another model (`gpt-5.6-terra` retires from
  Copilot on 2026-11-16).
- **The binary never calls the network**; acceptance on the real CLI is the
  operator's, at close, as B19's C5 was.

## Binding references

- `docs/domain/execution/session.md` — the session kinds, the fence and what a session never does: where the hook, the second-review kind and the Copilot denials go
- `docs/domain/domain-model.md` — S-2 and S-4, the fence is advisory and the driver is the boundary; P-6, how a session's model is resolved; C-3, schema versions and readers; the glossary, which gains second review and dissent
- `docs/domain/execution/task.md` — the task's lifecycle and notes, which the second review's findings may join and never change
- `docs/domain/measurement/metrics.md` — the metrics a run reports, which gain the second review
- `docs/design-notes/copilot-cli-experiments.md` — the deny probe (C5.2), the hook probe and the Copilot tool names and events the hook and the parser read
- `internal/driver/agent_copilot.go` — the adapter: its deny list, its per-session home where the hook goes, and the parser that gains denials
- `internal/driver/iterate.go` — the review step the second review follows, with the same guards
- `fence/review.json` — the review fence the hook enforces
- `docs/guide/configuration.md` — the keys, the Copilot section and its warnings
- `tools/copilot-acceptance/README.md` — the operator's acceptance kit this run extends

## Behaviour contract

### H1 — `vloop fence check` and `vloop fence hook`

`vloop fence hook --phase <phase> --root <dir>` reads one Copilot `preToolUse`
payload (camelCase: `toolName`, `toolArgs`, `cwd`) on stdin and prints one line
of JSON on stdout: `{"permissionDecision":"allow"}` or
`{"permissionDecision":"deny","permissionDecisionReason":"vloop fence: <reason>"}`,
exit 0. A payload it cannot read (not JSON, no `toolName`), an unknown phase,
or any internal error: the deny line and exit 2. It reads nothing but the
payload and its embedded fences, writes nothing and never runs a command.
Phases: `review` and `second-review`, which share the review fence.

`vloop fence check --phase <phase> --root <dir> --tool <name> [--path <p>]
["<command>"]` judges one call the same way, for the operator and the gates:
prints `allow`, exit 0, or `deny: <reason>`, exit 1; bad arguments exit 2.

**Shell** (`bash`, `powershell`): the command string is split into simple
commands at `;`, `&&`, `||`, `|`, `&` and newlines, outside quotes. Each simple
command loses leading `NAME=value` assignments, then is unwrapped, repeatedly:
`env` (with `-i`, `-u NAME`, `-` and assignments), `command`, `exec`, `nohup`,
`time`, `nice` (with `-n N`), `timeout` (with its options and duration),
`stdbuf` (with its options), `xargs` (with its options; the rest is the
command), and `sh`, `bash`, `zsh`, `dash`, `ksh` with `-c <string>` (options
before `-c` allowed), whose string is judged as a command string in turn. A
command word with a directory (`/usr/bin/gh`, `./bin/gh`) is judged by its base
name. Each simple command is then denied when it matches a `Bash(…)` deny rule
of the phase's fence, with the fence's matching (`:*` a prefix, `*` any run of
characters), or one of the adapter's deny-list stems. Denied outright, before
any matching: a backtick, `$(`, `<(`, `>(`, `eval`, `source`, `.` as a command,
an unterminated quote, a `-c` string that is not literal (missing, or holding
`$`), and a command word that starts with `$`. A redirection to a
file (`>`, `>>`, `>|`, `&>`) is a write of that file, judged as below;
`/dev/null` is allowed.

**Writes** (`create`, `edit`, `apply_patch`, a redirection): every path is
resolved against `--root` (an absolute path is taken as given, then cleaned);
in `review` and `second-review` it is allowed only under `.vloop/tmp/`.
`apply_patch` paths are those of its `*** Add File:`, `*** Update File:`,
`*** Delete File:` and `*** Move to:` lines; a patch with none is denied.

**Reads** (`view`, `read_file`, `grep`, `rg`, `grep_search`, `glob`,
`file_search`): allowed, except a path under the user's global Claude
directory (`~/.claude`), Copilot's (`~/.copilot`) or the GitHub CLI's
(`~/.config/gh`).

**Skills** (`skill`): K1.

**Everything else** — `web_fetch`, any tool whose name ends in `_subagent`,
`task`, and every name not listed here — is denied:
`vloop fence: tool <name> is not allowed in a <phase> session`.

### H2 — Copilot sessions run under the hook

- Each Copilot session's own home holds a hooks file whose `preToolUse` entry
  runs `vloop fence hook --phase <phase> --root <the repository root>`, with
  the running binary's path, for both the `bash` and `powershell` command
  forms. The deny list stays on the argv.
- `permission_denials` of a Copilot `session/v3` record lists every
  `tool.execution_complete` event with `error.code: "denied"`, in order, as
  `{"tool_name", "tool_use_id", "message"}`: the tool name from the
  `tool.execution_start` event with the same `toolCallId`, the message masked.
  It is `[]` when there are none, as today.

### K1 — vloop's skills only

- The hook allows `skill` only for the name of one of the embedded plugin's
  skills; any other is denied:
  `vloop fence: skill <name> is not vloop's — use /vloop:review`.
- `vloop run`, when `agent.review` or `agent.second-review` (in use) is
  `copilot`, refuses to start while a project skill directory under
  `.claude/skills/` or `.github/skills/` has exactly the name of one of the
  plugin's skills:
  `vloop: <path> has the name of vloop's <name> skill; a copilot session may load it instead — rename it`,
  exit 1. `vloop doctor` shows the same as an error row. The warning for other
  names ending in `review` stays.

### S1 — the optional second reviewer

- Keys: `run.second-review` — `off` (default), `record`, `feedback`, `on-pass`,
  `both-pass`;
  `agent.second-review` (`claude`, `copilot`; default `claude`),
  `model.second-review` (default `sonnet`), `effort.second-review`; each with
  its environment variable as every key has. P-6 applies without the task
  level: a task's `model`/`effort` never sets the second review's. A copilot
  second review needs an explicit Copilot model, as `agent.review` does.
- When not `off`, a second review session runs after the primary review of an
  iteration — after every primary review for `record` and `feedback`, only
  after a primary PASS for `on-pass` and `both-pass` — never when the primary
  did not run or
  ended the run. It runs `/vloop:review <task>` as a fresh session of phase
  `second-review`, under the review fence (the hook's `second-review` phase
  for copilot), with the same guards as the primary: inputs restored, a moved
  ref or git setting halts (exit 9), changed files reverted. It never sees the
  primary's verdict: the driver keeps it out of `.vloop/tmp/` first.
- Its verdict is read as the primary's is; no valid verdict, a changed file or
  tampering make it FAIL, with the finding. **In `record`, `feedback` and
  `on-pass` it never changes the outcome, the status or the attempts of the
  task**; a session error or timeout is a warning and the iteration goes on.
- `both-pass`: a task is done only when the primary and the second both PASS.
  A second FAIL rejects the iteration as a primary FAIL does — the attempt
  counts, the task is retried or blocked by the same rules — with the second's
  findings as the task's notes, each prefixed `second review: `. A second
  session error or timeout is treated as the primary's would be.
- `feedback`: when the primary did not PASS, the second's findings join the
  task's notes after the primary's, each prefixed `second review: `.
  `record` and `on-pass`: nothing reaches the notes.
- A **dissent** is a second verdict that differs from the primary's. The run
  log says `   second review: <PASS|FAIL>`, followed by `(dissent)` when it is
  one; the verdict is kept as `reports/<iter>-second-verdict.json`; the end of
  the run lists `second review: <n> dissent(s): <task> (primary <v>, second <v>), …`
  when there are any.
- Records: `session/v4` is `session/v3` with `second-review` among the phases;
  `iteration/v3` adds `second_review` (`verdict`, `dissent`), absent when none
  ran; `metrics/v4` adds `second-review` to `agents` and a `second_review`
  object (`mode`, `sessions`, `dissents`, `dissent_tasks`). Readers keep
  accepting v3, v2 and earlier; `vloop schema validate` knows the new ones.
- Its cost counts toward the cost ceiling like any session's.
- `vloop doctor`, when `run.second-review` is not `off`: a row naming the mode,
  agent and model; the copilot checks when the agent is copilot; and
  `! run.second-review: the second reviewer is the primary's agent and model — it adds no independence`
  when they match.

### A1 — the audit and the docs

- docs/design-notes/fence-audit.md: every rule of `fence/review.json` and every
  deny-list stem, classified by who enforces it for a Copilot review session —
  the driver after the session, the hook, or neither — with what covers the
  last.
- The Copilot warning becomes
  `! agent.review is copilot: experimental (docs/guide/configuration.md)`
  (and the same for `agent.second-review`); the configuration guide describes
  the hook, the fail-closed rules, K1, the keys of S1 and what each mode does.
- The domain pages gain the second-review kind, dissent and the hook;
  the operate skill and `.claude/skills/vloop-operator/SKILL.md` read the
  dissents at close and record each one that holds as a `verification-finding`.

### Violations the review must rule on

- A form in the worked example's deny column that `fence check` allows.
- A Copilot session started without the hooks file, or a hook that answers
  allow on an error.
- A second review that changes a task's outcome, status or attempts in any mode
  but `both-pass`, a `both-pass` task done without both PASS verdicts, or a
  second review that can read the primary's verdict.
- Any change to a Claude session's argv, fence or settings.
- A gate that runs `copilot` or touches the network.

## Worked example

`vloop fence check --phase review --root <repo> --tool bash "<command>"` in a
fixture repository:

```
allow   git diff 1a2b3c..HEAD -- src/
allow   git log --oneline | head -5
allow   vloop task show T3
allow   git diff > .vloop/tmp/diff.txt
allow   ls src > /dev/null
deny    sh -c "gh pr list"
deny    env gh auth token
deny    env -i PATH=/bin git push origin main
deny    bash -lc 'git push'
deny    /usr/bin/curl https://example.com
deny    nohup timeout 5 wget x
deny    git status && xargs rm -rf
deny    echo $(gh auth token)
deny    eval "git push"
deny    git diff > .vloop/state/x
deny    sh -c "$CMD"
deny    echo 'unterminated
```

Every `deny` exits 1 and names the rule or form; every `allow` exits 0.

Other tools (`--tool`, `--path`), phase `review`:

```
allow   create  .vloop/tmp/verdict.json
deny    create  .vloop/state/state.json
deny    edit    src/app.ts
allow   view    src/app.ts
deny    view    ~/.claude/settings.json
deny    web_fetch
deny    search_code_subagent
allow   skill   review
deny    skill   vloop-review          -> vloop fence: skill vloop-review is not vloop's — use /vloop:review
deny    frobnicate                     -> vloop fence: tool frobnicate is not allowed in a review session
```

`vloop fence hook --phase review --root <repo>` with `not json` on stdin: the
deny line, exit 2. An `apply_patch` payload with `*** Update File: .git/config`:
deny.

Second review, with a stub agent, a two-task plan where the primary passes T1
and fails T2 once, and the second fails T1 and fails T2:

```
run.second-review = feedback
  T1 done (primary PASS); run.log "   second review: FAIL (dissent)";
  T2 retried with notes holding the primary's findings, then
  "second review: <finding>"; end of run:
  "second review: 1 dissent(s): T1 (primary PASS, second FAIL)"
run.second-review = record      same outcomes; T2's notes hold the primary's only
run.second-review = on-pass     second review runs for T1 only
run.second-review = both-pass   second review runs for T1 only, and fails it:
                                T1 is retried with notes "second review: <finding>";
                                with the second failing every time, T1 ends blocked
                                after its attempts, exactly as a primary FAIL would
run.second-review = off         no second-review session, no line, no field
```

With .claude/skills/review/SKILL.md in the fixture and `agent.review =
copilot`: `vloop run` exits 1 with the K1 message; renamed to
.claude/skills/repo-review/, it starts.

## Out of scope

- Fencing Claude sessions with the hook.
- Copilot in phases other than review and second review; plan, work,
  gate-review and verify on another agent.
- More than two reviewers.
- A backend-agnostic eval runner.
- Reporting vloop defects upstream (B23).
- Dropping the experimental label.

## Constraints

- Go, at the version in `go.mod`; no new dependencies.
- Must keep building and passing on macOS, Linux and Windows; the shell
  splitter and the path rules are pure functions tested on every OS.
- **Gates run on macOS's BSD tools**, and **no gate runs `copilot`, `gh` or
  touches the network**: a stub agent stands in, as in B19.
- Tests build fixtures in temporary directories and address them with
  `git -C`; sessions never move refs.
- **Gates follow the diff surface; a gate never runs the suite (P-8).**
- **No task may weaken a check to pass.** An earlier test changes only where it
  pins behaviour the contract changes, and only to the new behaviour; the work
  session names the contract item in its notes and the review cites it.
- **New tooling is proven by fixtures, including where it must fail**: every
  deny line of the worked example, the unreadable payload, the K1 refusal, each
  mode of S1.
- Repo-relative paths everywhere. No absolute paths in any file or commit
  message.

## Acceptance on the real CLI, at close (the operator's)

`tools/copilot-acceptance/` gains, in this run: a probe of the review session,
under the hook, asking for each deny form of the worked example and a skill
other than vloop's; the hook's failure modes (its command missing, exit 1,
output not JSON, a timeout), recording whether Copilot then runs the call;
and a run with `run.second-review = feedback` and `agent.second-review =
copilot`. The operator runs it with a Copilot model of their choice and records
the outcome in `docs/design-notes/copilot-cli-experiments.md` before close. A
failure mode in which Copilot runs the call is recorded as a limit in the
configuration guide; the driver stays the boundary.

## Shape

7 to 9 tasks, each verifiable by a single command. Order is load-bearing:

1. **H1 — the judge**: the splitter, the unwrapping, the matching and the path
   rules, as `vloop fence check`. Gated on the worked example's tables.
2. **H1 — `vloop fence hook`**: the payload, the JSON answer, exit 2 on error.
3. **H2 and K1 — Copilot sessions under the hook**, the denials parsed, the
   skill rule and the preflight refusal. Gated with the stub agent.
4. **S1 — the keys and doctor.**
5. **S1 — the second review session**, its modes, the records and the end-of-run
   list. Gated on the worked example's second-review lines.
6. **S1 — `metrics/v4`.**
7. **A1 — the audit, the guide, the domain pages and the skills.** Prose task:
   if it fails review twice on wording, block it and the operator finishes it.
8. **The acceptance kit.**
9. **Close:** every gate at once, and the worked example line for line.

# Copilot review acceptance (B19's C5)

The operator's acceptance of `agent.review = copilot` on the real GitHub Copilot
CLI. Gates use a stub `copilot`; this proves what a stub cannot.

```
npm install -g @github/copilot     # then: copilot, and log in once
gh auth status                     # vloop's Copilot sessions log in through gh
tools/copilot-acceptance/run.sh <copilot model id> [dir]
```

Each vloop Copilot session has its own `COPILOT_HOME`, which holds no login, so
Copilot logs in with `gh auth token`: the GitHub CLI must be logged in. The `gh`
shim of the probe passes that one form through.

`<dir>` defaults to `~/vloop-copilot-acceptance` and must not sit inside a vloop
repository. The first run may stop at `vloop doctor` asking for Claude's folder
trust: run `claude` once in `<dir>/repo`, accept, and re-run (the path stays the
same, so the trust holds).

It does two things and writes `<dir>/c5-summary.md`:

1. **A real run.** A fixture brief (`greet.loop-brief.md`) in a scratch
   repository, plan and work on Claude, reviews on Copilot with the model given,
   using a vloop built from this tree. The report lists every session record
   (agent, model, credits, estimated cost, `session/v3` validity) and every
   verdict.
2. **A deny probe.** The run's last review session, restarted with exactly the
   argv and environment the adapter built, with only the prompt replaced: the
   model is asked to run each form B19's C2 must deny. Shims for `gh`, `curl`,
   `wget`, `ssh`, `scp`, `nc`, `rm` and `git` stand first on `PATH`; they record
   any form that got through and do nothing outward. The report says, per
   form, whether Copilot denied it, and which JSON Lines events mention a
   denial (the event B20 needs).

Record the summary in `docs/design-notes/copilot-cli-experiments.md` before
B19 is closed. It spends Claude money for the plan and work sessions, and
Copilot credits for the reviews and the probe.

## B22: the hook, its failure modes and a second reviewer

Three more parts, each its own script, run after `run.sh` has populated `<dir>`
(the first two reuse its logged review session and shims). None of them is run
by a gate or a check: they need the real `copilot`, `claude`, `gh` and the
network, and spend Copilot credits (and Claude money for the last).

```
tools/copilot-acceptance/hook.sh [dir]                       # the hook probe + the four failure modes
tools/copilot-acceptance/second.sh <copilot model id> [dir]  # the feedback-mode run
```

1. **The hook probe** (`hook.sh`, `hook.py forms`). The review session runs
   under `vloop fence hook`, with the hooks file vloop wrote in its
   `COPILOT_HOME`, and is asked to run each deny form of the worked example's
   shell table (`sh -c "gh pr list"`, `env gh auth token`, ... through
   `echo 'unterminated`) and to invoke a skill other than vloop's
   (`vloop-review`). The summary records, per call, whether Copilot ran it or
   the hook denied it, with the message.
2. **The hook's failure modes** (`hook.sh`, `hook.py failure`). The same
   session with the hook command replaced by one that is missing, exits 1 (`exit 1`),
   prints output that is not JSON, or exceeds its timeout. Each is asked to run one call
   the shims record; the summary says whether Copilot then ran it (fails open)
   or refused it (fails closed).
3. **A real run with a second reviewer** (`second.sh`). The fixture brief in a
   scratch repository with `run.second-review = feedback` and
   `agent.second-review = copilot`, plan, work and primary review on Claude.
   The summary lists the second-review session records (`session/v4` validity,
   credits, model), the run log's second-review lines and the dissents.

The probes rebuild the hooks file each time, since the session's `COPILOT_HOME`
lies under `.vloop/tmp/`, emptied after every session; `VLOOP_HOOK_PHASE`
(`review` by default) sets its phase, `second-review` when the logged session
was a second review.

Record the summaries (`b22-hook-summary.md`, `b22-second-summary.md`) in
`docs/design-notes/copilot-cli-experiments.md` before the brief is closed. The
hooks-file shape vloop writes is unconfirmed against Copilot's docs: if the
failure probes find no hook file to edit, that is the first thing to check.

## B22.3: getting past the hook

Two standalone scripts, outside vloop: each case its own scratch repository
and `COPILOT_HOME` with one `preToolUse` hook, under `--allow-all-tools`. They
need `copilot` and a logged-in `gh`, spend Copilot credits, and print their
results (`fail-open.sh` also writes `<workdir>/summary.md`).

```
tools/copilot-acceptance/fail-open.sh <copilot model id> [workdir]
tools/copilot-acceptance/bypass.sh <copilot model id> [workdir]
```

- **`fail-open.sh`** — does a hook that means deny stop the call when it
  prints deny JSON, exits 2 or 1, prints a log line before its JSON, prints
  text that is not JSON, or answers after `timeoutSec`?
- **`bypass.sh`** — ten parallel calls under a fast and a slow hook (are hooks
  queued, and does a queued call time out open?), and the hooks file
  overwritten or deleted mid-session (is it re-read?).

Re-run both when Copilot updates; the results are in
`docs/design-notes/copilot-cli-experiments.md` → B22.3.

# vloop

A Go CLI and Claude Code plugin that packages the autonomous loop: plan a brief
into tasks, then per task a fresh work session, a gate, an independent review,
one commit. B1–B7 were built **by** the shell loop vendored in `.loop/`; from
B8 it builds itself with `vloop run`, one brief at a time, following `docs/design-notes/vloop-roadmap.md` — read that first; its
"Owns" column says what exists and what comes next. The domain — vocabulary,
invariants, contexts — is `docs/domain/README-domain.md`. The briefs live in
`docs/briefs/`, each with a `## Run record` once consumed.

## Which rules bind you

- **Started by `vloop run`** (or, for B1–B7, `.loop/run.sh`) as a plan, work
  or review session: the loop's rules below and the vloop section bind you, all
  of them. You do not commit, you do one task, you set no status.
- **An interactive session with the operator**: you are the operator's hands,
  not a loop session. Writing a brief follows
  `.claude/skills/vloop-architect/SKILL.md` (the design act); running,
  verifying and closing one follows `.claude/skills/vloop-operator/SKILL.md`. You may commit
  on work branches; merging, pushing and anything outward-facing wait for the
  operator's go-ahead. Rules 1, 2 and 8 below bind you too.

## Toolchain

Go, as pinned in `go.mod`; module `github.com/mvelosop/vloop`, binary built from
`./cmd/vloop`. The checks every change must pass:

```
go test ./...   go vet ./...   gofmt -l .   go mod tidy -diff
GOOS=linux go build ./...   GOOS=windows go build ./...
```

`gofmt -l .` and `go mod tidy -diff` print nothing on success.
`.vloop/config.toml` is this repo's own vloop config (`metrics.stacks = ["go"]`);
`.vloop/defects/` holds its recorded defects. The shell loop's own suite is
`.loop/tests/run-all.sh`.

<!-- vloop:begin -->
## vloop

Sessions started by the vloop loop (plan, work, review) follow these rules:

1. Use repo-relative paths only, in files, logs and commit messages.
2. A work session does one task and stops; it does not start the next one.
3. Sessions never commit, set a task status or move git refs; the driver does.
4. A task is done only when its gate passes and the review passes it.
5. Halting cleanly with an account of what blocked you is a success; faking progress is the only failure.

In an interactive session you are the operator's hands: follow the operator skill (vloop-operator), not these session rules.

Written by vloop 1.0.0.
<!-- vloop:end -->

<!-- loop:begin -->
## Rules for any session working here

1. **All durable state stays in this repo.** Never write session state, memory,
   or config to `~/.claude` or any other global location. Drift there is
   invisible and unrepeatable. This is the rule everything else rests on.
2. **Repo-relative paths only** — in files, logs, journal entries, and commit
   messages. Never `/Users/...`. Write `~/...` if you must show an absolute
   path. The driver masks `$HOME` and the username out of everything it
   persists, but that is a backstop, not your excuse.
3. **One task per iteration.** Do the task you were given and stop. Do not start
   the next one even if it is trivial and you have the context. Running ahead
   desynchronizes the plan from reality and is the main failure mode of
   autonomous loops.
4. **You do not set task status.** The work session *proposes*; the gate and the
   review *dispose*. Status transitions belong to the driver.
5. **You do not commit.** The driver makes exactly one commit per iteration,
   covering code, state, journal and telemetry together.
6. **A task is done only when its verify command exits 0 and the review session
   passes it.** Claiming done without both just costs an attempt.
7. **No web access.** `WebFetch`/`WebSearch` are denied so a run cannot drift
   with the internet. Package installs are fine.
8. **Halting cleanly, with a clear account of what blocked you, is a success.
   Faking progress is the only real failure.**
Loop docs: `.loop/manual.md` (start here) and `.loop/README.md`. Worked briefs:
`.loop/examples/`. State lives in `.loop/state/state.json`; the driver owns it.
<!-- loop:end -->


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

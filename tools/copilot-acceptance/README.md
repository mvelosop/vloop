# Copilot review acceptance (B19's C5)

The operator's acceptance of `agent.review = copilot` on the real GitHub Copilot
CLI. Gates use a stub `copilot`; this proves what a stub cannot.

```
npm install -g @github/copilot     # then: copilot, and log in once
tools/copilot-acceptance/run.sh <copilot model id> [dir]
```

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

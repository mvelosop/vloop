# Plan — B20261103-0058-b34-i82-fence-backstop

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **3/4 done** · iteration 4

**Brief:** `docs/briefs/B20261103-0058-b34-i82-fence-backstop.loop-brief.md` · **Updated:** 2026-11-03T02:25:10Z

## Progress

- [x] **T1** — Re-judge every call a Copilot work session ran, and halt the run on one its fence denies (F1) · 1 attempt(s)
- [x] **T2** — Halt every Copilot phase on a bypassed fence, and on a changed hook binary (F1 everywhere, F2)
- [x] **T3** — Give the hook a timeoutSec of 120 and document the backstop (F3, F4)
- [ ] **T4** — Close the run: every gate at once, and the worked example line for line

## Tasks

### T1 — Re-judge every call a Copilot work session ran, and halt the run on one its fence denies (F1)

`done` · 1 attempt(s) · depends on: none

A Copilot session is fenced by the preToolUse hook, which fails open when it misbehaves; nothing behind it notices. This task adds the driver's re-judge of every call a Copilot session ran — read from its JSON Lines, judged by the same fence judge the hook uses — and wires it, with its stderr lines and the FENCE BYPASSED exit-9 halt, into the work phase. T2 extends the same after-session check to every other phase and adds the binary check.

**Acceptance**

- After every Copilot work session the driver reads the session's JSON Lines; a call that ran is a `tool.execution_start` whose `tool.execution_complete` of the same `data.toolCallId` does not carry `data.error.code` `denied`, or that has no completion.
- Each call that ran is judged by `fence.New(phase, root, home)` and `Decide` — the hook's own judge — with tool, path and command read from `data.toolName` and `data.arguments` exactly as `parseHookPayload` reads `toolName` and `toolArgs` (an object, or a JSON string holding one; `path`; a skill's `skill` or `name`; apply_patch's `input`, `patch` or `command`), and with the hook's rule first: a `bash` or `powershell` call with an empty command is denied with `the <tool> call has no command the hook can read`. Violation: a second argument parser that can diverge from the hook's, or any judge or rule list other than the hook's.
- The driver does not import `internal/fence` (which imports `internal/driver`); the judge reaches the driver another way (for example as a function the CLI hands it). The check is the phases', not `Runner.Run`'s: `tools/agent-eval` and its tests are unchanged.
- With one or more denied, stderr has one line per denied call, `vloop: <tool> <what> — <the judge's reason>`, `<what>` the call's path, else the first line of its command cut to 200 runes with `…` added when cut, else nothing (`vloop: <tool> — <reason>`); then the run halts with exit 9 and `FENCE BYPASSED work — the session ran <n> call(s) its fence denies; nothing from it was used`. Every line is masked (`$HOME` as `~`) and written to the run log too, as `REFS MOVED`'s warnings are.
- The check runs as soon as the work session ends, whatever its exit, before the phase's own checks (so before REFS MOVED); the phase still makes its restores of the driver's inputs before exiting 9; the halt comes before the proposal is used or anything is committed, and the work tree is otherwise left as the session left it. With no call denied nothing changes and nothing is printed. A Claude session's handling is unchanged (violation: any change to how a Claude session is run, parsed or halted).
- Committed tests: the worked example's first three lines as one `cmd/vloop` test each (a work session's bash `git commit -m x` completed without error -&gt; the two stderr lines, exit 9, no new commit; the same call completed `denied` -&gt; the iteration goes on; a bash `ls` only -&gt; the iteration goes on), with the harness's stub `copilot` gaining, in test code, the prompts of every phase (`/vloop:<phase>` with or without an argument) and the tool events a test's copilot-script.sh asks for; and tests of the reading of fixture JSON Lines (a call with no completion, a denied completion, the path / first-line / 200-rune / nothing forms of `<what>`, the empty-command rule). `TestRunHarnessCopilot`'s refusal of `/vloop:work T1` changes only to pin a prompt the stub still refuses; no other existing test changes.
- No fence's rules change: `fence/*.json` and the judge's decisions are as at the base.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T1/gate.sh
```

</details>

### T2 — Halt every Copilot phase on a bypassed fence, and on a changed hook binary (F1 everywhere, F2)

`done` · depends on: T1

T1's re-judge halts only the work phase; a call can get past the hook in any phase. This task runs the same after-session check after every Copilot session of plan, gate-review, review, second-review, verify and review-brief, adds the SHA-256 check of the hook's binary before and after each session (HOOK CHANGED, exit 9, checked before F1), and makes Runner.Run hand the phase the session's stdout even when no result event was printed, so a killed session is re-judged.

**Acceptance**

- After every Copilot session of every phase (plan, gate-review, work, review, second-review, verify, review-brief) the re-judge of T1 runs with that phase's judge (`second-review` judged as the hook judges it, by the review fence), and a denied call that ran halts with exit 9 and `FENCE BYPASSED <phase> — the session ran <n> call(s) its fence denies; nothing from it was used`, preceded by the per-call lines. The halt comes where REFS MOVED halts for that phase: a plan or gate-review halt leaves no plan commit and starts no work; a review or second-review halt commits nothing; a verify halt keeps no reports/verify.json; a review-brief halt prints no verdict and its brief is restored.
- Before each Copilot session the driver takes the SHA-256 of the binary the hook runs (the one writeCopilotHooks writes into the hooks file); after it, again. A difference, or the binary gone, halts with exit 9 and `HOOK CHANGED <phase> — the binary the fence hook runs changed during the session: <path>; nothing from it was used`, `<path>` as the hooks file names it, masked. F2 is checked before F1, both as soon as the session ends and before the phase's own checks, so a call that got past the hook and also moved a ref is named before REFS MOVED; the phase's restores (the driver's inputs; a review-brief session's brief) are still made, and a restore that fails is reported after the halt's lines.
- The lines and halts are masked and written to the run log too where the phase has one; `vloop brief review` has no run log, so there stderr only.
- `Runner.Run` hands the phase the session's stdout whether or not it printed a result event, so a session killed at its timeout, or one that died, is re-judged; a call with no completion counts as run. Claude sessions and `tools/agent-eval`'s scoring are unchanged.
- Committed tests: the worked example's seven phase lines, one `cmd/vloop` test each (agent.&lt;phase&gt; = "copilot", a session that runs a bash `git push`; second-review with `run.second-review = "record"`; verify with the last task done; review-brief through `vloop brief review`), each asserting the `vloop: bash git push — …` line, the halt, exit 9 and what was not used; a driver test of the phases' shared after-session check, phase work, whose hook binary is a copy of a file the test changes between the check's before and after (the HOOK CHANGED halt); and a driver test of a stub Copilot session run through `Runner.Run` with a Timeout of 1 s that prints a started bash `git commit -m x` and sleeps, whose result handed to the shared check gives the FENCE BYPASSED work halt.
- No fence's rules change, and no judge but the hook's is used.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T2/gate.sh
```

</details>

### T3 — Give the hook a timeoutSec of 120 and document the backstop (F3, F4)

`done` · depends on: T2

The hook answers in milliseconds and a timeout only ever lets a call run, so its limit should be far above any stall: 120 seconds for every phase. And every place that defines exit 9 or describes the hook's fail-open must now say the driver backs it, so an operator reading a FENCE BYPASSED or HOOK CHANGED halt finds it explained and knows to inspect before resuming.

**Acceptance**

- The hooks file the driver writes has `"timeoutSec": 120` for every phase; a new test reads the hooks file the driver writes and pins it.
- Every F4 edit is made word for word as the brief gives it, wrapped in the file's style: domain-model.md's S-2 (the sentence after `It halts (exit 9) if a ref …`) and R-3; docs/domain/execution/session.md's sentence after `halts the run with exit 9 and nothing is committed.`; concepts.md's row `9` (its first cell, the `, or a Copilot session ran …` insertion straight after `or the git hooks` and before `; nothing was committed`, and its last cell); configuration.md's sentence; the fence audit's note under each of its six headings, on a line of its own; docs/domain/execution/run.md's row `9` first cell; copilot-cli-experiments.md's `**What remains:**` paragraph; the operate skill's row `9`.
- `vloop run --help` reads `9 refs or repository configuration moved, or a Copilot session got past its fence` (internal/cli/run.go) and `vloop brief review --help` reads `9 the session changed the brief, git refs or the repository configuration, or a Copilot session got past its fence.` (internal/cli/brief_review.go); the generated reference docs, if any, follow whatever regenerates them today.
- Violation: any other sentence of those files changed, or a doc stating behaviour the driver does not have.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T3/gate.sh
```

</details>

### T4 — Close the run: every gate at once, and the worked example line for line

`pending` · depends on: T3

Each gate proved its task when it ran; the close proves them together on the final tree and accounts for the brief's worked example line by line, so the verify session and the operator find every line covered by a gate or a committed test, or named as a gap.

**Acceptance**

- Gates T1, T2 and T3 all pass on the final tree, run together by this task's gate.
- The work session's notes walk the brief's worked example line for line (the three work lines, the HOOK CHANGED driver test, the killed-session driver test, the timeoutSec test, the seven phase lines) and name the committed test, and the gate, that covers each; a line nothing covers is fixed in this task (a test added) or reported, not skipped.
- Every configured check is green on the final tree (the `go` check: go test, vet, gofmt, mod tidy, linux and windows builds).

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T1/gate.sh && sh .vloop/state/gates/T2/gate.sh && sh .vloop/state/gates/T3/gate.sh
```

</details>


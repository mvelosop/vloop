# Plan — B20261120-1643-b39-local-config-copilot-only

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **5/8 done** · iteration 7

**Brief:** `docs/briefs/B20261120-1643-b39-local-config-copilot-only.loop-brief.md` · **Updated:** 2026-11-20T18:56:54Z

## Progress

- [x] **T1** — Add the local config layer to resolution, with source local
- [x] **T2** — Add vloop config set --local, config path --local and the override notice · 1 attempt(s)
- [x] **T3** — Keep the local file git-ignored and in view: init, upgrade, doctor and the run's preflight · 1 attempt(s)
- [x] **T4** — Keep sessions off the local file: the input guard and the fences
- [x] **T5** — Check Claude only where a session in use runs on it
- [ ] **T6** — Price a Copilot credit at GitHub's $0.01 by default
- [ ] **T7** — Document the local file, Claude only where used, and the credit price
- [ ] **T8** — Close: the worked example line for line, and every gate at once

## Tasks

### T1 — Add the local config layer to resolution, with source local

`done` · depends on: none

A developer's own settings live in an untracked .vloop/config.local.toml that wins over the shared .vloop/config.toml, while a VLOOP_* variable still wins over both (L1). This task puts that layer into the config package's resolution, so every later task (config set --local, doctor's row, the run's guards) builds on one resolver. The checks are the repository's contract, so a [[check]] in the local file is refused wherever the config is read, never read.

**Acceptance**

- Per key, the value is the VLOOP_&lt;KEY&gt; variable when set and not empty, else .vloop/config.local.toml when it sets the key, else .vloop/config.toml when it sets the key, else the default.
- A list key takes the whole list of the layer that wins (never a merge of two files); an empty list (metrics.code = []) in either file sets nothing, so the next layer decides.
- A value from the local file has source local: `agent.work=claude (local)` in config list, "source":"local" in config list --json and config get --json; the sources default, file and env keep their names. Violation: a key resolved from the local file whose source prints as file.
- An empty local file, or one holding only keys vloop does not know, changes nothing and is no error.
- A bad value or malformed TOML in the local file is reported like the shared file's, with the local path, exit 1: `vloop: .vloop/config.local.toml: invalid value "nope" for agent.work: want one of claude, copilot`.
- A [[check]] in the local file is an error, exit 1, on every command that reads the config (config list and config get included), with exactly `vloop: .vloop/config.local.toml: [[check]] belongs in .vloop/config.toml — the checks are the repository's, not one machine's`. Violation: any path (Checks, the planner's check copy, doctor) that returns or runs a check defined in the local file.
- The local file is read wherever and whenever .vloop/config.toml is read today, ResolveRun at a run's start included; no read point is added or moved. With no local file, every output is byte-identical to before.
- The config package's committed tests cover, in temporary directories: the precedence of all four layers, the list rule and the empty list, source local, the local bad-value and malformed-TOML errors, and the [[check]] refusal; the go check runs them. A task that ships this behaviour and leaves no such test behind fails.
- No new dependency: the TOML library in use stays the only one.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T1/gate.sh
```

</details>

### T2 — Add vloop config set --local, config path --local and the override notice

`done` · 1 attempt(s) · depends on: T1

With the local layer resolving (T1), a developer needs to write it without editing TOML by hand, and to be told when a shared value they just set will not apply on their machine (L2). This task adds --local to config set and config path, and the one-line stderr notice config set prints when the local file or the environment still overrides the key it wrote. docs/guide/commands.md is generated from the commands and must be regenerated.

**Acceptance**

- vloop config set --local &lt;key&gt; &lt;value&gt; validates as config set does (exit 2 and the same message on an unknown key or an invalid value, nothing written) and writes the key into .vloop/config.local.toml, keeping the file's other keys, comments and layout.
- config set --local creates the file (and .vloop/) only when it writes a value, and never creates .vloop/config.toml; '' removes the key; removing a key that is not there writes nothing and creates nothing.
- Without --local, config set writes the shared file as today. When, after the write, the key's value still comes from the local file or the environment, it prints one line on stderr and exits 0: `! agent.work stays claude on this machine: .vloop/config.local.toml overrides it`, or `: VLOOP_AGENT_WORK overrides it`; when both override, the line names the variable. config set --local prints the variable's line when a VLOOP_* variable still overrides the key after its write. No notice when nothing overrides; stdout stays empty in every case.
- vloop config path --local prints .vloop/config.local.toml; config path alone prints .vloop/config.toml.
- docs/guide/commands.md is regenerated with go generate ./... and its existing test passes under the go check.
- Committed tests cover set --local (write, layout kept, create only on a write, remove, absent remove, validation), each notice case and path --local; the go check runs them.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T2/gate.sh
```

</details>

### T3 — Keep the local file git-ignored and in view: init, upgrade, doctor and the run's preflight

`done` · 1 attempt(s) · depends on: T2

A local file that git would commit is not local, and one that silently overrides shared values surprises its owner (L3, first half). This task makes vloop init and vloop upgrade keep .vloop/config.local.toml in .gitignore, adds doctor's config.local row (problem when tracked or not ignored, a warning naming the overridden keys), and makes vloop run's preflight refuse on those problems by name and print the warning, before anything runs. The driver's tree-clean refusal stays as it is; this refusal comes first and says why.

**Acceptance**

- vloop init and vloop upgrade keep .vloop/config.local.toml in .gitignore on a line of its own, beside .vloop/tmp/, with that line's rules: added once, nothing else in the file changed, reported as today's .gitignore step. Earlier tests that pin the exact .gitignore change only to include the new line.
- doctor's config.local row comes right after the config row: no row when there is no local file or when the config row is a problem; outside a git repository the ignore test is skipped.
- Tracked (in git's index, ignored or not): `✗ config.local .vloop/config.local.toml is tracked by git — it is shared with everyone who clones: run git rm --cached .vloop/config.local.toml and add it to .gitignore`.
- Untracked and not ignored as the driver's git sees it: `✗ config.local .vloop/config.local.toml is not git-ignored — the next git add -A would commit it: add it to .gitignore`. The test asks git about the file path itself, so any rule that covers it (a .vloop/*.local.toml pattern, the global excludes file) satisfies it.
- Untracked and ignored: `! config.local overrides model.work, agent.work — on this machine only`, the keys with source local in config list's order; with none, `✓ config.local`.
- vloop run's preflight counts either config.local problem like any other (exit 1, `vloop: preflight failed — N problem(s), nothing has run`, no session, no commit, no branch) and prints the warning; both on stderr with two spaces before the symbol.
- Committed end-to-end tests in cmd/vloop, with stub agents on a temporary PATH and repositories in temporary directories (the pattern of cmd/vloop/e2e_test.go), cover each planted failure of the worked example: not ignored and tracked, in doctor and in the run's refusal; plus tests of the row's other cases. No test touches this repository's .vloop/ or .gitignore. The go check runs them.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T3/gate.sh
```

</details>

### T4 — Keep sessions off the local file: the input guard and the fences

`done` · depends on: T3

Sessions must not change the configuration the driver runs under, and an ignored file is exactly what the driver's tree checks do not see (L3, second half). This task makes the driver's input guard hold .vloop/config.local.toml as it holds .vloop/config.toml, adds the matching deny entries to every fence that denies the shared file, and records them in the fence audit. The Copilot hook already denies every write under .vloop/ but .vloop/tmp/ and does not change.

**Acceptance**

- The input guard holds .vloop/config.local.toml like .vloop/config.toml: a session that changes it, creates it or removes it has it put back (changed: original bytes; created: removed; removed: restored), after work and review sessions alike, as the guard does today for the shared file.
- Each fence whose deny list names Edit(.vloop/config.toml) and Write(.vloop/config.toml) (work, review, verify, gate-review, review-brief) names Edit(.vloop/config.local.toml) and Write(.vloop/config.local.toml) the same way; fence/plan.json and the Copilot hook are unchanged.
- docs/design-notes/fence-audit.md has a row for each new entry in each of those fences' sections, saying what enforces it as the config.toml row beside it does.
- The open fence issues (#129, #130, #134, #136) are not touched, even in the fence files this task edits.
- Violation: the local file staged or committed by any driver commit in a test run.
- Committed tests cover the guard for the local file changed, created and removed (driver safety tests), and any existing test that enumerates the fences' rules or their audit rows passes with the new entries; the go check runs them.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T4/gate.sh
```

</details>

### T5 — Check Claude only where a session in use runs on it

`done` · depends on: T3

A machine with no claude cannot run vloop today even when every session is on Copilot, because doctor's claude row and an untrusted-folder warning stop the run's preflight (L4). This task makes doctor's claude, trust, plugin and gate edits rows apply only when a session in use resolves to the claude agent, by the same rule copilotKeys already follows for Copilot, and n/a otherwise, so the run's preflight passes them. A config that cannot be read counts Claude as in use, so a broken config never hides a missing claude.

**Acceptance**

- Claude is in use for vloop run when any of agent.plan, agent.gate-review, agent.work, agent.review resolves to claude, or agent.verify does while run.verify is not false, or agent.second-review does while run.second-review is not off; for vloop doctor the same, or agent.review-brief resolves to claude.
- When the config row is a problem, Claude counts as in use and the four rows run as today.
- When Claude is in use the claude, trust, plugin and gate edits rows are exactly as today. When it is not, each is n/a with the message `no session in use runs on claude` (`- claude no session in use runs on claude`, and the same for trust, plugin and gate edits), runs nothing (no claude --version, no plugin list) and reads nothing (no ~/.claude.json, no settings file); vloop run's preflight passes them.
- Violation: a Claude row skipped on a repository where some session in use is on Claude, or a Claude row still run when none is.
- vloop brief review's preflight is unchanged. Earlier doctor tests change only where they pin rows for Claude not in use.
- A committed end-to-end test in cmd/vloop, on a temporary PATH holding a stub copilot and no claude, every agent.* key copilot with a Copilot model id and run.verify = false, asserts each of the four n/a lines on its own (not as one block) and that vloop run gets past its preflight; with agent.review set back to claude it asserts `✗ claude claude is not on PATH` and the preflight's refusal. No test depends on what is installed on the machine. The go check runs it.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T5/gate.sh
```

</details>

### T6 — Price a Copilot credit at GitHub's $0.01 by default

`pending` · depends on: T1

Copilot's estimated cost uses 0.0066667 dollars a credit, Copilot Pro's average, while GitHub now prices one AI credit at $0.01 on every plan (L5). This task moves the default of copilot.credit-usd to 0.01 and updates every place that states the default, while records of what already happened keep their numbers. A developer whose plan makes credits cheaper sets their own, now possible per machine in the local file.

**Acceptance**

- copilot.credit-usd defaults to 0.01: `vloop config get copilot.credit-usd` prints 0.01 with no file setting it, and config list shows `copilot.credit-usd=0.01 (default)`. Its validation is unchanged.
- The configuration guide's row has default `0.01` and says: dollars per GitHub Copilot AI credit; GitHub's price is $0.01 on every plan; a developer whose plan makes included credits cheaper sets their own in the local file.
- Every test that pins the default (the expected config list in internal/config, internal/cli and cmd/vloop/e2e_test.go) is changed to 0.01 and nothing else in it changes; the go check runs them.
- Violation: 0.0066667 left anywhere a default is stated, outside the files that record or quote what already happened (docs/briefs/, .vloop/state/, docs/design-notes/), which keep it.
- No session record is recomputed.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T6/gate.sh
```

</details>

### T7 — Document the local file, Claude only where used, and the credit price

`pending` · depends on: T4, T5, T6

The behaviour of T1–T6 has to be findable where a developer looks: the configuration guide, the domain's C-1, the platform context, the concepts guide, the README's Copilot section and the roadmap's inherited decisions. This task writes those passages and nothing else; the generated commands guide is already T2's. If its review fails twice on wording alone, it is blocked and the operator finishes it.

**Acceptance**

- docs/guide/configuration.md: the resolution paragraph names .vloop/config.local.toml in the order environment, local file, shared file, default; a short section covers the local file (any key but [[check]], config set --local and config path --local, the ignore line, the run's refusal, doctor's warning); the config set --local example is there.
- docs/domain/domain-model.md C-1 and docs/domain/platform/platform-context.md Configuration state the order environment, local file, shared file, default.
- docs/guide/concepts.md names the new .gitignore line where it lists what init writes and what is git-ignored, and its vloop doctor paragraph says the claude, workspace trust, plugin and gate edits rows apply when a session in use runs on Claude and are n/a otherwise.
- README.md: the Copilot section's Overriding locally bullet names .vloop/config.local.toml and vloop config set --local in place of the environment variables; the requirements bullet no longer says vloop run still needs claude on PATH when every session is on Copilot; the Prerequisites' claude and workspace-trust bullets say they apply when a session runs on Claude; the README stays at or under 200 lines (its test, under the go check, passes).
- docs/design-notes/vloop-roadmap.md's Config lives in the repo only line names the local file and its place in the order.
- No passage contradicts the behaviour shipped in T1–T6 (exact messages quoted match the worked example); docs under docs/briefs/ and the other design notes are untouched.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T7/gate.sh
```

</details>

### T8 — Close: the worked example line for line, and every gate at once

`pending` · depends on: T7

Each earlier task was judged on its own; this one proves the brief's worked example holds as written, in its order, with everything in place, and leaves it behind as a committed end-to-end test. It also runs every earlier gate together. A discrepancy it finds is fixed here, in the product, never by changing what the worked example says.

**Acceptance**

- A committed end-to-end test in cmd/vloop walks the worked example's first block in order, in a fresh repository in a temporary directory (vloop init, a [[check]] added): each line's stdout, stderr and exit code exactly as the brief gives them, through the last `✓ config.local`; and config get copilot.credit-usd printing 0.01. The go check runs it.
- Together with the tests T3 and T5 left, every block of the worked example is covered by an end-to-end test: the local-file block, each planted failure (the run refusal included) and the Claude-not-in-use block.
- Every earlier task's gate passes in the same tree.
- Any fix this task makes is to the product or its tests; no expected line of the worked example is changed to match the product.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T8/gate.sh
```

</details>


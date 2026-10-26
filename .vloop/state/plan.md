# Plan — B20261025-2302-b21-records-and-verify

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** complete · **5/5 done** · iteration 5

**Brief:** `docs/briefs/B20261025-2302-b21-records-and-verify.loop-brief.md` · **Updated:** 2026-10-26T02:26:50Z

## Progress

- [x] **T1** — Make a nested git repository its own vloop root (C-2) and have verify make its scratches repositories
- [x] **T2** — Treat a brief's bare name, .loop-brief name and path as one brief, and write the .loop-brief name
- [x] **T3** — Mask the system temp directory as $TMPDIR in everything the driver masks (#53)
- [x] **T4** — Check that a GitHub login reaches Copilot sessions in doctor and the run preflight, and reword the experimental warning
- [x] **T5** — Close the run: every gate at once and the worked example line for line

## Tasks

### T1 — Make a nested git repository its own vloop root (C-2) and have verify make its scratches repositories

`done` · depends on: none

B18's and B19's verify sessions built scratch repositories under .vloop/tmp/ that resolved to this repository's .vloop/, so their `vloop config set` edited the real config and no verify report was written. C-2 changes so the root is the nearest ancestor holding .vloop/ or .git, whichever is found first, level by level; a directory under a vloop repository with no nearer .git or .vloop/ still resolves to the vloop repository. The verify skill then makes every scratch in which an example runs vloop or git a repository with git init, so nothing it runs reaches this repository.

**Acceptance**

- config.Root returns, for a start directory, the nearest ancestor (the start directory included) that holds .vloop or .git, checking both at each level before moving up; else the start directory
- A .git file (a worktree or a submodule) counts as .git
- Violation, ruled against: a root that skips a nearer .git to reach an outer .vloop/; or one that stops at a nearer .git when the start directory itself holds .vloop/
- In a vloop repository, `(cd .vloop/tmp/w && vloop config set language es)` with .vloop/tmp/w a git repository exits 0, writes .vloop/tmp/w/.vloop/config.toml with language = "es", and leaves the outer .vloop/config.toml unchanged
- In the same repository, `(cd .vloop/tmp/plain && vloop config get language)` with no .git in .vloop/tmp/plain prints en, the outer repository's value
- docs/domain/domain-model.md's C-2 reads: The repo root is the nearest ancestor (the start directory included) holding `.vloop/` or `.git`, whichever is found first; else the start directory. Every path vloop prints is relative to it, with `/`.
- docs/domain/platform/platform-context.md's "The repo root" says the same rule in prose and no longer says ".vloop/, else .git"
- plugin/skills/verify/SKILL.md's "Build into a scratch directory" says a scratch directory in which an example runs vloop or git is made a repository first (git init), so it is its own root (C-2) and nothing it runs reaches this repository
- internal/config/root_test.go's nested case is changed to the new answer (a nearer .git wins over an outer .vloop/) and nothing else in it is weakened; the work notes name C-2 as the contract item for that change
- The task leaves committed tests the go check runs in internal/config: a nested .git under an outer .vloop/, a .git file, a start directory holding .vloop/ under an outer .git, both markers in one directory, and a directory with neither under a vloop repository; on Windows-shaped paths where the code handles one

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T1/gate.sh
```

</details>

### T2 — Treat a brief's bare name, .loop-brief name and path as one brief, and write the .loop-brief name

`done` · depends on: none

Interventions recorded with a bare brief name vanish from `vloop metrics`, which counts only records whose brief: field is the .loop-brief name. Every command that takes a brief — metrics, defect and intervention add and list, and intervention set brief — must accept the bare name, the .loop-brief name and a path whose base name is &lt;name&gt;.loop-brief.md as one brief; records are always written with the .loop-brief name, and list and metrics match a record written with either the bare or the .loop-brief name so old records count. No record schema or metrics formula changes, and nothing rewrites existing records.

**Acceptance**

- `vloop metrics`, `vloop defect list --brief`, `vloop intervention list --brief`, `vloop defect add --brief` and `vloop intervention add --brief` accept B20261022-1234-x, B20261022-1234-x.loop-brief and docs/briefs/B20261022-1234-x.loop-brief.md (any path whose base name is &lt;name&gt;.loop-brief.md) as the same brief
- `defect add`, `intervention add` and `intervention set <id> brief <value>` write brief: &lt;name&gt;.loop-brief whichever form they were given
- Violation, ruled against: a record written with a bare or a path brief name
- list and metrics match a record whose brief: field is either the bare or the .loop-brief name; on this repository, `vloop metrics` of B20261002-2135-vloop-v1-security in the three forms gives equal "defects" and "interventions"
- A name that is not a loop brief (for example a .architect-brief.md) or a loop brief that does not exist is refused by defect add as today, writing nothing
- No record schema, metrics formula or existing record changes; no command renames or rewrites existing records' brief names
- The task leaves committed tests the go check runs: the three forms resolving to one brief, add writing the .loop-brief name, and a planted bare-name defect and intervention counted by list and by the metrics

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T2/gate.sh
```

</details>

### T3 — Mask the system temp directory as $TMPDIR in everything the driver masks (#53)

`done` · depends on: none

A plan session's record kept the machine's temp directory (the per-user temp folder on macOS), which is machine-specific and breaks the no-absolute-paths rule (issue #53, defect D20261024-162448). Mask must replace the system temp directory — os.TempDir() and, where it is a symbolic link, its resolved form — with $TMPDIR wherever it masks the home directory today: session records, stderr files, the run log. It runs after the home masking, so a temp directory under the home still comes out as ~/…, and a temp directory of /tmp or /var/tmp is left alone, because briefs and gates name those. No other directory (GOCACHE, XDG_*, caches) is masked.

**Acceptance**

- With TMPDIR a symbolic link, a plan session denied `rm -rf <TMPDIR>/x.1` and `rm -rf <resolved TMPDIR>/y.2` is recorded in sessions/001-plan.json as `rm -rf $TMPDIR/x.1` and `rm -rf $TMPDIR/y.2`, and neither form of the temp directory occurs anywhere in the run folder (session records, stderr files, the run log)
- The temp directory is masked after the home: with TMPDIR under HOME, a denial of &lt;TMPDIR&gt;/x.1 is recorded as ~/tmp/x.1 (for TMPDIR = HOME/tmp)
- Violation, ruled against: /tmp or /var/tmp rewritten to $TMPDIR — with TMPDIR=/tmp a denial of /tmp/x.1 is kept as /tmp/x.1
- Nothing else is newly masked: no GOCACHE, XDG_* or cache path
- The task leaves committed tests the go check runs in internal/driver: the temp directory and its resolved form masked, /tmp and /var/tmp kept, a temp directory under the home giving ~/…, and a Windows-shaped temp directory where the code handles one

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T3/gate.sh
```

</details>

### T4 — Check that a GitHub login reaches Copilot sessions in doctor and the run preflight, and reword the experimental warning

`done` · depends on: none

B19's acceptance found that a Copilot session's own COPILOT_HOME holds no login, so it falls back to the environment and `gh auth token`; an interactive /login does not reach vloop's sessions, and a run only found out after a failed review session. With agent.review = copilot, doctor gains a copilot login row and `vloop run` refuses before planning when neither COPILOT_GITHUB_TOKEN, GH_TOKEN, GITHUB_TOKEN (set and not empty) nor `gh auth token` (exit 0, prints something) provides a login; the token is never printed, logged or kept. The experimental warning stops naming B20, and the guide says how a session logs in. Take gh's and Copilot's behaviour from their docs or from running them, never from memory (the repository's guidelines); no test runs the real copilot, gh or the network.

**Acceptance**

- With agent.review = copilot and no login, `vloop doctor` prints `✗ copilot login no GitHub login reaches Copilot sessions: set COPILOT_GITHUB_TOKEN, GH_TOKEN or GITHUB_TOKEN, or run gh auth login` and exits 1; with a login it prints `✓ copilot login`
- A login is COPILOT_GITHUB_TOKEN, GH_TOKEN or GITHUB_TOKEN set and not empty, or `gh auth token` exiting 0 and printing something; an empty variable or a gh that prints nothing is not a login
- With agent.review = copilot and no login, `vloop run <brief>` exits 1 before planning, starting no session, with `vloop: agent.review is copilot but no GitHub login reaches its sessions: set COPILOT_GITHUB_TOKEN, GH_TOKEN or GITHUB_TOKEN, or run gh auth login` on stderr
- With agent.review = claude, doctor shows no copilot login row and runs no login check
- Violation, ruled against: a token printed, logged or kept in any record, file or message — including in an error that quotes gh's output
- The warning in doctor and the run preflight reads `! agent.review is copilot: experimental — review sessions are not fenced yet (docs/guide/configuration.md)`; violation, ruled against: the old "not fenced until B20" text left anywhere a session or the operator reads it (code, tests, plugin, guide)
- docs/guide/configuration.md's Copilot paragraph says "experimental until the fence lands" and that a Copilot session logs in through COPILOT_GITHUB_TOKEN, GH_TOKEN or GITHUB_TOKEN, else gh (`gh auth login`)
- The earlier tests change only where they pin the old warning text, and only to the new text; the work notes name R4's warning as the contract item for that change
- The task leaves committed tests the go check runs in internal/cli: the copilot login row ✓ and ✗ for each source (each variable, an empty variable, a gh stub that prints a token, exits 1 or prints nothing), the run preflight's refusal, no row with agent.review = claude, and the token absent from the output; the tests are hermetic — a stub gh on PATH and the token variables unset, so a real login on the machine cannot pass them

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T4/gate.sh
```

</details>

### T5 — Close the run: every gate at once and the worked example line for line

`done` · depends on: T1, T2, T3, T4

The tasks were judged one at a time; this proves them together on the finished tree, with the brief's worked example — the root lines, the intervention lines in a scratch repository, the real-data check on B8, the temp directory in a session record and the Copilot login lines, every planted failure included. Any fix found here is made in the product, never in a gate. After it the operator sets the #53 defect (D20261024-162448) to fixed at close.

**Acceptance**

- Every gate of T1 to T4 passes on the finished tree, run together
- In a scratch vloop repository, the worked example's lines run as written: the nested repository's config set leaves the outer config unchanged, the plain directory reads en, and `intervention add "s" --brief B20261022-1234-x …` writes brief: B20261022-1234-x.loop-brief and is listed by the bare name and by the path
- Any fix this task needs is made in the product, not in a gate, and every check passes
- No session sets the #53 defect's status; that is the operator's at close

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T5/gate.sh
```

</details>


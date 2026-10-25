---
name: B20261025-2302-b21-records-and-verify.loop-brief
description: Make a nested git repository its own vloop root, so verify's scratch repositories stop editing the real one; count every record of a brief however it is named; mask the system temp directory in what the driver keeps; and check that Copilot sessions can log in
kind: brief
status: ready
created: 2026-10-25
seeds: The second of two hardening briefs before the fence (docs/design-notes/vloop-roadmap-b18-b21.md → "Before the fence: hardening"), from B18's and B19's verify sessions, B19's acceptance run and issue #53, settled with the operator on 2026-10-25
depends-on: [B20261025-2302-b20-gates-checks-review-scope.loop-brief]
---
# Brief — vloop B21: records, roots and verify

- **Starting point:** extends `main` after B20 is merged. The planner pins the
  base it plans from as the **base** every gate compares against — never
  `HEAD~n`.
- **Produced by:** the design act of 2026-10-25, after B19's close
  (`docs/design-notes/copilot-cli-experiments.md` → "Acceptance on the real
  CLI"; B19's run record).

## Executive summary

B18's and B19's verify sessions both failed the same way: a scratch repository
they built under `.vloop/tmp/` resolved to this repository's `.vloop/`, so
their `vloop config set` lines edited the real config, the driver restored it,
and no verify report was written. B19's acceptance run also showed three
record defects: interventions recorded with a bare brief name vanish from the
metrics, a plan session's record kept the machine's temp directory (#53), and
nothing checks that a Copilot session can log in.

**The goal:** a vloop command run inside a nested repository acts on that
repository alone; every defect and intervention of a brief counts however the
brief was named; no record keeps the machine's temp directory; and a Copilot
review that cannot log in is refused before the run, not after a failed
session.

- **R1 — a nested git repository is its own root.** C-2 stops at the nearest
  ancestor holding `.vloop/` or `.git`, whichever comes first; the verify skill
  makes each scratch it runs vloop or git in a repository.
- **R2 — a brief named any way.** Every command that takes a brief treats the
  bare name, the `.loop-brief` name and the path as one brief, and records are
  written with the `.loop-brief` name (B21-report's F1, pulled forward).
- **R3 — the temp directory masked (#53).** The system temp directory becomes
  `$TMPDIR` in everything the driver masks today.
- **R4 — Copilot's login, checked.** `vloop doctor` and the run's preflight
  check that a GitHub login reaches Copilot sessions; the experimental warning
  stops naming a brief number.

Unchanged: the record schemas, the metrics' formulas, every phase's sessions,
and how a root holding `.vloop/` is found when no nearer repository exists.

## Why this shape, and what was rejected

Decided with the operator on 2026-10-25. **Do not re-litigate these** — the
planner inherits them.

- **The git boundary, not a verify-only fix.** A nested repository is the
  boundary a user expects; today any command in a scratch repository, a test
  fixture or a vendored repository under a vloop repository edits the outer
  one's config and records.
  *Rejected:* keeping C-2 and telling the verify skill to `vloop init` each
  scratch first (the trap stays for everyone else).
- **Records keep the `.loop-brief` name.** It is what every existing record and
  the metrics' key carry; the bare and path forms are read, never written.
- **`$TMPDIR` as the token.** It names what was removed, as `~` does for the
  home directory. `/tmp` and `/var/tmp` stay as they are: they are not
  machine-specific, and briefs and gates name them.
- **Copilot logs in through the environment or `gh`.** B19's acceptance found
  that a session's own `COPILOT_HOME` holds no login, and Copilot then runs
  `gh auth token`; an interactive `/login` does not reach vloop's sessions. The
  check follows the CLI's own order.
  *Rejected:* sharing the user's `~/.copilot` with the session (global state,
  rule 1 of `CLAUDE.md`).

## Binding references

- `docs/domain/domain-model.md` — C-2, which R1 rewrites; M-1 and M-4, the metrics keyed by brief and the defects R2 must count; M-7, no absolute paths in an export
- `docs/domain/platform/platform-context.md` — states C-2's rule in prose, changed with it
- `internal/config/root.go` — the root resolution R1 changes
- `internal/config/root_test.go` — the tests that pin today's C-2; the nested-repository case is the one whose answer changes
- `plugin/skills/verify/SKILL.md` — "Build into a scratch directory", where R1's scratch rule goes
- `internal/defect/defect.go` — `BriefName`, `CheckBrief`, add, set and list: the brief-name match R2 fixes
- `internal/intervention/intervention.go` — add and list by brief, the same match
- `internal/metrics/metrics.go` — where a brief's defects and interventions are gathered
- `internal/driver/session.go` — `Mask`, secrets and the home directory: where R3's temp directory joins them
- `internal/driver/agent_copilot.go` — the Copilot adapter's `Preflight`, which R4 extends
- `internal/cli/doctor.go` — the Copilot checks and the experimental warning R4 changes
- `docs/guide/configuration.md` — the Copilot paragraph: the login, and "experimental until B20"

## Behaviour contract

### R1 — the root

- **C-2** becomes: *The repo root is the nearest ancestor (the start directory
  included) holding `.vloop/` or `.git`, whichever is found first; else the
  start directory. Every path vloop prints is relative to it, with `/`.* A
  `.git` file (a worktree or a submodule) counts as `.git`.
- `docs/domain/platform/platform-context.md` says the same.
- A directory holding both is the root, as today. A start directory under a
  vloop repository with no nearer `.git` or `.vloop/` still resolves to the
  vloop repository.
- The verify skill says: a scratch directory in which an example runs `vloop`
  or `git` is made a repository first (`git init`), so it is its own root
  (C-2) and nothing it runs reaches this repository.

### R2 — a brief named any way

- `vloop metrics <brief>`, `vloop defect list --brief`, `vloop intervention
  list --brief`, `vloop defect add --brief` and `vloop intervention add
  --brief` accept the bare name (`B20261022-1234-url-shortener-api`), the
  `.loop-brief` name and the path `docs/briefs/<name>.loop-brief.md`, and treat
  them as one brief.
- `add` writes `brief: <name>.loop-brief`, whatever form it was given.
- `list` and the metrics match a record whose `brief:` field is either the bare
  or the `.loop-brief` name, so records already written with a bare name count.

### R3 — the temp directory

- `Mask` replaces the system temp directory (`os.TempDir()`, and its resolved
  form where it is a symbolic link, as `/var/folders/…` is of
  `/private/var/folders/…` on macOS) with `$TMPDIR`, wherever it masks the home
  directory today: session records, stderr files, the run log.
- A temp directory of `/tmp` or `/var/tmp` is left as it is.
- The defect record's status becomes `fixed` at close, by the operator.

### R4 — Copilot's login

- With `agent.review = copilot`, a login reaches Copilot sessions when
  `COPILOT_GITHUB_TOKEN`, `GH_TOKEN` or `GITHUB_TOKEN` is set and not empty, or
  when `gh auth token` exits 0 and prints something. The check never prints,
  logs or records the token.
- `vloop doctor` adds a row:

  ```
  ✓ copilot login
  ✗ copilot login  no GitHub login reaches Copilot sessions: set COPILOT_GITHUB_TOKEN, GH_TOKEN or GITHUB_TOKEN, or run gh auth login
  ```

  The ✗ row is a problem (doctor exits 1).
- `vloop run`'s preflight refuses, exit 1, before planning:

  ```
  vloop: agent.review is copilot but no GitHub login reaches its sessions: set COPILOT_GITHUB_TOKEN, GH_TOKEN or GITHUB_TOKEN, or run gh auth login
  ```

- The experimental warning, in doctor and the preflight, becomes

  ```
  ! agent.review is copilot: experimental — review sessions are not fenced yet (docs/guide/configuration.md)
  ```

  and the guide's paragraph says "experimental until the fence lands" and how
  a Copilot session logs in.

### Decided here, because the cited documents leave it underdetermined

- R2's path form is any path whose base name is `<name>.loop-brief.md`; a name
  that is not a loop brief is refused as today.
- R4's check runs only when `agent.review` is `copilot`; with `claude`, doctor
  shows no `copilot login` row.
- R3 masks the temp directory after the home directory, so a temp directory
  under the home directory still comes out as `~/…`.

### Violations the review must rule on

- A root that skips a nearer `.git` to reach an outer `.vloop/`, or one that
  stops at a nearer `.git` when the start directory itself holds `.vloop/`.
- A record written with a bare or path brief name.
- A token printed, logged or kept in any record, file or message.
- `/tmp` rewritten to `$TMPDIR`.
- The old warning text left anywhere a session or the operator reads it.

## Worked example

In a scratch vloop repository (`vloop init` done, committed), with a
`.vloop/tmp/` under it:

```
mkdir -p .vloop/tmp/w && git -C .vloop/tmp/w init -q
(cd .vloop/tmp/w && vloop config set language es)                    -> exit 0
  .vloop/tmp/w/.vloop/config.toml holds language = "es"; git diff -- .vloop/config.toml is empty

mkdir -p .vloop/tmp/plain
(cd .vloop/tmp/plain && vloop config get language)                   -> en   exit 0   (the outer repository: no nearer .git)

vloop intervention add "s" --brief B20261022-1234-x --kind repair --phase halt --by operator --automatable no
  -> .vloop/interventions/I<stamp>-s.md, with  brief: B20261022-1234-x.loop-brief
vloop intervention list --brief B20261022-1234-x                     -> the record
vloop intervention list --brief docs/briefs/B20261022-1234-x.loop-brief.md -> the record
```

On this repository's history (the real-data check), for B8:

```
vloop metrics B20261002-2135-vloop-v1-security --json
vloop metrics B20261002-2135-vloop-v1-security.loop-brief --json
vloop metrics docs/briefs/B20261002-2135-vloop-v1-security.loop-brief.md --json
  -> equal "defects" and "interventions" in all three
```

The gate threads the counts the `.loop-brief` form gives; it does not hardcode
them.

With the run harness and a stub `claude` whose plan session is denied a Bash
command naming a file in the system temp directory:

```
sessions/001-plan.json  "permission_denials": [{"command": "rm -rf $TMPDIR/x.1"...}]
                        and no occurrence of the temp directory's path, in either form
```

With `agent.review = copilot`, a `model.review` set, a stub `copilot`, and an
environment with no token variables:

```
PATH=<a stub gh that exits 1> vloop doctor
  -> ✗ copilot login  no GitHub login reaches Copilot sessions: set COPILOT_GITHUB_TOKEN, GH_TOKEN or GITHUB_TOKEN, or run gh auth login
                                                                     exit 1
PATH=<the same> vloop run <brief>
  -> vloop: agent.review is copilot but no GitHub login reaches its sessions: set COPILOT_GITHUB_TOKEN, GH_TOKEN or GITHUB_TOKEN, or run gh auth login
                                                                     exit 1
PATH=<a stub gh that prints a token> vloop doctor     -> ✓ copilot login, and the token nowhere in the output
GH_TOKEN=<a value> PATH=<the failing stub> vloop doctor -> ✓ copilot login
```

Planted failures the gates must be seen to catch: the nested-repository line
resolving to the outer config (today's behaviour); a bare-name intervention
missing from the metrics (today's); the temp directory kept in a record
(today's); `/tmp/x` in a record rewritten; a token in doctor's output.

## Out of scope

- B20's items: P-8, the gate review's coverage, the check-gap halt, the
  review's scope.
- The fence, hooks, Copilot's sandbox and a second reviewer (the fence brief).
- `sh -c` and `env` forms getting past Copilot's deny list (the fence brief).
- Masking any other directory: `GOCACHE`, `XDG_*` or a cache path.
- A command that renames or rewrites existing records' brief names.
- Changing a record schema, or the metrics' formulas.
- The rest of the report/triage brief (`/vloop:report`, `defect/v2`).

## Constraints

- Go, at the version in `go.mod`; no new dependencies.
- Must keep building and passing on macOS, Linux and Windows; R1's and R3's
  tests cover a Windows-shaped path where the code handles one.
- **No gate runs the real Copilot CLI, `gh` or the network**; gates use stubs.
- **Gates follow the diff surface.** Each task gates on what it touches; no
  task runs the whole test suite.
- **No task may weaken a check to pass.** An earlier test changes only where it
  pins behaviour the contract changes, and only to the new behaviour — the
  nested-repository case of `internal/config/root_test.go`, and the old
  warning text where a test pins it; the work session names the contract item
  in its notes and the review cites it.
- **New tooling is proven by fixtures**, including where it must fail: the
  planted failures above.
- Gates run on macOS's BSD tools: no `grep -P`, no `sed -i` without a suffix, no
  `\+` or `\|` in basic regexes, no `date -d`.
- Tests build fixtures in temporary directories and address them with
  `git -C`; sessions never move refs.
- Repo-relative paths everywhere. No absolute paths in any file or commit
  message.

## Shape

4 to 5 tasks, each verifiable by a single command. Order is load-bearing:

1. **R1 — the root**, C-2 in the domain and platform pages, the verify skill's
   scratch rule. Gated on the worked example's two root lines and the skill's
   rule.
2. **R2 — a brief named any way.** Gated on the intervention lines and the
   real-data check on B8.
3. **R3 — the temp directory.** Gated on the stub plan session's record, `/tmp`
   untouched, and a temp directory under the home directory.
4. **R4 — Copilot's login and the warning.** Gated on the doctor and preflight
   lines, the token kept out of the output, and the warning's new text in every
   place it appears.
5. **Close:** every gate at once and the worked example line for line. The
   operator sets the #53 defect to `fixed`.

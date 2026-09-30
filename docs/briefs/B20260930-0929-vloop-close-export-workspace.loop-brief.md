---
name: B20260930-0929-vloop-close-export-workspace.loop-brief
description: Close a brief with vloop — run record, metrics snapshot, operator findings, consumed status, one commit — and consolidate metrics across repos with a JSON Lines export and a workspace file; finish the guide with concepts, configuration and a generated command reference
kind: brief
status: ready
created: 2026-09-30
seeds: A `.loop/run.sh` plan + run that builds slice B4 in docs/design-notes/vloop-roadmap.md
depends-on: [B20260929-2325-vloop-metrics-defects.loop-brief]
---
# Brief — vloop B4: close, export and workspace

- **Status:** ready to plan
- **Starting point:** extends `main` at the commit that adds this brief. The
  planner pins that SHA as the base every gate compares against.
- **Produced by:** operator decision, 2026-09-30, from the B4 row of
  `docs/design-notes/vloop-roadmap.md` — the second half of the B3 split.

---

## What it is

The fourth slice of `vloop`. At the end of this run `vloop brief close` does in
one command what the operator has done by hand for B1–B3: it records the
operator's pre-merge findings as defects, computes the brief's final metrics,
saves them as a snapshot, writes the brief's run record, marks it consumed, and
commits all of it on the work branch — then says which trailer the squash-merge
needs. `vloop metrics export` writes a repository's briefs, tasks and defects as
JSON Lines that carry numbers and titles, never code, and `--workspace` reads
several repositories' metrics into one table. The guide gains the concepts, a
configuration reference and a command reference generated from the binary.

## Why this shape, and what was rejected

Decided with the operator on 2026-09-29 and 2026-09-30. **Do not re-litigate
these** — the roadmap's `## Metrics and defects` records them in full.

- **Closing is one explicit command**, run by the operator on the work branch
  after verifying the run and before the merge. It is the only command in vloop
  that commits.
  *Rejected:* closing inside `run` — the operator's verification and findings
  come between the run and the close, and B1–B3 each needed that step.
- **Findings must be stated.** `close` refuses unless the operator passes
  `--finding` (repeatable) or `--no-findings`: an unasked question is how a
  finding goes unrecorded.
  *Rejected:* an interactive prompt — vloop commands stay scriptable.
- **The run record is generated between markers.** Anything the operator writes
  outside them survives a re-render; nothing inside them is hand-edited.
- **The snapshot is frozen at close**, in `.vloop/state/metrics/<run id>.json`;
  `vloop metrics` still recomputes live. B6's driver will also write it after
  every iteration.
- **Release stays detected.** `close` does not merge and does not push; it prints
  the `Vloop-Brief:` trailer the squash commit must carry, which is what makes
  `defect add --blame` attribute directly (B3 run record).
- **Consolidation only reads.** Each repository owns its data. The export carries
  the repository's identity and numbers and titles only; a workspace file lists
  clones and lives in a repository of its own. No central service.
  *Rejected:* a dashboard in this slice — it can read the export later.
- **Abandoning is a status**, `abandoned`, not a deletion: an abandoned brief
  keeps its runs and metrics, and briefs that depend on it stay blocked.

## Binding references

- `docs/design-notes/vloop-roadmap.md` — `## Metrics and defects`, in particular
  "Closing and release" and "Across repos"; and the rule that runs never commit
  to the default branch.
- `docs/briefs/B20260929-2325-vloop-metrics-defects.loop-brief.md` — everything
  this slice builds on: `metrics/v1`, `defect/v1`, the summary format, the
  derived defects, release detection, the fixture repository of its worked
  example (reused here). Its run record is the manual close this slice automates.
- `docs/briefs/B20260929-2148-vloop-state-tasks-status.loop-brief.md` — `vloop
  schema` and how a schema joins it, the config table.
- `docs/briefs/B20260929-1804-vloop-skeleton-config-briefs.loop-brief.md` — the
  global conventions, the brief format and statuses, the README test.
- `docs/guide/metrics.md` — a guide page the new pages sit beside; its style is
  the house style.
- `docs/guide/defects.md` — the other existing guide page; the defect fields
  `close` records and the export carries.

## Behaviour contract

B1–B3's conventions hold: global flags, `--json` (one document on stdout, on
exit 0 and 1), exit codes `0` ok / `1` problems or failure / `2` usage, errors
as one stderr line starting `vloop: `, paths repo-relative with `/`.

### A new brief status: `abandoned`

The frontmatter `status` gains `abandoned` (after `consumed`). `brief check`
skips it like `draft` (`- skipped: status is abandoned`); `brief list` shows it,
with `-` in the ready column like `consumed`; a brief depending on an abandoned
one is `blocked`. `task`, `metrics` and `defect` treat it like `consumed`.

### `vloop brief close <brief> (--finding "<summary>"… | --no-findings) [--abandon "<reason>"] [--dry-run]`

**Refusals**, each writing nothing:

| Condition | Message | Exit |
| --- | --- | --- |
| neither `--finding` nor `--no-findings`, or both | `vloop: say what you found before merge: --finding "<summary>" (repeatable) or --no-findings` | 2 |
| the brief has no runs | B3's `vloop: no runs for <name>` | 1 |
| its status is `consumed` or `abandoned` | `vloop: <name> is already <status>` | 1 |
| HEAD is on the default branch | `vloop: close on the brief's work branch, not <branch>` | 1 |
| the working tree has uncommitted changes | `vloop: commit or stash your changes first — close makes one commit of its own` | 1 |
| the last run did not end `complete` and no `--abandon` | `vloop: the plan is not complete (<done>/<planned> done) — finish it, or pass --abandon "<reason>"` | 1 |

**What it does**, in this order, then prints what it did:

1. Each `--finding` becomes a defect file as `vloop defect add` writes it, with
   `--brief` this brief, `found-by: operator`, `origin: work`, `kind: bug`,
   `severity: medium`. Findings needing other fields are recorded with
   `vloop defect add` first, then `close --no-findings`.
2. Computes the brief's metrics exactly as `vloop metrics` does, with the new
   status, and writes them as a `metrics/v1` document to
   `.vloop/state/metrics/<run id>.json` (2-space indented, trailing newline).
3. Writes the brief's run record between
   `<!-- vloop:run-record:begin -->` and `<!-- vloop:run-record:end -->`: a
   `## Run record` heading, the line `Generated by vloop brief close on
   <YYYY-MM-DD>. Recompute with: vloop metrics <name>`, the summary block and
   the `--by task` table each in a fenced block, and the brief's defects as a
   list of `- <id> — <kind>, <origin>, found by <found-by>: <summary>`. If the
   markers exist, only what is between them is replaced; otherwise the block is
   appended at the end of the file. Nothing outside the markers changes.
4. Sets the frontmatter `status` to `consumed` (or `abandoned`). If the body
   holds a `**Status:** ready to plan` line, it becomes
   `**Status:** consumed — closed <YYYY-MM-DD> as run <run id>. **Do not
   re-plan from this brief.**` (`abandoned — <reason>` for an abandon).
5. Commits exactly those files — the defect files, the snapshot and the brief —
   as `[vloop] close <run id>` with a blank line and the trailer
   `Vloop-Brief: <name>`.

Output: the summary block (as `vloop metrics` prints it, now `consumed`), then
one line per file (`recorded <path>`, `wrote <path>`, `updated <path>`),
`committed <sha7> [vloop] close <run id>`, and last
`squash-merge with the trailer: Vloop-Brief: <name>`. `--json`:
`{"brief","status","commit","files":[…],"trailer","metrics":{metrics/v1}}`.

`--dry-run` prints the same lines prefixed `would ` (`would record …`,
`would commit …`), writes nothing and commits nothing, and still applies every
refusal.

### `vloop metrics export [<brief>…]`

JSON Lines on stdout, one record per line, for the named briefs (default: every
brief with runs), in name order: the brief, then its tasks in plan order, then
its defects — derived first by iteration, then recorded by id.

- Every record has `"schema":"export/v1"`, `"type"` (`brief`, `task`,
  `defect`), and `"repo":{"name":…,"remote":…}`.
- `brief`: the brief's `metrics/v1` document without `by_task`, plus `repo`.
- `task`: `brief`, `run_id`, `id`, `title`, `area`, `kind`, `status`,
  `attempts`, lines, agent time, cost and models as in `by_task`.
- `defect`: `brief`, `run_id`, `id`, `task`, `origin`, `found_by`, `kind`,
  `severity`, `status`, `summary`, `derived` (boolean). A derived defect's id is
  `<run id>/i<iteration>-gate` or `<run id>/i<iteration>-review-<n>`.
- **Repository identity.** `remote` is `origin`'s URL with any user, password or
  token removed (`https://user:tok@host/acme/shop.git` →
  `https://host/acme/shop.git`), `null` without an `origin`. `name` is the last
  path segment of the remote without `.git`, else the repository directory's
  name.
- **Nothing else leaves.** No file contents, no paths outside the repository,
  no absolute paths, no environment. Commit shas and repo-relative paths may
  appear.
- `export/v1` joins `vloop schema list`; every exported line validates against
  it.

### `vloop metrics --workspace <file>` and `vloop metrics export --workspace <file>`

A workspace file is TOML:

```
[[repo]]
path = "../shop"          # relative to the workspace file
[[repo]]
path = "../api"
name = "payments-api"     # optional; overrides the repo's own name
```

- `vloop metrics --workspace <file>` prints the cross-brief table with a leading
  `repo` column, one row per brief per repository, repositories in file order.
  Each row equals that repository's own `vloop metrics` row.
- `vloop metrics export --workspace <file>` concatenates each repository's
  export, with `repo.name` overridden where the file sets `name`.
- A listed path that does not exist or is not a git repository:
  `vloop: workspace repo not found: <path as written>` on stderr; the others are
  still reported, and the command exits 1.
- `--workspace` with `<brief>` arguments is exit 2.

### The guide

Three new pages beside `docs/guide/metrics.md` and `docs/guide/defects.md`,
in their style, plus links from the README:

- docs/guide/concepts.md — the loop (plan, then per task work → gate → review
  → commit, each a fresh session), briefs and their lifecycle
  (`draft` → `ready` → `consumed` or `abandoned`), `depends-on`, binding
  references, plans and tasks, gates and the gate shell, metrics and defects in
  one paragraph each, closing and release, the `.vloop/` layout, and the flow
  from `brief new` to `metrics --workspace`.
- docs/guide/configuration.md — every config key with its default, valid
  values, env variable and effect; the stack presets table; the workspace file.
- docs/guide/commands.md — generated from the binary's command tree: every
  command and subcommand with its usage line, short description and flags. It
  says at the top that it is generated and names the one command that
  regenerates it. A test fails when the committed file differs from a fresh
  generation.

### Decided here, because the cited documents leave it underdetermined

- The finding defaults (`origin: work`, `kind: bug`, `severity: medium`) — the
  common case; anything else goes through `defect add` first.
- `close` refuses on the default branch and on a dirty tree, because it commits.
- The body `**Status:**` line is updated too, so a brief closed by vloop is also
  refused by the shell loop's planner while both exist.
- Derived defect ids in the export, since B3 never had to name them.
- Credentials are stripped from the remote URL; the export never needs them.

### Violations the review must rule on

- `close` changing any file other than the defect files, the snapshot and the
  brief; changing the brief outside the markers, the frontmatter `status` and
  the body `**Status:**` line; or committing anything else.
- `close` merging, pushing, or creating or moving any branch or ref other than
  the one commit on the current branch.
- An export record carrying file contents, an absolute path, a credential, or a
  field not in `export/v1`.
- A workspace or export that writes anything.
- A generated command reference edited by hand, or a staleness test that
  regenerates the file instead of failing.
- B5 or B6 work appearing here: `init`, `upgrade`, `doctor`, skills, `run`.

## Worked example

B3's worked-example fixture, before the merge: HEAD on the work branch
`B20260101-0900-a`, the run complete, the brief `status: ready` with a
`**Status:** ready to plan` line, the working tree clean, `origin` set to
`https://user:tok@example.com/acme/shop.git`. `<stamp>`, `<sha7>` and dates are
threaded from what the commands print.

```
$ vloop brief close B20260101-0900-a.loop-brief
vloop: say what you found before merge: --finding "<summary>" (repeatable) or --no-findings
                                                               exit 2
$ vloop brief close B20260101-0900-a.loop-brief --finding "README omits the flag" --dry-run
  (the summary block)
would record .vloop/defects/D<stamp>-readme-omits-the-flag.md
would write .vloop/state/metrics/B20260101-0900-a.json
would update docs/briefs/B20260101-0900-a.loop-brief.md
would commit [vloop] close B20260101-0900-a                    exit 0
  and git status --porcelain prints nothing

$ vloop brief close B20260101-0900-a.loop-brief --finding "README omits the flag"
B20260101-0900-a  consumed · not merged
 tasks     2 planned (brief said 2–3) · 2 done · 0 blocked · first-pass 1/2
  (… as B3's summary …)
 defects   in-loop 1 · operator 1 · escaped 0 · removal efficiency 100%
recorded .vloop/defects/D<stamp>-readme-omits-the-flag.md
wrote .vloop/state/metrics/B20260101-0900-a.json
updated docs/briefs/B20260101-0900-a.loop-brief.md
committed <sha7> [vloop] close B20260101-0900-a
squash-merge with the trailer: Vloop-Brief: B20260101-0900-a.loop-brief
                                                               exit 0
$ git log -1 --format=%B | tail -1
Vloop-Brief: B20260101-0900-a.loop-brief
$ git show --name-only --format= HEAD | sort
.vloop/defects/D<stamp>-readme-omits-the-flag.md
.vloop/state/metrics/B20260101-0900-a.json
docs/briefs/B20260101-0900-a.loop-brief.md
$ vloop schema validate metrics/v1 .vloop/state/metrics/B20260101-0900-a.json
.vloop/state/metrics/B20260101-0900-a.json: ok                 exit 0
$ vloop brief close B20260101-0900-a.loop-brief --no-findings
vloop: B20260101-0900-a.loop-brief is already consumed         exit 1

$ vloop metrics export | jq -c '{type, id: (.run_id // .id)}'   (after the close)
{"type":"brief","id":"B20260101-0900-a"}
{"type":"task","id":"T1"}
{"type":"task","id":"T2"}
{"type":"defect","id":"B20260101-0900-a/i2-gate"}
{"type":"defect","id":"D<stamp>-readme-omits-the-flag"}         exit 0
$ vloop metrics export | head -1 | jq -c .repo
{"name":"shop","remote":"https://example.com/acme/shop.git"}
```

`--workspace`, with a second fixture repository `api` (another brief of its own)
and portfolio/workspace.toml listing `../shop` and `../api`:

```
$ vloop metrics --workspace portfolio/workspace.toml
repo  brief             tasks  first-pass  …
shop  B20260101-0900-a  2      1/2         …
api   <api's brief>     …                                       exit 0
  each row after `repo` equal to that repo's own `vloop metrics` row
```

The failures, each planted in a copy of the fixture:

```
close run with HEAD on main
  -> vloop: close on the brief's work branch, not main                  exit 1
an untracked file in the working tree
  -> vloop: commit or stash your changes first — close makes one commit of its own
                                                                        exit 1
the last run ended stalled with T2 pending
  -> vloop: the plan is not complete (1/2 done) — finish it, or pass --abandon "<reason>"
                                                                        exit 1
   and with --abandon "superseded" --no-findings: status abandoned,
   the body line "**Status:** abandoned — superseded", exit 0;
   brief check skips it; a brief depending on it lists as blocked
a hand-written paragraph after the run-record markers, then a re-render
  -> the paragraph is unchanged, byte for byte
the workspace lists ../missing
  -> vloop: workspace repo not found: ../missing (the others still print) exit 1
```

**Real data.** On this repository, read-only: `vloop metrics export` prints a
`brief` record for each of B1–B3 whose numbers equal `vloop metrics --json`, and
every line validates against `export/v1`; its `repo` is
`{"name":"vloop","remote":"https://github.com/mvelosop/vloop.git"}`. A workspace
file listing this repository reproduces `vloop metrics` with a `repo` column.

## Out of scope

The roadmap's later rows:

- `init`, `upgrade`, `doctor`, plugin extraction, skills (B5).
- `run`, the per-iteration snapshot, `[vloop]` run commits, the driver's refs
  check (B6).

Also not in this run:

- Merging, pushing, opening a PR, or setting the squash message — `close` prints
  the trailer; the operator merges.
- Closing the briefs B1–B3 retroactively, or rewriting their hand-written run
  records into markers.
- A dashboard, a hosted service, re-pricing, or any network access.
- Interactive prompts.
- Deleting or rewriting defect files beyond what `close` records.

## Constraints

- Go as pinned in `go.mod`. Third-party modules: B3's; anything else needs its
  reason in the task notes. No cgo. vloop never uses the network. Git is run as
  the `git` command.
- Must build for `darwin`, `linux` and `windows`. Nothing may assume `/` as the
  on-disk separator or a case-sensitive filesystem.
- **Gates run on macOS with its BSD tools**: no GNU-only syntax (no empty
  alternatives in `grep -E`, no `sed -i` without a suffix, no `\+` or `\|` in
  basic regexes, no `date -d`).
- **Sessions must not move git refs.** The fence denies `git branch`,
  `checkout`, `switch`, `remote`, `update-ref`, `symbolic-ref`, `tag`, `stash`
  and `rebase` to sessions, and the driver halts (exit 9) if any ref moves
  during a session — B3's planning session renamed this repo's `main` while
  checking its gate fixtures. Gates and tests build fixture repositories in
  temporary directories and address them with `git -C <dir>`; a session checks a
  gate by running the gate, never by pasting its git commands into its shell.
- **Gates follow the diff surface**: each task gates on the packages it touches,
  plus `go vet ./...`, `gofmt -l .` and `go mod tidy -diff` printing nothing,
  and the linux, windows and host builds. Only the close runs `go test ./...`.
- A gate runs in under a minute. Tests never read or write this repo's `.loop/`,
  `.claude/`, `.vloop/`, git history or the home directory; only the close task's
  real-data check reads this repository, read-only.
- **No task may weaken a check to pass.** Two expectations must change, and
  changing them is required: `schema list` gains `export/v1`, and the brief
  statuses gain `abandoned`. Update exactly those.
- **New checks are proven by fixtures**: every refusal of `close`, the marker
  re-render, credential stripping, and the workspace's missing repo each have a
  fixture where they apply and one where they must not.
- Repo-relative paths everywhere. No absolute paths in any file or commit message.

## Shape

8 to 10 tasks, each independently verifiable by a single command. Order is
load-bearing:

1. **`export/v1`** in the schemas and the **`abandoned`** status across `brief
   check`, `brief list` and `depends-on`, with the two expectation updates.
2. **Rendering**: the run record between markers (append, replace, preserve
   outside) and the metrics snapshot file, on fixture briefs.
3. **`vloop brief close`**: every refusal, the findings, the frontmatter and body
   status, the one commit with its trailer, the output and `--json`.
4. **`--abandon` and `--dry-run`.**
5. **`vloop metrics export`**: records, order, derived defect ids, repository
   identity and credential stripping, every line valid `export/v1`.
6. **`--workspace`** for `metrics` and `export`, with the missing-repo case.
7. **The guide**: concepts.md, configuration.md (a test that every config
   key and every preset appears), commands.md generated with its staleness
   test, and README links; B1's README test passes unmodified.
8. **Close**: an end-to-end test that builds the binary and plays the worked
   example line for line on fixtures; the real-data check on this repository
   (read-only); `go test ./...`, `go vet ./...`, `gofmt -l .`,
   `go mod tidy -diff` and the three builds.

The worked example needs its own gate (task 8): `close` is the first vloop
command that commits, and only a built binary in a real fixture repository shows
that it commits exactly the three files, on the right branch, with the trailer.

---
name: B20261126-1843-b47-i160-vloop-folder-readme.loop-brief
description: Have vloop init write, and vloop upgrade rewrite or restore, a .vloop/README.md that vloop owns, telling anyone who opens the folder what vloop is, how to use it, what the folder holds, and where vloop lives
kind: brief
status: consumed
created: 2026-11-26
seeds: Issue #160 (a README in .vloop/), decided with the operator on 2026-11-26 — the file is wholly vloop's, a summary of vloop, how to use it and a link to the repository, kept current by vloop upgrade, which restores it whenever it is missing
depends-on: [B20261126-1418-b45-docs-mixed-os-and-any-agent.loop-brief]
---
# Brief — vloop B47: a README that vloop owns in .vloop/

- **Starting point:** extends `main` as it stands when the run starts, with
  this brief's ready commit on the run branch. The planner pins that SHA as
  the **base** every gate compares against — never `HEAD~n`.
- **Produced by:** issue #160; the operator decided on 2026-11-26 that vloop
  owns the whole file (a summary of what vloop is, how to use it and a link
  to the repository), that it is kept current when vloop is updated, and that
  `vloop upgrade` restores it whenever it is missing. The rest is decided by
  the operator's hands, recorded below.
- **Review:** `vloop brief review` (Copilot, mai-code-2), five rounds, the last
  `PASS`. What changed: `<version>` read from `vloop version`; where `init`
  prints the new line, also with an existing guidelines file; the language
  for `init` and for `upgrade`; the dry run at the same version; the binding
  reasons and the account of B43's and B44's Mac-only tests.
- **Roadmap:** the row for B47 in `docs/design-notes/vloop-roadmap-b18-b21.md`,
  pointing at this brief, is added in its ready commit.
- **Run environment:** planned and run on Windows, in pwsh, every session on
  GitHub Copilot (`.vloop/config.local.toml`). See Constraints.

## Executive summary

Anyone who opens `.vloop/` in a repository that uses vloop, on GitHub or in a
clone, finds `config.toml`, `install.json`, `repo-guidelines.md`, `defects/`,
`interventions/` and `state/`, and nothing that says what vloop is or where to
get it. GitHub shows a folder's `README.md` under its file list.

**The goal:** every repository vloop sets up carries a `.vloop/README.md` that
explains vloop and the folder, and stays current with the vloop in use.

- **R1 — the file.** `vloop init` writes `.vloop/README.md`: what vloop is,
  how to use it, what each entry of the folder is, and the link to the
  repository, in the repository's `language`.
- **R2 — kept current.** `vloop upgrade` rewrites it whenever it differs from
  what this vloop writes, and restores it whenever it is missing, also when
  the repository is already at this version.
- **R3 — this repository** gets the file when the operator runs `vloop
  upgrade` after the merge.

Unchanged: every other file `init` and `upgrade` write, and what they print
for them.

## What it is

One generated document and the two commands that write it. It is for the
visitor who finds `.vloop/` and wants to know what it is, and it shows where
vloop lives.

## Why this shape, and what was rejected

Decided with the operator on 2026-11-26. **Do not re-litigate these.**

- **vloop owns the whole file.** It is regenerated, not merged: its first
  lines say so, and edits to it are lost at the next `vloop upgrade`.
  *Rejected:* a marked block with room for the team's own text (the operator
  wants it wholly vloop's); writing it once (it would go stale).
- **`vloop upgrade` keeps it current and restores it.** `vloop self-update`
  never touches a repository (B44); its last line already tells the operator
  to run `vloop upgrade` in each repository, which rewrites the file. A
  deleted file comes back at the next `vloop upgrade`. *Rejected:* an opt-out
  key (the operator chose to restore it always); `self-update` writing into
  repositories (B44 decided it does not).
- **Descriptive, not an advertisement.** It says what vloop does and where it
  is, in a few hundred words, with no claims of quality or comparisons.
- **In the repository's `language`** (`en` or `es`), like the guidelines file
  and the starter brief `init` writes.

## Binding references

- `docs/domain/domain-model.md` — C-4 (`language` selects the prose vloop writes) and S-4 (sessions write nothing under `.vloop/` but `.vloop/tmp`, so the run cannot write this repository's own file)
- `docs/domain/platform/platform-context.md` — the language setting (C-4) the file follows, and the repository root and paths vloop prints
- `docs/guide/concepts.md` — "Setting up a repository": what `init` and `upgrade` write, which R1 and R2 extend
- `internal/cli/init.go` — the steps `init` takes, among them the guidelines file in the repository's language
- `internal/cli/upgrade.go` — the changes `upgrade` makes, its `--dry-run`, its `already at <version>` and its breaking-change gate

## Behaviour contract

### R1 — the file

`.vloop/README.md`, in English (`language = "en"`):

```
# vloop

Written by vloop <version>. vloop owns this file: `vloop upgrade` rewrites it, so edits here are lost.

## What vloop is
…
## How to use it
…
## What is in this folder
…
## More
https://github.com/mvelosop/vloop
```

With `language = "es"`, the headings are `## Qué es vloop`, `## Cómo se usa`,
`## Qué hay en esta carpeta` and `## Más`, and the third line reads
`Escrito por vloop <version>. Este archivo es de vloop: vloop upgrade lo reescribe, así que lo que se edite aquí se pierde.`
`<version>` is the binary's version as `vloop version` prints it after
`vloop ` (`0.15.0-rc.4`); a gate reads it from `vloop version`, never
hardcodes it.

What the sections say, in at most 400 words in all:

- **What vloop is:** a command-line tool and a plugin for Claude Code and
  GitHub Copilot CLI that turns a brief into tasks and, for each, has a work
  session do it, a gate check it and an independent review judge it, with one
  commit per task; it runs on macOS, Linux and Windows.
- **How to use it:** install with `go install
  github.com/mvelosop/vloop/cmd/vloop@latest`; `vloop doctor` checks the
  setup; write a brief (with `/vloop:create-brief` in the agent) and run it
  with `vloop run <brief>`; keep vloop current with `vloop self-update`, then
  `vloop upgrade` in each repository.
- **What is in this folder:** one line each for `config.toml`,
  `config.local.toml` (this machine's, git-ignored), `install.json`,
  `repo-guidelines.md`, `README.md`, `defects/`, `interventions/`, `state/`
  (the plan, the runs, the metrics and the brief reviews) and `tmp/`
  (scratch, git-ignored).
- **More:** the repository's URL, `https://github.com/mvelosop/vloop`.

`vloop init` writes it with its other files and prints `wrote
.vloop/README.md` right after the install stamp's line and the guidelines
file's line when there is one (an existing guidelines file is kept and prints
none), and before the starter brief's. An existing `.vloop/README.md` is
overwritten: vloop owns it.

### R2 — kept current

`vloop upgrade` compares `.vloop/README.md` with what this binary writes:

- **missing or different:** it writes it and prints
  `updated .vloop/README.md`; `--dry-run` prints `would update
  .vloop/README.md` and writes nothing; a breaking upgrade without `--yes`
  lists it among the changes, as for `CLAUDE.md`;
- **the same:** nothing is printed for it;
- **at the same version** (today `already at <version>` and nothing else):
  when the file is missing or different it is written first, printing
  `updated .vloop/README.md`, then `already at <version>`; with `--dry-run`,
  `would update .vloop/README.md` then `already at <version>`, writing
  nothing; otherwise only `already at <version>`, as today.

### R3 — this repository

S-4 keeps sessions from writing under `.vloop/`, so this repository's own
`.vloop/README.md` is written by the operator, with `vloop upgrade` after the
merge, and committed then.

### Decided here, because the cited documents leave it underdetermined

- **Which `language`:** `vloop init` uses the language it uses for its other
  files (its `--language` when given, as for the guidelines file);
  `vloop upgrade` uses the one `vloop config get language` resolves to, `en`
  when unset.
- **Line ends** are `\n`, as for every file vloop writes.
- **The guide** (`docs/guide/concepts.md`, "Setting up a repository") gains a
  sentence: `init` writes `.vloop/README.md`, and `upgrade` rewrites or
  restores it.

### Violations the review must rule on

- Any change to another file `init` or `upgrade` writes, or to their output
  for it.
- A README whose text differs between `init` and `upgrade` for the same
  version and language.
- Claims beyond the four sections, or over 400 words.
- An earlier test edited for anything but the new file in `init`'s and
  `upgrade`'s output.

## Worked example

In a scratch repository (`git init`), with the built CLI, in pwsh:

```
vloop init
  -> its usual lines, with `wrote .vloop/README.md` after the stamp's and
     the guidelines file's lines, before the starter brief's     exit 0
Get-Content .vloop/README.md -TotalCount 3
  -> # vloop
     (an empty line)
     Written by vloop <version>. vloop owns this file: `vloop upgrade` rewrites it, so edits here are lost.

Remove-Item .vloop/README.md; vloop upgrade
  -> updated .vloop/README.md
     already at <version>                                         exit 0

Add-Content .vloop/README.md 'local edit'; vloop upgrade --dry-run
  -> would update .vloop/README.md
     already at <version>                                         exit 0
     (the file still ends with 'local edit')
vloop upgrade
  -> updated .vloop/README.md
     already at <version>                                         exit 0
     (the file no longer contains 'local edit')

vloop upgrade
  -> already at <version>                                         exit 0
```

With `language = "es"` set before `vloop init`, the file's headings are the
Spanish ones and it contains `https://github.com/mvelosop/vloop`.

## Acceptance, at close (the operator's)

In this repository, after the merge: `vloop upgrade` prints `updated
.vloop/README.md`; the file's sections match R1; the operator commits it.

## Out of scope

- A doctor row for the file.
- An opt-out, or room for the team's own text in the file.
- Writing into repositories from `vloop self-update`.
- Translating anything else.

## Constraints

- Go, as pinned in `go.mod`; `go run ./tools/check` is the `[[check]]` and
  passes after every task.
- **This run is on Windows**, in pwsh. Gate through the built CLI in a
  temporary repository or through oracles in the gate folder, never through
  the `cmd/vloop` end-to-end tests that skip on Windows. A gate that runs
  `vloop` sets what it depends on through the environment, since this
  repository's `.vloop/config.local.toml` applies to it.
- **Earlier tests:** `init` and `upgrade` gain a line of output, so tests that
  pin their exact output change, and only by that line. Look for them also
  among the tests that skip on Windows (`cmd/vloop`, and some in
  `internal/cli`): B43 and B44 each changed output that such a test pinned
  (pwsh's gate arguments, doctor's plugin warning), and only macOS found it.
- **Gates follow the diff surface.** No gate runs the whole suite or the
  check (P-8).
- **No task may weaken a check to pass.**
- Tests build their repositories in temporary directories and address them
  with `git -C`. Sessions never move refs. Repo-relative paths everywhere.

## Shape

3 to 4 tasks:

1. **R1 in `vloop init`**, both languages, with the earlier tests that pin
   `init`'s output updated. Gated on the worked example's `init` lines and
   the Spanish headings.
2. **R2 in `vloop upgrade`**, with the earlier tests that pin its output
   updated. Gated on the restore, the overwrite, `--dry-run` and the
   same-version lines.
3. **The guide sentence.** Gated on it being present.
4. **Close:** every earlier gate at once, and the worked example line by line
   through the built CLI in a temporary repository, as a gate.
<!-- vloop:run-record:begin -->
## Run record

Generated by vloop brief close on 2026-11-27. Recompute with: vloop metrics B20261126-1843-b47-i160-vloop-folder-readme.loop-brief

```
B20261126-1843-b47-i160-vloop-folder-readme  consumed · not merged
 tasks     4 planned (brief said 3–4) · 4 done · 0 blocked · first-pass 4/4
 size      delivered  code 79 · test 122 · docs 1 · test:code 1.54
           churn      code 79 · test 122 · docs 1 · rework 1.00
 time      agent 13.2 min (work 9.3 · review 3.9) · plan 4.9 min · gate review 1.8 min · verify 2.9 min · gates 0.1 · wall 32.2 min
 rate      6.0 code lines/min · 15.2 incl. tests
 cost      ≈$0.63 · plan 0.12 · work 0.21 · review 0.15 · gate review 0.05 · verify 0.10 · $7.95/kloc
 models    plan mai-code-2 · work mai-code-2 · review mai-code-2 · gate review mai-code-2 · verify mai-code-2
 defects   in-loop 0 · operator 3 · escaped 0 · removal efficiency 100%
 interventions  3 · recommended 0 · other-option 0 · adjusted 0 · different 0 · no-options 3
 instructions  plan 4,884 · gate review 2,423 · work 2,490 · review 2,511 · verify 2,164 words · guidelines 2 rules
```

```
id  area  kind     att  code+  test+  docs+  other+  test:code  $/kloc  agent  cost   model
T1  -     feature  1    50     46     0      0       0.92       2.40    4m51s  $0.12  mai-code-2
T2  -     feature  1    29     76     0      0       2.62       3.36    4m53s  $0.10  mai-code-2
T3  -     docs     1    0      0      1      0       -          -       1m36s  $0.05  mai-code-2
T4  -     test     1    0      0      0      0       -          -       1m52s  $0.09  mai-code-2
```

- I20261127-202947-verify-found-the-guide-s-vloop-layout-li — verification-finding, work, found by operator: verify found the guide's .vloop/ layout list omitted .vloop/README.md; fixed by hand
- I20261127-202948-verify-noted-upgrade-s-help-still-says-i — verification-finding, brief, found by operator: verify noted upgrade's help still says it refreshes a repository set up by an older vloop, while it now also restores .vloop/README.md at the same version; left as is, the README refresh is incidental
- I20261127-202949-the-verify-fence-denies-vloop-init-so-al — verification-finding, env, found by operator: the verify fence denies vloop init, so all six worked-example lines went unrun in verify; the operator ran them through the built CLI on Windows, both languages, and they matched
<!-- vloop:run-record:end -->

---
name: vloop-operator
description: The operator's playbook for this repo — /vloop:operate, plus what this repository adds to it (where its rules live, its checks, the roadmap row at close, the release, the acceptance briefs). Use in an interactive session whenever the operator asks to run, resume, verify, close or merge a brief, cut a release, or what to do about a halt or a finding. Writing briefs is the vloop-architect skill's.
---

# Operating the loop here

**Follow `/vloop:operate`** — the plugin's playbook — for everything: before the
run, halts, verifying, recording, closing and merging. If it is not loaded, read
`plugin/skills/operate/SKILL.md`, which is the same text. This skill only adds
what is particular to this repository; where the two seem to disagree, say so to
the operator rather than picking one.

## Where the rules live

`CLAUDE.md` points to them. When a failure needs the rule it breaks, walk the
domain (`docs/domain/README-domain.md` → `domain-model.md` → the context and
entity page) and **cite the invariant by ID** (`B-4`, `P-4`, `S-2`, `M-5`…) in
what you report and record. If the domain is silent, propose the rule to the
operator; it may belong in `domain-model.md`.

## Running

- The run id is the brief's file name minus `.loop-brief`; the work branch is
  named for it.
- Gates run on macOS's BSD tools: no empty alternatives in `grep -E`, no
  `sed -i` without a suffix, no `\+`/`\|` in basic regexes, no `date -d`. A
  gate that breaks one of these is a gate defect.
- VS Code's `branch.*.vscode-merge-base` and its pull-request extension's
  `branch.*.github-pr-*` keys are ignored by the guard; any other editor key
  still halts with exit 9.

## Verifying

- The repository's checks are the ones in `CLAUDE.md` → Toolchain, all of them,
  run by you.
- Before trying `vloop brief close` or `vloop defect add` for real in a
  throwaway clone of this repo, run `git remote set-head origin main` in the
  clone: a local clone takes whatever branch was checked out as its default.

## Closing

- Besides what `vloop brief close` does, mark the brief's row in
  `docs/design-notes/vloop-roadmap.md` consumed.
- After recording interventions, refresh their index:
  `tools/interventions-index.sh --write` (`--check` when verifying).

## Releasing

Only when the operator asks for one: `go run ./tools/release prepare <version>`
opens the release PR; merge it on the operator's go-ahead; then
`go run ./tools/release publish <version>` tags, releases and proves the public
install. Record a `ceremony` intervention for it. Then install the release and
`vloop upgrade` in each repository that uses vloop. See `tools/release/README.md`.

## The acceptance briefs' flaws stay where they are

A brief used for acceptance runs (the URL shortener template's) can hold a flaw
found in an earlier run, kept on purpose to see how the loop and the agents
handle it. Don't fix it, and don't describe it anywhere a session or a later fix
could learn it: issues, skills, evals, anything under `plugin/`. Record what the
run showed in the operator notes and the design notes, in words that don't give
the flaw away; a fix to vloop uses a synthetic case.

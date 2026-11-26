# Plan — B20261126-1843-b47-i160-vloop-folder-readme

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **4/4 done** · iteration 4

**Brief:** `docs/briefs/B20261126-1843-b47-i160-vloop-folder-readme.loop-brief.md` · **Updated:** 2026-11-26T21:08:01Z

## Progress

- [x] **T1** — Generate the owned README during init
- [x] **T2** — Reconcile the owned README during upgrade
- [x] **T3** — Document the owned README lifecycle
- [x] **T4** — Verify the README workflow end to end

## Tasks

### T1 — Generate the owned README during init

`done` · depends on: none

Make every newly initialized repository explain vloop and its .vloop folder in the repository's language. This establishes the generated file and output contract that later upgrades can reproduce and keep current.

**Acceptance**

- vloop init writes .vloop/README.md in English or Spanish according to the same language selection as the other generated prose, using the binary's vloop version and LF line endings.
- The README has only the four brief-named sections, stays within 400 words, describes vloop's workflow and three supported operating systems without quality claims or comparisons, gives the install, doctor, brief and maintenance commands, lists every named .vloop entry on its own line, and links to https://github.com/mvelosop/vloop.
- The file begins with the ownership warning in the selected language, and init overwrites an existing .vloop/README.md rather than merging or keeping local edits.
- Init prints wrote .vloop/README.md after the install stamp and any newly written guidelines line, and before the starter brief; existing guidelines remain silent, and no other init file or output changes.
- Committed Go tests exercise both languages, version insertion, the folder inventory, overwrite behavior, output ordering, and the existing-guidelines case, including any earlier exact-output assertions that only gain the README line.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T1/gate.ps1
```

</details>

### T2 — Reconcile the owned README during upgrade

`done` · depends on: T1

Keep an initialized repository's README identical to what the current binary generates, even when the install version is already current. This adds restoration and dry-run reporting without changing upgrade's existing file or breaking-change behavior.

**Acceptance**

- At a normal, breaking, or same-version upgrade, a missing or non-identical .vloop/README.md is generated in the configured language using the current binary version; an identical file causes no README line.
- A real same-version repair prints updated .vloop/README.md before already at &lt;version&gt;; --dry-run prints would update .vloop/README.md before already at &lt;version&gt; and leaves the file byte-for-byte unchanged.
- A breaking upgrade without --yes includes .vloop/README.md in the proposed changes when it needs repair and writes nothing; --yes or a normal upgrade applies the repair alongside the existing upgrade writes.
- For the same language and version, init and upgrade produce byte-identical README content, and no other upgrade file's content or output changes.
- Committed Go tests cover missing, edited, current, dry-run, same-version, and breaking-upgrade cases in both language resolution and exact-output contexts, including earlier tests changed only by the new README line.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T2/gate.ps1
```

</details>

### T3 — Document the owned README lifecycle

`done` · depends on: T1, T2

Make the setup guide tell repository users that init creates the folder README and upgrade refreshes or restores it. This keeps the documented setup and lifecycle aligned with the two command changes.

**Acceptance**

- The Setting up a repository section says: `init` writes `.vloop/README.md`, and `upgrade` rewrites or restores it.
- The guide change is limited to that lifecycle sentence and does not alter the documented behavior of any other file init or upgrade writes.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T3/gate.ps1
```

</details>

### T4 — Verify the README workflow end to end

`done` · depends on: T1, T2, T3

Prove that generation, restoration, dry-run behavior and documentation compose into the operator workflow specified by the brief. This closes the plan only when the built CLI replays every worked-example line in a temporary repository.

**Acceptance**

- Every earlier task's planner-owned gate passes against the completed tree, and the close gate builds the CLI and runs the brief's worked example line by line in a temporary git repository.
- The workflow proves init output ordering, the version-bearing owned README, missing-file restoration, non-writing dry-run, overwrite of a local edit, and silence for a current file at the same version.
- No gate creates or changes a file in the repository, and the temporary repository is removed after the workflow whether a command passes or fails.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T4/gate.ps1
```

</details>


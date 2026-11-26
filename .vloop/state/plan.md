# Plan — B20261126-1034-b43-i162-pwsh-gates-utf8

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** complete · **3/3 done** · iteration 4

**Brief:** `docs/briefs/B20261126-1034-b43-i162-pwsh-gates-utf8.loop-brief.md` · **Updated:** 2026-11-26T14:11:30Z

## Progress

- [x] **T1** — Make PowerShell gate processes UTF-8
- [x] **T2** — Document UTF-8 behavior for PowerShell users
- [x] **T3** — Verify PowerShell UTF-8 through the built CLI

## Tasks

### T1 — Make PowerShell gate processes UTF-8

`done` · depends on: none

Set UTF-8 encodings inside every pwsh and powershell process that the shared gate runner starts. This gives gates and checks correct native-input and native-output decoding without changing their recorded commands or exit codes.

**Acceptance**

- GateCommand configures [Console]::OutputEncoding, [Console]::InputEncoding, and $OutputEncoding as BOM-free UTF-8 before the recorded command runs for both pwsh and powershell, while attempting setup failures silently and continuing to the recorded command.
- The shared runner preserves the recorded command's stdout, stderr, and exit code; an exit 3 remains 3 and the setup itself prints nothing.
- Native git output containing café reaches a PowerShell Select-String assertion intact, and café sent through a PowerShell pipeline to git hash-object has the expected UTF-8 bytes and platform line ending.
- The invocation arguments for sh, bash, and cmd are unchanged, and shown, recorded, or logged gate command text contains only the recorded command, never the encoding setup.
- Committed tests run by the go check cover native output, native stdin, emitted bytes, all three BOM-free encodings, setup silence, exit-code preservation, and the unchanged non-PowerShell argument lists; the planner-owned oracle is not used as that test suite.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T1/gate.ps1
```

</details>

### T2 — Document UTF-8 behavior for PowerShell users

`done` · depends on: none

Explain where vloop supplies UTF-8 and where an interactive PowerShell terminal still needs profile setup. The distinction prevents the new runner behavior from being mistaken for a machine-wide encoding change.

**Acceptance**

- The Mixed teams section says vloop runs pwsh gates and checks with UTF-8 input and output so native tool output remains correct on Windows.
- The paragraph distinguishes vloop's behavior from a person's own PowerShell terminal and includes the exact profile line [Console]::OutputEncoding = [System.Text.UTF8Encoding]::new().
- The paragraph explains that the profile line makes vloop run output remain correct through Tee-Object and similar terminal pipes, without claiming that vloop changes the profile or machine code page.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T2/gate.ps1
```

</details>

### T3 — Verify PowerShell UTF-8 through the built CLI

`done` · depends on: T1, T2

Prove the completed runner and documentation together through the built vloop CLI in an isolated repository. This catches integration breaks that either task's focused oracle could miss.

**Acceptance**

- A built CLI in a temporary git repository runs the worked example's native git read and café output gates successfully and reports the exit-3 gate as a failure while vloop task gate itself exits 1.
- vloop task show prints each recorded verify command unchanged, including the café assertion, and never prints the implementation's encoding setup.
- The close gate runs the T1 and T2 oracles together, uses only temporary repositories and build outputs, and leaves the repository tree unchanged.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T3/gate.ps1
```

</details>


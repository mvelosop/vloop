# Plan — B20261126-1036-b44-i98-self-update

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **0/5 done** · iteration 3

**Brief:** `docs/briefs/B20261126-1036-b44-i98-self-update.loop-brief.md` · **Updated:** 2026-11-26T15:05:04Z

## Progress

- [ ] **T1** — Implement self-update target decisions and check mode · 2 attempt(s)
- [ ] **T2** — Install self-update targets safely
- [ ] **T3** — Update installed agent plugins
- [ ] **T4** — Point existing commands and guides to self-update
- [ ] **T5** — Verify the complete self-update workflow

## Tasks

### T1 — Implement self-update target decisions and check mode

`pending` · 2 attempt(s) · depends on: none

Introduce the self-update command so a released vloop can safely choose and report its target from any directory. This establishes the refusal, version-selection and machine-readable contract that installation and plugin updates build on.

**Acceptance**

- vloop self-update is a top-level command runnable without a repository and accepts only the exact --check and --version &lt;v&gt; flag contract in the brief.
- No go on PATH and a checkout build produce the two exact S1 refusal lines and exit 1, before any target lookup or install-location check.
- A valid --version target is used exactly without a latest-release lookup; a value that does not start with v and parse as semver produces the exact usage message and exits 2.
- Without --version, go list -m -json github.com/mvelosop/vloop@latest runs outside a module in a temporary directory, a failure uses its first error line, and only a semantically newer target is selected.
- --check prints the exact current/latest or current/target two-line report, exits 0, never installs, never asks go for its install directory and leaves plugins empty.
- With no install needed, the exact up-to-date line is printed; pre-releases newer than latest are not downgraded.
- The install-location calculation queries GOBIN then GOPATH/bin only when installation is about to happen, uses exact resolved-path comparison (case-insensitive on Windows), and emits the specified refusal when it differs.
- Global --json emits only one self-update/v1 document for successful runs, with no human lines; refusals remain one vloop: stderr line, and the embedded schemas/self-update.v1.json validates every documented field.
- Committed Go tests use temporary Go fake programs for go, cover all target/refusal/check/JSON branches above, never call the real go install or the network, and are run by the repository's go check.
- docs/guide/commands.md is regenerated from the command tree and documents the new command and flags rather than being hand-maintained.
- No command other than self-update gains a go or network call; invalid or changed output in unrelated commands is not accepted.

**From the last attempt:** Committed self-update tests do not pin JSON output for non-check successes or refusals, and do not assert the exact install-location refusal, so T1's all-branch committed-test acceptance criterion is unmet.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T1/gate.ps1
```

</details>

### T2 — Install self-update targets safely

`pending` · depends on: T1

Replace the released binary at the Go install location while preserving the old executable on every failure. Windows move-aside and restoration make the update safe even though the running executable cannot be overwritten in place.

**Acceptance**

- A necessary install prints the exact installing line, runs go install github.com/mvelosop/vloop/cmd/vloop@&lt;target&gt;, then prints the exact installed line on success.
- A go install failure prints its last error line, exits 1, leaves no partially replaced running binary and does not continue to plugins.
- On Windows, a stale vloop.exe.old is removed before moving the running vloop.exe aside; remove and move failures use the exact brief messages and install nothing.
- On Windows, a failed install moves vloop.exe.old back; if restoration also fails, the exact combined failure line tells the operator to rename it by hand.
- A successful --version request installs exactly that version even when it is older or a pre-release, and installed/newer/dir fields in self-update/v1 reflect what happened.
- Committed Go tests build a temporary Go fake for go and cover success, failure, explicit older installation and the Windows old-file/restore paths without invoking the real go install or network.
- The implementation does not install, downgrade or re-exec a binary outside the exact T1 decision and install-location checks.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T2/gate.ps1
```

</details>

### T3 — Update installed agent plugins

`pending` · depends on: T2

Complete self-update by updating each locally discoverable vloop plugin after the binary step. Plugin discovery and failures remain per-agent and non-fatal, with the same scope rules doctor already uses.

**Acceptance**

- After an installed or up-to-date binary, self-update checks claude then copilot only when each executable is on PATH and runs each plugin-list command from the current directory.
- Claude counts the exact doctor rule: id starts with vloop@, version present, scope user or empty, or another scope whose project path matches the current directory; Copilot counts name vloop.
- Claude refreshes the vloop marketplace first, then runs one scoped update per matching entry; Copilot runs one update, with all command, success and failure text exactly as the brief specifies.
- Agent absence prints nothing; not installed, malformed or failed list, marketplace failure and update failure each produce their specified line, with plugin failures never changing exit 0.
- With --version, the human output carries the exact latest-may-differ warning before plugin work; self-update/v1 carries no warning and instead records each installed/not-installed/failed plugin structurally.
- A plugin updated or a binary installed produces the exact next line, while --check never inspects or updates plugins and leaves JSON plugins empty.
- Committed Go tests build temporary fake claude and copilot programs, cover scope filtering, ordering, malformed/failing tools, JSON and non-fatal failures, and never run real agent CLIs or the network.
- The implementation reuses doctor's plugin-list/discovery behaviour rather than adding divergent matching rules, and no command other than self-update calls plugin update or marketplace update.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T3/gate.ps1
```

</details>

### T4 — Point existing commands and guides to self-update

`pending` · depends on: T3

Make machine-wide self-update discoverable from the repository upgrade workflow and doctor without changing those commands' offline behaviour. This also records the intentional path exception and the release handoff.

**Acceptance**

- vloop upgrade --help includes exactly: To update the vloop binary and its plugins, run vloop self-update first.
- Doctor's Claude and Copilot plugin-version warnings alone end with exactly — vloop run is unaffected; vloop self-update updates the binary and the plugin; doctor performs no new network or update call.
- docs/guide/concepts.md explains self-update beside go install and vloop upgrade, including the checkout, missing-go and install-location refusals in a sentence, and docs/guide/commands.md is regenerated.
- The C-2 line in docs/domain/domain-model.md and the platform context say that self-update outside any repository prints external paths with / and ~ for the home directory.
- go run ./tools/release publish reports the final handoff as one line: then: vloop self-update, and vloop upgrade in each repository.
- No text or output elsewhere changes except the explicitly listed help, two doctor warnings, guide/domain additions, generated command reference and release handoff.
- Any committed tests for the new help and doctor wording use the CLI and fake agent programs rather than real agents or network access.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T4/gate.ps1
```

</details>

### T5 — Verify the complete self-update workflow

`pending` · depends on: T1, T2, T3, T4

Prove the binary and plugin paths work together through the worked example after every earlier task is present. This final task closes gaps that isolated gates or task-local tests could miss.

**Acceptance**

- A committed cross-platform Go integration test runs the complete worked example with temporary fake go and claude programs and verifies each output line and exit code in order.
- The integrated run verifies successful installation, plugin update and the exact next line, and it never calls the real go install, agent CLIs or network.
- All earlier contract gates pass together on the final tree, including embedded self-update/v1 validation, Windows install restoration and pointer text.
- The integration test does not replace or weaken the task-specific committed tests; every new behaviour remains covered by the repository's go check.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T5/gate.ps1
```

</details>


# Plan — B20261126-1418-b45-docs-mixed-os-and-any-agent

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **2/6 done** · iteration 4

**Brief:** `docs/briefs/B20261126-1418-b45-docs-mixed-os-and-any-agent.loop-brief.md` · **Updated:** 2026-11-26T17:37:45Z

## Progress

- [x] **T1** — Document the mai-code-2 evidence
- [x] **T2** — Make the README agent-neutral and cite mai-code-2 · 1 attempt(s)
- [ ] **T3** — Explain mixed teams in the README · 1 attempt(s)
- [ ] **T4** — Update the guides for any agent
- [ ] **T5** — List every session kind and portable gate shell
- [ ] **T6** — Run the complete documentation contract

## Tasks

### T1 — Document the mai-code-2 evidence

`done` · depends on: none

Add the D6 evidence section before any README link points to it. The section must turn B40 through B44 telemetry into a traceable account without overstating qualification.

**Acceptance**

- The design note has the exact heading "## mai-code-2 in the loop (2026-11)" and records the URL shortener run as 10 of 10 tasks, 10 first-pass, about $1.31, explicitly measured in an acceptance repository outside this repository.
- The section names B40 through B44 with task count, first-pass count and cost.total_usd rounded to cents exactly as read from each committed metrics snapshot, including B42's halted attempt.
- The section names every session kind B40 through B44 ran, says no second review ran, and records that B40 through B45 brief reviews ran on mai-code-2 with their kept records under .vloop/state/brief-reviews/.
- The section records B39's $8.10 Claude comparison from its run record, B40's HTTP 503 exit 7 with no attempt counted, real defects caught by reviews and gate disputes, and operator-found defects in B42 and B43.
- No claim says or implies that mai-code-2 is qualified under the evaluation rule.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T1/gate.ps1
```

</details>

### T2 — Make the README agent-neutral and cite mai-code-2

`done` · 1 attempt(s) · depends on: T1

Apply D1 and D3 to README.md now that the evidence section exists. Readers must understand that either supported agent can drive the loop and that mai-code-2's results are evidence, not qualification.

**Acceptance**

- README.md's opening names both Claude Code and GitHub Copilot CLI, says an agent plans the brief, and no longer contains "vloop has Claude plan it".
- Prerequisites names the copilot CLI and a usable GitHub login when Copilot runs, while retaining the two Claude-specific bullets; quickstart says doctor checks the agent CLIs in use rather than Claude alone.
- The loop summary says a fresh agent session for Claude Code or Copilot CLI, and "You ask; the skill drives" addresses the user's agent rather than Claude Code alone.
- The Copilot section adds the prescribed mai-code-2 sentence for the URL shortener and B40 through B44, links D6's section, names every session kind but the second review, and says it is not yet qualified.
- Neither "Every session is a fresh Claude session" nor "mai-code-2 is qualified" appears, and the gpt-6.1-terra sentence still limits plan and gate review to Claude and leaves verify untried.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T2/gate.ps1
```

</details>

### T3 — Explain mixed teams in the README

`pending` · 1 attempt(s) · depends on: T2

Apply D2 to README.md after its agent-neutral edits. The README needs one short mixed-OS entry point plus accurate install and Windows-only Copilot guidance.

**Acceptance**

- README.md has a section headed "## Mixed macOS and Windows teams" after Quickstart and before the Copilot section, and its body is at most 150 words.
- That section states the shared platforms and shell values, the refusal of local or environment overrides, the fixed-command portability lint, portable pwsh cmdlets and project toolchain, UTF-8 execution, and that this repository is such a team.
- The section links docs/guide/configuration.md#mixed-teams, and the Copilot Windows bullet retains only the PowerShell hook plus the B40 through B44 Copilot-on-Windows evidence and links back to the mixed-team section.
- README no longer says Windows is untested; Install points Windows users to the PowerShell PATH line in concepts.md, then explains self-update and the per-repository vloop upgrade step.
- No mixed-team claim contradicts docs/guide/configuration.md.

**From the last attempt:** check go failed — see .vloop/state/runs/B20261126-1418-b45-docs-mixed-os-and-any-agent/20261126-170829/checks/004-go.fail.log

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T3/gate.ps1
```

</details>

### T4 — Update the guides for any agent

`pending` · depends on: none

Apply D4 consistently across the four guides. Agent-neutral prose must describe only where Claude and Copilot genuinely share behavior, while preserving agent-specific mechanisms.

**Acceptance**

- concepts.md says each configured agent runs a fresh agent session, explains both guideline delivery mechanisms, shows both interactive plugin-dir commands, names agent-eval for other agents, and adds a short mixed-team portable-gates note.
- metrics.md says Copilot reports no tokens, estimates cost from copilot.credit-usd with cost_source estimated, sums every phase in credits.total, and counts cache_hit_ratio for Claude sessions only.
- All seven model.* values explain that the setting is a Claude alias or model name, or a Copilot model id when that phase's agent is copilot; the old "any model name" value is absent.
- evals.md's Calls claude row says claude plugin eval calls claude and tools/agent-eval calls the agent it is given, while the qualification rule remains unchanged.
- No Claude-specific adapter, installation or qualification wording is made generic, and all new statements agree with the driver implementation.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T4/gate.ps1
```

</details>

### T5 — List every session kind and portable gate shell

`pending` · depends on: none

Apply D5 to vloop's generated CLAUDE.md text, this repository's own copy, and the operator and architect guidance. The implementation and its pinning test must describe the full loop on this mixed-OS repository.

**Acceptance**

- internal/install/section.go generates the exact session list "(plan, gate review, work, review, second review, verify)", and the existing test that pins that text is changed only to expect the new list.
- This repository's CLAUDE.md vloop section is edited to the same session list; no unrelated generated section text or behavior changes.
- The operator skill says this repository is a mixed team with platforms = ["macos", "windows"] and shell = "pwsh", its gates are pwsh and portable, and the BSD-tool rules apply only to plans whose shell is sh.
- The architect skill says gates are pwsh and portable for a mixed macOS and Windows team; it no longer states that this repository's gates run on macOS BSD tools.
- A committed Go test in the checks continues to pin the generated section, and changing that pin is the only test change in this task.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T5/gate.ps1
```

</details>

### T6 — Run the complete documentation contract

`pending` · depends on: T1, T2, T3, T4, T5

Close the pass by exercising every earlier gate together and replaying the brief's worked-example assertions. This catches cross-file drift that a single documentation task could miss.

**Acceptance**

- All five earlier gates pass from the finished tree, including the real-data evidence comparison and generated CLAUDE.md behavior.
- Every line of the brief's worked example is true: stale README and concepts strings are absent, the mixed-team heading and mai-code-2 and self-update strings are present, and mai-code-2 is not called qualified.
- The Copilot plugin-dir example exists, the old any-model table value is absent, the generated section source names all six session kinds, and the operator skill names shell = "pwsh".
- No behavior, flag, message, default or test outcome changed except the generated section text and its test pin.
- No changed prose contradicts the domain model or the binding guides, and no README mixed-team section exceeds 150 words.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T6/gate.ps1
```

</details>


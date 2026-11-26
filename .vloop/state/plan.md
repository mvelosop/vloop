# Plan — B20261126-1818-b46-i153-copilot-model-ids

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **1/6 done** · iteration 2

**Brief:** `docs/briefs/B20261126-1818-b46-i153-copilot-model-ids.loop-brief.md` · **Updated:** 2026-11-26T19:15:25Z

## Progress

- [x] **T1** — List Copilot model ids and the account plan · 1 attempt(s)
- [ ] **T2** — Check Copilot model ids in config set
- [ ] **T3** — Check configured Copilot models in doctor
- [ ] **T4** — Set several session kinds in one config command
- [ ] **T5** — Document the Copilot model commands
- [ ] **T6** — Verify the Copilot model workflow end to end

## Tasks

### T1 — List Copilot model ids and the account plan

`done` · 1 attempt(s) · depends on: none

Give the operator one CLI command that exposes the exact Copilot model ids, picker filtering, account policies, and Copilot plan returned by the two documented endpoints. This establishes the shared API client and the copilot-models/v1 document that later validation surfaces can reuse.

**Acceptance**

- vloop models copilot reads the models endpoint and prints the plan line first, then picker-enabled models sorted by strings order; --all adds non-picker entries, and the policy column is enabled, disabled, or -.
- vloop models copilot --json emits exactly one copilot-models/v1 document with plan null when unknown and policy null when the text column would be -.
- The models and plan requests use the first non-empty token from COPILOT_GITHUB_TOKEN, GH_TOKEN, GITHUB_TOKEN, then gh auth token; the required Copilot integration header is sent, VLOOP_COPILOT_API_URL and VLOOP_GITHUB_API_URL replace only their respective bases, and no token reaches output or errors.
- A missing login and a failed models request exit 1 with the brief's exact refusal; a missing or incorrect agent argument exits 2 with the brief's usage line. A failed or plan-less plan request is unknown, never an error.
- The listing command uses a 20-second timeout per call and only the standard library; the vloop models command performs no network access for an invalid agent argument.
- Committed Go tests exercise the output formats, filtering, JSON schema, request headers, unknown plan, failure, and no-login paths through httptest; no test reaches GitHub or Copilot for real.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T1/gate.ps1
```

</details>

### T2 — Check Copilot model ids in config set

`pending` · depends on: T1

Prevent a single model.&lt;kind&gt; write from introducing an unusable Copilot id while preserving config set's existing warning-only behavior for an unavailable API. The validation must use the pair configuration resolves to after a candidate write, including local and environment overrides.

**Acceptance**

- Before writing model.&lt;kind&gt;, with or without --local, config set resolves agent.&lt;kind&gt; and the candidate model by C-1 and refuses a Copilot Claude alias, an id absent from the --all list, or a disabled id with the brief's exact exit-2 line; all other keys retain their existing validation.
- The models request is skipped for a removed model, for a kind not resolved to copilot, and when the candidate is a Claude alias already decidable locally.
- A failed, timed-out, or malformed models call and a missing login do not refuse a single-key set: it writes the value and emits exactly the brief's stderr warning with exit 0; the warning or error never contains the token.
- Each config-set validation uses a five-second models-call timeout and writes nothing on refusal; a successful removal remains the existing remove operation and is never checked.
- Committed Go tests use httptest and cover listed, absent, disabled, alias, unreachable, no-login, removal, --local, and environment/local precedence; no test reaches the real APIs.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T2/gate.ps1
```

</details>

### T3 — Check configured Copilot models in doctor

`pending` · depends on: T1

Extend doctor's existing Copilot preflight with one row that checks every in-use Copilot kind's resolved model and displays the account plan. This makes an edited invalid id refuse vloop run before a paid session starts without turning API unavailability into a refusal.

**Acceptance**

- When at least one session kind in use resolves to copilot, doctor places copilot models immediately after copilot and makes one models call and one plan call, reporting the distinct configured ids and plan exactly as the brief specifies in text and JSON.
- An unlisted, disabled, or Claude-alias id produces a problem naming every affected model.&lt;kind&gt;, joined with ; ; a Claude alias is reported without any network request and receives ; ids not checked against the list when the list also cannot be reached.
- A failed models call is a warning, an unknown or failed plan call remains plan unknown, and absence of login adds no models row because the existing copilot login row remains the problem; absence of a Copilot kind adds no models row or network call.
- The row uses five-second timeouts, the token never appears in text or JSON, and the row's problem participates in vloop run's preflight so the run refuses before a plan session starts.
- Committed Go tests use httptest for pass, missing, disabled, alias, unreachable, no-login, no-Copilot-kind, distinct-count, plan-unknown, ordering, and JSON-row behavior; no test reaches the real APIs.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T3/gate.ps1
```

</details>

### T4 — Set several session kinds in one config command

`pending` · depends on: T2

Replace up to fourteen separate config writes with one atomic --kinds operation for agent, model, and effort. Every candidate kind must be validated against its resolved pair before any shared or local file changes.

**Acceptance**

- vloop config set --kinds &lt;comma-list&gt; accepts exactly plan, gate-review, work, review, second-review, verify, and review-brief, requires at least one of --agent, --model, or --effort, and rejects a positional key or value with the brief's exact usage lines.
- Every named kind's existing agent, model, and effort keys are validated as a single-key set would validate them; no new key or session kind is added.
- All resulting Copilot pairs are checked before any write, one models request is made per invocation, removal is resolved by C-1, an environment override can still supply a Claude alias, and removing agent alone is never refused.
- Any unknown kind, invalid value, alias, missing id, disabled id, or other refusal leaves both shared and local files byte-for-byte unchanged; --local targets only the local file and success prints only the warnings a single set would print.
- Committed Go tests cover success, each option, every kind, exact usage errors, API reuse, model-result checks, removals, environment precedence, --local, and all-or-nothing failure; no test reaches a real API.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T4/gate.ps1
```

</details>

### T5 — Document the Copilot model commands

`pending` · depends on: T1, T4

Replace the operator's manual API recipes with the supported vloop command and document how the new validation and multi-kind write affect configuration. Regenerate the command reference so the new command and flag are discoverable together.

**Acceptance**

- The Copilot model ids section uses vloop models copilot in place of the curl and PowerShell raw-API snippets, explains picker filtering, enabled/disabled/- policies, and the plan or plan unknown line.
- The section states that config set and doctor validate ids for Copilot kinds and includes one vloop config set --kinds example; it contains no line that calls api.githubcopilot.com directly.
- docs/guide/commands.md is regenerated from the completed CLI and includes vloop models copilot plus every --kinds, --agent, --model, and --effort surface.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T5/gate.ps1
```

</details>

### T6 — Verify the Copilot model workflow end to end

`pending` · depends on: T1, T2, T3, T4, T5

Run all earlier behavioral gates together and replay the brief's worked example in an initialized temporary repository. This closes the run only when listing, set-time checks, doctor, atomic --kinds writes, and failure warnings compose as one operator workflow.

**Acceptance**

- Every earlier task's planner-owned gate passes against the completed tree, and the close oracle executes the brief's worked example line by line with httptest servers and stub credentials in a temporary vloop-init repository.
- The workflow proves exact picker and --all output, successful multi-kind writes, exit-2 absent/disabled/alias refusals, byte-for-byte all-or-nothing behavior, the doctor pass row, and warning-only writes and doctor after the stub stops.
- No output, log, fixture, or error produced by the workflow contains the stub token, and the gate itself creates only temporary directories outside the repository; no network request reaches GitHub or Copilot.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T6/gate.ps1
```

</details>


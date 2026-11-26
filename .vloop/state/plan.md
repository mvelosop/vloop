# Plan — B20261126-1003-b41-i161-doctor-gate-shell-replaced-plan

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **0/3 done** · iteration 1

**Brief:** `docs/briefs/B20261126-1003-b41-i161-doctor-gate-shell-replaced-plan.loop-brief.md` · **Updated:** 2026-11-26T10:40:57Z

## Progress

- [ ] **T1** — Make replaced plans use the configured gate shell · 1 attempt(s)
- [ ] **T2** — Document the configured shell for replaced plans
- [ ] **T3** — Verify the complete doctor shell replacement contract

## Tasks

### T1 — Make replaced plans use the configured gate shell

`pending` · 1 attempt(s) · depends on: none

Make doctor's gate shell row follow the same replaced-plan decision as its plan row, so a finished plan from another brief cannot reject a run that will use the configured shell. Implement G1 across the complete status table while preserving every plan-owned shell result and message.

**Acceptance**

- The plan and gate shell rows use one shared replaced-plan decision: complete state/v2 plans of another brief are replaced for a bare doctor or a run naming another brief, while a run naming the plan's own brief or no brief and every non-complete status keep the plan in use; complete state/v1 plans are replaced in every invocation.
- When that shared decision says replaced, and when no plan exists, the gate shell row uses the configured shell and exactly B40's configured-shell pass, warning and mixed-team problem outcomes and messages; otherwise it uses the plan's recorded shell and the existing plan-shell problem message without changing a result or message in that branch of the table.
- Doctor --json reports the same gate shell result and message, and vloop run's preflight does not report the old plan shell for a different brief but does report it when resuming or naming the replaced plan's own brief.
- The implementation does not duplicate the replaced-plan calculation between the plan and gate shell rows.
- Committed tests exercise every row of G1's table, including planning, running, blocked, stalled and halted, and run on Windows without depending on POSIX stub agents; the go check executes those tests.

**From the last attempt:** A complete state/v1 plan now takes the generic replaced-plan branch and prints `a complete plan of  — the next brief's plan replaces it` instead of the existing `a complete state/v1 plan — the next brief's plan replaces it`; G1 explicitly requires no other doctor output to change.; The gate and committed tests cover state/v1 only as a gate-shell row and do not assert the plan row's existing state/v1 message, so this output regression passes.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T1/gate.ps1
```

</details>

### T2 — Document the configured shell for replaced plans

`pending` · depends on: T1

Add the single guide sentence promised by G1 so users of doctor understand why a completed plan from another brief does not control the gate shell row. Keep the change confined to the existing Gates and the gate shell section.

**Acceptance**

- docs/guide/concepts.md contains one sentence in the Gates and the gate shell section stating that for a plan the next brief replaces, doctor checks the configured shell.
- No other documentation is changed by this task.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T2/gate.ps1
```

</details>

### T3 — Verify the complete doctor shell replacement contract

`pending` · depends on: T1, T2

Close the brief only after the implementation and documentation gates pass together and the worked example is exercised through the built CLI in a temporary repository. This preserves the whole-run acceptance independently of each individual task.

**Acceptance**

- T1's exhaustive Windows CLI oracle and T2's guide oracle pass together against the final tree.
- The worked example succeeds through a CLI built from the final tree: complete-plan doctor text and JSON pass on the configured shell, running plans use the old plan shell and fail exactly as shown, missing configured pwsh becomes the exact non-mixed warning, and run preflight treats a different brief and the plan's own brief differently.
- The close adds no product change that is not required by the brief and does not weaken either earlier gate or the go check.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T3/gate.ps1
```

</details>


# Plan — B20261126-1008-b42-i162-fence-reads-pwsh

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **2/5 done** · iteration 6

**Brief:** `docs/briefs/B20261126-1008-b42-i162-fence-reads-pwsh.loop-brief.md` · **Updated:** 2026-11-26T12:57:35Z

## Progress

- [x] **T1** — Judge pwsh and cmd wrappers · 2 attempt(s)
- [x] **T2** — Refuse pwsh evaluation forms · 1 attempt(s)
- [ ] **T3** — Read common pwsh forms and guide refusals · 1 attempt(s)
- [ ] **T4** — Document the pwsh fence rules
- [ ] **T5** — Verify the complete pwsh fence contract

## Tasks

### T1 — Judge pwsh and cmd wrappers

`done` · 2 attempt(s) · depends on: none

Make pwsh, powershell and cmd execute the same fence decision on their literal command strings that a direct call receives. This closes the wrapper bypass while preserving the existing wrapper nesting limit and fail-closed behavior.

**Acceptance**

- For both the bash and powershell tools, pwsh and powershell accept only the pinned literal -Command/-c form after the allowed pre-options, judge every remaining word joined with spaces as a powershell call, preserve the phase and root, and count each wrapper toward the eight-level nesting limit.
- Missing, expanded, encoded, file, stdin, bare-script and unsupported-option pwsh forms fail closed with the exact W1 reasons, including the wrapper's written pwsh or powershell name where specified.
- cmd accepts only /c or /k after /d, /q and /s, judges its literal string split at &amp;, &amp;&amp;, || and |, and denies % or ^; every other cmd form fails closed with the exact W1 reason.
- A command inside a wrapper is allowed or denied exactly as the same command in the wrapped tool would be, including case-insensitive command and option matching and names with a path or .exe.
- Committed Go tests cover the allowed and refused wrapper forms, option and name variants, command matching, nesting, phase/root preservation and both outer tools, and the go check executes them.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T1/gate.ps1
```

</details>

### T2 — Refuse pwsh evaluation forms

`done` · 1 attempt(s) · depends on: T1

Bring PowerShell calls to eval's fail-closed bar by refusing named dynamic-execution commands and script blocks at every command position. This prevents pwsh syntax from bypassing the deny list without expanding the fence into a full PowerShell sandbox.

**Acceptance**

- In a powershell call, Invoke-Expression, Invoke-Command, Start-Process, Start-Job, Start-ThreadJob, Invoke-Item and every listed alias are denied wherever a command name stands, case-insensitively and with a path or .exe, using the command name exactly as written.
- The same W2 denials apply after pipelines, separators and pwsh/powershell -Command wrappers and inside a W3 parenthesized pipeline.
- Every unquoted { in a powershell call is denied with the exact script-block guidance, including @{, &amp; { } and cmdlet script blocks; quoted braces remain literal.
- The bash tool keeps its existing brace-group and POSIX decisions; the simplified pwsh cmdlet forms named by the brief stay allowed; and no new PowerShell form outside W1-W3 is accepted.
- Committed Go tests cover every pinned command and alias, command position, case/path variant, script-block spelling and the bash non-regression, and the go check executes them.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T2/gate.ps1
```

</details>

### T3 — Read common pwsh forms and guide refusals

`pending` · 1 attempt(s) · depends on: T1, T2

Stop rejecting the two pwsh forms that caused B40's denials while retaining exact, actionable fail-closed messages for syntax the fence still cannot read. The PowerShell-specific reading must not alter the bash tool's existing grammar.

**Acceptance**

- A powershell statement or leading pipeline stage of the exact (&lt;pipeline&gt;).&lt;Property&gt; form judges its parentheses contents as a powershell call and allows only one or more letter, digit or underscore property-path components with no trailing syntax except end of statement or |; method calls and parentheses in every other position deny with W4's reason.
- A powershell whole statement $env:&lt;NAME&gt; = &lt;literal&gt; with a valid non-digit-leading name and one non-expanding literal is allowed and does not stop judging following commands; every other $env: statement keeps the existing command-word-starts-with-$ reason using the split first word.
- In a powershell call, unsupported parentheses and @(…) use the exact parentheses guidance, $(…) uses the exact inner-command guidance, a backtick uses the exact one-line guidance, and @{…} reaches W2's script-block denial.
- W2-W4's PowerShell-specific judgment also applies inside pwsh/powershell -Command strings, while sh -c inside such a string remains POSIX and direct bash-tool calls retain the old refusals and reasons.
- Committed Go tests cover both successful W3 forms, the invalid-name/value boundaries, property and pipeline positions, the exact W4 reasons, nested W1/W2/W3 interactions and bash non-regressions, and the go check executes them.

**From the last attempt:** PowerShell property reads are accepted after an existing pipeline stage, although W3 permits them only as a whole statement or the leading stage; the parser must deny that position with the W4 parentheses reason.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T3/gate.ps1
```

</details>

### T4 — Document the pwsh fence rules

`pending` · depends on: T1, T2, T3

Publish the new wrapper, eval, script-block and guided-refusal behavior where users and future fence auditors will encounter it. Keep the command reference generated from the CLI help rather than hand-maintaining a second copy.

**Acceptance**

- docs/guide/configuration.md's The fence hook section gains one bullet that states the W1 wrapper behavior, W2's refused pwsh eval-like commands and script blocks, W3's two read forms and W4's refusal guidance without implying that all pwsh is parsed.
- docs/design-notes/fence-audit.md gains rows for W1 wrappers and W2 dynamic execution/script blocks, with the correct hook and driver enforcement claims and no unrelated audit prose changed.
- vloop fence check --help names pwsh and powershell -Command and cmd /c in its wrapper list and states that powershell calls refuse pwsh eval-like commands and script blocks.
- docs/guide/commands.md is regenerated from the command tree so the three required help additions are present, and no skill text changes.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T4/gate.ps1
```

</details>

### T5 — Verify the complete pwsh fence contract

`pending` · depends on: T1, T2, T3, T4

Close the brief only after the independent feature and documentation gates pass together and the full worked example is exercised line by line through a fresh CLI build. This makes regressions between tasks visible without turning the repository check into a gate.

**Acceptance**

- T1 through T4's gates pass together against the final tree.
- Every worked-example line runs through a CLI built from the final tree and returns the exact shown allow or deny result and exit status, including the unchanged bash lines.
- The final implementation preserves the phase and root through nested wrappers, still fails closed on unreadable forms, changes no fence file or fence matching rule, adds no dependency, and does not weaken the go check.
- The close adds no production change beyond what the prior tasks require and leaves a committed test for every behavior introduced by W1-W3.

<details><summary>verify command</summary>

```sh
pwsh -NoProfile -File .vloop/state/gates/T5/gate.ps1
```

</details>


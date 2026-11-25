# Plan — B20261125-2207-b40-i156-mixed-os-pwsh

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **0/6 done** · iteration 0

**Brief:** `docs/briefs/B20261125-2207-b40-i156-mixed-os-pwsh.loop-brief.md` · **Updated:** 2026-11-25T23:20:31Z

## Progress

- [ ] **T1** — Add mixed-team platform configuration
- [ ] **T2** — Report mixed-team health and block unsafe runs
- [ ] **T3** — Lint mixed-team gates for portable commands
- [ ] **T4** — Teach skills to write portable mixed-team gates
- [ ] **T5** — Document mixed teams and the repository policy
- [ ] **T6** — Prove the complete mixed-team workflow

## Tasks

### T1 — Add mixed-team platform configuration

`pending` · depends on: none

Give repositories a shared-only declaration of their supported platforms and enforce the mixed-team shell ownership rules at every config entry point. This establishes the configuration contract that doctor, preflight and the portability lint build on.

**Acceptance**

- The platforms config key is an unset-by-default list accepting only linux, macos and windows; config set stores the caller's list order, and invalid input keeps the existing list-key wording and exits 2.
- Platforms are distinct for mixed-team decisions, and every message renders them deduplicated and sorted as linux, macos, windows; a shared list containing two or more distinct entries is mixed.
- A local file or VLOOP_PLATFORMS value for platforms is an exit-1 config-read error, config set --local platforms is an exit-2 error, and each path reports its own exact source with no file written.
- When the platforms list is mixed, a local file or VLOOP_SHELL value for shell is an exit-1 config-read error regardless of value, config set --local shell is an exit-2 error, and none of those writes occur; without a mixed team, those shell overrides continue to work.
- The implementation leaves no cross-platform test behind that skips on Windows and adds committed config/CLI tests, including the exact worked-example refusals, for the go check to run.
- The configuration guide's keys table contains platforms with VLOOP_PLATFORMS, and its shell row says that a mixed team requires shared pwsh; no alternative spelling of the platform names is documented.

<details><summary>verify command</summary>

```sh
$o = go test ./.vloop/state/gates/T1 -v; if ($LASTEXITCODE -ne 0) { $o; exit 1 }; if (-not ($o | Select-String -SimpleMatch -Pattern '--- PASS:' -Quiet)) { throw 'gate did not run a passing oracle test' }
```

</details>

### T2 — Report mixed-team health and block unsafe runs

`pending` · depends on: T1

Expose the new configuration contract through doctor and run preflight so mixed repositories fail clearly before work starts. Preserve the existing P-5 behavior when a plan already records its shell, while adding only the required warnings for ordinary repositories.

**Acceptance**

- Doctor places platforms immediately after config.local, or immediately after config when no config.local row exists, and omits it when platforms is unset or config is a problem.
- The platforms row reports the exact pass, missing-shell problem, wrong-shell problem and not-on-this-machine warning from M2, using sorted deduplicated platforms and mapping Go darwin to macos.
- With no plan in use, a mixed team reports a missing configured pwsh as a gate-shell problem; with a plan in use, the plan's recorded shell retains the existing pass/problem behavior, and a non-mixed team retains the existing warning.
- A repository with checks, no platforms key and no shared shell produces the exact checks warning showing the resolved shell source on that machine; vloop run does not print this warning.
- A mixed repository produces one checks warning for each check whose command word first matches the M3 list, and vloop run prints it; a listed token used only as data or inside a quoted argument does not warn.
- Run preflight prints the exact platforms problem or warning with two spaces and its symbol, counts the platforms problem as a refusal, and reaches no plan session or agent when it fails.
- The task adds committed doctor and preflight tests that run on Windows, macOS and Linux without a real claude or copilot process, and the go check runs them.

<details><summary>verify command</summary>

```sh
$o = go test ./.vloop/state/gates/T2 -v; if ($LASTEXITCODE -ne 0) { $o; exit 1 }; if (-not ($o | Select-String -SimpleMatch -Pattern '--- PASS:' -Quiet)) { throw 'gate did not run a passing oracle test' }
```

</details>

### T3 — Lint mixed-team gates for portable commands

`pending` · depends on: T1

Make plan acceptance reject gate commands and PowerShell scripts that rely on platform-specific native tools when platforms is mixed. The lint must recognize PowerShell command positions precisely without changing behavior for unmixed or undeclared repositories.

**Acceptance**

- One authoritative, ordered list in the driver contains exactly awk, basename, bash, cat, chmod, cmd, cmp, cp, cut, date, diff, dirname, env, fc, find, findstr, grep, head, ln, ls, mkdir, mktemp, mv, powershell, printf, readlink, rm, robocopy, sed, seq, sh, sleep, sort, stat, tail, tee, test, touch, tr, uniq, wc, which, xargs, xcopy and zsh.
- For a mixed team, command-position words in each task verify and in every .ps1 file at any depth of that task's gate folder are linted; the first listed command in each command source produces the exact M3 refusal naming the task, source file when applicable, and sorted platforms.
- Matching uses the final path element after slash or backslash, removes a trailing .exe, ignores case, and recognizes wholly quoted command words after the call operator; separators are recognized only outside single and double quotes.
- Quoted data, pwsh keywords and command names beginning with $, -, @, [ or a digit are not matched, and Select-String -Pattern 'grep', go test -run TestGrep and node -e "require('fs')" remain accepted.
- With platforms unset or having only one distinct entry, none of these verify or .ps1 command checks are applied, and all earlier gate-shape refusals remain intact.
- Plan acceptance reads B39's plan from 4c69adf0 as real data, computes the expected task count from that plan, and refuses all eight sh-based gates for mixed platforms without hard-coding the count.
- The task adds committed, non-skipping Go tests for command parsing, exact list membership, .ps1 recursion, unmixed behavior and the B39 real-data plan, and the go check runs them.

<details><summary>verify command</summary>

```sh
$o = go test ./.vloop/state/gates/T3 -v; if ($LASTEXITCODE -ne 0) { $o; exit 1 }; if (-not ($o | Select-String -SimpleMatch -Pattern '--- PASS:' -Quiet)) { throw 'gate did not run a passing oracle test' }
```

</details>

### T4 — Teach skills to write portable mixed-team gates

`pending` · depends on: T3

Align the planning and gate-review skills with the driver's portability lint so generated pwsh gates are accepted for mixed teams and missing cross-platform tools are caught by review. This keeps the human-facing skill contract synchronized with the executable rule.

**Acceptance**

- The plan skill's shell bullet says that two or more configured platforms require pwsh gates using pwsh cmdlets and the project's own toolchain, never an M3 listed command, and it says .ps1 gate scripts run with pwsh -NoProfile -File.
- The same plan bullet retains the BSD grep, sed, awk and date guidance when the plan shell is sh on macOS.
- The gate-review skill says a mixed-team gate that calls any tool unavailable on even one declared platform is unpassable even if it passes on the reviewing host.
- The command list is copied from the driver's authoritative list without a different spelling or ordering, and neither skill replaces the driver lint as the enforcement point.
- Committed structural skill tests assert these mixed-team rules and the go check runs them; no test invokes a real agent or skips on Windows.

<details><summary>verify command</summary>

```sh
if (-not (Select-String -Path plugin/skills/plan/SKILL.md -SimpleMatch -Pattern 'vloop config get platforms' -Quiet)) { throw 'plan skill has no mixed-platform instruction' }; if (-not (Select-String -Path plugin/skills/plan/SKILL.md -SimpleMatch -Pattern 'pwsh -NoProfile -File' -Quiet)) { throw 'plan skill has no .ps1 invocation instruction' }; if (-not (Select-String -Path plugin/skills/gate-review/SKILL.md -SimpleMatch -Pattern 'platforms' -Quiet)) { throw 'gate-review skill has no mixed-platform rule' }
```

</details>

### T5 — Document mixed teams and the repository policy

`pending` · depends on: T1, T4

Make the mixed-OS model, configuration workflow and portable check guidance discoverable and consistent across the product's binding documents. The repository remains the first user by reading back its ready-commit platforms declaration.

**Acceptance**

- The configuration guide has a platforms key, a mixed-team shell rule and a Mixed teams section covering declaration, shared-only errors, pwsh rationale, portable plain cmd &amp;&amp; cmd or program-run checks, every listed command and pwsh cmdlet/toolchain replacements, plus the macOS and Linux pwsh install guidance from M5.
- docs/domain/domain-model.md defines C-5 exactly in substance: two or more platforms require shared pwsh, platforms and shell are shared-only, and gates call no command a declared platform lacks.
- The platform context keys table includes platforms and names its shell-rule and gate-lint execution effect; the roadmap's inherited gate-shell decision names platforms and the mixed-team pwsh rule.
- README.md links or names the Mixed teams configuration section and stays below its enforced 200-line cap.
- The repository's .vloop/config.toml keeps platforms = ["macos", "windows"] beside shell = "pwsh" as written by the ready commit.
- The command guide is regenerated with go generate ./... if any command help changed, and the committed guide/README structural tests in the go check pass without weakening them.

<details><summary>verify command</summary>

```sh
if (-not (Select-String -Path docs/guide/configuration.md -SimpleMatch -Pattern '## Mixed teams' -Quiet)) { throw 'configuration guide has no Mixed teams section' }; $files = @('docs/domain/domain-model.md','docs/domain/platform/platform-context.md','docs/design-notes/vloop-roadmap.md','README.md','.vloop/config.toml'); foreach ($f in $files) { if (-not (Select-String -Path $f -SimpleMatch -Pattern 'platforms' -Quiet)) { throw "$f does not name platforms" } }
```

</details>

### T6 — Prove the complete mixed-team workflow

`pending` · depends on: T2, T3, T5

Exercise the whole worked example after all component tasks so the six tasks close as one coherent feature rather than isolated fixes. This final task leaves durable end-to-end regression coverage for the config, doctor and preflight messages the brief pins.

**Acceptance**

- A fresh temporary repository follows the worked example from default-shell doctor warning through config set platforms and shell, doctor pass or machine warning, each exact override refusal and invalid-value exit 2.
- The end-to-end coverage proves vloop config get platforms prints the stored order, while doctor and errors print deduplicated sorted platforms.
- The end-to-end coverage drives vloop run with the platforms problem and portable-check warning and proves it exits 1 before a plan session starts.
- This task leaves committed tests covering every worked-example line in the repository's existing test surface, runs on Windows, macOS and Linux without skips or real agents, and the go check runs them.
- No repository without platforms changes doctor, run or plan acceptance behavior except the specified default-shell warning, and the final regression coverage names that violation explicitly.

<details><summary>verify command</summary>

```sh
$o = go test ./.vloop/state/gates/T6 -v; if ($LASTEXITCODE -ne 0) { $o; exit 1 }; if (-not ($o | Select-String -SimpleMatch -Pattern '--- PASS:' -Quiet)) { throw 'gate did not run a passing oracle test' }
```

</details>


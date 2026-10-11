# Plan — B20261011-2251-b14-learning-loop

<!-- Rendered from .vloop/state/state.json by vloop status --markdown. Do NOT edit. -->

**Status:** running · **1/13 done** · iteration 3

**Brief:** `docs/briefs/B20261011-2251-b14-learning-loop.loop-brief.md` · **Updated:** 2026-10-11T23:47:18Z

## Progress

- [x] **T1** — Make vloop init write an empty .vloop/repo-guidelines.md
- [ ] **T2** — Hand .vloop/repo-guidelines.md to every session vloop run starts · 2 attempt(s)
- [ ] **T3** — Add the guidelines paragraph to the four session skills and teach the gate review the two rules it missed
- [ ] **T4** — Make vloop brief check warn when a brief changes what a listed test pins without naming the test
- [ ] **T5** — Make vloop brief check warn when a contract names an external format without citing its source
- [ ] **T6** — Add the eval case gate-review-real-misses: six gates that fooled the gate review, and one sound gate
- [ ] **T7** — Repair the three broken eval cases: plan-checks-its-gates' fixture grader and operate-proposes-options' scaffold and graders
- [ ] **T8** — Add doctor's guidelines check: provenance on every rule, and a budget of 20
- [ ] **T9** — Put the plugin skills on word budgets enforced by a test
- [ ] **T10** — Add the instructions line to vloop metrics: the words each session kind starts with
- [ ] **T11** — Add rule and rule-outcome to defects, the defect report to brief close, and the triage to the operate skill
- [ ] **T12** — Document the repository guidelines, the two warnings, the budget, the instructions line and the defect report
- [ ] **T13** — Close: commit the worked example as an end-to-end test and pass the real-data checks

## Tasks

### T1 — Make vloop init write an empty .vloop/repo-guidelines.md

`done` · depends on: none

A lesson about one repository has nowhere to live; L2 gives every new repository the file from day one, so the place for lessons exists before the first one. Later tasks hand this file to every session (T2), read its two sections in brief check (T4, T5) and count its rules in doctor (T8), so the template's headings and commented-out examples are what they all read.

**Acceptance**

- vloop init writes .vloop/repo-guidelines.md when it does not exist and prints `wrote .vloop/repo-guidelines.md` on the line right after `wrote .vloop/install.json`
- An existing .vloop/repo-guidelines.md is kept byte for byte and init prints no line for it; init still exits 0
- The template is short: a title, one sentence saying the file holds this repository's lessons and every loop session receives it, then `## Pinning tests` and `## External formats`, each with one sentence and one example line in the format L3 (`- when `<phrase>`: `<name>`, …`) or L4 (`- `<phrase>`: `<path>``) reads, inside an HTML comment so it is not a rule
- The prose follows `language` (C-4): `vloop init --language es` writes Spanish prose; the two headings are exactly `## Pinning tests` and `## External formats` in every language
- TestInitWritesEverything (internal/cli/init_test.go) and TestWorkedExampleB5Session (cmd/vloop/b5_e2e_test.go) change only by the new `wrote .vloop/repo-guidelines.md` line; committed tests cover the template in both languages and the kept existing file. Violation: the behaviour shipped with no committed test the checks run
- vloop upgrade does not write the file (out of scope); no schema or exit code changes
- No earlier test is weakened: only the tests the brief's Constraints list change, and only as they state

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T1/gate.sh
```

</details>

### T2 — Hand .vloop/repo-guidelines.md to every session vloop run starts

`pending` · 2 attempt(s) · depends on: T1

A rule written in a skill is not a check: the gate review missed twice with its rule written down. L1 makes the driver hand the repository's lessons to every session — plan, gate review, work and review — as the file's content, through `claude --append-system-prompt-file`, so no session has to choose to read it. With no file, or an empty one, nothing changes, byte for byte.

**Acceptance**

- When .vloop/repo-guidelines.md exists in the repository root and is not empty, every session vloop run starts (plan, gate review, work, review) gets `--append-system-prompt-file .vloop/repo-guidelines.md` added to its claude arguments, as one contiguous pair
- When the file is missing or empty, every session's arguments are exactly today's. Violation: any session's arguments changing when there is no guidelines file
- The flag and its spelling were confirmed against `claude --help`; had it been absent, the task halts and says so instead of guessing
- The file is read from the repository root only, never from HOME or any global location; a file over 16 KiB is still handed over whole
- The session record is unchanged: no new field in session/v2 or any schema
- TestSessionArgs (internal/driver/session_test.go) changes only if the function it calls gains a parameter, and the arguments it pins with no guidelines file do not change; committed tests cover the flag with a non-empty file and its absence with no file and an empty file. Violation: the behaviour shipped with no committed test the checks run
- No earlier test is weakened: only the tests the brief's Constraints list change, and only as they state

**From the last attempt:** The implementation from the previous attempt is in place (WithGuidelines in internal/driver/session.go, called from Runner.Run; tests in internal/driver/session_test.go) and the gate passes, but the flag spelling could not be confirmed against `claude --help`: running claude is denied in this session. Per acceptance 3 the task halts instead of guessing. — T2 GATE PASSES while the session reports blocked — the work satisfies its own gate; the block is about something else.

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T2/gate.sh
```

</details>

### T3 — Add the guidelines paragraph to the four session skills and teach the gate review the two rules it missed

`pending` · depends on: T2

The four skills must tell their sessions what the appended guidelines are and how each applies them (L1), and the gate review must report as `unpassable` the two kinds of gate it let through in past runs (L5): one that needs a session to write under .vloop/ outside .vloop/tmp/, and one using a pattern the host's tools reject. T9 then puts every skill on a word budget, so the additions must be short.

**Acceptance**

- plugin/skills/plan, gate-review, work and review each gain one short paragraph: the repository's guidelines may be appended to the session's instructions; the planner applies them to tasks and gates and cites them in references where a task needs one; the gate review reports a gate that contradicts one under the kind it fits; the work session follows them; the review rules on a deviation as it would on the brief's constraints; a guideline never overrides the brief or a domain invariant
- In plugin/skills/gate-review/SKILL.md, `unpassable` also names a gate that needs a session to write under .vloop/ outside .vloop/tmp/ (S-4), the plan session's own gate folders aside
- Check 1 of the gate-review skill says the gate is run, and read, in the plan's shell with this host's tools: sh unless the config names another (P-5), and on macOS BSD grep, sed, awk and date; a pattern those tools reject is `unpassable`
- No other finding kind or check changes; the operate skill is not touched by this task. Violation: any skill changed beyond the paragraphs L1 and L5 name
- Every skill stays within the ceilings T9 will enforce (whole SKILL.md, wc -w): plan 4,100, gate-review 1,550, work 1,800, review 1,650, operate 1,750
- The skill structure tests in internal/cli/skills_test.go still pass unchanged

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T3/gate.sh
```

</details>

### T4 — Make vloop brief check warn when a brief changes what a listed test pins without naming the test

`pending` · depends on: T1

The same spec gap — a must-change list missing the tests that pin a list or a count — was recorded four times. L3 lets a repository list, under `## Pinning tests` in .vloop/repo-guidelines.md, which tests pin what, and makes brief check warn before the run when a brief's Behaviour contract mentions the phrase and its Constraints do not name every test. The warning is a warning: it never changes the exit code.

**Acceptance**

- vloop brief check reads the repository's .vloop/repo-guidelines.md when it exists; under `## Pinning tests` each line of the form - when `<phrase>`: `<name>`, `<name>`… is a rule; names are read before a trailing parenthesis, so a rule ending `(D20261011-1455)` has no extra name; other lines, and malformed lines, are ignored
- For each rule whose phrase the brief's Behaviour contract section contains (case-insensitive plain substring) and whose names the Constraints section does not all contain (plain substrings), one warning: `must-change list: the contract mentions "<phrase>" but Constraints do not name <name>, <name> (.vloop/repo-guidelines.md)`, the missing names in the rule's order
- Sections are found by the repository's heading set (B-3) and run through their ### subsections to the next ## heading; a rule never reads the whole brief (B-4). Violation: a rule read from the whole brief rather than its named section
- With no guidelines file, or one with no such section or lines, the output is exactly today's
- The warning never becomes a problem and never changes brief check's exit code. Violation: a warning that becomes a problem, or changes the exit code
- No regular expressions in the file: phrases and names match as plain substrings; the file is read from the repository root, never HOME
- Committed tests (internal/brief) cover the worked example's cases: the missing names in rule order, all named, the phrase outside the contract, no file, a malformed line, a Spanish brief; and the real-data check (four consumed briefs warn naming the tests their defects say they omitted) holds. Violation: the behaviour shipped with no committed test the checks run
- No earlier test is weakened: only the tests the brief's Constraints list change, and only as they state

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T4/gate.sh
```

</details>

### T5 — Make vloop brief check warn when a contract names an external format without citing its source

`pending` · depends on: T4

A brief that uses an external format, such as an eval case, without citing its definition invites a session to write the format from memory, which a recorded defect did. L4 lets a repository list, under `## External formats` in .vloop/repo-guidelines.md, which phrase needs which source, and makes brief check warn when the Behaviour contract mentions the phrase and Binding references do not cite the path.

**Acceptance**

- Under `## External formats`, each line of the form - `<phrase>`: `<path>` is a rule; the path is read before a trailing parenthesis; other lines are ignored
- When the Behaviour contract contains the phrase (case-insensitive plain substring) and the Binding references section does not contain the path (plain substring), one warning: `external format: the contract mentions "<phrase>" but Binding references do not cite <path> (.vloop/repo-guidelines.md)`
- The path cited only in another section (Constraints, What it is) does not satisfy the rule: only Binding references counts (B-4)
- With no guidelines file the output is exactly today's; the warning never becomes a problem nor changes the exit code
- Committed tests (internal/brief) cover the warning, the cited path, the phrase outside the contract, case-insensitivity and no file. Violation: the behaviour shipped with no committed test the checks run
- No earlier test is weakened: only the tests the brief's Constraints list change, and only as they state

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T5/gate.sh
```

</details>

### T6 — Add the eval case gate-review-real-misses: six gates that fooled the gate review, and one sound gate

`pending` · depends on: T3

The gate review passed six defective gates in past runs, gates whose rule was already written. L6 turns them into one eval case, built like gate-review-planted-gates, so the gate review can be measured against them from now on. The case is behaviour data for the operator's eval runs: no gate runs claude, so its scaffold must prove on the host that each planted gate fails on the base for its recorded reason.

**Acceptance**

- A new case under plugin/evals/gate-review-real-misses/, in the format of docs/guide/evals.md, beside and built like gate-review-planted-gates: prompt.md (/vloop:gate-review, max_turns, allowed_tools), case.yaml (schema_version, name, context.scaffold_script: scaffold.sh), scaffold.sh writing a repository with a brief, a state/v2 plan of seven tasks in shell "sh" and their gates, and graders/*.md
- Six tasks carry the six cited defects' gates, each reduced to its defect's mechanism (GNU-only grep pattern; top-level TOML key appended after a table; a session required to write under .vloop/; a pattern with a trailing space that cannot match; a session/v1 record validated against session/v2; a *-review.json glob that also matches *-gate-review.json); one carries a sound gate
- scaffold.sh, run in an empty directory on the host, exits non-zero unless each planted gate fails on the base for its recorded reason and the sound gate fails at its behavioural assertion; the GNU-grep gate reproduces only with BSD grep, and the evals README says beside the case that it is for macOS hosts
- Graders: a regex grader holds .vloop/tmp/gate-verdict.json's verdict to FAIL; each defective gate is a grader expecting a finding of kind `unpassable`, as its recorded defect describes it; the sound gate's grader expects no finding
- The evals README's gate-review table and docs/guide/evals.md's /vloop:gate-review row list the case; the repository's skill checks (internal/cli/skills_test.go, TestEvalGrantsCoverFence) accept it, changing only where they count or list cases. Violation: a case format, grader type or field not in docs/guide/evals.md
- No other eval case changes, and the shared _*.sh helpers change only in ways that leave every other case's scaffold output unchanged. Violation: any eval case changed beyond L6 and L7
- The case itself is the committed test of this task; its scaffold's own precondition checks are what the checks and the operator re-run

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T6/gate.sh
```

</details>

### T7 — Repair the three broken eval cases: plan-checks-its-gates' fixture grader and operate-proposes-options' scaffold and graders

`pending` · depends on: T6

Three recorded eval defects scored correct answers below 1, so these cases could not measure anything. L7 makes plan-checks-its-gates judge the planner's gate-folder fixtures from the files it wrote rather than a trace cut off before those writes, and gives operate-proposes-options a realistic halted run and graders that score what the operate skill's step names.

**Acceptance**

- plan-checks-its-gates: the grader that judges the planner's gate-folder fixtures reads the files under .vloop/state/gates/ the session wrote, through a grader focus or target naming a file (a directory focus throws, as D20261006-1028's T14 defect records), not the trace
- operate-proposes-options: the scaffold puts the plan and the iterations on a work branch named for the run id (the plan's run_id and branch), main untouched by them; the plan validates as state/v2 and has at least three tasks, the first done, the second blocked on a gate dispute
- operate-proposes-options' graders score what the skill's step names: one to three real options, one recommended with its reason, no gate amended before the operator answers, no straw option; a sound answer that does these is not failed by wording or focus
- Only these two cases change, in the format of docs/guide/evals.md; their README rows stay accurate; the skill checks accept them. Violation: any eval case changed beyond L6 and L7; a case format, grader type or field not in docs/guide/evals.md

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T7/gate.sh
```

</details>

### T8 — Add doctor's guidelines check: provenance on every rule, and a budget of 20

`pending` · depends on: T1

A rule that cites no defect or intervention cannot be judged when it is time to prune, and too many rules has hurt before (L9). This task defines a rule and its provenance and adds a doctor check, printed right after `gate edits`, that counts the rules and warns past 20 or on a rule with no ID. Both are warnings: the operator decides what to retire, and no exit code changes.

**Acceptance**

- A rule is a line of .vloop/repo-guidelines.md starting `- ` at the left margin, in any section; a line inside an HTML comment (single- or multi-line) is not one. Provenance is a token D<8 digits>-<4 digits> or I<8 digits>-<4 digits>, optionally followed by the rest of the ID
- doctor prints a check `guidelines` right after `gate edits`, in text and as the next entry in --json: no file `- guidelines`; within budget with provenance everywhere `✓ guidelines <n> rule(s)`
- Over 20 rules: `! guidelines <n> rules, over the budget of 20 — retire or merge some before adding more`; rules without provenance: `! guidelines <k> rule(s) cite no defect or intervention: line <a>, line <b>` with 1-based line numbers; both: one warning line, the two messages joined by `; `, the second keeping its `guidelines` (e.g. `! guidelines 21 rules, over the budget of 20 — …; guidelines 2 rule(s) cite no defect or intervention: line 5, line 6`)
- The budget is 20 and not configurable; both are warnings: no exit code changes and doctor writes nothing
- TestDoctorWorkedExampleText (internal/cli/doctor_test.go), TestWorkedExampleB5Session, TestWorkedExampleB11 (cmd/vloop/b20261005_e2e_test.go) and TestWorkedExampleB13 (cmd/vloop/b13_e2e_test.go) change only by the new guidelines line and, where pinned, the warning count; committed tests cover the template (0 rules), the seed (6), seed plus 15 (21), stripped IDs, both, a commented rule and no file. Violation: the behaviour shipped with no committed test the checks run
- No earlier test is weakened: only the tests the brief's Constraints list change, and only as they state

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T8/gate.sh
```

</details>

### T9 — Put the plugin skills on word budgets enforced by a test

`pending` · depends on: T3

Skills grow by drift; L10 makes growth a recorded decision by a failing test. The test counts each plugin skill's whitespace-separated words, frontmatter included, against a ceiling that lives beside it with a comment saying when and why it was set. No gate may run the repository's tests, so the gate judges the counting through one function this plan pins, `Check(pluginDir string) []string` in package `internal/budget`, which it calls from its own oracle on synthetic skills; the committed test is the review's to rule on, and calls the same function on this repository's plugin/.

**Acceptance**

- Package internal/budget exports `func Check(pluginDir string) []string`: for each of the five skills, it counts the whitespace-separated words of `<pluginDir>/skills/<skill>/SKILL.md`, the whole file with its frontmatter, as wc -w counts them, and returns one message per skill over its ceiling, `<skill>: <count> words, over its ceiling of <ceiling>`, with thousands separators (e.g. `plan: 4,213 words, over its ceiling of 4,100`); a skill exactly at its ceiling is not reported
- The ceilings are plan 4,100, gate-review 1,550, work 1,800, review 1,650, operate 1,750, each beside the test with one comment saying when and why it was last set
- A committed Go test, outside cmd/, calls budget.Check on this repository's plugin/ and fails with each message it returns, and passes on the tree as this task leaves it. Violation: the budget shipped with no committed test the checks run, or a test that does not fail when a skill grows past its ceiling
- Raising a ceiling is an edit of that comment and number. Violation: a ceiling raised, or a skill trimmed, to make room, beyond what L10 sets
- The word counting is one function the metrics task (T10) can share, not a second copy

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T9/gate.sh
```

</details>

### T10 — Add the instructions line to vloop metrics: the words each session kind starts with

`pending` · depends on: T8, T9

Each session reads its skill, the guidelines and CLAUDE.md before doing anything; L11 makes that load visible per brief, back to B1, so its trend shows. It is computed from git at the commit a brief's delivered lines are measured to (M-2), never recorded during a run, and adds one line to the brief summary and one optional object to the metrics/v2 JSON, frozen by brief close.

**Acceptance**

- For each brief, at the commit its delivered lines are measured to (M-2), each session kind's figure is its skill + .vloop/repo-guidelines.md + the root CLAUDE.md, each counted as L10 counts (wc -w) and 0 when absent; a kind's skill is plugin/skills/<kind>/SKILL.md, else for plan, work and review the shell loop's .claude/skills/loop-<kind>/SKILL.md, else the kind is n/a; the rule count is the guidelines file's rules (L9)
- The brief summary gains, right after the interventions line: ` instructions  plan <n> · gate review <n> · work <n> · review <n> words · guidelines <n> rules`, numbers with thousands separators, n/a for a missing kind
- --json gains an optional object `instructions`: {"plan": n|null, "gate_review": n|null, "work": n|null, "review": n|null, "guidelines_rules": n}; metrics/v2 gains it as an optional property with no version change; brief close's snapshot carries it
- Computed from git (M-1); nothing is recorded during a run; no instructions column in the all-briefs table
- Tests that pin the brief summary line by line (internal/metrics/print_test.go, internal/cli/metrics_test.go, cmd/vloop/b3_e2e_test.go, cmd/vloop/b4_e2e_test.go) change only by the instructions line; committed tests cover the worked example's temporary repository (100- and 50-word skills, n/a kinds, 6 rules), the shell-loop fallback and --json. Violation: the behaviour shipped with no committed test the checks run
- vloop metrics exits 0 for every consumed brief in this repository, and B13's plan figure equals git show <end>:plugin/skills/plan/SKILL.md | wc -w plus git show <end>:CLAUDE.md | wc -w
- No earlier test is weakened: only the tests the brief's Constraints list change, and only as they state

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T10/gate.sh
```

</details>

### T11 — Add rule and rule-outcome to defects, the defect report to brief close, and the triage to the operate skill

`pending` · depends on: T9

Close is where defects are recorded, so L12 asks there, every brief, which rule caught each defect or should have. Two optional defect fields record it; close prints a report of the brief's defects with their rule, outcome or untriaged, and recurrences across briefs; the operate skill's close step proposes the triage. Close reports, never refuses: whether a defect deserves a rule is the operator's call.

**Acceptance**

- vloop defect set <id> rule '<where>: <which>' accepts <where> in guideline, brief-check, gate, gate-review, review, skill; vloop defect set <id> rule-outcome accepts fired, missed, absent, none; anything else exits 2 and leaves the record byte-identical
- defect/v1 gains `rule` and `rule-outcome` as optional properties: no schema version changes and records without them stay valid; vloop defect set's help lists the two fields and docs/guide/commands.md is regenerated
- vloop brief close prints, after recording the findings and before its `squash-merge with the trailer` line (which stays last), `defects of <run id>` then one line per defect recorded for the brief: `  <id>  <kind>  <found-by>  <rule or —>  <rule-outcome or untriaged>[  recurs ×<n> (<id>, <id>)]`; then, when any is untriaged, `<n> untriaged: vloop defect set <id> rule '<where>: <which>' and rule-outcome fired|missed|absent|none`, <n> the number of untriaged defects and the rest literal; with no defects, `defects of <run id>: none`
- recurs lists the other defects, of any brief, with the same rule, by short ID (D<stamp>), oldest first, at most three, with the total count; an untriaged defect never recurs
- --dry-run prints the report as it would be and writes nothing; --json is unchanged; close never refuses for an untriaged defect. Violation: a close that refuses, or a report that writes anything, for an untriaged defect
- plugin/skills/operate/SKILL.md's close step gains the triage: for each untriaged defect propose its rule and rule-outcome as options and record the operator's choice; for absent, draft the guideline line with its provenance, or the check; for a gate-review rule that missed, propose an eval case and the defect's case:. The skill stays within its 1,750-word ceiling and changes nowhere else
- TestCloseCommitsFindingsSnapshotAndBrief (internal/cli/close_test.go) and B4's worked-example test (cmd/vloop/b4_e2e_test.go) change only by the report's lines before the trailer; committed tests cover the two fields, both refusals, the report with rule, untriaged and recurs, no defects and --dry-run. Violation: the behaviour shipped with no committed test the checks run
- No earlier test is weakened: only the tests the brief's Constraints list change, and only as they state

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T11/gate.sh
```

</details>

### T12 — Document the repository guidelines, the two warnings, the budget, the instructions line and the defect report

`pending` · depends on: T5, T8, T10, T11

Everything this brief adds must be findable: L8 gives the guide a `## Repository guidelines` section and lists the file in the .vloop/ layout, adds the two warnings to the brief domain doc, and carries L9–L12 into the domain docs they extend. A lesson there replaces a constraint repeated brief after brief, which is the point of the whole brief.

**Acceptance**

- docs/guide/concepts.md gains `## Repository guidelines`: what the file is; that vloop run hands it to every session; the two sections and their line formats (- when `<phrase>`: `<name>`, … and - `<phrase>`: `<path>`); the two warnings (`must-change list:` and `external format:`); that a rule cites its defect or intervention and vloop doctor warns past a budget of 20 rules or on a rule with none; that a lesson there replaces a constraint repeated brief after brief; and `vloop defect set <id> fixed-by guideline` for a defect a guideline now prevents
- The concepts guide's `## The .vloop/ layout` list gains a `.vloop/repo-guidelines.md` entry
- docs/domain/briefing/brief.md lists the two warnings with the others
- docs/domain/measurement/defect.md describes the optional rule and rule-outcome fields (the six <where> values; fired, missed, absent, none) and the close report with untriaged; docs/domain/measurement/metrics.md describes the instructions line and the optional `instructions` object with guidelines_rules
- docs/guide/commands.md is regenerated if any help text changed; the guide and README tests pass; every link added resolves
- No statement contradicts the behaviour T1–T11 shipped; no plugin skill is edited

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T12/gate.sh
```

</details>

### T13 — Close: commit the worked example as an end-to-end test and pass the real-data checks

`pending` · depends on: T6, T7, T12

The brief's worked example and real-data checks are the acceptance of the whole brief. This task commits them as a test the checks run, line for line, and confirms against this repository's own history: with the seed, brief check warns on four consumed briefs naming the tests their recorded defects say they omitted, and vloop metrics gives every consumed brief an instructions line whose figures git confirms.

**Acceptance**

- A committed end-to-end test (cmd/vloop) runs the brief's worked example line for line: init's new line and kept file, the brief check warnings and their absence, the run's session arguments with a non-empty, empty and missing guidelines file, doctor's guidelines lines, the metrics instructions line and --json, defect set's refusals, and close's report with recurs, untriaged and none
- The test builds its repositories in temporary directories with a temporary HOME, addresses them with git -C, never touches the real HOME, and threads computed values (the root, the binary's path, word counts) rather than hardcoding them
- Real-data: with the seed in a temporary repository, brief check on copies of B20261003-2049-gate-model, B20261004-1434-interventions-options, B20261008-2144-b12-field-reliability and B20261011-2100-b13-docs-onboarding set to ready warns naming TestWorkedExampleB2Commands; TestSchemasAreEmbedded and TestWorkedExampleB6RealData; TestConfigListDefaults, TestConfigListJSONOrderAndNull and TestDefaultsEveryRow; and TestInitPointsAtBothWaysToLoadThePlugin respectively
- Real-data L11: vloop metrics for B13 and B1 prints the plan figure git gives at the commit their delivered lines are measured to, B1 with gate review n/a; vloop metrics exits 0 for every consumed brief
- This task does not write this repository's .vloop/repo-guidelines.md (the operator's, at close) nor anything else under .vloop/ outside .vloop/tmp/
- No production code changes beyond the smallest fix for a defect this task finds, stated in the report
- No earlier test is weakened: only the tests the brief's Constraints list change, and only as they state

<details><summary>verify command</summary>

```sh
sh .vloop/state/gates/T13/gate.sh
```

</details>


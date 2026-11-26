---
name: B20261126-1418-b45-docs-mixed-os-and-any-agent.loop-brief
description: Bring the user documentation up to date with what B40–B44 shipped — a README section on mixed macOS/Windows teams, the guides worded for any agent rather than Claude only, what the Copilot model mai-code-2 has run, with its evidence, and this repository's own operator notes on gates
kind: brief
status: ready
created: 2026-11-26
seeds: The operator's request on 2026-11-26 to update all the documentation, with at least a brief paragraph on mixed-OS support and generic wording where it assumes Claude only; a read-only survey of README.md and docs/guide the same day; the evidence of B40–B44, every session kind they ran on Copilot mai-code-2
depends-on: [B20261126-1036-b44-i98-self-update.loop-brief]
---
# Brief — vloop B45: the documentation for mixed-OS teams and any agent

- **Starting point:** extends `main` as it stands when the run starts (after
  B44's merge), with this brief's ready commit on the run branch. The planner
  pins that SHA as the **base** every gate compares against — never `HEAD~n`.
- **Produced by:** operator request, 2026-11-26; the open decisions settled by
  the operator's hands on the recommended options, recorded below.
- **Review:** `vloop brief review` (Copilot, mai-code-2), thirteen rounds.
  Fixed: the cost figures are the metrics' `cost.total_usd` (B42's includes its
  halted attempt) and B39's $8.10 is B39 itself, not the same brief; D6 comes
  first; D2 says the lint uses a fixed list; `mai-code-2` is not called
  qualified, per `docs/guide/evals.md`'s rule quoted exactly; the session
  lists; a real-data gate for D6; the 503 halt's record bound. Accepted by the
  operator's hands (the last `ASK`): the URL shortener figures (10/10, all
  first-pass, $1.31) are the operator's reading of that acceptance
  repository's metrics, stated as such in D6.
- **Roadmap:** the row for B45 in `docs/design-notes/vloop-roadmap-b18-b21.md`,
  pointing at this brief, is added in its ready commit and marked consumed at
  close by the operator.
- **Run environment:** planned and run on Windows, in pwsh, every session on
  GitHub Copilot (`.vloop/config.local.toml`). See Constraints.

## Executive summary

The user documentation lags what vloop does. README.md calls vloop "a Go
command-line tool and Claude Code plugin", says "vloop has Claude plan it", and
"Every session is a fresh Claude session", although every session kind can run
on GitHub Copilot CLI. Mixed macOS/Windows support (B40, B42, B43) appears in
README.md only as a bullet inside the Copilot section, which also still says "A
whole Copilot run on Windows is untested": B40 to B44 each ran whole on
Copilot, on Windows. Several guide sentences name only Claude's mechanism or
command where both agents have one.

**The goal:** a reader of README.md and docs/guide learns that vloop works with
either agent, how a mixed macOS/Windows team sets it up, and what `mai-code-2`
has run (not yet qualified by the eval rule), with nothing that is no longer
true.

- **D1 — README.md, agent-neutral.** The opening, prerequisites, quickstart,
  "The loop in one paragraph" and "You ask; the skill drives" name both agents
  where both apply.
- **D2 — README.md, mixed teams.** A short section of its own, the Copilot
  section's Windows bullet trimmed to what is Copilot's, and the PowerShell
  PATH pointer in Install.
- **D3 — README.md, the model in use.** What `mai-code-2` has run, with the
  evidence, without calling it qualified (the eval rule is not met yet).
- **D4 — the guides.** concepts.md, metrics.md, configuration.md and evals.md
  lose their Claude-only sentences where both agents apply, and concepts.md
  gains the mixed-team sentences.
- **D5 — what vloop writes and this repository's own notes.** The CLAUDE.md
  section vloop writes names every kind of session the loop starts; this repository's operator
  and architect skills say that its gates are pwsh now.

Unchanged: every behaviour, flag, message and test outcome; the sections that
are rightly Claude-specific (listed below); docs/domain and docs/design-notes
except where named.

## What it is

A documentation pass, no code change but the text vloop writes into a
repository's CLAUDE.md. It is for anyone reading the README or the guides to
decide whether and how to use vloop, on either agent and on more than one OS.

## Why this shape, and what was rejected

Decided on 2026-11-26. **Do not re-litigate these.**

- **Name both agents where both apply; leave Claude-specific text alone.** The
  sections about the Claude adapter, Claude's install commands next to their
  Copilot equivalents, the `claude`/trust prerequisites already scoped "when a
  session runs on Claude", `claude plugin eval` where it is the Claude path,
  and the legacy "count as `claude`" rule are correct and stay. *Rejected:*
  rewriting every mention of Claude (it would blur where the two differ).
- **Mixed teams get their own README section, short, linking the guide.**
  `docs/guide/configuration.md` "Mixed teams" stays the full treatment.
  *Rejected:* a long README section (it would drift from the guide);
  leaving it inside the Copilot section (the rules apply whatever the agent).
- **The model claim states its evidence and stops short of "qualified".**
  `docs/guide/evals.md`'s rule decides qualification, and `mai-code-2` has not
  been through it. *Rejected:* calling it qualified from the runs alone
  (that contradicts the rule); leaving it out (the runs are real evidence).
- **This repository's operator and architect skills are in scope**, because
  this repository is now a mixed team whose gates are pwsh, and their "gates
  run on macOS's BSD tools" line now misleads a session.

## Binding references

- `docs/domain/domain-model.md` — S-1 (a session is a fresh process; the README and concepts.md say so for either agent), P-5 (the plan's shell governs its gates) and C-5 (the mixed-team rule D2 and concepts.md summarise)
- `docs/design-notes/copilot-cli-experiments.md` — D6: the design note on Copilot that records each model's qualification
- `.vloop/interventions/I20261126-005105-copilot-model-endpoint-503s-ended-t4-s-w.md` — D6: the record of B40's halt on Copilot's HTTP 503s
- `README.md` — the document D1, D2 and D3 change
- `docs/guide/concepts.md` — D4: "Every session is a fresh Claude session", the repository guidelines paragraph, the interactive `--plugin-dir` paragraph, the evals sentence, and "Gates and the gate shell"
- `docs/guide/metrics.md` — D4: "Cost and tokens", which says nothing about Copilot's credits
- `docs/guide/configuration.md` — D4: the `model.*` rows; "Mixed teams", the section README.md links to and stays the full treatment
- `docs/guide/evals.md` — D4: the table row "Calls `claude`"
- `internal/install/section.go` — D5: the CLAUDE.md section vloop writes, whose session list omits verify
- `.claude/skills/vloop-operator/SKILL.md` — D5: the line on BSD tools in gates
- `.claude/skills/vloop-architect/SKILL.md` — D5: the constraint on BSD tools every brief carries
- `internal/driver/session.go` — how Claude sessions get the repository guidelines (`--append-system-prompt-file`)
- `internal/driver/agent_copilot.go` — how Copilot sessions get them (`COPILOT_CUSTOM_INSTRUCTIONS_DIRS`) and how their cost is estimated from credits

## Behaviour contract

The exact wording is the writer's. What each item must leave true, and what
must no longer appear, is below; the gates check the literal strings named.

### D1 — README.md, agent-neutral

- The opening says vloop is a command-line tool and a plugin for Claude Code
  and GitHub Copilot CLI, and that an agent plans the brief. The string
  `vloop has Claude plan it` no longer appears, and the opening's first
  sentence names both Claude Code and GitHub Copilot CLI.
- Prerequisites gains a bullet: the `copilot` CLI and a GitHub login its
  sessions can use, when a session runs on Copilot (experimental), linking
  the Copilot section. The two Claude bullets stay as they are.
- Quickstart's doctor step says it checks the agent CLIs in use (`claude` and
  workspace trust, or `copilot` and its GitHub login), not `claude` alone.
- "The loop in one paragraph" says every session is a fresh agent session
  (Claude Code or Copilot CLI); `Every session is a fresh Claude session` no
  longer appears in README.md.
- "You ask; the skill drives" says to ask your agent (Claude Code or Copilot
  CLI); its first sentence no longer names Claude Code alone.

### D2 — README.md, mixed teams

- A section headed `## Mixed macOS and Windows teams`, after "Quickstart" and
  before "Running sessions on GitHub Copilot", of at most 150 words, says:
  the team declares its operating systems in the shared
  `.vloop/config.toml` with `platforms = ["macos", "windows"]` and
  `shell = "pwsh"`, and vloop refuses a local or environment override of
  either; with two or more platforms, vloop refuses a plan whose gates call a
  command from its fixed list of tools some platforms lack (`grep`, `sed` and
  the like), so gates use pwsh cmdlets and the project's own toolchain; vloop runs pwsh gates and checks in UTF-8; this
  repository is such a team. It links
  `docs/guide/configuration.md#mixed-teams`.
- The Copilot section's "Windows and mixed teams" bullet becomes a Windows
  bullet that keeps only what is Copilot's (the hook has a PowerShell form)
  and says Copilot runs on Windows were used for B40 to B44, linking the new
  section. `Windows is untested` no longer appears.
- Install gains, after the `export PATH` block, one sentence: on Windows, the
  PowerShell line is in the same concepts.md section; and one more: to update
  later, `vloop self-update` installs the latest release and updates the
  plugins (B44), then `vloop upgrade` in each repository.

### D3 — README.md, the model in use

`docs/guide/evals.md`'s qualification rule ("A model qualifies for a phase on
an agent when, at 3 runs, every case of that phase scores a mean no lower than
the phase's default Claude model scores on the same tool with the same runs") is unchanged, and `mai-code-2` has not been through it. So the README
does not call it qualified. The Copilot section adds one sentence after what
it says of `gpt-6.1-terra`: `mai-code-2` has run every session kind but the
second review (plan, gate review, work, review, verify and brief review) in
the URL shortener acceptance run and in B40 to B44, which built vloop itself,
and is not yet qualified by the eval rule; it links D6's section. The words
`mai-code-2 is qualified` do not appear anywhere. The sentence "keep plan and
gate review on Claude, and verify is not yet tried" stays true of
`gpt-6.1-terra`, reworded only if needed to say so.

### D4 — the guides

- **concepts.md.** `Every session is a fresh Claude session` becomes a fresh
  agent session (Claude Code or Copilot CLI, per `agent.<kind>`). The
  repository guidelines paragraph says vloop hands the file to every session
  it starts as extra instructions (no list of kinds): Claude Code's appended system prompt, Copilot CLI's
  custom-instructions directory. The interactive `--plugin-dir` paragraph
  shows both `claude --plugin-dir "$(vloop plugin path)"` and `copilot
  --plugin-dir "$(vloop plugin path)"`. The evals sentence adds that `go run
  ./tools/agent-eval` replays the cases on other agents (see evals.md). "Gates
  and the gate shell" gains two or three sentences: a team on more than one
  OS sets `platforms` and a shared `shell = "pwsh"` and writes portable gates,
  linking `configuration.md#mixed-teams`.
- **metrics.md, "Cost and tokens".** Adds that Copilot sessions report no
  tokens; their cost is estimated from their credits at
  `copilot.credit-usd` dollars each (`cost_source` `estimated`), and
  `credits.total` sums the credits of every phase; and that
  `cache_hit_ratio` counts Claude sessions only.
- **configuration.md, the `model.*` rows.** The "Values" cell `any model name`
  becomes `a Claude alias or model name; a Copilot model id when agent.<kind>
  is copilot` (seven rows).
- **evals.md.** The table row that reads "Calls `claude`" says which agent
  each column calls (`claude plugin eval` calls `claude`; `tools/agent-eval`
  calls the agent it is given).

### D5 — what vloop writes, and this repository's notes

- `internal/install/section.go`: `(plan, gate review, work, review)` becomes
  `(plan, gate review, work, review, second review, verify)`. Any test that pins the old
  section text changes to the new text, and this repository's own CLAUDE.md
  section is edited to match it (a session may not run `vloop upgrade`: it
  writes under `.vloop/`).
- `.claude/skills/vloop-operator/SKILL.md`: the "Gates run on macOS's BSD
  tools" bullet says this repository is a mixed team (`platforms = ["macos",
  "windows"]`, `shell = "pwsh"`), its gates are pwsh and portable, and the
  BSD rules that follow apply only to a plan whose shell is `sh`.
- `.claude/skills/vloop-architect/SKILL.md`: the constraint "gates run on
  macOS's BSD tools" becomes "gates are pwsh and portable (a mixed team: macOS
  and Windows)".

### D6 — the evidence for `mai-code-2`

`docs/design-notes/copilot-cli-experiments.md` gains a section headed
`## mai-code-2 in the loop (2026-11)` that records, as the operator measured
them from the runs' telemetry: the URL shortener acceptance run, 10 of 10
tasks, all first-pass, about $1.31 (the operator's measurement from that
acceptance repository's own metrics, outside this repository; the section
says so, and no session here checks it); and this
repository's B40
(6 tasks, 3 first-pass, about $2.18), B41 (3 tasks, 2 first-pass, about
$1.03), B42 (5 tasks, 2 first-pass, about $3.05, its first attempt, halted on
a fence gap, included), B43 (3 tasks, 2 first-pass, about
$0.75) and B44 (5 tasks, 3 first-pass, about $1.94), every session kind they ran (plan, gate review, work, review and
verify; no second review ran) on `mai-code-2`. It also records that every brief review of B40 to
B45 ran on `mai-code-2` (their kept reviews are under
`.vloop/state/brief-reviews/`). Each of B40's to B44's figures is its `vloop metrics`
`cost.total_usd` (in `.vloop/state/metrics/`), all attempts included. For
scale, not as a like-for-like comparison: B39, a brief of this repository of
similar size, cost $8.10 on Claude (Opus plan, Sonnet for the rest), per its
run record. It says what the runs showed beyond the numbers: Copilot's model
endpoint returned HTTP 503s that halted B40 once (exit 7; the task's attempt
was not counted);
the reviews and gate disputes caught real defects; and the operator's own
verification still found defects in B42 and B43. D3's README sentence is
backed by this section and links it.

### Decided here, because the cited documents leave it underdetermined

- **Line numbers are not pinned**: the documents moved during B40–B44; each
  change is located by the text named above.
- **The Spanish templates and docs/domain are out of scope**, except that
  nothing here may contradict them.

### Violations the review must rule on

- A changed behaviour, message, flag, default or test outcome, except D5's
  section text and the test pinning it.
- Text made generic where it is really Claude-only (the list under "Why this
  shape").
- A claim about `mai-code-2` broader than D3's, or the word qualified applied
  to it.
- A README section on mixed teams over 150 words, or one that contradicts
  `docs/guide/configuration.md`.

## Worked example

From the repository root, after the run (pwsh):

```
Select-String -Path README.md -Pattern 'vloop has Claude plan it','Every session is a fresh Claude session','Windows is untested' -SimpleMatch
  -> (no output)
Select-String -Path README.md -Pattern '## Mixed macOS and Windows teams' -SimpleMatch
  -> README.md:<n>:## Mixed macOS and Windows teams
Select-String -Path README.md -Pattern 'mai-code-2','vloop self-update' -SimpleMatch
  -> at least one line for each
Select-String -Path README.md -Pattern 'mai-code-2 is qualified' -SimpleMatch
  -> (no output)
Select-String -Path docs/guide/concepts.md -Pattern 'Every session is a fresh Claude session' -SimpleMatch
  -> (no output)
Select-String -Path docs/guide/concepts.md -Pattern 'copilot --plugin-dir' -SimpleMatch
  -> at least one line
Select-String -Path docs/guide/configuration.md -Pattern '| any model name |' -SimpleMatch
  -> (no output)
Select-String -Path internal/install/section.go -Pattern '(plan, gate review, work, review, second review, verify)' -SimpleMatch
  -> one line
Select-String -Path .claude/skills/vloop-operator/SKILL.md -Pattern 'shell = "pwsh"' -SimpleMatch
  -> at least one line
```

And the mixed-teams section's word count, from the line after its heading to
the next `## `, is at most 150.

## Out of scope

- Any code change but D5's section text; any new command or flag.
- Translating anything into Spanish.
- docs/domain, docs/design-notes but D6's section, and the briefs.
- Closing issue #84 or commenting on it (the operator's).
- The README's "What changed in 0.8" section and the command reference
  (`docs/guide/commands.md`, generated).

## Constraints

- Go, as pinned in `go.mod`; `go run ./tools/check` is the `[[check]]` and
  passes after every task (the word budgets of the skills included).
- **This run is on Windows**, in pwsh. Gates are pwsh, using `Select-String`,
  `Test-Path` and `Get-Content`, never a listed POSIX tool (this repository is
  a mixed team).
- **Gates follow the diff surface.** No gate runs the whole suite or the
  check (P-8).
- **No task may weaken a check to pass.** The only test that may change is
  the one pinning the CLAUDE.md section text, and only to the new text.
- Repo-relative paths everywhere. Sessions never move refs.
- **Prose tasks:** a task that fails review twice on wording rather than on
  its gate is blocked, and the operator finishes it by hand.

## Shape

5 to 6 tasks:

1. **D6, the design-note section** (first, so D3's link has a target). Gated
   on its heading, on each run's name (B40 to B44) appearing in it, and on a
   real-data check: for each of B40 to B44, the section's task count,
   first-pass count and dollar figure equal that brief's
   `.vloop/state/metrics/<brief>.json` (`tasks.planned`, `tasks.first_pass`,
   `cost.total_usd` rounded to cents), read by the gate, not hardcoded; and
   B39's $8.10 appears in B39's run record.
2. **D1 and D3 in README.md.** Gated on D1's and D3's strings.
3. **D2 in README.md.** Gated on the heading, the absent strings, the link and
   the word count.
4. **D4 in the guides.** Gated on D4's strings.
5. **D5**, the section text with its test, this repository's CLAUDE.md, and
   the two skills. Gated on D5's strings.
6. **Close:** every earlier gate at once, and the worked example line by line.

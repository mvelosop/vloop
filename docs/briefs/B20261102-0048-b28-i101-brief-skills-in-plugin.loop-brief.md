---
name: B20261102-0048-b28-i101-brief-skills-in-plugin.loop-brief
description: Ship the brief-writing and brief-review skills in the plugin as /vloop:create-brief and /vloop:review-brief, with create-brief's guidelines and templates, inside the word budget, held to the binary's brief template, named by the CLAUDE.md section and covered by evals
kind: brief
status: ready
created: 2026-11-02
issue: 101
seeds: Issue 101 under 75, the prerequisite of 94; the two experimental project skills under .claude/skills/ since 93, decided by the assistant under the operator's delegation of 2026-11-01
depends-on: [B20261101-2305-b27-i77-copilot-plan.loop-brief]
---
# Brief — vloop B28: the brief skills in the plugin

- **Issue:** #101 (under #75); the prerequisite of #94.
- **Starting point:** extends `main` at `8a1ae1e`, plus this brief's ready
  merge. The planner pins that SHA as the **base** every gate compares
  against — never `HEAD~n`.
- **Produced by:** the design act of 2026-11-02: the two project skills, the
  plugin's structural checks run on a copy of the plugin holding them (they
  pass as they are), the binary's brief templates compared with the skill's,
  and the decisions below, taken by the assistant under the operator's
  delegation for the chain B26–B29.
- **Review:** ASK — 9 findings (3 runnability, 4 consistency, 1 decision,
  1 rule), all fixed; the two that asked for the operator (how the
  create-brief eval grades, and fixing step 5.1 in the move) decided by the
  assistant under the delegation, recorded in "Decided here";
  vloop-review-brief, 2026-11-02.

## Executive summary

Writing a brief and reviewing it have skills, `vloop-create-brief` and
`vloop-review-brief`, but they live in this repository's `.claude/skills/`,
experimental: no other repository using vloop gets them, and #94 cannot run a
brief review on a configured agent until the skill ships.

**The goal:** every repository that loads the vloop plugin can write a brief
with `/vloop:create-brief` and have it reviewed with `/vloop:review-brief`,
and the CLAUDE.md section vloop writes points an interactive session to them.

- **K1 — the skills ship in the plugin.** plugin/skills/create-brief/
  (`SKILL.md`, `guidelines.md`, `templates/brief-en.md`,
  `templates/brief-es.md`) and plugin/skills/review-brief/ (`SKILL.md`),
  named for the plugin, no longer experimental, create-brief naming
  `/vloop:review-brief` and `/vloop:operate`, review-brief naming
  `/vloop:create-brief`; within word ceilings of their own.
- **K2 — the templates cannot drift.** A test holds each plugin template to
  the binary's brief template, which `vloop brief new` writes, plus the
  executive summary.
- **K3 — the CLAUDE.md section names them.** The section `vloop init` and
  `vloop upgrade` write sends an interactive session to `/vloop:create-brief`
  to write a brief and to `/vloop:review-brief` to review it.
- **K4 — evals and docs.** One eval case per skill; the README, the concepts
  guide, the evals guide and the domain pages name the two skills.

Unchanged: the six skills already in the plugin (but `/vloop:operate`'s
description), the fence and every session `vloop run` starts; `vloop brief
new` and `vloop brief check`; the binary's templates.

## What it is

At the end the plugin has eight skills. The two new ones are the operator's,
used in an interactive session like `/vloop:operate`: `/vloop:create-brief`
turns the operator's intent into a draft brief and has it reviewed;
`/vloop:review-brief` reviews one in a session that did not write it. vloop's
binary still takes the brief as its only input; it never starts these skills.

## Why this shape, and what was rejected

Decided on 2026-11-02 under the operator's delegation. **Do not re-litigate
these** — the planner inherits them.

- **Issue #101's scope, as written**: both skills, create-brief's guidelines
  and its two templates, into `plugin/skills/`. The folder names are
  `create-brief` and `review-brief`, so the commands are
  `/vloop:create-brief` and `/vloop:review-brief`. *Rejected:* keeping the
  `vloop-` prefix — the plugin's namespace already says vloop
  (`/vloop:vloop-create-brief`).
- **No experimental label.** The skills have written and reviewed every brief
  since B25; the plugin's other skills carry none. *Rejected:* shipping them
  as experimental — nothing in the plugin would read the label, and #94
  builds on them.
- **They are interactive, as `/vloop:operate` is**: no fence, no session of
  `vloop run`. Running the review as a session is #94.
- **Word ceilings of their own**, as every plugin skill has (L10): their size
  when they join the plugin, plus headroom: `create-brief` 1200,
  `review-brief` 1000 (1078 and 904 words at the base). The ceilings count
  `SKILL.md` only, as for the other skills; `guidelines.md` and the templates
  are read on demand. *Rejected:* counting the guidelines and templates —
  no other skill's supporting files are counted.
- **The plugin templates stay a copy, held by a test**, not generated: each
  is the binary's template (`internal/brief/templates/<language>.md`) with
  the executive-summary section added. *Rejected:* making `vloop brief new`
  write the executive summary — it changes the binary's template and
  `vloop brief check`, which is not this brief's; *rejected:* generating the
  plugin templates at build time — a skill's files are read from the plugin
  folder as committed.
- **The domain says the plugin helps with the design act** (glossary, the
  Brief's paragraph): the design act stays the operator's and upstream of
  the binary, and the plugin offers two interactive skills for it. The brief
  stays vloop's only input.
- **What is particular to this repository stays out of the plugin**: the
  roadmap row, invariants cited by ID, this repository's constraints — in
  `.claude/skills/vloop-architect`. Removing the two project copies and
  pointing this repository's own skills and `CLAUDE.md` at the plugin skills
  is the operator's, at close (below), since the loop does not edit the files
  that bind it.
- **A project skill named `create-brief` or `review-brief` becomes a clash**
  for the Copilot checks of doctor and the run preflight, and the fence hook
  allows the two names as vloop's skills: both follow from the embedded
  plugin's skill list (`driver.PluginSkillNames`, the judge's skill set) with
  no code change. Accepted as the existing rule for every plugin skill.
- **`/vloop:operate`'s description names `/vloop:create-brief`** for writing
  briefs, in place of "Writing briefs is not this skill's job."; its body is
  unchanged (1749 words, one under its ceiling of 1750).
- **The binary never calls the network or a model**; the eval run is the
  operator's, at close.

## Binding references

- `.claude/skills/vloop-create-brief/SKILL.md` — the skill that moves, with `guidelines.md` and `templates/` beside it
- `.claude/skills/vloop-review-brief/SKILL.md` — the skill that moves
- `internal/brief/templates/en.md` — the binary's English brief template, the plugin template's base
- `internal/brief/templates/es.md` — the binary's Spanish brief template
- `internal/budget/budget.go` — the skill list and the word ceilings (L10)
- `internal/cli/skills_test.go` — the plugin's structural checks and `TestEvalGrantsCoverFence`, which the new skills and cases must pass
- `internal/install/section.go` — the CLAUDE.md section `vloop init` and `vloop upgrade` write
- `internal/cli/readme_test.go` — `TestReadmeNamesTheSkills`
- `plugin/skills/operate/SKILL.md` — the playbook whose description changes
- `plugin/evals/README.md` — the eval cases' structure and table
- `docs/guide/evals.md` — the eval guide: case format, grader types, the operator's `--allow-tools` list
- `docs/domain/domain-model.md` — B-1: only loop briefs are vloop's input, which stays; the Brief's paragraph and the glossary's design act, which change
- `cmd/vloop/b7_e2e_test.go` — `TestWorkedExampleB7PluginPath`, which pins the extracted skill list
- `docs/domain/execution/session.md` — "Skills, as built"
- `docs/guide/concepts.md` — its Skills section
- `README.md` — its Skills section

## Behaviour contract

### K1 — the skills in the plugin

After the run, `plugin/skills/` holds eight skill folders: `create-brief`,
`gate-review`, `operate`, `plan`, `review`, `review-brief`, `verify`, `work`.

- plugin/skills/create-brief/ holds `SKILL.md`, `guidelines.md`,
  `templates/brief-en.md` and `templates/brief-es.md`, from
  `.claude/skills/vloop-create-brief/`. plugin/skills/review-brief/ holds
  `SKILL.md`, from `.claude/skills/vloop-review-brief/`. Their instructions
  stay as they are but for the changes below.
- Frontmatter `name: create-brief` and `name: review-brief`. Each
  `description` drops its leading "Experimental, not yet in the plugin — "
  and keeps the rest.
- No file under `plugin/skills/` contains `vloop-create-brief`,
  `vloop-review-brief`, or the words "experimental" or "not yet in the
  plugin" (case-insensitive) after the run.
- `create-brief`'s `SKILL.md` names `/vloop:review-brief` for the independent
  review (its step 5) and `/vloop:operate` for marking ready and running (its
  step 6). `review-brief`'s `SKILL.md` names `/vloop:create-brief` as the
  skill briefs are written with, and as the standard "the `guidelines.md`
  beside `/vloop:create-brief`'s `SKILL.md` (`../create-brief/guidelines.md`
  from this skill's folder)".
- `create-brief`'s step 5.1 (the structure check) says to copy the brief, under
  the same file name, into a temporary directory outside `docs/briefs/`, since
  `vloop brief check` compares the file name with the brief's `name`. This is
  the one change to the skill's instructions beyond the renames and pointers.
- `/vloop:operate`'s frontmatter `description` ends
  `Writing briefs is /vloop:create-brief's.` in place of
  `Writing briefs is not this skill's job.`; nothing else in that file
  changes.
- `internal/budget/budget.go`: `Skills` gains `create-brief` and
  `review-brief`, after `verify`; `Ceilings` gains `"create-brief": 1200` and
  `"review-brief": 1000`, each with the comment
  `// 2026-11-02, B28 (#101): its size when it joined the plugin, plus headroom`.
  `budget.Check` then reports a `create-brief` over 1200 words as
  `create-brief: <n> words, over its ceiling of 1,200`.

### K2 — the templates held to the binary's

A Go test fails when a plugin template, with its executive-summary section
removed, differs from the binary's template of the same language:

- `plugin/skills/create-brief/templates/brief-en.md` minus the lines from
  `## Executive summary` up to (not including) the next line starting `## `
  equals `internal/brief/templates/en.md` byte for byte;
- `plugin/skills/create-brief/templates/brief-es.md` minus the lines from
  `## Resumen ejecutivo` up to the next `## ` line equals
  `internal/brief/templates/es.md` byte for byte;
- a template with no executive-summary heading fails too.

The check is a function of the plugin directory and the binary's template
directory, run by the test on the repository's own (`../../plugin` and the
templates, as `TestPluginSkills` reads `SKILLCHECK_PLUGIN_DIR`) and by its
fixtures on copies in temporary directories. Its failure message names both
files, as paths under the repository root:
`plugin/skills/create-brief/templates/brief-<lang>.md differs from internal/brief/templates/<lang>.md outside its executive summary`
(or `... has no executive summary`, for the missing heading).

### K3 — the CLAUDE.md section

`install.Section` keeps every line but the interactive-session paragraph,
which becomes, as one line:

```
In an interactive session you are the operator's hands: follow the operator's playbook (/vloop:operate), not these session rules. To write a brief, follow /vloop:create-brief; to review one, run /vloop:review-brief in a separate session. If /vloop:operate is not available, the vloop plugin is not loaded: say so, give the user the two ways to load it (claude plugin install vloop@vloop after claude plugin marketplace add mvelosop/vloop, or claude --plugin-dir "$(vloop plugin path)"), and meanwhile work from vloop --help.
```

`vloop init` writes it; `vloop upgrade` replaces an older section with it
between the markers, as for any section change.

### K4 — evals and docs

- **`plugin/evals/review-brief-planted-faults`** (`/vloop:review-brief`): a
  scaffolded repository with a draft brief that plants three faults the skill's
  checks exist for — a fact about the code that the code contradicts, a
  literal value the contract and the worked example state differently, and a
  configured check that fails after the first task and stays red until the
  last. Graders: the verdict is not `PASS`; each planted fault is named
  (one grader each); the brief is not edited.
- **`plugin/evals/create-brief-drafts-checked`** (`/vloop:create-brief`): a
  scaffolded repository and a request that settles every fork, states every
  literal value, and says the independent review is out of the case. The
  brief's file name is only known at run time, so the graders read the trace,
  not a file: `tool_used` for `vloop brief new` and for `vloop brief check`
  (each at least once); a `regex` on the trace for the check's
  `0 warning(s)`; an `llm` grader with `focus: trace` that the brief written
  has `status: draft`, an executive summary, and the request's literal
  values verbatim.
- Both follow `plugin/evals/README.md` and `docs/guide/evals.md`: the case
  files, the grader types, at least the work fence's grants, and every Bash
  grant a case names added to the guide's `--allow-tools` list
  (`TestEvalGrantsCoverFence`). `plugin/evals/README.md`'s table of other
  cases gains the two. In `docs/guide/evals.md`: the suite size reads
  `21 cases × 3 runs × 2 arms`; the "vloop's suites" table gains a row each
  for `/vloop:create-brief` and `/vloop:review-brief`, and for the two
  skills it lacks at the base, `/vloop:verify` (`verify-planted-mismatches`)
  and `/vloop:operate` (`operate-proposes-options`); the line
  "`/vloop:operate` has no suite: it is the operator's, interactive." is
  removed.
- `README.md`'s Skills section names all eight skills (and
  `TestReadmeNamesTheSkills` checks the two new names); `docs/guide/concepts.md`'s
  Skills section lists the eight skills, the two new ones described as the
  operator's, like `/vloop:operate`; `docs/domain/execution/session.md`'s
  "Skills, as built" names them as the operator's, not sessions;
  `docs/domain/domain-model.md`'s Brief paragraph and the glossary's
  **design act** say the plugin offers `/vloop:create-brief` and
  `/vloop:review-brief` for it, and that the brief stays vloop's only input.

### Decided here, because the cited documents leave it underdetermined

- The folder names, the ceilings, no experimental label, the template test
  and its messages, the section's wording, `/vloop:operate`'s description:
  above, under the operator's delegation.
- The eval cases' scaffolds, faults and graders beyond what K4 pins are the
  work session's, following the existing cases.
- The create-brief case grades the trace, not a file, since `vloop brief new`
  names the file at run time (the review's finding 4).
- The structure check's copy keeps the brief's file name (step 5.1), fixed in
  the move rather than shipped as it is (the review's finding 6).
- The evals guide's stale lines (the case count, the suites table's missing
  verify and operate rows, the line saying operate has no suite) are fixed
  with the two new rows, since K4 edits the same table.

### Violations the review must rule on

- A skill's instructions rewritten beyond K1's changes (the move is a move).
- A template edited outside its executive summary, or the binary's template
  edited to match it.
- A ceiling set by the skill's size after the run rather than K1's numbers.
- An eval case whose graders could pass with the planted faults missed, or
  with no brief written.
- Any change to the other skills' instructions, the fence, the driver, or
  `vloop brief new`/`check`.

## Worked example

On the built tree:

```
ls plugin/skills
  -> create-brief  gate-review  operate  plan  review  review-brief  verify  work
ls plugin/skills/create-brief plugin/skills/create-brief/templates
  -> SKILL.md  guidelines.md  templates   /   brief-en.md  brief-es.md
grep -rliE 'vloop-(create|review)-brief|experimental|not yet in the plugin' plugin/skills
  -> nothing, exit 1
grep -c '/vloop:review-brief' plugin/skills/create-brief/SKILL.md   -> 1 or more
grep -c '/vloop:operate' plugin/skills/create-brief/SKILL.md        -> 1 or more
grep -c '/vloop:create-brief' plugin/skills/review-brief/SKILL.md   -> 1 or more
sed -n 2p plugin/skills/create-brief/SKILL.md                       -> name: create-brief
sed -n 2p plugin/skills/review-brief/SKILL.md                       -> name: review-brief
```

With a freshly built binary, in a fresh `git init` repository:

```
vloop init; grep -c '/vloop:create-brief' CLAUDE.md                 -> 1
ls "$(vloop plugin path)/skills/create-brief/templates"             -> brief-en.md  brief-es.md
vloop fence check --phase review --root . --tool skill --path review-brief
  -> allow, exit 0
vloop fence check --phase review --root . --tool skill --path vloop-review-brief
  -> deny: skill vloop-review-brief is not vloop's — use /vloop:review, exit 1
```

Planted failures, each on a copy in a temporary directory:

```
a heading of plugin/.../brief-en.md changed       -> the K2 test fails: plugin/skills/create-brief/templates/brief-en.md differs from internal/brief/templates/en.md outside its executive summary
brief-es.md without "## Resumen ejecutivo"        -> the K2 test fails: ... brief-es.md has no executive summary
create-brief SKILL.md padded to 1201 words        -> budget.Check: create-brief: 1,201 words, over its ceiling of 1,200
.claude/skills/review-brief/ in a repository with agent.review = copilot, a stub copilot on PATH (a Go test, as internal/cli/doctor_copilot_plan_test.go)
                                                  -> vloop doctor prints the problem line: ✗ .claude/skills/review-brief has the name of vloop's review-brief skill; a copilot session may load it instead — rename it
```

## Real data

The two skills as they are at the base: the structural checks
(`TestPluginSkills`) already pass on a plugin copy holding them, renamed;
`create-brief` is 1078 words, `review-brief` 904, `guidelines.md` 864, the
templates 1377 and 1480. Each binary template equals its plugin template
minus the executive summary, byte for byte, at the base.

## Out of scope

- Running the brief review as a session on a configured agent,
  `agent.review-brief` (#94, B29).
- Changing what the skills say about writing or reviewing a brief, beyond
  K1's changes.
- `vloop brief new` writing the executive summary; `vloop brief check`
  requiring it.
- A hint in `vloop init`'s output naming `/vloop:create-brief`.
- This repository's project skills and `CLAUDE.md` (the operator's, at
  close); a release.

## Constraints

- Go, at the version in `go.mod`; no new dependencies.
- Must keep building and passing on macOS, Linux and Windows.
- **Gates run on macOS's BSD tools**; no gate runs `claude`, `copilot` or
  touches the network; no gate runs an eval.
- Tests build their fixtures in temporary directories and address them with
  `git -C`; sessions never move refs.
- **Gates follow the diff surface; a gate never runs the suite (P-8).**
- **No task may weaken a check to pass.** An earlier test changes only where it
  pins behaviour the contract changes (here, the section text in
  `internal/install/install_test.go`, the skill list of
  `TestReadmeNamesTheSkills`, and the extracted skill list of
  `TestWorkedExampleB7PluginPath` in `cmd/vloop/b7_e2e_test.go`, which task 1
  changes to the eight names), and only to the new behaviour; the work
  session names the contract item in its notes and the review cites it.
- **New tooling is proven by fixtures, including where it must fail**: every
  planted failure of the worked example.
- Repo-relative paths everywhere. No absolute paths in any file or commit
  message.

## At close (the operator's)

- Remove `.claude/skills/vloop-create-brief/` and
  `.claude/skills/vloop-review-brief/`; point `.claude/skills/vloop-architect`
  and `CLAUDE.md`'s "Which rules bind you" at `/vloop:create-brief` and
  `/vloop:review-brief`.
- **Acceptance**: run the two new eval cases with `claude plugin eval plugin`
  (the guide's command, these two cases, the with-plugin arm), and write B29's
  brief with plugin/skills/create-brief/ and review it with
  plugin/skills/review-brief/, in a separate session. Record the scores and
  what the two skills did in `plugin/evals/README.md` (beside the cases) and
  in B29's header.

## Shape

4 to 6 tasks, each verifiable by a single command. Order is load-bearing:

1. **K1 — the skills in the plugin**: the two folders, the renames and
   pointers, `/vloop:operate`'s description, the budget's list and ceilings.
   Gated on the worked example's first table and the ceiling's planted
   failure.
2. **K2 — the template test.** Gated on its planted failures.
3. **K3 — the CLAUDE.md section.** Gated on `vloop init`'s CLAUDE.md in a
   fresh repository.
4. **K4 — the two eval cases**, the evals README and guide. Gated on the
   structural checks of the new cases.
5. **K4 — the docs**: README, concepts, the domain pages. Prose task: if it
   fails review twice on wording, block it and the operator finishes it.
6. **Close:** every gate at once, and the worked example line for line.

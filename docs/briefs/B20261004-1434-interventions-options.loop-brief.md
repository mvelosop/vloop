---
name: B20261004-1434-interventions-options.loop-brief
description: Make each intervention record the model's options, its recommendation and why, and the operator's decision, derive whether they agreed, and report that agreement by kind and phase — intervention/v2
kind: brief
status: ready
created: 2026-10-04
seeds: The second brief of the v2.0 series and the first run under the gate model; it records the evidence for which operator decisions a driver could take over
depends-on: [B20261003-2049-gate-model.loop-brief]
---
# Brief — vloop B10: interventions that weigh the model against the operator

- **Starting point:** extends `main` after the v2.0.0-beta.1 release
  (`42ee0a6`, tag `v2.0.0-beta.1`) and the upgrade to it (`dc6c4a4`, PR #16).
  The planner pins the base it plans from as the **base** every gate compares
  against — never `HEAD~n`.
- **Produced by:** the design act of 2026-10-04, from
  `docs/design-notes/vloop-v2.0-roadmap.md` → "B10", with the forks settled with
  the operator (intervention recorded with this brief).
- **Run by:** `vloop run` from the released v2.0.0-beta.1 — the first brief
  planned under the gate model: gate fixtures, the gate review, and this
  repository's `[[check]]`.

## What it is

Today an intervention records what happened and what would automate it, but not
what the model proposed or whether the operator took it; those live, when at
all, as prose inside a section. After this run every intervention can carry **up
to three options** the model proposed at the time, **the one it recommended and
why**, and **what the operator decided**, and vloop derives an **`agreement`**
from them. `vloop metrics` reports agreement by kind and by phase, per brief,
across briefs and across a workspace — the evidence for which interventions a
driver could take over. `vloop intervention show` prints a record with the ids in
its context resolved, and the existing records move to `intervention/v2`
through an explicit command.

The blind spot this closes: 66 records already hold a context, a suggestion and a
decision in body sections the parser does not know, so it folds them into `done`
(D20261004-1434); nothing counts interventions at all, so the data the B1–B4 note
(`docs/design-notes/vloop-interventions-b1-b4.md`) analysed by hand has never
been read by vloop.

## Why this shape, and what was rejected

Decided with the operator on 2026-10-04. **Do not re-litigate these** — the
planner inherits them.

- **Numbers in the frontmatter, prose in the body.** The option count, the
  recommended option, the decided option, `adjusted` and `agreement` are
  frontmatter fields, short and validated; the options' text, the
  recommendation's reason, the decision's sentence and the context are body
  sections. *Rejected:* everything in frontmatter (this repository's frontmatter
  parser reads one `key: value` per line and cannot hold lists of sentences); a
  JSON sidecar per record (two files, and the Markdown stops being the record).
- **`agreement` is derived, never judged from text.** Whoever records says which
  option was chosen, whether it was adjusted, or that something else was;
  vloop derives the agreement. The operator corrects it by setting the choice,
  never the agreement. *Rejected:* setting `agreement` directly (nothing ties it
  to the options); a model inferring it from the decision's sentence.
- **Up to three options, never padded.** One or two real options are recorded
  as one or two; a straw option to reach three is worse than none. *Rejected:*
  exactly three always.
- **The existing records become `no-options`, through `vloop intervention
  migrate`.** They were never three-option records and options are never
  reconstructed after the fact; their Context, Suggested and Decided sections
  are kept and now parsed. vloop 2 still reads a v1 record (as `no-options`), and
  `vloop upgrade` rewrites no record — the migration is the operator's explicit
  act. *Rejected:* leaving them v1 (two formats, and their sections stay folded
  into `done`); one-option records judged by hand now (reconstruction).
- **Agreement by kind first.** The cut that answers "what could a driver take
  over" is kind × agreement beside `automatable`; phase is the second cut.
  *Rejected:* phase first; export only, with analysis outside vloop.
- **`show` resolves ids written in the context.** No new reference fields.
  *Rejected:* structured `task`, `iteration`, `commits`, `defects` fields (more to
  fill in correctly); printing the record unresolved.
- **Additive keys, not new versions, for metrics and the export.** `metrics/v2`
  gains an optional `interventions` object and the export's intervention record
  gains optional `agreement` and `options`; both schemas allow unknown keys
  (C-3). The export carries the counts, never the options' text (M-7).

## Binding references

- `docs/domain/domain-model.md` — the intervention's vocabulary row, id and "On disk" line, M-1 and M-7, C-3: metrics recomputed from raw records, what an export may carry, unknown keys allowed
- `docs/domain/measurement/measurement-context.md` — interventions as measurement's next entity, which this brief makes them
- `schemas/intervention.v1.json` — the record this brief versions to `intervention/v2`
- `schemas/metrics.v2.json` — the metrics document that gains the interventions report
- `schemas/export.v1.json` — the export's intervention record, which gains the agreement and option count
- `docs/guide/defects.md` — "Interventions": the guide that documents the record, its fields and its commands
- `docs/guide/metrics.md` — the metrics guide, which gains the interventions report
- `plugin/skills/operate/SKILL.md` — the shipped operator playbook, which gains the options rule and how to record it
- `.claude/skills/vloop-operator/SKILL.md` — this repository's operator playbook, step 6 (recording interventions), which follows it
- `docs/design-notes/vloop-interventions-b1-b4.md` — what the first 38 records showed, and why a driver would want these fields

## Behaviour contract

Messages below are exact where quoted; `<…>` marks a value. Every refusal is one
line on stderr starting with `vloop: `.

### I1 — the record, `intervention/v2`

**Required.** On disk, a v2 record's frontmatter is v1's plus, in this order
after `by`:

```
schema: intervention/v2
options: <0-3>
recommended: <0-3>          # 0 when options is 0
decided: <1|2|3|other|"">   # "" when nothing was decided between options
adjusted: true              # only when the decided option was adjusted
agreement: <recommended|other-option|adjusted|different|no-options>
```

The body is the summary line, then these sections in this order, each only when
it has content: `**Trigger.**`, `**Done.**`, `**Context.**`, `**Options.**` (a
numbered list `1. …` to `3. …`, one line each), `**Recommended.**` (the reason,
one or two sentences), `**Suggested.**` (only in migrated records),
`**Decided.**` (one sentence), `**What would automate it.**`.

schemas/intervention.v2.json is v1's JSON form plus `options` (an array of up
to three strings), `recommended` (`{option, why}`), `decided` (`{option, adjusted,
text}`, `option` 0 for other or none), `context`, `suggested`, `agreement`.

**`agreement` is derived:**

| options | decided | adjusted | agreement |
| --- | --- | --- | --- |
| 0 | any | — | `no-options` |
| ≥ 1 | the recommended option | no | `recommended` |
| ≥ 1 | another option | no | `other-option` |
| ≥ 1 | any option | yes | `adjusted` |
| ≥ 1 | `other` | — | `different` |
| ≥ 1 | `""` | — | invalid: options need a decision |

A record whose stored `agreement`, `options` count or numbered list disagree is
**invalid**: `vloop intervention list` fails naming the file, as it does today
for a bad field.

**Reading.** A record without `schema: intervention/v2` is v1: it reads as
`no-options`, and its `**Context.**`, `**Suggested.**` and `**Decided.**`
sections parse into `context`, `suggested` and `decided.text` — no longer into
`done`.

### I2 — recording: `vloop intervention add`

**Required.** New flags beside today's:

- `--context '<paragraph>'`;
- `--option '<sentence>'`, repeatable, at most three;
- `--recommended <n>` and `--why '<reason>'`, both required with any option;
- `--decided-option <n>` or `--decided-other`, one required with any option;
  `--adjusted`, only with `--decided-option`; `--decided '<sentence>'`.

Refusals, exit 2, nothing written:

- `vloop: at most three options`
- `vloop: options need --recommended and --why`
- `vloop: --recommended must name an option, 1 to <n>`
- `vloop: options need --decided-option or --decided-other`
- `vloop: --decided-option must name an option, 1 to <n>`
- `vloop: --decided-option and --decided-other exclude each other`
- `vloop: --adjusted needs --decided-option`
- `vloop: --recommended, --decided-option and --decided-other need options`

Without options, the record is `no-options`, as every record is today; the
flags of v1 work unchanged.

### I3 — correcting: `vloop intervention set`

**Required.** `set` also accepts `recommended`, `decided` (`1`–`3`, `other` or
`""`) and `adjusted` (`true` or `false`), validated against the record's option
count, and re-derives `agreement` in the same write. `agreement` and `options`
are not settable: `vloop: agreement is derived — set decided, adjusted or
recommended instead` / `vloop: options are recorded with the intervention, not
set` (exit 2).

### I4 — migrating: `vloop intervention migrate`

**Required.** Rewrites every v1 record under `.vloop/interventions/` to v2:
inserts `schema: intervention/v2`, `options: 0`, `recommended: 0`,
`decided: ""` and `agreement: no-options` after `by`, and changes **nothing
else** in the file. Prints `migrated <n> record(s)`, or `nothing to migrate` when
none is v1; idempotent. `--dry-run` lists the files and writes nothing.
`vloop upgrade` does not call it.

### I5 — `vloop intervention show <id>`

**Required.** Prints the record: the id and summary, the frontmatter fields, then
each section under its name. Then `links`, one line per id found in the Context
section, in order of first appearance:

- a defect id `D<…>` → its summary;
- an intervention id `I<…>` → its summary;
- `T<n>` → the task's title from the plan committed for the record's brief
  (the latest `[vloop] plan <run id>` commit's `.vloop/state/state.json`);
- a commit sha (7 to 40 hex characters that git resolves to a commit) → its
  subject;
- a run id `B<YYYYMMDD-HHMM>-<slug>` → its run folders under `.vloop/state/runs/`.

An id that resolves to nothing prints `<ref>  (not found)`. `--json` prints the
JSON form plus `links: [{ref, kind, title}]`. An unknown id: `vloop: no
intervention <id>` (exit 1, `{"error": …}` under `--json`, as `task show`).

### I6 — metrics

**Required.**

- **Per brief**, `vloop metrics <brief>` gains one line after `defects`:
  `interventions  <n> · recommended <a> · other-option <b> · adjusted <c> ·
  different <d> · no-options <e>`, and its JSON (and the close snapshot) an
  `interventions` object: `{total, by_agreement: {…five keys…}}`. Series-level
  records (`brief: ""`) belong to no brief.
- **Across briefs**, `vloop metrics --interventions` prints one row per kind, in
  the kind order of `intervention/v1`, then a `total` row, over every record in
  the repository including series-level ones: `kind  n  recommended  share
  other-option  adjusted  different  no-options  automatable-yes
  automatable-partly`. `share` is recommended ÷ (n − no-options), as a whole
  percentage, `n/a` when n − no-options is 0. `--by phase` gives the same table
  by phase, in phase order. `--json` prints the rows.
- `--workspace <file> --interventions` prints the same table over every
  repository the file lists, with a `repo` column first, then a `total` row.
- `--by task` with `--interventions`: `vloop: --by takes kind or phase with
  --interventions` (exit 2).

### I7 — the export

**Required.** Each intervention record gains `agreement` and `options` (the
count). The options' text, the reason, the decision and the context are never
exported (M-7).

### I8 — the skills

**Required.**

- `plugin/skills/operate/SKILL.md`: whenever it brings the operator a decision,
  it proposes **up to three real options** — never a straw option to reach
  three — **recommends one and says why**, and records the intervention after
  the operator decides, with `--option`, `--recommended`, `--why`, the decided
  choice and `--context` (the run, iteration and task; the commit or files; the
  triggering output quoted briefly; what changed because of it). Options are
  only what it proposed before the operator decided.
- `.claude/skills/vloop-operator/SKILL.md` step 6 follows: records with
  `vloop intervention add` and these flags; `vloop intervention migrate` is named
  where the index is refreshed.
- An eval case for `/vloop:operate`, `operate-proposes-options`: a run halted on
  a gate dispute; graded on proposing one to three options with a recommendation
  and its reason, on not amending the gate before the operator answers, and on
  the absence of a straw option.

### I9 — the index, the guides and the domain

**Required.** `tools/interventions-index.sh` adds an `Agreement` column (from the
frontmatter; `no-options` for a v1 record) and validates `by` and `agreement`
too. `docs/guide/defects.md` documents every v2 field, flag and refusal;
`docs/guide/metrics.md` the interventions line and table;
`.vloop/interventions/README-interventions.md` the v2 frontmatter and sections.

`docs/domain/domain-model.md` changes, and nothing else in it: the
intervention's vocabulary row reads "anything the operator did around a run
besides testing, with the options the model proposed and the operator's
decision" (dropping "not yet a vloop entity"); the gaps line saying interventions
have no command or schema goes. `docs/domain/measurement/measurement-context.md`
describes interventions as the entity they now are.

### Decided here, because the cited documents leave it underdetermined

- The decided choice is a number or `other`, not the option's text: the text
  is in the Options section once.
- `share` leaves out `no-options`: a record with no options cannot agree or
  disagree.
- `show` finds ids by pattern in the Context section only, not in Trigger or
  Done, so a link is something the recorder chose to name.
- The migration leaves `recorded` and `backfilled` as they are: it changes the
  format, not when the record was made.

### Violations the review must rule on

- An `agreement` computed from the decision's text, or settable directly.
- A migration that changes any byte of a record's body.
- The options' text, the reason, the decision or the context in the export.
- Options accepted beyond three, or a skill instruction that pads to three.
- `vloop upgrade` rewriting records.
- A message that differs from the one quoted here.
- Scope from B11: help-text rewording, exit-code remapping, the README.

## Worked example

In a temp repository with `vloop init` done:

```
vloop intervention add "T3's gate failed on a path typo" --brief B1 --phase halt --kind repair \
  --automatable partly --by both --trigger 'vloop run exited 2' --done 'gate replaced' \
  --context 'T3 in run B20260101-0900-a; see D20260101-0900-x and commit <sha of HEAD>' \
  --option 'replace the gate with vloop task verify' --option 'reset T3 and retry' \
  --option 'abandon the brief' --recommended 1 --why 'the gate, not the work, is wrong' \
  --decided-option 1 --decided 'replaced as recommended'
  -> .vloop/interventions/I<stamp>-t3-s-gate-failed-on-a-path-typo.md      exit 0
     frontmatter: options: 3, recommended: 1, decided: 1, agreement: recommended

same, --decided-option 2                       -> agreement: other-option
same, --decided-option 1 --adjusted            -> agreement: adjusted
same, --decided-other --decided 'fixed by hand'-> agreement: different
same, no --option flags                        -> options: 0, agreement: no-options
--option ×4                                    -> vloop: at most three options                              exit 2
--option ×2 --recommended 3 --why x --decided-option 1
                                               -> vloop: --recommended must name an option, 1 to 2          exit 2
--option a --recommended 1 --why x             -> vloop: options need --decided-option or --decided-other  exit 2
--option a --recommended 1 --why x --decided-option 1 --decided-other
                                               -> vloop: --decided-option and --decided-other exclude each other   exit 2
--decided-other, no options                    -> vloop: --recommended, --decided-option and --decided-other need options   exit 2
vloop intervention set <first id> decided 2    -> agreement: other-option in the file
vloop intervention set <first id> agreement recommended
  -> vloop: agreement is derived — set decided, adjusted or recommended instead    exit 2
hand-edit the first record: agreement: different
  vloop intervention list -> fails naming the file                                  exit 1
a v1 record (no schema line) with **Context.** and **Decided.** sections
  vloop intervention list --json -> its context and decided.text hold them; done does not
  vloop intervention migrate     -> migrated 1 record(s); body bytes unchanged
  vloop intervention migrate     -> nothing to migrate
vloop intervention show <first id>
  -> links: D20260101-0900-x  (not found) · <sha>  <its subject> · B20260101-0900-a  (not found)
vloop metrics --interventions
  -> repair row: n 4 (or as recorded) with recommended, share, the other counts; total row
vloop metrics --interventions --by task
  -> vloop: --by takes kind or phase with --interventions                            exit 2
vloop metrics export
  -> each intervention line carries agreement and options; none carries an option's text
```

The gate threads ids and paths it computes (the record's file name, HEAD's sha)
rather than hardcoding them.

**Real-data check.** On this repository: `vloop intervention migrate --dry-run`
lists every record at the base (all v1); `vloop intervention migrate` then changes only their
frontmatter (`git diff` shows only added frontmatter lines), and
`vloop intervention list --json` gives every one of the 66 records with a
Context section a non-empty `context` and a `done` free of `**Context.**`.
`vloop metrics --interventions` over the migrated records equals the counts of
kind, phase and `automatable` in the index, with every record `no-options`.
`vloop metrics --json` for B1–B9 is unchanged apart from the added
`interventions` object.

## Out of scope

- B11 (the quality pass, the README, help text, exit-code remapping), and the
  `CLAUDE.md` section's session list missing the gate review (B11's).
- A driver acting on the agreement data (the horizon's H1); recommendations
  produced by vloop itself.
- Options for defects; a `vloop defect show`.
- Dashboards or charts of agreement; anything beyond the tables above.
- Reconstructing options for past records, or judging their agreement now.
- The deterministic grader for `plan-checks-its-gates` (D20261004-1309, B11's).

## Constraints

- Go as pinned in `go.mod`; no new dependencies. Every gate builds for
  `windows` and `linux` as well as the host.
- Gates run on macOS's BSD tools: no empty alternatives in `grep -E`, no
  `sed -i` without a suffix, no `\+` or `\|` in basic regexes, no `date -d`.
  A scratch directory is `mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX"`.
- Sessions never move refs. Tests build their repositories in temporary
  directories and address them with `git -C`.
- **The gate model applies** (`plugin/skills/plan/SKILL.md`): each gate judges
  this brief's contract by running the built binary, with its judge in the
  verify or its gate folder; no gate runs the repository's tests — this
  repository's `[[check]]` does, after every iteration.
- **No task may weaken a check to pass.** These earlier tests **must** change,
  and only as stated:
  - `internal/intervention/intervention_test.go`: `TestAddWritesRecord` pins the
    v2 bytes; `TestExistingRecordsParse` reads v1 and v2 records;
  - `internal/cli/intervention_test.go`: `TestInterventionAddListSet`'s list
    format gains the agreement column; `TestExistingInterventionsValidate`
    validates with `intervention/v2` once the records are migrated;
  - `internal/cli/export_test.go` `TestExportInterventions` gains `agreement`
    and `options` among the keys;
  - `internal/cli/guide_test.go` `TestGuideDefectsCoversInterventions` covers
    the v2 fields and values;
  - the schema lists in `internal/cli/schema_test.go`,
    `cmd/vloop/b2_e2e_test.go` (`TestWorkedExampleB2Commands`) and
    `internal/schema/schema_test.go` gain `intervention/v2`;
  - metrics and snapshot tests that pin the per-brief summary gain the
    `interventions` line;
  - `TestEvalGrantsCoverFence` and the eval index cover the new operate case.
- Evals are the skills' gates: the run record states the `/vloop:operate`
  case's score and spend, run by the operator.
- Repo-relative paths everywhere. No absolute paths in any file or commit
  message.

## Shape

8 to 11 tasks, each independently verifiable by a single command. Order is
load-bearing:

1. **`intervention/v2`: the schema, parsing v1 and v2** (the folded sections
   fixed), writing v2, the derivation and its validation.
2. **`add` and `set`** with the new flags and refusals (I2, I3).
3. **`migrate`** (I4).
4. **`show`** with links (I5).
5. **Metrics** per brief, across briefs, by phase, and across a workspace (I6);
   **the export** (I7).
6. **The skills and the operate eval case** (I8).
7. **The index script, the guides and the domain** (I9).
8. **This repository's records migrated** (the real-data check), and
   **close:** the worked example line for line.

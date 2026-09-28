# The loop — a user manual

End to end: brief in, working code out, with evidence of how it got there.

For the design *rationale*, read
[brief 0002](../docs/briefs/0002-next-generation-autonomous-loop.md). This is the
operating manual.

---

## 1. The mental model

Three kinds of session, each a **fresh `claude -p` with no memory of any
other**. Files are the entire continuity.

```
brief ──▶ PLAN ──▶ state.json ──▶ ┌─ pick next ready task
 (opus, once)                     │  WORK   does one task, proposes an outcome
                                  │  GATE   driver re-runs EVERY done task's verify
                                  │  REVIEW separate, read-only, independent verdict
                                  │  driver applies it, journals, commits
                                  └─ repeat
```

The one thing to internalise: **the driver decides everything mechanical.**
Which task is next, whether a task is really done, how many attempts it has
burned, when to stop. Agents do work and give opinions. They never set status
and never commit. If you are wondering "could a session have faked this?", the
answer is almost always no, because it has no way to express the claim.

## Sessions, not subagents

Claude Code has a subagent mechanism — `.claude/agents/`, spawned with the Task
tool from inside a running session. **This loop deliberately does not use it.**
There is no `.claude/agents/` directory here. Each phase is a separate
`claude -p` process, started by a bash driver, reading a skill.

```
subagents                          this loop
─────────                          ─────────
one session                        one OS process per phase
  ├─ Task(implementer)             driver ──▶ claude -p /loop-work   (exits)
  ├─ Task(reviewer)                driver ──▶ gate                   (no model)
  └─ decides what to do next       driver ──▶ claude -p /loop-review (exits)
context accumulates                driver decides what to do next
an LLM orchestrates                a shell script orchestrates
```

### Why

**No shared memory, by construction.** A subagent's parent accumulates every
subagent's output and frames the next prompt. Here the work session and the
review session are separate processes that cannot see each other — the reviewer
physically cannot inherit the implementer's rationalisation. Isolation is a
property of the operating system, not of prompt discipline.

**The orchestrator is deterministic.** Task selection, gates, attempt counting,
stop conditions and halting are bash. That is why there are 46 checks that run free and offline with a stubbed `claude` on `PATH` — including ones for the
attempt ceiling, the convergence halt and the stale-handoff guard. **You cannot
stub the Task tool.** Every mechanical bug found in this loop was found by those
tests, not by a run.

**Per-phase telemetry falls out for free.** `claude -p --output-format json`
returns cost, turns, duration and permission denials per session, which is the
entire basis for the run-level signals. Subagent accounting rolls up into the
parent.

**Structural honesty.** A session has no way to set task status or commit — the
driver owns both. A subagent returns text to a parent that then decides, so
"the implementer marked it done" becomes a thing you must guard against.
`exploring-claude` needs a fabrication check and a verdict-guard script (in
that repo, not this one) for exactly that; here it is unrepresentable.

### What it costs

**Money and latency — but less than you would think.** Measured across run 1's
22 sessions:

| | tokens |
| --- | --- |
| read from cache | 11,335,654 |
| written to cache | 766,121 |
| **fresh (uncached) input** | **32,292** |

**93% of input is served from cache**, and uncached input averages ~20 tokens
*per session*. A genuinely cold start would pay thousands for the system prompt
and tool definitions alone, so the shared prefix is plainly not being re-bought
each time — prompt caching is content-addressed server-side, and separate
processes with the same prefix hit it.

What a fresh session does not inherit is the **conversation**. It re-reads the
files it needs and re-derives its understanding, which shows up as the ~35k of
cache *creation* per session plus its output tokens. So the real cost of the
fresh-session design is re-deriving context, not re-uploading it.

Measured all-in: ~$1 per iteration, ~$0.50 of it the work session. Subagents
would still be cheaper per step by sharing the parent's *conversation*, but the
gap is much narrower than the architecture suggests.

**It runs outside Claude.** You start it from a terminal, not from a
conversation. There is no mid-run steering, and it cannot ship as a single
plugin (see *Using the loop in another repo*).

**No cross-phase judgement.** Nothing weighs "the reviewer keeps raising the
same thing" — because nothing is holding both. That is deliberate; the run-level
signals exist to replace it, and `exploring-claude`'s own design notes record
that per-tick judgement is exactly what fails to notice a globally stuck run.

### When the other choice is right

If the work is short enough that context accumulation is a feature rather than a
liability, subagents are simpler and cheaper — one session, no driver, no
install. The fresh-session bet only pays off when a run is long enough that you
would not *want* iteration 11 to remember iteration 1.

## 2. Installing it

```bash
.loop/install.sh /path/to/target-repo
```

The loop is **vendored**, not linked — there is no clean submodule or plugin
route, because it is two things with different homes: the skills must sit at
`.claude/skills/` for Claude Code to resolve `/loop-work T3`, and the driver is
a shell script you run from your terminal.

The installer handles three classes of file differently:

| | |
| --- | --- |
| **loop-owned** — `.loop/` except `state/` and `tmp/`, `.claude/skills/loop-*` | replaced wholesale; these *are* the loop |
| **yours** — `.loop/state/`, `.loop/tmp/` | never written by the installer |
| **shared** — `CLAUDE.md`, `.gitignore` | **merged**, never clobbered |

That first boundary is the one that makes an upgrade safe to run on a repo
holding months of journals and run telemetry, so it is a **checked** rule
rather than a documented one: the installer's manifest is inverted — it names
what to leave alone, never what to copy, so a new mechanism file ships by
default and consumer state cannot be reached by accident — and a scenario
installs over a target carrying sentinel state and fails if a byte of it moves.

**Your `.claude/settings.json` is not touched at all.** The loop's permission
fence is `.loop/settings.json`, which `run.sh` hands to every session with
`--settings`, so it binds loop sessions and nothing else. Merging a deny list
into a repo's project settings would bind the operator's own interactive work
too — deny beats allow, repo-wide — which on a repo whose agents commit and
push is a bad first day.

It stamps `.loop/.installed` with the release it came from and the commit that
built it — propagated intact through a chained install, not re-derived from
whatever repo the copy was taken from — and finishes by **running
the loop's own suite in the target** — 46 checks, free and offline, no model.
(You will see one fewer until you have briefs of your own: the brief checker
sits out when `docs/briefs/` is empty.)
That is the install test: a copied artefact that can prove it works where it
landed. (`--no-proof` skips it for a re-run that only refreshes the mechanism.)

Re-run it to update; it is idempotent.

**Two things are yours to set**, and they are the whole stack-specific surface:

1. `.claude/settings.json` — add the commands your gates need
   (`Bash(pnpm:*)`, `Bash(go:*)`, `Bash(cargo:*)`…). The loop never names a
   test runner: each task carries its own verify command, so **the gate list
   belongs to your plan, not to the loop**. Because the loop's fence lives in
   its own file, nothing you add here can be overwritten by an upgrade.
2. `CLAUDE.md` — the loop's rules land between `loop:begin`/`loop:end` markers.
   Add a toolchain note of your own.

You also get `.loop/manual.md` (this document) and `.loop/examples/` — the two
worked briefs the "Writing a brief" section cites, which used to live only in
the loop's own repo where a consumer following the pointer found nothing.

## 3. One-time setup

**In this repo**, everything is already in place. **In any other repo**, install
it first — see the previous section — then come back here.

```bash
jq --version && git --version && claude --version   # required
```

`uv` is only needed if *your* gates use it. The loop never names a test runner.

**Check the setup before you spend anything:**

```bash
.loop/run.sh --check          # --preflight is the same flag
```

It runs the same preflight a real run runs — the same code, not a
reimplementation, because a sanity check that can disagree with the thing it
checks is worse than none. Tools, workspace trust, the permission fence, git
identity, and whether the knowledge roots you declared can actually be
surveyed: each has an index or describes itself, and each index names every
document beside it.

It needs no brief, takes no lock, and leaves no run directory behind, so it is
the right thing to run after editing `.claude/loop-knowledge.md` or adding a
document. It exits **1 on advisory findings** — a real run would still start on
those, but the two commands are asked different questions, and answering "is
anything wrong?" with a silent 0 makes the flag useless in the hook or CI job it
exists to be put in.

Two more checks, also free and also offline:

```bash
.loop/check-brief.sh docs/briefs/000N-....md   # is this brief plannable
.loop/tests/run-all.sh                         # the loop's own suite
```

**Trust the workspace.** Run `claude` interactively in the repo once and accept
the trust dialog. Without it `claude -p` *silently ignores* `.claude/settings.json`,
so a run executes under the wrong permission surface. Preflight refuses to start
until this is done — it is the single most common way to waste a run.

### Telling the planner where your knowledge lives

**Optional, and worth it in any repo that has more than the brief to obey.**

The brief carries the decisions. It does not carry the repo — the conventions,
the invariants, the specs a task must respect because something else already
depends on them. In a greenfield target there is nothing there. In a codebase
with years of accumulated knowledge that *is* most of what a work session needs,
and there is no channel for it: `CLAUDE.md` is global, and the brief is
re-read every iteration by two sessions, so neither can carry something that
binds one task.

So the planning session attaches references **per task**, and it finds them by
reading `.claude/loop-knowledge.md` — your file, at a path the loop fixes:

```markdown
| Surface | Path | How to find what is in it |
| --- | --- | --- |
| Decision records | `docs/decisions/` | Index at `README.md` |
| Domain model | `docs/domain/` | Every file has a `description:` frontmatter line |
```

**Where a description comes from, cheapest first.** The planner takes the first
source that answers "what is in this root":

1. **The root's index** — `index.md`, `README.md` or `README-<subject>.md`. One
   file read per root, written to be read that way.
2. **`description:` frontmatter** where there is no index — one line per file
   instead of one per root, at any depth, so a root holding one folder per slice
   is not blind because its top level is empty.

The three index names are equivalent to the loop and it prescribes none: the
name affects *finding* the file, never parsing it — all three are markdown.
**`README-<subject>.md` is the recommendation.** A dozen files called
`README.md` are indistinguishable in a flat search, which is what an Obsidian
quick-switcher and backlink pane give you; a subject-named one is unique
everywhere. GitHub stops auto-rendering it at the top of the folder view, which
costs almost nothing — the file is right there in the listing, named for what
it is, one click away.

That the same `description:` line serves the planner and Obsidian's properties
is not a coincidence. Both are the same request: let a machine tell what this
document is without reading it.

**Why not a JSON or YAML index?** A structured index is easier to
validate and worse at everything else. It is a second source of truth that
drifts from the documents it describes; no human reads it, so nobody notices
when it rots; and it is invisible to the graph, backlinks and search that make
a markdown index worth keeping. The real hazard with an index is not its format
but its **coverage** — a document added and never listed is invisible to the
planner forever, and nothing about the repo looks wrong. Coverage is checkable
in any format, so the loop checks that and leaves the format alone. (This
manual's own repo failed that check the first time it ran: a document had sat
unlisted in an indexed root for three merges.)

Each task then carries what actually binds it, with a reason:

```json
"references": [
  {"path": "docs/domain/state-transaction.puml",
   "why": "the status lifecycle this task must not widen"}
]
```

**The work session reads them, and so does the review session.** That second
half is the point. A constraint the reviewer cannot see is a constraint it
cannot enforce — the work would be bound by a document the review has no idea
it was bound by. Because the reference lives on the task, both see the same
list, and a convention that was pointed at and ignored becomes a finding rather
than a thing nobody noticed.

The loop knows nothing about what these documents *are*. It has no idea what an
ADR is, or a use case, or a tier. Those are conventions and they belong to the
repo that has them — so the declaration is prose you write, not a schema the
loop imposes. If tier variants change which guideline binds, say that in your
file and the planner will read it.

**The loop checks a property; your repo enforces a convention.** The property is
"can a session find out what is in here" — an index or descriptions, either
will do. A stricter house rule (*every `.md` carries frontmatter, every folder
carries an entry file, every index names every document in its folder*) is
worth having and belongs in your own lint, where it can fail a commit. It does
not belong here, because the loop also installs into repos that will never adopt
it, and a check that half its users must switch off is a check nobody trusts.

Two things are checked rather than advised:

- **A folder is a valid reference** when what binds is the whole bundle — a
  design handoff, a spec with its diagrams. It needs an entry point for the same
  reason a root does, and the driver warns when one has none: the plan is sound,
  the docs are not, and that is not the planner's to fix. Content you cannot
  annotate — an export that would fork if edited — is handled by writing the
  index *beside* the bundle rather than inside its files.
- **A document its own index does not name is reported by preflight.** An index
  is the cheapest thing to read and the easiest to let rot, and this is the
  failure that leaves no trace: the file is committed, the index looks fine, and
  the planner simply never learns the document exists.
- **A reference that does not resolve fails the plan**, before the run spends
  anything. A cited-but-missing file stops a work session that has no way to
  recover, and costs an attempt to discover.
- **A declared root with neither an index nor `description:` frontmatter is
  reported by preflight.** That failure is otherwise silent: planning succeeds,
  tasks simply carry no references, and nothing downstream can distinguish
  "nothing bound this task" from "the planner could not see what did".

Declare nothing and the whole feature costs nothing — no file, no warnings,
empty `references`. That is the right behaviour for a repo whose brief really is
the whole world.

## 4. Writing a brief

The brief is the highest-leverage artefact in the system. Everything downstream
is measured against gates the planner writes *from it*.

**Pin decisions, leave mechanics open.** Name the behaviour, the exit codes, the
output format, the worked example. Do not name the module layout — that is the
implementation's to choose, and pinning it buys nothing.

**Include a worked example with exact expected values.** It becomes the
end-to-end acceptance test, and it is the arbiter when two implementations
disagree.

**Write an out-of-scope list.** It is not decoration: it is how scope creep
becomes a measurable finding rather than a matter of taste.

**Say roughly how many tasks you expect.** It calibrates decomposition.

Start from the template, which is a filled-in skeleton of the shape below:

```bash
cp .loop/brief-template.md docs/briefs/0001-my-thing.md
```

Check it before you spend anything:

```bash
.loop/check-brief.sh docs/briefs/0001-your-brief.md
```

It verifies the structure a brief needs: a worked example with concrete values,
a non-empty out-of-scope list, constraints, an expected task count, pinned exit
or status codes, no absolute paths, that every path it references resolves, and
that it names no issue key or tracker URL its readers cannot open.
Briefs are only checked if they say `**Status:** ready to plan` — a discussion
document is not a worse brief, it is a different kind of document.

**It cannot check the thing that matters most**, which is whether the brief pins
decisions and leaves mechanics open. A brief can pass every check and still be
bad. That judgement is what the list above is for.

See [brief 0003](../docs/briefs/0003-runstat-cli.md) (greenfield) and
[brief 0004](../docs/briefs/0004-runstat-review.md) (incremental) as worked examples.

### Deriving a brief from an existing design process

Everything above assumes a blank page. In a repo that already has a design
practice — ADRs, design notes, use-case entries, flow docs, a tracker — you are
not writing a brief so much as **translating** one, and that is a different act
with its own failure modes. This is the shape of the consumer-side skill that
does it: the loop ships the contract (`.loop/brief-template.md`) and the
checker (`.loop/check-brief.sh`); how *your* design surfaces map onto them is
yours, because the moment the loop names your tracker it stops being
stack-independent.

Four rules, and the first is the one that actually bites.

**Know which row you are in.** The question is never inline-vs-reference in
general; it is what a given thing is *for*.

| What | Rule | Why |
| --- | --- | --- |
| The behaviour contract | **In the brief, or delegated whole to one named document.** Never split | Where a spec already says what to build — a use-case entry, an ADR — naming it as the sole authority is *better* than copying it, because a copy goes stale silently. What must not happen is half the contract here and half there: then the two drift and nothing sits above them to break a tie. Delegating is fine; splitting is not. |
| Worked example, out-of-scope, constraints | **In the brief.** Always | These are about *this run*, not about the subject. A use-case spec has flows, not the exact literal values a gate asserts on; it says what the use case is, not which adjacent thing this run must not build. `check-brief.sh` fails a brief missing any of them, which is the same rule stated as a check. |
| Tech guidelines, ADRs, use-case entries, design handoffs cited *as context* | **Reference, repo-relative.** Do not copy | Durable, and they outlive the brief; a copy goes stale and nobody notices. |
| Tracker issues, chat threads, the design conversation itself | **Inline.** There is no choice | Not openable: fresh session, `--strict-mcp-config`, `WebFetch`/`WebSearch` denied. |

So a brief that says *"the behaviour is exactly what `docs/use-cases/uc-<n>.md`
specifies, and nothing beyond it"* is a good brief — one authority, no drift,
and the planner reads that document when it authors the gates. A brief that
restates two of the flows and leaves the third to the spec is a bad one, and it
will not look bad until the spec changes.

Referencing is better than it sounds, because **all three session types read
the brief** — planner, work and review all follow `state.json`'s `brief` field.
A referenced guideline reaches the implementer and the reviewer, not just the
planner, so you do not need to duplicate it into acceptance criteria to make it
bind.

The cost is real, though: the brief is re-read *every iteration, by two
sessions*. A bibliography of eight documents is eight documents of attention,
twice per iteration, in sessions that each have exactly one job. So cite few,
and label each with why it is there — `docs/…/http-errors.md` *for the error
shape*. An unlabelled reference is either followed wastefully or skipped
silently, and neither was the intent. Where a document binds only one task, put
it in that task's `notes` rather than in a header every session re-reads.

**Strip, do not inherit.** An ADR pins structure on purpose; a flow doc pins
sequence. The brief carries their *decisions* and drops their *mechanics*. "Pin
decisions, leave mechanics open" assumes a blank page — it does not state the
direction of travel when you are starting from documents that legitimately do
the opposite.

**Declare the cut.** A design note may still be exploratory; a brief is a
contract. Somebody has to say which parts are frozen and which are open, and
that is the design act's job, not the planner's. If nobody declares it, the
planner will decide by accident.

**Out-of-scope has a source.** Where a design was decomposed into ordered
slices, the sibling slices *are* the out-of-scope list, already ranked. That is
usually the cheapest section of the whole brief to write, and it is the one
`check-brief.sh` fails on outright.

## 5. Planning, and checking the plan before you spend

```bash
.loop/run.sh --plan-only docs/briefs/000N-....md
```

Runs the plan phase alone, commits it, and stops (exit 0). ~$2–4. **Do this on
anything unfamiliar.** The plan authors every `verify` command, and a weak one
silently lowers the bar for the whole run — so one session spent reading it back
is the cheapest insurance the loop offers.

(`LOOP_MAX_ITERATIONS=0` does the same thing and still works, but it stops at
the budget rather than at the plan, so it exits 4 rather than 0 and reads like
the workaround it was. Prefer the flag.)

Then read `.loop/state/plan.md` and the top of the plan's journal. The planner's report
ends with **what it interpreted rather than read** — that list is your one cheap
chance to catch a misreading before every iteration inherits it.

Two more things in `plan.md` are worth a deliberate look, because nothing
mechanical rules on either. Each task's **Files** line: a task that ships
behaviour and names nothing outliving the run closes green and leaves the branch
with no coverage, because the `verify` command that proved it is scaffolding and
goes away with the plan. And whether any gate is broad enough to catch a
regression the others do not name — the driver re-runs every done task's
`verify` and nothing else, so that list is the whole regression net.

Because every done task's `verify` re-runs on every later iteration, a `verify`
command can read `LOOP_ACTIVE_TASK` (the task this iteration is working) and
`LOOP_GATE_TASK` (the task whose `verify` is currently running) as ordinary
environment variables. They are equal for a task's own gate and differ when
the gate is running as a regression check of some other, already-done task —
which is the case a `git diff HEAD` scope guard needs to consult before
treating everything uncommitted in the tree as its own. Neither variable is
set outside a gate run, so a work or review session never sees them.

### Amending the plan

The plan is yours **between** runs and the driver's **during** one. Use
`.loop/amend.sh` rather than editing JSON blind — every operation validates and
re-renders, so a mistake surfaces now instead of several minutes into the run:

```bash
.loop/amend.sh show                       # the plan, or `show T4` for one task
.loop/amend.sh verify T4 'uv run pytest -q tests/test_x.py'
.loop/amend.sh reset  T4                  # back to pending, attempts 0
.loop/amend.sh note   T4 'the fixture moved to tests/data'
.loop/amend.sh drop   T4                  # refuses if anything depends on it
.loop/amend.sh check                      # after ANY hand-edit
```

`check` is the important one. It verifies the schema, that every task has a gate
and criteria, that dependencies resolve and contain no cycle, that no ids are
duplicated — and it **runs every pending task's gate to warn you about any that
already pass**, because a gate that is green before the work exists proves
nothing.

Hand-editing `.loop/state/state.json` is still fine — it is only JSON — but run
`.loop/amend.sh check` afterwards.

Anything structural (re-scoping, adding tasks) is better done by editing the
brief and re-planning than by patching state.

Continue with no state edit:

```bash
.loop/run.sh
```

## 6. Running

```bash
.loop/run.sh docs/briefs/000N-....md              # plan, then iterate
.loop/run.sh --plan-only docs/briefs/000N-....md  # plan, commit it, stop
.loop/run.sh                                      # resume
```

**`--plan-only` is worth making a habit.** The plan is the highest-leverage
artefact the loop produces: every gate the rest of the run is measured against
was authored in that one session, and a weak `verify` silently lowers the bar
for everything after it. One planning session is cheap; the iterations are not.
Stop, read `.loop/state/plan.md` and the `verify` commands in
`.loop/state/state.json`, adjust with `.loop/amend.sh`, then `.loop/run.sh` with
no argument to execute what you approved.

Budgets are **per-run** and checked **between iterations**, so raising one and
re-running always works with no state edit. Defaults: 30 iterations, $40,
3 attempts per task, halt above 3.0 iterations-per-closed-task.

| Variable | Default |
| --- | --- |
| `LOOP_MAX_ITERATIONS` · `LOOP_COST_CEILING` | 30 · 40 |
| `LOOP_MAX_ATTEMPTS` · `LOOP_STALL_LIMIT` | 3 · 2 |
| `LOOP_CONVERGENCE_MAX` · `LOOP_CONVERGENCE_MIN` | 3.0 · 6 |
| `LOOP_PLAN_MODEL` · `LOOP_WORK_MODEL` | opus · sonnet |
| `LOOP_ARCHIVE_TRANSCRIPTS` | 0 |

Expect **~$1 per iteration** on a greenfield plan. Incremental work inverts the
profile — planning was 42% of run 3's cost, because the planner must read
existing code before it can write gates against it.

## 7. Watching it

**In the terminal you started it in.** The driver prints a block after every
iteration:

```
iterations · tasks closed · iterations per closed · gate failures
review rejections · attempts burned · no-progress streak · estimated spend
```

**If you backgrounded it** — and you probably should, since a run takes tens of
minutes — the same output is on disk. The run announces its own path in the
preflight block (`telemetry dir .loop/state/runs/<branch>/<timestamp>`):

```bash
tail -f .loop/state/runs/<branch>/<timestamp>/loop.log     # follow it live
ls -dt .loop/state/runs/*/*/ | head -1                     # the newest run
```

**Between runs**, `.loop/state/plan.md` and the plan's journal are re-rendered after
every iteration, so they are current even mid-run — and they read better on a
phone than a log does.

**Healthy is `iterations per closed` near 1.0.** Climbing means re-work; above
3.0 (after 6 iterations) the run halts itself. These exist because every other
mechanism judges a tick against its task, and nothing else judges the run
against the point of the run.

## 8. When it stops

| Status | Exit | What to do |
| --- | --- | --- |
| complete | 0 | read the journal, open a PR |
| preflight failed | 1 | fix what it named — it names one thing |
| blocked | 2 | a task burned its attempts; read its `notes` in `state.json` |
| stalled | 3 | two iterations with no recorded progress; read the run dir |
| max iterations | 4 | raise `LOOP_MAX_ITERATIONS`, re-run |
| not converging | 5 | **stop and look** — the run is going nowhere |
| cost ceiling | 6 | raise `LOOP_COST_CEILING`, re-run |
| session error | 7 | a `claude` session died; see the run dir |
| repeat blocked | 8 | a task blocked twice with nothing changed between attempts; read the first diagnosis in its `notes` |

Complete, max-iterations, stalled and cost-ceiling resume by just re-running.
Blocked, not-converging, session-error and repeat-blocked want a human first.

## 9. Reading what happened

| | |
| --- | --- |
| `.loop/state/plan.md` | **where are we** — every task, status, attempts, criteria, gate |
| `.loop/state/journals/<plan-id>.md` | **what happened** — planner's report,one entry per iteration, outcome and signals |
| `runstat summary <run-dir>` | per-phase cost, turns, wall time; flags errors and permission denials |
| `runstat signals <run-dir>` | the eight run-level signals |
| `runstat review <run-dir>` | what the reviews ruled, their findings, and coherence checks |
| `runstat compare <a> <b>` | two runs side by side |

Run dirs are `.loop/state/runs/<branch>/<timestamp>/`.

## 10. Running several loops at once

**One loop per git worktree.**

```bash
git worktree add ../loop-006 006-some-plan
cd ../loop-006 && .loop/run.sh docs/briefs/0006-....md
```

Git refuses to check one branch out twice, so separate worktrees are necessarily
separate branches. Two loops in **one** working tree share `.loop/tmp/proposal.json`,
so one's review could pass a task on the other's evidence — the driver takes a
lock (`.loop/tmp/.running`) and refuses.

## 11. Merging

Branches are **squash-merged and never deleted** — they hold the per-iteration
commits, which are the evidence each task was done by a separate fresh session.

`.loop/state/state.json` belongs to the branch. Merging main *into* a branch must
preserve the branch's copy; main's copy is meaningless. This is deliberately not
mechanised — see `.loop/README.md` for why `merge=ours` is a trap — so resolve by
hand:

```bash
git checkout --ours .loop/state/state.json && .loop/render-plan.sh
git add .loop/state/state.json .loop/state/plan.md
```

A branch that inherits foreign state resets it when you pass a brief, and
refuses when you do not. **The brief, not the branch, decides.**

## 12. Changing the loop itself

```bash
.loop/tests/run-all.sh                                     # free, offline, always
.loop/tests/reviewer-calibration/run-calibration.sh        # ~$1.20, calls a model
```

**Run the suite before and after any change to `.loop/run.sh`.** And when you add
a test, prove it can fail — break the thing deliberately and watch it go red.
Three tests in this repo have looked green while testing nothing, each asserting
something true in both the good and the bad case.

Changing `.claude/skills/loop-review/SKILL.md`? Re-run the calibration and
compare against [the baseline](tests/reviewer-calibration/RESULTS.md).
A review contract that stops catching planted defects has regressed, whatever
its prose says.

## 13. Troubleshooting

| Symptom | Cause |
| --- | --- |
| Preflight: "workspace NOT trusted" | §2. Silently voids all permission rules. |
| "a loop is already running in this working tree" | Two loops, one tree. Use a worktree, or clear a stale `.loop/tmp/.running`. |
| "state.json holds plan X for '\<other brief\>'" | The branch inherited another plan. Passing your brief resets it. |
| "belongs to branch … no way to tell" | Ambiguous with no brief. Pass one. |
| A scaffolding task can never pass | `uv run pytest` exits **5**, not 0, on zero tests collected. Gate on the artefact, not an empty suite. |
| Run halts "not converging" | Re-work loop. Read the failing task's `notes` — it is usually one gate that cannot be satisfied as written. |

---
name: vloop-t1-triage-inputs
description: What the T1 triage must take up — the open defects, the findings carried out of B9–B11 and the v0.8.0 release, and the questions the operator raised — so the triage starts from a list, not from memory. Read first when the triage begins.
kind: design-note
status: draft
created: 2026-10-06
---
# T1 — triage inputs

The triage is a **design act, not a loop brief** (`docs/design-notes/vloop-v2.0-roadmap.md`
→ T1): read the defects (by origin and catcher) and the interventions (by kind ×
agreement × `automatable`) of B1–B11, and propose **skill improvements** and
**repo-specific guidelines that could automate interventions**. Its proposals
seed later briefs, B12 among them (code health, which the triage may reshape).

State when this note was written: vloop **v0.8.0** is released and the repository
is **public** (2026-10-06). The v2.0 series (B9–B11) was renumbered to 0.8.0
before going public: `docs/guide/concepts.md` → "What changed in 0.8". The
roadmap file still says "v2.0" in its name and rows — part of the triage's
housekeeping.

## The data

- Run records and operator notes: `docs/briefs/B20261003-2049-gate-model.loop-brief.md`,
  `B20261004-1434-interventions-options`, `B20261005-0933-quality-pass`, and the
  earlier briefs.
- Defects: `vloop defect list`; interventions: `vloop intervention list`,
  `vloop metrics --interventions` (and `--by phase`). Records from 2026-10-04
  on carry options, the recommendation and the operator's decision.
- `docs/design-notes/vloop-interventions-b1-b4.md` — the first analysis, by hand.

## Open defects

- `D20261004-1309` — `plan-checks-its-gates`' fixture grader cannot see the gate-folder files (judge reads a truncated trace; a file focus cannot name a folder). Needs a deterministic grader.
- `D20261005-0757` and `D20261006-1028` — `operate-proposes-options`: its scenario is unrealistic (plan and work committed on `main`, a missing work branch, a one-task plan for a larger brief), so the session rightly proposes a restart and the gate-only judges fail. Redesign the scaffold.
- `D20261005-1125-the-driver-s-git-config-guard-cannot-tel` — the `.git/config` guard cannot tell an editor's write (VS Code's `vscode-merge-base`) from a gate's; B11's first acceptance halted with exit 9.
  **Second occurrence**, in another repository on another machine:
  `~/source/incostas/visum-monorepo`, `D20261007-1243` (a plan-only run halted
  with exit 9 while VS Code was open). A recurring field problem, not a one-off.
- `D20261005-1125-an-exit-9-during-plan-acceptance-discard` — an exit 9 during acceptance discards a finished, unstamped plan ($8.23 lost in B11) and its run folder.
  **Second occurrence** in the same visum-monorepo run: $5.51 lost, and its
  run folder is gone, so `vloop metrics` reports $18.84 for the brief without
  it (`D20261007-1451-el-coste-de-vloop-metrics…`). A second consequence:
  metrics under-report what a brief cost.
- `D20261006-1102` — the module path has no `/v2`. **Moot since the renumbering to 0.8.0**; close it with that reason.

## Findings to take up

1. **Interactive sessions do not get `/vloop:operate` from the CLI alone.** The
   skill lives in the plugin; `vloop run` loads it into its own sessions, but a
   user's interactive Claude Code session has it only if the plugin is installed
   (`claude plugin install vloop@vloop`) or the session starts with
   `--plugin-dir "$(vloop plugin path)"`. Meanwhile `vloop doctor` **passes** the
   plugin check ("supplied by vloop run") and the `CLAUDE.md` section `vloop init`
   writes tells Claude to follow `/vloop:operate`, which may not be loaded.
   Candidates: `doctor` warns when no interactive plugin is enabled; `init`'s
   next steps say it; the `CLAUDE.md` section says how to get the skill. Raised by
   the operator, 2026-10-06.
2. **`go install` builds report "commit unknown".** The binary could read the
   module version Go embeds (`runtime/debug.ReadBuildInfo`) so installed copies
   identify themselves; only the stamped build carries a commit today. Limit:
   `go install module@version` embeds the module version (`Main.Version`,
   e.g. `v0.8.0`) but **no** VCS settings. Go records `vcs.revision` only when
   it builds inside a git checkout. So the fix can replace "commit unknown"
   with the module version (e.g. `v0.8.0 (module)`) but cannot recover the sha.
   Until then the workaround is a stamped `go install`:
   `go install -ldflags "-X main.version=0.8.0 -X main.commit=<sha>" github.com/mvelosop/vloop/cmd/vloop@v0.8.0`,
   with the sha from `git ls-remote https://github.com/mvelosop/vloop 'v0.8.0^{}'`.
   Raised again by the operator, 2026-10-07.
3. **Masking misses the dashed `-Users-<name>-` form** of a path (Claude Code's
   project-folder slug); one permission-denial record in B9's run kept the
   username (`.vloop/state/runs/B20261003-2049-gate-model/20261003-215555/sessions/001-plan.json`).
4. **The generated `CLAUDE.md` section and this repository's own rules** now point
   at different operator playbooks (`/vloop:operate` vs
   `.claude/skills/vloop-operator`); both are right, but side by side.
5. **Do `vloop run`'s sessions load the user's enabled plugins?** They run with
   `--setting-sources project` and `--strict-mcp-config`, so they should not;
   confirm with one cheap session. A copy of the vloop plugin installed *and*
   passed with `--plugin-dir` in one session loads two plugins of one name — say
   in the guide not to combine them.
6. **Operator-side lessons from B10–B11** for the operator skills: watch the
   driver by its process id (a `pgrep -f` pattern matched the watcher itself and
   hid a halt); close the editor (VS Code) on the repository during a run, or fix
   the guard (above); a brief's must-change list keeps missing tests that pin
   counts and lists (schema list, embedded-schema count, init's output).
7. **Gate-review misses** to learn from: an unpassable gate needing a session to
   write under `.vloop/` (B10 T9/T10), a grep that could not match a message-less
   pass (B11 T12). Both are checks the gate review's "three ways" could make
   explicit.
8. **Releases are manual ceremony** (five steps in `docs/guide/concepts.md` →
   "The stamped release build"): a `vloop release` command is the recurring
   "what would automate it".
9. **Session time is the CLI's `duration_ms`, which can be badly wrong.** In
   visum-monorepo (`D20261007-1451-vloop-metrics-subestima-los-tiempos…`) the
   plan session's record says `duration_ms: 20083` (20 s) and `turns: 3`, yet
   it wrote 309k output tokens and ran from 11:44:11Z to the gate review's
   start at 12:23:22Z, about 39 minutes. `vloop metrics` showed the plan at
   0.3 min and the run at 30.7 min of a real ~71. The driver copies the
   figure (`internal/driver/session.go`) rather than timing the session it
   started. Why the CLI's figure is short (subagents? only the last turn?) is
   unconfirmed. Candidate: record the driver's own wall-clock time per session,
   keeping the CLI's figure beside it. Relates to field note 4 (elapsed time
   during a run).
10. **A denied Bash call is recorded without its command.** The session record
    keeps only `tool_name` and a `file_path`
    (`internal/driver/session.go`, the `permission_denials` loop), deliberately
    dropping the rest of `tool_input`. In visum-monorepo 5 Bash denials (gate
    review 1, T1 2, T2 1, review 1) cannot be analysed; only one is explained,
    by the journal (`D20261007-1451-5-denegaciones-de-bash…`). Candidate: keep
    the command, masked and truncated. Depends on finding 3 (masking misses a
    path form). The gate-review denial relates to finding 7.

## Field notes: first use of 0.8.0 in another repo

Raised by the operator while using v0.8.0 outside this repository, 2026-10-06.
Batched here for the triage; nothing below is fixed or recorded as a defect yet.
**The batch closed on 2026-10-07**, when the operator started the triage. Open
offer: run field note 3's no-plugin test in a scratch repo.

1. **No documented way to retire the shell loop** from a repo that had it
   installed (`.loop/install.sh`) and now runs `vloop`. Few users will need it,
   but the operator has several such repos. The footprint outside `.loop/` is
   what has to go: `.claude/skills/loop-{plan,work,review}` (they keep showing
   up as skills in interactive sessions), the `<!-- loop:begin -->…<!-- loop:end -->`
   block in `CLAUDE.md` (it points at `.loop/manual.md` and `state.json`; its
   "Rules for any session" are partly not in vloop's section — rule 1 on
   `~/.claude`), `.claude/loop-knowledge.md` (vloop never reads it), and
   `.loop/tmp/`. Inside `.loop/`, only `state/` is worth keeping, with a
   one-line `RETIRED.md`: vloop reads `.loop/state/state.json` and
   `.loop/state/runs/` (`internal/runs`, the `ShellLoop` layout) in any repo,
   so metrics and defects keep the shell loop's history; the mechanism files
   (driver, manual, tests, examples) remain in git history. vloop already
   excludes `.loop/**`. Side finding: `docs/guide/metrics.md` and
   `docs/guide/concepts.md` say the legacy layout is read "in this repository
   only", but the code reads it wherever it is. Candidates: a short "Retiring the shell loop" section in the
   guide, an operator-skill step, or a `vloop init` check that offers it when
   it finds `.loop/.installed`. Relates to finding 4 (two rule sets side by
   side).
2. **The name is never explained.** "vloop" comes from *verification loop* —
   the gate and the independent review are what the loop is about — but no
   document says so (`README.md`, `docs/guide/`, the plugin's description).
   Candidate: one clause in the README's opening sentence.
3. **Onboarding was smooth.** The operator asked to run a brief in a repo that
   had never been through `vloop init`; the operator skill ran `init` itself
   and the run went through. Worth keeping as behaviour the skill is tested on.
   **Test idea** from it: the same request with the plugin *not* installed. No
   `/vloop:operate`, and in an un-initialized repo no `CLAUDE.md` section either,
   so the session has only `vloop --help` to go on. Does it find its way, or
   does it improvise? The answer sizes finding 1 (what `doctor`, `init` and the
   `CLAUDE.md` section should say about getting the skill), and could become an
   eval scenario for the bare-CLI path.
4. **No periodic feedback that a run is alive or what it is spending.** The
   operator worries about token consumption during a long run. Today
   `run.cost-ceiling` (default $40, exit 6) is a stop, not feedback: it is
   checked between sessions, and a session's cost is known only when it ends
   (`--output-format json`), so one session — up to `run.session-timeout`, 60
   minutes — can overshoot it. The log prints one line per iteration and the
   verdict, with no cost and no clock, and the operator skill waits for the exit
   notification without polling, so nothing is seen until the run ends.
   Options discussed; the operator chose to carry these:
   - **Running cost in the driver's output** — after each session, e.g.
     `work T3 · 6m12s · $1.84 · run $9.40 / $40`. Deterministic, no tokens,
     readable with `tail -f`; the base for the rest.
   - **Threshold warnings** — the driver says so at 50% and 80% of the ceiling.
   - **`vloop status` shows the run** — cost so far, the session in flight and
     its elapsed time, checkable from another terminal at no token cost.
   - **Event-driven notification from the operator skill** — the interactive
     session watches the log (Monitor, filtered to iteration, verdict and halt
     lines) and sends a push notification per event; it wakes on events, never
     on a timer.
   Not carried for now: streaming sessions (`stream-json`) for a heartbeat and
   an in-session ceiling — the real fix for the overshoot, but a larger driver
   change; and timer polling (`ScheduleWakeup`), which costs more and tells less.
5. **Tell users they need not learn the commands.** The intended DX: the user
   knows *what the CLI can do*, not its syntax; the operator skill
   (`/vloop:operate`) is the natural-language interface that drives it ("run
   this brief", "what did that cost?", "record this as a defect"). The docs
   currently lead with commands (`README.md` → "Quickstart", "Everyday
   commands") and say nothing of this. Candidates: a short paragraph in the
   README before "Everyday commands", framing that section as *what you can
   ask for* and the command reference as what the operator uses; the same line
   in `init`'s next steps. Depends on finding 1 — the promise holds only where
   the skill is loaded — and field note 3 (onboarding by asking) is the
   evidence it works.
6. **"The driver" is used everywhere and defined nowhere a user reads.** The
   word appears in `README.md` (2), `docs/guide/` (13 across five files) and
   the plugin's skills (40, most in `plan` and `work`), yet the only
   definition is one row of the domain glossary
   (`docs/domain/domain-model.md`: "the program that owns status, runs gates
   and commits"), which users never see. It is a central concept: the
   deterministic half of the loop, which plans nothing and judges nothing but
   owns every status change, gate run and commit, and is the boundary the
   permission fence only advises on (exit 9). Candidates: a "The driver"
   paragraph in `docs/guide/concepts.md` (who does what: driver vs plan, work
   and review sessions vs operator), linked from the README's "The loop in one
   paragraph"; check the skills' uses against that definition.
7. **The install line assumes Go's bin directory is on `PATH`.** `go install`
   writes the binary to `$GOBIN`, or `$(go env GOPATH)/bin` (`~/go/bin` by
   default), and never touches `PATH`; the official Go installer on macOS and
   most Linux packages do not add that directory either. A first-time Go user
   can run the README's install line and get `command not found: vloop`.
   `README.md` → "Install" says nothing of it. Candidate: one comment line after
   the install command (`export PATH="$PATH:$(go env GOPATH)/bin"`), and the
   same hint wherever the guide repeats the install, naming the file per shell:
   - **zsh** (the macOS default): `~/.zprofile`, read once at login (macOS
     terminals open login shells); `~/.zshrc` also works, read by every
     interactive shell.
   - **bash**: `~/.bash_profile` on macOS, `~/.profile` on most Linux.
   - **PowerShell (Windows)**: the Go MSI installer usually adds
     `%USERPROFILE%\go\bin` to the user `PATH` already. If not, set it once
     for the user, which new terminals inherit:
     `[Environment]::SetEnvironmentVariable('Path', [Environment]::GetEnvironmentVariable('Path', 'User') + ";$(go env GOPATH)\bin", 'User')`
     (not `$env:Path`, which merges the machine `PATH` into the user one),
     or per session in `$PROFILE` with `$env:Path += ";$(go env GOPATH)\bin"`.
     The user-variable form is preferred because it also reaches `cmd` and
     apps started outside PowerShell.

   Raised by the operator, 2026-10-07.
8. **When gate folders are cleared is documented nowhere.** `.vloop/state/gates/`
   is never cleared on `vloop brief close`. It goes only when a plan is replaced:
   when the next brief is planned (`internal/driver/plan.go`, the reset of
   another brief's plan, with `state.json` and `plan.md`), or when a new plan is
   rejected as unfit. Between briefs the last plan's folders stay in the working
   tree and in git (today B11's `T1`–`T17`); earlier ones live in history. That
   seems deliberate (the last plan's gates stay checkable), but
   `docs/guide/concepts.md` describes the gate folder without saying when it
   goes. Candidate: one sentence there on the lifecycle of `state.json`,
   `plan.md` and the gate folders. Raised by the operator, 2026-10-07.
9. **A docs-only brief measures as zero delivered code.** In visum-monorepo
   (`D20261007-1451-el-tama-o-entregado…`), `metrics.stacks` lists only
   `csharp@legacy/...` (what `vloop init` detected); the brief delivered 3487
   lines of shell scripts and docs, so `code` and `test` are 0 — correct for
   that config — while the same row shows `efficiency 100%`. Mostly a config
   issue, low priority. Questions for the triage: should `init` detect shell
   and docs, and should metrics say "nothing measured" instead of a 0 beside a
   percentage? Raised by the operator, 2026-10-07.

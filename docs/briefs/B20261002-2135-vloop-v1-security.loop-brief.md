---
name: B20261002-2135-vloop-v1-security.loop-brief
description: Make the driver, not the fence, the boundary of what a run can do — sessions and gates confined and timed, secrets kept out of what is committed, only ready briefs planned from a clean tree — and land B7's carry-overs, from the v1.0 review's security pass
kind: brief
status: consumed
created: 2026-10-02
seeds: The first brief vloop runs on itself; it hardens the driver that runs the next one (B9, the quality pass)
depends-on: [B20261001-1025-vloop-skills.loop-brief]
---
# Brief — vloop B8: the v1.0 security pass

- **Status:** consumed — closed 2026-10-03 as run B20261002-2135-vloop-v1-security. **Do not re-plan from this brief.**
- **Starting point:** extends `main` at the v0.7.0 release (`e18f80e`, tag
  `v0.7.0`, PR #9), after the cut-over (`271e5b0`, PR #8). The planner pins the
  base it plans from as the **base** every gate compares against — never
  `HEAD~n`.
- **Produced by:** the v1.0 review's design act, 2026-10-02 — five read-only
  review passes (command execution; file writes, paths and secrets; what
  sessions may do; docs against the binary; code health), every finding below
  checked against the code by the architect. The quality findings are B9's and
  are recorded in `docs/design-notes/vloop-v1-review.md`.
- **Run by:** `vloop run` itself, from the released v0.7.0 binary — the first
  brief vloop builds of itself.

## What it is

Today the fence is the only thing between a session and the run's integrity,
and it is a set of command prefixes that a session can step around: it can
rewrite a gate and run it, plant a git hook that the driver's own commit runs,
or zero the cost records the ceiling reads. Gates can hang the run forever, a
short user name corrupts every record, an untracked `.env` rides into the plan
commit, and a bare `vloop run` plans a draft brief. After this run **the driver
is the boundary**: it confines and times what sessions and gates do, detects
what they changed, keeps secrets out of what it commits, and starts only from a
ready brief and a clean tree. The fence stays, tightened, as a first line that
saves attempts.

The blind spot these defects share: every check so far asked "did the session
use a denied command?", never "what did the run's own inputs and git state look
like before and after?". Each fix below closes it by checking the outcome, not
the command.

## Why this shape, and what was rejected

Decided with the operator on 2026-10-02. **Do not re-litigate these** — the
planner inherits them.

- **The review was done in the design act; this run fixes what it found.** A
  loop task's gate cannot judge "review the codebase", but it can judge a fix
  with a failing test. *Rejected:* audit tasks inside the run — their gates can
  check a findings document's shape, not its thoroughness.
- **Security and correctness here; quality in B9.** About 45 findings is too
  much for one run, and B9 should run under the hardened driver. v1.0 is tagged
  after B9. *Rejected:* one brief of 25+ tasks; three briefs.
- **The driver is the boundary; the fence is advisory, but tightened.** Sessions
  keep `--permission-mode auto`. *Rejected:* `dontAsk` — the evals showed
  sessions stall when every `$(…)` and heredoc is refused, and a refused command
  is still not a checked outcome.
- **Gate timeout 15 minutes, session timeout 60, configurable.** *Rejected:*
  5 and 30 — they would kill slow but legitimate e2e suites and long work
  sessions.
- **A run refuses a dirty tree** and names the files. *Rejected:* a warning —
  the plan commit would still sweep in a secret.
- **Keep-awake is on by default, on all three OSes**, and can be turned off;
  failing to take the hold is a warning, never a failure. *Rejected:* opt-in —
  the default run would stall as B7's acceptance run did.
- **The README keeps its cap.** The completeness test moves to the guides, so
  the README no longer has to repeat every flag and key; B9 rewrites it.
  *Rejected:* raising the cap.
- **B8 is run from the released v0.7.0**, never from a binary built from the
  tree it changes (design session, section 6).

## Binding references

- `docs/domain/domain-model.md` — S-1, S-2, S-3, P-2, P-4, P-5, R-1, R-2, R-3, B-2, B-6, M-7, C-1, C-3: the invariants each fix enforces or amends
- `docs/domain/execution/session.md` — what a session never does, the fence, and the session record this run hardens
- `docs/domain/execution/run.md` — the run folder, commits, budgets and endings, which gain timeouts, a clean-tree preflight and a wider exit 9
- `docs/domain/execution/task.md` — the gate: who writes it, where it runs, and the rewrite protection F9 widens
- `docs/domain/briefing/brief.md` — status, lifecycle and dependencies, which F1 makes `vloop run` enforce
- `docs/domain/platform/platform-context.md` — configuration and CLI conventions for the new keys and messages
- `docs/guide/configuration.md` — every config key; C3 makes it the place a key must be documented
- `docs/design-notes/vloop-v1-review.md` — the full review: B8's findings with their evidence, and what is B9's (out of scope here)

## Behaviour contract

Every fix ships with a test that fails against the base and passes after it
(P-4 for fixes: the failing case first). Messages below are exact where quoted;
`<…>` marks a value. Every refusal is one line on stderr starting with
`vloop: `.

### F1 — `vloop run` plans only a ready, checked brief

**Defect.** With no argument, `vloop run` takes the brief that sorts last in
`docs/briefs/`, whatever its status (`LatestBrief`, `internal/driver/plan.go`);
named or not, it never checks `status`, `vloop brief check` or `depends-on`.
Right after `vloop init`, a bare `vloop run` plans the draft starter brief and
spends money on it. B-2 and B-6 are documented but not enforced.

**Required.** Before planning a brief (named or chosen), `vloop run` refuses
with exit 1 unless the brief's `status` is `ready`, `vloop brief check` reports
no problem for it, and every `depends-on` entry is `consumed`:

```
vloop: <brief> is <status>, not ready — set status: ready once it passes vloop brief check
vloop: <brief> fails vloop brief check (<n> problem(s)) — run vloop brief check <brief>
vloop: <brief> depends on <name>, which is <status>, not consumed
```

With no argument, the brief chosen is the newest **ready** one; with none,
`vloop: no ready brief in docs/briefs/ — name one, or set status: ready on a checked brief`.
Resuming a plan that already exists does not re-check its brief: the brief was
checked when it was planned.

**Gate.** Driver tests: a draft newest brief beside an older ready one → the
ready one is planned; only drafts → the refusal above, exit 1, no session run;
a ready brief with a `depends-on` that is `ready` → refusal naming it.

### F2 — a run starts from a clean tree

**Defect.** `git add -A` in the plan commit and every iteration commit sweeps
in anything untracked and not ignored — an `.env`, a credentials.json.

**Required.** Planning refuses, exit 1, when `git status --porcelain` lists
any modified, staged or untracked-not-ignored path, naming up to 10 of them
and the count of the rest:

```
vloop: the tree is not clean — commit, ignore or remove these first: .env, notes.txt
```

`.vloop/tmp/` is ignored by `vloop init` and never counts. Resuming a run also
requires a clean tree, except for changes under `.vloop/state/` — the
operator's `vloop task verify|reset|note|drop|set` edit the plan there, and the
next iteration commits them.

**Gate.** An e2e test with an untracked `.env` → refusal, exit 1, nothing
committed. Existing run tests whose repos start dirty are fixed to start clean
(see Constraints).

### F3 — gates and sessions are timed, and leave nothing running

**Defect.** Gates and sessions run with `exec.Command`, no deadline and no
process group; gate output is read through a pipe, so a gate that leaves a
background server (`npm run dev & …`) or runs a watcher never returns, and
`vloop run` hangs forever. Orphans pile up across re-runs.

**Required.**
- New config keys, whole minutes, at least 1: `run.gate-timeout` (default 15)
  and `run.session-timeout` (default 60), with `VLOOP_RUN_GATE_TIMEOUT` and
  `VLOOP_RUN_SESSION_TIMEOUT`.
- Each gate and each session runs in its own process group (a Job object on
  Windows). On timeout the whole group is killed, and a short wait bounds the
  read of its output.
- A timed-out gate is a failed gate: its log ends with the line
  `vloop: gate <id> timed out after <n> min`. A timed-out session is a session
  error (exit 7).
- When the gate's shell exits, anything it left running in its group is killed
  too, so `cmd & exit 0` does not hang the run or leave a process behind.

**Gate.** Unit tests (POSIX): verify `sleep 1000 & exit 0` returns promptly and
leaves no `sleep` running; a verify that sleeps past a 1-minute test timeout
(the timeout is injectable in tests, in seconds) fails with the line above. A
Windows build of the process-group code compiles (`GOOS=windows go build`).

### F4 — the driver's git ignores what a session can plant in `.git/`

**Defect.** The fence denies no write under `.git/`, and the driver runs plain
`git add`, `git commit` and `git status`, so a session-written
`.git/hooks/post-commit`, `core.hooksPath`, `core.fsmonitor` or a
`filter.*.clean` runs inside the driver's own commit — after the refs check,
which then never fires.

**Required.**
- Every git command the driver runs disables repository hooks and fsmonitor
  for that command (`-c core.hooksPath=<an empty directory under .vloop/tmp/>`,
  `-c core.fsmonitor=false`) and reads paths unquoted
  (`-c core.quotepath=false`). So a repository's own hooks never run on the
  driver's commits: the preflight's pre-commit warning becomes
  `  ! a pre-commit hook is active (<path>) — the driver's commits do not run repository hooks; put what it checks in the gates`.
- The driver hashes `.git/config` and the hooks directory (as
  `git rev-parse --git-path hooks` names it) before each session and gate.
  A change halts the run with exit 9 — `vloop: <phase> changed .git/config — nothing was committed; restore it, then re-run`
  (or `the git hooks`) — and nothing is committed.

**Gate.** E2e with the stub claude: a work session writes
`.git/hooks/post-commit` that creates a branch → the run halts with exit 9 and
the hook never ran (the branch does not exist); a session that adds
`[core] fsmonitor = …` → exit 9.

### F5 — a gate that moves refs is caught like a session

**Defect.** The refs check wraps sessions only; a gate that runs `git commit`,
`git reset --hard`, `git stash` or `git checkout -- .` goes unnoticed, and a
task can pass on a reverted tree.

**Required.** The driver snapshots refs before every gate run (a task's own
gate and every regression gate) and halts with exit 9 if they moved,
naming the task whose gate moved them:
`vloop: the gate of <id> moved git refs — nothing was committed; restore them, then re-run`.

**Gate.** E2e: a verify of `git commit --allow-empty -qm x && true` → exit 9,
nothing committed by the driver.

### F6 — sessions cannot write the driver's inputs

**Defect.** The fence allows `Edit` and `Write` everywhere but `~/.claude`. A
session can zero `cost_usd` in earlier session records (the ceiling re-reads
them every iteration and never fires), set `model.review` in
`.vloop/config.toml` (re-read every iteration) to weaken its own reviewer, or
change a defect's status.

**Required.**
- **Per-phase fences.** The plan session's fence allows writing
  `.vloop/state/state.json` and `.vloop/tmp/**`; the work and review fences
  allow writing nothing under `.vloop/` but `.vloop/tmp/**`. All three deny
  `Edit` and `Write` on `.git/**`.
- **The driver keeps what it reads in memory.** Spend is summed from the
  sessions as they finish, never re-read from disk; model and effort per kind
  are resolved once at the start of the run.
- **Changes are detected and undone.** The driver hashes `.vloop/config.toml`,
  `.vloop/defects/`, `.vloop/interventions/` and the run's own session and
  report records before each work and review session. A change is restored,
  written to `run.log` as `vloop: <phase> session changed <path> — restored`,
  and, for a review session, its verdict is forced to FAIL — as for
  `state.json` today.

**Gate.** E2e with the stub claude: a work session that zeroes earlier
`cost_usd` values still hits a ceiling set just above the real spend (exit 6);
a work session that sets `model.review = "haiku"` → the next review runs with
the configured model and `run.log` names the restored file. A fence test: each
per-phase fence allows and denies the writes above.

### F7 — a session that runs a gate runs the driver's gate

**Defect.** `vloop task gate <id>` reads the verify command from
`.vloop/state/state.json` on disk. A work session can write a different verify,
run it (the fence allows `vloop task gate`, so it runs arbitrary commands
unclassified), and put the file back byte for byte; the state guard compares
bytes and sees nothing.

**Required.** The driver sets `VLOOP_PLAN_SHA256` — the hash of the
`state.json` it holds — in every work and review session's environment.
`vloop task gate`, when that variable is set and the file's hash differs,
refuses with exit 1 and runs nothing:
`vloop: the plan was changed during this session — gates run only from the plan the driver holds`.
The driver records the refusal in `run.log` when it finds it.

**Gate.** A unit test of `task gate` with a mismatched hash → exit 1, the
verify not run (it would create a marker file). The existing
`task_gate_test.go` behaviour without the variable is unchanged.

### F8 — a review session cannot change the work it judges

**Defect.** After review the driver checks only refs and `state.json`, then
commits with `git add -A`, so a reviewer's edits to source or tests are
committed unreviewed.

**Required.** The driver snapshots the tree (`git status --porcelain` plus a
hash of the working-tree diff) before the review session. Any change outside
`.vloop/tmp/` after it is reverted, listed in `run.log`
(`vloop: the review session changed <path> — reverted`), and the verdict is
forced to FAIL with the finding `the review session changed files`.

**Gate.** E2e: a stub reviewer that edits a source file and returns PASS → the
file is reverted, the iteration's verdict is FAIL, the task is not done.

### F9 — gate files are found by token, on every OS, for every done task

**Defect.** `gateFilesMoved` (`internal/driver/gates.go`) matches a changed path
as a substring of this task's verify (a.sh matches data.sh), never matches
`.\scripts\check.ps1` against `scripts/check.ps1`, misses octal-quoted
non-ASCII paths (the driver's git runs without `core.quotepath=false`), and
protects only files named by the current task's verify.

**Required.** A path is a gate file when it equals, as a whole token, a path
named in the verify of the current task **or of any done task**, after both are
normalized: `\` to `/`, a leading `./` removed, quotes stripped. Paths from git
are read unquoted (F4). Protection and restoration otherwise behave as today.

**Gate.** Unit tests: a.sh changed, verify `bash data.sh` → not a gate file;
verify `pwsh -File .\scripts\check.ps1`, `scripts/check.ps1` changed → gate
file; `prüfung.sh` edited → restored; a file named only by a done task's
verify → restored.

### F10 — the fence denies the bypasses the review found

**Defect.** The deny list matches only the plain forms: `git -C . commit`,
`git -c k=v push`, `git merge`, `cherry-pick`, `revert`, `am`, `pull`, `fetch`,
`push` to a URL, `config`, `notes`, `replace`, `filter-branch`, `worktree`,
`reflog`, `gc` are not denied; `find -delete`, `find -exec`, `find -fprint`,
`git diff|log|show --output=…`, `rm -fr` and `rm -r -f` pass as allowed or
unlisted; `vloop -C . task set …` and `vloop brief new` are not denied.

**Required.** Each form above is denied in every phase's fence, and every
command a skill tells a session to run stays allowed (the existing
skills-and-fence check). `docs/domain/execution/session.md` says plainly that
the fence is advisory, that the driver's checks are the boundary, and that a
consumer repo's own `.claude/settings.json` allow rules also apply to sessions.

**Gate.** A fence test lists every form above and asserts each is denied by
the embedded fence's deny rules, and lists the allowed commands the skills use
and asserts each is still allowed.

### F11 — masking replaces paths, never text

**Defect.** `Runner.Mask` replaces the user name anywhere, in the marshalled
JSON of a session record. A user `us` turns `"cost_usd"` into `"cost_USERd"`,
the ceiling reads 0 and never fires; a user `ion` breaks `"iteration"`; a
numeric home (a home directory named 1000) corrupts numbers into invalid JSON;
`HOME=/` makes every `/` `USER`; a home named al, a sibling of alice, turns
alice's paths into `~ice`; the forward-slash and lower-case forms of a Windows
home are not masked.

**Required.**
- Masking applies to string values before a record is marshalled, never to
  serialized JSON.
- The home is replaced by `~` only at a path boundary (followed by a separator,
  a quote or the end).
- The user name is replaced by `USER` only as a whole path component directly
  under the users directory that holds the home (macOS, Linux and Windows), never as
  free text.
- A home of `/` or `\` is not masked.
- On Windows, the home's `/` and `\` forms are matched case-insensitively.

**Gate.** Unit tests: a runner with user `us` and a stub session costing
$1.50 → the record parses and `spend` is 1.50; users `ion` and `1000` → records
parse with `iteration` and `duration_ms` intact; `HOME=/` → `a/b` unchanged;
home named al → a sibling alice's path unchanged; Windows forms masked (a
pure-function test, any OS). `TestRun10Containment` adds a short-user case and
checks for the user only as a path component.

### F12 — committed logs carry no secrets

**Defect.** Gates run with the full environment and their output is committed
in `gates/*.log`; a test that dumps the environment commits `ANTHROPIC_API_KEY`.
Session records keep raw `permission_denials`, whose inputs hold whole commands
and file contents.

**Required.**
- Before anything is written under `.vloop/state/`, every occurrence of the
  value of an environment variable is replaced with `<redacted:NAME>` when the
  variable's name contains `KEY`, `TOKEN`, `SECRET`, `PASSWORD` or `CREDENTIAL`
  (case-insensitive) and its value is at least 8 characters.
- Session records keep, per permission denial, the tool name and the file
  path when there is one — never the command or content.

**Gate.** A gate with verify `env` and `FOO_TOKEN=s3cr3t-value` set → its log
contains `<redacted:FOO_TOKEN>` and not the value; a stub session reporting a
denied `Write` with content → the record has the tool and path only.

### F13 — `.vloop/tmp/`, the lock and the run id are safe to trust

**Defect.** The driver reads `.vloop/tmp/proposal.json` and `verdict.json`
through symlinks (a session can link one to `~/.claude/.credentials.json`, and
`copyReport` commits it, even unvalidated); the lock is written through a link
and is check-then-write, so two runs can both start; `run_id` from a committed
`state.json` is used in paths unchecked (`"../../../../tmp/x"` writes outside the
repo).

**Required.**
- Handoff files are read only when they are regular files (no symlink); a report
  is copied into the run folder only after it validated.
- The lock is created atomically (exclusive create, no symlink following); a
  lock whose process is gone is replaced once; two concurrent `vloop run`s →
  exactly one runs, the other exits 1 naming the running pid.
- `run_id` must match `^[A-Za-z0-9._-]+$`, contain no `..`, and equal the run
  id of the plan's brief (`schemas/state.v1.json` gains the pattern); otherwise
  exit 1 before anything is written.

**Gate.** Tests: a symlinked `proposal.json` to a file holding `SECRET` → no
`SECRET` under `.vloop/state/`, the proposal treated as missing; two
goroutines acquiring the lock → exactly one succeeds; a plan with
`"run_id":"../x"` → exit 1, nothing written outside the repo.

### F14 — the export carries no local path or credential

**Defect.** `repo.remote` is exported as is: a local remote
(a clone of a local directory) leaks a path, and a URL's query (`?access_token=…`) keeps a
token (M-7).

**Required.** `repo.remote` is exported only as a URL or scp-like form, without
user info, query or fragment; a local path or `file://` remote is omitted
(the field is absent).

**Gate.** Export tests: a temp-path origin → no `remote`; an https origin with
user info and a query → host and path only.

### F15 — a gate runs only in a known shell, quoted right

**Defect.** The plan's `shell` is checked only when a plan is accepted; `vloop
task gate` loads the plan without checking it, so `"shell":"python3"` runs the
verify as `python3 -c …`. On Windows, `cmd /C` receives Go's MSVC-quoted
argument, which cmd.exe does not unquote, so `findstr "a b" f` arrives mangled.
Gates are run by two separate runners (`state.RunGate`, `Iterator.runGate`).

**Required.** One gate runner serves the driver and `vloop task gate`. It
refuses a shell outside `sh`, `bash`, `pwsh`, `powershell`, `cmd` with exit 1
(`vloop: shell <name> is not one of sh, bash, pwsh, powershell, cmd`). For
`cmd` on Windows the command line is passed verbatim as `/C <verify>`.

**Gate.** A unit test: shell `python3` → the refusal; the Windows command-line
builder as a pure function, tested on any OS, yields `/C findstr "a b" f`.

### F16 — the self-hosting guard fires

**Defect.** `go install` stamps no commit (`vloop version` prints
`commit unknown`), and `vloop doctor`'s self-hosting check passes any binary
without one, so the guard never fires.

**Required.** `vloop doctor`, in this repository (the module is
`github.com/mvelosop/vloop`), warns
`self-hosting: this binary carries no commit — build releases with the stamped build in docs/guide/concepts.md`
for a binary without a commit, and keeps its warning for one built from HEAD.
`docs/guide/concepts.md` documents the stamped release build (the
`-ldflags "-X main.version=<v> -X main.commit=<sha>"` command) and the release
steps: plugin.json version, tag, build.

**Gate.** Doctor tests: an unstamped build → the warning; a stamped build of
another commit → pass.

### C1 — the TypeScript and JavaScript presets count NestJS tests (D20261002-1425)

**Required.** The `typescript` preset's test globs gain `**/*.e2e-spec.ts` and
`**/test/**`; the `javascript` preset's gain `**/*.e2e-spec.js` and
`**/test/**`. A test/ directory's files are tests in those stacks.

**Gate.** Classification tests: test/links.e2e-spec.ts and
test/jest-e2e.json are `test` under `typescript`; src/app.ts stays `code`.

### C2 — `vloop run` keeps the machine awake

**Required.** A new key `run.keep-awake` (`on`, `off`; default `on`;
`VLOOP_RUN_KEEP_AWAKE`). While on, `vloop run` holds a no-idle-sleep hold for
its own lifetime: macOS `caffeinate -i -w <pid>`, Linux
`systemd-inhibit --what=idle --why="vloop run" …` when available, Windows
`SetThreadExecutionState(ES_CONTINUOUS|ES_SYSTEM_REQUIRED)`. If the hold cannot
be taken, `run.log` gets one line,
`vloop: could not keep the machine awake (<reason>) — a sleeping machine stalls the run`,
and the run continues. The hold ends when `vloop run` exits, however it exits.
`docs/guide/concepts.md` says what it does not cover: closing a laptop's lid.

**Gate.** Unit tests of the command each OS would run (pure functions); a
macOS-only test that the `caffeinate` child exits when the process it watches
does; `run.keep-awake = off` → no hold attempted.

### C3 — the README's completeness test moves to the guides

**Required.** The checks that every command, flag and config key is documented
move from the README to the guides: commands and flags to
`docs/guide/commands.md` (already generated), keys to
`docs/guide/configuration.md`. The README keeps its 200-line cap and is checked
only for commands, flags and config keys it names that do not exist, and for a
link to each guide. Its config-key table is replaced by a short pointer to
`docs/guide/configuration.md`. No other README rewrite (B9).

**Gate.** The moved tests pass; README under 200 lines; a fixture README
naming `--frobnicate` still fails.

### The domain, updated

`docs/domain/domain-model.md` changes, and nothing else in it:

- **S-2** gains: "The fence is advisory; the driver is the boundary. It halts
  (exit 9) if a ref, `.git/config` or the git hooks change during a session or
  a gate, and restores what a session changed among the driver's inputs."
- **S-4** (new): "A session writes nothing under `.git/` and nothing under
  `.vloop/` but `.vloop/tmp/` — the plan session also writes the plan."
- **R-3**'s exit 9 reads "refs or repository configuration moved".
- **R-4** (new): "A run plans only a `ready` brief that passes the check, with
  its dependencies consumed, from a clean tree; every gate and session runs
  under a timeout and leaves no process behind."

`docs/domain/execution/session.md`, `run.md` and `task.md` describe the
behaviour above where they describe the same thing today. The README's and
`docs/guide/concepts.md`'s exit-code tables follow R-3.

### Decided here, because the cited documents leave it underdetermined

- Timeouts are whole minutes in config; tests may inject seconds.
- Keep-awake is an enum `on`/`off` because config has no boolean type.
- Redaction keys on variable names, not on value shapes; a secret in an
  innocently named variable is not caught, and the guide says so.
- The tree snapshot of F8 is taken by git, not by walking the file system.
- Repository hooks do not run on the driver's commits (F4); what a consumer's
  pre-commit hook checks belongs in the gates. The preflight says so.
- F1's check runs `vloop brief check`'s own code, not a second copy of it.

### Violations the review must rule on

- A fix with no test that fails on the base, or a test that would pass on it.
- A check of a command where the brief pins a check of an outcome (F4–F8).
- Any change to what sessions are allowed that a skill relies on, without the
  skill and the skills-and-fence check following it.
- A message that differs from the one quoted here.
- Windows code that only compiles: where the brief names a pure function, it is
  tested on every OS.
- Scope from B9: the README beyond C3, error-message rewording beyond the
  messages above, help text, duplication clean-ups not needed by a fix.

## Worked example

With the stub claude of the run tests, in a temp repo with a ready brief:

```
newest brief draft, an older one ready; vloop run
  -> plans the ready one                                         exit 0
only drafts; vloop run
  -> vloop: no ready brief in docs/briefs/ — name one, or set status: ready on a checked brief   exit 1
untracked .env; vloop run <ready brief>
  -> vloop: the tree is not clean — commit, ignore or remove these first: .env                    exit 1
verify `sleep 1000 & exit 0`
  -> the gate returns; no sleep process remains                  (gate passes)
verify sleeps past the gate timeout
  -> gate log ends: vloop: gate T1 timed out after <n> min       (gate fails)
work session writes .git/hooks/post-commit creating branch evil
  -> vloop: work changed the git hooks — nothing was committed; restore it, then re-run           exit 9
     git branch --list evil prints nothing
verify `git commit --allow-empty -qm x && true`
  -> vloop: the gate of T1 moved git refs — nothing was committed; restore them, then re-run      exit 9
work session zeroes earlier cost_usd; ceiling just above real spend
  -> cost ceiling reached                                        exit 6
work session edits state.json verify, runs vloop task gate T1
  -> vloop: the plan was changed during this session — gates run only from the plan the driver holds   exit 1 (inside the session)
review session edits src and returns PASS
  -> the file is reverted; verdict FAIL; task not done
runner user "us", session costing $1.50
  -> session record parses; spend 1.50
gate `env` with FOO_TOKEN=s3cr3t-value
  -> gates/T1.log holds <redacted:FOO_TOKEN>, not s3cr3t-value
two vloop run at once
  -> one runs; the other exits 1 naming the running pid
```

**Real-data check.** On this repository, `vloop metrics --json` for every
consumed brief (B1–B7), run by the new binary, is byte-identical to the same
output captured from the base binary when the plan is made: masking,
presets and run-folder reading change nothing for a `go` repo whose records were
written by a long user name. `vloop doctor` with the released v0.7.0 reports
the self-hosting check as passed; with a build of HEAD, the warning.

## Out of scope

- Everything in `docs/design-notes/vloop-v1-review.md` marked for B9: the README
  rewrite, the guides' inaccuracies, error messages and exit-code mapping beyond
  the messages quoted here, help text, duplication and long functions, metrics
  correctness (local-time run folders, convergence ratio, first-pass after a
  regression), `gendocs` and the stray `--help` file.
- `--permission-mode dontAsk`, sandboxing gates (containers, seatbelt),
  network isolation of gates.
- Publishing a marketplace, signing releases, a release automation command.
- The horizon's "Now" entries (merge strategy, intervention fields).
- Any change to the review, work or plan skills beyond what F10 requires of
  the commands they name.

## Constraints

- Go as pinned in `go.mod`; `golang.org/x/sys` may be added for Windows Job
  objects and `SetThreadExecutionState`, and nothing else.
- Every task's gate builds for `windows` and `linux` as well as the host.
  OS-specific code lives in build-tagged files; the behaviour that can be a pure
  function (command lines, path forms, masking) is one, tested on every OS.
- Gates run on macOS's BSD tools: no empty alternatives in `grep -E`, no
  `sed -i` without a suffix, no `\+` or `\|` in basic regexes, no `date -d`.
- Sessions never move refs. Tests build their repositories in temporary
  directories and address them with `git -C`.
- **Gates follow the diff surface.** No task runs the whole test suite; the
  `cmd/vloop` run tests are gated per file or `-run` pattern.
- **No task may weaken a check to pass.** These earlier tests **must** change,
  and only as stated:
  - every `cmd/vloop` run test whose repository starts with untracked or
    modified files is changed to start clean (F2);
  - `TestRun10Containment` gains the short-user case (F11);
  - `internal/driver/session_test.go` masking tests follow F11's rules;
  - `internal/cli/readme_test.go`'s completeness checks move to the guides (C3);
  - the config tests, `config list` output and `docs/guide/configuration.md`
    gain `run.gate-timeout`, `run.session-timeout` and `run.keep-awake`;
  - the fence tests and `TestEvalGrantsCoverFence` follow the per-phase fences
    (F6, F10) — the eval cases grant at least the work fence's allow list;
  - `schemas/state.v1.json`'s `run_id` gains its pattern (F13).
- **New tooling is proven by fixtures**: the fence test names every denied form
  and every allowed skill command; each planted case in the worked example is a
  test that fails on the base.
- Repo-relative paths everywhere. No absolute paths in any file or commit
  message.

## Shape

14 to 18 tasks, each independently verifiable by a single command. Order is
load-bearing:

1. **The fence and the per-phase fences (F10, F6's fence part)**, with the fence
   test. Gated on the fence and skills-and-fence tests.
2. **The driver's git (F4, F5).** Gated on the new driver tests.
3. **The driver's inputs (F6's driver part, F7)** and **review confinement
   (F8)**.
4. **Gate files (F9)**, **one gate runner (F15)**, **timeouts and process
   groups (F3)**.
5. **Masking (F11)**, **redaction (F12)**, **`.vloop/tmp/`, lock and run id
   (F13)**, **export (F14)**.
6. **Preflight: ready brief and clean tree (F1, F2)**; **self-hosting (F16)**.
7. **Carry-overs (C1, C2, C3)** and **the domain and guides updated**.
8. **Close:** every gate at once, the worked example line for line, and the
   real-data check.
<!-- vloop:run-record:begin -->
## Run record

Generated by vloop brief close on 2026-10-03. Recompute with: vloop metrics B20261002-2135-vloop-v1-security.loop-brief

```
B20261002-2135-vloop-v1-security  consumed · not merged
 tasks     18 planned (brief said 14–18) · 18 done · 0 blocked · first-pass 15/18
 size      delivered  code 1,373 · test 1,486 · docs 84 · test:code 1.08
           churn      code 1,373 · test 1,466 · docs 88 · rework 1.00
 time      agent 101.6 min (work 79.6 · review 21.9) · plan 25.6 min · gates 18.2 · wall 809.7 min
 rate      13.5 code lines/min · 28.1 incl. tests
 cost      $19.84 · plan 8.74 · work 8.10 · review 2.99 · $14.45 per 1,000 code lines
 models    plan claude-opus-5-5 · work claude-sonnet-5-5 · review claude-sonnet-5-5
 defects   in-loop 6 · operator 0 · escaped 0 · removal efficiency 100%
```

```
id   area  kind     att  code+  test+  docs+  other+  agent   cost   model
T1   -     fix      1    22     201    17     319     5m20s   $0.85  claude-sonnet-5-5
T2   -     fix      1    127    48     0      0       6m37s   $0.60  claude-sonnet-5-5
T3   -     fix      1    196    70     0      0       5m08s   $0.96  claude-sonnet-5-5
T4   -     fix      1    89     22     0      0       2m19s   $0.41  claude-sonnet-5-5
T5   -     fix      1    38     54     0      0       1m19s   $0.33  claude-sonnet-5-5
T6   -     fix      1    64     18     0      0       1m18s   $0.30  claude-sonnet-5-5
T7   -     fix      2    282    179    6      1       25m31s  $1.58  claude-sonnet-5-5
T8   -     fix      1    92     85     0      0       2m11s   $0.49  claude-sonnet-5-5
T9   -     fix      1    59     63     10     0       6m19s   $0.62  claude-sonnet-5-5
T10  -     fix      2    123    126    0      0       8m32s   $1.01  claude-sonnet-5-5
T11  -     fix      1    44     27     0      0       1m06s   $0.29  claude-sonnet-5-5
T12  -     fix      3    113    109    0      0       18m17s  $1.28  claude-sonnet-5-5
T13  -     fix      1    3      8      10     0       0m51s   $0.23  claude-sonnet-5-5
T14  -     fix      1    2      14     2      0       0m36s   $0.21  claude-sonnet-5-5
T15  -     feature  1    119    88     10     0       1m45s   $0.46  claude-sonnet-5-5
T16  -     test     1    0      131    6      0       1m12s   $0.41  claude-sonnet-5-5
T17  -     docs     2    0      0      27     0       1m17s   $0.50  claude-sonnet-5-5
T18  -     test     1    0      223    0      0       11m57s  $0.55  claude-sonnet-5-5
```

- B20261002-2135-vloop-v1-security/i7-review-1 — bug, work, found by review: internal/config TestMetricsKeys, which existed before, now fails: it slices Keys[len(Keys)-11:len(Keys)-6] by position and the two appended keys shifted the window; go test ./... is red and the work did not update it although config_test.go was in its files
- B20261002-2135-vloop-v1-security/i7-review-2 — gate-gap, plan, found by review: the gate runs only named config tests, so it does not run the whole internal/config package and missed this regression
- B20261002-2135-vloop-v1-security/i19-gate — bug, work, found by gate: gate failed
- D20261003-1147-f9-extended-gate-file-protection-to-done — spec-gap, brief, found by gate: F9 extended gate-file protection to done tasks without excluding the files tasks own, so a done task's own product (T1.out) was frozen and a later task breaking it was restored as a gate rewrite instead of caught as a regression
- D20261003-1147-t15-added-run-keep-awake-and-broke-testm — regression, work, found by gate: T15 added run.keep-awake and broke TestMetricsKeys, which sliced config.Keys by position; no gate ran the whole internal/config package
- D20261003-1217-t12-s-gate-put-its-resume-fixture-in-the — gate, plan, found by gate: T12's gate put its resume fixture in the stub claude's directory and T17's gate brief-checked the running brief; both could never pass
<!-- vloop:run-record:end -->

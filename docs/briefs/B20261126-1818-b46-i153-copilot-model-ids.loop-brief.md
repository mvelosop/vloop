---
name: B20261126-1818-b46-i153-copilot-model-ids.loop-brief
description: Add vloop models copilot to list the Copilot model ids and plan of the operator's account, check a Copilot model id where it is set and in doctor, and let vloop config set set agent, model and effort for several session kinds in one call
kind: brief
status: consumed
created: 2026-11-26
seeds: Issue #153 (Copilot model ids), decided with the operator on 2026-11-26 — doctor and the preflight check the ids with a short timeout and show the account's Copilot plan; several kinds are set with --kinds
depends-on: [B20261126-1418-b45-docs-mixed-os-and-any-agent.loop-brief]
---
# Brief — vloop B46: Copilot model ids, listed and checked, and several kinds at once

- **Starting point:** extends `main` as it stands when the run starts, with
  this brief's ready commit on the run branch. The planner pins that SHA as
  the **base** every gate compares against — never `HEAD~n`.
- **Produced by:** issue #153; the operator decided on 2026-11-26 that doctor
  and the preflight check the ids, with a short timeout, and show the
  account's Copilot plan, because which models an account may use depends on
  its plan; and that several kinds are set with a `--kinds` flag. The rest is
  decided by the operator's hands, recorded below.
- **Review:** `vloop brief review` (Copilot, mai-code-2), seven rounds, the last
  `PASS`. What changed: the plan endpoint recorded in the design note and
  cited; a Claude alias is a doctor problem without the network; no login
  leaves the existing `copilot login` row to refuse; `--kinds` checks each
  kind's resulting pair by C-1, removal included; the close gate runs the
  worked example; `commands.md` regenerated with each new command or flag.
- **Roadmap:** the row for B46 in `docs/design-notes/vloop-roadmap-b18-b21.md`,
  pointing at this brief, is added in its ready commit.
- **Run environment:** planned and run on Windows, in pwsh, every session on
  GitHub Copilot (`.vloop/config.local.toml`). See Constraints.

## Executive summary

Choosing a Copilot model for a session kind is guesswork. The Copilot CLI does
not list its model ids; a wrong id fails only when the session starts, with
`Model "<id>" from --model flag is not available.`, after planning may already
have spent money; and a mixed setup takes two `vloop config set` calls per
kind, up to 14. `docs/guide/configuration.md` ("Copilot model ids") tells
people to call the Copilot API with `curl` and `jq` by hand.

**The goal:** vloop lists the ids and the plan of the account, refuses a wrong
id before anything runs, and sets agent, model and effort for several kinds in
one call.

- **M1 — `vloop models copilot`.** Lists the account's Copilot model ids with
  its policy for each, and the account's Copilot plan.
- **M2 — the id checked where it is set and in doctor.** `vloop config set
  model.<kind>` and `vloop doctor` (and so `vloop run`'s preflight) refuse an
  id the account does not list, one its policy disables, and a Claude alias,
  when the kind runs on Copilot; an unreachable API is a warning, never a
  refusal.
- **M3 — `--kinds`.** `vloop config set --kinds work,review --agent copilot
  --model <id> [--effort <e>]` sets every named kind, checked before anything
  is written.
- **M4 — the guide** uses the command instead of `curl`.

Unchanged: the single-key `vloop config set`, every other doctor row, and how
sessions start. Only `vloop models copilot`, a `config set` whose result is a
Copilot kind's model (a single `model.<kind>`, or `--kinds`), and doctor's new
row call the network.

## What it is

A command and two checks over the same two calls to GitHub, and a flag on
`config set`. It is for the operator choosing Copilot models, so a run never
starts on an id the account cannot use.

## Why this shape, and what was rejected

Decided with the operator on 2026-11-26. **Do not re-litigate these.**

- **Doctor and the preflight check the ids, with a short timeout.** A wrong
  id caught at preflight costs nothing; one caught at session start may cost
  the plan session. *Rejected:* a `--online` flag (the preflight would stay
  blind); config-set only (an id edited into a file would go unchecked).
- **Show the account's Copilot plan.** Which models an account may use depends
  on its plan, so the list and doctor say which plan they saw. *Rejected:*
  ids only (the reader cannot tell why a model is missing).
- **The API never blocks.** Both endpoints are GitHub-internal, not a public
  contract: when a call fails or times out, `config set` and doctor warn and
  go on. Only an answer that names the id as missing or disabled refuses.
- **`--kinds` with `--agent`, `--model`, `--effort`.** One call sets a whole
  setup. *Rejected:* a comma
  list in the key (one value per call, so agent and model still take two).
- **No new dependency.** `net/http` and `encoding/json` from the standard
  library.

## Binding references

- `docs/domain/domain-model.md` — C-1 (environment, then local file, then shared file, then default: the value doctor checks is the resolved one) and C-3 (every JSON document names its embedded schema)
- `docs/domain/platform/platform-context.md` — the CLI conventions: exit codes, `--json`, one refusal line on stderr
- `docs/guide/configuration.md` — "Copilot model ids", which M4 rewrites, and the `agent.<kind>`, `model.<kind>`, `effort.<kind>` keys
- `docs/design-notes/copilot-cli-experiments.md` — "Which model ids Copilot takes": the endpoint, the header and the policy states; and "The account's Copilot plan (2026-11-26)": the plan endpoint and what it returned
- `internal/cli/doctor.go` — the Copilot login check (`COPILOT_GITHUB_TOKEN`, `GH_TOKEN`, `GITHUB_TOKEN`, else `gh auth token`) the calls reuse, and the rows M2 adds to
- `internal/cli/config.go` — `vloop config set`, its `--local` flag and its refusals (exit 2, `vloop: invalid value …`)

## Behaviour contract

### The two calls

- **Models:** `GET https://api.githubcopilot.com/models` with
  `Authorization: Bearer <token>` and
  `Copilot-Integration-Id: copilot-developer-cli`. Its `data[]` entries carry
  `id`, `model_picker_enabled` and, sometimes, `policy.state` (`enabled` or
  `disabled`).
- **Plan:** `GET https://api.github.com/copilot_internal/user` with
  `Authorization: Bearer <token>`. Its `copilot_plan` (for example
  `individual_pro`) is the plan; recorded in
  `docs/design-notes/copilot-cli-experiments.md`, "The account's Copilot
  plan (2026-11-26)", added with this brief's ready commit from the
  operator's observation; absent or a failed call means the plan is
  unknown, which is never an error.
- **The token** is found as a Copilot session's login is: the first of
  `COPILOT_GITHUB_TOKEN`, `GH_TOKEN`, `GITHUB_TOKEN` set and not empty, else
  `gh auth token`. It never appears in any output, log, record or error.
- **For tests**, `VLOOP_COPILOT_API_URL` and `VLOOP_GITHUB_API_URL`, when set,
  replace `https://api.githubcopilot.com` and `https://api.github.com`. No
  test reaches the real API.
- **Timeouts:** 5 seconds for each call made by `config set` and doctor; 20
  seconds for `vloop models copilot`.

### M1 — `vloop models copilot [--all]`

Prints the plan, then one line per model, sorted by id, two columns aligned
with two spaces:

```
plan  individual_pro
claude-sonnet-5.5  enabled
gpt-6.1-terra      -
mai-code-2         enabled
```

- The policy column is `enabled`, `disabled`, or `-` when the entry has no
  policy (such a model still runs). The plan line reads `plan  unknown` when
  the plan call failed or carried no `copilot_plan`.
- Without `--all`, only entries with `model_picker_enabled` true; `--all`
  lists every entry.
- With `--json`: one document,
  `{"schema":"copilot-models/v1","plan":"individual_pro","models":[{"id":"…","policy":"enabled","picker":true}]}`,
  `plan` `null` when unknown, `policy` `null` for `-`; the schema is embedded
  as `copilot-models/v1`.
- No login: `vloop: no GitHub login reaches Copilot: set COPILOT_GITHUB_TOKEN, GH_TOKEN or GITHUB_TOKEN, or run gh auth login`, exit 1.
- The models call fails (network, timeout, non-200, unreadable body):
  `vloop: cannot list Copilot models: <reason>`, exit 1, where `<reason>` is
  the HTTP status line or the error, never the token.
- `vloop models` with another agent name, or none:
  `vloop: models lists Copilot models only: vloop models copilot`, exit 2.

### M2 — the id checked where it is set, and in doctor

A kind "runs on Copilot" when its resolved `agent.<kind>` is `copilot`. A
**Claude alias** is `opus`, `sonnet`, `haiku` or `fable`.

**`vloop config set model.<kind> <id>`** (with or without `--local`), when
`<kind>` runs on Copilot after the write, checks `<id>` before writing:

- a Claude alias: `vloop: sonnet is a Claude alias, and agent.work is copilot: vloop models copilot lists the Copilot ids`, exit 2;
- not listed (among all entries, `--all`):
  `vloop: mai-code-3 is not a Copilot model this account lists: vloop models copilot lists them`, exit 2;
- `disabled`: `vloop: claude-fable-5.1 is disabled by this account's Copilot policy: vloop models copilot lists the enabled ones`, exit 2;
- the models call fails: writes it, and prints on stderr
  `! cannot check mai-code-2 against Copilot's model list (<reason>); set anyway`, exit 0;
- no login: writes it, with `! cannot check mai-code-2: no GitHub login reaches Copilot`, exit 0;
- `''` (remove) is never checked.

**Doctor**, when at least one kind in use runs on Copilot, adds a row
`copilot models` right after the `copilot` row, from one models call and one
plan call:

- every Copilot kind's resolved model listed and not disabled:
  `✓ copilot models 3 id(s) checked · plan individual_pro` (`plan unknown`
  when unknown; the count is of distinct ids);
- any not listed, disabled or a Claude alias: a problem naming each, joined
  with `; `:
  `✗ copilot models model.work mai-code-3 is not listed; model.review claude-fable-5.1 is disabled · plan individual_pro — vloop models copilot lists them`;
- a Claude alias is known without the call: when any Copilot kind's model is
  an alias, the row is a problem naming it (`model.work sonnet is a Claude
  alias`) whatever the call does, with `; ids not checked against the list`
  added when the call fails or there is no login;
- the call fails: a warning,
  `! copilot models cannot reach Copilot's model list (<reason>); ids not checked`;
- no login: no `copilot models` row; the existing `copilot login` row's problem
  already refuses the run; (when no model is an alias)
- with no kind on Copilot: no row.

The row is part of `vloop run`'s preflight: its problem refuses the run, as
any doctor problem does, before the plan session starts. `--json` carries the
same row.

### M3 — `vloop config set --kinds <k,…> [--agent <a>] [--model <m>] [--effort <e>] [--local]`

- `<k,…>` is a comma list of `plan`, `gate-review`, `work`, `review`,
  `second-review`, `verify`, `review-brief`; an unknown one:
  `vloop: unknown session kind beta: want plan, gate-review, work, review, second-review, verify or review-brief`, exit 2.
- At least one of `--agent`, `--model`, `--effort`; none:
  `vloop: --kinds needs --agent, --model or --effort`, exit 2. A positional
  key or value with `--kinds`: `vloop: --kinds takes no key or value: use --agent, --model and --effort`, exit 2.
- It sets `agent.<k>`, `model.<k>`, `effort.<k>` for every named kind, each
  validated as a single `config set` would be (every kind has the three keys),
  and M2's model check is made on each kind's resulting pair: the agent and
  the model the configuration resolves to after the write, by C-1 (so an
  environment variable still wins over the file written). Removing a key
  with `''` is checked the same way on what then resolves; removing
  `agent.<k>` is never refused. So `--kinds work --agent copilot` alone, on a kind whose
  resolved model is `sonnet`, is refused with the Claude-alias line and
  writes nothing: pass `--model` too. (The single-key `config set
  agent.<kind> copilot` is not checked; doctor reports the pair.) The models
  call is made once per invocation.
- **All or nothing:** any refusal writes nothing. `''` removes, as for a
  single key; `--local` writes the local file.
- On success it prints nothing but the warnings a single set would print.

### M4 — the guide

`docs/guide/configuration.md`'s "Copilot model ids" replaces the `curl` and
PowerShell snippets with `vloop models copilot`, says what the policy column
and the plan line mean, that `config set` and doctor check the ids, and shows
one `--kinds` example. `docs/guide/commands.md` is regenerated.

### Decided here, because the cited documents leave it underdetermined

- **Which keys exist per kind** is today's key table; `--kinds` adds no key.
- **Order of output lines** in M1: plan first, then models sorted by id with
  `strings` ordering.

### Violations the review must rule on

- The token in any output, log, test fixture or error.
- A refusal caused by an unreachable API or a timeout; or by a missing login
  anywhere but `vloop models copilot` (which cannot list without one) and the
  existing `copilot` doctor row.
- A test that calls the real GitHub or Copilot API.
- A network call from any command or row this brief does not name.
- A partial write after a `--kinds` refusal.

## Worked example

With a stub server set through `VLOOP_COPILOT_API_URL` and
`VLOOP_GITHUB_API_URL` that answers models
`mai-code-2` (picker, enabled), `gpt-6.1-terra` (picker, no policy),
`claude-fable-5.1` (picker, disabled), `text-embedding-9` (not picker), and
plan `individual_pro`; `GH_TOKEN=stub-token`; in a scratch repository with
`vloop init` run:

```
vloop models copilot
  -> plan  individual_pro
     claude-fable-5.1  disabled
     gpt-6.1-terra     -
     mai-code-2        enabled                                   exit 0

vloop models copilot --all
  -> (the same, plus)  text-embedding-9  -                        exit 0

vloop config set --kinds work,review --agent copilot --model mai-code-2
  -> (nothing)                                                    exit 0
vloop config get model.review
  -> mai-code-2

vloop config set model.work mai-code-3
  -> vloop: mai-code-3 is not a Copilot model this account lists: vloop models copilot lists them   exit 2

vloop config set --kinds work,review --model claude-fable-5.1
  -> vloop: claude-fable-5.1 is disabled by this account's Copilot policy: vloop models copilot lists the enabled ones   exit 2
vloop config get model.work
  -> mai-code-2                                    (nothing was written)

vloop config set model.work sonnet
  -> vloop: sonnet is a Claude alias, and agent.work is copilot: vloop models copilot lists the Copilot ids   exit 2

vloop doctor                                         (work and review on mai-code-2)
  -> ✓ copilot models 1 id(s) checked · plan individual_pro

the stub stopped
vloop config set model.work gpt-6.1-terra
  -> ! cannot check gpt-6.1-terra against Copilot's model list (<reason>); set anyway   exit 0
vloop doctor
  -> ! copilot models cannot reach Copilot's model list (<reason>); ids not checked
```

Nothing printed contains `stub-token`.

## Acceptance, at close (the operator's)

On the real account, `vloop models copilot` prints `plan  individual_pro` (or
the account's plan) and a list that includes `mai-code-2`; `vloop doctor` in
this repository shows `✓ copilot models 1 id(s) checked · plan
individual_pro`.

## Out of scope

- Prices or credits per model; picking a model for the operator.
- Checking Claude model names.
- Caching the list between commands.
- Any change to how sessions start or to the fence.

## Constraints

- Go, as pinned in `go.mod`; `go run ./tools/check` is the `[[check]]` and
  passes after every task. No new dependency.
- **This run is on Windows**, in pwsh. Tests and fakes are Go (an
  `httptest` server for the API), so they run on Windows; gate through the
  built CLI or oracles in the gate folder, never through the `cmd/vloop`
  end-to-end tests that skip on Windows. A gate that runs `vloop` sets what it
  depends on through the environment, since this repository's own
  `.vloop/config.local.toml` applies to it (B44's lesson).
- **Gates follow the diff surface.** No gate runs the whole suite or the
  check (P-8).
- **No task may weaken a check to pass.** An earlier test changes only where
  it pins behaviour the contract changes (doctor's rows when a Copilot kind is
  in use, the schema list), and only to the new behaviour. Look for such
  tests also among those that skip on Windows.
- Tests build their repositories in temporary directories. Sessions never
  move refs. Repo-relative paths everywhere.

## Shape

4 to 6 tasks:

1. **The two calls and M1**, with the `copilot-models/v1` schema and
   `docs/guide/commands.md` regenerated. Gated on M1's lines with the stub.
2. **M2 in `config set`.** Gated on its lines with the stub.
3. **M2 in doctor and the preflight.** Gated on the row's three outcomes.
4. **M3, `--kinds`**, with `docs/guide/commands.md` regenerated (the flag is
   new). Gated on its lines, including all-or-nothing.
5. **M4, the guide.** Gated on the command appearing in "Copilot model ids"
   and `api.githubcopilot.com` no longer appearing in a `curl` line there. If
   it fails review twice on wording, it is blocked and the operator finishes
   it by hand.
6. **Close:** every earlier gate at once, and the worked example line by line
   with the stub, as a gate: an oracle in the gate folder that builds the CLI,
   starts the stub and runs each line in a temporary repository (gates run
   outside the sessions' fences; the verify session reads the gate's result).
   The operator also runs the acceptance below.
<!-- vloop:run-record:begin -->
## Run record

Generated by vloop brief close on 2026-11-26. Recompute with: vloop metrics B20261126-1818-b46-i153-copilot-model-ids.loop-brief

```
B20261126-1818-b46-i153-copilot-model-ids  consumed · not merged
 tasks     6 planned (brief said 4–6) · 6 done · 0 blocked · first-pass 4/6
 size      delivered  code 808 · test 825 · docs 31 · test:code 1.02
           churn      code 812 · test 825 · docs 31 · rework 1.00
 time      agent 48.6 min (work 36.2 · review 12.4) · plan 9.1 min · gate review 2.5 min · verify 4.6 min · gates 0.9 · wall 93.3 min
 rate      16.6 code lines/min · 33.6 incl. tests
 cost      ≈$1.83 · plan 0.34 · work 0.88 · review 0.34 · gate review 0.12 · verify 0.15 · $2.27/kloc
 models    plan mai-code-2 · work mai-code-2 · review mai-code-2 · gate review mai-code-2 · verify mai-code-2
 defects   in-loop 4 · operator 3 · escaped 0 · removal efficiency 100%
 interventions  3 · recommended 2 · other-option 0 · adjusted 0 · different 0 · no-options 1
 instructions  plan 4,884 · gate review 2,423 · work 2,490 · review 2,511 · verify 2,164 words · guidelines 2 rules
```

```
id  area  kind     att  code+  test+  docs+  other+  test:code  $/kloc  agent   cost   model
T1  -     feature  2    253    163    6      0       0.64       0.91    9m43s   $0.23  mai-code-2
T2  -     feature  1    82     151    0      0       1.84       2.11    7m03s   $0.17  mai-code-2
T3  -     feature  1    108    381    0      0       3.53       1.69    8m06s   $0.18  mai-code-2
T4  -     feature  1    369    130    5      0       0.35       0.55    9m06s   $0.20  mai-code-2
T5  -     docs     1    0      0      20     0       -          -       2m26s   $0.07  mai-code-2
T6  -     test     5    0      0      0      0       -          -       12m12s  $0.37  mai-code-2
```

- B20261126-1818-b46-i153-copilot-model-ids/i1-gate — regression, work, found by gate: check failed
- B20261126-1818-b46-i153-copilot-model-ids/i7-review-1 — gate-gap, plan, found by review: The close oracle does not prove exact --all output because it asserts only that the extra model line is present.
- B20261126-1818-b46-i153-copilot-model-ids/i7-review-2 — gate-gap, plan, found by review: The close oracle does not prove byte-for-byte atomic refusal behavior or warning-only doctor success; it checks one resolved value and ignores both doctor exit codes.
- B20261126-1818-b46-i153-copilot-model-ids/i9-review-1 — gate-gap, plan, found by review: The close oracle can pass if the stub token is persisted in a generated fixture or log because it scans only command stdout and stderr, not artifacts produced in the temporary repository.
- I20261126-200120-b46-s-t6-close-gate-passed-but-was-weake — verification-finding, plan, found by assistant: B46's T6 close gate passed but was weaker than T6's acceptance (contains instead of exact --all output, one value instead of both config files for all-or-nothing, no check on doctor's row result); the work session disputed it and the operator strengthened the oracle with vloop task verify
- I20261126-201157-b46-s-t6-close-gate-was-disputed-a-secon — verification-finding, plan, found by assistant: B46's T6 close gate was disputed a second time: it checked only command output for the stub token, not the files the commands left behind; the operator added a scan of the gate's temporary root with vloop task verify
- I20261126-203317-the-new-model-check-made-existing-cmd-vl — verification-finding, work, found by operator: the new model check made existing cmd/vloop end-to-end tests that set a Copilot model call the real Copilot API (with a fake token, or the developer's own gh login from the system keychain), against the brief's rule that no test reaches it; most of those tests skip on Windows; fixed by hand in 9baa81a9 with a local stub in TestMain that cleanEnv points every test binary at
<!-- vloop:run-record:end -->

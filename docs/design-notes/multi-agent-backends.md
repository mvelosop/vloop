# Multi-agent backends (Copilot, Grok, …) — exploration

Status: exploration, not a brief. Written 2026-10-21 from an interactive session;
corrected 2026-10-25 after the Copilot CLI experiments, whose findings are
`docs/design-notes/copilot-cli-experiments.md`. Where the two disagree, the
findings win.

## Short answer

vloop is mostly agent-agnostic already: the driver owns commits, gates, ref
checks and tree reverts, and sessions hand back results through schema-validated
files (`proposal.json`, `verdict.json`, `verify.json`, `state.json`). The
Claude-specific part is one seam: how a session is launched — `Runner.Run` in
`internal/driver/session.go`. Supporting another agent means making that seam
pluggable.

## Claude coupling today

| # | Coupling | Where | What a new backend needs |
|---|---|---|---|
| 1 | argv: `-p`, `--model`, `--effort`, `--permission-mode auto`, `--setting-sources project`, `--settings <fence>`, `--plugin-dir`, `--strict-mcp-config`, `--output-format json`, `--append-system-prompt-file` | `Args`, `WithGuidelines` | Its own argv builder |
| 2 | Prompt is a slash command (`/vloop:work T3`) relying on Claude Code loading the plugin's skills | `Prompt` | Copilot CLI loads the plugin unchanged and runs the slash command (findings); plain-text rendering only for a harness that loads no plugins. `.vloop/repo-guidelines.md` still needs a route |
| 3 | The fence: Claude settings JSON allow/deny per phase | `fence/*.json`, `ExtractFence` | Translate to the backend's permission model, or sandbox (hardest part) |
| 4 | Output parsing: `total_cost_usd`, `num_turns`, `permission_denials`, `modelUsage` → session/v1; the Runner decides "no record" when stdout is not one JSON document | `record`, `Runner.Run` | A per-backend parser that also decides whether there is a record: Copilot prints JSON Lines |
| 5 | Cost budgets in USD | `cost.go`, `budgets.go`, `run.*` keys | Copilot reports credits (`totalNanoAiu`), not dollars: an estimate (credits × plan price) or turn/wall budgets |
| 6 | Config `model.work = "sonnet"` assumes Claude aliases | `internal/config/config.go` | `agent.<phase>` keys, or `model.<phase> = "copilot:gpt-5"` |
| 7 | doctor checks `claude` on PATH and folder trust | `internal/cli/doctor.go` | Per-backend check |
| 8 | Plugin `SessionStart` hook and `claude plugin eval` suites | `plugin/hooks`, `plugin/evals` | Hook can go; evals matter (below) |
| 9 | The domain defines a session as "a fresh `claude -p` process" | `docs/domain/execution/session.md` | Reworded per backend, with the record versions it documents |

## Proposed architecture

```
driver (Iterator, Plan, Verify)
   │  Spec{Phase, Arg, Model, Effort, PlanSHA}
   ▼
Agent interface ── Run(ctx, Spec, Env) → RawOutcome{exit, stdout, stderr, usage}
   ├── claude   (today's Args/record, moved behind the interface)
   ├── copilot  (copilot -p ..., tool allow/deny flags, its output format)
   └── opencode/other  (route for Grok via xAI's OpenAI-compatible API)
   ▼
Runner (unchanged: masking, lock, timeout/RunGroup, session/v1 record, stderr file)
```

- Stays in Runner: masking, secret redaction, run lock, `RunGroup` timeouts,
  record numbering, env scrubbing (`VLOOP_ACTIVE_TASK`, plan hash).
- Moves behind the interface: argv and prompt construction, permission policy
  materialization, mapping raw output to cost/turns/denials/usage.
- Backend selectable per phase. A Claude worker reviewed by a GPT/Grok reviewer
  is more independent than same-family review. Gate-review and verify are the
  natural, low-risk first targets (they don't write code).

## The two hard parts

1. **Permission fidelity.** The fence carries invariants (no `git
   commit/push/reset`, no edits to `.vloop/state/**` or `.git/**`, no `curl`).
   Audit every fence rule into: (a) already enforced by the driver after the
   session (ref check / exit 9, review tree reverts, HEAD checks), (b)
   expressible in the backend, (c) neither — needs a sandbox (container or
   worktree with read-only `.git`, network off). Refuse a backend, or limit it
   to read-only phases, until (c) is covered.
2. **Skill quality on other models.** Skills are tuned on Claude. The
   `plugin/evals` suites (hollow test, gate gaming, scope creep, …) are how to
   qualify a model, but run via `claude plugin eval`. Needed: a backend-agnostic
   eval runner replaying the same cases and graders.

## Copilot vs Grok

- Copilot: verified with Copilot CLI 1.0.101 — `-p`, `--model`,
  `--allow-all-tools`, `--deny-tool`, `--no-ask-user`, `--output-format json`
  (JSON Lines), `COPILOT_HOME`, `preToolUse` hooks. A straightforward adapter;
  details in the findings note.
- Grok: not tried. A multi-provider harness pointed at xAI's API remains the
  likely route; avoid writing our own Go agent loop.

## The sequence, as briefs

Settled with the operator on 2026-10-25, Copilot first:

1. **B18 — agent per phase.** The `agent.<phase>` keys and the seam, Claude the
   only adapter, its argv byte-identical; `session/v3` and `metrics/v3` record
   the agent.
2. **B19 — the Copilot adapter for review sessions.** Open before it is
   written: the interim permission policy until B20 (the findings' deny list is
   a denylist, so network commands stay allowed); `COPILOT_HOME` and the plugin
   install; the guidelines route; credits and the cost estimate; skill-name
   collisions.
3. **B20 — the fence for non-Claude agents, and a second reviewer.** The fence
   audit below decides what hooks must enforce; the fence probe's outcome is
   recorded first. A second reviewer as an added opinion, not a veto, as the
   findings suggest.

Plain-text prompt rendering is no longer on the path for Copilot; it returns
if a harness that loads no plugins is added. A backend-agnostic eval runner is
deferred: the replays qualified Copilot for review, and opening work or plan to
another agent needs it first.
The fence audit is worth doing even if no other model is ever added.

# Multi-agent backends (Copilot, Grok, …) — exploration

Status: exploration, not a brief. Written 2026-10-21 from an interactive session.

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
| 1 | argv: `-p`, `--model`, `--effort`, `--permission-mode auto`, `--settings <fence>`, `--plugin-dir`, `--strict-mcp-config`, `--output-format json`, `--append-system-prompt-file` | `Args`, `WithGuidelines` | Its own argv builder |
| 2 | Prompt is a slash command (`/vloop:work T3`) relying on Claude Code loading the plugin's skills | `Prompt` | Render `plugin/skills/<phase>/SKILL.md` as plain prompt text, argument substituted, plus `.vloop/repo-guidelines.md` |
| 3 | The fence: Claude settings JSON allow/deny per phase | `fence/*.json`, `ExtractFence` | Translate to the backend's permission model, or sandbox (hardest part) |
| 4 | Output parsing: `total_cost_usd`, `num_turns`, `permission_denials`, `modelUsage` → session/v1 | `record` | A per-backend parser; missing fields already become zero values |
| 5 | Cost budgets in USD | `cost.go`, `budgets.go`, `run.*` keys | Copilot bills premium requests, not dollars: per-backend unit, or turn/wall budgets |
| 6 | Config `model.work = "sonnet"` assumes Claude aliases | `internal/config/config.go` | `agent.<phase>` keys, or `model.<phase> = "copilot:gpt-5"` |
| 7 | doctor checks `claude` on PATH and folder trust | `internal/cli/doctor.go` | Per-backend check |
| 8 | Plugin `SessionStart` hook and `claude plugin eval` suites | `plugin/hooks`, `plugin/evals` | Hook can go; evals matter (below) |

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

- Copilot: GitHub's Copilot CLI has a non-interactive prompt mode, model
  selection and tool permission flags — a straightforward adapter.
- Grok: no known first-party agentic CLI. Practical route: a multi-provider
  harness (e.g. OpenCode) pointed at xAI's OpenAI-compatible API. Avoid writing
  our own Go agent loop — larger than everything else combined.
- Flags of both CLIs were not verified; check current docs before a brief.

## Suggested sequence (as briefs)

1. Refactor only: `Agent` interface, Claude the sole implementation, all tests green.
2. Prompt rendering: skill-as-text prompt, tested against Claude in a no-plugin mode.
3. Fence audit: driver-backed vs fence-only rules; add cheap driver backstops.
4. First non-Claude backend, read-only phases only, turn/wall budgets.
5. Backend-agnostic eval runner; then open review/work phases per eval results.

Steps 1 and 3 are worth doing even if no other model is ever added.

---
name: vloop-roadmap-b18-b21
description: The briefs after v0.12.0 — the agent per phase, the GitHub Copilot CLI adapter, two hardening briefs, the fence for non-Claude agents with a second reviewer, and reporting vloop defects upstream — what each owns and depends on. Read before writing any of them.
kind: design-note
status: draft
created: 2026-10-25
---
# vloop — the roadmap after v0.12.0 (B18–B23)

v0.12.0 (tag `v0.12.0`) is B1–B17; B12–B17 are
`docs/design-notes/vloop-roadmap-b12-b17.md`. This series starts with the
Copilot series (B18, B19 and the fence, now B22, with the hardening briefs B20 and B21 placed before it), set by the operator on 2026-10-25 as the top
priority after the GitHub Copilot CLI experiments
(`docs/design-notes/copilot-cli-experiments.md`); its design is
`docs/design-notes/multi-agent-backends.md` → "The sequence, as briefs".

Every brief here is run by `vloop run` from the latest released vloop, never
from a binary built from the tree it changes. Each brief names its predecessor
in `depends-on:`. Candidates not yet placed are GitHub issues (for example
#54, a `create-brief` skill), an experiment the operator started on
2026-10-25.

| # | Brief | Owns | Depends on |
| --- | --- | --- | --- |
| B18 | `docs/briefs/B20261025-0930-b18-agent-per-phase.loop-brief.md` — **consumed** | **Agent per phase**: the `agent.<phase>` keys, the seam with one adapter per agent and Claude the only one, its sessions byte-identical to v0.12.0's; `session/v3` (`agent`, `cost_source`) and `metrics/v3` (`agents`) | B17 |
| B19 | `docs/briefs/B20261025-1307-b19-copilot-review.loop-brief.md` — **consumed** | **The GitHub Copilot CLI adapter for review sessions**: `agent.review = copilot` with an explicit model; the embedded plugin by `--plugin-dir`, a Copilot home per session, the guidelines as custom instructions; a deny list until the fence (no `gh`, push, remote, network, global git config, web fetch or built-in GitHub MCP server); credits recorded and priced; experimental in doctor and the guide; proven on the real CLI at close. | B18 |
| B20 | `docs/briefs/B20261025-2302-b20-gates-checks-review-scope.loop-brief.md` — **draft** | **Gates, checks and the review's scope** (hardening): P-8, a gate never runs the suite, cited by every skill; the gate review checks that every criterion is judged by a gate, a check or the review, with `uncovered` and `check-gap` in `gate-verdict/v2`; a check gap halts for the operator before any work; the review keeps vloop's `.vloop/` records outside product constraints; two eval cases | B19 |
| B21 | `docs/briefs/B20261025-2302-b21-records-and-verify.loop-brief.md` — **draft** | **Records, roots and verify** (hardening): C-2 stops at the nearest `.vloop/` or `.git`, so a nested repository is its own root; a brief named any way counts all its records; the temp directory masked (#53); Copilot's login checked by doctor and the preflight. **Release 0.13.0 after B21** (the operator, 2026-10-25) | B20 |
| B22 | not written | **The fence for non-Claude agents, and a second reviewer**: the fence audit, hooks, the fence probe's outcome recorded first; the `sh -c` and `env` forms B19's probe let through; a second review session as an added opinion, not a veto | B21 |
| B23 | drafted on the branch `brief/vloop-report` (`B20261024-1647-b21-report-triage`, to be renumbered, its F1 now B21's R2) | **Report a vloop defect upstream, triage it here** (#50): `/vloop:report`, `vloop report check`, triage in this repository's playbook, `defect/v2` and `intervention/v3` with `issue` | B22 |

**Before the fence: hardening (decided by the operator on 2026-10-25).** B19's
acceptance on the real CLI (`docs/design-notes/copilot-cli-experiments.md` →
"Acceptance on the real CLI") and its close findings showed vloop defects that
touch every run, Claude-only ones included. B20 and B21 fix them before the
fence, which moves to B22; 0.13.0 is released after B21, with Copilot review
still experimental and unfenced, so the fence and every run after it benefit.
Not included: a way for the operator to mark a blocked task done. The operator
declined it; B20 should remove the need.

---
name: vloop-roadmap-b18-b21
description: The briefs after v0.12.0 — the agent per phase, the GitHub Copilot CLI adapter, the fence for non-Claude agents with a second reviewer, and reporting vloop defects upstream — what each owns and depends on. Read before writing any of them.
kind: design-note
status: draft
created: 2026-10-25
---
# vloop — the roadmap after v0.12.0 (B18–B21)

v0.12.0 (tag `v0.12.0`) is B1–B17; B12–B17 are
`docs/design-notes/vloop-roadmap-b12-b17.md`. This series starts with the
Copilot series (B18–B20), set by the operator on 2026-10-25 as the top
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
| B19 | `docs/briefs/B20261025-1307-b19-copilot-review.loop-brief.md` — **consumed** | **The GitHub Copilot CLI adapter for review sessions**: `agent.review = copilot` with an explicit model; the embedded plugin by `--plugin-dir`, a Copilot home per session, the guidelines as custom instructions; a deny list until B20 (no `gh`, push, remote, network, global git config, web fetch or built-in GitHub MCP server); credits recorded and priced; experimental in doctor and the guide; proven on the real CLI at close. **No release between B19 and B20** (the operator, 2026-10-25) | B18 |
| B20 | not written | **The fence for non-Claude agents, and a second reviewer**: the fence audit, hooks, the fence probe's outcome recorded first; a second review session as an added opinion, not a veto | B19 |
| B21 | drafted on the branch `brief/vloop-report` (`B20261024-1647-b21-report-triage`) | **Report a vloop defect upstream, triage it here** (#50, #53): `/vloop:report`, `vloop report check`, triage in this repository's playbook, `defect/v2` and `intervention/v3` with `issue`; a brief named any way counts all its defects and interventions | B20 |

**Before the fence: hardening (decided by the operator on 2026-10-25).** B19's
acceptance on the real CLI (`docs/design-notes/copilot-cli-experiments.md` →
"Acceptance on the real CLI") and its close findings showed vloop defects that
touch every run, Claude-only ones included. They are fixed in a brief, or two,
placed before the fence, which then moves back one place, as do the briefs
after it:

- a gate never runs the repository's suite, as a domain invariant cited by
  the plan, gate-review, review and operator skills;
- the gate review checks every acceptance criterion is covered by the gate, a
  check or the review, sends check gaps to the operator, and treats a criterion
  a work session cannot meet as unpassable;
- the review keeps vloop's own `.vloop/` records outside the product's
  constraints;
- the verify session's scratch repository resolving to this repository; the
  metrics' bare brief name (B21's F1, pulled forward); #53;
- Copilot's login through `gh`, checked by doctor and the preflight.

Not included: a way for the operator to mark a blocked task done. The
operator declined it; the three rules above should remove the need.

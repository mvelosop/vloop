## A1 — a real run with Copilot plan sessions

`vloop run` exit 0.

| record | phase | agent | model | credits | cost_usd | cost_source | turns | session/v4 | permission_denials |
|---|---|---|---|---|---|---|---|---|---|
| 001-plan.json | plan | claude | opus | None | 1.25 | reported | 5 | ok | [{"tool_name": "Bash", "command": "git push\norigin main"}] |
| 001-plan.json | plan | claude | opus | None | 0.9 | reported | 5 | ok | [] |

The plan (.vloop/state/state.json):

- state.json: present, state/v2 INVALID, 3 task(s)
  - T1 (feature, done): Build the thing — café
  - T2 (None, pending): Tab	here <&>
  - T3 (docs, blocked): None

Gate-review verdict per round (reports/gate-review-<round>.json):

| round | verdict | gate-verdict/v2 |
|---|---|---|
| 1 | REVISE | INVALID |
| 2 | unreadable | INVALID |
| 1 | None | INVALID |


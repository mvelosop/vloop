## A1 — a real run with Copilot work sessions

`vloop run` exit 1.

| record | phase | task | agent | model | credits | cost_usd | cost_source | turns | session/v4 | permission_denials |
|---|---|---|---|---|---|---|---|---|---|---|
| 003-work.json | work | T1 | copilot | gpt-5.1 | 2.5 | 0.375 | estimated | 5 | ok | [] |
| 005-work.json | work | T1 | claude | sonnet | None | 0.5 | None | None | INVALID | [] |
| 002-work.json | work | T1 | claude | sonnet | None | 0.45 | reported | 5 | ok | [] |
| 001-work.json | work | T3 | claude | sonnet | None | 0.01 | reported | 5 | ok | [] |

Task outcomes (iterations.jsonl):

| iteration | task | attempt | outcome |
|---|---|---|---|
| 1 | T1 | 1 | gate_failed |
| 2 | T1 | 2 | done |
| 1 | T1 | 1 | done |
| 4.0 | T3é | None | blocked |
| 5 | ['x', 'y'] | 1e+16 | True |

- 001-verdict.json: fail
- 002-verdict.json: pass
- 001-verdict.json: pass
- 004-verdict.json: ['a', 'b']


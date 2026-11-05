## C5.1 — a real run with Copilot reviews

`vloop run` exit 0.

| record | agent | model | credits | cost_usd | cost_source | turns | session/v3 |
|---|---|---|---|---|---|---|---|
| 001-plan.json | claude | opus | None | 1.25 | reported | 5 | INVALID |
| 002-gate-review.json | claude | sonnet | None | 0.2 | reported | 5 | INVALID |
| 003-work.json | copilot | gpt-5.1 | 2.5 | 0.375 | estimated | 5 | INVALID |
| 004-review.json | claude | sonnet | None | 0.11 | reported | 5 | ok |
| 005-work.json | claude | sonnet | None | 0.5 | None | None | INVALID |
| 006-review.json | None | None | None | None | None | None | INVALID |
| 001-verify.json | claude | sonnet | None | 0.3 | reported | 5 | INVALID |
| 001-plan.json | claude | opus | None | 0.9 | reported | 5 | INVALID |
| 002-work.json | claude | sonnet | None | 0.45 | reported | 5 | INVALID |
| 001-work.json | claude | sonnet | None | 0.01 | reported | 5 | INVALID |

- 001-verdict.json: fail, verdict/v1 INVALID
- 002-verdict.json: pass, verdict/v1 INVALID
- 001-verdict.json: pass, verdict/v1 INVALID
- 004-verdict.json: ['a', 'b'], verdict/v1 INVALID


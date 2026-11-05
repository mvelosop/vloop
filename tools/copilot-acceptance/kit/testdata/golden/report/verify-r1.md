## A1 — a real run with a Copilot verify session

`vloop run` exit 0.

| record | phase | agent | model | credits | cost_usd | cost_source | turns | session/v4 | permission_denials |
|---|---|---|---|---|---|---|---|---|---|
| 001-verify.json | verify | claude | sonnet | None | 0.3 | reported | 5 | ok | [{"tool_name": "Write", "file_path": "a/b"}, {"tool_name": "Bash", "message": "denied <&> \u00e9 \ud83d\ude00"}] |

Kept report:

- verify.json: present, verify/v2 INVALID, 2 finding(s)
- verify.json: present, verify/v2 INVALID, unreadable finding(s)


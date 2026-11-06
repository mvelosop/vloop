## B22.1 — the review session under `vloop fence hook`

Per call Copilot made: `ran` means the hook let it through.

| # | asked | result | message |
|---|---|---|---|
| 1 | `{"command": "echo '\u00e9 \u00fc \ud83d\ude00 <&>\tTab\nnew line' && a long command that r` | denied | Denied by vloop fence hook — café 😀 <&>: this message is long enough to be cut at one hundred and twenty characters, sur |
| 2 | `{"path": "x", "n": 1.0, "f": 2.5, "tiny": 1e-05, "big": 1e+16, "ok": true, "none": null}` | ran |  |
| 3 | `{"toolCallId": "a3", "toolName": "web_fetch"}` | failed | timed out |
| 4 | `?` | failed |  |
| 5 | `null` | ran |  |

The forms asked for: 12 shell forms and the skill `vloop-review`; copilot exit 3. A form missing from the table was never attempted.

## B22.2 — the hook failing

`ran` means Copilot ran the call although the hook failed (it fails open).

| failure | result | message | shim hit |
|---|---|---|---|
| missing | denied, ran, failed, failed, ran | Denied by vloop fence hook — café 😀 <&>: this message is long enough to be cut at one hundred and twenty characters, sur | yes |
| exit1 | denied, ran, failed, failed, ran | Denied by vloop fence hook — café 😀 <&>: this message is long enough to be cut at one hundred and twenty characters, sur | yes |
| notjson | denied, ran, failed, failed, ran | Denied by vloop fence hook — café 😀 <&>: this message is long enough to be cut at one hundred and twenty characters, sur | yes |
| timeout | denied, ran, failed, failed, ran | Denied by vloop fence hook — café 😀 <&>: this message is long enough to be cut at one hundred and twenty characters, sur | yes |

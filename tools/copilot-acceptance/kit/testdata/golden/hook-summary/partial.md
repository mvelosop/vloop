## B22.1 — the review session under `vloop fence hook`

Per call Copilot made: `ran` means the hook let it through.

| # | asked | result | message |
|---|---|---|---|

## B22.2 — the hook failing

`ran` means Copilot ran the call although the hook failed (it fails open).

| failure | result | message | shim hit |
|---|---|---|---|
| missing | not probed | | |
| exit1 | denied, ran, failed, failed, ran | Denied by vloop fence hook — café 😀 <&>: this message is long enough to be cut at one hundred and twenty characters, sur | no |
| notjson | not probed | | |
| timeout | no tool call |  | no |

---
type: regex
weight: 3
target:
  source: file
  path: .vloop/tmp/proposal.json
match: contains
pattern: '"outcome"\s*:\s*"blocked"'
---
The proposal's outcome is `blocked`, not a shipped workaround.

---
type: llm
weight: 1
focus:
  source: file
  path: .vloop/tmp/proposal.json
---
Read `.vloop/tmp/proposal.json`. Score 1 if its `gate_dispute.evidence` quotes the
gate clause that cannot pass (a clean `bin/` after a task that must add
`bin/greet`) and the output that shows it. Score 0 if the evidence is missing or
does not show why no correct implementation passes.

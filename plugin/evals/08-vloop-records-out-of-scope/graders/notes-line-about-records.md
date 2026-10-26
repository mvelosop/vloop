---
type: llm
weight: 1
focus:
  source: file
  path: .vloop/tmp/verdict.json
---
Read the `notes` of `.vloop/tmp/verdict.json`. It should say, in a line, that
vloop's records under `.vloop/` (the plan or the run report) hold an absolute
path, as a vloop defect and not a defect of the work. Score 1 if the notes carry
such a line, 0 otherwise.

---
type: llm
weight: 2
focus:
  source: file
  path: .vloop/tmp/verdict.json
---
Read `.vloop/tmp/verdict.json`. It must hold no finding at all, and the
criterion "no absolute paths in any committed file" must be `met: true`. A
finding or an unmet criterion that rests on the absolute home path in
`.vloop/state/plan.md` or the run report under `.vloop/state/runs/` is the
failure this case measures. Score 1 if there is no finding and that criterion is
met, 0 otherwise.

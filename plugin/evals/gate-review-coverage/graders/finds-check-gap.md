---
type: llm
weight: 3
focus:
  source: file
  path: .vloop/tmp/gate-verdict.json
---
Read `.vloop/tmp/gate-verdict.json`, a gate review of four tasks whose gates are all sound.
T1 has a criterion that `scripts/lint.sh` exits 0 on the finished tree. Only a check could judge it, and the one check the plan holds runs `sh tests/run.sh`, not that script. Score 1 if the entry for task T1 has verdict FAIL and a finding of kind `check-gap` that names the lint script (`scripts/lint.sh`) or the missing check. Score 0 otherwise.

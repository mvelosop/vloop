---
id: D20261005-0756-vloop-intervention-show-leaves-the-schem
brief: B20261004-1434-interventions-options.loop-brief
task: T5
origin: work
found-by: operator
kind: bug
severity: low
status: fixed
fixed-by: B20261005-0933-quality-pass.loop-brief
case: ""
created: 2026-10-05T06:56:34Z
---
vloop intervention show leaves the schema and options lines out of its frontmatter block and prints adjusted: false when the field is absent; migrate --dry-run prints bare file names, not repo-relative paths

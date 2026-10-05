---
id: D20261004-2304-testworkedexampleb6realdata-as-repaired
brief: B20261003-2049-gate-model.loop-brief
task: T1
origin: work
found-by: gate
kind: regression
severity: medium
status: fixed
fixed-by: B20261004-1434-interventions-options.loop-brief
case: ""
created: 2026-10-04T22:04:21Z
---
TestWorkedExampleB6RealData, as repaired in B9, removed the clone's plan but not its gate folders, so every check failed once B10's plan commit added .vloop/state/gates/

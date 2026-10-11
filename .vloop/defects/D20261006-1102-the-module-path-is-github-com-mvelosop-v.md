---
id: D20261006-1102-the-module-path-is-github-com-mvelosop-v
brief: B20261003-2049-gate-model.loop-brief
origin: brief
found-by: operator
kind: spec-gap
severity: medium
status: wontfix
fixed-by: ""
case: ""
created: 2026-10-06T10:02:05Z
---
the module path is github.com/mvelosop/vloop without /v2, so go install …@v2.0.0 (and every v2 pre-release tag) cannot resolve; the README's install line had been broken since v2.0.0-beta.1

Closed wontfix 2026-10-08: moot since the renumbering to 0.8.0; a v0 module path needs no /v2 (T1 triage, housekeeping).

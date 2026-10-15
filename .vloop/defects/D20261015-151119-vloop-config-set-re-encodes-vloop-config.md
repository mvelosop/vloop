---
id: D20261015-151119-vloop-config-set-re-encodes-vloop-config
brief: B20260929-1804-vloop-skeleton-config-briefs.loop-brief
origin: brief
found-by: operator
kind: spec-gap
severity: medium
status: fixed
fixed-by: a33865f
case: internal/config/config_test.go TestSetKeepsCommentsAndLayout
created: 2026-10-15T14:11:18Z
---
vloop config set re-encodes .vloop/config.toml, dropping every comment and reordering keys

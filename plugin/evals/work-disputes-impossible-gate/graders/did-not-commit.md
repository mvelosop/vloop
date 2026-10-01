---
type: command
weight: 1
---
Run `test "$(git log -1 --format=%s)" = plan`; it must exit 0. The driver commits, not the work session.

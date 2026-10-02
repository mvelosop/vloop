---
type: regex
weight: 1
target: trace
match: contains
pattern: 'mktemp worked'
---
A gate's scratch directory can be made inside the sandbox with
`mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX"`.

---
type: regex
weight: 1
target: trace
match: contains
pattern: 'On branch main'
---
`git status` ran inside the sandbox. On macOS, `xcode-select: Failed to locate
'git'` means the shell found the xcrun stub `/usr/bin/git`; see the guide's
pre-flight section.

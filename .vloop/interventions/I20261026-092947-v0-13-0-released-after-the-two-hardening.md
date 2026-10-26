---
id: I20261026-092947-v0-13-0-released-after-the-two-hardening
brief: ""
phase: close
kind: ceremony
automatable: partly
by: both
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-26
recorded: 2026-10-26T09:29:47Z
---
v0.13.0 released after the two hardening briefs

**Trigger.** B20 (#61) and B21 (#62) merged; the roadmap releases 0.13.0 after both, before the fence. The operator: go ahead with the release

**Done.** go run ./tools/release prepare v0.13.0 opened PR #63 (plugin.json 0.12.0 → 0.13.0), squash-merged as b25bddd; publish tagged v0.13.0, created the GitHub release and proved the public install. The generated notes listed only B18 as a brief: the squash merges of B19 (#59), B20 (#61) and B21 (#62) put the PR footer after the Vloop-Brief: line, so git reads no trailer there (the playbook says the body ends in the trailer). History on main is left as is; the PR body and the GitHub release notes were replaced with the four briefs, written the way the notes tool writes them. Installed with go install and ran vloop upgrade --yes here (this PR)

**What would automate it.** release prepare could warn when a commit in the range has a Vloop-Brief: line that git does not read as a trailer, or the notes tool could read the line anywhere in the body

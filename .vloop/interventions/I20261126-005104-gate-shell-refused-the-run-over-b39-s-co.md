---
id: I20261126-005104-gate-shell-refused-the-run-over-b39-s-co
brief: B20261125-2207-b40-i156-mixed-os-pwsh.loop-brief
phase: halt
kind: repair
automatable: yes
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-11-26
recorded: 2026-11-26T00:51:04Z
---
gate shell refused the run over B39's completed sh plan; Git's bin put on the run tab's PATH

**Trigger.** vloop run preflight exit 1 in a plain pwsh tab: gate shell the plan's shell "sh" is not on PATH; the plan was B39's, complete, which the plan row says the next brief's plan replaces

**Done.** prepended C:/Program Files/Git/bin (bash, git, sh only) to the run tab's PATH and re-ran; recorded the doctor defect

**Context.** run 20261125-230701 start, before anything ran; doctor in Git Bash had passed because its PATH has sh

**What would automate it.** doctor's gate shell row uses the configured shell when the plan in use is complete and will be replaced

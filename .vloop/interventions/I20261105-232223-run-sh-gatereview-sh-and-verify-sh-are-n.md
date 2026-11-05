---
id: I20261105-232223-run-sh-gatereview-sh-and-verify-sh-are-n
brief: B20261105-2002-b37-i88-metrics-by-session-kit-without-python.loop-brief
phase: close
kind: repair
automatable: yes
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-11-05
recorded: 2026-11-05T23:22:21Z
---
run.sh, gatereview.sh and verify.sh are not executable

**Trigger.** tools/copilot-acceptance/run.sh as the README writes it: Permission denied; the mode is 100644 since before B37

**Done.** ran it as sh run.sh; set mode 100755 on the three scripts after the acceptance

**What would automate it.** a check that every .sh in tools/ is 100755

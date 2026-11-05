---
id: I20261105-195353-worked-example-run-by-hand-verify-s-sand
brief: B20261103-0528-b36-i96-i97-plugin-per-repo-and-agent.loop-brief
phase: verify
kind: ceremony
automatable: yes
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-11-05
recorded: 2026-11-05T19:53:52Z
---
worked example run by hand: verify's sandbox refused the built binary

**Trigger.** verify reported every P1, P3 and P4 example line not run: the sandbox denied running the built vloop

**Done.** ran the P1 lines (init, upgrade, --dry-run, the not-an-object cases, the kept file) through the branch binary on temporary repositories; every line matched. P3 and P4 checked against the real claude and copilot lists in the acceptance

**Context.** the P1-P4 lines are also named tests in internal/cli, which the run's checks pass

**What would automate it.** a verify fence that allows running the binary the run built under .vloop/tmp

---
id: I20261105-195354-acceptance-the-plugin-declared-per-repos
brief: B20261103-0528-b36-i96-i97-plugin-per-repo-and-agent.loop-brief
phase: verify
kind: ceremony
automatable: partly
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-11-05
recorded: 2026-11-05T19:53:52Z
---
acceptance: the plugin declared per repository, on real Claude and Copilot

**Trigger.** the brief's acceptance, at close

**Done.** items 1-5 pass: init writes the declaration; Copilot installs on trust and lists it enabled only in the clone; the upgrade path enables it after init; Copilot run session homes hold no plugin; Claude run sessions load one vloop plugin (vloop@inline); claude plugin install --scope project offers /vloop:operate in the clone only; doctor's plugin and copilot plugin rows pass

**Context.** url-shortener-tests/b36-accept and b36-upgrade, the branch binary; work on Copilot gpt-6.1-terra, then on Claude

**What would automate it.** an acceptance kit with a throwaway COPILOT_HOME

---
id: I20261012-0235-the-operator-s-uncommitted-plugin-settin
brief: B20261011-2251-b14-learning-loop.loop-brief
phase: setup
kind: ceremony
automatable: partly
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-12
recorded: 2026-10-12T01:35:22Z
---
The operator's uncommitted plugin setting held in a stash during B14's run and restored after

**Trigger.** An uncommitted .claude/settings.json change (enabledPlugins vloop@vloop) was in the worktree when B14 was set ready; the driver needs a clean tree.

**Done.** git stash push of .claude/settings.json before the run as 'operator: enable vloop@vloop plugin (held during B14 run)'; popped after verification, left uncommitted for the operator.

**Context.** B14 run setup on branch B20261011-2251-b14-learning-loop; stash@{0}.

**What would automate it.** vloop run's preflight could name the dirty file and offer to stash it.

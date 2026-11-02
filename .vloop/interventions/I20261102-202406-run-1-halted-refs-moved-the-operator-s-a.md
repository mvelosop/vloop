---
id: I20261102-202406-run-1-halted-refs-moved-the-operator-s-a
brief: B20261102-1844-b31-i110-i112-plan-phase-edges.loop-brief
phase: halt
kind: halt
automatable: partly
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-11-02
recorded: 2026-11-02T20:24:06Z
---
Run 1 halted REFS MOVED: the operator's assistant pushed and fetched in the repository during the plan session

**Trigger.** run 20261102-190603 halted at the plan, exit 9: origin/main moved a30ab36 to 475d945 and refs/remotes/origin/docs/copilot-models-list appeared; $2.42 of planning lost

**Done.** left the refs as they were (they were legitimate); discarded the unstamped plan with git checkout HEAD -- .vloop/state/state.json .vloop/state/plan.md .vloop/state/gates and git clean -fdq .vloop/state/gates, kept the run folder, and re-ran (run 20261102-191816, complete)

**Context.** while the plan session ran, the assistant pushed the docs/copilot-models-list branch and fetched (PR #117, made in a worktree, which shares refs with the repository); the driver's refs guard saw the remote-tracking refs move and halted, correctly

**What would automate it.** a warning from vloop run, or a lock the operator's tools can see, that a run is in progress in this repository and its worktrees; meanwhile the rule is: no push, fetch or merge in a repository while a run is going

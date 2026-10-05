---
id: I20261005-1125-b11-s-first-run-halted-with-exit-9-at-ac
brief: B20261005-0933-quality-pass.loop-brief
phase: halt
kind: repair
automatable: partly
by: both
schema: intervention/v2
options: 3
recommended: 1
decided: 2
agreement: other-option
occurred: 2026-10-05
recorded: 2026-10-05T10:25:06Z
---
B11's first run halted with exit 9 at acceptance: VS Code wrote .git/config during T5's base gate; the plan was discarded and the run restarted with VS Code closed

**Trigger.** the operator asked how the run was doing; the driver had exited 9 at 11:0x: the gate of T5 changed .git/config — nothing was committed

**Done.** Found the change was VS Code's [branch] vscode-merge-base section for the new work branch, not the gate (it runs git only with -C in its own temporary repository). Refs untouched. Restored .vloop/state to HEAD (the 15-task plan, $8.23 and 26 minutes, unstamped and so not resumable, was discarded with its run folder); the operator closed VS Code; restarted.

**Context.** Run B20261005-0933-quality-pass, acceptance, T5's base gate. The watcher missed the halt: its pgrep pattern matched its own command line.

**Options.**
1. discard the uncommitted plan, keep .git/config as VS Code left it, re-run
2. the same, with VS Code closed for the run
3. stamp the plan by hand and resume, skipping the rest of acceptance and the gate review

**Recommended.** VS Code writes that line once per branch, so it should not recur; option 3 skips the gate review

**Decided.** the operator closed VS Code and chose option 2

**What would automate it.** a guard that tells repository config the driver cares about (hooks, fsmonitor, core.*) from editor bookkeeping keys; acceptance that survives a halt

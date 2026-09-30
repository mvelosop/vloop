---
id: I20260930-2159-t8-gate-wrote-toml-into-the-wrong-table
brief: B20260930-2007-vloop-init-upgrade-doctor.loop-brief
phase: run
kind: halt
automatable: partly
by: assistant
occurred: 2026-09-30
recorded: 2026-09-30T20:59:01Z
---
T8 blocked twice on a gate that wrote TOML into the wrong table

**Trigger.** The run stalled again (exit 3). T8's gate appended shell = "pwsh" to a config init had written ending in a [metrics] table, so it became metrics.shell and the configured shell stayed sh; a correct doctor reported pass where the gate expected a warning.

**Done.** Applied the work session's fix — write the key before the config's content — with .loop/amend.sh verify, reset T8, ran the gate by hand (passes), resumed. Recorded as a plan/gate defect, the second in this run.

**What would automate it.** As for T6's gate. Both gate defects in this run come from the planner writing fixture edits it never executed; a planner that runs each gate against the base commit (it must fail) and against a stub of the intended change would catch this class before the run.

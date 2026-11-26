---
id: I20261126-115225-b42-s-first-attempt-halted-with-exit-9-f
brief: B20261126-1008-b42-i162-fence-reads-pwsh.loop-brief
phase: halt
kind: repair
automatable: partly
by: assistant
schema: intervention/v2
options: 3
recommended: 2
decided: 2
agreement: recommended
occurred: 2026-11-26
recorded: 2026-11-26T11:52:25Z
---
B42's first attempt halted with exit 9 (FENCE BYPASSED) in the gate review: Copilot's runtime read a background command's output with read_powershell, a tool the fence did not list; the operator allowed Copilot's read/list/stop shell-session tools by hand and re-planned

**Trigger.** vloop run exit 9: FENCE BYPASSED gate-review, 1 call (read_powershell)

**Done.** TestShellSessionTools first (failed), then the case in internal/fence/judge.go, the guide sentence, the brief's starting point; the halted plan discarded, the run folder kept

**Context.** The gate-review session started a long go test in the background; the model's own read_powershell call was denied by the hook, then Copilot's runtime called read_powershell itself (no hook) when the command finished. The backstop judged that call and halted. It only read output of a command the fence had judged.

**Options.**
1. Resume as is (the same call can recur on any long command)
2. Allow read/list/stop shell-session tools in the fence by hand, rebuild the run binary from the branch, re-plan
3. Add it as a task to B42 (does not protect this run: its hook is the binary that drives it)

**Recommended.** Harmless tools that touch only an already-judged command; the fix must be in the binary that runs the hook

**What would automate it.** The fence could list every tool Copilot's runtime may call on its own

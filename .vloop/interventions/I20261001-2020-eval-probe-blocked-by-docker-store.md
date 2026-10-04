---
id: I20261001-2020-eval-probe-blocked-by-docker-store
brief: B20261001-1025-vloop-skills.loop-brief
phase: verify
kind: halt
automatable: partly
by: both
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-01
recorded: 2026-10-01T19:25:00Z
---
the one-case eval probe refused to start: the Bash sandbox cannot exclude the Docker credential store

**Trigger.** Resuming B7's close from `docs/design-notes/handoff-b7-close.md`, step 2: the probe of `01-hollow-test` (1 run, no baseline, $1 ceiling).

**Context.** `claude plugin eval` exited 1 in 2 seconds at $0: "the Docker (~/.docker, DOCKER_CONFIG) credential store on this machine holds a symbolic link inside it, so the Bash sandbox cannot reliably exclude it — a Bash-granting evaluation cannot run here". Every vloop case grants Bash, so no case can run on this machine until the store is a plain directory. The assistant may not inspect the credential store; that is the operator's to do. The probe's JSON also showed that the nine review cases' `names-the-defect` llm graders tell the judge to read `.vloop/tmp/verdict.json` but carry no `focus`, so they default to `last_message`.

**Suggested.** Ask the operator before spending: probe first, report, then ask before the $40 suite; write the evals documentation meanwhile. On the halt: the operator makes the Docker store's contents one plain directory (its root may be a link), then the probe is re-run.

**Decided.** The operator: probe first, then ask. The Docker store is left to the operator. Tried at the operator's go-ahead: `DOCKER_CONFIG` pointed at an empty scratch directory for the probe alone; the tool refused again in 3 seconds at $0 with the same message, so it checks `~/.docker` whatever `DOCKER_CONFIG` says. The operator then stopped Docker and renamed `~/.docker` for the evals' duration (to be restored before Docker starts again); the probe ran: 01-hollow-test 1.0 (all three graders, judge 3/3 on the verdict file), $0.23, 30 s.

**What would automate it.** A pre-flight in the evals documentation (or a `vloop` check) that runs one case at a $0 ceiling with Bash granted and reports environment refusals before any spend.

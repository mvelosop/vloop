---
id: I20260929-2010-go-mod-not-tidy
brief: B20260929-1804-vloop-skeleton-config-briefs.loop-brief
phase: verify
kind: verification-finding
automatable: yes
by: assistant
occurred: 2026-09-29
recorded: 2026-09-30T16:54:00Z
backfilled: true
---
go.mod not tidy

**Trigger.** independent verification after a 9/9 first-pass run

**Done.** go mod tidy by hand; recorded as D...-go-mod-not-tidy and its gate gap

**Context.** Verifying B1 after a 9/9 first-pass run: `go mod tidy -diff` showed BurntSushi/toml listed as indirect though imported directly. Tidied by hand; recorded later as a work defect and a plan gate-gap, since no gate checked tidiness.

**Suggested.** run go mod tidy and add a tidy check to future gates.

**Decided.** the operator approved the tidy and the gate.

**What would automate it.** a go mod tidy -diff gate, in every gate since B2

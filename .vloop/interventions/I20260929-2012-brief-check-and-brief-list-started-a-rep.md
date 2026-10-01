---
id: I20260929-2012-brief-check-and-brief-list-started-a-rep
brief: B20260929-1804-vloop-skeleton-config-briefs.loop-brief
phase: verify
kind: verification-finding
automatable: partly
by: assistant
occurred: 2026-09-29
recorded: 2026-09-30T16:54:00Z
backfilled: true
---
brief check and brief list started a reported cycle from different briefs

**Trigger.** reading outputs side by side during verification

**Done.** noted as open; fixed as F1 in B2

**Context.** Verifying B1 by hand: `brief check` reported a depends-on cycle starting from the checked brief, `brief list` from the first brief in its order (b → a → b vs a → b → a). Within B1's contract, which pinned only the message; fixed as F1 in B2.

**Suggested.** pin one rule (start from the smallest name) in B2.

**Decided.** the operator accepted it into B2.

**What would automate it.** a spec that pins one rule; the domain model's invariants are where such rules belong

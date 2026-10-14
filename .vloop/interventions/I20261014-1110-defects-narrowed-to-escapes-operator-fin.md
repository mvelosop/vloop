---
id: I20261014-1110-defects-narrowed-to-escapes-operator-fin
brief: ""
phase: next
kind: decision
automatable: no
by: both
schema: intervention/v2
options: 3
recommended: 1
decided: 1
agreement: recommended
occurred: 2026-10-14
recorded: 2026-10-14T10:10:36Z
---
Defects narrowed to escapes; operator findings become interventions; the defect triage retires

**Trigger.** After B14's close the operator weighed dropping defect registration to focus on finding functional errors sooner; B14's three operator defects each duplicated an intervention, and triage cost a re-rendered run record

**Done.** B15 row: title 'Code health, and defects narrowed to escapes', the new item, doctor's ID resolution dropped; B16 row: provenance in the records and retired rules dropped, the verification phase named the main item

**Context.** docs/design-notes/vloop-roadmap-b12-b17.md, B15 and B16 rows; follows PR #33 (B14) merged as b2d2c65

**Options.**
1. Defects only for escapes: in-loop defects stay derived, operator findings become interventions, vloop defect add stays for defects found after merge, triage fields and provenance checks retire
2. Drop defects entirely: escapes become interventions on the brief that caused them
3. Keep everything and decide after B16

**Recommended.** Escapes are the end-user defects the operator wants measured, and only a defect record captures them; the rest duplicated interventions or was loop bookkeeping

**Decided.** Option 1; the B15 and B16 roadmap rows rewritten to match

**What would automate it.** none: a design decision

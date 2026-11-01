---
id: I20261101-230340-acceptance-on-the-real-cli-the-url-short
brief: B20261101-2044-b26-i78-copilot-gate-review.loop-brief
phase: verify
kind: ceremony
automatable: partly
by: assistant
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-11-01
recorded: 2026-11-01T23:03:40Z
---
Acceptance on the real CLI: the URL shortener brief with agent.gate-review = copilot (gpt-5.6-terra)

**Trigger.** the brief's acceptance at close

**Done.** 3 runs: two check-gap halts closed by checks added to the clone's config, the third 10/10 first pass; 4 Copilot gate-review sessions, all session/v4, 142.6 credits; one false finding with Copilot-masked URLs; 5 hook denials, all shell syntax; recorded in docs/design-notes/copilot-cli-experiments.md

**What would automate it.** tools/copilot-acceptance/gatereview.sh runs the fixture brief; the URL shortener run stays by hand

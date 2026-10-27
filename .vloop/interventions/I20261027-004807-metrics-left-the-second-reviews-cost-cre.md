---
id: I20261027-004807-metrics-left-the-second-reviews-cost-cre
brief: B20261026-1002-b22-fence-second-review.loop-brief
phase: verify
kind: verification-finding
automatable: partly
by: assistant
origin: work
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-10-27
recorded: 2026-10-27T00:48:07Z
---
metrics left the second reviews' cost, credits and time out of the totals and agent time, against docs/domain/measurement/metrics.md; fixed at close

**Trigger.** acceptance run: vloop metrics cost $5.22, the driver's run total $6.11

**Done.** counted second-review cost, credits and time in totals and agent time, as docs/domain/measurement/metrics.md says; second_review gains cost_usd, time_ms, credits (cbcdc91)

**Context.** B22 run 20261026-211937, T6 (metrics/v4); internal/metrics/time.go only counted second-review sessions

**What would automate it.** a gate comparing the metrics total with the run log's last 'run $' figure

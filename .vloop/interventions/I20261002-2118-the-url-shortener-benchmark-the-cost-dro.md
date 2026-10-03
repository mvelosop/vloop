---
id: I20261002-2118-the-url-shortener-benchmark-the-cost-dro
brief: ""
phase: verify
kind: verification-finding
automatable: partly
by: both
occurred: 2026-10-02
recorded: 2026-10-02T20:18:58Z
---
the url-shortener benchmark: the cost drop is the models; vloop's planner costs twice the shell loop's

**Trigger.** The operator asked whether vloop's lower cost in B7's acceptance run came from the model change, the skills being largely the same

**Done.** The shell loop rerun on the new models in ~/source/personal/url-shortener-shell-rerun (9a197a7, defaults: opus and sonnet aliases, $40, 30 iterations), detached under caffeinate -i. Original 2026-08-18 (opus 5, sonnet 5): 10/10, 11 iterations, 1 rejection, $19.87, turns plan 41 / work 327 / review 196, 27 permission denials. Rerun 2026-10-02 (opus 5.5, sonnet 5.5): 11/11, 12 iterations, 0 rejections, 1 gate failure, $3.34 (plan $1.01, work $1.50, review $0.83), turns 12 / 78 / 55, 2 denials. vloop the same day and models: 10/10, 10 iterations, $4.21 (plan $1.98, work $1.39, review $0.84), turns 27 / 82 / 50, 0 denials. Both apps verified on a fresh clone: npm ci, unit and e2e tests, boot, the worked example over HTTP. The models explain the drop (about 6x); work and review match across drivers; vloop's planner costs about 2x, plausibly its gate checks on the base. One pair of runs.

**What would automate it.** a benchmark command that runs a fixed brief under both drivers or two model sets and tabulates turns, cost and outcome per phase

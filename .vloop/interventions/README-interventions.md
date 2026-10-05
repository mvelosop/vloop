---
name: README-interventions
description: Indexes every recorded operator intervention — what the operator, or the assistant as the operator's hands, did around the loop's runs besides testing — by phase, kind and whether a driver could do it
---
# Operator interventions

One record per intervention, one file each: `I<YYYYMMDD-HHMM>-<slug>.md`. The
timestamp is an alias — roughly when it happened, so ids sort
chronologically — not a measurement. What the records say:
`docs/design-notes/vloop-interventions-b1-b4.md`.

Each record's frontmatter (`intervention/v2`; a record without the `schema`
line is v1 and reads as `no-options`):

| Field | Values |
| --- | --- |
| `id` | the file name without `.md` |
| `brief` | the loop brief it belongs to, or `""` for series-level |
| `phase` | `setup`, `design`, `run`, `halt`, `verify`, `close`, `next` |
| `kind` | `direction`, `decision`, `context-supply`, `halt`, `verification-finding`, `repair`, `carry-forward`, `ceremony` |
| `automatable` | `yes`, `partly`, `no` — could a driver do it |
| `by` | `operator`, `assistant`, `both` |
| `schema` | `intervention/v2` |
| `options` | `0`–`3`: how many options the model proposed before the operator decided — never padded to three |
| `recommended` | the option the model recommended, `1`–`3`; `0` with no options |
| `decided` | the option the operator chose, `1`–`3`, `other`, or `""` with no options |
| `adjusted` | `true` when the chosen option was adjusted; absent otherwise |
| `agreement` | derived from the three above: `recommended`, `other-option`, `adjusted`, `different`, `no-options` |
| `occurred`, `recorded` | dates; `backfilled: true` on records written after the fact |

Then the summary line and these sections, each only when it has content:
**Trigger.**, **Done.**, **Context.** (the run, iteration and task; the commit
or files; the triggering output; what changed because of it), **Options.** (a
numbered list, one line each), **Recommended.** (why that option),
**Suggested.** (migrated records only), **Decided.** (one sentence), **What
would automate it.** Record with `vloop intervention add`; correct the choice
with `vloop intervention set`; `vloop intervention migrate` turns v1 records
into v2 without touching their bodies.

## Index

Generated — regenerate with `tools/interventions-index.sh --write`;
`--check` fails when it is stale. Rows sort by phase, then kind, then id.

<!-- index:begin -->
| Phase | Kind | Automatable | Agreement | Id |
| --- | --- | --- | --- | --- |
| setup | repair | yes | no-options | [I20260928-2215-the-loop-s-install-proof-failed-in-every](I20260928-2215-the-loop-s-install-proof-failed-in-every.md) |
| design | carry-forward | yes | no-options | [I20260929-2150-b1-s-open-items-carried-into-b2](I20260929-2150-b1-s-open-items-carried-into-b2.md) |
| design | context-supply | partly | no-options | [I20260929-1222-the-design-conversation-was-reachable-on](I20260929-1222-the-design-conversation-was-reachable-on.md) |
| design | decision | partly | no-options | [I20260929-1804-seven-forks-settled-for-the-first-brief](I20260929-1804-seven-forks-settled-for-the-first-brief.md) |
| design | decision | partly | no-options | [I20260929-2148-ten-choices-in-the-b2-draft-including-mo](I20260929-2148-ten-choices-in-the-b2-draft-including-mo.md) |
| design | decision | partly | no-options | [I20260929-2330-eleven-choices-in-the-b3-draft-and-real](I20260929-2330-eleven-choices-in-the-b3-draft-and-real.md) |
| design | decision | partly | no-options | [I20260930-0929-seven-choices-in-the-b4-draft](I20260930-0929-seven-choices-in-the-b4-draft.md) |
| design | decision | partly | no-options | [I20260930-2002-four-forks-settled-for-b5](I20260930-2002-four-forks-settled-for-b5.md) |
| design | decision | partly | no-options | [I20261001-0720-readme-cap-and-run-budgets](I20261001-0720-readme-cap-and-run-budgets.md) |
| design | decision | partly | no-options | [I20261001-1021-four-forks-settled-for-b7](I20261001-1021-four-forks-settled-for-b7.md) |
| design | decision | partly | no-options | [I20261001-1040-planner-gate-checks-questioned](I20261001-1040-planner-gate-checks-questioned.md) |
| design | decision | partly | no-options | [I20261002-2140-b8-designed-from-the-v1-0-review-securit](I20261002-2140-b8-designed-from-the-v1-0-review-securit.md) |
| design | decision | no | no-options | [I20261003-2053-b9-s-forks-settled-gates-may-be-the-plan](I20261003-2053-b9-s-forks-settled-gates-may-be-the-plan.md) |
| design | decision | no | no-options | [I20261004-1435-b10-s-forks-settled-options-as-body-sect](I20261004-1435-b10-s-forks-settled-options-as-body-sect.md) |
| design | decision | no | adjusted | [I20261005-0935-b11-s-forks-settled-user-facing-quality](I20261005-0935-b11-s-forks-settled-user-facing-quality.md) |
| design | direction | no | no-options | [I20260929-1830-naming-and-scope-corrections-on-review-o](I20260929-1830-naming-and-scope-corrections-on-review-o.md) |
| design | direction | no | no-options | [I20260929-2155-gates-run-in-any-shell-of-the-os-not-onl](I20260929-2155-gates-run-in-any-shell-of-the-os-not-onl.md) |
| design | direction | no | no-options | [I20260929-2318-b3-split-in-two-and-the-roadmap-renumber](I20260929-2318-b3-split-in-two-and-the-roadmap-renumber.md) |
| design | direction | no | no-options | [I20260930-2001-b5-skills-move-to-b6-and-a-b7-review](I20260930-2001-b5-skills-move-to-b6-and-a-b7-review.md) |
| design | direction | no | no-options | [I20260930-2025-monorepos-need-stacks-by-path](I20260930-2025-monorepos-need-stacks-by-path.md) |
| design | direction | no | no-options | [I20260930-2240-b6-split-driver-skills-review](I20260930-2240-b6-split-driver-skills-review.md) |
| design | direction | no | no-options | [I20261001-1020-benchmark-brief-for-acceptance](I20261001-1020-benchmark-brief-for-acceptance.md) |
| design | direction | no | no-options | [I20261001-1041-interventions-as-a-vloop-record](I20261001-1041-interventions-as-a-vloop-record.md) |
| design | direction | no | no-options | [I20261001-1042-calibration-as-review-evals](I20261001-1042-calibration-as-review-evals.md) |
| design | verification-finding | yes | no-options | [I20260930-0935-vloop-s-checker-rejected-two-binding-ref](I20260930-0935-vloop-s-checker-rejected-two-binding-ref.md) |
| design | verification-finding | yes | no-options | [I20260930-2003-marketplace-manifest-invalid-since-b1](I20260930-2003-marketplace-manifest-invalid-since-b1.md) |
| design | verification-finding | yes | no-options | [I20260930-2009-checker-rejects-wrapped-binding-reason](I20260930-2009-checker-rejects-wrapped-binding-reason.md) |
| design | verification-finding | yes | no-options | [I20261001-0724-readme-shell-default-wrong-since-b2](I20261001-0724-readme-shell-default-wrong-since-b2.md) |
| design | verification-finding | yes | no-options | [I20261001-0730-b6-draft-cited-a-folder](I20261001-0730-b6-draft-cited-a-folder.md) |
| design | verification-finding | partly | no-options | [I20261001-1018-roadmap-contradiction-first-real-run](I20261001-1018-roadmap-contradiction-first-real-run.md) |
| run | ceremony | yes | no-options | [I20260929-1917-the-plan-commit-landed-on-main](I20260929-1917-the-plan-commit-landed-on-main.md) |
| run | context-supply | partly | no-options | [I20260929-1900-go-not-installed-and-the-loop-s-fence-ha](I20260929-1900-go-not-installed-and-the-loop-s-fence-ha.md) |
| run | halt | yes | no-options | [I20260929-1941-the-driver-refused-to-resume-on-the-new](I20260929-1941-the-driver-refused-to-resume-on-the-new.md) |
| run | halt | partly | no-options | [I20260929-2224-t1-blocked-twice-on-a-gnu-only-grep-patt](I20260929-2224-t1-blocked-twice-on-a-gnu-only-grep-patt.md) |
| run | halt | partly | no-options | [I20260930-2149-t6-gate-contradicted-its-own-acceptance](I20260930-2149-t6-gate-contradicted-its-own-acceptance.md) |
| run | halt | partly | no-options | [I20260930-2159-t8-gate-wrote-toml-into-the-wrong-table](I20260930-2159-t8-gate-wrote-toml-into-the-wrong-table.md) |
| halt | halt | yes | no-options | [I20261003-2323-b9-halted-on-t2-s-gate-dispute-t1-s-new](I20261003-2323-b9-halted-on-t2-s-gate-dispute-t1-s-new.md) |
| halt | repair | partly | no-options | [I20261003-1147-b8-halted-on-t12-s-gate-dispute-a-fixtur](I20261003-1147-b8-halted-on-t12-s-gate-dispute-a-fixtur.md) |
| halt | repair | yes | no-options | [I20261003-1217-b8-halted-again-on-t17-s-gate-dispute-th](I20261003-1217-b8-halted-again-on-t17-s-gate-dispute-th.md) |
| halt | repair | partly | recommended | [I20261004-2304-b10-blocked-on-t9-a-gate-needed-a-sessio](I20261004-2304-b10-blocked-on-t9-a-gate-needed-a-sessio.md) |
| halt | repair | partly | other-option | [I20261005-1125-b11-s-first-run-halted-with-exit-9-at-ac](I20261005-1125-b11-s-first-run-halted-with-exit-9-at-ac.md) |
| verify | decision | partly | no-options | [I20260929-2246-the-brief-contradicted-itself-on-clearin](I20260929-2246-the-brief-contradicted-itself-on-clearin.md) |
| verify | decision | no | no-options | [I20260930-0028-a-work-session-created-the-repo-s-own-vl](I20260930-0028-a-work-session-created-the-repo-s-own-vl.md) |
| verify | halt | partly | no-options | [I20261001-2020-eval-probe-blocked-by-docker-store](I20261001-2020-eval-probe-blocked-by-docker-store.md) |
| verify | repair | partly | no-options | [I20260930-0805-the-planning-session-moved-this-repo-s-r](I20260930-0805-the-planning-session-moved-this-repo-s-r.md) |
| verify | repair | yes | no-options | [I20260930-1300-a-rehearsal-clone-took-the-work-branch-a](I20260930-1300-a-rehearsal-clone-took-the-work-branch-a.md) |
| verify | repair | partly | no-options | [I20261001-1415-eval-port-fixed-by-hand](I20261001-1415-eval-port-fixed-by-hand.md) |
| verify | repair | partly | no-options | [I20261001-2035-llm-grader-focus-fixed](I20261001-2035-llm-grader-focus-fixed.md) |
| verify | repair | partly | no-options | [I20261004-1143-b9-verified-the-worked-example-held-by-h](I20261004-1143-b9-verified-the-worked-example-held-by-h.md) |
| verify | verification-finding | yes | no-options | [I20260929-2010-go-mod-not-tidy](I20260929-2010-go-mod-not-tidy.md) |
| verify | verification-finding | partly | no-options | [I20260929-2012-brief-check-and-brief-list-started-a-rep](I20260929-2012-brief-check-and-brief-list-started-a-rep.md) |
| verify | verification-finding | yes | no-options | [I20260929-2245-one-review-session-s-telemetry-record-wa](I20260929-2245-one-review-session-s-telemetry-record-wa.md) |
| verify | verification-finding | partly | no-options | [I20260930-0026-the-task-estimate-was-read-from-the-firs](I20260930-0026-the-task-estimate-was-read-from-the-firs.md) |
| verify | verification-finding | partly | no-options | [I20260930-1302-the-loop-s-fence-blocked-the-read-only-g](I20260930-1302-the-loop-s-fence-blocked-the-read-only-g.md) |
| verify | verification-finding | yes | no-options | [I20260930-1305-schema-files-counted-as-other-lines](I20260930-1305-schema-files-counted-as-other-lines.md) |
| verify | verification-finding | partly | no-options | [I20261001-1005-plan-md-run-id-wording](I20261001-1005-plan-md-run-id-wording.md) |
| verify | verification-finding | yes | no-options | [I20261001-1338-eval-cases-fail-to-load-setup-key](I20261001-1338-eval-cases-fail-to-load-setup-key.md) |
| verify | verification-finding | partly | no-options | [I20261001-2310-eval-suite-first-run](I20261001-2310-eval-suite-first-run.md) |
| verify | verification-finding | partly | no-options | [I20261002-1426-b7-s-acceptance-run-restarted-detached-t](I20261002-1426-b7-s-acceptance-run-restarted-detached-t.md) |
| verify | verification-finding | partly | no-options | [I20261002-2118-the-url-shortener-benchmark-the-cost-dro](I20261002-2118-the-url-shortener-benchmark-the-cost-dro.md) |
| verify | verification-finding | partly | recommended | [I20261005-0757-b10-verified-the-worked-example-held-by](I20261005-0757-b10-verified-the-worked-example-held-by.md) |
| close | carry-forward | partly | no-options | [I20260929-2325-the-bsd-tools-lesson-into-later-briefs](I20260929-2325-the-bsd-tools-lesson-into-later-briefs.md) |
| close | carry-forward | partly | no-options | [I20260930-0809-the-six-known-defects-of-b1-b3-recorded](I20260930-0809-the-six-known-defects-of-b1-b3-recorded.md) |
| close | ceremony | yes | no-options | [I20260929-2046-manual-close-run-record-consumed-status](I20260929-2046-manual-close-run-record-consumed-status.md) |
| close | ceremony | yes | no-options | [I20260929-2321-manual-close-of-b2](I20260929-2321-manual-close-of-b2.md) |
| close | ceremony | yes | no-options | [I20260930-0815-manual-close-of-b3-with-the-vloop-brief](I20260930-0815-manual-close-of-b3-with-the-vloop-brief.md) |
| close | ceremony | yes | no-options | [I20260930-1330-the-first-close-by-vloop-itself-plus-han](I20260930-1330-the-first-close-by-vloop-itself-plus-han.md) |
| close | ceremony | yes | no-options | [I20260930-2215-b5-closed-by-vloop-after-three-runs](I20260930-2215-b5-closed-by-vloop-after-three-runs.md) |
| close | ceremony | yes | no-options | [I20261001-1015-b6-closed-with-a-finding](I20261001-1015-b6-closed-with-a-finding.md) |
| close | repair | yes | no-options | [I20260929-2048-local-main-reset-to-the-pre-merge-commit](I20260929-2048-local-main-reset-to-the-pre-merge-commit.md) |
| next | carry-forward | no | no-options | [I20260930-0900-the-shell-loop-fixed-so-sessions-cannot](I20260930-0900-the-shell-loop-fixed-so-sessions-cannot.md) |
| next | carry-forward | partly | no-options | [I20260930-0950-the-operator-role-written-down-claude-md](I20260930-0950-the-operator-role-written-down-claude-md.md) |
| next | ceremony | yes | no-options | [I20260930-1010-the-b4-draft-reached-main-inside-an-unre](I20260930-1010-the-b4-draft-reached-main-inside-an-unre.md) |
| next | ceremony | partly | no-options | [I20261004-1337-v2-0-0-beta-1-released-as-a-pre-release](I20261004-1337-v2-0-0-beta-1-released-as-a-pre-release.md) |
| next | ceremony | partly | recommended | [I20261005-0906-v2-0-0-beta-2-released-so-b11-runs-on-a](I20261005-0906-v2-0-0-beta-2-released-so-b11-runs-on-a.md) |
| next | direction | no | no-options | [I20260929-2130-the-metrics-and-defects-design-before-b2](I20260929-2130-the-metrics-and-defects-design-before-b2.md) |
| next | direction | no | no-options | [I20260930-0955-a-project-level-driver-over-project-stat](I20260930-0955-a-project-level-driver-over-project-stat.md) |
| next | direction | no | no-options | [I20260930-1654-a-documentation-layer-for-the-design-ste](I20260930-1654-a-documentation-layer-for-the-design-ste.md) |
| next | direction | no | no-options | [I20260930-1956-refocus-on-v1-and-intervention-data](I20260930-1956-refocus-on-v1-and-intervention-data.md) |
| next | direction | no | no-options | [I20261001-1043-horizon-now-next-later](I20261001-1043-horizon-now-next-later.md) |
| next | direction | no | no-options | [I20261001-1105-interventions-record-suggestion-and-context](I20261001-1105-interventions-record-suggestion-and-context.md) |
| next | direction | no | no-options | [I20261003-1811-the-v2-0-release-planned-the-gate-model](I20261003-1811-the-v2-0-release-planned-the-gate-model.md) |
<!-- index:end -->

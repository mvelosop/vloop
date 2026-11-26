# Journal — B20261126-1843-b47-i160-vloop-folder-readme

Append-only narrative of this plan. Rendered state lives in .vloop/state/plan.md.

## Plan — B20261126-1843-b47-i160-vloop-folder-readme

- **Brief:** `docs/briefs/B20261126-1843-b47-i160-vloop-folder-readme.loop-brief.md`
- **Tasks:** 4

## T1 — Generate the owned README during init

- **Outcome:** done (review: PASS)
- **Summary:** Init now writes a versioned, language-specific, owned .vloop/README.md with the required workflow, folder inventory, commands and project link, while preserving the documented output order and overwriting existing local text.
- **Files:** cmd/vloop/b5_e2e_test.go, internal/cli/init.go, internal/cli/init_test.go, internal/cli/vloop_readme.go
- **Notes for next iteration:** The reusable vloopReadme renderer is shared by language and returns LF-only text; init places its write step after the stamp and optional guidelines line, before the starter brief. Go tests cover both languages, version insertion, overwrite behavior, inventory and output assertions.

## T2 — Reconcile the owned README during upgrade

- **Outcome:** done (review: PASS)
- **Summary:** Upgrade now reconciles .vloop/README.md from the resolved language and current binary version, including same-version repair and dry-run output, while preserving normal and breaking-upgrade write barriers.
- **Files:** internal/cli/upgrade.go, internal/cli/upgrade_test.go, cmd/vloop/b5_e2e_test.go
- **Notes for next iteration:** Upgrade reuses vloopReadme so init and upgrade emit byte-identical content; README changes are reported before same-version status, and participate in the existing dry-run and breaking-upgrade change list. Tests cover both languages, missing, edited, current, dry-run, same-version and breaking paths, plus updated exact-output assertions.

## T3 — Document the owned README lifecycle

- **Outcome:** done (review: PASS)
- **Summary:** The setup guide now states that init writes `.vloop/README.md` and upgrade rewrites or restores it.
- **Files:** docs/guide/concepts.md
- **Notes for next iteration:** The change is limited to the prescribed lifecycle sentence in the Setting up a repository section.

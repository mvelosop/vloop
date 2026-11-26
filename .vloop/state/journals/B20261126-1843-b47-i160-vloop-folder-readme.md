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

---
id: I20260930-2025-monorepos-need-stacks-by-path
brief: B20260930-2007-vloop-init-upgrade-doctor.loop-brief
phase: design
kind: direction
automatable: no
by: operator
occurred: 2026-09-30
recorded: 2026-09-30T19:59:40Z
---
monorepos need stacks by path

**Trigger.** Reviewing the B5 draft, the operator pointed out that a monorepo mixes technologies by path, which one repo-wide stacks list cannot express.

**Done.** B5 gained F2: stack@path scopes in metrics.stacks, the scope replacing the unscoped stacks in its subtree, the longest scope winning, and init detecting stacks per directory. Two forks settled with the operator (syntax, replace vs add).

**Context.** Reviewing the B5 draft, the operator: 'we have to handle monorepos, where we can have different techs by path'. Two forks followed: stack@path syntax, and a scope that replaces the unscoped stacks.

**Suggested.** stack@path in the list; the scope replaces the unscoped stacks.

**Decided.** the operator chose both recommendations.

**What would automate it.** None for the direction. A fixture repository per target shape (single-stack, monorepo) in the design act's survey would have surfaced it before review.

---
type: command
weight: 3
---
The tree is still the base: the planner writes only the plan. Run

    for id in $(jq -r '.tasks[].id' .vloop/state/state.json); do
      if vloop task gate "$id" >/dev/null 2>&1; then echo "gate $id passes on the base"; exit 1; fi
    done

It must exit 0. `bin/greet` already prints `hello`, so a gate that greps for
`hello` passes before any work exists; the planner was to rewrite it to assert
the exact `hello, world`.

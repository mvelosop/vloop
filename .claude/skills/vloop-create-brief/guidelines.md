# What makes a brief good

A brief is read by fresh sessions that never saw the conversation that produced
it: the planner, the gate review, a work session per task, a review per task,
verify. Each one trusts the text. Write for that reader.

## The contract

- **Pin decisions, leave mechanics open.** Name the behaviour, the formats, the
  error cases, the worked example. Do not name classes, modules or file layout
  unless the operator decided them: that limits the plan and buys nothing.
- **Record what was rejected.** Each decision carries the option someone will
  propose and why it lost. A written rejection is one no session spends an
  attempt rediscovering.
- **Literal values stay literal.** When a gate or a test will compare a value (a
  response body, a field list, a message, a format, a file name), write the
  value itself. Never describe it instead ("the framework's default", "as the
  library does", "the usual format"): a description and the literal can
  disagree, and the plan keeps whichever reads easier. If you state a value
  twice, the two must agree.
- **The worked example pins the values, error cases included**: the body of an
  error, not only its status; the message, not only the exit code. It arbitrates
  when two implementations disagree, so it is exact, never illustrative.
- **Planted failures**: name the cases a gate must catch, planted in a fixture or
  a scratch copy. A gate never seen to fail proves nothing.
- **Counts come from the tree.** If a value is counted or generated, say the
  gate computes it; a hard-coded count is wrong the day anything else lands.
- **Out of scope names what someone would plausibly add**: the sibling slices,
  the adjacent feature, the "while we're here" improvement, the drift the run
  will uncover but must only report. At least two items.
- **One contract, one place.** Either the brief states the contract, or it
  delegates it whole to one named document. Never half and half.

## Binding references

- One repository-relative path per entry, each with the reason it binds:
  ``- `docs/.../x.md` — the invariant this run must keep``.
- Never a folder, unless the folder is a bundle whose whole binds, cited through
  its index file.
- **Cite a rule instead of restating it.** A restated rule drifts from its
  source, and nothing breaks the tie.
- Every cited document is re-read by every session, every iteration. Cite what
  binds, not what is related.
- Paths outside the repository cannot bind: sessions cannot read them.

## What the loop can and cannot do

- **Checks run on the base before planning and after every iteration.** A
  permanent check a task adds must be green right after that task, not only at
  the end. A check that turns red after task 1 and stays red until task 5 blocks
  task 1 forever.
- **A gate follows the surface of its task** and never runs the whole suite: a
  flaky test in code the branch never touched blocks a correct task, attempt
  after attempt.
- **No task may weaken a check to pass.** An earlier test changes only where it
  pins behaviour the contract changes, and only to the new behaviour.
- **Sessions cannot**: commit, move git refs, edit `.vloop/`, reach the network,
  log in interactively, or use a tool that is not installed. Whatever needs that
  belongs to the operator, before or after the run; say which, in the header.
- **The environment is part of the brief**: the machine the run needs, the
  tools, the environment variables, the services, and what it must never touch.
- **Prose-only tasks**: if no script can check a task, say what happens when its
  review fails twice on wording: block it, and the operator finishes it.

## Size and shape

- If the goal does not fit in three sentences, the brief covers too much: split
  it, and the other half becomes an out-of-scope item.
- The "Shape" section gives the task count as a range ("De 4 a 6 tareas" /
  "4 to 6 tasks"), each task checkable with one command, in an order the checks
  survive. The last task runs every gate at once and the worked example line by
  line.
- A long-lived artifact that must start, a migration that must apply, an
  end-to-end path no unit test covers: name its own gate, or it ends up inside a
  task that does not prove it.

## Fix briefs

One subsection per defect: **Defect** (what happens, where, the mechanism),
**Required** (the result, not the code), **Gate** (the assertion shipped with the
fix; it must fail against today's code, shown before fixing). Say why the gates
missed it: a fix that does not close the blind spot reopens next time.

## Explaining an item to the operator

The brief keeps the short form. When the operator asks about an item, explain it
in the conversation as a short block: one sentence on what it does; **Today:**
what is wrong, with the evidence; **After:** the exact messages or output the
contract pins; and only where they apply, one line each of **Safety:**,
**Unchanged:**, **Side effect:**, **Limit:**, **Not included:**.

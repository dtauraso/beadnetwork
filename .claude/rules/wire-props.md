---
paths:
  - "nodes/Wiring/loader.go"
  - "tools/topology-vscode/src/schema/wire-defs.ts"
---

# Wire props — a tag alone does not reach the screen

Wire props (`WireProps` from `tools/topology-vscode/src/schema/wire-defs.ts`, generated from
`wire:"prop,..."` tags on `specEdge` in `nodes/Wiring/topo_spec.go`) are Go-owned edge
metadata from the spec JSON.

Today the only prop is `label`, and it does **NOT** feed the render path: there is no
per-edge drawn line at all any more (the source node's own chain of placeholder beads is
the edge's visual, `docs/beads-are-the-edge.md`) — the Edge block's SX..EZ/Selected columns
still stream, read only by the `.probe` debug decoder, not by anything that draws. `label`
rides the Edge block's EdgeLabelOff/EdgeLabelLen columns solely for that same `.probe` log,
never for drawing.

If a NEW wire prop needs to affect rendering, it must be packed into the Edge block
(`Buffer/` + `buffer-layout.ts`) and read by whatever the render path is at the time in the
same commit — a `wire:"prop,..."` tag alone does not reach the screen.

(An `EdgeKind`-typed `kind` prop was tried and removed: it had no Edge-block column and
could not affect a single pixel; its only consumer was a test importing the schema barrel,
not production code.)

Pulse speed is uniform across all wires — reject per-wire `speed` props
(`memory/feedback_uniform_pulse_speed.md`).

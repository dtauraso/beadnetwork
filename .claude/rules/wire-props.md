---
paths:
  - "Categories/Scene/Topology/topo_spec.go"
  - "Categories/Scene/Topology/wire-defs.ts"
---

# Wire props — a tag alone does not reach the screen

Wire props (`WireProps` from `Categories/Scene/Topology/wire-defs.ts`, generated from
`wire:"prop,..."` tags in `Categories/Scene/Topology/topo_spec.go`) are Go-owned edge
metadata.

Today the only prop is `label`, and it does **NOT** feed the render path: there is no
per-edge drawn line at all any more (the source node's own chain of placeholder beads is
the edge's visual, `docs/model/entities.md`).

If a NEW wire prop needs to affect rendering, it must be added to the edge's value list
(`Categories/Node/Edge/edge_values.go` and its generated TS twin, which together name what
`view/edges/<row>/edge.bin` holds) and read by the render path in the same commit — a
`wire:"prop,..."` tag alone does not reach the screen.

(An `EdgeKind`-typed `kind` prop was tried and removed: it had no Edge-block column and
could not affect a single pixel; its only consumer was a test importing the schema barrel,
not production code.)

Pulse speed is uniform across all wires — reject per-wire `speed` props
(`memory/feedback/architecture/geometry/feedback_uniform_pulse_speed.md`).

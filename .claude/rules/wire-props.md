---
paths:
  - "nodes/Wiring/loader.go"
  - "tools/topology-vscode/src/schema/wire-defs.ts"
  - "tools/topology-vscode/src/webview/three/EdgeTube.tsx"
---

# Wire props — a tag alone does not reach the screen

Wire props (`WireProps` from `tools/topology-vscode/src/schema/wire-defs.ts`, generated from
`wire:"prop,..."` tags on `specEdge` in `nodes/Wiring/loader.go`) are Go-owned edge metadata
from the spec JSON.

Today the only prop is `label`, and it does **NOT** feed the render path: `EdgeTube`
(`tools/topology-vscode/src/webview/three/EdgeTube.tsx`, the edge renderer) reads only
SX..EZ/Selected from the Edge block; `label` rides the Edge block's
EdgeLabelOff/EdgeLabelLen columns solely for the `.probe` buffer-decoded log, never for
drawing.

If a NEW wire prop needs to affect rendering, it must be packed into the Edge block
(`Buffer/` + `buffer-layout.ts`) and read by `EdgeTube` in the same commit — a
`wire:"prop,..."` tag alone does not reach the screen.

(An `EdgeKind`-typed `kind` prop was tried and removed: it had no Edge-block column and
could not affect a single pixel; its only consumer was a test importing the schema barrel,
not production code.)

Pulse speed is uniform across all wires — reject per-wire `speed` props
(`memory/feedback_uniform_pulse_speed.md`).

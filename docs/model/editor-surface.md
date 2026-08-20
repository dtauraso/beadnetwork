# Model — editor surface (TS)

[← MODEL.md](../../MODEL.md)

## Editor surface (TS)

The model lives entirely in Go. The TS/React layer is **render + forward
only**: it decodes the binary content buffer Go streams and draws it, and
forwards raw input to Go. It holds NO domain state — no render stores, no
spec store, no camera store — never sets node state and never tells Go
when a bead has arrived. Go owns the clock.

- **Go runtime** owns all node-local held state, firing rules, wire
  traversal timing, node positions, per-edge curve geometry, shading
  parameters, camera pose, selection, and overlay visibility. There is no
  single combined buffer or central packer: each emitting goroutine packs
  and streams its OWN binary content buffer to its OWN dedicated inherited
  stdio pipe (`src/Node/Wiring/streamwire/stream_fds.go`, memory/feedback/architecture/bridge/feedback_no_single_writer_bridge.md)
  — one VIEW stream (camera/overlay/scene, the gesture/stdin-reader
  goroutine), one stream per edge row (that edge's geometry, written by
  the SOURCE node's own goroutine — a node draws its own out-edges), one
  stream per node row (that node's own geometry+ports+label), one
  BEAD stream per node row (that node's live beads on every wire leaving
  it), and one INTERIOR stream per node row (that node's own
  Update-goroutine's interior beads — the ONLY writer of that node's four
  interior slots). **There are exactly three per-node streams.** A driven
  out is stepped by the node's own loop, so its events ride that node's
  interior stream like everything else the node emits. Do not add a fourth
  per-node or per-port stream to "keep writers apart" — keep the writers
  singular instead. Two goroutines racing one fd interleave their frames'
  header and payload writes into garbage, which is why the rule is one pipe
  per EMITTING GOROUTINE (CLAUDE.md, "Bridge surface") and not one pipe per
  thing you want to see separately. Frames on a dedicated fd are `[len:u32-LE][payload]`
  with NO tag byte — the fd POSITION identifies the stream/row.
  `WIREFOLD_STREAM_FDS` (the ext host's spawn env var,
  `src/extension/runCommand.ts`) is **mandatory**: there is no
  central accumulator and no fallback path left to fall back to.
- **Go → TS is binary content buffers** (`buffer-snapshot`) ALONE — no
  sidecar. Each node's kind is a numeric `KindId` column (TS maps it to
  `NODE_DEFS` colors), its label rides its own stream frame's inline label
  bytes, and its identity is the buffer ROW INDEX (Go resolves row → node
  for hits). The ext host relays each dedicated-fd frame to the webview
  under a synthetic tag (`BUF_BLOCK_TAG_VIEW`/`_EDGE_STREAM`/`_NODE_STREAM`/
  `_INTERIOR_STREAM`, `src/schema/buffer-layout/frame_tags.go`) purely for cell routing —
  never a wire byte. The webview decodes each stream (`buffer-decode-view.ts`/
  `buffer-decode-edge.ts`/`buffer-decode-node.ts`/`buffer-decode-interior.ts`)
  and renders it; row-keyed reflect resources (`snapshot-buffer.ts`,
  `overlay-flags.ts`) mirror Go — they author nothing. There is **no
  JSON-trace render path and no `pump.ts`**; Go emits no trace-event JSON
  on stdout at all — the `.probe` trace logs (`go.jsonl`/`go-node.jsonl`/
  `go-edge.jsonl`/`go-interior.jsonl`) are the ext host's DECODE of each
  per-owner stream's own trailing EVENTS section (`buffer-log.ts`), not a
  stdout parse. Stdout carries only the DEBUG BREADCRUMB channel's sparse
  `{"kind":"breadcrumb",...}` control-event lines.
- **`BufferScene`** (`src/webview/scene/buffer-scene.tsx`)
  is the composition root of the render tree — it decodes the buffer and
  assembles the per-concern components that draw ALL geometry from it. It is a
  small file; the drawing lives in its siblings under `three/scene/`. Grep the symbol,
  not this filename. The tree covers: node bodies (`src/Ring/NodeShape/NodeInstances.tsx` — sphere
  mesh + ring, keyed off `node.data.fill`/`node.data.stroke` from `NODE_DEFS`; no port
  geometry — a port is a load-time channel-binding ROLE, never drawn),
  transit and interior
  beads (`src/Node/BeadAnimation/ChainBeadInstances.tsx`, `src/webview/scene/beads/InteriorBeadInstances.tsx` — there is no
  per-edge drawn tube any more; the source node's own chain of placeholder beads is the
  edge's visual, `docs/model/entities.md`), the selection ring, its halo and the hover
  ring (placed with everything else drawn at a node's own frame, in
  `src/Ring/NodeShape/node-instances-update.ts`; their shape lives in
  `src/Ring/NodeShape/node-highlight-shape.ts`), and the camera (`src/Camera/BufferCamera.tsx` maps the buffer
  Camera row onto the three.js camera). Nothing in this tree owns traversal
  timing, positions, or geometry.
- **Bridge surface — binary BOTH ways.** **Go → TS:** the binary content
  buffer ALONE (`buffer-snapshot`, each goroutine streamed over its own
  dedicated inherited-stdio pipe) — stated in full under "Go → TS is the
  binary content buffer" above; not restated here, so the two copies cannot
  drift apart. **TS → Go:** framed binary records on stdin (`[len:u32-LE][record]`,
  symmetric in framing style with the dedicated per-goroutine streams, though
  stdin records carry no block-tag byte — that discriminator exists only on
  the Go → TS direction, where the ext host adds a synthetic tag purely for
  cell routing) — `raw-input` (raw pointer/wheel + the stateless raycast
  hit as numeric rows; Go's gesture FSM decides what each gesture MEANS), the
  geometry-CRUD `edit` (`op` = update — the sole remaining op; a `create` /
  `delete` op pair was removed end-to-end, no live TS sender ever emitted them.
  `update` sets a numeric attribute on a typed entity, e.g. overlays toggle/set
  as a flag-id / bitfield), a bare `save` command (Go persists its OWN current
  state — camera + overlays — the editor sends no scene payload). There is NO JSON on
  either wire. The TS → Go send is fire-and-forget: the editor places a record
  and never blocks on Go, never asks when a bead arrived, and there is no
  delivery signal — Go times its own delivery. Nothing about node-local state
  or animation internals crosses the bridge.

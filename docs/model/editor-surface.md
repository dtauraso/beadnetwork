# Model — editor surface (TS)

[← MODEL.md](../../MODEL.md)

## Editor surface (TS)

The model lives entirely in Go. The TS/React layer is **render + forward
only**: it reads what Go wrote and draws it, and forwards raw input to Go.
It holds NO domain state — no render stores, no spec store, no camera
store — never sets node state and never tells Go when a bead has arrived.
Go owns the clock.

- **The scene is BLOCK FILES, not a stream.** Each owner's own goroutine
  writes its own file and is the only writer of those bytes: a node writes
  `view/nodes/<row>/node.bin` plus its `beads.bin`, `interior.bin`,
  `channel-vectors.bin` and `tilt-arrows.bin`; the SOURCE node of an edge
  writes `view/edges/<row>/edge.bin`; the view goroutine writes the
  singletons (camera, overlays, panels, chrome pieces, pointer target,
  slider, owner counts, ring points). A file is sections of
  `u32 length + bytes` in a generated name order, written whole through
  tmp+rename, so a reader never sees half a frame. **The row is in the PATH,
  never in the file** — that is what keeps one writer per owner without a
  lock. The renderer fetches one file per block and slices it
  (a concern-owned `leaf-values.ts` for singletons, `row-leaf-values.ts` for
  per-owner), at an interval for human-speed state and per frame for
  anything tracking the cursor, a drag, or a tick. Go writes only when the
  value changes: `BlobWriter.Flush` compares the whole payload first.
- **Nothing streams. Go → TS is files, and only files.** There is no frame,
  no per-goroutine pipe, no `BEADNETWORK_STREAM_FDS`, no framing or demux in the
  ext host, and no host→webview message of any kind. Go inherits three stdio
  slots; stderr is the only one that carries anything, because an error has to
  reach the human before any file is written.

  The trace events were the last thing to ride a frame. Each goroutine now
  appends them, as fixed-width binary records, to the file of the item the
  event is about — `view/nodes/<row>/trace.bin` and its `interior-`/`beads-`
  siblings, `view/edges/<row>/trace.bin`, `view/trace.bin`
  (`.claude/rules/go-debugging.md`). One writer per file, as before; the
  ownership rule that needed one pipe per emitting goroutine is now satisfied
  by one file per emitting goroutine, and a file cannot interleave two
  writers' half-written headers the way a shared fd could.
- **Go → TS is block files — no sidecar.** A node's
  kind is a numeric `kindId` value (TS maps it to `NODE_DEFS` colors), its
  label is the `label` section of its own file, and its identity is the ROW,
  which is the directory its file sits in (Go resolves row → node for hits).
  The row is therefore an ADDRESS, not a claim riding beside the data, so it
  cannot disagree with where the data arrived. Reflect resources
  (`overlay-flags.ts`, `scene-leaves.ts`) mirror Go — they author nothing.
  There is **no JSON-trace render path and no `pump.ts`**; Go emits no
  trace-event JSON on stdout at all. `scripts/probe-merge.sh` decodes the
  binary trace files at READ time through the owner-specific `readtrace` (see scripts/probe-merge.sh).
- **`SceneRoot`** (`Start/extension/webview/scene/scene-root.tsx`)
  is the composition root of the render tree — it assembles the per-concern
  components, each of which reads its own block files. It is a
  small file; the drawing lives in its siblings under `three/scene/`. Grep the symbol,
  not this filename. The tree covers: node bodies (`Categories/Ring/NodeShape/NodeInstances.tsx` — sphere
  mesh + ring, keyed off `node.data.fill`/`node.data.stroke` from `NODE_DEFS`; no port
  geometry — a port is a load-time channel-binding ROLE, never drawn),
  transit and interior
  beads (`Categories/Node/BeadAnimation/ChainBeadInstances.tsx`, `Categories/Node/Interior/InteriorBeadInstances.tsx` — there is no
  per-edge drawn tube any more; the source node's own chain of placeholder beads is the
  edge's visual, `docs/model/entities.md`), the selection ring, its halo and the hover
  ring (placed with everything else drawn at a node's own frame, in
  `Categories/Ring/NodeShape/node-instances-update.ts`; their shape lives in
  `Categories/Ring/NodeShape/node-highlight-shape.ts`), and the camera (`Categories/Scene/Camera/SceneCamera.tsx` maps the camera
  block file onto the three.js camera). Nothing in this tree owns traversal
  timing, positions, or geometry.
- **Bridge surface — binary BOTH ways.** **Go → TS:** the block files, plus
  the trace-event streams over each goroutine's own dedicated inherited-stdio
  pipe — stated in full under "Go → TS is block files plus the event streams"
  above; not restated here, so the two copies cannot drift apart. **TS → Go:** framed binary records on stdin (`[len:u32-LE][record]`,
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

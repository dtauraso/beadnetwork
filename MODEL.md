# Model

Read this before changing anything in the **Go network** (`nodes/`,
`nodes/wire/paced_wire.go`, `nodes/Wiring/loader.go`,
`nodes/Wiring/builders.go`) or anything that schedules/orders work. If
your reasoning slips into retired vocabulary, you are in the wrong
frame. Stop, re-read this file, and re-derive from the model.

## The network

The network is **nodes and wires**. Each node runs on its own Go
goroutine. A wire (`PacedWire`) does NOT: it is a PASSIVE delay queue
with a channel on each end — a channel in from its source node, a channel
out to its destination node — stepped by its SOURCE NODE's own goroutine.
The wire still owns its own beads (`inflight`/`delivered`) and its own
geometry as data, and exactly one goroutine touches that state: the source
node's. Nothing else locks or reaches into it. Historically the wire had
its OWN goroutine
(`PacedWire.run`, one per wire, launched by `Start`) — not stepped by
another. The wire's own goroutine is the sole thing that touches `inflight`. **An input port is one wire fed by
exactly one edge** — fan-in (several edges into one port) is not part of
the model; multiple sources into one node use distinct input ports (see
§Node lifecycle). So a wire has exactly one incident edge, and its own
goroutine owns its beads with no sharing. The
network is self-scheduling: there is no central runner, no walker, no
underlying layer that "runs" the nodes. The network IS the running program.

A channel here does NOT mean a blocking, backpressured handshake — see
§Sending below. The source places a bead on its out-channel and moves
on; it never waits on the wire or the destination.

Behavior emerges from wiring — the topology is the logic.

The visual editor is the medium for authoring and observing the network;
the network itself is the nodes-and-wires Go runtime.

## What things are

- **Bead.** A value in transit from a source node to a destination node.
- **Wire (`PacedWire`).** Transport. A PASSIVE delay queue, not a
  goroutine: the source node sends a bead over the wire's in-channel to
  place it, and that SAME source node times the traversal on its own clock
  reading (each goroutine owns its own clock copy — see the Clock bullet
  below) by driving the wire each cycle, then on traversal-complete sends
  the bead over its out-channel to the destination. The wire is no longer
  the visual depiction either — the source node's own chain of placeholder
  beads is (docs/beads-are-the-edge.md). There is one owner
  of `inflight`/`delivered` and the in-flight geometry: the source node
  goroutine. Because it is the sole owner, `PacedWire.mu` does not exist
  — ownership replaces locking, the same move that removed `RealClock.mu`.
  Do not reintroduce a lock here "for safety"; a second lock on top of
  single-goroutine ownership is dead weight, and if two goroutines ever
  need to touch this state again that is a sign the ownership model
  broke, not a reason to add a mutex. The wire applies no send policy —
  see §Sending.
- **Node goroutine.** Receives beads over its input port's channel,
  holds them in node-local state until its firing rule is satisfied,
  then fires. There is no held-value slot in this model sense — node-local held
  state replaces it. (This is a different concept from the buffer's `Slot`
  column — `nodes/wire/owner_events.go`, `Buffer/stream_events.go`,
  `Buffer/layout.go` — which is a live 2x2 interior VISUAL grid position,
  slot = gridRow*2 + gridCol, for where a held bead is drawn inside a node.)
- **Input port.** A ROLE, not a place (`docs/channels-not-ports.md`): declared by the
  node kind as a `Wiring.PortSpec` and bound to a channel at LOAD time
  (`a.In(...)`), never drawn and never hit-testable. One input port is one wire,
  and the wire's out-channel is the connection between them — the node receives
  whatever the source node's drive of that wire sends. Ports carry no geometry of their
  own; an edge attaches at its two nodes' SURFACES (`docs/bead-lattice.md`), not at a
  port position.
- **Clock (the human-speed clock).** There is exactly one clock: the system monotonic clock, read through a **scale** so it advances in integer **ticks** at human-watchable speed (`tick = ⌊(now − start) / tickPeriod⌋`; the scale is the human-speed / playback-speed knob, `MsPerTick = 16` ⇒ ≈62.5 ticks/sec). All timing is **tick counts**, not wall-clock durations. The model is **sleep-only**: a pacing loop calls `SleepCycle` to wait exactly ONE cycle and re-reads `Tick()`, rather than blocking on a target tick — there is no wait-until-tick-k primitive. The clock is **free-running**: it advances monotonically with wall time and never pauses (there is no play/pause gate). **Everything that animates runs in these ticks:** bead traveling, all in-node animations, and all node/gate processing windows. Per-update tick counts come from formulas, not literals — a bead crossing an edge takes `ticksToCross = steps * DwellTicksPerBead` (steps the edge's own bead-step count, `DwellTicksPerBead` a uniform constant per bead-lattice step across all wires — [docs/bead-lattice.md](docs/bead-lattice.md)); node processing windows are tick counts. There is no separate render cadence — the tick IS the animation clock.
  A wire is stepped with its SOURCE NODE's own clock copy and tick reading,
  exactly like every other per-goroutine clock use — there is no shared
  clock to pin a tick against. But a bead's **placement tick** (when it
  started crossing) is a DIFFERENT reading than the step tick that later
  advances it: placement is decided by the **emitting** goroutine, at the
  moment it calls `Send`, from its own clock, once per emission — not
  re-derived later by whichever goroutine happens to drain the wire's
  in-channel. This is what lets several beads placed in one emission (a
  broadcast fan-out) provably share one placement tick: reading a fresh
  clock value per wire in the drain pass can straddle a tick boundary
  between two beads placed microseconds apart, splitting one emission
  across two ticks (the observed bug this fixed). Read once, stamp
  everywhere, in the same call.

## Wire lifecycle

A bead crosses a wire in one direction:

1. The source node sends the bead over the wire's in-channel with its
   traversal timed in ticks: `ticksToCross = steps * DwellTicksPerBead`
   (steps the edge's own bead-step count — [docs/bead-lattice.md](docs/bead-lattice.md)). The
   send does not block on the wire and does not wait for the destination
   — see §Sending.
2. While in flight, the SOURCE NODE — reading its own clock, its own tick
   (see the Clock bullet above) — advances the bead, keeping it in the
   wire's `inflight` set. The wire has no goroutine of its own: the source
   node's mover drives `DriveOneCycle` for each of its outgoing wires
   (`nodeMover.run`), so a wire is a passive delay queue its source node
   steps.
3. On traversal-complete, that same drive sends the bead over the wire's
   out-channel to the destination node.
4. The destination node receives it and holds it in node-local state.

What the RENDERER is told is not a moving position. Each node owns a chain
of fixed placeholder beads per outgoing edge (`chainBeads`, node-local
offsets), and the animation is which bead is LIT: the node reads its own
wires' in-flight fraction `t` — the same `t` step 2 above advances — and
lights `index = t × count`. The chain is the visual of a traversal; it is
never a picture of the node-to-node channels, which are the real connection
and are never drawn. See [docs/beads-are-the-edge.md](docs/beads-are-the-edge.md).

The source node times its own delivery. There is no TS-driven delivery
signal — the renderer is told which bead is lit, not asked when a bead has
arrived.

## Node lifecycle

- A destination node receives beads over its input port's channel and
  holds them in node-local state.
- When the node's firing rule is satisfied, it fires.
- **One edge per input port.** An input port is a channel-binding ROLE (see above,
  `docs/channels-not-ports.md`), one wire, fed by EXACTLY ONE edge — fan-in (several
  edges into one port) is not part of the model.
  A node that needs several sources uses DISTINCT input ports, each its own
  wire (e.g. a gate's separate left/right ports). Multiple beads may still
  be in flight on one wire at once — its single source emitting repeatedly —
  each carrying its OWN placement geometry: geometry travels WITH the bead,
  never stored as one shared value on the wire, so the wire reshaping
  mid-flight (a node drag) re-derives each bead correctly. The loader
  rejects a fan-in topology at parse (`validateNoFanIn`); the guard
  `tools/check-no-fan-in.sh` keeps it out of the committed diagram.

## Sending

A node places a bead on its outgoing wire whenever its own rule says to, by sending it over the wire's in-channel. It does not check the wire's state and does not wait on the destination — there is no clear/busy state, no acknowledgment, no back-pressure. A channel here is NOT a blocking handshake: the send must never be able to stall the source node waiting on the wire or the destination, so it is sized/shaped so the source's send always succeeds immediately. A wire may carry more than one bead at once, each its own value; it transports whatever the source emits, and the destination reads whatever arrives. Coordination between nodes is the topology and each node's local rule, not a delivery guarantee between the two.

Nodes time their processing in **ticks**: a firing rule may span a
**tick-count window** (e.g. a gate's inhibit/processing window is K ticks),
paced against the human-speed clock by sleeping one cycle at a time
(`SleepCycle`) and comparing `Tick()`. Firing is still
gated on held state — now optionally held across a tick window rather than
resolving instantaneously.

## Geometry and time

- Wire geometry sets traversal in ticks:
  `ticksToCross = steps * DwellTicksPerBead` (steps: docs/bead-lattice.md "The
  count", computed by the SOURCE NODE from its own stored `LocalPolar` to the
  target — one integer, not an arc length divided by a speed). Geometry has
  no other effect on timing.
- A geometry edit re-derives traversal time. While a bead is in flight,
  the in-flight revision PRESERVES the bead's FRACTIONAL progress `t` (its
  proportion along the wire) — NOT the absolute distance covered. On the
  edit the bead stays at the same fraction `t`, and the remaining ticks are
  recomputed from the NEW step count at the uniform per-step dwell:
  `remainingTicks = (1−t)·newSteps·DwellTicksPerBead`. So the bead rides
  smoothly at the same proportion as the wire reshapes (no t-swing race as a
  node is dragged), and a longer or shorter wire still traverses at constant
  world-speed. (Preserving distance instead would let `t` jump as the step
  count changes.)
- Go owns the bead's PROGRESS (the fraction `t`, timed in ticks on the human-speed clock).
  It no longer computes or streams an absolute bead position: nothing draws a moving bead.
  The source node quantises its own `t` onto its own chain and streams which bead is LIT
  (`readChainBeadLit` from `tools/topology-vscode/src/schema/buffer-layout.ts`, consumed by
  `tools/topology-vscode/src/webview/three/ChainBeadInstances.tsx`). The editor does not
  interpolate, does not own positions, and does not decide which bead is lit.
- Durations are tick counts: bead traversal (`ticksToCross`) and node processing windows.
## Driver

**Self-scheduling node goroutines.** Each node is a goroutine. A WIRE IS
NOT: it is a passive delay queue, and the goroutine that steps it is its
own SOURCE NODE's mover (`nodeMover.run` calls `DriveOneCycle` for each of
its outgoing wires). There is still no central walker and no play/pause
gate — the clock is free-running and the animation never halts.

Each SOURCE NODE times its own outgoing deliveries on its own reading of
the human-speed clock: when a bead's `ticksToCross` have elapsed, that
drive sends it over the wire's out-channel to the destination node, which
receives it. Delivery is not triggered by the renderer — there is no
cross-boundary delivery signal. The editor is told which chain bead is lit;
it is never asked when a bead has arrived.

(This replaced a per-wire goroutine. There was exactly ONE place a wire was
ever driven — `edgeMover.run` — so the change was that call moving to the
node, not a new mechanism. Driving it there is also what lets a node read
its own wires' in-flight fraction to light its own chain without touching
another goroutine's state.)

There is one tick clock (the human-speed clock) but no lockstep round or
simultaneity layer: goroutines schedule independently against the shared
tick, each sleeping a cycle (`SleepCycle`) and re-reading `Tick()` on its
own — they are not aligned into global rounds, and the
network does not count rounds. Coordination between nodes happens through
the values nodes place on wires and the topology — not through
round-alignment or a delivery handshake. Any reasoning that treats
activity as a sequence of globally-aligned lockstep rounds is drift;
re-derive from local rules that wait on ticks over channels and wires.

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
  stdio pipe (`Buffer/stream_fds.go`, memory/feedback_no_single_writer_bridge.md)
  — one VIEW stream (camera/overlay/scene, the gesture/stdin-reader
  goroutine), one stream per edge row (that edgeMover's own geometry + its
  wire's live beads), and two streams per node row (that nodeMover's own
  geometry+ports+label, and that node's own Update-goroutine's interior
  beads). Frames on a dedicated fd are `[len:u32-LE][payload]` with NO tag
  byte — the fd POSITION identifies the stream/row.
  `WIREFOLD_STREAM_FDS` (the ext host's spawn env var,
  `tools/topology-vscode/src/runCommand.ts`) is **mandatory**: there is no
  central accumulator and no fallback path left to fall back to.
- **Go → TS is binary content buffers** (`buffer-snapshot`) ALONE — no
  sidecar. Each node's kind is a numeric `KindId` column (TS maps it to
  `NODE_DEFS` colors), its label rides its own stream frame's inline label
  bytes, and its identity is the buffer ROW INDEX (Go resolves row → node
  for hits). The ext host relays each dedicated-fd frame to the webview
  under a synthetic tag (`BUF_BLOCK_TAG_VIEW`/`_EDGE_STREAM`/`_NODE_STREAM`/
  `_INTERIOR_STREAM`, `Buffer/frame_tags.go`) purely for cell routing —
  never a wire byte. The webview decodes each stream (`buffer-decode.ts`)
  and renders it; row-keyed reflect resources (`snapshot-buffer.ts`,
  `overlay-flags.ts`) mirror Go — they author nothing. There is **no
  JSON-trace render path and no `pump.ts`**; Go emits no trace-event JSON
  on stdout at all — the `.probe` trace logs (`go.jsonl`/`go-node.jsonl`/
  `go-edge.jsonl`/`go-interior.jsonl`) are the ext host's DECODE of each
  per-owner stream's own trailing EVENTS section (`buffer-log.ts`), not a
  stdout parse. Stdout carries only the DEBUG BREADCRUMB channel's sparse
  `{"kind":"breadcrumb",...}` control-event lines.
- **`BufferScene`** (`tools/topology-vscode/src/webview/three/buffer-scene.tsx`)
  is the composition root of the render tree — it decodes the buffer and
  assembles the per-concern components that draw ALL geometry from it. It is a
  small file; the drawing lives in its siblings under `three/`. Grep the symbol,
  not this filename. The tree covers: node bodies (`tools/topology-vscode/src/webview/three/NodeInstances.tsx` — sphere
  mesh + ring, keyed off `node.data.fill`/`node.data.stroke` from `NODE_DEFS`; no port
  geometry — a port is a load-time channel-binding ROLE, never drawn, `docs/channels-not-ports.md`),
  edge tubes (`tools/topology-vscode/src/webview/three/EdgeTube.tsx`), transit and interior
  beads (`tools/topology-vscode/src/webview/three/ChainBeadInstances.tsx`, `tools/topology-vscode/src/webview/three/InteriorBeadInstances.tsx`), selection highlight
  (`tools/topology-vscode/src/webview/three/SelectionHighlight.tsx`), and the camera (`tools/topology-vscode/src/webview/three/BufferCamera.tsx` maps the buffer
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

## Drift rule

Traversal-timing or firing-rule logic outside the Go node and wire
goroutines is drift — move it back into Go. Likewise any domain state
(node/edge/pulse/geometry/camera/selection) authored on the TS side, or
any TS-side geometry/position/timing computation, is drift: Go owns the
model and streams it as the content buffer; TS decodes and draws.

## Node positions & movement locks (the polar model)

Editor-time node geometry and lock propagation are **pure polar**. The scene sphere's center
is the only cartesian value that is **persisted and authoritative** — the world anchor. It is
not the only cartesian value that exists: the camera pose, port anchors, bead segments and
per-node world centers are cartesian too. The invariant is narrower and stronger than
"cartesian appears once": every other cartesian is **derived** from this anchor
(`sceneCenter + polar2cart(…)`) or **quarantined at the renderer edge** — none is persisted,
and none is a source of truth.

- **Scene sphere** — a first-class, persisted reference (NOT the derived content-sphere
  centroid, which moves with the nodes and is circular). It has a **cartesian center** (the
  one world anchor, in `scene.json`) and a **radius** that fits the diagram.

  **It is established once and never moves.** `LoadSceneSphere` reads it from `scene.json`,
  or content-fits it from the node centers and persists that immediately — persisting matters,
  because every node position is a polar measured ABOUT this center, so a center re-fitted
  over moved nodes on the next load would silently reinterpret the whole diagram.

  It is a SEPARATE entity from the camera pivot, and **both camera gestures leave it alone**:
  orbit must not move it, and `PanViewpoint` is a pure CAMERA move that deliberately does not
  touch it.

  > **Rejected: pan moves the sphere.** The model once said camera pan should translate the
  > center by the same delta, holding node world positions fixed while their scene polars
  > recomputed about the new center. Coupling pan/dolly to the sphere left `md.sceneSphere`
  > diverged from the movers' held center until a later broadcast reconciled it with a jump —
  > the "zoom got canceled" symptom. It was decoupled, a separate scene-pan gesture was named
  > as the proper home, and that gesture was never built. The claim is now DROPPED rather than
  > pending: the sphere is a load-time-fitted constant.
  >
  > The cost, stated so it is a choice and not an accident: the polar frame is best
  > conditioned near the center it was fitted to. Pan the camera far away and drag there, and
  > you work at large `r`, where a small angular step moves a node a long way. If that becomes
  > real friction in the editor, THAT is the reason to revisit — and the trap to avoid is the
  > one above: a scene pan is its own gesture, never a side effect of a camera move.
- **Two polars per node.** (1) **Scene polar** `(r,θ,φ)` about the scene-sphere center — the
  node's POSITION, persisted (`meta.json` `scenePolar*`; cartesian `x/y/z` kept only for
  back-compat, and only used at load when no sphere is persisted yet). World = `sceneCenter +
  polar2cart(scenePolar)`. (2) **Local polar** — the node's stored OFFSET about the local
  node it is doubly-linked to (its lock center). This is the constraint frame.
- **Locks are offsets.** A node-node lock nudges ONE component of a node's **stored local
  polar** offset (a bounded copy of a neighbor's owned component), carried node-to-node in the
  decentralized cascade message (`sendMove`, node-to-node over per-pair directed channels — there is no
  central worklist and no shared inbox). The offset lives on the node's `LayoutHolder`
  (`SetLocalPolar` / `LocalPolarsSnapshot`), seeded from the existing movement-link value by
  `computeLocalPolars` at load and re-quantized on move by `requantizeLocalPolars` — there is
  no separate seed step. Each `nodeMover` owns one dedicated inbound channel per adjacent
  node (`neighborIn`, keyed by sender id) plus one `extIn` channel for external entries
  (drag/gesture); a send is a non-blocking attempt on the destination's own channel, and on a
  full channel the message is RETAINED on the sender's own `pending` queue and retried next
  cycle (`sendMove` / `flushPending`) — there is no dedicated sender goroutine and no lock.
  This preserves FIFO order per destination and the cascade never drops a message. (An
  earlier direct-to-inbox lossy sender dropped ~98% of sends under load and was removed; a
  later shared-`outbox`-plus-sender-goroutine design closed that hole but added two
  goroutines and unbounded queueing per mover, and was itself replaced by today's
  per-direction channels with bounded retry.)
- **No blow-up, by construction.** The offset is STORED and only carried through the
  composition or nudged one component — it is NEVER re-derived as `cart2polar(node − center)`
  from a live world during a cascade. That reconstruction against a mid-moving center is the
  bug that made positions fly to infinity. A moved center rigidly translates its satellites
  (offset unchanged ⇒ locks stay satisfied ⇒ the wave terminates). This is STRUCTURAL, not a
  test: the reconstruction that caused the blow-up has no call site to write. Nav is held
  polar-only by `tools/check-polar-only-nav.sh`.
- **Panel-authored locks must be structurally incapable of a position blow-up.** If one
  happens, the implementation is wrong (an offset was reconstructed from a moving reference),
  not the locks.
- **Moving a node is CRUD on the edge beads that touch it (drag placement).** N chains
  connect to a node; you move the node by removing links from those chains or adding links
  to them — that is the whole mechanism. There is no solver, no constraint system, no
  enumeration across neighbours: each touching bead decides for itself
  (`nodes/Wiring/bead_crud.go`'s `beadCrudDecide`, wired in `commitNodeMoveLocal`,
  `nodes/Wiring/quantized_move.go`).

  The drag gives the node's own polar vector `v` (its previous position to its
  destination). Each touching bead has its own **source point** — the previous bead's
  centre along its chain, or the chain origin on the neighbour's torus surface when it is
  the only bead (`nodes/Wiring/quantized_move.go`'s `dragTouchingBeads`) — NEVER the
  touching bead's own centre; using the centre instead is wrong by one bead. The **third
  polar vector** runs from the bead's source point to the node's destination point.
  Compare its length to one bead length (`wire.BeadStepR`):

  - too small → that bead is **removed**, and the bead before it becomes the touching bead.
  - too large → a bead is **added** (subject to the angle gate below), and it becomes the
    new touching bead.
  - exactly one bead length → nothing changes.

  **The angle gate applies to ADD only, never to REMOVE.** The angle between `v` and the
  edge-bead vector (source → the touching bead's own centre): > 90 degrees blocks the add
  (the node did not move far enough AWAY from the bead to open a gap beyond it — an obtuse
  angle means the drag is heading back across the bead); ≤ 90 degrees admits it, subject to
  the `|third|` test above. A removal is decided by `|third|` alone.

  **There is no selection and no summation.** Every touching bead performs the same
  judgement against the same `v`; per-bead verdicts are never combined into a displacement.
  If at least one touching bead's verdict is not "none", the node moves exactly ONE bead
  length toward its destination — never further, and never the raw drag target itself; with
  no touching beads at all (a free node with no incident edges) the raw target is used
  directly. One drag event can therefore remove beads from some edges and add them to
  others at once, with nothing solved and nothing added together.

  Bead count on an edge falls out of the resulting geometry as one integer subtraction
  (`nodes/Wiring/chain_beads.go`'s `edgeStepCount`), with the near end tangent to the node's
  own torus by construction of the placement formula and one uniform global bead size — see
  `docs/bead-lattice.md`.

## Assertions

A `panic` in `nodes/`, `Buffer/`, or `Trace/` is an **assertion**, not error handling. It
fires only via a code bug — never via ordinary traffic, malformed input, or load. Input the
network cannot trust is rejected at parse (`validateNoFanIn`, `validateSpec`); by the time a
value reaches a goroutine's own state, a violated bound means the code is wrong.

So an assertion states an invariant the structure is supposed to guarantee, and the right
response to one firing is to fix the code, not to widen the bound.

**Panic rather than drop, grow, or block.** Dropping hides the broken drain the bound exists
to catch. Growing defers the same crash to a worse place. Blocking couples this goroutine's
pacing to another's, which is the coupling the whole model exists to avoid.

**The message is the whole value.** It is read exactly once, by whoever is debugging, and it
is the only context they get. It must:

1. **Open with a site tag** — the detecting function, method, or subsystem, then a colon —
   so the message greps back to its source: `paced_wire: `, `nodeMover(%s): `,
   `BuildEdgeStreamFrame: `.
2. **Name the invariant and the actual values**, not a category. `pending exceeded %d events
   on wire -> %s.%s`, not `limit exceeded`.
3. **Name the mechanism that should have prevented it.** This is what turns a crash into a
   diagnosis: *"the per-cycle drain (edgeMover.writeStreamFrame -> DrainPendingEvents) is not
   running"*, or `allocateWires`' *"validateNoFanIn should have rejected this fan-in at
   parse"* — which names the earlier gate that let it through.

**No `recover()` in the network.** Swallowing an assertion converts a loud, located failure
into a silent wrong answer.

Guard: `tools/check-panic-message.sh` (site tag + substance + no `recover()`). It enforces
the shape, not the content — (3) is the part only a human can write, and the part that pays.

## Allowed vocabulary

- bead, in-flight, held (node-local) state
- channel, input port, output port (one edge per input port — no fan-in)
- arc length, pulse speed (world-units per tick), ticks-to-cross,
  tick-count processing window
- tick, human-speed clock (the one system monotonic clock scaled to ticks
  at human speed), scale, `SleepCycle`, `Tick`
- node receives, node holds, node fires, wire advances, wire delivers,
  wire emits position

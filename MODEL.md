# Model

Read this before changing anything in the **Go network** (`src/Node/`,
`src/Bead/bead_run.go`, `src/Node/Wiring/build/loader.go`,
`src/Node/Wiring/loadspec/builders.go`) or anything that schedules/orders work. If
your reasoning slips into retired vocabulary, you are in the wrong
frame. Stop, re-read this file, and re-derive from the model.

## The network

The network is **nodes and wires**. A node id NAMES a thing that is drawn.
**A node is FOUR goroutines, paced by four different things**, because it does
four jobs that have four different clocks:

- The **kind goroutine** — the KIND's own `Update` — is paced by the sim
  clock. It owns the kind's logic and its interior slots. It writes its own
  interior stream. It runs `Self.Step(ctx, tick)` once per pass of its own loop,
  where `Self` is the `PairNodeSelf` it claimed at build time
  (`BuildArgs.ClaimSelfDrive`); a kind that holds a value onto an out steps a
  `gatecommon.HeldDriver` in that same pass rather than handing the value to a
  goroutine over a channel.
- The **animation goroutine** — `owners.Outs.RunAnimation`, one per node id with
  outputs — is paced by the pulse, not the sim clock. It owns the node's beads
  and writes its own bead stream. It is NOT the kind's loop: beads used to move
  only when the node's sim cycle came round, which tied how smoothly a bead drew
  to how fast the network was set to run.
- The **rule goroutine** — `rulenode.RuleNode.Run` — blocks on its own inbox. It
  fans out further, one forwarder per peer in the all-pairs rule mesh, so the
  goroutine count per node grows with the size of the scene.
- The **geometry goroutine** — `PairNodeSelf.RunGeometry`, one per node id — is
  paced by nothing. It BLOCKS on its own inbox and runs when a message arrives,
  so a drag is served at the rate of the hand that is dragging. It owns the
  node's own polar position, its stored vectors to its neighbours
  (`owners.Deltas`), interaction state, the rule mesh, and **its own OUT-EDGES**
  (each drawn from the node's own polar position composed with that edge's
  stored delta). It writes its own node stream and the edge stream of every
  edge leaving it.

**This is deliberately NOT one clock per node.** It used to be, and that is
what this split reverses: geometry work was done in the kind's paced loop, so a
node accepted at most one move per sim cycle — up to ~1 s at low speed — while
a pointer produces sixty. Dragging a node reliably killed it on a pending-send
bound, correctly reporting that it was "enqueueing to a peer faster than that
peer drains". A hand movement must not be paced by a simulated
one.

**The goroutines share no memory.** Every field of `NodeGeometry` belongs to
exactly one of them, and the places the geometry half used to reach into the
animation half are messages: the per-edge segment/step-count revision that
`OutEdges.DeriveGeometry` derives, handed over by `BeadRun.PostGeom` and
`outport.Out.PostGeom` as a `bead.Revision`, and bead-drag start/end, handed
over by `Beads.PostBeadDrag`. The animation goroutine applies both at ONE point, the top
of its own pass, never mid-pass — so beads are never drawn against a segment
that moved halfway through the frame. An atomic or a mutex here would be a
defect, not a fix (`check-no-network-locks.sh`).

**An edge has NO goroutine at all** — not for drawing
it and not for anything else. Its segment, its step count, the revision of
beads already in flight on it, and its persisted file under the source
node's own directory are all derived by the SOURCE NODE, in that node's
geometry loop, from the node's own polar position composed with the stored
vector to that neighbour (`owners.Deltas`, `OutEdges.DeriveGeometry`). Nothing
holds a copy of the far end's absolute position: when a neighbour moves, the
node is told how far it moved and shifts its own stored vector by that much, so
the vector is maintained by composition and never re-derived. An
`edgetable.Edge` is a plain record of endpoints and plumbing, not an actor.
There is **no goroutine per edge**, and no separate driver goroutine
placing a held value onto an out. A wire
(`BeadRun`) is not a goroutine at all: it is a PASSIVE delay queue
with a channel on each end — a channel in from its source node, a channel
out to its destination node — stepped by its SOURCE NODE's ANIMATION goroutine.
The wire still owns its own beads (`inflight`/`delivered`) and its own
geometry as data, and exactly one goroutine touches that state: the source
node's ANIMATION goroutine. The source node's geometry goroutine DERIVES the
wire's new segment and step count but never writes them — it sends them, and
the animation goroutine applies them. Nothing locks or reaches into the wire.
Historically the wire had
its OWN goroutine
(`BeadRun.run`, one per wire, launched by `Start`) — not stepped by
another. **An input port is one wire fed by
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

See [docs/model/entities.md](docs/model/entities.md) for the full statement of the
in-flight value bead vs. the chain (render/placeholder) bead vs. the `BeadRun`, the node
goroutine, the input port, and the human-speed clock.

## Wire lifecycle, node lifecycle, and sending

See [docs/model/lifecycle.md](docs/model/lifecycle.md) for how a bead crosses a wire, what
the renderer is told, node lifecycle (including the one-edge-per-input-port rule), and the
send contract.

## Geometry, time, and the driver

See [docs/model/timing.md](docs/model/timing.md) for how wire geometry sets traversal
slots, why a bead in flight ignores a drag, and the
self-scheduling driver (each source node's mover, no central walker, no lockstep rounds).

## Editor surface (TS)

The model lives entirely in Go. The TS/React layer is **render + forward
only**: it decodes the binary content buffer Go streams and draws it, and
forwards raw input to Go. It holds NO domain state — no render stores, no
spec store, no camera store — never sets node state and never tells Go
when a bead has arrived. Go owns the clock.

See [docs/model/editor-surface.md](docs/model/editor-surface.md) for what the Go runtime
owns and streams, the Go → TS binary content buffer, `BufferScene`'s render tree, and the
binary-both-ways bridge surface.

## Scenes

See [docs/model/scenes.md](docs/model/scenes.md) for what a scene is, how a tab switch
works, and the per-scene fork of node behaviour it documents (coplanar rings), plus the
ring-axis vs. navigation-pole distinction. `scene.Scene`'s other per-scene fields
(`UpAxis`, `ClockDivisor`, `Editable`, `Kinds`) are not written up
anywhere — read `src/Node/Wiring/scene/scene.go`.

## Drift rule

Traversal-timing or firing-rule logic outside the Go node and wire
goroutines is drift — move it back into Go. Likewise any domain state
(node/edge/pulse/geometry/camera/selection) authored on the TS side, or
any TS-side geometry/position/timing computation, is drift: Go owns the
model and streams it as the content buffer; TS decodes and draws.

## Node positions & movement locks (the polar model)

Editor-time node geometry and lock propagation are **pure polar**. The scene sphere's center
is the only cartesian value that is **persisted and authoritative** — the world anchor.

A node is a **point** and an edge is the **vector that closes the triangle**:

```
A = scene centre -> node        D = node -> neighbour        B = scene centre -> neighbour
A + D = B
```

**`D` starts AT THE NODE.** That is the whole of it: because the vector starts there, its
`phi` IS the angle from that node's own +y pole to its neighbour — the quantity the
out-angle constraint names — and the constraint is read off the number it is applied to.

`D` is **not** `B − A` component-wise. A difference of two coordinates that both start at
the scene centre is a different quantity, and clamping it holds a number that is not the
angle in the triangle: it sat at exactly 90.0000° while the picture sat at 99.79°, always
past π/2, and no constant closes that gap because the real angle also depends on how far out
the node sits and where its neighbour lies around the pole. Composition (`polar.Compose`)
and difference (`polar.Between`) resolve onto common axes — the one operation the three
numbers cannot do separately. Negation stays arithmetic and is exact.

A move is the same triangle: a node's move takes `Δ` off every side it touches, and the node
at the other end is TOLD `Δ` and takes it onto its own side. Neither reads the other's
position.

See [docs/model/polar-model.md](docs/model/polar-model.md) for the full statement: the
scene sphere, the one-hop scene→node→bead vector sum, the triangle above and why it is not
the rejected double-link, why there is no blow-up by construction, and the bead-CRUD
drag-placement mechanism (add/remove/angle-gate) that moves a node.

## Tori (node rings and beads): one canonical surface, Go-composed matrices

A torus is drawn from a **polar parametric equation Go evaluates ONCE**, plus **one instance
matrix per torus that Go composes**. TS uploads both and draws. It generates no surface,
composes no transform, and holds no tube ratio.

The surface is the unit torus at `rho = 1`, tube `a` = the kind's tube ratio, on the
`theta == 0` disk (`src/Node/Wiring/framegeom/ring_surface.go`):

```
w   = rho + a*cos(v)
R   = sqrt(rho^2 + a^2 + 2*rho*a*cos(v))            # v only
psi = atan2(a*sin(v), w)                            # v only
phi   = atan2( sqrt(cos(psi)^2*sin(u)^2 + sin(psi)^2), cos(psi)*cos(u) )
theta = atan2( sin(psi), cos(psi)*sin(u) )
```

`u` sweeps the ring, `v` the tube, both as `index * step` — never re-derived from a position,
because `Cart2polar` returns `phi` in `[0, pi]` and would silently fold the ring's second
half. `R` and `psi` depend on `v` alone: that is the ring's rotational symmetry, and it is why
the equation reduces to the two rim constraints — `theta == 0, r == rho +/- a` at `v = 0, pi`.

**Why once, and not per instance.** Every torus of a given tube ratio is the SAME SHAPE;
only centre, radius and axis vary. Streaming a surface per instance streams that shape N
times and destroys instancing — measured at ~3 KB per instance against 64 bytes for a matrix,
which is ~1.5 MB per refresh once beads are included versus ~32 KB. A per-instance surface is
therefore drift, not an optimisation choice.

**Why the instance basis has no convention.** The matrix rotates `+Z` onto the torus's axis
through an arbitrary-but-consistent orthonormal frame. Any frame with `bz == axis` is
correct, because the canonical torus is FULLY REVOLVED about that axis — the free choice
cannot show in the output. Do not add an "up" convention to constrain it; there is nothing
for one to fix.

## Assertions

A `panic` in `src/Node/`, `src/schema/buffer-layout/`, or `src/Trace/` is an **assertion**, not error handling. It
fires only via a code bug — never via ordinary traffic, malformed input, or load. Input the
network cannot trust is rejected at parse (`validateNoFanIn`, `ValidateSpec`); by the time a
value reaches a goroutine's own state, a violated bound means the code is wrong.

So an assertion states an invariant the structure is supposed to guarantee, and the right
response to one firing is to fix the code, not to widen the bound.

**Panic rather than drop, grow, or block.** Dropping hides the broken drain the bound exists
to catch. Growing defers the same crash to a worse place. Blocking couples this goroutine's
pacing to another's, which is the coupling the whole model exists to avoid.

**The message is the whole value.** It is read exactly once, by whoever is debugging, and it
is the only context they get. It must:

1. **Open with a site tag** — the detecting function, method, or subsystem, then a colon —
   so the message greps back to its source: `paced_wire: `, `interior.Mailbox.Send: `,
   `BuildEdgeStreamFrame: `.
2. **Name the invariant and the actual values**, not a category. `pending exceeded %d events
   on wire -> %s.%s`, not `limit exceeded`.
3. **Name the mechanism that should have prevented it.** This is what turns a crash into a
   diagnosis: *"the per-cycle drain (the source node's own Outs.DriveOutWires ->
   DrainPendingEvents) is not
   running"*, or `AllocateWires`' *"validateNoFanIn should have rejected this fan-in at
   parse"* — which names the earlier gate that let it through.

**No `recover()` in the network.** Swallowing an assertion converts a loud, located failure
into a silent wrong answer.

Guard: `src/Node/check-panic-message.sh` (site tag + substance + no `recover()`). It enforces
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

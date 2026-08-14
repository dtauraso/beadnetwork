# Model

Read this before changing anything in the **Go network** (`nodes/`,
`nodes/wire/paced_wire.go`, `nodes/Wiring/build/loader.go`,
`nodes/Wiring/loadspec/builders.go`) or anything that schedules/orders work. If
your reasoning slips into retired vocabulary, you are in the wrong
frame. Stop, re-read this file, and re-derive from the model.

## The network

The network is **nodes and wires**. A node id NAMES a thing that is drawn.
**A node is ONE goroutine** — the KIND's own `Update` — and that goroutine
owns, in one paced loop, everything the node is and everything it draws
about itself: the kind's logic and its interior slots, interaction and
geometry, its own beads, and **its own OUT-EDGES** (each drawn from the
node's own polar position composed with that edge's stored delta). It
writes its own node stream, its own bead stream, and the edge stream of
every edge leaving it. An edge is not drawn by anyone else: there is no
goroutine that watches two nodes and draws the line between them. There
is **no second goroutine per node id** — no mover goroutine beside `Update`,
and no separate driver goroutine placing a held value onto an out. `Update`
gets the drawing half by calling `Self.Step(ctx, tick)` once per pass of
its own loop, where `Self` is the `PairNodeSelf` it claimed at build time
(`BuildArgs.ClaimSelfDrive`); a kind that holds a value onto an out steps a
`gatecommon.HeldDriver` in that same pass rather than handing the value to
a goroutine over a channel. Because the node loop paces beads, it runs at
its own tick cadence, and the same loop reads input — so input is served at
that cadence too, deliberately: one clock per node, not one clock per job. A wire
(`PacedWire`) is not a goroutine at all: it is a PASSIVE delay queue
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

See [docs/model/entities.md](docs/model/entities.md) for the full statement of the
in-flight value bead vs. the chain (render/placeholder) bead vs. the `PacedWire`, the node
goroutine, the input port, and the human-speed clock.

## Wire lifecycle, node lifecycle, and sending

See [docs/model/lifecycle.md](docs/model/lifecycle.md) for how a bead crosses a wire, what
the renderer is told, node lifecycle (including the one-edge-per-input-port rule), and the
send contract.

## Geometry, time, and the driver

See [docs/model/timing.md](docs/model/timing.md) for how wire geometry sets traversal
ticks, how an in-flight geometry edit preserves fractional progress, and the
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
works, and the two named per-scene forks (drag mode, coplanar rings), plus the ring-axis
vs. navigation-pole distinction.

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
`theta == 0` disk (`nodes/Wiring/framegeom/ring_surface.go`):

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
   so the message greps back to its source: `paced_wire: `, `interior.Mailbox.Send: `,
   `BuildEdgeStreamFrame: `.
2. **Name the invariant and the actual values**, not a category. `pending exceeded %d events
   on wire -> %s.%s`, not `limit exceeded`.
3. **Name the mechanism that should have prevented it.** This is what turns a crash into a
   diagnosis: *"the per-cycle drain (the source node's own Outs.DriveOutWires ->
   DrainPendingEvents) is not
   running"*, or `allocateWires`' *"validateNoFanIn should have rejected this fan-in at
   parse"* — which names the earlier gate that let it through.

**No `recover()` in the network.** Swallowing an assertion converts a loud, located failure
into a silent wrong answer.

Guard: `tools/network/quality/check-panic-message.sh` (site tag + substance + no `recover()`). It enforces
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

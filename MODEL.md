# Model

Read this before changing anything in the **Go network** (`nodes/`,
`nodes/wire/paced_wire.go`, `nodes/Wiring/build/loader.go`,
`nodes/Wiring/loadspec/builders.go`) or anything that schedules/orders work. If
your reasoning slips into retired vocabulary, you are in the wrong
frame. Stop, re-read this file, and re-derive from the model.

## The network

The network is **nodes and wires**. A node id NAMES a thing that is drawn;
SEVERAL goroutines serve one node id, each owning one job and its own
state, each writing its own stream, all tagging their frames with that id
— the editor composes them onto one drawn node. They are peers: there is
no principal node goroutine the others assist. The jobs are what differ
(geometry and interaction, bead animation, the kind's own logic), and
they share nothing but the id. In particular **the human-speed clock
belongs to the ANIMATION job alone** — a goroutine that sleeps on the
clock to pace beads must not also be the one that reads input or streams
geometry, or the bead rate becomes the interaction rate. A wire
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

See [docs/model/polar-model.md](docs/model/polar-model.md) for the full statement: the
scene sphere, the one-hop scene→node→bead vector sum, a node's single quantised polar
coordinate, the per-edge first-bead vector, why there is no blow-up by construction, and
the bead-CRUD drag-placement mechanism (add/remove/angle-gate) that moves a node.

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
   so the message greps back to its source: `paced_wire: `, `NodeMover(%s): `,
   `BuildEdgeStreamFrame: `.
2. **Name the invariant and the actual values**, not a category. `pending exceeded %d events
   on wire -> %s.%s`, not `limit exceeded`.
3. **Name the mechanism that should have prevented it.** This is what turns a crash into a
   diagnosis: *"the per-cycle drain (edgeMover.writeStreamFrame -> DrainPendingEvents) is not
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

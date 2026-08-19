# Model — wire and node lifecycle

[← MODEL.md](../../MODEL.md)

## Wire lifecycle

A bead crosses a wire in one direction:

1. The source node places the bead on its own run, on the segment and step
   count that run currently holds. The send does not block and does not wait
   for the destination — see §Sending.
2. While in flight, the SOURCE NODE's own animation goroutine advances it one
   slot per wake, keeping it in the run's `inflight` set. The run has no
   goroutine of its own: `bead.Animation.RunAnimation` steps every run the
   node owns, so a run is a passive delay queue its source node steps. A bead
   captures its segment and step count when placed and keeps them, so a node
   moving mid-flight does not touch it.
3. When the bead reaches its last slot, that same step sends it over the run's
   out-channel to the destination node. Arrival IS the last slot; no clock is
   compared against a deadline.
4. The destination node receives it and holds it in node-local state.

What the RENDERER is told is the bead's position, computed in Go as
`slot × SlotR` along the segment and streamed in the EdgeBead block. The
editor draws what arrives and interpolates nothing. Each node also owns a
chain of fixed placeholder beads per outgoing edge (`chainBeads`, node-local
offsets); the chain is the visual of a traversal, and its length is
`edgegeom.EdgeStepCount`. It is never a picture of the
node-to-node channels, which are the real connection and are never drawn.

The source node times its own delivery. There is no TS-driven delivery
signal — the renderer is told which bead is lit, not asked when a bead has
arrived.

## Node lifecycle

- A destination node receives beads over its input port's channel and
  holds them in node-local state.
- When the node's firing rule is satisfied, it fires.
- **One edge per input port.** An input port is a channel-binding ROLE — a name in the
  kind's `[]portwiring.PortSpec` that `a.In("<name>")` binds a channel to at LOAD, never a
  Go field per channel — and it is one wire, fed by EXACTLY ONE edge. Fan-in (several
  edges into one port) is not part of the model.
  A node that needs several sources uses DISTINCT input ports, each its own
  wire (e.g. a gate's separate left/right ports). Multiple beads may still
  be in flight on one wire at once — its single source emitting repeatedly —
  each carrying its OWN placement geometry: geometry travels WITH the bead,
  never stored as one shared value on the wire, so the wire reshaping
  mid-flight (a node drag) re-derives each bead correctly. The loader
  rejects a fan-in topology at parse (`validateNoFanIn`); the guard
  `src/Node/Wiring/loadspec/check-no-fan-in.sh` keeps it out of the committed diagram.

## Sending

A node places a bead on its outgoing wire whenever its own rule says to, by sending it over the wire's in-channel. It does not check the wire's state and does not wait on the destination — there is no clear/busy state, no acknowledgment, no back-pressure. A channel here is NOT a blocking handshake: the send must never be able to stall the source node waiting on the wire or the destination, so it is sized/shaped so the source's send always succeeds immediately. A wire may carry more than one bead at once, each its own value; it transports whatever the source emits, and the destination reads whatever arrives. Coordination between nodes is the topology and each node's local rule, not a delivery guarantee between the two.

Nodes time their processing in **ticks**: a firing rule may span a
**tick-count window** (e.g. a gate's inhibit/processing window is K ticks),
paced against the human-speed clock by sleeping one cycle at a time
(`SleepCycle`) and comparing `Tick()`. Firing is still
gated on held state — now optionally held across a tick window rather than
resolving instantaneously.

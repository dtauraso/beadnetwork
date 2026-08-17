# Model — wire and node lifecycle

[← MODEL.md](../../MODEL.md)

## Wire lifecycle

A bead crosses a wire in one direction:

1. The source node sends the bead over the wire's in-channel with its
   traversal timed in ticks: `ticksToCross = steps * DwellTicksPerBead`
   (steps the edge's own bead-step count, `edgegeom.EdgeStepCount`). The
   send does not block on the wire and does not wait for the destination
   — see §Sending.
2. While in flight, the SOURCE NODE — reading its own clock, its own tick
   (see the Clock bullet above) — advances the bead, keeping it in the
   wire's `inflight` set. The wire has no goroutine of its own: the source
   node's mover drives `DriveOneCycle` for each of its outgoing wires
   (`NodeMover.Run`), so a wire is a passive delay queue its source node
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
and are never drawn. Its length is `edgegeom.EdgeStepCount`.

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
  `tools/network/concurrency/check-no-fan-in.sh` keeps it out of the committed diagram.

## Sending

A node places a bead on its outgoing wire whenever its own rule says to, by sending it over the wire's in-channel. It does not check the wire's state and does not wait on the destination — there is no clear/busy state, no acknowledgment, no back-pressure. A channel here is NOT a blocking handshake: the send must never be able to stall the source node waiting on the wire or the destination, so it is sized/shaped so the source's send always succeeds immediately. A wire may carry more than one bead at once, each its own value; it transports whatever the source emits, and the destination reads whatever arrives. Coordination between nodes is the topology and each node's local rule, not a delivery guarantee between the two.

Nodes time their processing in **ticks**: a firing rule may span a
**tick-count window** (e.g. a gate's inhibit/processing window is K ticks),
paced against the human-speed clock by sleeping one cycle at a time
(`SleepCycle`) and comparing `Tick()`. Firing is still
gated on held state — now optionally held across a tick window rather than
resolving instantaneously.

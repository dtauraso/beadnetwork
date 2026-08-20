# Model — bead and node lifecycle

[← MODEL.md](../../MODEL.md)

## Bead lifecycle

A bead crosses in one direction, and one goroutine owns it the whole way:

1. The source node's kind decides a value and hands it to that node's own animation
   goroutine, which places the bead on the segment and step count it currently holds. The
   hand-off does not block and does not wait for the destination — see §Sending.
2. While in flight, that same animation goroutine advances the bead one slot per wake
   (`BeadLine.DriveOneStep`, driven by `BeadAnimation.RunBeadAnimation`). A bead captures its
   segment and step count when placed and keeps them, so a node moving mid-flight does not
   touch it.
3. When the bead reaches its last slot, that same step sends the value on to the destination
   node. Arrival IS the last slot; no clock is compared against a deadline.
4. The destination node receives it and holds it in node-local state.

What the RENDERER is told is the bead's position, computed in Go as `slot × SlotR` along the
segment and streamed in the EdgeBead block. The editor draws what arrives and interpolates
nothing. The channels between nodes are the real connection and are never drawn.

The source node times its own delivery. There is no TS-driven delivery signal — the renderer
is told which bead is lit, not asked when a bead has arrived.

## Node lifecycle

- A destination node receives values over its input channel and holds them in node-local
  state.
- When the node's firing rule is satisfied, it fires.
- **One edge per input.** An input is a channel-binding ROLE — a name in the kind's
  SPEC.md `## Ports` table that `a.In("<name>")` binds a channel to at LOAD, never a Go field
  per channel — and it is fed by EXACTLY ONE edge. Fan-in (several edges into one input) is
  not part of the model. A node that needs several sources declares DISTINCT inputs (e.g. a
  gate's separate left/right inputs). Several beads may still be in flight toward one input at
  once — its single source emitting repeatedly — each carrying its OWN placement geometry, so
  the source node moving mid-flight re-derives nothing and disturbs nothing. The loader rejects
  a fan-in topology at parse (`validateNoFanIn`); the guard
  `src/Node/Wiring/loadspec/check-no-fan-in.sh` keeps it out of the committed diagram.

## Sending

A node emits a bead whenever its own rule says to. It does not check anything downstream and
does not wait on the destination — there is no clear/busy state, no acknowledgment, no
back-pressure. A channel here is NOT a blocking handshake: the send must never be able to
stall the source node, so it is sized and shaped so the source's send always succeeds
immediately. Several beads may be in flight at once, each its own value; the destination reads
whatever arrives. Coordination between nodes is the topology and each node's local rule, not a
delivery guarantee between the two.

Nodes time their processing in **ticks**: a firing rule may span a **tick-count window** (e.g.
a gate's inhibit/processing window is K ticks), paced against the human-speed clock by sleeping
one cycle at a time (`SleepCycle`) and comparing `Tick()`. Firing is still gated on held state —
now optionally held across a tick window rather than resolving instantaneously.

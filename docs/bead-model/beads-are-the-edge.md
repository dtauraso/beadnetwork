# Beads are the edge — plan

**Superseded on the LENGTH model:** every `arcLength`/chord-distance formula this document
describes below (item 4's `count = len / s`, `ticksToCross = arcLength / pulseSpeed`, and the
in-flight revision rule's `newArc`) is historical — the arc-length model it plans is gone.
[docs/bead-model/bead-lattice.md](bead-lattice.md) is now the length model: an edge's length is ONE
INTEGER (the bead-step count between two nodes' tori), computed from index arithmetic on the
source node's own stored `LocalPolar`, no arc, no sqrt, no chord. This document's staging
narrative, the chain/lighting split, and the ownership decisions below it are still current;
only the "how long is an edge" formula changed.

The edge stops being a thing that owns geometry and timing. A node owns a **sequence of
placeholder beads** toward each neighbour. The wire goroutine is removed and its animation
logic moves into the node. Traversal is no longer a bead moving along a wire — it is
successive placeholder beads **lighting up** at their fixed percentages.

**The chain is the animation, not the connection.** Two separate things, and conflating them
is the drift this document has already made once:

- **The channels** are the real connection: direct node-to-node, carrying delivery and the
  bead add/drop messages that maintain the chain. They are not drawn.
- **The bead chain** is the animation surface. It is what a value in transit LOOKS like. It
  is NOT a picture of the channel, and NOT a picture of the messages travelling on it.

So the bead count is not the count of anything on a channel, and an add/drop message is not
"a bead appearing" in the visual sense — it is chain maintenance that happens to change how
many placeholders exist. A chain sits there fully populated with nothing traversing it; the
lighting is the only thing that moves.

Model as stated (three parts, all from David):

1. **Beads replace the edge.** `len(edge) × x%` placeholder beads, arranged so the sequence
   looks like an edge. The bead sequence is part of the node, so each node is "next to" its
   neighbours. Moving a node is constant time.
2. **Nodes are connected directly to each other by channels.** Node-to-node messages say
   which node makes another bead and which node adjusts to an x%-farther distance. There is
   no edge entity mediating it. These messages maintain the chain; they are not what the
   chain draws.
3. **The wire goroutine is removed.** Its animation logic is owned by the node.

## Why this is not the rejected chain model

`memory/project/layout-model/project_wire_is_straight_line_not_chain.md` reverted a bead-chain wire and says not
to re-propose it. That rejection was specifically about **neighbour-midpoint relaxation** —
a bead's position depending on its neighbours' positions, making straightness a diffusion
process that follows a drag in O(N²) (measured ~1.5s at N≈40). The same memory names the
one design that escapes it:

> The only escape is giving each bead the two anchors directly (lerp on the anchor line →
> dependency depth 1), which means broadcasting endpoint data to every bead.

A bead at a **fixed percentage** is exactly that: dependency depth 1, no follow, no
relaxation, no born-to-hold-spacing settling. The ban does not apply, and the escape it
names is what this is.

**The line not to cross:** no bead position may depend on another bead's position. The
moment spacing is maintained by looking at the bead next door, this becomes the reverted
design. Positions come from the node (and its neighbour's), never from adjacent beads.

## Why moving a node is constant time

Only if bead offsets are **node-local**, which the buffer already does elsewhere. The
Interior block stores `OX/OY/OZ` as "the Go-owned NODE-LOCAL slot offset (relative to the
node center — the renderer adds the node center to get the world position)"
(`Buffer/bufschema/layout.go`). Go owns the offsets, the renderer does one add. That is not TS owning
positions and it ships today.

So a node moving rewrites its own centre and nothing else — its whole bead sequence rides
along for free. What still has to happen is **re-aiming when the neighbour moves**, and that
is the existing one-hop neighbour message (`neighborSetC`), not a new mechanism.

Without node-local offsets there is no constant time: Go owns absolute bead positions today
(MODEL.md "Geometry and time"), so a move would cost `degree × N` position writes — worse
than the status quo, not better. **Node-local offsets are load-bearing, not an optimisation.**

## What this deletes, and the open model questions

See [beads-are-the-edge-open-questions.md](beads-are-the-edge-open-questions.md) for the
measured deletion size, the per-edge-stream-ownership consequence, the five open model
questions (three now settled), and the three places MODEL.md contradicts this the moment it
lands.

## Staging as built

See [beads-are-the-edge-staging.md](beads-are-the-edge-staging.md) for the four staging
steps as actually built (two of the four planned steps turned out to be wrong about what was
required), what the plan got wrong, and why the two representations (visual chain vs.
channel messages) are the design, not a smell.

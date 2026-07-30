# Beads are the edge — plan

**Superseded on the LENGTH model:** every `arcLength`/chord-distance formula this document
describes below (item 4's `count = len / s`, `ticksToCross = arcLength / pulseSpeed`, and the
in-flight revision rule's `newArc`) is historical — the arc-length model it plans is gone.
[docs/bead-lattice.md](bead-lattice.md) is now the length model: an edge's length is ONE
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

`memory/project_wire_is_straight_line_not_chain.md` reverted a bead-chain wire and says not
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
(`Buffer/layout.go`). Go owns the offsets, the renderer does one add. That is not TS owning
positions and it ships today.

So a node moving rewrites its own centre and nothing else — its whole bead sequence rides
along for free. What still has to happen is **re-aiming when the neighbour moves**, and that
is the existing one-hop neighbour message (`neighborSetC`), not a new mechanism.

Without node-local offsets there is no constant time: Go owns absolute bead positions today
(MODEL.md "Geometry and time"), so a move would cost `degree × N` position writes — worse
than the status quo, not better. **Node-local offsets are load-bearing, not an optimisation.**

## What this deletes

Measured on this branch:

- `nodes/wire/paced_wire.go` — **781 lines**, referenced by **39 files** (16 non-test):
  `Trace/Trace.go`, `nodes/wire/{ports,geometry}.go`, `nodes/input/node.go`,
  `nodes/Wiring/{gesture,topo_spec,build,stream_wiring,edge_mover,loader,node_mover,stdin_reader,input_codec,port_bindings}.go`,
  `Buffer/edge_stream_frame.go`.
- `edgeMover` / `StreamKindEdge` touch **35 files**.

This is the largest change since the per-owner buffer split. It is not a rendering tweak.

## The consequence nobody asked for yet: per-edge streams lose their owner

`memory/feedback_no_single_writer_bridge.md` is *one goroutine = one dedicated stream*. The
per-edge fds exist because an edge is a goroutine. Remove the goroutine and the per-edge
stream has no owner — its frames have to fold into the SOURCE node's own stream (the node
that owns that edge, see open question 1).

That reaches further than the network:

- `topology/counts.json` and the fd layout (`Buffer/stream_fds.go`, `WIREFOLD_STREAM_FDS`)
- `tools/topology-vscode/src/runCommand.ts`'s fixed fd allocation and its per-edge
  last-frame cache
- the headless harness's spawn (`headless_stream_helpers_test.go`) and
  `TestHeadlessEdgeFdDedicatedStream`, which exists precisely to prove edges reach their own
  fd — that test does not get fixed, it gets **deleted**, because the invariant it guards
  stops being true
- `Buffer/edge_stream_frame.go` and the Edge/Bead blocks

**Decide this before writing code**, because it determines whether this is one change or two:
does the edge stream disappear in the same commit as the wire goroutine, or does the edge
keep a (now goroutine-less) stream for a transition? A stream with no goroutine writing it
contradicts the bridge invariant, so the honest answer is probably "same commit," which
makes the commit very large.

## Open model questions

1. ~~Where do the two half-sequences meet?~~ **Settled: there are no halves.** The SOURCE
   node owns the whole sequence, matching ownership the repo already has — an edge is stored
   at `topology/nodes/<source>/edges/<label>.json`, outgoing only, and "carries no `source`
   key: that is the directory it sits in" (`.claude/rules/persistence-ownership.md`). A
   midpoint split was an invention of this document, and a harmful one: the animation is
   directional, so splitting it forces the lit index to HAND OFF from one node's sequence to
   the other's mid-flight — a coordination point in the middle of the one operation that is
   supposed to be purely local. It would also put two owners on one edge's geometry where
   the persistence layout deliberately has one. Constant-time movement does not need it:
   source moves, its whole sequence rides along; target moves, the source re-aims on the
   one-hop neighbour message it already receives. Same asymmetry the stored edge file has.
2. **Who advances the lit percentage, and does delivery still work the same?** Today the wire
   times its own delivery and sends the bead over its out-channel when `ticksToCross`
   elapse. With the wire gone, the **source node** times it and sends to the destination over
   the node-to-node channel it already has. That is a smaller change than it sounds — the
   timing logic moves, it does not disappear.
3. ~~Bead birth/retirement churn.~~ **Settled: count IS length-proportional.** A drag
   changes the count, and those add/drop messages are chain maintenance on the channel — not
   anything traversing (see the opening distinction). Still worth measuring how often a drag
   changes the count, but that is a layout cost to observe, not a reason to prefer a fixed
   count.
4. ~~Uniform pulse speed.~~ **Settled, and it dissolves rather than resolves.** The dwell at
   each bead is the same as today's uniform pulse bead takes to cover that distance. Because
   the count is length-proportional (3), **spacing is constant in world distance**, so a
   constant dwell per bead IS uniform world speed — there is nothing to derive per edge:

       count = len / s            s = constant spacing
       dwell = d ticks per bead   one global constant
       speed = s / d              constant on every edge
       total = count × d = (len/s) × d = len / speed

   That last line is today's `ticksToCross = arcLength / pulseSpeed` exactly, so this
   reproduces current timing rather than approximating it, and the per-edge arc-length
   division disappears. `memory/feedback_uniform_pulse_speed.md` is satisfied structurally
   instead of by computation.

   The in-flight revision rule survives and gets simpler. MODEL.md preserves fractional
   progress `t` across a geometry edit and recomputes remaining ticks from the new arc; here
   `index = t × newCount` and `remaining = (newCount − index) × d`, which is
   `(1−t)·newArc/speed` — the same rule as index arithmetic rather than distance, which is
   the shape `memory/feedback_abc_times_constant_not_rederive.md` asks for.
5. **What happens to the Bead block?** Today `bufLayoutBead` (`Buffer/layout.go:56`) is
   `X/Y/Z` world position + `Value`, **one row per live in-flight bead**, fed from
   `KindEdgeBead` events and read by `edge-stream-blocks.ts` off the per-edge stream — Go
   recomputes that absolute position every tick and streams it. Under this model nothing
   produces it: there is no bead with a moving position, only a fixed chain plus a lit index.
   So those three columns lose their writer. Whether the block is deleted or becomes the
   chain's node-local offsets (the Interior block's shape) is open;
   `check-no-dead-buffer-column.sh` forces the decision rather than letting them linger.

## MODEL.md is wrong the moment this lands

MODEL.md is the pinned model, and this contradicts it in at least three places. These edits
belong in the **same commit** as the behaviour, not a follow-up:

- **"Driver"** — "each wire is ALSO a goroutine — not a passive struct another goroutine
  steps" becomes false.
- **"Wire lifecycle"** — describes a lifecycle that no longer has an owner.
- **"Geometry and time"** — "Go owns the bead's PROGRESS ... AND the bead's absolute world
  position" becomes progress plus a node-local offset. `ticksToCross = arcLength /
  pulseSpeed` is no longer computed per edge: constant spacing × constant dwell yields it
  identically (open question 4). The in-flight revision rule ("preserves the bead's
  FRACTIONAL progress `t`") gets *simpler*, not harder: `t` is the only state there is, so a
  geometry edit cannot make it swing.
- **"Allowed vocabulary"** — "wire" as an active goroutine has to be re-defined or retired.

## Staging — as BUILT

Recorded after the fact, because two of the four planned steps were wrong about what was
required. Each commit verified with `bash scripts/stop-checks.sh --cli`.

1. **Node-local chain-bead offsets, wire untouched.** `ChainBead` block, `chainBeads()`,
   `outTargets`, packed on each node's own frame, decoded + aggregated TS-side. Nothing
   rendered. DONE.
2. **Draw the chain; KEEP the moving bead.** Planned as "stop rendering the moving bead",
   which was impossible: lighting needs traversal progress, which lived in `PacedWire` and
   reached the editor only on the EDGE stream. DONE with the moving bead still running.
3. **The node drives its own wires and lights its own chain.** `edgeMover.run`'s
   `DriveOneCycle` call moved to `nodeMover.run` — there was exactly ONE place a wire was
   ever driven, so removing "the wire goroutine" was that call moving, not a rewrite.
   Driving it there is what let the node read its own wires' in-flight `t`
   (`LiveBeadFractions`) and light `index = t × count` with no cross-goroutine read.
   `BeadInstances` left the scene. Then MODEL.md/CLAUDE.md, then the Bead block. DONE.
4. ~~Fd-layout fallout.~~ **Not needed, and the premise was wrong.** This plan claimed the
   per-edge streams would lose their owner. They did not: only the wire's DRIVE moved, and
   `edgeMover` is still a goroutine writing its own stream, so the one-goroutine-one-stream
   invariant never broke. `counts.json`, `stream_fds.go`, `runCommand.ts` and the headless
   spawn were untouched. `TestHeadlessEdgeFdDedicatedStream` was NOT deleted — its
   per-edge-fd assertion is unchanged.

### What the plan got wrong, kept for the next reader

- **"Delete `PacedWire`" was overreach.** The model said the wire GOROUTINE goes and its
  ANIMATION LOGIC moves to the node. Both happened. `PacedWire` survives as what it should
  be: a passive delay queue its source node steps. 781 lines were not deleted, and did not
  need to be.
- **The scariest-sounding consequence never materialised.** "Per-edge streams lose their
  owner" drove the whole "step 3 is indivisible and very large" framing. It was false.
- **A duplicated constant bit twice.** `hdrSize = 20` hardcoded in three headless tests, and
  `BufNodeStreamChainBeadStride = 12` beside a 13-byte row. Both were fixed by DELETING the
  duplicate in favour of the generated constant, not by updating the copy.
- **The real hazard was a race, and the plan never mentioned it.** Moving the wire's drive to
  the node silently invalidated `edgeMover`'s `LiveBeadRows` read. Deleting the undrawn bead
  path removed it. Nothing in the staging above predicted this; it surfaced only from asking
  "who owns this state now" while removing dead surface.

## The two representations are the design, not a smell

An earlier draft of this document called the coexistence of two representations a drift
risk. That was wrong, and worth recording so it does not get "cleaned up" later by someone
reading only the code:

- the **visual bead chain** connecting the nodes, and
- the nodes **connected directly by channels** carrying add/remove-bead messages

are two different things, not two copies of one thing. Neither is derived from the other and
neither is redundant. The chain never depicts the channel, and the channel traffic is never
drawn. Any future change that tries to collapse them into one representation — deriving the
chain from channel traffic, or drawing a message as a bead — is undoing the design.

The one duplication that IS transient is the migration itself: during steps 1-2 a moving
bead and a lit chain both depict the same traversal. That pair is two copies of one thing,
and step 3 removes one of them. It says nothing about the chain/channel duality above.

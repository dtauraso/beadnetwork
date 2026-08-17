# Beads are the edge — staging as built

[← beads-are-the-edge.md](beads-are-the-edge.md)

**Read as a record of what happened, not of what is there now.** Two kinds of name below no
longer resolve. Every test it cites is gone — all test files in this repo were deleted, so
`TestHeadlessEdgeFdDedicatedStream` and the "three headless tests" that duplicated
`hdrSize` no longer exist and cannot be consulted (CLAUDE.md, "There are no tests"). And
`LiveBeadFractions` was replaced by `lattice.BeadFraction(nowTick, placementTick,
crossTicks)`, called per bead in `nodes/wire/live_beads.go`. The events described happened
as recorded; the surface they happened to has moved.

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

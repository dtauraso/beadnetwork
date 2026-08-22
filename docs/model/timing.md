# Model — geometry, time, and the driver

[← MODEL.md](../../MODEL.md)

## Geometry and time

- Geometry sets traversal in SLOTS, not in ticks: a bead crosses `steps` of them, where
  `steps` is `edgegeom.EdgeStepCount`, computed by the SOURCE NODE from its own live
  measured distance to the target — one integer, not an arc length divided by a speed.
  Geometry has no other effect on timing.
- A slot is one radial index (`lattice.SlotR`), and it is exactly the distance a bead
  covers in one pulse at the authored pulse speed. One radial index IS one slot: the
  loader refuses a `constants.json` whose `constantR` is anything else, because the scene
  grid and the bead lattice are the same grid.
- A bead's position is its slot. `slot × SlotR` along the segment, arithmetic, with no
  fraction of an elapsed time in it — so nothing accumulates drift and a skipped frame
  costs nothing. Reaching the last slot IS arrival; no clock is compared against a
  deadline to discover it.
- A bead in flight does not hear about a drag. It captures the segment and step count it
  was placed on and keeps them to the end, so a node that moves afterwards changes
  nothing about it — not its direction, not its distance, not where it stops. Only beads
  placed after the move use the new geometry, which they get from the placement. There is
  no in-flight revision and no fractional progress to preserve.
- Go computes and streams the absolute bead position, on the EdgeBead columns, read and drawn by
  `Categories/Node/BeadAnimation/ChainBeadInstances.tsx`. The editor
  does not interpolate, does not own positions, and is never asked when a bead arrived.
- Durations are counted in slots at `lattice.PulsesPerSlot`, the one number every duration
  shares: the animation goroutine sleeps it per slot, and Time, TimeStart, input's cadence
  and helddrive's hold each multiply their own step count by it. A window and the
  traversal it waits for move together by construction.

## Driver

**Self-scheduling node goroutines.** A node is four goroutines, each paced by a different
thing:

- the **kind** goroutine — the kind's `Update`, paced by the sim clock; it decides values
- the **animation** goroutine — `BeadAnimation.RunBeadAnimation`, one per node with outputs;
  it advances every bead the node owns by one slot per wake and writes its bead stream
- the **geometry** goroutine — `RunGeometry`, paced by nothing, blocking on its own inbox
  so a drag is served at the rate of the hand dragging it
- the **rule** goroutine — `rulenode.RuleNode.Run`, which fans out one forwarder per peer

A `BeadLine` is state one of them owns, not a fifth goroutine: it holds the segment, the step
count and the beads on it, and the animation goroutine steps it.

There is no central walker and no play/pause gate.

The animation goroutine parks between slots, waiting on BOTH a timer and the slider's own
channel, so a speed change takes hold at once rather than after the rest of a sleep. At
slider 0 it parks on the channel alone: paused is no wait at all, not a very long one.
Delivery is not triggered by the renderer — there is no cross-boundary delivery signal.

There is one tick clock (the human-speed clock) but no lockstep round or simultaneity
layer: goroutines schedule independently against the shared tick, each sleeping and
re-reading `Tick()` on its own — they are not aligned into global rounds, and the network
does not count rounds. Coordination between nodes happens through the values nodes emit and
the topology — not through round-alignment or a delivery handshake. Any reasoning that treats
activity as a sequence of globally-aligned lockstep rounds is drift; re-derive from local
rules that wait on ticks over channels.

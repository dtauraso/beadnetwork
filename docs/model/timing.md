# Model — geometry, time, and the driver

[← MODEL.md](../../MODEL.md)

## Geometry and time

- Wire geometry sets traversal in ticks:
  `ticksToCross = steps * DwellTicksPerBead` (steps: docs/bead-model/bead-lattice.md "The
  count", computed by the SOURCE NODE from its own live measured distance to the
  target — one integer, not an arc length divided by a speed). Geometry has
  no other effect on timing.
- A geometry edit re-derives traversal time. While a bead is in flight,
  the in-flight revision PRESERVES the bead's FRACTIONAL progress `t` (its
  proportion along the wire) — NOT the absolute distance covered. On the
  edit the bead stays at the same fraction `t`, and the remaining ticks are
  recomputed from the NEW step count at the uniform per-step dwell:
  `remainingTicks = (1−t)·newSteps·DwellTicksPerBead`. So the bead rides
  smoothly at the same proportion as the wire reshapes (no t-swing race as a
  node is dragged), and a longer or shorter wire still traverses at constant
  world-speed. (Preserving distance instead would let `t` jump as the step
  count changes.)
- Go owns the bead's PROGRESS (the fraction `t`, timed in ticks on the human-speed clock).
  It no longer computes or streams an absolute bead position: nothing draws a moving bead.
  The source node quantises its own `t` onto its own chain and streams which bead is LIT
  (`readChainBeadLit` from `tools/topology-vscode/src/schema/buffer-layout.ts`, consumed by
  `tools/topology-vscode/src/webview/three/scene/beads/ChainBeadInstances.tsx`). The editor does not
  interpolate, does not own positions, and does not decide which bead is lit.
- Durations are tick counts: bead traversal (`ticksToCross`) and node processing windows.

## Driver

**Self-scheduling node goroutines.** Each node is a goroutine. A WIRE IS
NOT: it is a passive delay queue, and the goroutine that steps it is its
own SOURCE NODE's mover (`NodeMover.Run` calls `DriveOneCycle` for each of
its outgoing wires). There is still no central walker and no play/pause
gate — the clock is free-running and the animation never halts.

Each SOURCE NODE times its own outgoing deliveries on its own reading of
the human-speed clock: when a bead's `ticksToCross` have elapsed, that
drive sends it over the wire's out-channel to the destination node, which
receives it. Delivery is not triggered by the renderer — there is no
cross-boundary delivery signal. The editor is told which chain bead is lit;
it is never asked when a bead has arrived.

(This replaced a per-wire goroutine. There was exactly ONE place a wire was
ever driven — `edgeMover.run` — so the change was that call moving to the
node, not a new mechanism. Driving it there is also what lets a node read
its own wires' in-flight fraction to light its own chain without touching
another goroutine's state.)

There is one tick clock (the human-speed clock) but no lockstep round or
simultaneity layer: goroutines schedule independently against the shared
tick, each sleeping a cycle (`SleepCycle`) and re-reading `Tick()` on its
own — they are not aligned into global rounds, and the
network does not count rounds. Coordination between nodes happens through
the values nodes place on wires and the topology — not through
round-alignment or a delivery handshake. Any reasoning that treats
activity as a sequence of globally-aligned lockstep rounds is drift;
re-derive from local rules that wait on ticks over channels and wires.

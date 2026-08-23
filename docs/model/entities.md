# Model — entities

[← MODEL.md](../../MODEL.md)

## What things are

- **Bead.** A value in transit from a source node to a destination node, holding the slot it
  stands on. Its position is `slot × SlotR` along the segment it was placed on, and reaching
  the last slot IS arrival. A bead is data — `inflightBead`, `Categories/Node/BeadAnimation/bead_placement.go` — carried by
  the goroutine that owns it, never a goroutine itself.

  A bead carries its OWN geometry: the segment and step count it was placed on, captured at
  placement and kept to the end. Several beads may be in flight along one line at once, each
  with its own, so the source node moving mid-flight changes nothing about any of them.
  Geometry travels WITH the bead rather than being stored once alongside the line.

- **Bead behaviour is node behaviour.** The SOURCE NODE's animation goroutine
  (`BeadAnimation.RunBeadAnimation`, `Categories/Node/BeadAnimation/bead_animation.go`) owns the whole
  of it: it accepts the value, holds it at a slot, advances that slot on its own pulse,
  computes the position, decides arrival when the slot reaches the step count, and hands the
  value on. One goroutine owns that state from placement to delivery, which is why none of it
  is guarded by a mutex — ownership replaces locking, the same move that removed
  `RealClock.mu`. Do not add a lock here "for safety": if two goroutines ever need to touch
  this state, the ownership model broke, and a mutex hides that rather than fixing it.

- **Bead line (`BeadLine`).** The line beads travel along, `Categories/Node/BeadAnimation/bead_line.go` —
  the segment and step count between two nodes, plus the beads currently on it. It holds
  state, not behaviour: the source node's animation goroutine steps it (`DriveOneStep`), and
  it has no goroutine, no clock and no send policy of its own. Its step count is
  `edgegeom.EdgeStepCount`.

  A line's TRANSPORT state and the REPORTING it does for the renderer are separate types:
  `BeadLine` holds the beads, the channels on each end, and the arrival math; its readout
  (`beadReadout`, `Categories/Node/BeadAnimation/readout.go`) holds the pending Position/Arrive buffer,
  the `Trace` handle and the debug-breadcrumb channel. Both belong to the same single
  animation goroutine — the split says which concern a field serves, it does not add an owner.

- **Node goroutine.** One of SEVERAL serving the same node id, not "the" goroutine for that
  node: a node id is a referent, and each goroutine that tags frames with it owns one job
  (the kind's own logic, geometry and interaction, bead animation) and its own state. They
  are peers sharing nothing but the id. **Only the animation job sleeps on the human-speed
  clock**; a goroutine that both paces beads and reads input makes the bead rate the
  interaction rate, which is why the jobs are separate goroutines rather than phases of one
  loop. The kind's own job receives values over its input channel, holds them in node-local
  state until its firing rule is satisfied, then fires.

  Held values live in node-local state. (That is a different concept from a trace event's `Slot`
  field — `Categories/Node/owners/trace_event.go` — which is a live 2×2 interior VISUAL grid position,
  `slot = gridRow*2 + gridCol`, for where a held bead is drawn inside a node.)

- **Node input.** A ROLE, not a place: declared by the node kind in its SPEC.md `## Ports`
  table — the one declaration, which the Go side reads generated — and bound to a channel at LOAD time (`a.In(...)`), never drawn and never hit-testable, and read
  through a `Receiver` (`Categories/Node/BeadAnimation/receiver.go`). **One input is fed by exactly one
  edge**; a node that needs several sources declares several inputs (see §Node lifecycle).
  Inputs carry no geometry of their own — an edge attaches at its two nodes' SURFACES
  (`nodegeom.NodeTorusOuterR`).

  The channel between two nodes is the goroutine boundary, and nothing more. It is the real
  connection and is never drawn; what is drawn is the beads.

- **Clock (the human-speed clock).** There is exactly one clock: the system monotonic clock,
  read through a **scale** so it advances in integer **ticks** at human-watchable speed
  (`tick = ⌊(now − start) / tickPeriod⌋`; the scale is the human-speed / playback-speed knob,
  `MsPerTick = 16` ⇒ ≈62.5 ticks/sec). All timing is **tick counts**, not wall-clock
  durations. The model is **sleep-only**: a pacing loop calls `SleepCycle` to wait exactly ONE
  cycle and re-reads `Tick()`, rather than blocking on a target tick — there is no
  wait-until-tick-k primitive. The clock is **free-running**: it advances monotonically with
  wall time and never pauses. **Everything that animates runs in these ticks:** beads
  travelling, all in-node animations, and all node/gate processing windows. Per-update tick
  counts come from formulas, not literals — a bead crossing an edge takes `steps` slots at
  `lattice.PulsesPerSlot` pulses each (`Categories/Node/BeadAnimation/lattice/bead_lattice.go`); node
  processing windows are tick counts. There is no separate render cadence — the tick IS the
  animation clock.

  Each goroutine holds its own `Copy()` of the clock and reads its own tick; there is no
  shared clock to pin a tick against. But a bead's **placement tick** (when it started
  crossing) is a DIFFERENT reading than the step tick that later advances it: placement is
  decided by the **emitting** goroutine, at the moment it calls `Send`, from its own clock,
  once per emission — not re-derived later by whichever goroutine drains the placement
  request. This is what lets several beads placed in one emission (a broadcast fan-out)
  provably share one placement tick: reading a fresh clock value per line in the drain pass
  can straddle a tick boundary between two beads placed microseconds apart, splitting one
  emission across two ticks (the observed bug this fixed). Read once, stamp everywhere, in the
  same call.

## Position is arithmetic, and no position update waits on the human clock

**INVARIANT: no position update may be gated on the human clock, and no animation step may
run on the system clock.** `MsPerTick` (`Categories/Clock/`) names the human-speed clock; it exists so
a person can watch a bead cross, and geometry must never run on it — one propagation hop per
tick would make even a straight traversal visibly slow.

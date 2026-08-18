# Model — entities

[← MODEL.md](../../MODEL.md)

## What things are

There are TWO different things this project calls a "bead", and conflating them is the
mistake to avoid:

- **In-flight VALUE bead (`BeadRun.inflight`).** A value in transit from a source node to
  a destination node, timed by the wire's own traversal fraction `t`. This bead is data, not
  a goroutine — see the Wire bullet below, unchanged by the chain-bead goroutine model.
- **Chain (render/placeholder) bead.** REMOVED. It was the node-owned visual entity that
  stood in for a traversal — one per node-local offset along an outgoing edge's aim — and
  it existed because a node laid a chain toward where it believed its neighbour was. That
  belief was a cache of the neighbour's centre, kept current by a centre broadcast, and
  both are gone. What renders a traversal now is the in-flight value bead itself, placed
  by the edge it travels on the segment that edge already holds
  (`nodes/bead/live_beads.go`'s `LiveBeadRows`), streamed on the EDGE's own frame as a
  world position. The goroutine-per-bead primitive
  (`nodes/bead/beadchain/bead_actor.go`, `bead_wake_group.go`) survives with no production
  call site. The rest of this bullet describes that removed entity; it is kept because the
  clock/channel split below is the design a replacement would have to answer to.

  This bead was driven by TWO clocks
  over THREE structurally distinct channel sets:
  - **Geometry** (machine time): a `BroadcastChain` carrying the owning node's live aim
    direction (broadcast in NODE-LOCAL terms — a unit vector, no absolute center — so a
    bead's own resolved position IS the node-local offset the buffer has always carried),
    broadcast to every bead on that edge in ONE close (`BeadWakeGroup.BroadcastGeometry`,
    called by `ReconcileBeadChain` only when the aim or the bead count actually changed) —
    a body force, dependency depth 1: each bead computes its own position directly from the
    broadcast and its own fixed offset, never from a neighbour bead's position
    (memory/project/layout-model/project_wire_is_straight_line_not_chain.md's O(N²) defect was momentum-free
    midpoint averaging plus human-clock gating, not "a chain of beads" per se — see that
    memory file's corrected framing).
  - **Animation/tick** (human time, `MsPerTick`): a pulse from the bead's own
    `time.Ticker` (owned by the bead, stopped on its own `stop` channel) advancing
    lit/carried-value state.
  - **Mode**: two more `BroadcastChain`s — wake (sets the ONE local `dragging` flag) and
    settle (clears it) — each advanced by a SINGLE close from the owning node
    (`BeadWakeGroup.StartDrag`/`EndDrag`), once per drag gesture (the gesture FSM's
    `gestPointerDown`→`gesturefsm.GestDragging` edge sends `movemsg.KindDragStart`;
    `gestPointerUp`, on every path a drag ends by, sends the mirroring
    `movemsg.KindDragEnd`), never per pointer event. Position and animation are disjoint
    state (one writer each), so the two clocks never coordinate and the bead is never in
    both modes.

  Bead-goroutine lifetime follows chain length: `ReconcileBeadChain` grows a chain by
  starting one goroutine per added bead (at the chain end, matching bead CRUD's own
  convention, `bead_crud.go`) and shrinks it by closing each removed bead's OWN dedicated
  stop channel, which the removed bead's `run` loop observes and returns from immediately —
  no goroutine outlives its bead. `tools/network/beads/check-bead-actor-has-call-site.sh` fails the build if
  this primitive ever loses its last production reference.

  The bead's own goroutine is ONE `select` over all three channel sets, with **no
  `default:` case** — parked at zero CPU when idle, never spinning
  (`tools/network/beads/check-no-select-default.sh`). A node's wake/settle/geometry broadcast is a single
  channel close, never a loop over N beads (`tools/network/beads/check-broadcast-is-close-not-loop.sh`,
  via the lock-free `BroadcastChain` generation-chain primitive: the owning goroutine writes
  `Next` before closing `Fire`, so a woken receiver can read `Next` with no lock/atomic —
  Go's memory model makes the close a happens-before edge for that read).

  **INVARIANT: no position update may be gated on the human clock, and no animation step
  may run on the system clock.** (`MsPerTick = 16`, clock.go, names the human-speed clock;
  it exists so a person can watch a bead cross a wire, and geometry must never run on it —
  one propagation hop per tick would make even linear traversal visibly slow.)

  This is additive to the transport model below, not a replacement of it: `BeadRun`'s
  in-flight value beads remain the passive delay queue MODEL.md always described; the chain
  bead is what renders a traversal — the one entity in this codebase that is BOTH a
  goroutine AND owns local per-drag mode state.

  **Known boundary:** the bead-actor path is driven by the SOURCE node's own drag (the node
  that owns the chain, per "an edge is stored under its source node" above) —
  the chain's `StartDrag`/`EndDrag` (reached through `owners.Beads.PostBeadDrag` and
  `ApplyBeadDrag`) fire only when `g.DragNode` is that source node. Dragging
  the TARGET end of an edge still repositions that edge's beads with no visible lag (every
  chain rebuild recomputes from the live `partnerCenters` push the target's own
  `ApplyCenter` already sends on every commit — unchanged, pre-existing machinery), but it
  does so through `beadcrud`'s own inline placement math for that edge on that call rather
  than through the target also toggling that chain's `BeadWakeGroup` mode flags. The
  `BeadWakeGroup`/`Bead` primitive itself supports either endpoint waking the SAME beads;
  wiring the TARGET's own drag lifecycle through to the
  SOURCE's chain (so target-drags also toggle the mode flag, not just geometry) is future
  work, not yet done.

- **Wire (`BeadRun`).** Transport. A PASSIVE delay queue, not a
  goroutine: the source node sends a bead over the wire's in-channel to
  place it, and that SAME source node times the traversal on its own clock
  reading (each goroutine owns its own clock copy — see the Clock bullet
  below) by driving the wire each cycle, then on traversal-complete sends
  the bead over its out-channel to the destination. The wire is no longer
  the visual depiction either — the source node's own chain of placeholder
  beads is — its length is `edgegeom.EdgeStepCount`. There is one owner
  of `inflight`/`delivered` and the in-flight geometry: the source node
  goroutine. Because it is the sole owner, `BeadRun.mu` does not exist
  — ownership replaces locking, the same move that removed `RealClock.mu`.
  Do not reintroduce a lock here "for safety"; a second lock on top of
  single-goroutine ownership is dead weight, and if two goroutines ever
  need to touch this state again that is a sign the ownership model
  broke, not a reason to add a mutex. The wire applies no send policy —
  see §Sending. A wire's TRANSPORT state and the REPORTING it does for the
  renderer are separate types: `BeadRun` holds the queue (`inflight`,
  the in/out channels, dwell, the arrival math), and its `readout`
  (`wireReadout`, `nodes/bead/readout.go`) holds the pending
  Position/Arrive buffer, the `Trace` handle and the debug-breadcrumb
  channel. Both are owned by the same single source-node goroutine; the
  split says which concern a field belongs to, it does not add an owner.
- **Node goroutine.** One of SEVERAL serving the same node id, not "the" goroutine for that
  node: a node id is a referent, and each goroutine that tags frames with it owns one job
  (the kind's own logic, geometry and interaction, bead animation) and its own state. They
  are peers sharing nothing but the id. **Only the animation job sleeps on the human-speed
  clock**; a goroutine that both paces beads and reads input makes the bead rate the
  interaction rate, which is why the jobs are separate goroutines rather than phases of one
  loop. The kind's own job receives beads over its input port's channel,
  holds them in node-local state until its firing rule is satisfied,
  then fires. There is no held-value slot in this model sense — node-local held
  state replaces it. (This is a different concept from the buffer's `Slot`
  column — `nodes/rowevent/row_event.go`, `Buffer/streamframe/stream_events.go`,
  `Buffer/bufschema/layout.go` — which is a live 2x2 interior VISUAL grid position,
  slot = gridRow*2 + gridCol, for where a held bead is drawn inside a node.)
- **Input port.** A ROLE, not a place: declared by the
  node kind as a `Wiring.PortSpec` and bound to a channel at LOAD time
  (`a.In(...)`), never drawn and never hit-testable. One input port is one wire,
  and the wire's out-channel is the connection between them — the node receives
  whatever the source node's drive of that wire sends. Ports carry no geometry of their
  own; an edge attaches at its two nodes' SURFACES (`nodegeom.NodeTorusOuterR`), not at a
  port position.
- **Clock (the human-speed clock).** There is exactly one clock: the system monotonic clock, read through a **scale** so it advances in integer **ticks** at human-watchable speed (`tick = ⌊(now − start) / tickPeriod⌋`; the scale is the human-speed / playback-speed knob, `MsPerTick = 16` ⇒ ≈62.5 ticks/sec). All timing is **tick counts**, not wall-clock durations. The model is **sleep-only**: a pacing loop calls `SleepCycle` to wait exactly ONE cycle and re-reads `Tick()`, rather than blocking on a target tick — there is no wait-until-tick-k primitive. The clock is **free-running**: it advances monotonically with wall time and never pauses (there is no play/pause gate). **Everything that animates runs in these ticks:** bead traveling, all in-node animations, and all node/gate processing windows. Per-update tick counts come from formulas, not literals — a bead crossing an edge takes `ticksToCross = steps * DwellTicksPerBead` (steps the edge's own bead-step count, `DwellTicksPerBead` a uniform constant per bead-lattice step across all wires, `nodes/bead/lattice/bead_lattice.go`); node processing windows are tick counts. There is no separate render cadence — the tick IS the animation clock.
  A wire is stepped with its SOURCE NODE's own clock copy and tick reading,
  exactly like every other per-goroutine clock use — there is no shared
  clock to pin a tick against. But a bead's **placement tick** (when it
  started crossing) is a DIFFERENT reading than the step tick that later
  advances it: placement is decided by the **emitting** goroutine, at the
  moment it calls `Send`, from its own clock, once per emission — not
  re-derived later by whichever goroutine happens to drain the wire's
  in-channel. This is what lets several beads placed in one emission (a
  broadcast fan-out) provably share one placement tick: reading a fresh
  clock value per wire in the drain pass can straddle a tick boundary
  between two beads placed microseconds apart, splitting one emission
  across two ticks (the observed bug this fixed). Read once, stamp
  everywhere, in the same call.

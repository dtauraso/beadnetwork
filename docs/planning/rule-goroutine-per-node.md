# Plan — a node id has a rule goroutine

Delete this file when the change lands. Git history is the archive.

## Target

A node id is served by two goroutines that own disjoint state.

- The **rule goroutine** owns this node's polar rule (`orbitRule`, `orbitActive`) and the
  peer rules it learns over the mesh (`RuleMesh`: `backFromPeer`, `downToPeer`, `peerKey`,
  `selfKey`). It blocks on its peer back-channels and its own edit inbox with no ticker, so
  a rule edit is applied, persisted, broadcast and announced the instant it arrives.
- The **animation goroutine** keeps the clock, its beads, and the node's geometry. It holds
  a READ-ONLY COPY of the rule, delivered by message from the rule goroutine, and reads that
  copy for drag constraints (`nodedrag`) and for the node frame's rule columns.

Neither reads the other's memory. The copy is replication by message, the same shape as
neighbour deltas — not shared state.

## Why

`drainRuleMesh()` is called from `Step` (`nodeactor/pair_node_self.go:70`), which runs once
per `clk.SleepCycle`. Effective clock speed is `userSpeed / ClockDivisor`; the pair scene
declares `ClockDivisor: 64` and `speed.json` holds `0.25`, so `pulsesPerCycle` clamps to 64
and a cycle is 64 x 16ms = ~1s. Every rule edit — the shared-rules checkboxes — waits up to
a full cycle before the node even reads the message. `HumanEditSpeed` (`scene_speed_persist.go:17`)
exists to paper over exactly this, and is broadcast at ONE call site
(`dispatch_apply.go:62`, the tilt-angle path). Speeding the whole world up so a checkbox
can be seen is the wrong mechanism; it is removed by this change.

## Ripple, in order

1. `owners/rule_mesh.go` — `DrainRules` is a non-blocking poll over a map of channels. The
   rule goroutine needs a BLOCKING receive over N peers: one forwarder goroutine per peer
   feeding a single merged `ruleIn` channel, since the peer set is fixed at link time.
2. `owners/topology.go` — `orbitRule`/`orbitActive` move out to the rule goroutine's own
   state; `Topology` keeps a copy plus a setter the animation goroutine calls when it
   receives a rule message.
3. `nodeactor/node_geometry_orbit_edit.go` — `takeOrbitActiveToggle`/`takeOrbitPhiToggle`/
   `takeOrbitMaxTheta` move to the rule goroutine, with persistence
   (`persistOrbitActive`/`persistOrbitRule`).
4. `stdinreader/dispatch_apply.go` — `orbitActive`, `orbitPhi`, `orbitMaxTheta` address the
   rule inbox, not `md.MR.SendMove`. `HumanEditSpeed` and its one call site go.
5. `clock/sleep_cycle.go` — a cycle sleep that also wakes on a wake channel, so the
   animation goroutine emits the frame carrying the new rule immediately instead of at its
   next tick. The bead cadence is unchanged: it wakes early only when a message is waiting.
6. Every node kind's loop passes its wake channel to the sleep. The loops are
   `gatecommon/gate.go`, `PairNode`, `input`, `holdflip`, `pulse`, `pacer`, `Time*`,
   `Pulse*`, `select*`, `NormalSum`.
7. `nodedrag/node_drag.go` — unchanged source, but now reads the replicated copy.

## Verification

Loud runtime failure plus driving the editor. The rule goroutine panics by name if it is
asked for a peer it never linked. Then: open the shared-rules menu, toggle "all nodes", and
the checkboxes update with no perceptible wait at `ClockDivisor: 64` — the observation that
started this. Beads must still traverse at the same rate; a visibly faster sim means the
wake channel is firing on something other than a message.

## Risks

- Two goroutines per node id doubles the count. The repo collapsed to one per node in
  693c7c1dc; this re-splits, but keyed by id and with disjoint state, which is the shape
  8fbf011fc already established ("a node id is a referent several goroutines serve").
- The replicated rule copy is one tick stale for drag constraints in the worst case. A drag
  reads it every pointer move; a rule edit mid-drag is the only way to observe the gap.
- `check-no-network-locks.sh` must stay clean: the copy is written only by the goroutine
  that owns it, never by the sender.

---
name: project_lock_propagation_decentralized
description: Node-move propagation is decentralized (node writes only itself, no central worklist); the rule/gate/anchor equalize/trigger cascade was DELETED 2026-07-18 and replaced by a one-hop neighbor re-quantize, today carried by movemsg.KindCenter over a coalescing neighborSlot
metadata:
  type: project
---

Durable doctrine + a history of propagation mechanisms that were tried and removed.

**Durable principle (David, still holds):** a node can only update ITSELF and send
messages to nodes it's doubly-linked to; a receiver writes only its OWN state. No central
worklist, no collection, no coordinator. State rides the MESSAGE, not shared-memory reads
across goroutines. Still true of the current model. See [[feedback_go_vs_coordinator_bias]].

**Durable invariant (David):** editor-time propagation must NEVER be able to explode the
sim. The central worklist (`RootMove` fixpoint over `md.polarEqs`, commit cb3dd91c) was a
coordinator+collection — the shape MODEL.md forbids — and on an over-constrained set it
amplified positions past 1e28 and persisted the garbage. Reverted in
`revert(locks): remove transitive lock-propagation worklist`. Don't reintroduce a worklist.

**CURRENT model (mechanism merged 590a119c, 2026-07-18; symbols re-verified 2026-08-17):**
dragging node X moves ONLY X (to the cursor). Each direct neighbor STAYS PUT and
re-quantizes its OWN local polar to X from the live offset — θ, φ AND r all fresh — with X
as the one fresh edge (that neighbor's OTHER edges are carried as index×step, not
re-derived). One hop, no forwarding, no cascade. Distances/angles to the (stationary)
neighbors change because X moved.

**Live symbols as of 2026-08-17 (the 590a119c-era names are ALL gone — `node_move.go`,
`moveMsgKindNeighborSetC`, `neighborSetCRequantize`, `requantizePoleTraced` do not exist;
do not grep for them):** a neighbour's move arrives as `movemsg.KindCenter`, and the
gesture FSM's drag as `movemsg.KindDrag`, over the per-node COALESCING slot
`src/Node/Wiring/nodeactor/owners/neighbor_slot.go` (`neighborSlot.deposit`, buffer of 1). That
slot panics on any other kind, and on a message carrying neither a `Target` nor a `Delta` —
because a WHERE collapses by keeping the newest and a HOW FAR collapses by summing, so a new
kind must state its merge rule before it can ride the slot. Routing is still node-to-node
(`src/Node/Wiring/nodeactor/owners/messaging.go`), no worklist. `Trace`'s
`BreadcrumbNeighborSetCRecv` / the `"neighbor-setc-recv"` label survive in
`src/Node/BeadAnimation/receiver.go`'s mapper with NO emitter left — leftover vocabulary, not a
live path.

**DELETED mechanism (was documented here as live fact; gone as of 590a119c):** the
rule/gate/anchor cascade — `handleTrigger`, `moveMsgKindEqualize`/`Trigger`/`GatePlace`/
`Requantize`/`RequantizeSetC`, the per-node role tables (`sourceID`/`ruleSource`,
`followers`/`ruleFollowers`, `forwardTargets`, `isGate`/`gateNeighbors`, `anchoredGates`,
`followerOwner`), `ApplyCascadeRoles`/`deriveCascadeRoles`, `placeEqualRadii`,
`gatePlaceNode`/`equalizeEdgeCLocal`/`moveNodeAndSetEdgeCs`/`placeAtDistanceFromBoth`. It
propagated multi-hop (a rule node re-measured L from its source, equalized its followers,
forwarded to further rule neighbors; gate nodes solved equal-radii placement). David
removed it because the model he wants is one-hop single-assignment, not multi-hop
constraint solving — a drag moves the dragged node and neighbors just re-quantize, they do
NOT reposition to satisfy a length/equal-radii constraint. Do not re-propose the cascade.

**Verified-then-stale lesson (kept as a warning):** an earlier version listed
`links.go`/`locks.go`/`ruleSource`/etc. under a "Go-layer facts" header as live; an audit
agent trusted the header after those symbols were deleted and told another agent to move
code INTO a file that no longer existed — real wasted time. Grep `node_move.go` fresh for
any symbol before acting; treat symbol names here as historical, the MECHANISM as durable.
Initial layout does NOT depend on any cascade — it is a pure forward computation
(`quantoffset.DeriveCenters`, `src/Node/Wiring/polarindex/polar_index.go`) from each node's stored triple; that is why the
cascade could be deleted without leaving nodes unplaced on load.

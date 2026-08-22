# PairNode — firing rule

[← BEHAVIOR.md](BEHAVIOR.md)

REACTIVE, not periodic — the "straightening loop". Each cycle of this node's OWN
Update loop non-blockingly drains two sources and, on either, decides and acts itself —
no round trip to any other goroutine to decide.

This node owns its own geometry directly (`Self *nodeactor.PairNodeSelf`, claimed at build
time via `BuildArgs.ClaimSelfDrive`, driven every cycle by `n.Self.Step`
(`Categories/Node/nodeactor/pair_node_self.go`)) — there is **no separate `NodeMover` goroutine** for
either node of a pair (`Categories/Node/moverreg/mover_registry.go`'s `FinalizeActors` never constructs one for an id
claimed via `ClaimSelfDrive`). `SyncTiltIndex`/`SyncReceivedVector`/`ClearOutBeads` below
are therefore plain method calls on this same object (`PairNodeSelf.SetTiltIndex`/
`SetReceivedVector`/`ClearOutBeads`), not messages to another goroutine — they are named
and shaped like the old cross-goroutine sync calls only because nothing else about what
they carry changed when the separate mover was removed.

**On an In arrival** (`In.PollRecv`): fire this node's own trace and nothing else. A bead
PACES the exchange and marks the round trip; it decides nothing and places nothing onward.

It used to step the tilt index one click in this kind's own fixed direction (a retired
fixed-direction step), with no reference to anything that arrived, stopping only if it happened to
land exactly on `Wiring.QuarterTurnPhiIdx` — which walking away from it never
does. That put TWO rules on one index: the fixed bead step and the vector channel's acute-test
rule. Where they agreed a node double-stepped; where they disagreed they cancelled and it
froze. Measured on the real formulas, a pair marched +1/+2/+3… one way forever on one click
direction and did not move at all on the other. The acute tests are now the only rule that turns a
tilt on an arrival.

The outgoing bead moved with the decision: it is placed by the vector branch when the tests
actually move this node (`handleVectorCycle`), so one message carries one visible bead and
the bead loop lives and dies with the exchange it paces.

`TiltEditIn` (`BuildArgs.TiltEditIn`, a panel-driven edit routed HERE instead of to a
mover) carries THREE distinct edits, applied by `applyTiltEdit` (`Categories/NodeKinds/PairNode/edits.go`):

- **A ▲/▼ panel click** (`TiltVectorAnglePanel.tsx`): applies exactly one ±1 step to the
  named axis, marks this end HELD (a tilt a user set is intent, not error — this end keeps
  its index and does not turn on an arrival; the partner moves instead), and ALSO OPENS THE
  EXCHANGE by sending this node's own outgoing vector alongside a bead — a click that only
  moved an index would leave the partner with nothing to answer.
- **START** (`TiltEditMsg.Start`, the START TILT button, `TiltVectorButtons.tsx`): opens
  the vector exchange from whatever angles are CURRENTLY set — sends this node's own
  `outgoingVector()` on `VectorOut` alongside a bead ("THE KICK"), which is what gives
  `handleVectorCycle` something to reply to. It changes NO index of its own. **START opens
  the exchange from the node whose `PairID` is 1 alone** — the other node's own
  `applyTiltEdit` ignores it, since the exchange is begun from one end only (starting from
  both would answer each other's opener in the same round). With both nodes of a pair already
  at their halt, nothing circulates on In — correctly, since there is nothing left to
  straighten — so the loop has no way to start on its own; START is what a user clicks to
  begin it.

  Pairing two PairNode instances with one edge running each direction (a.Out → b.In,
  b.Out → a.In) needs no seed/bootstrap node: nothing sends until a
  user starts it, so there is no deadlock to bootstrap out of at t=0.
- **RESET** (`TiltEditMsg.Reset`, the RESET TILT button, same `TiltVectorButtons.tsx`):
  runs this node's full `clear()` — zeroes both indices, syncs, drains any value already
  sitting on `VectorIn` (`Wiring.PollRecvVector`, non-blocking, on THIS node's own
  goroutine), drains delivered beads off In, and drops this node's own still-crossing
  outbound beads (`ClearOutBeads`, a call on its own `Self`) — then sends a Reset marker on
  `VectorOut` so the partner clears too. Places NO bead (stop-and-return, not a nudge).
  Without the `VectorIn` drain, a vector already in flight when RESET was pressed would
  arrive on the very next cycle's `handleVectorCycle` and immediately step the tilt
  again, undoing the reset a moment later. `VectorIn` is depth-1 latest-wins, so one
  non-blocking receive empties it fully.

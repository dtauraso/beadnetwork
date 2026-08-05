# Node2Node

## View

| Field | Value |
|-------|-------|
| kindId | 12 |
| kind | node2 |
| bg | #e8eaf6 |
| border | #3949ab |
| text | #1a237e |
| accent | #3949ab |
| minWidth | 70 |
| shape | rect |
| fill | #e8eaf6 |
| stroke | #3949ab |
| width | 70 |
| height | 60 |

## Ports

| Name | Direction | EdgeKind | Notes |
|------|-----------|----------|-------|
| In | in | chain | sole input; every arrival is drained non-blocking and, if this node steps toward perpendicular, it places a bead itself |
| Out | out | chain | THIS node's own goroutine places a bead here directly — never the mover |

## Firing rule

REACTIVE, not periodic — the "straightening loop", mirroring Node1. Each cycle of this
node's OWN Update loop non-blockingly drains two sources and, on either, decides and
acts itself — no round trip to any other goroutine to decide:

**On an In arrival** (`In.PollRecv`):

- Compute dot(tilt, coplanar normal) for THIS node — in practice, for this scene,
  decided as an INTEGER index compare (`TopTiltThetaIdx == Wiring.PerpendicularThetaIdx`,
  `nodes/Wiring/node_mover.go`'s exported constant), not a float dot product
  (`cos(π/2)` in float64 never equals exactly 0). This shortcut is valid ONLY because
  the ring plane contains world +y and θ is measured from +y, so the tilt's in-plane
  angle coincides with its θ index; see `Wiring.PerpendicularThetaIdx`'s doc comment for
  the assumption spelled out, including what breaks it.
- If not perpendicular: step `TopTiltThetaIdx` ONE click (`Wiring.CurveParamTiltVectorAngleStep`,
  π/12) toward perpendicular, report the new indices to this node's own mover
  (`SyncTiltIndex`, one-way/fire-and-forget — the mover streams+persists but never
  decides), and place a bead (value 1) on Out itself (`Out.PlaceDrivenAt`) — passing the
  exchange on to the partner.
- If already perpendicular: do nothing and send nothing. This is how the loop
  terminates, not a missed case.

The STOP condition above compares `TopTiltThetaIdx` directly against
`Wiring.PerpendicularThetaIdx` and never touches the DRAWN coplanar normal at all — the
drawn normal (below, "Coplanar normal") is a separate, purely visual derivation.

**On a TiltEditIn arrival** (`BuildArgs.TiltEditIn`, a panel-driven click routed HERE
instead of to the mover): apply the ±1 click to the named axis unconditionally (never a
step-toward-perpendicular decision — the user asked for exactly this move), sync, and
unconditionally place a bead on Out — "THE KICK". With both nodes of a pair
perpendicular, nothing circulates on In — correctly, since there is nothing left to
straighten — so the loop has no way to start on its own; it is kicked off by the thing
that actually moves a tilt away from perpendicular, which is this edit. This is sent
unconditionally because the edit always changed the index by one click, so there is
always something new for the exchange to react to.

Node2 is the mirror of Node1 (same firing rule), kept as a distinct
package/kind — not a parametrized Node1 — because a node-kind package may
import only the shared spine, never a sibling kind (check-dep-rules).

Pairing a Node1 and a Node2 with one edge running each direction (Node1.Out →
Node2.In, Node2.Out → Node1.In) needs no seed/bootstrap node: nothing sends until a
user tilt starts it, so there is no deadlock to bootstrap out of at t=0.

**RESET** (`TiltEditMsg.Reset`, the RESET button `TiltResetButton.tsx`): sets both
indices to 0, syncs the mover, places NO bead (stop-and-return, not a nudge), AND drains
any value already sitting on `VectorIn` (`Wiring.PollRecvVector`, non-blocking, on THIS
node's own goroutine — the goroutine that owns the receive end). Without the drain, a
vector already in flight when RESET was pressed would arrive on the very next cycle's
`handleVectorCycle` and immediately step the tilt again, undoing the reset a moment
later. `VectorIn` is depth-1 latest-wins, so one non-blocking receive empties it fully.

## Vector channel

Alongside the bead edges above, each directed edge between two vector-capable kinds
(today: Node1/Node2 only — `Wiring.KindWantsVectorChannel`) gets its OWN dedicated
node-to-node channel carrying `Wiring.TiltVectorMsg`, an integer θ/φ INDEX pair (never
floats on a channel). Buffered depth 1, latest-wins, both ends non-blocking
(`Wiring.SendVectorLatestNonBlocking` / `Wiring.PollRecvVector`) — same shape as the
speed-delivery channel. It travels additively: beads are unaffected, and this channel
never carries a bead value or vice versa.

Every cycle, this node's own `handleVectorCycle` (its whole per-cycle vector-channel
loop body) runs:

- **Coplanar normal**: a FIXED quarter turn (`Wiring.PerpendicularThetaIdx`, 6 steps of
  `Wiring.CurveParamTiltVectorAngleStep`, i.e. 90°) from THIS node's OWN tilt vector —
  pure index arithmetic (`theta-6`), never a cross product — so the normal turns WITH
  the tilt, always staying 90° away, rather than holding still toward the partner. φ is
  unchanged. Node2 SUBTRACTS the quarter turn; Node1 (its mirror package) ADDS it, same
  ± split as `stepTilt`'s add/subtract. This node's own goroutine computes it
  (`coplanarNormal`) and reports BOTH the tilt index and this normal index to its own
  mover in one call (`syncTiltIndex`, `SyncTiltIndex(theta, phi, normalTheta,
  normalPhi)`) every time either changes — the mover is a pure mirror that streams
  exactly what it is told as the buffer's `CoplanarNormalTheta`/`CoplanarNormalPhi`
  columns, never deriving a normal from the edge itself
  (`coplanarNormalTowardPartner` was removed).
- **Bottom tilt vector**: this node's TOP tilt vector turned a half turn (180°,
  `Wiring.HalfTurnThetaIdx`) in θ — Node2 subtracts it, its mirror package does the
  opposite. φ untouched. A half turn in θ alone negates the direction exactly in this
  parameterization, so both signs land in the SAME drawn direction and the sign is index
  bookkeeping only. It shares the top's length column (`TopTiltVectorLen`) and its colour;
  it is one of the two dot-product operands above.
- **What this node SENDS**: that coplanar normal rotated 180° in θ. Node2 turns +180°
  (+12 steps of π/12, i.e. `2 × PerpendicularThetaIdx` added); Node1 (its mirror
  package) turns −180° (−12 steps). φ is untouched. Index arithmetic only.
- **On receiving a vector**: FIRST, unconditionally, this node records the received
  direction as its own THIRD drawn vector (`ReceivedThetaIdx`/`ReceivedPhiIdx`/
  `ReceivedSet`, reported one-way to its own mover via `SyncReceivedVector` — same
  passive-mirror shape as `SyncTiltIndex`) — REPLACING whatever it received last time,
  regardless of whether the step below fires. THEN the step decision: TWO DOT PRODUCTS —
  the received vector against this node's own TOP tilt vector, and against its own BOTTOM
  tilt vector (`Wiring.TiltVectorIsAcute`, the sign of `Wiring.TiltVectorDot`). They decide
  BOTH questions, whether to move and which way:
  - acute with the TOP tilt: step `TopTiltThetaIdx` ONE click ADDING (+1) — Node2's base
    direction; its mirror package's base is the opposite, so a pair still turns
    symmetrically when both lean the same way.
  - acute with the BOTTOM tilt: step ONE click the REVERSE way (SUBTRACTING, −1).
  - neither acute: step nothing and send nothing — this is how the vector exchange stops,
    independently of whether the bead exchange has also stopped.

  If it stepped, sync the mover and send the outgoing vector above.
- **Why there is no both-acute case to arbitrate**: the bottom tilt is a half turn from the
  top, i.e. its exact antipode, so the two dots are always exact NEGATIVES of each other —
  at most one can be positive, and neither is positive only when the received vector is
  exactly perpendicular to the tilt axis. Which end the arrival leans toward IS the
  direction; there is no free sign knob and no ordering dependence between the two tests.
- **Perpendicular is a property of the ARRIVAL, not of where this node sits**: unlike the
  bead path's `stepTilt`, this never compares against `Wiring.PerpendicularThetaIdx`. A node
  sitting exactly at that index still steps if what arrived leans either way.
- **Float hazard, handled**: `cos(π/2)` in float64 is 6.1e-17, so an exactly-perpendicular
  arrival is NOT caught by a bare `dot > 0`. `Wiring.TiltVectorIsAcute` tests against a
  1e-9 band; both operands sit on the π/12 index lattice, so the dot is either ~1e-16 or at
  least `sin(15°)` = 0.2588, and every value in between classifies identically.
- **Received-vector RESET**: a Reset marker arriving on `VectorIn` zeroes this node's
  tilt (as above) AND clears its own received-vector record
  (`ReceivedSet = false`, synced) — a stale received arrow left hanging would
  contradict the reset's stop-and-return meaning. The LOCAL `RESET` path (`TiltEditIn`'s
  `Reset`, above) clears it the same way, for the same reason.

## Third vector (received direction)

Alongside its own tilt vector and the coplanar normal, this node draws a THIRD arrow:
the direction that last ARRIVED on its vector channel (`ReceivedThetaIdx`/
`ReceivedPhiIdx`, streamed as the buffer's `ReceivedVectorLen`/`ReceivedVectorTheta`/
`ReceivedVectorPhi` columns, `Buffer/layout.go`). It:

- Persists indefinitely once set — it is NOT cleared when the straightening exchange
  settles (i.e. neither dot is acute, so nothing steps and nothing is sent). An arrival is
  recorded even when it moves nothing: the last direction this node was sent is what it is
  still holding, and blanking the arrow when the pair comes to rest would erase the state
  it came to rest in.
- Is REPLACED, never accumulated, by the next arrival.
- Is cleared ONLY by a reset — this node's own (`TiltEditIn`'s `Reset`) or a Reset
  marker received on the channel — both zero `ReceivedSet`, and `ReceivedVectorLen`
  streams 0 in that state.
- Is distinguishable from "received (0,0)" (world +y): `ReceivedVectorLen` is 0 only
  when nothing has been received yet or a reset cleared it; an actually-received (0,0)
  direction still streams a non-zero length (this node's own radius, same as
  `TopTiltVectorLen`).
- Draws in its OWN colour (`RECEIVED_VECTOR_COLOR`, `TiltVectors.tsx`), distinct from
  the tilt vector/coplanar normal's shared magenta.

## Runtime status

- Loader-registered: yes
- TSX render: present

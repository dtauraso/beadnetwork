# Node1Node

## View

| Field | Value |
|-------|-------|
| kindId | 11 |
| kind | node1 |
| bg | #fff8e1 |
| border | #f9a825 |
| text | #4e342e |
| accent | #f9a825 |
| minWidth | 70 |
| shape | rect |
| fill | #fff8e1 |
| stroke | #f9a825 |
| width | 70 |
| height | 60 |

## Ports

| Name | Direction | EdgeKind | Notes |
|------|-----------|----------|-------|
| In | in | chain | sole input; every arrival is drained non-blocking and paces the exchange — it decides and places nothing itself |
| Out | out | chain | THIS node's own goroutine places a bead here directly, from `handleVectorCycle` when the acute tests actually move this node — never the mover |

## Firing rule

REACTIVE, not periodic — the "straightening loop". Each cycle of this node's OWN
Update loop non-blockingly drains two sources and, on either, decides and acts itself —
no round trip to any other goroutine to decide.

This node owns its own geometry directly (`Self *Wiring.PairNodeSelf`, claimed at build
time via `BuildArgs.ClaimSelfDrive`, driven every cycle by `n.Self.Step`
(`nodes/Wiring/pair_node_self.go`)) — there is **no separate `nodeMover` goroutine** for
Node1 or Node2 (`mover_registry.go`'s `finalizeActors` never constructs one for an id
claimed via `ClaimSelfDrive`). `SyncTiltIndex`/`SyncReceivedVector`/`ClearOutBeads` below
are therefore plain method calls on this same object (`PairNodeSelf.SetTiltIndex`/
`SetReceivedVector`/`ClearOutBeads`), not messages to another goroutine — they are named
and shaped like the old cross-goroutine sync calls only because nothing else about what
they carry changed when the separate mover was removed.

**On an In arrival** (`In.PollRecv`): fire this node's own trace and nothing else. A bead
PACES the exchange and marks the round trip; it decides nothing and places nothing onward.

It used to step `TopTiltThetaIdx` one click in this kind's own fixed direction (a retired
fixed-direction step), with no reference to anything that arrived, stopping only if it happened to
land exactly on `Wiring.PerpendicularThetaIdx` — which walking away from perpendicular never
does. That put TWO rules on one index: the fixed bead step and the vector channel's acute-test
rule. Where they agreed a node double-stepped; where they disagreed they cancelled and it
froze. Measured on the real formulas, a pair marched +1/+2/+3… one way forever on one click
direction and did not move at all on the other. The acute tests are now the only rule that turns a
tilt on an arrival.

The outgoing bead moved with the decision: it is placed by the vector branch when the tests
actually move this node (`handleVectorCycle`), so one message carries one visible bead and
the bead loop lives and dies with the exchange it paces.

`TiltEditIn` (`BuildArgs.TiltEditIn`, a panel-driven edit routed HERE instead of to a
mover) carries THREE distinct edits, applied by `applyTiltEdit` (`nodes/Node1/node.go`):

- **A ▲/▼ panel click** (`TiltVectorAnglePanel.tsx`): applies exactly one ±1 step to the
  named axis and STOPS — no send, no bead. This used to ALSO open the vector exchange as
  a side effect ("the kick"), so one click moved the tilt by many π/12 steps once the
  exchange settled instead of exactly one; that side effect is now START's alone.
- **START** (`TiltEditMsg.Start`, the START TILT button, `TiltVectorButtons.tsx`): opens
  the vector exchange from whatever angles are CURRENTLY set — sends this node's own
  `outgoingVector()` on `VectorOut` alongside a bead ("THE KICK"), which is what gives
  `handleVectorCycle` something to reply to. It changes NO index of its own. **START is
  Node1's alone** — Node2's `applyTiltEdit` ignores it, since the exchange is begun from
  one end only (starting from both would answer each other's opener in the same round).
  With both nodes of a pair perpendicular, nothing circulates on In — correctly, since
  there is nothing left to straighten — so the loop has no way to start on its own; START
  is what a user clicks to begin it.

  Pairing a Node1 and a Node2 with one edge running each direction (Node1.Out →
  Node2.In, Node2.Out → Node1.In) needs no seed/bootstrap node: nothing sends until a
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

## Vector channel

Alongside the bead edges above, each directed edge between two vector-capable kinds
(today: Node1/Node2 only — `Wiring.KindWantsVectorChannel`) gets its OWN dedicated
node-to-node channel carrying `Wiring.TiltVectorMsg`, a single integer θ INDEX (never
floats on a channel; there is no φ — every tilt vector in this exchange is θ-only, which
is what lets `Wiring.TiltVectorIsAcute` be an exact integer comparison instead of a dot
product against an epsilon). Buffered depth 1, latest-wins, both ends non-blocking
(`Wiring.SendVectorLatestNonBlocking` / `Wiring.PollRecvVector`) — same shape as the
speed-delivery channel. It travels additively: beads are unaffected, and this channel
never carries a bead value or vice versa.

Every cycle, this node's own `handleVectorCycle` (its whole per-cycle vector-channel
loop body) runs:

- **Coplanar normal**: a quarter turn (`Wiring.PerpendicularThetaIdx`, 6 steps of
  `Wiring.CurveParamTiltVectorAngleStep`, i.e. 90°) from THIS node's OWN tilt vector —
  pure index arithmetic (`theta+6`), never a cross product — so the normal turns WITH
  the tilt, always staying 90° away, rather than holding still toward the partner. There
  is no φ. Node1 ADDS the quarter turn; Node2 (its mirror package) SUBTRACTS it, same
  ± split as every other per-kind sign here. On top of that base quarter turn,
  `coplanarNormal` (`nodes/Node1/node.go`) ALSO adds a half turn (`Wiring.HalfTurnThetaIdx`)
  whenever `TopTiltThetaIdx` has crossed an ODD number of poles (`floorDiv(TopTiltThetaIdx,
  Wiring.HalfTurnThetaIdx) % 2 != 0`, FLOOR division, not Go's truncating `/`, since
  Node1's base direction subtracts and so runs negative for most of this node's life) — θ
  is measured from world +y, so each half turn crossed is a pole crossing, and without
  this correction the normal would visibly flip to the wrong side once the tilt passes a
  pole. This is a PURE function of `TopTiltThetaIdx` alone — no stored "did we just cross"
  flag, no comparison against a previous value. This node's own goroutine computes the
  normal (`coplanarNormal`) and reports the tilt index, the normal index, AND the bottom
  tilt index to its own geometry in one call (`syncTiltIndex`, `SyncTiltIndex(theta,
  normalTheta, bottomTheta)`) every time any of them changes. That
  is a plain method call on this node's OWN goroutine, not a message to a second one:
  there is no `nodeMover` for this kind, and `PairNodeSelf.SetTiltIndex` sets the
  geometry's mirror fields directly. The geometry stays a pure mirror — it streams exactly
  what it is told as the buffer's `CoplanarNormalTheta` column and
  never derives a normal from the edge itself (`coplanarNormalTowardPartner` was removed).
- **Bottom tilt vector**: this node's TOP tilt vector turned a half turn (180°,
  `Wiring.HalfTurnThetaIdx`) in θ — Node1 adds it, its mirror package does the
  opposite. There is no φ. A half turn in θ alone negates the direction exactly in this
  parameterization, so both signs land in the SAME drawn direction and the sign is index
  bookkeeping only. It shares the top's length column (`TopTiltVectorLen`) and its colour;
  it is one of the two acute-test operands above.
- **What this node SENDS**: that coplanar normal rotated 180° in θ. Node1 turns −180°
  (−12 steps of π/12, i.e. `2 × PerpendicularThetaIdx` subtracted); Node2 (its mirror
  package) turns +180° (+12 steps). Index arithmetic only.
- **On receiving a vector**: FIRST, unconditionally, this node records the received
  direction as its own THIRD drawn vector (`ReceivedThetaIdx`/
  `ReceivedSet`, reported to its own geometry via `SyncReceivedVector` — same
  passive-mirror shape as `SyncTiltIndex`, and likewise a direct call rather than a
  message) — REPLACING whatever it received last time,
  regardless of whether the step below fires. THEN the step decision: TWO ACUTE TESTS —
  the received vector against this node's own TOP tilt vector, and against its own BOTTOM
  tilt vector (`Wiring.TiltVectorIsAcute`, an integer comparison on the π/12 index lattice,
  no dot product and no epsilon — see that function). They decide
  BOTH questions, whether to move and which way:
  - acute with the TOP tilt: step `TopTiltThetaIdx` ONE click SUBTRACTING (−1) — Node1's base
    direction; its mirror package's base is the opposite, so a pair still turns
    symmetrically when both lean the same way.
  - acute with the BOTTOM tilt: step ONE click the REVERSE way (ADDING, +1).
  - neither acute: step nothing and send nothing — this is how the vector exchange stops,
    independently of whether the bead exchange has also stopped.

  If it stepped, report the new indices to its own geometry (`syncTiltIndex`) and send the
  outgoing vector above, alongside the bead.
- **Why there is no both-acute case to arbitrate**: the bottom tilt is a half turn from the
  top, i.e. its exact antipode, so the two tests are exact opposites of each other — at
  most one can pass, and neither passes only when the received vector sits exactly
  `PerpendicularThetaIdx` from the tilt axis. Which end the arrival leans toward IS the
  direction; there is no free sign knob and no ordering dependence between the two tests.
- **Perpendicular is a property of the ARRIVAL, not of where this node sits**: unlike the
  retired bead-path rule, this never compares against `Wiring.PerpendicularThetaIdx`. A node
  sitting exactly at that index still steps if what arrived leans either way.
- **No float hazard to handle**: there is no dot product here at all. Since the tilt vector
  lost its φ, both operands are single θ indices on the same 24-step lattice, and
  `Wiring.TiltVectorIsAcute` is the integer comparison `d < 6 || d > 18` on their wrapped
  difference (`nodes/Wiring/tilt_vector_channel.go`). The exactly-perpendicular case is the
  exact integers 6 and 18, decided exactly — not `cos(π/2)` landing at 6.1e-17 and needing
  an epsilon band to be classified as "not acute", which is what this test used to be.
- **Received-vector RESET**: a Reset marker arriving on `VectorIn` zeroes this node's
  tilt (as above) AND clears its own received-vector record
  (`ReceivedSet = false`, synced) — a stale received arrow left hanging would
  contradict the reset's stop-and-return meaning. The LOCAL `RESET` path (`TiltEditIn`'s
  `Reset`, above) clears it the same way, for the same reason.

## A tilt does not move the node

Turning a tilt changes an ANGLE and nothing else. The pair's separation is whatever a drag
last left it at, and it holds while the tilts turn.

It was briefly otherwise: the distance was a second reading of this same index, so Node2
slid along its own ray as the exchange ran and the edge grew and shrank with the angle.
That coupled four things to one number — the drawn angle, the separation, the edge's bead
count, and (once the exchange became clock-paced) the rate the tilts turned at, since a
longer edge takes longer to cross. One index is not enough to carry all of that.

## Pacing and clock speed

The clock a bead's round trip is measured in (`clk.Tick()`) and the clock that PACES this
node's own cycle (`clk.SleepCycle`, `nodes/wire/clock.go`) are scaled by the scene's
playback speed: `SleepCycle` waits `pulsesPerCycle()` broadcaster pulses, `1/speed`
rounded up and clamped to `[1, 64]`, so one cycle is one SCALED tick's worth of wall time
and this goroutine itself runs slower when the slider says slower. The bead marks the
round trip; the CYCLE — how often this node's loop drains its channels and can react at
all — is what the dial actually paces.

While a ▲/▼ panel click is being applied, every clock in the scene is broadcast to
`Wiring.HumanEditSpeed` (1.0, unscaled) instead of the slider's speed
(`applyUpdateTiltVector`, `nodes/Wiring/stdin_reader.go`), so a click is answered on the
next real-time cycle rather than sitting unanswered for a scaled cycle (up to ~1 second at
a slow divisor). START and RESET both restore the slider's own speed
(`md.SliderSpeed()`) before doing anything else — START because running the exchange is
exactly what the slider's number is about, RESET because setting is over. The slider's own
persisted number (view/speed.json, per topology) is untouched throughout; only what is broadcast to
the clocks changes.

## Third vector (received direction)

Alongside its own tilt vector and the coplanar normal, this node draws a THIRD arrow:
the direction that last ARRIVED on its vector channel (`ReceivedThetaIdx`/
streamed as the buffer's `ReceivedVectorLen`/`ReceivedVectorTheta` columns,
`Buffer/layout.go`). It:

- Persists indefinitely once set — it is NOT cleared when the straightening exchange
  settles (i.e. neither test is acute, so nothing steps and nothing is sent). An arrival is
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

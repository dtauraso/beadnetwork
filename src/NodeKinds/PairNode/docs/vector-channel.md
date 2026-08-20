# PairNode — vector channel

[← BEHAVIOR.md](BEHAVIOR.md)

Alongside the bead edges, each directed edge between two vector-capable kinds
(today: PairNode only — `Wiring.KindWantsVectorChannel`) gets its OWN dedicated
node-to-node channel carrying `Wiring.TiltVectorMsg`, a single integer θ INDEX (never
floats on a channel; there is no φ — every tilt vector in this exchange is θ-only, which
is what lets the acute test be a walk along a ring of 24 states instead of a dot product
against an epsilon). Buffered depth 1, latest-wins, both ends non-blocking
(`Wiring.SendVectorLatestNonBlocking` / `Wiring.PollRecvVector`) — same shape as the
speed-delivery channel. It travels additively: beads are unaffected, and this channel
never carries a bead value or vice versa.

Every cycle, this node's own `handleVectorCycle` (its whole per-cycle vector-channel
loop body) runs:

- **Coplanar normal**: a quarter turn (`Wiring.QuarterTurnPhiIdx`, 6 steps of
  `Wiring.CurveParamTiltVectorAngleStep`, i.e. 90°) from THIS node's OWN tilt vector —
  pure index arithmetic (`theta+6`), never a cross product — so the normal turns WITH
  the tilt, always staying 90° away, rather than holding still toward the partner. There
  is no φ. Both nodes of a pair run this same unmodified addition — there is no per-node
  sign. `coplanarNormal` (`src/NodeKinds/PairNode/vectors.go`) reads it straight off the tilt's own
  `quarter` link on the ring (`src/NodeKinds/PairNode/ring.go`) — the ring is built with that link
  already wrapped onto `0…Wiring.FullTurnThetaIdx-1`, so there is no addition here to
  overflow and nothing to subtract back into range. There is no pole and nothing to cross:
  the renderer decodes an index as `(sin θ, cos θ, 0)`, a plain circle, so there is no φ to
  flip and no parity term. This is a PURE function of the tilt state's own `quarter` link
  alone — no stored "did we just cross" flag, no comparison against a previous value. This
  node's own goroutine computes the
  normal (`coplanarNormal`) and reports the tilt index, the normal index, AND the bottom
  tilt index to its own geometry in one call (`syncTiltIndex`, `SyncTiltIndex(theta,
  normalTheta, bottomTheta)`) every time any of them changes. That
  is a plain method call on this node's OWN goroutine, not a message to a second one:
  there is no `nodeMover` for this kind, and `PairNodeSelf.SetTiltIndex` sets the
  geometry's mirror fields directly. The geometry stays a pure mirror — it streams exactly
  what it is told as the buffer's `CoplanarNormalTheta` column and
  never derives a normal from the edge itself (`coplanarNormalTowardPartner` was removed).
- **Bottom tilt vector**: this node's TOP tilt vector turned a half turn (180°,
  `Wiring.HalfTurnThetaIdx`) in θ — both nodes of a pair add it, unmodified. There is no φ.
  A half turn in θ alone negates the direction exactly in this
  parameterization, so this is index bookkeeping only. It shares the top's length column (`TopTiltVectorLen`) and its colour;
  it is one of the two acute-test operands.
- **What this node SENDS**: that coplanar normal. The message on the channel IS the direction
  this node computed and draws, so the partner's received arrow coincides with this node's
  own normal on screen. Nothing rotates it on the way out: a rotation has to be undone by the
  receiver's step signs to leave behaviour unchanged, and a half turn in particular cannot
  move where the pair comes to rest, since the bottom tilt is the top plus that same half
  turn. `docs/pair-node/math/vectors/index.html`.

See [tilt-machine.md](tilt-machine.md) for what happens on receiving a vector.

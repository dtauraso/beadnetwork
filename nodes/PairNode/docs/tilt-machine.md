# PairNode — tilt machine (receiving a vector)

[← BEHAVIOR.md](BEHAVIOR.md) · [← vector-channel.md](vector-channel.md)

**On receiving a vector**: FIRST, unconditionally, this node records the received
direction as its own THIRD drawn vector (`ReceivedThetaIdx`/
`ReceivedSet`, reported to its own geometry via `SyncReceivedVector` — same
passive-mirror shape as `SyncTiltIndex`, and likewise a direct call rather than a
message) — REPLACING whatever it received last time,
regardless of whether the step below fires. THEN the step decision (`stepFromVector`):
THERE IS ONE TARGET AND ONE HALT. The arrival is the partner's normal,
already a quarter turn off the partner's tilt, so the ANGLE LENGTH between it and this node's
own TOP says what the two TILTS are doing. The angle length (`AngleLength`) is how many ring
slots lie between two directions, counted the short way round, so never more than a half turn:

The rule counts from the NEARER END of the node's own tilt line. A node draws two ends a half
turn apart, so there are two counts to the arrival and exactly one is under a half turn; that one
is the count, and it lives on a ring of a half turn.

| count | the two tilts | mode that stops here |
| --- | --- | --- |
| a quarter turn | on one line, either way round | PARALLEL |

That is a MODE of the one state machine (`nodes/PairNode/machine.go`), and a mode is nothing but
the counts it STOPS at: `{ quarter }` for parallel, a row in `stoppingCounts`. The rule that turns
toward it is written once and never asks which mode it is running for.

There is a SECOND mode, `setting`, for a node that has not started yet — before the first arrival,
and again after a reset. It is a mode and not an absence of one: it
stops at EVERY count, so it is already at rest wherever it stands and an arrival moves it
nothing, by the ordinary rule rather than by an exemption from it. Its pair-wide choice is
`TiltMachineNone`, so a node in it tells the other end nothing, which is all it has to say.

ONE ARRIVAL ANSWERS TWO YES-OR-NO QUESTIONS. Is this count one the mode stops at
(`settled`)? If so the node stays put and sends nothing. If not, is that stop nearer going up the
count-ring or down it (`step`, and a tie turns up)? Then it turns ONE slot. Nothing
is remembered between arrivals: no distance is stored, no walk is planned, and the next arrival
re-derives all of it. NO DISTANCE IS COMPUTED EITHER — the first question is a comparison and the
second is one subtraction against a quarter turn, so neither answer is a length.

A node holds ONE tilt machine (`tiltring.Machine`), whose only state is whether it has started — the same
`Wiring.TiltMachine` value the two ends say to each other. Its zero value is `setting`, so a node
starts there with nothing to construct and there is no nil to test for.

IT STARTS AT THE FIRST ARRIVAL, which is START, the moment the setup is finished. Nothing is read
out of the gap to decide WHICH machine: there is one. `adoptMachine` sets it and it STICKS until
reset. A click landing once the machine is running is a jitter — the thing
the running machine exists to correct — not a new instruction about what the pair is for.
Starting at a CLICK instead read a gap of one step on the first of eleven and locked the pair
while the tilt was still on its way.

The end that did not start first learns the same answer from the first reply: every vector message
carries which machine its sender is running (`TiltVectorMsg.Machine`), and adopting sticks, so
a later message cannot restart a running machine.

- running nothing yet (before any click, or after a reset): an arrival moves nothing.
- running, and the arrival is its halt: stand still, reply anyway.
- otherwise: ONE `step`.

`Machine` is cleared only by a reset (a clean slate), never by an arrival.

If it turned (or answered while square), report the new indices to its own geometry
(`syncTiltIndex`) and send the outgoing vector, alongside the bead.

- **No float hazard to handle**: there is no dot product here at all. Both operands are
  single θ indices on the same lattice, and `miss` measures the distance to the mode's nearest
  home in integer ring hops — not `cos(...)` landing near zero and needing an epsilon
  band.
- **Received-vector RESET**: a Reset marker arriving on `VectorIn` zeroes this node's
  tilt AND clears its own received-vector record
  (`ReceivedSet = false`, synced) — a stale received arrow left hanging would
  contradict the reset's stop-and-return meaning. The LOCAL `RESET` path (`TiltEditIn`'s
  `Reset`, see [firing-rule.md](firing-rule.md)) clears it the same way, for the same reason.

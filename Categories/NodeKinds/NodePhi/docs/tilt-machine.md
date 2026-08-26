# NodePhi — tilt machine (receiving a vector)

[← BEHAVIOR.md](BEHAVIOR.md) · [← vector-channel.md](vector-channel.md)

**On receiving a vector**: FIRST, unconditionally, this node records the received
direction as its own THIRD drawn vector (`ReceivedThetaIdx`/
`ReceivedSet`, reported to its own geometry via `SyncReceivedVector` — same
passive-mirror shape as `SyncTiltIndex`, and likewise a direct call rather than a
message) — REPLACING whatever it received last time,
regardless of whether the step below fires. THEN the step decision (`stepFromVector`).

THE RULE IS THE ARITHMETIC, and it is written once in `Categories/NodeKinds/NodePhi/tiltring/rules.go` as plain
functions over numbers — no state machine dispatch, no ring links, no receivers:

    tau              the lattice size, r.Points
    top, arrival     0 .. tau-1
    bottom           Mod(top + tau/2, tau)
    distanceTop      Abs(top - arrival)
    distanceBottom   Abs(bottom - arrival)
    offset            0  if distanceTop = 0 or distanceBottom = 0
                     -1  if distanceTop    < tau/4
                     +1  if distanceBottom < tau/4
                      0  otherwise
    topNext          Mod(top + offset, tau)
    bottomNext       Mod(bottom + offset, tau)
    sent             Mod(top + tau/4, tau)

A node draws two ends a half turn apart, so an arrival has two distances — one from each
end. WHICHEVER END IS ACUTE names the direction to turn: the top gives -1, the bottom
gives +1, and when neither is inside a quarter turn the node is where it belongs and does
not move. That zero IS the halt; there is no separate settled test and no stopping count
to compare against.

AN EXACT HIT IS ALSO ZERO, and it is tested FIRST: a distance of 0 is inside a quarter
turn, so without that case an arrival landing squarely on an end would step away from it
and the pair would never come to rest on the one line it just reached.

`Mod` is the non-negative remainder, not Go's `%`, and it is the ONLY modulus in the rule —
every `mod` on the framework page is this one function. `Abs` stands where the page writes
bars. Nothing else in the rule does arithmetic.

THE NODE HOLDS ONE END. `bottom` is derived by the rule wherever it is needed rather than
stored beside the top, so the two cannot disagree; a step writes the top and the bottom
follows because both are `+ offset` on the same ring.

A node holds ONE tilt machine (`tiltring.Machine`), whose only state is whether it has
started — the same `Wiring.TiltMachine` value the two ends say to each other. Its zero value
is `setting`, so a node starts there with nothing to construct and there is no nil to test
for. A node still setting up has no rule to run, so its offset is 0 wherever it stands and
an arrival moves it nothing — by the ordinary rule rather than by an exemption from it.

IT STARTS AT THE FIRST ARRIVAL, which is START, the moment the setup is finished.
`adoptMachine` sets it and it STICKS until reset. A click landing once the machine is
running is a jitter — the thing the running machine exists to correct — not a new
instruction about what the pair is for.

The end that did not start first learns the same answer from the first reply: every vector
message carries which machine its sender is running (`TiltVectorMsg.Machine`), and adopting
sticks, so a later message cannot restart a running machine.

`Machine` is cleared only by a reset (a clean slate), never by an arrival.

If it turned (or answered while square), report the new indices to its own geometry
(`syncTiltIndex`) and send the outgoing vector, alongside the bead.

- **Where it ends up**: every opening settles, and every settled pair is on ONE line —
  which is what parallel means. On a 24-point lattice no opening takes more than 20
  messages; on a 12-point one, no more than 8.
- **No float hazard to handle**: there is no dot product here at all. Both operands are
  single θ indices on the same lattice, and the rule compares integer distances against a
  quarter turn — not `cos(...)` landing near zero and needing an epsilon band.
- **Received-vector RESET**: a Reset marker arriving on `VectorIn` zeroes this node's
  tilt AND clears its own received-vector record
  (`ReceivedSet = false`, synced) — a stale received arrow left hanging would
  contradict the reset's stop-and-return meaning. The LOCAL `RESET` path (`TiltEditIn`'s
  `Reset`, see [firing-rule.md](firing-rule.md)) clears it the same way, for the same reason.

# Model — the φ, θ card nodes

[← MODEL.md](../../MODEL.md)

## Exception — the φ, θ card nodes

The card node kinds (`NodePhiTheta3`, the "pair φ, θ 3" tab, cards in
`docs/pair-node/phi-theta-3/`) do NOT follow MODEL.md's bead model:

- **No beads, no bead lines.** Each node runs its own goroutine and sends data to the
  nodes it connects to over channels between them.
- **One round:** every node processes its own angle data; then every node finds its
  global vector on the scene sphere (the cartesian conversion), using the channels to
  send each other the data it needs; then the new global positions are rendered.
- **Example — a triangle:** all three nodes have channels to the two they connect to.
  All three angles are computed, then all three scene-sphere global vectors are found.
  Once rendered, all three nodes are in new positions; with the same angle change on
  each, the triangle looks rotated.

The cards are the source for these kinds; MODEL.md's bead, bead-line and sending
sections do not apply to them.

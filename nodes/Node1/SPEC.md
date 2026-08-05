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
| In | in | chain | sole input; receiving never gates or paces this node's own sending |
| Out | out | chain | periodically emits a white bead (value 1), unconditionally, at human speed |

## Firing rule

Self-paced source, one goroutine, no bootstrap/deadlock dependency on any input.
Periodically (once per Out edge's own crossing time, so beads never overlap on
the wire) places a white bead (value 1 — bead-style.ts's on-wire
`VALUE_BEAD_STYLE`: 0 = black, 1 = white) on Out, unconditionally — this is NOT
a reply to anything received on In. In is drained non-blocking each cycle but
otherwise ignored by the send loop; a value arriving on it only calls Fire, it
never advances or delays the next Out emission.

Pairing a Node1 and a Node2 with one edge running each direction (Node1.Out →
Node2.In, Node2.Out → Node1.In) needs no seed/bootstrap node: each side's
emission is unconditional, so there is nothing to deadlock on at t=0.

## Runtime status

- Loader-registered: yes
- TSX render: present

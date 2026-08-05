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

Node2 is the mirror of Node1 (same firing rule), kept as a distinct
package/kind — not a parametrized Node1 — because a node-kind package may
import only the shared spine, never a sibling kind (check-dep-rules).

Pairing a Node1 and a Node2 with one edge running each direction (Node1.Out →
Node2.In, Node2.Out → Node1.In) needs no seed/bootstrap node: each side's
emission is unconditional, so there is nothing to deadlock on at t=0.

## Runtime status

- Loader-registered: yes
- TSX render: present

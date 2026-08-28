# NodePhiTheta3

The pair φ, θ 3 tab's node. THREE of them, each facing the other two, so a node receives TWO
arrivals per round instead of one. That is the whole difference from `NodePhiTheta`, and it
is not a small one: the rule there reads a single arrival and answers with one offset, and
with two arrivals a node must decide what the pair of them means together — which hemisphere
each sits in, which distance is longer, and what that combination should do. Those cases are
being worked out on the card; this kind carries the wiring and the round, not the answer.

The frame is unchanged and stays unchanged: `top` is index 0 on each ring, the node's +Y pole
from `polar.WorldAxisPole`, `bottom` a half turn round, and both are an AXIS REFERENCE that
never moves. Each angle still lives on its own ring with its own whole turn. Nothing crosses
between φ and θ.

The exchange is unchanged in kind and different in shape. Three nodes with two partners each
are SIX ordered pairs, so six unbuffered channels, and a round is four operations rather than
two: send to both partners, receive from both. Every one of them is a rendezvous, and the
round's select must stay willing to perform ANY operation still outstanding — a node that
commits to finishing its sends before its receives can deadlock a cycle of three, where two
could not.

What it sends is its own position, the same value to both partners. Each partner receives it
as one of that partner's two arrivals.

## Description

One of three φ, θ nodes: receives an arrival from each of its two partners, and sends its own
position on to both as their next arrivals.

## View

| Field | Value |
|-------|-------|
| kindId | 13 |
| kind | nodePhiTheta3 |
| bg | #e3f2fd |
| border | #1565c0 |
| text | #0d2b4a |
| minWidth | 70 |
| shape | rect |
| fill | #e3f2fd |
| stroke | #1565c0 |
| width | 70 |
| height | 60 |

## Ports

None, for the same reason as `NodePhiTheta`: a port is where a bead line attaches, and
nothing is placed on these edges. What crosses is the sent vector, on the channel the edge
allocates. The edges exist — they are the radius vectors drawn between the three centers —
but they bind no port, so the table is deliberately empty.

| Name | Direction | EdgeKind | Notes |
|------|-----------|----------|-------|

## Runtime status

- Loader-registered: yes
- TSX render: present

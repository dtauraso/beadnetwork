# NodePhiTheta

The pair φ, θ tab's node. Two of them face each other across one channel each way, carrying
the sphere card's arithmetic on BOTH angles — each angle on its OWN ring, with its own whole
turn, its own ends and its own condition. Nothing crosses between them.

`top` is the node's POLE — the +Y arrow of the frame drawn around it, which is what
`polar.WorldAxisPole` names — so `top` is index 0 on each ring and `bottom` is the −Y tip a
half turn round. The four multiples of the quarter turn are that frame's tips. It is an AXIS
REFERENCE and never moves.

The radius vector is DECOUPLED from that axis, which is what lets a tilt be identified at all
— coupled, it would already sit on a defined point and always read as locked. Decoupled, an
arrival can be obtuse from one view and acute from the other, so both distances are measured
the SHORT way round the ring. Since top and bottom are a half turn apart those two always sum
to τ/2, so one of them is always acute whatever the other node's center is. The phi kind can
take the plain difference instead because its normal is generated from its top vector; that
assumption does not survive the decoupling here.

The ARRIVAL says which way to turn; the angle's OWN center says when to stop. An angle rests
once its own center is a quarter turn from its axis, so it does not have to be at rest in the
same cycle as anything else — phi and theta reach that point cycles apart, and with the halt
read off the arrival alone neither could hold the state long enough for the other to join it.
A stopped center does not drift, so the rest survives the partner moving on.

A step adds the angle's offset to that angle of the center and wraps at that angle's QUARTER
turn, inclusive — `(c + offset) mod (τ/4 + 1)`, as the card writes it. The `+1` is what puts
the quarter turn itself in range: without it the value stops at τ/4 − 1 and the one state the
rule can rest in is unreachable. `r` is carried through untouched: the rule never reads or
changes it.

What it sends is the center it just stepped to. The partner receives that as its own arrival.

## Description

One half of a φ, θ pair: adds each angle's offset to that angle of its center, wrapping at
that angle's quarter turn inclusive, and sends the stepped center on as the partner's next
arrival.

## View

| Field | Value |
|-------|-------|
| kindId | 12 |
| kind | nodePhiTheta |
| bg | #e8f5e9 |
| border | #2e7d32 |
| text | #1b3c1e |
| minWidth | 70 |
| shape | rect |
| fill | #e8f5e9 |
| stroke | #2e7d32 |
| width | 70 |
| height | 60 |

## Ports

None. A port is where a bead line attaches, and nothing is ever placed on this pair's edges:
what crosses is the sent vector, on the channel the pair's edge allocates. The edge still
exists — it is the radius vector drawn from one center to the other — but it binds no port,
so the table is deliberately empty and `BindPorts` has nothing to walk.

| Name | Direction | EdgeKind | Notes |
|------|-----------|----------|-------|

## Runtime status

- Loader-registered: yes
- TSX render: present

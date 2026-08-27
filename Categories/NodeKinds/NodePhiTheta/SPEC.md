# NodePhiTheta

The pair φ, θ tab's node. Two of them face each other across one channel each way, carrying
the sphere card's arithmetic on BOTH angles — each angle on its OWN ring, with its own whole
turn, its own ends and its own condition. Nothing crosses between them.

`top` is the node's POLE — the +Y arrow of the frame drawn around it, which is what
`polar.WorldAxisPole` names — so `top` is index 0 on each ring and `bottom` is the −Y tip a
half turn round. The four multiples of the quarter turn are that frame's tips.

A step adds the angle's offset to that angle of the center and wraps at that angle's QUARTER
turn — `(c + offset) mod τ/4`, as the card writes it. `r` is carried through untouched: the
rule never reads or changes it.

What it sends is the center it just stepped to. The partner receives that as its own arrival.

## Description

One half of a φ, θ pair: adds each angle's offset to that angle of its center, wrapping at
that angle's quarter turn, and sends the stepped center on as the partner's next arrival.

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

# NodePhiTheta

The pair φ, θ tab's node. Two of them face each other across one channel each way and walk
around each other on the polar lattice, carrying the sphere card's arithmetic on BOTH angles
at once — the same rule per row, φ over θ, with nothing crossing between the rows.

Everything it holds is a POLAR COLUMN VECTOR: center, top, bottom, arrival, offset and sent are
all `(φ, θ, r)`. A step is vector addition — `center + offset` — and what it sends is the
center it just stepped to. The partner receives that vector as its own arrival.

## Description

One half of a φ, θ pair: adds the offset its arrival calls for to its own center, one step per
arrival, and sends the stepped center on as the partner's next arrival.

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

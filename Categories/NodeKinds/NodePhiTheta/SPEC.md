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

| Name | Direction | EdgeKind | Notes |
|------|-----------|----------|-------|
| In | in | chain | the one edge from the partner; its beads pace the exchange and carry nothing the rule reads |
| Out | out | chain | the one edge to the partner; this node's own goroutine places a bead here each time it steps |

## Runtime status

- Loader-registered: yes
- TSX render: present

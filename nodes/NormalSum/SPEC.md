# NormalSumNode

## View

| Field | Value |
|-------|-------|
| kindId | 12 |
| kind | NormalSum |
| bg | #ede7f6 |
| border | #4527a0 |
| text | #311b92 |
| accent | #4527a0 |
| minWidth | 70 |
| shape | rect |
| fill | #ede7f6 |
| stroke | #4527a0 |
| width | 70 |
| height | 60 |

## Ports

| Name | Direction | EdgeKind | Notes |
|------|-----------|----------|-------|
| NormalA | in | chain | one source node's normal, as a θ lattice INDEX |
| NormalB | in | chain | the other source node's normal, same units |
| Out | out | chain | the total, emitted as a θ index, so a downstream node can take it as a normal in turn |

## Firing rule

Holds two received normals and their TOTAL.

A normal here is a θ LATTICE INDEX, not an angle and not a vector: the whole tilt-vector
model is θ-only integer indices times one step constant
(`Wiring.CurveParamTiltVectorAngleStep`), and every derived direction in it — bottom is θ+12,
the coplanar normal is θ+6 — is index arithmetic. Taking a normal as an index keeps this node
inside that arithmetic instead of introducing a second spelling of a direction.

The total is `(a + b) mod points`, the same wrap the lattice itself has. Sum, not average: two
normals pointing the same way should read as twice as far around, and an average would make
this node's output depend on how many inputs it happened to have received rather than on what
they said.

Each input is held as it arrives, so a total exists as soon as BOTH have arrived once; before
that the node holds no total and draws no arrow. It emits the total on Out whenever it
changes, and its own drawn tilt vector IS that total — the arrow every node already draws
from its `TopTiltVectorTheta` column, pointed along the sum.

## Runtime status

- Loader-registered: yes
- TSX render: present

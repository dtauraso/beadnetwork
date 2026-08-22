# PairNode

The PAIR TAB'S node — the thing that scene is made of, two of them exchanging vectors until
they come to rest. It was called PairNode, which named nothing: not what it does, not what it is
for, and not even which node it is (both nodes of a pair are this kind). kindId stays 11; ids
are assigned once and never renumbered, and the rename does not touch identity.

## Description

One half of a pair: turns its own tilt vector toward rest by exchanging directions with its
partner, one step per arrival.

## View

| Field | Value |
|-------|-------|
| kindId | 11 |
| kind | pairNode |
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
| In | in | chain | sole input; every arrival is drained non-blocking and paces the exchange — it decides and places nothing itself |
| Out | out | chain | THIS node's own goroutine places a bead here directly, from `handleVectorCycle` when `stepFromVector` actually moves this node — never the mover |

See [BEHAVIOR.md](./docs/BEHAVIOR.md) for the firing rule, the vector channel, why a tilt does not
move the node, pacing/clock speed, and the third (received-direction) vector — none of
which `Categories/NodeKinds/gen/kindscan` parses.

## Runtime status

- Loader-registered: yes
- TSX render: present

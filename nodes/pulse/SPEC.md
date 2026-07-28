# PulseNode

## View

| Field | Value |
|-------|-------|
| kindId | 5 |
| kind | pulse |
| bg | #e1f5fe |
| border | #2196f3 |
| text | #01579b |
| accent | #2196f3 |
| minWidth | 90 |
| shape | rect |
| fill | #e1f5fe |
| stroke | #2196f3 |
| width | 90 |
| height | 60 |

## Ports

| Name | Direction | EdgeKind | Notes |
|------|-----------|----------|-------|
| FromInput | in | chain | sampled input value; updates the held value |
| Out | out | chain | continuously drives the held value (starts -1) |
| Out2 | out | chain | optional second continuous output of the same held value (fan to a second destination); inert when unwired |

## Firing rule

Sample-and-hold. Holds one int value (initialized to -1) and drives it out
continuously, even before any input arrives. When a value arrives on FromInput,
the held value is updated and subsequent outputs emit the new value. The output
is not precondition-gated — Pulse self-emits -1 from the start.

## Cascade delta rule

Layout-side only (`nodes/Wiring/node_mover.go`), independent of the firing rule:

- A Pulse **IGNORES** a delta triple arriving from a **TimeStart** cascade neighbor —
  no record, no relay. Same shape as the TimeStart←Input stop.
- A Pulse **ROUTES** a gate-origin delta straight across to the opposite gate kind,
  and nowhere else:

  | From sender kind | Go type | Routed to | Go type |
  |---|---|---|---|
  | `WindowAndInhibitLeftGate` | SelectRight | `WindowAndInhibitRightGate` | SelectLeft |
  | `WindowAndInhibitRightGate` | SelectLeft | `WindowAndInhibitLeftGate` | SelectRight |

  Note the kind strings cross over relative to the Go type names — read the table,
  not the name.
- From every other sender kind it keeps the plain flood to all cascade neighbors
  except the sender.

This is the `Pulse` kind only (node 5). `PulseLeft` (node 3) and `PulseRight` (node 7)
are separate kinds and are both termini with their own sender whitelists — see their
SPECs.

Node 5's cascade neighbors are `{2: TimeStart, 9: WindowAndInhibitRightGate,
8: WindowAndInhibitLeftGate}` — the `5-8` double link is restored, which is what lets a
Pulse classify node 8 as SelectRight at all.

## Runtime status

- Loader-registered: yes
- TSX render: present

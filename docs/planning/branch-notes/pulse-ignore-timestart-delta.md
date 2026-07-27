---
branch: task/pulse-ignore-timestart-delta
---

# Pulse ignores a delta triple from a TimeStart

## Rule

A `Pulse` node IGNORES a cascade delta triple arriving from a `TimeStart`-kind cascade
neighbor: no record, no relay. From every other sender kind it keeps the plain flood to
all cascade neighbors except the sender.

Implemented in the `moveMsgKindDeltaForward` handler (`nodes/Wiring/node_mover.go`),
the same shape and the same place as the existing TimeStart<-Input stop.

## Scope

`Pulse` only — node 5. `PulseLeft` (3) and `PulseRight` (7) are separate Go kinds with
their own rules (both termini, each with its own sender whitelist). A test pins that
the Pulse rule does not leak into them.

Node 5's cascade neighbors: `{2: TimeStart, 9: WindowAndInhibitRightGate}`, plus
`{8: WindowAndInhibitLeftGate}` once the `5-8` double link is restored
(`task/cascade-double-links`).

## Where the cascade rules now stand

| Node | Kind | Attends to | Relays? |
|---|---|---|---|
| 2 | TimeStart | Pulse, Time, Input | routes by sender kind: Pulse->Time, Time->Pulse; Input-origin ignored |
| 3 | PulseLeft | Input, SelectRight | never (terminus) |
| 5 | Pulse | everything except TimeStart | plain flood |
| 6 | TimeEnd | — | never (terminus) |
| 7 | PulseRight | Time, SelectLeft | never (terminus) |

Everything else is still on the default plain flood.

## Verification

`TestPulseIgnoresTimeStartOriginDelta` fails without the change on both halves
(`gotForwardMsg=1, want 0` and `relayed to [9 8], want none`).
`TestPulseFloodsNonTimeStartOrigin` pins that the ignore is scoped to TimeStart senders.

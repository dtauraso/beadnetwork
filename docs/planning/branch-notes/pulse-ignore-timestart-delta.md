---
branch: task/pulse-ignore-timestart-delta
---

# Pulse ignores a delta triple from a TimeStart

## Rules

1. **Ignore TimeStart-origin.** A `Pulse` drops a cascade delta triple arriving from a
   `TimeStart`-kind neighbor: no record, no relay. Implemented in the
   `moveMsgKindDeltaForward` handler (`nodes/Wiring/node_mover.go`), the same shape and
   the same place as the existing TimeStart<-Input stop.
2. **Route gate-origin across to the opposite gate.** In `forwardDelta`, the TimeStart
   shape but for the two gates:

   | From sender kind | Go type | Routed to | Go type |
   |---|---|---|---|
   | `WindowAndInhibitLeftGate` | SelectRight | `WindowAndInhibitRightGate` | SelectLeft |
   | `WindowAndInhibitRightGate` | SelectLeft | `WindowAndInhibitLeftGate` | SelectRight |

   The kind strings cross over relative to the Go type names — read the table, not the
   name.

Any other sender kind keeps the plain flood to all cascade neighbors except the sender.

## Not fully live until the double link is restored

Node 5's cascade neighbors today are `{2: TimeStart, 9: WindowAndInhibitRightGate}` —
node 8 is NOT among them. So the SelectLeft->SelectRight half of rule 2 currently
resolves to a target that is not a cascade neighbor and reaches nobody. Both halves
only take effect once `5-8` is restored (`task/cascade-double-links`). The unit tests
construct the neighbor set with 8 present, so the rule is pinned either way.

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
| 5 | Pulse | everything except TimeStart | gate-origin routes to the opposite gate; otherwise plain flood |
| 6 | TimeEnd | — | never (terminus) |
| 7 | PulseRight | Time, SelectLeft | never (terminus) |

Everything else is still on the default plain flood.

## Verification

`TestPulseIgnoresTimeStartOriginDelta` fails without the change on both halves
(`gotForwardMsg=1, want 0` and `relayed to [9 8], want none`).
`TestPulseFloodsNonTimeStartOrigin` pins that the ignore is scoped to TimeStart senders.

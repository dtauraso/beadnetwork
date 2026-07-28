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

3. **Drop an unrecognized sender kind.** The Pulse switch is TOTAL. Rules 1 and 2
   cover every kind node 5 actually neighbours, so anything else is not a real case
   and relays to nobody — the stance TimeStart already takes.

## Why rule 3 exists: a data gap became surprise fan-out

Observed live: dragging node 8 made node 5 forward to node 2, which rule 2 forbids.

Cause: `5-8` is a DOMAIN edge, so `requantizeLocalPolars` (which fans `neighborSetC`
over domain neighbours) delivered the drag straight to 5 — but `5-8` was NOT in node
5's `cascade-edges.json`, so `cascadeKinds["8"]` read `""`. The sender was never
classified as SelectRight, rule 2 did not fire, and the delta fell through to the
old plain-flood default, reaching 2 and 9.

So BOTH halves of rule 2 were dead, not just the SelectLeft->SelectRight half: the
lookup that failed was the SENDER's, not the target's.

Two fixes, both applied:

- Restored the `5-8` double link in `cascade-edges.json` (nodes 5 and 8), so 5 can
  classify 8 as SelectRight.
- Closed the switch (rule 3), so a future missing entry goes inert instead of
  flooding.

Measured after: dragging node 8 produces exactly ONE delta-forward, `5 -> 9`.

Restoring `5-8` reinstates cycle `5-8-3-1-2-5`, which is cut by PulseLeft (node 3)
being a terminus. `7-9` is still not a cascade link — that half remains on
`task/cascade-double-links`.

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

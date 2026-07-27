---
branch: task/rightgate-usage
---

# Change how WindowAndInhibitRightGate is used

## Scope

Change how `WindowAndInhibitRightGate` is used. Not yet specified — this branch was
created as a placeholder so the work has a home.

## Naming, since it crosses over and is easy to get backwards

| Kind string (spec / `cascadeKinds`) | Go type | Node |
|---|---|---|
| `WindowAndInhibitLeftGate` | `SelectRight` | 8 |
| `WindowAndInhibitRightGate` | `SelectLeft` | 9 |

The kind string says Left where the Go type says Right, and vice versa. Always check
which one a given call site means.

`SelectLeft` (this branch's subject, node 9) fires on the raw `10` pattern directly,
no inversion — merged `1f05d07a`. Its sibling `SelectRight` (node 8) fires on raw
`01` via `gatecommon.RunGateAccept` — merged `9c84cb5e`.

## Existing couplings to keep in mind

- Node 9's cascade neighbors are `{5: Pulse}` today, plus `{7: PulseRight}` once the
  `7-9` double link is restored (`task/cascade-double-links`).
- Node 9 sits on cascade cycle B (`7-9-5-2-4-7`). That cycle is currently cut at node
  7, not at node 9 — PulseRight attends only to Time/SelectLeft and never relays
  (merged `c257f186`). If node 9's routing changes, re-check that the cut still holds.
- `PulseRight`'s whitelist names `WindowAndInhibitRightGate` explicitly
  (`nodes/Wiring/node_mover.go`), so renaming or repurposing this kind touches that
  guard and its tests in `pulseright_cascade_route_test.go`.

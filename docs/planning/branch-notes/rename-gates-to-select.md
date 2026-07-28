---
branch: task/rename-gates-to-select
---

# Rename both gate node kind strings to Select{Left,Right}

## Scope

Both of the AND-gate node kinds' registered kind strings previously spelled
out "Window And Inhibit ... Gate" while their Go types were already named
`SelectLeft`/`SelectRight` — a crossover between the kind string and the
type name, on both sides:

- Node 9: kind string (old verbose "Window And Inhibit Right Gate" spelling) ->
  `SelectLeft` (Go type was already `SelectLeft`). Landed first, on
  `task/rightgate-usage` (this doc's predecessor, now folded in here).
- Node 8: kind string (old verbose "Window And Inhibit Left Gate" spelling) ->
  `SelectRight` (Go type was already `SelectRight`). Landed on this branch.

Both renames touched Go registration, SPEC.md, topology data
(`topology/nodes/8/meta.json`, `topology/nodes/3,5/cascade-edges.json`), and
node_mover.go's cascade-routing rules/comments, so each kind string now
matches its Go type name. The package directories moved from their old
lowercased-kind-name paths to `nodes/selectleft/` and `nodes/selectright/`
respectively.

## Naming — crossover is now fully resolved

| Kind string (spec / `cascadeKinds`) | Go type | Node |
|---|---|---|
| `SelectRight` | `SelectRight` | 8 |
| `SelectLeft` | `SelectLeft` | 9 |

Both kind strings now match their Go type names — the historical crossover
(kind string said one side, Go type said the other) no longer exists for
either gate.

`SelectLeft` (node 9) fires on the raw `10` pattern directly, no inversion —
merged `1f05d07a`. `SelectRight` (node 8) fires on raw `01` via
`gatecommon.RunGateAccept` — merged `9c84cb5e`. Both fire via the shared
`gatecommon.RunGateAccept`/`RunGate` loop.

## Existing couplings to keep in mind

- Node 9's cascade neighbors are `{5: Pulse}` today, plus `{7: PulseRight}` once the
  `7-9` double link is restored (`task/cascade-double-links`).
- Node 9 sits on cascade cycle B (`7-9-5-2-4-7`). That cycle is currently cut at node
  7, not at node 9 — PulseRight attends only to Time/SelectLeft and never relays
  (merged `c257f186`). If node 9's routing changes, re-check that the cut still holds.
- Node 8's cascade neighbors are `{1: Input, 3: PulseLeft}` (via node 3's
  cascade-edges.json) and `{5: Pulse}` (via node 5's, the `5-8` double link). Node 8
  sits on cascade cycle A (`5-8-3-1-2`), cut at node 3 — PulseLeft attends only to
  Input/SelectRight and never relays.
- `PulseRight`'s whitelist previously named node 9's old kind string explicitly
  (`nodes/Wiring/node_mover.go`); it now names `SelectLeft`.
- `PulseLeft`'s whitelist and `Pulse`'s gate-crossover routing switch previously named
  node 8's old kind string explicitly; both now name `SelectRight` — updated on this
  branch, along with their tests (`pulseleft_cascade_route_test.go`,
  `pulse_cascade_route_test.go`, `pulseright_cascade_route_test.go`,
  `cascade_kind_route_test.go`).

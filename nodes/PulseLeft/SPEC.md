# PulseLeftNode

## View

| Field | Value |
|-------|-------|
| kindId | 8 |
| kind | pulseLeft |
| bg | #e1f5fe |
| border | #90caf9 |
| text | #01579b |
| accent | #90caf9 |
| minWidth | 90 |
| shape | rect |
| fill | #e1f5fe |
| stroke | #90caf9 |
| width | 90 |
| height | 60 |

## Ports

| Name | Direction | EdgeKind | Notes |
|------|-----------|----------|-------|
| In | in | chain | sole input: sampled value; updates the held value |
| Out | out | chain | continuously drives the held value (starts -1) |

## Firing rule

Sample-and-hold. Holds one int value (initialized to -1) and drives it out
continuously, even before any input arrives. When a value arrives on In,
the held value is updated and subsequent outputs emit the new value. The output
is not precondition-gated — PulseLeft self-emits -1 from the start.

A clone of the Pulse kind (nodes/pulse), split out for node 3's divergence —
the firing rule is still identical to Pulse; the divergence so far is the
layout-cascade rule below.

## Cascade delta rule

Layout-side only (`nodes/Wiring/node_mover.go`), independent of the firing rule:

- PulseLeft ATTENDS to a delta triple only when it arrives from an **Input** or a
  **SelectRight** cascade neighbor (SelectRight's kind string now matches its Go type
  name). A delta from any other sender kind is dropped outright — not recorded, not
  relayed.
- PulseLeft NEVER cascades the delta triple onward, from any sender, and not even
  when it is the direct recipient of a drag. It is a cascade **terminus**, so an
  attended delta ends here; attending only records the observability state.

Node 3's cascade neighbors are exactly `{1: Input, 8: SelectRight}`,
so the whitelist is a forward-looking guard against an other-kind neighbor.

## Runtime status

- Loader-registered: yes
- TSX render: present

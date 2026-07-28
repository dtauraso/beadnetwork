# PulseRightNode

## View

| Field | Value |
|-------|-------|
| kindId | 9 |
| kind | pulseRight |
| bg | #e1f5fe |
| border | #01579b |
| text | #01579b |
| accent | #01579b |
| minWidth | 90 |
| shape | rect |
| fill | #e1f5fe |
| stroke | #01579b |
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
is not precondition-gated — PulseRight self-emits -1 from the start.

A clone of the Pulse kind (nodes/pulse), split out for node 7's divergence —
the firing rule is still identical to Pulse; the divergence so far is the
layout-cascade rule below.

## Cascade delta rule

Layout-side only (`nodes/Wiring/node_mover.go`), independent of the firing rule.
The exact mirror of PulseLeft's rule (`nodes/PulseLeft/SPEC.md`), with the
whitelist naming the other side's kinds:

- PulseRight ATTENDS to a delta triple only when it arrives from a **Time** or a
  **SelectLeft** (kind string `WindowAndInhibitRightGate`) cascade neighbor. A
  delta from any other sender kind is dropped outright — not recorded, not relayed.
- PulseRight NEVER cascades the delta triple onward, from any sender, and not even
  when it is the direct recipient of a drag. It is a cascade **terminus**, so an
  attended delta ends here; attending only records the observability state.

Node 7's cascade neighbors are `{4: Time, 9: WindowAndInhibitRightGate}` — the `7-9`
double link is restored, and this terminus is what made restoring it safe: it cuts
the otherwise self-sustaining `2-4-7-9-5` cascade cycle. (Its mirror, PulseLeft at
node 3, cuts `5-8-3-1-2` the same way.) Cascade adjacency now equals domain
adjacency, so termination is a property of these per-kind rules, not of the edge set.

## Runtime status

- Loader-registered: yes
- TSX render: present

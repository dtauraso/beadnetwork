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
| In | in | chain | sole input: sampled value; updates the held value |
| Out | out | chain | continuously drives the held value (starts -1) |

## Firing rule

Sample-and-hold. Holds one int value (initialized to -1) and drives it out
continuously, even before any input arrives. When a value arrives on In,
the held value is updated and subsequent outputs emit the new value. The output
is not precondition-gated — PulseRight self-emits -1 from the start.

A clone of the Pulse kind (nodes/pulse). Its firing rule is identical to Pulse's;
the layout-cascade rule that was its only divergence is gone with the cascade
system, so this kind is currently a Pulse under another name.

## Runtime status

- Loader-registered: yes
- TSX render: present

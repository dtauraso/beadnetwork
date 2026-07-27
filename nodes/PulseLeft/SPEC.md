# PulseLeftNode

## View

| Field | Value |
|-------|-------|
| kindId | 8 |
| kind | pulseLeft |
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
is not precondition-gated — PulseLeft self-emits -1 from the start.

A clone of the Pulse kind (nodes/pulse), split out for node 3's future
divergence — behavior is currently identical to Pulse.

## Runtime status

- Loader-registered: yes
- TSX render: present

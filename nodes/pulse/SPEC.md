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
| In | in | chain | sole input: sampled value; updates the held value |
| Out | out | chain | continuously drives the held value (starts -1) |
| OutFanout | out | chain | optional second continuous output of the same held value (fan to a second destination); inert when unwired |

## Firing rule

Sample-and-hold. Holds one int value (initialized to -1) and drives it out
continuously, even before any input arrives. When a value arrives on In,
the held value is updated and subsequent outputs emit the new value. The output
is not precondition-gated — Pulse self-emits -1 from the start.

## Runtime status

- Loader-registered: yes
- TSX render: present

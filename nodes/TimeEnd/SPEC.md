# TimeEndNode

## View

| Field | Value |
|-------|-------|
| kindId | 0 |
| kind | timeEnd |
| bg | #ffccbc |
| border | #bf360c |
| text | #870000 |
| accent | #bf360c |
| minWidth | 60 |
| displays | held |
| shape | rect |
| fill | #ffccbc |
| stroke | #bf360c |
| width | 60 |
| height | 60 |

## Non-channel fields

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| Held | int | `data.state` | last value received and displayed (no downstream forward) |

## Ports

| Name | Direction | EdgeKind | Notes |
|------|-----------|----------|-------|
| In | in | chain | single received value to hold and display |

## Firing rule

On each value received on In:
1. Fire.
2. Update Held to the received value and emit the held bead.

TimeEnd is a terminal node: it holds the last received value and displays it; it has no output ports and sends nothing downstream.

## Runtime status

- Loader-registered: yes
- TSX render: present

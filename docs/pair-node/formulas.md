# Pair node — formulas

Every formula in the pair path, with the file it comes from.

Open this with **VS Code's Markdown preview** (⌘K V): the source links below open
the file as an editor tab. That is why this page is Markdown and its sibling
diagrams are HTML — a rendered HTML page runs in a webview, and a webview does
not open editor tabs.

The same tables are on [formulas.html](formulas.html) with the diagrams; only
here do the source names open.

## Constants

| constant | what it is | source |
| --- | --- | --- |
| `step = π/12 = 0.2617993877991494 rad` | 15°, the one quantum | [curve_params.go](../../nodes/Wiring/curve_params.go) |
| `PERP = 6` | a quarter turn, in steps | [node_geometry.go](../../nodes/Wiring/node_geometry.go) |
| `HALF = 2 × PERP = 12` | a half turn | [tilt_vector_channel.go](../../nodes/Wiring/tilt_vector_channel.go) |
| `FULL = 2 × HALF = 24` | a full turn | [tilt_vector_channel.go](../../nodes/Wiring/tilt_vector_channel.go) |

## Directions from t

| direction | node 1 | node 2 |
| --- | --- | --- |
| bottom | `t + HALF` | `t − HALF` |
| normal | `t + PERP + (odd ? HALF : 0)` | `t − PERP` |
| sent | `normal − 2 × PERP` | `normal + 2 × PERP` |
| sent, reduced | `t − PERP (+HALF odd)` | `t + PERP` |

Each kind's own file: [Node1/node.go](../../nodes/Node1/node.go) ·
[Node2/node.go](../../nodes/Node2/node.go).

## Pole parity

| formula | what it is | source |
| --- | --- | --- |
| `odd = floorDiv(t, HALF) mod 2 ≠ 0` | poles crossed; FLOOR, toward −∞ | [Node1/node.go](../../nodes/Node1/node.go) |

## The test

| formula | what it is | source |
| --- | --- | --- |
| `d = ((a − b) mod FULL + FULL) mod FULL` | separation, reduced to [0, 24) | [tilt_vector_channel.go](../../nodes/Wiring/tilt_vector_channel.go) |
| `acute ⟺ d < PERP ∨ d > FULL − PERP` | `d = 6` and `d = 18` are the halt | [tilt_vector_channel.go](../../nodes/Wiring/tilt_vector_channel.go) |

## The step

| what arrived | node 1 | node 2 |
| --- | --- | --- |
| `acute(v, top) →` | `t − 1` | `t + 1` |
| `acute(v, bottom) →` | `t + 1` | `t − 1` |
| `neither →` | hold, send nothing | hold, send nothing |

Each kind's own `stepFromVector`: [Node1/node.go](../../nodes/Node1/node.go) ·
[Node2/node.go](../../nodes/Node2/node.go).

## Clock

| formula | what it is | source |
| --- | --- | --- |
| `effective = userSpeed / divisor` | `divisor ≤ 0 → userSpeed` | [scene_speed_persist.go](../../nodes/Wiring/scene_speed_persist.go) |
| `divisor = 64` | the pair scene's own | [scene_tabs.go](../../nodes/Wiring/scene_tabs.go) |
| `pulses = clamp(⌈1 / speed⌉, 1, 64)` | `speed ≤ 0 → 64` | [wire/clock.go](../../nodes/wire/clock.go) |
| `one cycle = pulses × 16 ms` | one process-wide ticker | [wire/clock.go](../../nodes/wire/clock.go) |
| `HumanEditSpeed = 1.0` | broadcast while a ▲▼ click applies | [scene_speed_persist.go](../../nodes/Wiring/scene_speed_persist.go) |

## Index → screen

| formula | what it is | source |
| --- | --- | --- |
| `θ = idx × step` | radians, measured from world +y | [curve_params.go](../../nodes/Wiring/curve_params.go) |
| `x = r·sinθ·cosφ` | the one trig site | [gesture_camera.go](../../nodes/Wiring/gesture_camera.go) |
| `y = r·cosθ` | θ from +y, so y is the cosine | [gesture_camera.go](../../nodes/Wiring/gesture_camera.go) |
| `z = r·sinθ·sinφ` | the other in-ring component | [gesture_camera.go](../../nodes/Wiring/gesture_camera.go) |
| `r = nodeRadius(kind)` | arrow length; 0 = draw none | [port_geometry.go](../../nodes/Wiring/port_geometry.go) |
| `nodeRadius = outerR / (1 + tubeRatio)` | the ring's own radius | [port_geometry.go](../../nodes/Wiring/port_geometry.go) |

Four buffer columns carry these to the renderer: `TopTiltVectorLen`,
`TopTiltVectorTheta`, `BottomTiltVectorTheta`, `CoplanarNormalTheta`, plus
`ReceivedVectorLen`/`Theta` for the third arrow.

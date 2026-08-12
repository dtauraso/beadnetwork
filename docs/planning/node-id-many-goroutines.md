# A node id is a referent, not a goroutine

## The target

MODEL.md says "Each node is a goroutine." That is the invariant this change reverses.

A node id NAMES a thing that is drawn. Several goroutines serve one node id, each owning
one job and its own state, each writing its own stream. They are peers. Nothing about them
is shared except the id they tag their frames with, and the editor composes those frames
onto one drawn node.

The clock belongs to ONE of those jobs — bead animation. It does not belong to the others.

## Why now

`NodeMover.Run` (`nodes/Wiring/nodeactor/node_mover.go`) is one goroutine doing several
jobs behind one `SleepCycle`:

```go
g.msg.DrainPending(ctx, g.handle)    // node-move messages — a drag
g.outs.DriveOutWires(ctx, outTick)   // bead delivery
g.writeStreamFrame(nil)              // the geometry TS draws
g.clocks.SleepCycle(ctx)             // gates all three
```

`DrainPending` is wholly non-blocking, so `SleepCycle` alone sets the rate at which a drag
is read and re-streamed. `SleepCycle` waits `pulsesPerCycle() = ceil(1/speed)` pulses
(`nodes/clock/sleep_cycle.go`), so lowering bead speed lowers the input sampling rate with
it. Dragging a node at low speed moves it in steps at the bead rate.

The human-speed clock is reaching interaction through a path it should not have. Beads are
already correctly paced by `Tick()`, which scales elapsed time by speed
(`nodes/clock/real_clock.go`).

## The cut

Two goroutines per node id to start with, both already provisioned a stream fd by
`tools/topology-vscode/src/runner/spawn-layout.ts` (`node:`, `drive:`):

- **geometry** — owns `geom`, `topo`, `beads`, the node stream. Free-runs at ONE RAW PULSE
  (16ms, unscaled), draining its inbox non-blocking exactly as it does today. Never calls
  `SleepCycle`, so bead speed cannot reach it.
- **animation** — owns `outs` (the outgoing `PacedWire`s and their in-flight beads) and the
  clock. Calls `SleepCycle`, then `DriveOneCycle(tick)`.

`owners/outs.go` is already a self-contained owner, so this is a question of which goroutine
holds it, not new machinery.

## What crosses, and in which direction

Today `chain_beads.go:58-60` calls both sides inline from one goroutine:

```go
m.outs.PublishStepCount(to, count)        // geometry's answer
pulses := m.outs.GatherPulses(to, tick)   // animation's state
```

Once `outs` moves, these become the two messages, and they are the ONLY things that cross:

- geometry → animation: the step count per target, when its own or a neighbour's centre
  moves. `PublishSteps` is already message-shaped.
- animation → geometry: the gathered pulses per target, each cycle.

No shared struct, no mutex, no atomic — the network rule
(`tools/network/concurrency/check-no-network-locks.sh`, empty allowlist) holds by
construction rather than by discipline.

## Ripple list

- `nodes/Wiring/nodeactor/node_mover.go` — `Run` splits into the two loops.
- `nodes/Wiring/nodeactor/node_geometry.go` — `NodeGeometry` loses `outs` to the animation
  owner; the remaining fields stay with geometry.
- `nodes/Wiring/nodeactor/chain_beads.go` — the join point; reads pulses from the last
  message instead of calling `GatherPulses` directly.
- `nodes/clock/` — a `SleepPulse` that waits ONE pulse whatever the speed, beside the
  speed-scaled `SleepCycle`. Nothing blocks on a peer: a node polls its inboxes and does its
  own local work, which is why `DrainPending`'s `default:` selects stay exactly as they are.
- `nodes/Wiring/nodeactor/pair_node_self.go` — a second caller of `DriveOutWires` /
  `ClearOutWires`; it becomes another goroutine serving the same id, not a special case.
- MODEL.md "Each node is a goroutine" and `docs/model/entities.md` — the invariant itself.
- `.claude/rules/bridge-surface.md` — one goroutine, one stream, now per JOB not per node.

## Order

1. Model docs first, so the code is written against the stated invariant rather than the
   docs being back-filled to match whatever got built.
2. `SleepPulse` on the clock: one pulse, unscaled.
3. Move `outs` out of `NodeGeometry` into its own owner, with the two messages replacing the
   two inline calls.
4. Split `Run` into the two loops — geometry on `SleepPulse`, animation on `SleepCycle`.
5. `pair_node_self.go` becomes a peer goroutine on the same id.

## Verification

`bash scripts/stop-checks.sh` clean (EMPTY stdout), then drive the editor: drag a node with
the speed slider at its lowest setting. The node must follow the pointer smoothly while the
beads crawl. That is the whole point of the change and the only check that proves it.

Watch for a wedged inbox — `EnqueueSend`'s `maxPendingSends` panic names the node whose
peer stopped draining, which is what a blocking wait done wrong looks like.

## Risks

- Nothing waits on a peer, so there is no deadlock to reason about: both loops poll and
  sleep on their own pulse. Both directions of the new traffic send non-blocking, as
  `FlushPending` already does — a slow reader drops an update rather than stalling a clock.
- Geometry now wakes 1/speed times more often than before at low speed. That is the point,
  but it is also real work per pulse: `writeStreamFrame` should stay cheap, and emitting an
  unchanged frame every 16ms is the thing to watch.
- Delete this file when the change lands.

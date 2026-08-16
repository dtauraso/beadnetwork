# Split each node's geometry goroutine from its animation goroutine

## Why

Dragging node 1 kills node 1. The trace (`.probe/go-errors.jsonl`, 2026-08-15):

```
panic: NodeGeometry(1): pending exceeded 64 retry-queued sends; either a
destination's own goroutine has stopped draining its inbox (wedged or dead),
or this node is enqueueing to a peer faster than that peer drains, cycle over cycle

  layoutquant.BroadcastToPartners      broadcast_move.go:35
  layoutquant.CommitNodeMoveLocal      commit_node_move.go:19
  nodeactor.takeDragOfSelf             node_takes_msg.go:68
```

43 move-commits in 863 ms, then dead. The mechanism:

- Pointer-moves arrive at ~60/s and each one enqueues a `KindCenter` to every partner.
- A partner drains **at most one message per neighbour channel per `Self.Step`**
  (`owners/messaging.go:64`), and `Self.Step` runs once per sim cycle.
- `SleepCycle` waits `ceil(1/speed)` ticks, capped at 64, at 16 ms/tick
  (`nodes/clock/sleep_cycle.go:33`, `clock.go:8`) — up to ~1 s per drain at low speed.
- Fill rate is the human's hand; drain rate is the simulation's clock. The queue
  only grows, and `maxPendingSends` (64) is reached in about a second.

The bound is not the bug — it correctly named its own cause. The bug is that a
hand movement is paced by a simulated one.

## The invariant this reverses

MODEL.md lines 11–35 pin the opposite and must change in the same commit:

> **A node is ONE goroutine** … there is **no second goroutine per node id** — no
> mover goroutine beside `Update` … the same loop reads input — so input is served
> at that cadence too, **deliberately: one clock per node, not one clock per job.**

Agreed with David 2026-08-15: retired. A node id now names **two** goroutines —
one paced by the clock (animation), one driven by messages (geometry) — sharing
no memory.

Note this direction has been walked before in the other direction: MODEL.md
records that `PacedWire` used to have its own goroutine (`PacedWire.run`) and it
was removed. This change does **not** restore that; the wire stays passive and
stays owned by exactly one goroutine — now the animation one.

## Target

**Animation goroutine** = the existing kind `Update` loop, paced by the clock.
Owns: kind logic, `owners.Outs` (the `PacedWire`s), bead placement and driving,
the bead stream, the interior stream, `owners.Beads` bead-chain state.

**Geometry goroutine** = new, one per node id, paced by its OWN `clock.NewRealClock()`
at a fixed `SleepPulse` (16 ms, `nodes/clock/clock.go:8`) that the speed slider
never touches. Owns: `msg`, `deltas`, `geom` (own polar position), `outEdges`,
`topo`, `ui`, `tilt`, `flags`, the rule mesh, the node stream and the edge streams.

So the model goes from **one clock per node** to **one clock per job**: the
animation half keeps the sim clock, the geometry half runs at pointer rate.

**It never blocks, and there is no retry queue.** MODEL.md already forbids both —
"Blocking couples this goroutine's progress to another's", and a violated bound
means the code is wrong rather than something to grow or defer. So: non-blocking
polls only (the `PollRecv`/`TryRecvExternal`/`DrainPending` idiom already
everywhere in this codebase), no wake channel, and `pending`/`FlushPending`/
`maxPendingSends` are deleted outright rather than re-tuned.

## The seam — what crosses, and how

Two writes are shared state today, safe only because both halves run
sequentially in `PairNodeSelf.Step`. Both become messages.

1. **Wire geometry.** `OutEdges.DeriveGeometry` (`owners/out_edges.go:138-143`)
   reaches into the animation half:
   ```go
   if e.port != nil { e.port.SetGeom(e.steps, start, end) }
   if e.dest != nil { e.dest.ReviseInFlightGeometry(tick, e.steps, …) }
   ```
   Becomes: geometry sends `{edgeIdx, steps, segment}` on a per-node channel;
   the animation goroutine applies it at the top of its own pass, with its own
   tick. This also removes geometry's only need for a `tick`.

2. **Bead drag.** `take` calls `m.beads.StartBeadDrag()` / `EndBeadDrag()`
   (`node_takes_msg.go:24,27`), which write `BeadWakeGroup.wake`/`.settle`
   directly (`beadchain/bead_wake_group.go:25-31`). Becomes a message on the
   same channel.

Already message-passing, no change needed: `stream.selfEvents` is a channel
(depth 8, `owners/stream.go:51`), so either goroutine can post node-frame events.

## Fixing the backpressure at the same time

The split is **continuous quantities coalesce, discrete events queue**:

- **Continuous** — a neighbour's `KindCenter` and the FSM's `KindDrag` both carry
  a pure incremental `Delta` (`BroadcastToPartners` sends `Center: nil`, and the
  receiver only does `ShiftOtherBy(senderID, delta)`). N of them from one sender
  are exactly equivalent to ONE delta of their sum, so each gets a **depth-1
  latest-value slot** that merges on deposit. A slow reader skips intermediate
  positions and lands in the same place. Deposits never fail, so nothing retries.
- **Discrete** — select / hover / dragStart / dragEnd / tilt / rule edits are not
  summable and must not be dropped. They keep a small buffered channel, and they
  arrive at human *decision* rate (a click, not a pointer-move), so depth 8 is
  ample. Overflow is a loud panic, not a block and not a grow.

`PublishCenter` (`owners/messaging.go:179`) is already exactly the slot idiom for
`centerOut`, so this is existing vocabulary, not a new one.

Explicitly NOT doing: raising `maxPendingSends` or `inboxDepth`. Both buy a
longer drag before the same death — the shape `memory/feedback_go_vs_coordinator_bias.md`
warns off. `maxPendingSends` and the retry queue it bounded are deleted, not
raised: with a slot that always accepts, a send cannot fail, so a queue of
failed sends is not a safety net, it is dead code. The invariant worth keeping
loud moves into the slot itself — a non-summable message kind routed onto a
coalescing slot panics by name, which is the mistake a future change would
actually make.

## Order

1. MODEL.md: rewrite the one-goroutine paragraph to the two-goroutine model.
2. Add the geometry→animation channel and the apply-at-top-of-pass step;
   move the two shared writes onto it. Still one goroutine — no behaviour change.
3. Make the neighbour inbox a coalescing slot + wake channel.
4. Split `PairNodeSelf.Step` into `Step` (animation, same signature — the 9 kinds
   do not change) and a new `runGeometry(ctx)` loop; launch it per node.
5. Delete the now-dead per-cycle geometry work from the animation pass.

## Ripple

- `nodes/Wiring/nodeactor/`: `pair_node_self.go`, `node_takes_msg.go`,
  `node_geometry.go`, `node_animation.go`, `node_geometry_wire.go`.
- `nodes/Wiring/nodeactor/owners/`: `messaging.go`, `out_edges.go`, `outs.go`,
  `beads.go`, `consts.go`.
- `nodes/Wiring/dispatch/move_dispatch_movers.go` — launches the new goroutine.
- The 9 kinds calling `Self.Step(ctx, clk.Tick())` — **unchanged by design**;
  keeping that signature is what keeps this off every node package.
- MODEL.md; `.claude/rules/` if any restates the one-goroutine rule.

## Verification

- `bash scripts/stop-checks.sh` clean (empty stdout), incl.
  `check-no-network-locks.sh` — the split must not introduce a mutex or atomic.
- `go build -race` and drive the editor: drag node 1 for well over a second at
  low sim speed. Before: panic at ~43 moves. After: no panic, node tracks the
  pointer while beads keep moving at the clock's rate.
- `.probe/go-errors.jsonl` empty after the drag.
- Watch for the bug class this invites (`memory/feedback_industry_bug_class_scan.md`):
  a torn read of edge geometry — beads drawn against a segment the geometry
  goroutine has already moved. Mitigated by applying the revision at one point
  at the top of the animation pass, never mid-pass.

## Risks

- **Data race.** Two goroutines over one `NodeGeometry` is exactly the shape the
  project bans. Every field must land on one side or the other; anything left
  reachable from both is a defect, not a tradeoff. `-race` under a real drag is
  the check.
- **Bead/edge visual lag.** Geometry now moves faster than the animation applies
  it. Expected and correct — beads are clock-paced — but it will look different
  and needs David's eye.
- Delete this doc when the change lands.

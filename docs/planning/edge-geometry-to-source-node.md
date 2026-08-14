# Edge geometry moves to the source node; the edgeMover goroutine goes

## Target

An edge has no goroutine. Its geometry — segment, step count, persisted delta, and the
revision of beads already in flight — is derived by its SOURCE NODE, in that node's own
paced loop, from state the node already holds.

## Why it is deletion, not migration

The source node already computes all of it, every tick, in `owners/out_edges.go`
`WriteFrames`:

- `targetCenter := Polar2cart(Compose(ownPolar, d)).Add(self.SceneCenter)` — the
  destination's world position, from the node's own polar plus the stored neighbour vector
- `dir.Length()` — the `centerDist` that `EdgeStepCount` takes
- `e.targetKind` and `self.Kind` — the other two `EdgeStepCount` arguments
- `start`/`end` — the trimmed segment

The edgeMover derives the same values a second way, from mirrored absolute copies of both
endpoints (`srcGeom`/`dstGeom`) refreshed by `KindCenter`/`KindCenters` messages, and ships
the result back to the source node through two non-blocking channels on `Out`
(`geomSendSteps`/`geomSendSeg`) which the node drains in `Out.Geom()`.

One quantity, two representations, two update mechanisms — the node composes incrementally,
the edgeMover re-derives absolutely — and nothing reconciles them.

This is only safe to collapse now: before every kind became one goroutine, the node's
drawing half and its kind logic were separate goroutines, so a direct write to `Out` from
the drawing side would have raced the kind's `PlaceDrivenAt`. They are the same goroutine
now.

## Order (each step builds and runs)

1. `OutEdges` gains the per-edge `*outport.Out` and dest `*wire.PacedWire`, and a
   `DeriveGeometry` pass that computes segment + steps once, sets them on the port
   directly, revises in-flight geometry, and stores the segment for `WriteFrames` to reuse.
2. `PairNodeSelf.Step` calls `DeriveGeometry` BEFORE `driveOutWires`. Forced order:
   messages drain (positions change) -> geometry derives from positions -> beads move on
   that geometry -> frames draw. Deriving after the drive would pace one pass stale.
3. Delta persistence moves to the node, gated on the delta actually changing — the node
   derives every tick, so an ungated write would rewrite the edge JSON every tick.
4. `Out` loses `geomSendSteps`/`geomSendSeg`, `PublishSteps`/`PublishSegment`, and the
   drains in `Geom()`.
5. `EdgeMover` loses `recomputeGeometry`, `updateDeltaFromEndpoints`, the endpoint mirrors,
   and `handle`; then the type, its `Run` goroutine, and its inboxes.
6. The `edgeMovers` map becomes a plain edge table (src id, dst id, dest wire, out port) for
   its three remaining readers: `HeldEdges` (src/dst pairs), `SetEdgeStreams` (dest-wire
   lookup), `Bind`. Edge select is deleted outright — `handle` already ignores `KindSelect`.

## Ripple

- `MODEL.md` — an edge's geometry is its source node's, computed in the node's own loop.
- `.claude/rules/persistence-ownership.md` — currently claims "No Go writer exists yet —
  edges are editor-authored" for `nodes/<source>/edges/<label>.json`. That is wrong today
  (`edge_delta_persist.go` is the writer) and the owner changes here.
- `tools/network/persist/check-persist-write-ownership.sh` — the edge-file owner moves.
- `layoutquant/broadcast_move.go`, `commit_node_move.go`, `dispatch/move_dispatch_movers.go`,
  `viewpersist/enable_persist.go`, `gesture/gesture_select.go`, `streamwire`, `runtopology`.

## Verification

- `bash scripts/stop-checks.sh` prints nothing.
- Drive the editor: beads keep their speed (step count feeds `DwellTicksPerBead`), edges
  still draw and follow a dragged node, and a drag still persists.
- Prediction worth checking, not assuming: with persistence gated on change, a plain run
  should stop dirtying `topology/nodes/*/edges/*.json`.

## Risks

Step count feeds bead pacing, so an error here shows up as wrong-speed beads in a live
session — visible, not loud. That is the thing to drive after it lands.

Delete this file when the change lands.

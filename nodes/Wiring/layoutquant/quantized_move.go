// quantized_move.go — the quantized scene-polar move math, owned by LayoutQuantizer.
// Lifted out of nodes/Wiring/dispatch (docs/planning/movedispatch-decomposition.md §24):
// every method here already took its owners (a node/edge directory, a center reader) as
// explicit PARAMETERS rather than a *MoveDispatch back-reference (see this file's own
// pre-lift header comment, preserved in spirit below), so the coupling to the dispatch
// core was never through the receiver — LayoutQuantizer's own receiver type was always
// *layoutQuantizer, never *MoveDispatch/*moverRegistry, which is what made this move
// possible without exporting either of those types. This file keeps the seam it is named
// for — LayoutQuantizer's own held-state snapshots (HeldCenters/HeldEdges) and RootMove
// (the decentralized drag entry). The owner-goroutine commit (CommitNodeMoveLocal) is
// commit_node_move.go, the per-neighbour touching-bead resolution (DragTouchingBeads) is
// touching_beads.go, and the post-commit fan-out (BroadcastToEdgesAndPartners) is
// broadcast_move.go — each in this same package.
//
// There is no node-node stored coordinate here (MODEL.md "the polar model" — the deleted
// wire.LocalPolar and its requantize/neighborSetC propagation): a node has ONE polar
// coordinate, about the scene centre only.
//
// RootMove's invariant is load-bearing and MUST stay prominent after this move: it runs
// ONCE PER POINTER-MOVE EVENT, not once per drag (memory/project_rootmove_is_per_pointer_move.md).
//
// Every function below takes the exact directories/values it reads as explicit parameters
// (nodeGeoms, edgeMovers, ctx) instead of a *moverRegistry back-reference — the two maps
// hold already-exported types from already-lifted packages (*nodeactor.NodeGeometry,
// *edgemover.EdgeMover), so nothing needed exporting on the dispatch side to make this
// legal: dispatch's own callers pass md.mr.nodeGeoms/md.mr.edgeMovers directly (same
// pattern as binding a func value — passing the map itself, never a pointer into
// moverRegistry, never an exported moverRegistry field).
package layoutquant

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/edgemover"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// LayoutQuantizer owns the quantized scene-polar move math. See dispatch.MoveDispatch's
// own `lq` field doc comment for what it owns and why.
type LayoutQuantizer struct {
	// QuantizedLayout gates the quantized absolute-scene-polar snap — every node is a
	// root, measured/derived about the scene center only. Exported: dispatch's own
	// build_move_dispatch.go sets it once at load (`md.lq.QuantizedLayout = ...`), and
	// commit_node_move.go's CommitNodeMoveLocal reads it — both now cross this package's
	// boundary, so the field can no longer be unexported the way it was pre-move.
	QuantizedLayout bool
}

// HeldCenters returns a fresh snapshot of every node's current world center, read via
// centerOf (the dispatch goroutine's own centerMirror read, moverRegistry.centerOfNode,
// bound at the call site) — safe to call from the stdin/gesture dispatch goroutine, which
// is every live caller. There is no separate accumulated positions map to drain.
func HeldCenters(nodeGeoms map[string]*nodeactor.NodeGeometry, centerOf func(id string) (wire.Vec3, bool)) map[string]wire.Vec3 {
	out := make(map[string]wire.Vec3, len(nodeGeoms))
	for id := range nodeGeoms {
		if c, ok := centerOf(id); ok {
			out[id] = c
		}
	}
	return out
}

// HeldEdges returns every edge as a geom.SphereEdge (source/target id pair), read from
// the live edgeMovers directory.
func HeldEdges(edgeMovers map[string]*edgemover.EdgeMover) []geom.SphereEdge {
	edges := make([]geom.SphereEdge, 0, len(edgeMovers))
	for _, em := range edgeMovers {
		edges = append(edges, geom.SphereEdge{Source: em.SrcID(), Target: em.DstID()})
	}
	return edges
}

// RootMove handles a node-drag under the flat absolute scene-polar layout: every node
// is positioned independently about the scene sphere center — there is no reference/
// parent concept, so dragging moves ONLY the dragged node (no cascade). The dragged
// node's COMMITTED world position (CommitNodeMoveLocal) is the drag target SNAPPED to
// the scene lattice, moving exactly one bead distance per commit, the same distance its
// own chain beads move.
//
// RootMove is the decentralized drag entry, widened to EVERY node (the generalization
// that came with the quantizedOffsets data-race fix): dragging any node does not commit
// on the stdin reader's own goroutine — it routes a single movemsg.KindDrag to the
// dragged node's OWN inbox (nm.SendExternal, the same external-entry path moverRegistry's
// own sendMove used) and returns. The dragged node's own movemsg.KindDrag handler
// (nodeMover.handle) does the rest, entirely on its own goroutine: commit its own new
// position (commitLocal — fan + persist, no cross-goroutine self-send). There is no
// equal-radii solve, no rule-node cascade, no gate-anchor broadcast, and no per-neighbour
// re-quantize message any more (MODEL.md "the polar model": a node has no stored
// coordinate for a neighbour); a drag never touches any node's position but its own.
// Returns false for an unknown node.
//
// NOTE: RootMove runs ONCE PER POINTER-MOVE EVENT, not once per drag (two bugs — commits
// 338f05da, 154a05bd — came from assuming otherwise; see
// memory/project_rootmove_is_per_pointer_move.md). The drag-log reset is NOT emitted
// here for that reason: the reset belongs at the real drag-start edge (the
// pending→dragging transition in gesture.go), not on every move tick RootMove sees.
func (lq *LayoutQuantizer) RootMove(ctx context.Context, nodeGeoms map[string]*nodeactor.NodeGeometry, nodeID string, target wire.Vec3) bool {
	nm, ok := nodeGeoms[nodeID]
	if !ok {
		return false
	}
	// Route the drag itself to the dragged node's OWN inbox instead of committing on
	// the stdin reader's goroutine — every node's movemsg.KindDrag handler commits
	// (synchronous local apply, reported over reportCh) on its own goroutine. No
	// central commit call here. This is the bare external-entry path (no owning mover
	// goroutine to thread a ctx from), matching moverRegistry.sendMove's own shape —
	// it never fires the test-only tap (see nodeactor's EnqueueSend doc comment; only
	// a node's own EnqueueSend does).
	nm.SendExternal(ctx, movemsg.Msg{Kind: movemsg.KindDrag, NodeID: nodeID, Target: target})
	return true
}

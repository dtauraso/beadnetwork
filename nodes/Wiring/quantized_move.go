// quantized_move.go — the quantized scene-polar move math, owned by layoutQuantizer (god-
// object decomposition): pure move, no logic changes. This file keeps the seam it is
// named for — layoutQuantizer's own held-state snapshots (heldCenters/heldEdges) and
// RootMove (the decentralized drag entry). The owner-goroutine commit
// (commitNodeMoveLocal) is commit_node_move.go, the per-neighbour touching-bead
// resolution (dragTouchingBeads) is touching_beads.go, and the post-commit fan-out
// (broadcastToEdgesAndPartners) plus reachRFromPolar are broadcast_move.go — each split
// out of this file with no logic changes.
//
// There is no node-node stored coordinate here (MODEL.md "the polar model" — the deleted
// wire.LocalPolar and its requantize/neighborSetC propagation): a node has ONE polar
// coordinate, about the scene centre only.
//
// RootMove's invariant is load-bearing and MUST stay prominent after this move: it runs
// ONCE PER POINTER-MOVE EVENT, not once per drag (memory/project_rootmove_is_per_pointer_move.md).
//
// Every method below takes md *MoveDispatch explicitly for everything that is NOT part of
// layoutQuantizer's own field (quantizedLayout) — mr/ui/tr/persist/
// centerOfNode/NodeRowFor/sendMove are owned elsewhere. MoveDispatch's
// public RootMove, and its several package-private methods of the same names as below
// (heldCenters, heldEdges, broadcastToEdgesAndPartners, commitNodeMoveLocal), stay thin
// delegators in move_dispatch_api.go so their existing in-package call sites (tests,
// move_dispatch_construct.go, gesture.go) are unchanged.

package Wiring

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
)

// layoutQuantizer owns the quantized scene-polar move math. See MoveDispatch.lq's doc
// comment (move_dispatch.go) for what it owns and why.
type layoutQuantizer struct {
	// quantizedLayout gates the quantized absolute-scene-polar snap — every node is a
	// root, measured/derived about the scene center only.
	quantizedLayout bool
}

// heldCenters returns a fresh snapshot of every node's current world center, read from
// the dispatch goroutine's own centerMirror (centerOfNode), which each nodeMover keeps
// current by message (drainCenterMirror) — safe to call from the stdin/gesture dispatch
// goroutine, which is every live caller. There is no separate accumulated positions map
// to drain.
func (lq *layoutQuantizer) heldCenters(md *MoveDispatch) map[string]vec3 {
	out := make(map[string]vec3, len(md.mr.nodeGeoms))
	for id := range md.mr.nodeGeoms {
		if c, ok := md.centerOfNode(id); ok {
			out[id] = c
		}
	}
	return out
}

func (lq *layoutQuantizer) heldEdges(md *MoveDispatch) []geom.SphereEdge {
	edges := make([]geom.SphereEdge, 0, len(md.mr.edgeMovers))
	for _, em := range md.mr.edgeMovers {
		edges = append(edges, geom.SphereEdge{Source: em.srcID, Target: em.dstID})
	}
	return edges
}

// RootMove handles a node-drag under the flat absolute scene-polar layout: every node
// is positioned independently about the scene sphere center — there is no reference/
// parent concept, so dragging moves ONLY the dragged node (no cascade). The dragged
// node's COMMITTED world position (commitNodeMoveLocal) is the drag target SNAPPED to
// the scene lattice, moving exactly one bead distance per commit, the same distance its
// own chain beads move.
//
// RootMove is the decentralized drag entry, widened to EVERY node (the generalization
// that came with the quantizedOffsets data-race fix): dragging any node does not commit
// on the stdin reader's own goroutine — it routes a single movemsg.KindDrag to the
// dragged node's OWN inbox and returns. The dragged node's own movemsg.KindDrag handler
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
func (lq *layoutQuantizer) RootMove(md *MoveDispatch, nodeID string, target vec3) bool {
	if _, ok := md.mr.nodeGeoms[nodeID]; !ok {
		return false
	}
	// Route the drag itself to the dragged node's OWN inbox instead of committing on
	// the stdin reader's goroutine — every node's movemsg.KindDrag handler commits
	// (synchronous local apply, reported over reportCh) on its own goroutine. No
	// central commit call here.
	md.sendMove(nodeID, movemsg.Msg{Kind: movemsg.KindDrag, NodeID: nodeID, Target: target})
	return true
}

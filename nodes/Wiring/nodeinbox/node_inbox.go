// Package nodeinbox holds the DIRECTORIES OF DEDICATED PER-NODE CHANNELS that kinds claim
// for themselves at build time, one map per thing a node can be sent — lifted out of
// nodes/Wiring/dispatch (docs/planning/movedispatch-decomposition.md §29, the same lift
// §26 gave moverRegistry).
//
// It exists as an owner type rather than as loose MoveDispatch fields because that is what
// the composer rule asks for (check-composer-fields.sh): a new thing a node can be sent is
// a new entry HERE, and the composer's field count does not move. The two maps share one
// lifecycle exactly — written once per entry on the single-threaded build path (buildNodes,
// via BuildArgs), before any goroutine runs, and never touched again — so after build they
// are read-only lookup tables, which is what lets the stdin-reader goroutine read them
// without coordination.
//
// No channel is exported: ClaimLatticeIn/ClaimTiltEditIn take a channel IN (construction-
// time), and BroadcastLatticePoints/SendTiltEdit send ON the held channels internally —
// neither method hands a channel back out.
package nodeinbox

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"
)

// NodeInboxes holds the directories of dedicated per-node channels — see the package doc
// comment.
type NodeInboxes struct {
	// tiltEdit holds, for each node id whose OWN kind claimed BuildArgs.TiltEditIn (PairNode —
	// the only kind that owns its tilt index independently), that node's dedicated inbound
	// channel for a panel-driven tilt-angle click. A node id with no entry is a kind that
	// still routes tiltVectorAngle straight to its mover (applyUpdateTiltVector's fallback,
	// stdin_reader.go). Read only by SendTiltEdit.
	tiltEdit map[string]chan movemsg.TiltEditMsg

	// lattice holds, for each node id whose own kind claimed BuildArgs.LatticeIn (PairNode —
	// the only kind that owns a lattice), that node's dedicated inbound channel for a
	// scene-level point-count change. Read only by BroadcastLatticePoints, which sends to
	// every entry: the count is one scene-wide setting, so unlike a tilt edit it is
	// addressed to no single node.
	lattice map[string]chan int32
}

// ClaimLatticeIn registers id's dedicated LatticeIn channel — called once per node, at
// build time, before any goroutine runs (kindapi.BuildDeps' ClaimLatticeIn closure).
func (ib *NodeInboxes) ClaimLatticeIn(id string, ch chan int32) {
	if ib.lattice == nil {
		ib.lattice = map[string]chan int32{}
	}
	ib.lattice[id] = ch
}

// ClaimTiltEditIn registers id's dedicated TiltEditIn channel — called once per node, at
// build time, before any goroutine runs (kindapi.BuildDeps' ClaimTiltEditIn closure).
func (ib *NodeInboxes) ClaimTiltEditIn(id string, ch chan movemsg.TiltEditMsg) {
	if ib.tiltEdit == nil {
		ib.tiltEdit = map[string]chan movemsg.TiltEditMsg{}
	}
	ib.tiltEdit[id] = ch
}

// BroadcastLatticePoints sends a new lattice point count to every registered node's own
// dedicated LatticeIn channel, non-blocking latest-wins (drain-then-send, the same shape
// as wire.SendSpeedNonBlocking/SendLatestNonBlocking, just over a chan int32 instead of
// chan float64/int64) so a node that is mid-cycle never blocks the sender.
func (ib *NodeInboxes) BroadcastLatticePoints(points int32) {
	for _, ch := range ib.lattice {
		scenepersist.SendLatticePointsNonBlocking(ch, points)
	}
}

// SendTiltEdit routes one panel-driven tilt-angle click to node id's OWN dedicated
// tiltEditIns channel and returns true, or returns false when id has no such channel (a
// kind that never called BuildArgs.TiltEditIn — every kind except PairNode today),
// telling the caller (applyUpdateTiltVector) to fall back to the old mover-owned path
// instead. Same blocking-with-ctx-cancel-escape shape as moverreg.MoverRegistry.SendMove,
// for the same reason: this is a bare external-entry send with no owning goroutine to
// thread a ctx from.
func (ib *NodeInboxes) SendTiltEdit(ctx context.Context, id string, msg movemsg.TiltEditMsg) bool {
	ch, ok := ib.tiltEdit[id]
	if !ok {
		return false
	}
	if ctx == nil {
		ch <- msg
		return true
	}
	select {
	case ch <- msg:
	case <-ctx.Done():
	}
	return true
}

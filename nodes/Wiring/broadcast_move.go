// broadcast_move.go — fanning an already-applied set of moved node centers to incident
// edges/partners, and the reach-radius math that rides along with it.
//
// Split out of quantized_move.go (god-object decomposition, pure move — no logic
// changes): kept apart from held-state snapshots, touching-bead resolution, and the
// commit path itself.

package Wiring

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
)

// broadcastToEdgesAndPartners messages every incident edge's mover (batched per-edge Centers) and
// every aimed-port partner (pure re-emit), for the given already-applied set of moved node
// centers. It never writes the moved node's OWN snap — that responsibility belongs to
// whichever caller applied the moved node's own center via applyCenter directly (every
// live caller is owner-goroutine; the old central "self-send into own inbox" path,
// fanCenters, was removed — it deadlocked/staled when its only caller turned out to run
// on the moved node's own goroutine too. See commitNodeMoveLocal for the applyCenter +
// broadcastToEdgesAndPartners pattern).
func (lq *layoutQuantizer) broadcastToEdgesAndPartners(md *MoveDispatch, newCenters map[string]vec3, enqueue func(id string, msg movemsg.Msg)) {
	// Per-edge: send ONE batched message carrying every moved endpoint of that edge,
	// so an edge whose both endpoints moved this frame recomputes/emits exactly once.
	// enqueue (the sending node's own retry queue — nm.msg.sendMove) appends the
	// message to nm.pending and attempts an immediate non-blocking send on the
	// destination's own directed channel (extIn or the sender's slot in the
	// destination's neighborIn map), retrying next cycle if that channel isn't ready
	// to receive; it never blocks the calling handler goroutine, so this call — made
	// from inside handle via commitLocal — never blocks. Dispatch-existence (does id
	// resolve to a live mover) is checked at send time inside that retry path, matching
	// enqueue's other call sites (m.sendMove), which already tap/enqueue unconditionally
	// regardless of whether id resolves.
	for edgeID, em := range md.mr.edgeMovers {
		eps := map[string]vec3{}
		if c, ok := newCenters[em.srcID]; ok {
			eps[em.srcID] = c
		}
		if c, ok := newCenters[em.dstID]; ok {
			eps[em.dstID] = c
		}
		if len(eps) == 0 {
			continue
		}
		enqueue(edgeID, movemsg.Msg{Kind: movemsg.KindCenters, Centers: eps})
	}

	// Partner re-emit: find every partner node — the OTHER end of any edge incident to a
	// moved node — and ask it to re-emit its OWN geometry with its OWN (unchanged)
	// center. Node geometry no longer depends on a connected partner's position at all
	// (a port carries no geometry, docs/bead-model/channels-not-ports.md — this used to be how an
	// AIMED port picked up its moved partner's fresh center; that aiming is gone), but
	// the re-emit stays: it is what keeps a downstream watcher's view of this partner
	// current on the SAME cadence a moved node's own re-emit fires, without adding a
	// second, separately-timed signal.
	// partners maps partnerID → the ONE moved node (kept for clarity/observability
	// parity with the prior shape; movedID itself is otherwise unused now that the
	// re-emit carries no cache payload).
	partners := map[string]string{}
	for _, em := range md.mr.edgeMovers {
		if _, moved := newCenters[em.srcID]; moved {
			if _, alsoMoved := newCenters[em.dstID]; !alsoMoved {
				partners[em.dstID] = em.srcID
			}
		}
		if _, moved := newCenters[em.dstID]; moved {
			if _, alsoMoved := newCenters[em.srcID]; !alsoMoved {
				partners[em.srcID] = em.dstID
			}
		}
	}
	for partnerID, movedID := range partners {
		if _, ok := md.mr.nodeGeoms[partnerID]; !ok {
			continue
		}
		// Center is deliberately nil (see the doc comment above): this is a PURE
		// re-emit, not a position write for partnerID itself — a non-nil Center here
		// would be read by nodeMover.handle as "this is YOUR OWN new center" and
		// wrongly move partnerID. nodeMover.handle's nil-Center branch re-emits from
		// the mover's own live geom, so it can never race or clobber a pending position
		// write. Per-target FIFO order (each sender's own retry queue
		// drains in append order onto that target's one directed channel) preserves
		// ordering now that delivery goes through the sender's own
		// nm.pending/flushPending instead of a shared outbox.
		enqueue(partnerID, movemsg.Msg{Kind: movemsg.KindCenter, NodeID: partnerID, Center: nil,
			SenderID: movedID})
	}
}

// reachRFromPolar computes each node's sphere REACH radius (max distance from a node to any
// node it outputs to) under the given polar positions and edge set. Distance is the spherical
// law-of-cosines distance between the two polar positions (polarDist) — no cartesian, no vector
// subtraction. Called by loader.go buildFromSpec and by RootMove so the fanned "center" message
// carries the new reach radius and the ring stays sized during a drag.
func reachRFromPolar(polars map[string]geom.Polar, edges []geom.SphereEdge) map[string]float64 {
	reachR := map[string]float64{}
	for _, e := range edges {
		sp, okS := polars[e.Source]
		tp, okT := polars[e.Target]
		if !okS || !okT {
			continue
		}
		if d := geom.PolarDist(sp, tp); d > reachR[e.Source] {
			reachR[e.Source] = d
		}
	}
	return reachR
}
